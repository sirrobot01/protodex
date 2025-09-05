package pkg

import "time"

type Package struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Tags        []string  `json:"tags,omitempty"`
	OwnerID     string    `json:"owner_id"`
	CreatedAt   time.Time `json:"created_at"`
}

type SchemaVersion struct {
	ID        string       `json:"id"`
	PackageID string       `json:"package_id"`
	Version   string       `json:"version"`
	Checksum  string       `json:"checksum"`
	Metadata  string       `json:"metadata"`
	CreatedAt time.Time    `json:"created_at"`
	CreatedBy string       `json:"created_by"`
	Files     []SchemaFile `json:"files,omitempty"`
}

type SchemaFile struct {
	ID        string    `json:"id"`
	VersionID string    `json:"version_id"`
	Filename  string    `json:"filename"`
	FilePath  string    `json:"file_path"`
	Checksum  string    `json:"checksum"`
	SizeBytes int64     `json:"size_bytes"`
	CreatedAt time.Time `json:"created_at"`
}
