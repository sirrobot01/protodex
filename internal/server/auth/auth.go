package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	authstore "github.com/sirrobot01/protodex/internal/store/auth"
)

type Context struct {
	Session *authstore.Session `json:"session"`
	UserID  string             `json:"user_id"`
	User    *authstore.User    `json:"user,omitempty"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	OrgID    string `json:"org_id"`
}

type LoginResponse struct {
	User    *authstore.User `json:"user"`
	Token   string          `json:"token"`
	Expires *time.Time      `json:"expires_at,omitempty"`
}

type Service struct {
	store authstore.Store
}

func NewAuthService(store authstore.Store) *Service {
	return &Service{
		store: store,
	}
}

func (s *Service) Login(client, username, password string) (*LoginResponse, error) {
	user, err := s.store.Authenticate(username, password)
	if err != nil {
		return nil, err
	}

	// Create a session token
	token, err := generateToken()
	if err != nil {
		return nil, fmt.Errorf("failed to generate session token: %w", err)
	}

	tokenHash := hashToken(token)
	session, err := s.store.CreateSession(client, user.ID, tokenHash)
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	return &LoginResponse{
		User:    user,
		Token:   token,
		Expires: session.ExpiresAt,
	}, nil
}

func (s *Service) ValidateToken(token string) (*Context, error) {
	if token == "" {
		return nil, fmt.Errorf("token is required")
	}

	tokenHash := hashToken(token)

	session, err := s.store.GetSessionByHash(tokenHash)
	if err != nil {
		return nil, fmt.Errorf("invalid session")
	}

	if s.store.IsSessionExpired(session) {
		return nil, fmt.Errorf("session expired")
	}

	if err := s.store.UpdateSessionLastUsed(session.ID); err != nil {
		return nil, fmt.Errorf("failed to update session: %w", err)
	}

	// Get user info
	user, err := s.store.GetUserByID(session.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &Context{
		Session: session,
		UserID:  session.UserID,
		User:    user,
	}, nil
}

func (s *Service) GetUserByID(userID string) (*authstore.User, error) {
	return s.store.GetUserByID(userID)
}

func (s *Service) GetUserByUsername(username string) (*authstore.User, error) {
	return s.store.GetUserByUsername(username)
}

func (s *Service) CreateUser(username, password string) (*authstore.User, error) {
	return s.store.CreateUser(username, password)
}

func generateToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	return "pg_" + hex.EncodeToString(bytes), nil
}

func hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

func ExtractTokenFromHeader(authHeader string) string {
	if authHeader == "" {
		return ""
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return ""
	}

	return parts[1]
}
