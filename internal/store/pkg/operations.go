package pkg

import (
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/google/uuid"
)

func (s *packageStore) CreatePackage(name, description, ownerID string, tags []string) (*Package, error) {
	id := uuid.New().String()

	tagsJSON, err := json.Marshal(tags)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal tags: %w", err)
	}

	query := `INSERT INTO packages (id, name, description, tags, owner_id) VALUES (?, ?, ?, ?, ?)`
	_, err = s.db.Exec(query, id, name, description, string(tagsJSON), ownerID)
	if err != nil {
		return nil, fmt.Errorf("failed to create package: %w", err)
	}

	return &Package{
		ID:          id,
		Name:        name,
		Description: description,
		Tags:        tags,
		OwnerID:     ownerID,
	}, nil
}

func (s *packageStore) GetPackage(name string) (*Package, error) {
	pkg := &Package{}
	var tagsJSON string
	query := `SELECT id, name, description, tags, owner_id, created_at FROM packages WHERE name = ?`
	err := s.db.QueryRow(query, name).Scan(&pkg.ID, &pkg.Name, &pkg.Description, &tagsJSON, &pkg.OwnerID, &pkg.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("package %s not found", name)
		}
		return nil, fmt.Errorf("failed to get package: %w", err)
	}

	// Parse tags from JSON
	if tagsJSON != "" {
		if err := json.Unmarshal([]byte(tagsJSON), &pkg.Tags); err != nil {
			return nil, fmt.Errorf("failed to unmarshal tags: %w", err)
		}
	}

	return pkg, nil
}

func (s *packageStore) GetPackageByID(id string) (*Package, error) {
	pkg := &Package{}
	var tagsJSON string
	query := `SELECT id, name, description, tags, owner_id, created_at FROM packages WHERE id = ?`
	err := s.db.QueryRow(query, id).Scan(&pkg.ID, &pkg.Name, &pkg.Description, &tagsJSON, &pkg.OwnerID, &pkg.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("package with id %s not found", id)
		}
		return nil, fmt.Errorf("failed to get package: %w", err)
	}

	// Parse tags from JSON
	if tagsJSON != "" {
		if err := json.Unmarshal([]byte(tagsJSON), &pkg.Tags); err != nil {
			return nil, fmt.Errorf("failed to unmarshal tags: %w", err)
		}
	}

	return pkg, nil
}

func (s *packageStore) ListPackages() ([]*Package, error) {
	query := `SELECT id, name, description, tags, owner_id, created_at FROM packages ORDER BY name`
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to list packages: %w", err)
	}
	defer rows.Close()

	var packages []*Package
	for rows.Next() {
		pkg := &Package{}
		var tagsJSON string
		if err := rows.Scan(&pkg.ID, &pkg.Name, &pkg.Description, &tagsJSON, &pkg.OwnerID, &pkg.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan package: %w", err)
		}

		// Parse tags from JSON
		if tagsJSON != "" {
			if err := json.Unmarshal([]byte(tagsJSON), &pkg.Tags); err != nil {
				return nil, fmt.Errorf("failed to unmarshal tags: %w", err)
			}
		}

		packages = append(packages, pkg)
	}

	return packages, nil
}

func (s *packageStore) SearchPackages(query string, tags []string) ([]*Package, error) {
	var sqlQuery string
	var args []interface{}

	switch {
	case query != "" && len(tags) > 0:
		sqlQuery = `SELECT id, name, description, tags, created_at FROM packages 
					WHERE (name LIKE ? OR description LIKE ?) AND tags LIKE ? ORDER BY name`
		tagsPattern := "%\"" + tags[0] + "\"%"
		args = []interface{}{"%" + query + "%", "%" + query + "%", tagsPattern}
	case query != "":
		sqlQuery = `SELECT id, name, description, tags, created_at FROM packages 
					WHERE name LIKE ? OR description LIKE ? ORDER BY name`
		args = []interface{}{"%" + query + "%", "%" + query + "%"}
	case len(tags) > 0:
		sqlQuery = `SELECT id, name, description, tags, created_at FROM packages 
					WHERE tags LIKE ? ORDER BY name`
		tagsPattern := "%\"" + tags[0] + "\"%"
		args = []interface{}{tagsPattern}
	default:
		return s.ListPackages()
	}

	rows, err := s.db.Query(sqlQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to search packages: %w", err)
	}
	defer rows.Close()

	var packages []*Package
	for rows.Next() {
		pkg := &Package{}
		var tagsJSON string
		if err := rows.Scan(&pkg.ID, &pkg.Name, &pkg.Description, &tagsJSON, &pkg.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan package: %w", err)
		}

		// Parse tags from JSON
		if tagsJSON != "" {
			if err := json.Unmarshal([]byte(tagsJSON), &pkg.Tags); err != nil {
				return nil, fmt.Errorf("failed to unmarshal tags: %w", err)
			}
		}

		packages = append(packages, pkg)
	}

	return packages, nil
}

func (s *packageStore) StoreSchema(packageID, version, filePath, createdBy string) (*SchemaVersion, error) {
	checksum, err := calculateChecksum(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate checksum: %w", err)
	}

	id := uuid.New().String()

	query := `INSERT INTO schema_versions (id, package_id, version, file_path, checksum, created_by) VALUES (?, ?, ?, ?, ?, ?)`
	_, err = s.db.Exec(query, id, packageID, version, filePath, checksum, createdBy)
	if err != nil {
		return nil, fmt.Errorf("failed to store schema: %w", err)
	}

	return s.GetSchemaVersion(packageID, version)
}

func (s *packageStore) GetSchemaVersion(packageID, version string) (*SchemaVersion, error) {
	return s.GetSchemaVersionWithFiles(packageID, version)
}

func (s *packageStore) ListVersions(packageID string) ([]*SchemaVersion, error) {
	query := `SELECT id, package_id, version, checksum, metadata, created_at, created_by 
			  FROM schema_versions WHERE package_id = ? ORDER BY created_at DESC`
	rows, err := s.db.Query(query, packageID)
	if err != nil {
		return nil, fmt.Errorf("failed to list versions: %w", err)
	}
	defer rows.Close()

	var versions []*SchemaVersion
	for rows.Next() {
		schema := &SchemaVersion{}
		if err := rows.Scan(&schema.ID, &schema.PackageID, &schema.Version,
			&schema.Checksum, &schema.Metadata, &schema.CreatedAt, &schema.CreatedBy); err != nil {
			return nil, fmt.Errorf("failed to scan schema version: %w", err)
		}

		files, err := s.GetSchemaFiles(schema.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to get files for version %s: %w", schema.Version, err)
		}
		schema.Files = files

		versions = append(versions, schema)
	}

	return versions, nil
}

func (s *packageStore) GetSchemaPath(packageName, version string) string {
	return filepath.Join(s.dataDir, "schemas", packageName, version)
}

func (s *packageStore) GetDataDir() string {
	return s.dataDir
}

func calculateChecksum(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}
