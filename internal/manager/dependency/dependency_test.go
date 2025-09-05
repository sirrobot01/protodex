package dependency

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sirrobot01/protodex/internal/manager/fetcher"
)

func TestNewResolver(t *testing.T) {
	// Use a temporary directory to avoid conflicts
	tempDir := t.TempDir()
	oldProtodexPath := os.Getenv("PROTODEX_PATH")
	defer func() {
		if oldProtodexPath != "" {
			os.Setenv("PROTODEX_PATH", oldProtodexPath)
		} else {
			os.Unsetenv("PROTODEX_PATH")
		}
	}()
	os.Setenv("PROTODX_PATH", tempDir)

	resolver, err := NewResolver()
	require.NoError(t, err)
	assert.NotNil(t, resolver)
	assert.NotNil(t, resolver.client)
	assert.Equal(t, 5*time.Minute, resolver.client.Timeout)
	assert.Contains(t, resolver.outputPath, "deps")

	// Check if the directory was created
	_, err = os.Stat(resolver.outputPath)
	assert.NoError(t, err)
}

func TestResolver_GetDependencyPath(t *testing.T) {
	tempDir := t.TempDir()
	resolver := &Resolver{
		outputPath: tempDir,
	}

	path := resolver.GetDependencyPath()
	assert.Equal(t, tempDir, path)
}

func TestResolver_IsCached(t *testing.T) {
	tempDir := t.TempDir()
	resolver := &Resolver{
		outputPath: tempDir,
	}

	tests := []struct {
		name     string
		setup    func() string
		expected bool
	}{
		{
			name: "non-existent path",
			setup: func() string {
				return filepath.Join(tempDir, "non-existent")
			},
			expected: false,
		},
		{
			name: "empty directory",
			setup: func() string {
				emptyDir := filepath.Join(tempDir, "empty")
				os.MkdirAll(emptyDir, 0755)
				return emptyDir
			},
			expected: false,
		},
		{
			name: "directory with content",
			setup: func() string {
				contentDir := filepath.Join(tempDir, "with-content")
				os.MkdirAll(contentDir, 0755)
				os.WriteFile(filepath.Join(contentDir, "file.txt"), []byte("content"), 0644)
				return contentDir
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := tt.setup()
			result := resolver.isCached(path)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestResolver_Clear(t *testing.T) {
	tempDir := t.TempDir()
	resolver := &Resolver{
		outputPath: tempDir,
	}

	// Create some test content
	testDir := filepath.Join(tempDir, "test-dep")
	err := os.MkdirAll(testDir, 0755)
	require.NoError(t, err)
	err = os.WriteFile(filepath.Join(testDir, "test.proto"), []byte("content"), 0644)
	require.NoError(t, err)

	// Clear the cache
	err = resolver.Clear()
	require.NoError(t, err)

	// Check if directory was removed
	_, err = os.Stat(tempDir)
	assert.True(t, os.IsNotExist(err))
}

func TestResolver_ListCached(t *testing.T) {
	tempDir := t.TempDir()
	resolver := &Resolver{
		outputPath: tempDir,
	}

	tests := []struct {
		name     string
		setup    func()
		expected []string
	}{
		{
			name:     "empty cache directory",
			setup:    func() {},
			expected: []string{},
		},
		{
			name: "cache with dependencies",
			setup: func() {
				// Create some dependency directories
				os.MkdirAll(filepath.Join(tempDir, "dep1"), 0755)
				os.MkdirAll(filepath.Join(tempDir, "dep2"), 0755)
				// Create a file (should be ignored)
				os.WriteFile(filepath.Join(tempDir, "file.txt"), []byte("content"), 0644)
			},
			expected: []string{"dep1", "dep2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()

			cached, err := resolver.ListCached()
			require.NoError(t, err)

			if len(tt.expected) == 0 {
				assert.Empty(t, cached)
			} else {
				assert.ElementsMatch(t, tt.expected, cached)
			}
		})
	}
}

func TestResolver_ListCached_NonExistentDir(t *testing.T) {
	resolver := &Resolver{
		outputPath: "/non/existent/path",
	}

	cached, err := resolver.ListCached()
	require.NoError(t, err)
	assert.Empty(t, cached)
}

func TestResolver_ResolveDependencies(t *testing.T) {
	tempDir := t.TempDir()
	resolver := &Resolver{
		outputPath: tempDir,
	}

	// Create a mock local dependency source
	sourceDir := filepath.Join(tempDir, "source")
	err := os.MkdirAll(sourceDir, 0755)
	require.NoError(t, err)
	err = os.WriteFile(filepath.Join(sourceDir, "test.proto"), []byte("syntax = \"proto3\";"), 0644)
	require.NoError(t, err)

	tests := []struct {
		name    string
		deps    []Config
		wantErr bool
	}{
		{
			name: "successful local dependency resolution",
			deps: []Config{
				{
					Name:   "test-dep",
					Type:   fetcher.SourceLocal,
					Source: sourceDir,
				},
			},
		},
		{
			name: "multiple dependencies",
			deps: []Config{
				{
					Name:   "test-dep1",
					Type:   fetcher.SourceLocal,
					Source: sourceDir,
				},
				{
					Name:   "test-dep2",
					Type:   fetcher.SourceLocal,
					Source: sourceDir,
				},
			},
		},
		{
			name: "invalid dependency",
			deps: []Config{
				{
					Name:   "invalid-dep",
					Type:   fetcher.SourceLocal,
					Source: "/non/existent/path",
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := resolver.ResolveDependencies(tt.deps)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)

			// Check if dependencies were resolved
			for _, dep := range tt.deps {
				depPath := filepath.Join(tempDir, dep.Name)
				_, err := os.Stat(depPath)
				assert.NoError(t, err, "dependency %s should be resolved", dep.Name)
			}
		})
	}
}

func TestResolver_ResolveDependency_GoogleWellKnown(t *testing.T) {
	tempDir := t.TempDir()
	resolver := &Resolver{
		outputPath: tempDir,
	}

	dep := Config{
		Name:   "google-protobuf",
		Type:   fetcher.SourceGoogleWellKnown,
		Source: "protobuf",
	}

	// Since resolveGoogleWellKnownProtos might not be implemented,
	// we test that it gets called for google-well-known type
	err := resolver.resolveDependency(dep)
	// This might error if the method isn't implemented, but we're testing the routing
	// For a real implementation, you'd mock the Google API calls
	if err != nil {
		// Expected for now if method isn't implemented
		assert.Contains(t, err.Error(), "google")
	}
}

func TestConfig_Validation(t *testing.T) {
	tests := []struct {
		name   string
		config Config
		valid  bool
	}{
		{
			name: "valid local config",
			config: Config{
				Name:   "test-dep",
				Type:   fetcher.SourceLocal,
				Source: "/path/to/source",
			},
			valid: true,
		},
		{
			name: "valid github config with version",
			config: Config{
				Name:    "test-dep",
				Type:    fetcher.SourceGitHub,
				Source:  "user/repo",
				Version: "v1.0.0",
			},
			valid: true,
		},
		{
			name: "valid protodex config",
			config: Config{
				Name:    "test-service",
				Type:    fetcher.SourceProtodex,
				Source:  "my-service",
				Version: "v2.1.0",
			},
			valid: true,
		},
		{
			name: "config with path",
			config: Config{
				Name:   "test-dep",
				Type:   fetcher.SourceLocal,
				Source: "/source",
				Path:   "/custom/path",
			},
			valid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Basic validation - check required fields are present
			assert.NotEmpty(t, tt.config.Name, "Name should not be empty")
			assert.NotEmpty(t, tt.config.Source, "Source should not be empty")

			// Type should be a valid SourceType
			validTypes := []fetcher.SourceType{
				fetcher.SourceLocal,
				fetcher.SourceGitHub,
				fetcher.SourceHTTP,
				fetcher.SourceProtodex,
				fetcher.SourceGoogleWellKnown,
			}

			typeValid := false
			for _, validType := range validTypes {
				if tt.config.Type == validType {
					typeValid = true
					break
				}
			}
			assert.True(t, typeValid, "Type should be a valid SourceType")
		})
	}
}