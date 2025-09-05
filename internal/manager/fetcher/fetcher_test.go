package fetcher

import (
	"archive/zip"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewFetcher(t *testing.T) {
	tests := []struct {
		name       string
		sourceType SourceType
		source     string
		version    string
		dest       string
		wantErr    bool
	}{
		{
			name:       "local fetcher",
			sourceType: SourceLocal,
			source:     "/tmp/source",
			version:    "",
			dest:       "/tmp/dest",
		},
		{
			name:       "github fetcher",
			sourceType: SourceGitHub,
			source:     "user/repo",
			version:    "v1.0.0",
			dest:       "/tmp/dest",
		},
		{
			name:       "local fetcher with empty dest uses source",
			sourceType: SourceLocal,
			source:     "/tmp/source",
			version:    "",
			dest:       "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fetcher, err := NewFetcher(tt.sourceType, tt.source, tt.version, tt.dest)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.sourceType, fetcher.SourceType)
			assert.Equal(t, tt.source, fetcher.Source)
			assert.Equal(t, tt.version, fetcher.Version)

			// For local with empty dest, it should use source
			if tt.sourceType == SourceLocal && tt.dest == "" {
				assert.Equal(t, tt.source, fetcher.Dest)
			} else {
				assert.Equal(t, tt.dest, fetcher.Dest)
			}

			assert.NotNil(t, fetcher.client)
			assert.Equal(t, 30*time.Minute, fetcher.client.Timeout)
		})
	}
}

func TestNewFetcherFromURL(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		dest    string
		wantErr bool
	}{
		{
			name: "valid github url",
			url:  "github://user/repo@v1.0.0",
			dest: "/tmp/dest",
		},
		{
			name: "valid local path",
			url:  "./local/path",
			dest: "/tmp/dest",
		},
		{
			name:    "invalid url",
			url:     "invalid://url",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fetcher, err := NewFetcherFromURL(tt.url, tt.dest)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.NotNil(t, fetcher)
		})
	}
}

func TestLocalFetch(t *testing.T) {
	tests := []struct {
		name    string
		setup   func() (string, string) // returns source, dest
		wantErr bool
	}{
		{
			name: "valid local fetch with symlink",
			setup: func() (string, string) {
				tempDir := t.TempDir()
				sourceDir := filepath.Join(tempDir, "source")
				destDir := filepath.Join(tempDir, "dest")

				err := os.MkdirAll(sourceDir, 0755)
				require.NoError(t, err)

				testFile := filepath.Join(sourceDir, "test.proto")
				err = os.WriteFile(testFile, []byte("syntax = \"proto3\";"), 0644)
				require.NoError(t, err)

				return sourceDir, destDir
			},
		},
		{
			name: "same source and dest",
			setup: func() (string, string) {
				tempDir := t.TempDir()
				sourceDir := filepath.Join(tempDir, "source")

				err := os.MkdirAll(sourceDir, 0755)
				require.NoError(t, err)

				testFile := filepath.Join(sourceDir, "test.proto")
				err = os.WriteFile(testFile, []byte("syntax = \"proto3\";"), 0644)
				require.NoError(t, err)

				return sourceDir, sourceDir
			},
		},
		{
			name: "non-existent source",
			setup: func() (string, string) {
				tempDir := t.TempDir()
				return "/non/existent/path", filepath.Join(tempDir, "dest")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			source, dest := tt.setup()

			fetcher, err := NewFetcher(SourceLocal, source, "", dest)
			require.NoError(t, err)

			err = fetcher.Fetch()

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)

			// For same source and dest, no symlink should be created
			if source == dest {
				return
			}

			// Check if symlink was created
			linkInfo, err := os.Lstat(dest)
			require.NoError(t, err)
			assert.Equal(t, os.ModeSymlink, linkInfo.Mode()&os.ModeSymlink)
		})
	}
}

func TestIsFetched(t *testing.T) {
	tempDir := t.TempDir()

	tests := []struct {
		name     string
		setup    func(string) string // returns path to test
		expected bool
	}{
		{
			name: "non-existent directory",
			setup: func(dir string) string {
				return filepath.Join(dir, "non-existent")
			},
			expected: false,
		},
		{
			name: "empty directory",
			setup: func(dir string) string {
				emptyDir := filepath.Join(dir, "empty")
				os.MkdirAll(emptyDir, 0755)
				return emptyDir
			},
			expected: false,
		},
		{
			name: "directory without marker",
			setup: func(dir string) string {
				noMarkerDir := filepath.Join(dir, "no-marker")
				os.MkdirAll(noMarkerDir, 0755)
				os.WriteFile(filepath.Join(noMarkerDir, "file.txt"), []byte("content"), 0644)
				return noMarkerDir
			},
			expected: false,
		},
		{
			name: "directory with marker",
			setup: func(dir string) string {
				markerDir := filepath.Join(dir, "with-marker")
				os.MkdirAll(markerDir, 0755)
				os.WriteFile(filepath.Join(markerDir, "file.txt"), []byte("content"), 0644)
				os.WriteFile(filepath.Join(markerDir, FetchedMarkerFile), []byte(""), 0644)
				return markerDir
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testPath := tt.setup(tempDir)
			result := isFetched(testPath)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMarkAsFetched(t *testing.T) {
	tempDir := t.TempDir()
	destDir := filepath.Join(tempDir, "test-dest")
	err := os.MkdirAll(destDir, 0755)
	require.NoError(t, err)

	fetcher, err := NewFetcher(SourceLocal, "/dummy", "", destDir)
	require.NoError(t, err)

	err = fetcher.markAsFetched()
	require.NoError(t, err)

	// Check if marker file was created
	markerPath := filepath.Join(destDir, FetchedMarkerFile)
	_, err = os.Stat(markerPath)
	assert.NoError(t, err)
}

func TestDownload(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/test.zip":
			w.Header().Set("Content-Type", "application/zip")
			w.Write([]byte("fake zip content"))
		case "/notfound":
			w.WriteHeader(http.StatusNotFound)
		default:
			w.WriteHeader(http.StatusBadRequest)
		}
	}))
	defer server.Close()

	fetcher, err := NewFetcher(SourceHTTP, "", "", "")
	require.NoError(t, err)

	tests := []struct {
		name    string
		url     string
		wantErr bool
	}{
		{
			name: "successful download",
			url:  server.URL + "/test.zip",
		},
		{
			name:    "404 error",
			url:     server.URL + "/notfound",
			wantErr: true,
		},
		{
			name:    "invalid url",
			url:     "://invalid",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempFile, err := fetcher.download(tt.url)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.NotEmpty(t, tempFile)

			// Clean up
			if tempFile != "" {
				os.Remove(tempFile)
			}
		})
	}
}

func TestExtractZip(t *testing.T) {
	tempDir := t.TempDir()

	// Create a test zip file
	zipPath := filepath.Join(tempDir, "test.zip")
	file, err := os.Create(zipPath)
	require.NoError(t, err)

	zipWriter := zip.NewWriter(file)

	// Add a file to the zip
	f, err := zipWriter.Create("test/file.txt")
	require.NoError(t, err)
	_, err = io.WriteString(f, "test content")
	require.NoError(t, err)

	// Add a directory (will be skipped)
	_, err = zipWriter.Create("test/")
	require.NoError(t, err)

	err = zipWriter.Close()
	require.NoError(t, err)
	err = file.Close()
	require.NoError(t, err)

	// Test extraction
	extractDir := filepath.Join(tempDir, "extract")
	err = extractZip(zipPath, extractDir)
	require.NoError(t, err)

	// Check if file was extracted
	extractedFile := filepath.Join(extractDir, "test", "file.txt")
	content, err := os.ReadFile(extractedFile)
	require.NoError(t, err)
	assert.Equal(t, "test content", string(content))
}

func TestURLFetch(t *testing.T) {
	tempDir := t.TempDir()

	// Create a test zip server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/test.zip" {
			// Create a minimal zip file in memory
			w.Header().Set("Content-Type", "application/zip")

			zipWriter := zip.NewWriter(w)
			f, err := zipWriter.Create("test.proto")
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			io.WriteString(f, "syntax = \"proto3\";")
			zipWriter.Close()
		} else {
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	tests := []struct {
		name    string
		source  string
		wantErr bool
	}{
		{
			name:   "successful http fetch",
			source: server.URL + "/test.zip",
		},
		{
			name:    "invalid scheme",
			source:  "ftp://example.com/file.zip",
			wantErr: true,
		},
		{
			name:    "invalid url",
			source:  "://invalid",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			destDir := filepath.Join(tempDir, tt.name)
			fetcher, err := NewFetcher(SourceHTTP, tt.source, "", destDir)
			require.NoError(t, err)

			err = fetcher.urlFetch()

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)

			// Check if marker file was created
			markerPath := filepath.Join(destDir, FetchedMarkerFile)
			_, err = os.Stat(markerPath)
			assert.NoError(t, err)
		})
	}
}

func TestFetchUnsupportedType(t *testing.T) {
	fetcher := &Fetcher{
		SourceType: SourceType("unsupported"),
		Source:     "test",
		Dest:       "/tmp/test",
	}

	err := fetcher.Fetch()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported source type")
}

func TestFetchAlreadyFetched(t *testing.T) {
	tempDir := t.TempDir()
	destDir := filepath.Join(tempDir, "dest")

	// Set up a directory that appears to be already fetched
	err := os.MkdirAll(destDir, 0755)
	require.NoError(t, err)
	err = os.WriteFile(filepath.Join(destDir, "test.proto"), []byte("content"), 0644)
	require.NoError(t, err)
	err = os.WriteFile(filepath.Join(destDir, FetchedMarkerFile), []byte(""), 0644)
	require.NoError(t, err)

	// Create a simple test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("Server should not be called for already fetched content")
	}))
	defer server.Close()

	fetcher, err := NewFetcher(SourceHTTP, server.URL+"/test.zip", "", destDir)
	require.NoError(t, err)

	err = fetcher.urlFetch()
	assert.NoError(t, err) // Should succeed without calling server
}