package auth

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type Session struct {
	UserAgent string     `json:"user_agent" db:"user_agent"`
	ID        string     `json:"id" db:"id"`
	UserID    string     `json:"user_id" db:"user_id"`
	TokenHash string     `json:"-" db:"token_hash"`
	ExpiresAt *time.Time `json:"expires_at" db:"expires_at"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	LastUsed  *time.Time `json:"last_used" db:"last_used"`
}

type User struct {
	ID           string `json:"id"`
	Username     string `json:"username"`
	PasswordHash string `json:"-"`
	CreatedAt    string `json:"created_at"`
}

type Store interface {
	Authenticate(username, password string) (*User, error)
	CreateUser(username, password string) (*User, error)
	GetUserByUsername(username string) (*User, error)
	GetUserByID(userID string) (*User, error)

	CreateSession(userAgent, userID, tokenHash string) (*Session, error)
	GetSession(id string) (*Session, error)
	GetSessionByHash(tokenHash string) (*Session, error)
	UpdateSessionLastUsed(sessionID string) error
	DeleteSession(sessionID string) error
	IsSessionExpired(session *Session) bool
}

type dbStore struct {
	db *sql.DB
}

func NewStore(db *sql.DB) Store {
	return &dbStore{
		db: db,
	}
}

func (ds *dbStore) Authenticate(username, password string) (*User, error) {
	user, err := ds.GetUserByUsername(username)
	if err != nil {
		return nil, fmt.Errorf("authentication failed: %w", err)
	}

	if err := CheckPasswordHash(password, user.PasswordHash); err != nil {
		return nil, fmt.Errorf("authentication failed: %w", err)
	}

	return user, nil
}

func (ds *dbStore) CreateUser(username, password string) (*User, error) {
	id := uuid.New().String()

	passwordHash, err := HashPassword(password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	query := `INSERT INTO users (id, username, password_hash) VALUES (?, ?, ?)`
	_, err = ds.db.Exec(query, id, username, passwordHash)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return &User{
		ID:           id,
		Username:     username,
		PasswordHash: passwordHash,
		CreatedAt:    time.Now().Format(time.RFC3339),
	}, nil
}

func (ds *dbStore) GetUserByUsername(username string) (*User, error) {
	user := &User{}

	query := `SELECT id, username, password_hash, created_at FROM users WHERE username = ?`
	err := ds.db.QueryRow(query, username).Scan(&user.ID, &user.Username, &user.PasswordHash, &user.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}

func (ds *dbStore) GetUserByID(userID string) (*User, error) {
	user := &User{}

	query := `SELECT id, username, password_hash, created_at FROM users WHERE id = ?`
	err := ds.db.QueryRow(query, userID).Scan(&user.ID, &user.Username, &user.PasswordHash, &user.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}

func (ds *dbStore) CreateSession(userAgent, userID, tokenHash string) (*Session, error) {
	id := uuid.New().String()

	// Sessions expire in 30 days
	expiresAt := time.Now().Add(30 * 24 * time.Hour)

	query := `INSERT INTO sessions (id, user_id, token_hash, user_agent, expires_at) VALUES (?, ?, ?, ?, ?)`
	_, err := ds.db.Exec(query, id, userID, tokenHash, userAgent, expiresAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	return ds.GetSession(id)
}

func (ds *dbStore) GetSession(id string) (*Session, error) {
	session := &Session{}

	query := `SELECT id, user_id, token_hash, user_agent, expires_at, created_at, last_used FROM sessions WHERE id = ?`
	err := ds.db.QueryRow(query, id).Scan(&session.ID, &session.UserID, &session.TokenHash, &session.UserAgent, &session.ExpiresAt, &session.CreatedAt, &session.LastUsed)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("session not found")
		}
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	return session, nil
}

func (ds *dbStore) GetSessionByHash(tokenHash string) (*Session, error) {
	session := &Session{}

	query := `SELECT id, user_id, token_hash, user_agent, expires_at, created_at, last_used FROM sessions WHERE token_hash = ?`
	err := ds.db.QueryRow(query, tokenHash).Scan(&session.ID, &session.UserID, &session.TokenHash, &session.UserAgent, &session.ExpiresAt, &session.CreatedAt, &session.LastUsed)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("session not found")
		}
		return nil, fmt.Errorf("failed to get session by hash: %w", err)
	}

	return session, nil
}

func (ds *dbStore) UpdateSessionLastUsed(sessionID string) error {
	query := `UPDATE sessions SET last_used = CURRENT_TIMESTAMP WHERE id = ?`
	_, err := ds.db.Exec(query, sessionID)
	if err != nil {
		return fmt.Errorf("failed to update session last used: %w", err)
	}
	return nil
}

func (ds *dbStore) DeleteSession(sessionID string) error {
	query := `DELETE FROM sessions WHERE id = ?`
	result, err := ds.db.Exec(query, sessionID)
	if err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("session not found")
	}

	return nil
}

func (ds *dbStore) IsSessionExpired(session *Session) bool {
	if session.ExpiresAt == nil {
		return false
	}
	return session.ExpiresAt.Before(time.Now())
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func CheckPasswordHash(password, hash string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}
