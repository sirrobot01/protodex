package store

import (
	"database/sql"
	"fmt"
	"path/filepath"

	"github.com/rs/zerolog"
	"github.com/sirrobot01/protodex/internal/logger"
	"github.com/sirrobot01/protodex/internal/store/auth"
	"github.com/sirrobot01/protodex/internal/store/pkg"
	_ "modernc.org/sqlite"
)

type Store interface {
	Init() error
	Auth() auth.Store
	Package() pkg.Store

	Close() error
}

type dbStore struct {
	auth auth.Store
	pkg  pkg.Store

	db     *sql.DB
	logger zerolog.Logger
}

func New(dataDir string) (Store, error) {
	dbPath := filepath.Join(dataDir, "protodex.db")

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}
	// Create sub-stores
	authStore := auth.NewStore(db)
	pkgStore := pkg.NewStore(db, dataDir)

	return &dbStore{
		db:     db,
		logger: logger.Get(),
		auth:   authStore,
		pkg:    pkgStore,
	}, nil
}

func (s *dbStore) Auth() auth.Store {
	return s.auth
}

func (s *dbStore) Package() pkg.Store {
	return s.pkg
}

func (s *dbStore) Close() error {
	return s.db.Close()
}

func (s *dbStore) Init() error {
	if err := s.db.Ping(); err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	// Set journal mode to WAL for better concurrency
	_, err := s.db.Exec(`PRAGMA journal_mode = WAL;`)
	if err != nil {
		return fmt.Errorf("failed to set journal mode: %w", err)
	}

	// Migrate db
	if err := s.migrate(); err != nil {
		return fmt.Errorf("failed to migrate database: %w", err)
	}
	return nil
}
