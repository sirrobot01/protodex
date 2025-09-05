package pkg

import (
	"database/sql"
	"github.com/rs/zerolog"

	"github.com/sirrobot01/protodex/internal/logger"
)

type Store interface {
	CreatePackage(name, description, ownerID string, tags []string) (*Package, error)
	GetPackage(name string) (*Package, error)
	GetPackageByID(id string) (*Package, error)
	ListPackages() ([]*Package, error)
	SearchPackages(query string, tags []string) ([]*Package, error)

	StoreSchema(packageID, version, filePath, createdBy string) (*SchemaVersion, error)
	GetSchemaVersion(packageID, version string) (*SchemaVersion, error)
	ListVersions(packageID string) ([]*SchemaVersion, error)
	GetSchemaPath(packageName, version string) string

	SaveSchemaFiles(packageID, version string, filePaths []string, createdBy string) (*SchemaVersion, error)
	GetSchemaVersionWithFiles(packageID, version string) (*SchemaVersion, error)
	GetSchemaFiles(versionID string) ([]SchemaFile, error)
	DeleteSchemaVersion(packageID, version string) error

	GetDataDir() string
}

type packageStore struct {
	dataDir string
	db      *sql.DB
	logger  zerolog.Logger
}

func NewStore(db *sql.DB, dataDir string) Store {
	return &packageStore{
		db:      db,
		logger:  logger.Get(),
		dataDir: dataDir,
	}
}
