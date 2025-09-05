package client

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sirrobot01/protodex/internal/config"
	"github.com/sirrobot01/protodex/internal/logger"
)

func TestParsePackageRef(t *testing.T) {
	tests := []struct {
		name        string
		packageRef  string
		expectedPkg string
		expectedVer string
		expectError bool
	}{
		{
			name:        "Valid package with version",
			packageRef:  "user-service:v1.0.0",
			expectedPkg: "user-service",
			expectedVer: "v1.0.0",
			expectError: false,
		},
		{
			name:        "Valid package with latest",
			packageRef:  "user-service:latest",
			expectedPkg: "user-service",
			expectedVer: "latest",
			expectError: false,
		},
		{
			name:        "Package without version",
			packageRef:  "user-service",
			expectedPkg: "",
			expectedVer: "",
			expectError: true,
		},
		{
			name:        "Empty string",
			packageRef:  "",
			expectedPkg: "",
			expectedVer: "",
			expectError: true,
		},
		{
			name:        "Multiple colons",
			packageRef:  "user:service:v1.0.0",
			expectedPkg: "",
			expectedVer: "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pkg, ver, err := ParsePackageRef(tt.packageRef)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedPkg, pkg)
				assert.Equal(t, tt.expectedVer, ver)
			}
		})
	}
}

func TestNewClient(t *testing.T) {
	// Test with default config
	client, err := New()
	require.NoError(t, err)
	assert.NotNil(t, client)

	// Test with token
	client, err = NewWithToken("test-token")
	require.NoError(t, err)
	assert.NotNil(t, client)
}

func TestClientLogin(t *testing.T) {
	// Mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/auth/login", r.URL.Path)
		assert.Equal(t, "POST", r.Method)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"token": "test-token-123",
			"user": {
				"id": "user123",
				"username": "testuser"
			}
		}`))
	}))
	defer server.Close()

	client := newTestClient(server.URL, "")

	resp, err := client.Login("testuser", "password123")
	require.NoError(t, err)

	assert.Equal(t, "test-token-123", resp.Token)
	assert.Equal(t, "user123", resp.User.ID)
	assert.Equal(t, "testuser", resp.User.Username)
}

func TestClientLoginError(t *testing.T) {
	// Mock server with error response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error": "Invalid username or password"}`))
	}))
	defer server.Close()

	client := newTestClient(server.URL, "")

	_, err := client.Login("testuser", "wrongpassword")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "login failed")
}

func TestClientRegister(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/auth/register", r.URL.Path)
		assert.Equal(t, "POST", r.Method)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{
			"token": "test-token-456",
			"user": {
				"id": "user456",
				"username": "newuser"
			}
		}`))
	}))
	defer server.Close()

	client := newTestClient(server.URL, "")

	resp, err := client.Register("newuser", "password123")
	require.NoError(t, err)

	assert.Equal(t, "test-token-456", resp.Token)
	assert.Equal(t, "user456", resp.User.ID)
	assert.Equal(t, "newuser", resp.User.Username)
}

func TestClientListPackages(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/packages", r.URL.Path)
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`[
			{
				"id": "pkg1",
				"name": "user-service",
				"description": "User management service",
				"tags": ["grpc", "user"],
				"visibility": "public",
				"created_at": "2023-01-01T00:00:00Z"
			}
		]`))
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")

	packages, err := client.ListPackages()
	require.NoError(t, err)

	assert.Len(t, packages, 1)
	assert.Equal(t, "pkg1", packages[0].ID)
	assert.Equal(t, "user-service", packages[0].Name)
	assert.Equal(t, "User management service", packages[0].Description)
	assert.Equal(t, []string{"grpc", "user"}, packages[0].Tags)
}

func TestClientGetPackage(t *testing.T) {
	packageName := "test-package"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, fmt.Sprintf("/api/packages/%s", packageName), r.URL.Path)
		assert.Equal(t, "GET", r.Method)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(fmt.Sprintf(`{
			"id": "pkg123",
			"name": "%s",
			"description": "Test package",
			"tags": ["test"],
			"created_at": "2023-01-01T00:00:00Z"
		}`, packageName)))
	}))
	defer server.Close()

	client := newTestClient(server.URL, "test-token")

	pkg, err := client.GetPackage(packageName)
	require.NoError(t, err)

	assert.Equal(t, "pkg123", pkg.ID)
	assert.Equal(t, packageName, pkg.Name)
	assert.Equal(t, "Test package", pkg.Description)
	assert.Equal(t, []string{"test"}, pkg.Tags)
}

func TestClientSearchPackages(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/packages/search", r.URL.Path)
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "user", r.URL.Query().Get("q"))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`[
			{
				"id": "pkg1",
				"name": "user-service",
				"description": "User management service",
				"tags": ["grpc"],
				"created_at": "2023-01-01T00:00:00Z"
			}
		]`))
	}))
	defer server.Close()

	client := newTestClient(server.URL, "")

	packages, err := client.SearchPackages("user", []string{})
	require.NoError(t, err)

	assert.Len(t, packages, 1)
	assert.Equal(t, "user-service", packages[0].Name)
}

// Helper function to create test client
func newTestClient(baseURL, token string) *HTTPClient {
	tmpDir, _ := os.MkdirTemp("", "protodex-test-config-")
	configPath := filepath.Join(tmpDir, "config.yaml")

	return &HTTPClient{
		baseURL:    baseURL,
		token:      token,
		httpClient: &http.Client{},
		logger:     logger.Get(),
		config:     &config.Config{ConfigPath: configPath},
	}
}
