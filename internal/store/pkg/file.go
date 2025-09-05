package pkg

import (
	"crypto/sha256"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	"github.com/google/uuid"
)

func (s *packageStore) SaveSchemaFiles(packageID, version string, filePaths []string, createdBy string) (*SchemaVersion, error) {
	if len(filePaths) == 0 {
		return nil, fmt.Errorf("at least one file is required")
	}

	versionID := uuid.New().String()

	var allFiles []SchemaFile
	var combinedChecksum string

	for _, filePath := range filePaths {
		fileInfo, err := os.Stat(filePath)
		if err != nil {
			return nil, fmt.Errorf("failed to stat file %s: %w", filePath, err)
		}

		checksum, err := calculateChecksum(filePath)
		if err != nil {
			return nil, fmt.Errorf("failed to calculate checksum for %s: %w", filePath, err)
		}

		// Store relative path from data directory for consistent storage
		relPath, err := filepath.Rel(s.dataDir, filePath)
		if err != nil {
			// If we can't get relative path, use the full path
			relPath = filePath
		}

		schemaFile := SchemaFile{
			ID:        uuid.New().String(),
			VersionID: versionID,
			Filename:  filepath.Base(filePath),
			FilePath:  relPath,
			Checksum:  checksum,
			SizeBytes: fileInfo.Size(),
		}

		allFiles = append(allFiles, schemaFile)
		combinedChecksum += checksum
	}

	combinedHash := fmt.Sprintf("%x", sha256.Sum256([]byte(combinedChecksum)))

	tx, err := s.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	committed := false
	defer func() {
		if !committed {
			if err := tx.Rollback(); err != nil {
				s.logger.Error().Err(err).Msg("failed to rollback transaction")
			}
		}
	}()

	query := `INSERT INTO schema_versions (id, package_id, version, checksum, created_by) VALUES (?, ?, ?, ?, ?)`
	_, err = tx.Exec(query, versionID, packageID, version, combinedHash, createdBy)
	if err != nil {
		return nil, fmt.Errorf("failed to create schema version: %w", err)
	}

	for _, file := range allFiles {
		query := `INSERT INTO schema_files (id, version_id, filename, file_path, checksum, size_bytes) 
				  VALUES (?, ?, ?, ?, ?, ?)`
		_, err = tx.Exec(query, file.ID, file.VersionID, file.Filename, file.FilePath, file.Checksum, file.SizeBytes)
		if err != nil {
			return nil, fmt.Errorf("failed to store file %s: %w", file.Filename, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}
	committed = true

	return s.GetSchemaVersionWithFiles(packageID, version)
}

func (s *packageStore) GetSchemaVersionWithFiles(packageID, version string) (*SchemaVersion, error) {
	schema := &SchemaVersion{}
	query := `SELECT id, package_id, version, checksum, metadata, created_at, created_by 
			  FROM schema_versions WHERE package_id = ? AND version = ?`
	err := s.db.QueryRow(query, packageID, version).Scan(
		&schema.ID, &schema.PackageID, &schema.Version, &schema.Checksum,
		&schema.Metadata, &schema.CreatedAt, &schema.CreatedBy)
	if err != nil {
		return nil, fmt.Errorf("schema version not found: %w", err)
	}

	files, err := s.GetSchemaFiles(schema.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get schema files: %w", err)
	}

	schema.Files = files
	return schema, nil
}

func (s *packageStore) GetSchemaFiles(versionID string) ([]SchemaFile, error) {
	query := `SELECT id, version_id, filename, file_path, checksum, size_bytes, created_at 
			  FROM schema_files WHERE version_id = ? ORDER BY filename`
	rows, err := s.db.Query(query, versionID)
	if err != nil {
		return nil, fmt.Errorf("failed to query schema files: %w", err)
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			s.logger.Error().Err(err).Msg("failed to close rows")
		}
	}(rows)

	var files []SchemaFile
	for rows.Next() {
		var file SchemaFile
		err := rows.Scan(&file.ID, &file.VersionID, &file.Filename, &file.FilePath,
			&file.Checksum, &file.SizeBytes, &file.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan schema file: %w", err)
		}
		files = append(files, file)
	}

	return files, nil
}

func (s *packageStore) DeleteSchemaVersion(packageID, version string) error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	committed := false
	defer func() {
		if !committed {
			if err := tx.Rollback(); err != nil {
				s.logger.Error().Err(err).Msg("failed to rollback transaction")
			}
		}
	}()

	var versionID string
	query := `SELECT id FROM schema_versions WHERE package_id = ? AND version = ?`
	err = tx.QueryRow(query, packageID, version).Scan(&versionID)
	if err != nil {
		return fmt.Errorf("schema version not found: %w", err)
	}

	var filePaths []string
	query = `SELECT file_path FROM schema_files WHERE version_id = ?`
	rows, err := tx.Query(query, versionID)
	if err != nil {
		return fmt.Errorf("failed to get file paths: %w", err)
	}

	for rows.Next() {
		var filePath string
		if err := rows.Scan(&filePath); err != nil {
			rows.Close()
			return fmt.Errorf("failed to scan file path: %w", err)
		}
		filePaths = append(filePaths, filePath)
	}

	if err := rows.Close(); err != nil {
		return err
	}

	query = `DELETE FROM schema_files WHERE version_id = ?`
	if _, err := tx.Exec(query, versionID); err != nil {
		return fmt.Errorf("failed to delete schema files: %w", err)
	}

	query = `DELETE FROM schema_versions WHERE id = ?`
	if _, err := tx.Exec(query, versionID); err != nil {
		return fmt.Errorf("failed to delete schema version: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	committed = true

	for _, filePath := range filePaths {
		if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("failed to delete file %s: %w", filePath, err)
		}
	}

	return nil
}
