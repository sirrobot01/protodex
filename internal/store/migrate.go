package store

import (
	"fmt"
)

func (s *dbStore) migrate() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS packages (
			id TEXT PRIMARY KEY,
			name TEXT UNIQUE NOT NULL,
			description TEXT,
			tags TEXT DEFAULT '[]',
			owner_id TEXT REFERENCES users(id),
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS schema_versions (
			id TEXT PRIMARY KEY,
			package_id TEXT REFERENCES packages(id),
			version TEXT NOT NULL,
			checksum TEXT NOT NULL,
			metadata TEXT DEFAULT '{}',
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			created_by TEXT,
			UNIQUE(package_id, version)
		)`,
		`CREATE TABLE IF NOT EXISTS schema_files (
			id TEXT PRIMARY KEY,
			version_id TEXT REFERENCES schema_versions(id) ON DELETE CASCADE,
			filename TEXT NOT NULL,
			file_path TEXT NOT NULL,
			checksum TEXT NOT NULL,
			size_bytes INTEGER NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS users (
			id TEXT PRIMARY KEY,
			username TEXT UNIQUE NOT NULL,
			password_hash TEXT NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS sessions (
			id TEXT PRIMARY KEY,
			user_agent TEXT,
			user_id TEXT REFERENCES users(id) ON DELETE CASCADE,
			token_hash TEXT UNIQUE NOT NULL,
			expires_at TIMESTAMP,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			last_used TIMESTAMP
		)`,
	}

	for _, query := range queries {
		if _, err := s.db.Exec(query); err != nil {
			return fmt.Errorf("failed to execute migration: %w", err)
		}
	}
	return nil
}
