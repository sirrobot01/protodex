package dependency

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/sirrobot01/protodex/internal/config"
	"github.com/sirrobot01/protodex/internal/manager/fetcher"
)

type Resolver struct {
	outputPath string
	client     *http.Client
}

func NewResolver() (*Resolver, error) {
	outputPath := filepath.Join(config.GetProtodexPath(), "deps")
	if err := os.MkdirAll(outputPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create cache directory: %w", err)
	}

	return &Resolver{
		outputPath: outputPath,
		client:     &http.Client{Timeout: 5 * time.Minute},
	}, nil
}

func (dc *Resolver) ResolveDependencies(deps []Config) error {
	for _, dep := range deps {
		err := dc.resolveDependency(dep)
		if err != nil {
			return fmt.Errorf("failed to resolve dependency %s: %w", dep.Name, err)
		}
	}

	return nil
}

func (dc *Resolver) resolveDependency(dep Config) error {
	if dep.Type == fetcher.SourceGoogleWellKnown {
		return dc.resolveGoogleWellKnownProtos(dep)
	}
	return dc._resolveDependency(dep)
}

func (dc *Resolver) _resolveDependency(dep Config) error {
	targetDir := filepath.Join(dc.outputPath, dep.Name)

	// Create a fetcher based on source type
	fch, err := fetcher.NewFetcher(dep.Type, dep.Source, dep.Version, targetDir)
	if err != nil {
		return fmt.Errorf("failed to create fetcher: %w", err)
	}
	return fch.Fetch()
}

func (dc *Resolver) isCached(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}

	// Check if directory has content
	entries, err := os.ReadDir(path)
	if err != nil {
		return false
	}

	return len(entries) > 0
}

func (dc *Resolver) GetDependencyPath() string {
	return dc.outputPath
}

// Clear cache
func (dc *Resolver) Clear() error {
	return os.RemoveAll(dc.outputPath)
}

func (dc *Resolver) ListCached() ([]string, error) {
	entries, err := os.ReadDir(dc.outputPath)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, err
	}

	var cached []string
	for _, entry := range entries {
		if entry.IsDir() {
			cached = append(cached, entry.Name())
		}
	}

	return cached, nil
}
