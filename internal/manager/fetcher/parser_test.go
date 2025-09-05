package fetcher

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseSource(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected *SourceInfo
		wantErr  bool
	}{
		{
			name:  "local relative path",
			input: "./schemas",
			expected: &SourceInfo{
				Type:    SourceLocal,
				Source:  "./schemas",
				Version: "",
				Raw:     "./schemas",
			},
		},
		{
			name:  "local absolute path",
			input: "/home/user/schemas",
			expected: &SourceInfo{
				Type:    SourceLocal,
				Source:  "/home/user/schemas",
				Version: "",
				Raw:     "/home/user/schemas",
			},
		},
		{
			name:  "local parent relative path",
			input: "../schemas",
			expected: &SourceInfo{
				Type:    SourceLocal,
				Source:  "../schemas",
				Version: "",
				Raw:     "../schemas",
			},
		},
		{
			name:  "file scheme",
			input: "file:///home/user/schemas",
			expected: &SourceInfo{
				Type:    SourceLocal,
				Source:  "/home/user/schemas",
				Version: "",
				Raw:     "file:///home/user/schemas",
			},
		},
		{
			name:  "github with default version",
			input: "github://user/repo",
			expected: &SourceInfo{
				Type:    SourceGitHub,
				Source:  "user/repo",
				Version: "main",
				Raw:     "github://user/repo",
			},
		},
		{
			name:  "github with version",
			input: "github://user/repo@v1.0.0",
			expected: &SourceInfo{
				Type:    SourceGitHub,
				Source:  "user/repo",
				Version: "v1.0.0",
				Raw:     "github://user/repo@v1.0.0",
			},
		},
		{
			name:  "protodex with version",
			input: "protodex://my-service@v2.1.0",
			expected: &SourceInfo{
				Type:    SourceProtodex,
				Source:  "my-service",
				Version: "v2.1.0",
				Raw:     "protodex://my-service@v2.1.0",
			},
		},
		{
			name:  "protodex with default version",
			input: "protodex://my-service",
			expected: &SourceInfo{
				Type:    SourceProtodex,
				Source:  "my-service",
				Version: "latest",
				Raw:     "protodex://my-service",
			},
		},
		{
			name:  "http url",
			input: "https://example.com/schemas.zip",
			expected: &SourceInfo{
				Type:    SourceHTTP,
				Source:  "example.com/schemas.zip",
				Version: "",
				Raw:     "https://example.com/schemas.zip",
			},
		},
		{
			name:  "google well known",
			input: "google-well-known://protobuf",
			expected: &SourceInfo{
				Type:    SourceGoogleWellKnown,
				Source:  "protobuf",
				Version: "",
				Raw:     "google-well-known://protobuf",
			},
		},
		{
			name:    "empty input",
			input:   "",
			wantErr: true,
		},
		{
			name:    "unsupported scheme",
			input:   "ftp://example.com/file",
			wantErr: true,
		},
		{
			name:  "domain-like local path",
			input: "example.com/file",
			expected: &SourceInfo{
				Type:    SourceLocal,
				Source:  "example.com/file",
				Version: "",
				Raw:     "example.com/file",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseSource(tt.input)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, result)
			assert.Equal(t, tt.expected.Type, result.Type)
			assert.Equal(t, tt.expected.Source, result.Source)
			assert.Equal(t, tt.expected.Version, result.Version)
			assert.Equal(t, tt.expected.Raw, result.Raw)
		})
	}
}

func TestIsLocalPath(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"current directory", "./path", true},
		{"parent directory", "../path", true},
		{"absolute path", "/home/user", true},
		{"simple path", "simple", true},
		{"url with scheme", "http://example.com", false},
		{"url with at", "user@host.com", false},
		{"github style", "user/repo", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isLocalPath(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetDefaultVersion(t *testing.T) {
	tests := []struct {
		name         string
		sourceType   SourceType
		expectedVersion string
	}{
		{"github default", SourceGitHub, "main"},
		{"protodex default", SourceProtodex, "latest"},
		{"http default", SourceHTTP, ""},
		{"local default", SourceLocal, ""},
		{"google well known default", SourceGoogleWellKnown, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getDefaultVersion(tt.sourceType)
			assert.Equal(t, tt.expectedVersion, result)
		})
	}
}