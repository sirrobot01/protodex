package dependency

import (
	"archive/zip"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sirrobot01/protodex/internal/manager/fetcher"
)

func TestResolver_ResolveGoogleWellKnownProtos(t *testing.T) {
	tempDir := t.TempDir()

	// Create a mock zip file with protobuf includes
	mockZipPath := createMockProtocZip(t, tempDir)

	// Create a test server to serve the mock zip
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "protoc-") && strings.HasSuffix(r.URL.Path, ".zip") {
			// Serve the mock zip file
			http.ServeFile(w, r, mockZipPath)
		} else {
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	resolver := &Resolver{
		outputPath: tempDir,
		client:     &http.Client{},
	}

	tests := []struct {
		name    string
		dep     Config
		setup   func()
		wantErr bool
	}{
		{
			name: "successful google protobuf download with default version",
			dep: Config{
				Name: "google-protobuf",
				Type: fetcher.SourceGoogleWellKnown,
			},
			wantErr: false, // Actually works - downloads from real GitHub
		},
		{
			name: "already cached dependency",
			dep: Config{
				Name: "google-protobuf-cached",
				Type: fetcher.SourceGoogleWellKnown,
			},
			setup: func() {
				// Pre-create the cached directory
				cachedDir := filepath.Join(tempDir, "google-protobuf-cached")
				os.MkdirAll(cachedDir, 0755)
				os.WriteFile(filepath.Join(cachedDir, "test.proto"), []byte("content"), 0644)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup()
			}

			err := resolver.resolveGoogleWellKnownProtos(tt.dep)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)

			// Check if the target directory was created
			targetDir := filepath.Join(tempDir, tt.dep.Name)
			_, err = os.Stat(targetDir)
			assert.NoError(t, err)
		})
	}
}

func TestExtractIncludes(t *testing.T) {
	tempDir := t.TempDir()

	// Create a test zip file with include structure
	zipPath := filepath.Join(tempDir, "test.zip")
	targetDir := filepath.Join(tempDir, "extracted")

	createTestIncludeZip(t, zipPath)

	err := extractIncludes(zipPath, targetDir)
	require.NoError(t, err)

	// Check if files were extracted correctly
	expectedFiles := []string{
		"any.proto",
		"timestamp.proto",
		"compiler/plugin.proto",
	}

	for _, expectedFile := range expectedFiles {
		filePath := filepath.Join(targetDir, expectedFile)
		_, err := os.Stat(filePath)
		assert.NoError(t, err, "Expected file %s should exist", expectedFile)
	}

	// Check that non-include files were not extracted
	unexpectedFile := filepath.Join(targetDir, "bin", "protoc")
	_, err = os.Stat(unexpectedFile)
	assert.True(t, os.IsNotExist(err), "Unexpected file should not be extracted")
}

func TestExtractFile(t *testing.T) {
	tempDir := t.TempDir()

	// Create a test zip file
	zipPath := filepath.Join(tempDir, "test.zip")
	file, err := os.Create(zipPath)
	require.NoError(t, err)

	zipWriter := zip.NewWriter(file)
	testContent := "syntax = \"proto3\";\npackage test;"

	f, err := zipWriter.Create("test.proto")
	require.NoError(t, err)
	_, err = io.WriteString(f, testContent)
	require.NoError(t, err)

	err = zipWriter.Close()
	require.NoError(t, err)
	err = file.Close()
	require.NoError(t, err)

	// Open the zip for reading
	reader, err := zip.OpenReader(zipPath)
	require.NoError(t, err)
	defer reader.Close()

	// Extract the file
	targetPath := filepath.Join(tempDir, "extracted.proto")
	err = extractFile(reader.File[0], targetPath)
	require.NoError(t, err)

	// Verify content
	content, err := os.ReadFile(targetPath)
	require.NoError(t, err)
	assert.Equal(t, testContent, string(content))
}

func TestGoogleProtocVersionConstant(t *testing.T) {
	// Test that the constant is set to a reasonable value
	assert.NotEmpty(t, googleProtocVersion)
	assert.True(t, strings.HasPrefix(googleProtocVersion, "v"))
}

// Helper function to create a mock protoc zip file
func createMockProtocZip(t *testing.T, tempDir string) string {
	zipPath := filepath.Join(tempDir, "protoc-test.zip")
	file, err := os.Create(zipPath)
	require.NoError(t, err)

	zipWriter := zip.NewWriter(file)

	// Add some mock include files
	includeFiles := map[string]string{
		"include/google/protobuf/any.proto":              "syntax = \"proto3\";",
		"include/google/protobuf/timestamp.proto":        "syntax = \"proto3\";",
		"include/google/protobuf/compiler/plugin.proto": "syntax = \"proto3\";",
		"bin/protoc":                                     "binary content", // Should be ignored
	}

	for fileName, content := range includeFiles {
		f, err := zipWriter.Create(fileName)
		require.NoError(t, err)
		_, err = io.WriteString(f, content)
		require.NoError(t, err)
	}

	err = zipWriter.Close()
	require.NoError(t, err)
	err = file.Close()
	require.NoError(t, err)

	return zipPath
}

// Helper function to create a test zip with include structure
func createTestIncludeZip(t *testing.T, zipPath string) {
	file, err := os.Create(zipPath)
	require.NoError(t, err)
	defer file.Close()

	zipWriter := zip.NewWriter(file)
	defer zipWriter.Close()

	// Create mock include files
	includeFiles := map[string]string{
		"include/google/protobuf/any.proto":              "syntax = \"proto3\";",
		"include/google/protobuf/timestamp.proto":        "syntax = \"proto3\";",
		"include/google/protobuf/compiler/plugin.proto": "syntax = \"proto3\";",
		"bin/protoc":                                     "binary", // Should be ignored
		"readme.txt":                                     "info",   // Should be ignored
	}

	for fileName, content := range includeFiles {
		f, err := zipWriter.Create(fileName)
		require.NoError(t, err)
		_, err = io.WriteString(f, content)
		require.NoError(t, err)
	}
}