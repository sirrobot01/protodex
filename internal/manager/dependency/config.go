package dependency

import (
	"github.com/sirrobot01/protodex/internal/manager/fetcher"
)

type Config struct {
	Name    string             `yaml:"name"`
	Type    fetcher.SourceType `yaml:"type"`              // e.g., "git", "http", "google-well-known", "local", "protodex"
	Source  string             `yaml:"source"`            // e.g., repo URL, base URL, or local path
	Version string             `yaml:"version,omitempty"` // This can be a git tag, branch, protodex version, or URL version
	Path    string             `yaml:"path,omitempty"`
}
