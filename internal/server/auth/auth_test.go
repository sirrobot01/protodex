package auth

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	"github.com/sirrobot01/protodex/internal/store/auth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

func TestCreateUser(t *testing.T) {
	service := setupTestAuthService(t)

	user, err := service.CreateUser("testuser", "password123")
	require.NoError(t, err)

	assert.NotEmpty(t, user.ID)
	assert.Equal(t, "testuser", user.Username)
	assert.NotEmpty(t, user.PasswordHash)
	assert.NotZero(t, user.CreatedAt)
}

func TestLogin(t *testing.T) {
	service := setupTestAuthService(t)

	// Create a user
	_, err := service.CreateUser("testuser", "password123")
	require.NoError(t, err)

	// Test successful login
	loginResp, err := service.Login("test-user-agent", "testuser", "password123")
	require.NoError(t, err)

	assert.NotEmpty(t, loginResp.Token)
	assert.Equal(t, "testuser", loginResp.User.Username)

	// Test invalid password
	_, err = service.Login("test-user-agent", "testuser", "wrongpassword")
	assert.Error(t, err)
}

func TestValidateToken(t *testing.T) {
	service := setupTestAuthService(t)

	// Create a user and login
	_, err := service.CreateUser("testuser", "password123")
	require.NoError(t, err)

	loginResp, err := service.Login("test-user-agent", "testuser", "password123")
	require.NoError(t, err)

	// Test valid token
	authCtx, err := service.ValidateToken(loginResp.Token)
	require.NoError(t, err)

	assert.Equal(t, loginResp.User.ID, authCtx.UserID)
	assert.Equal(t, "testuser", authCtx.User.Username)

	// Test invalid token
	_, err = service.ValidateToken("invalid-token")
	assert.Error(t, err)
}

func TestTokenExtraction(t *testing.T) {
	tests := []struct {
		name     string
		header   string
		expected string
	}{
		{
			name:     "Bearer token",
			header:   "Bearer abc123",
			expected: "abc123",
		},
		{
			name:     "No Bearer prefix",
			header:   "abc123",
			expected: "",
		},
		{
			name:     "Empty header",
			header:   "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExtractTokenFromHeader(tt.header)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Helper function to set up test auth service
func setupTestAuthService(t *testing.T) *Service {
	tmpDir, err := os.MkdirTemp("", "protodex-auth-test-")
	require.NoError(t, err)

	dbPath := filepath.Join(tmpDir, "test.db")
	db, err := sql.Open("sqlite", dbPath)
	require.NoError(t, err)

	// Create auth tables
	_, err = db.Exec(`
		CREATE TABLE users (
			id TEXT PRIMARY KEY,
			username TEXT UNIQUE NOT NULL,
			password_hash TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`)
	require.NoError(t, err)

	_, err = db.Exec(`
		CREATE TABLE sessions (
			id TEXT PRIMARY KEY,
			user_id TEXT NOT NULL,
			token_hash TEXT UNIQUE NOT NULL,
			user_agent TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			expires_at DATETIME NOT NULL,
			is_active BOOLEAN DEFAULT 1,
			last_used TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users (id)
		)
	`)
	require.NoError(t, err)

	authStore := auth.NewStore(db)
	service := NewAuthService(authStore)

	t.Cleanup(func() {
		db.Close()
		os.RemoveAll(tmpDir)
	})

	return service
}
