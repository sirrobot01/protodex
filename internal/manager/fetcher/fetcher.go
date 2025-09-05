package fetcher

import (
	"archive/zip"
	"cmp"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"github.com/sirrobot01/protodex/internal/client"
)

const (
	FetchedMarkerFile = ".protodex_fetched"
)

type Fetcher struct {
	SourceType SourceType
	Source     string
	Version    string
	Dest       string
	client     *http.Client
}

func NewFetcherFromURL(sourceURL, dest string) (*Fetcher, error) {
	sourceInfo, err := ParseSource(sourceURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse source URL: %w", err)
	}
	return NewFetcher(sourceInfo.Type, sourceInfo.Source, sourceInfo.Version, dest)
}

func NewFetcher(sourceType SourceType, source, version, dest string) (*Fetcher, error) {
	if sourceType == SourceLocal {
		dest = cmp.Or(dest, source)
	}
	return &Fetcher{
		SourceType: sourceType,
		Source:     source,
		Dest:       dest,
		Version:    version,
		client:     &http.Client{Timeout: 30 * time.Minute},
	}, nil
}

func (f *Fetcher) Fetch() error {
	switch f.SourceType {
	case SourceLocal:
		return f.localFetch()
	case SourceGitHub:
		return f.githubFetch()
	case SourceProtodex:
		return f.protodexFetch()
	case SourceHTTP:
		return f.urlFetch()
	case SourceGoogleWellKnown:
		return f.urlFetch()
	default:
		return fmt.Errorf("unsupported source type: %s", f.SourceType)
	}
}

func (f *Fetcher) localFetch() error {
	absSourcePath, err := filepath.Abs(f.Source)
	if err != nil {
		return fmt.Errorf("failed to get absolute path of source: %w", err)
	}
	// Check if source path exists
	if _, err := os.Stat(absSourcePath); os.IsNotExist(err) {
		return fmt.Errorf("source path does not exist: %s", absSourcePath)
	}

	destAbsolute, err := filepath.Abs(f.Dest)
	if err != nil {
		return fmt.Errorf("failed to get absolute path of destination: %w", err)
	}

	// Check if source and dest are the same
	if absSourcePath == destAbsolute {
		return nil
	}

	// Try to create a symlink
	if err := os.Symlink(absSourcePath, f.Dest); err != nil {
		return fmt.Errorf("failed to create symlink from %s to %s: %w", absSourcePath, f.Dest, err)
	}
	return nil
}

func (f *Fetcher) protodexFetch() error {
	// Use client to pull the package
	c, err := client.New()
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	return c.PullVersion(f.Source, f.Version, f.Dest)
}

func (f *Fetcher) urlFetch() error {
	parsedURL, err := url.Parse(f.Source)
	if err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}
	// Only support http and https
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return fmt.Errorf("unsupported URL scheme: %s", parsedURL.Scheme)
	}
	// Let's check if it has been fetched already
	if isFetched(f.Dest) {
		fmt.Printf("Source already fetched to %s\n", f.Dest)
		return nil
	}
	fmt.Printf("Downloading from %s...\n", f.Source)

	if err := f.downloadAndExtract(); err != nil {
		return fmt.Errorf("failed to fetch from URL: %w", err)
	}
	// Mark as fetched
	if err := f.markAsFetched(); err != nil {
		// Marking as fetched failed, but we already fetched the content
		return nil
	}
	return nil
}

func (f *Fetcher) githubFetch() error {
	ref := f.Version
	if ref == "" {
		ref = "main" // default branch
	}
	gitURL := f.Source
	if gitURL == "" {
		return fmt.Errorf("github source URL is required")
	}

	if isFetched(f.Dest) {
		fmt.Printf("Source already fetched to %s\n", f.Dest)
	}

	// Now let's add github.com/ prefix if missing
	gitURL = fmt.Sprintf("%s%s", "github.com/", gitURL)
	// Download and extract

	if err := f.downloadAndExtractGithub(gitURL, ref); err != nil {
		return fmt.Errorf("failed to fetch from GitHub: %w", err)
	}
	// Mark as fetched
	if err := f.markAsFetched(); err != nil {
		return fmt.Errorf("failed to mark as fetched: %w", err)
	}
	return nil
}

func isFetched(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}

	// Check if directory has content
	entries, err := os.ReadDir(path)
	if err != nil {
		return false
	}

	if len(entries) == 0 {
		return false
	}

	// Check for fetched marker file
	markerPath := filepath.Join(path, FetchedMarkerFile)
	if _, err := os.Stat(markerPath); os.IsNotExist(err) {
		return false
	}
	return true
}

func (f *Fetcher) download(url string) (string, error) {
	resp, err := f.client.Get(url)
	if err != nil {
		return "", err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Printf("Warning: failed to close response body: %v\n", err)
		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	// Save to temp file
	tempFile, err := os.CreateTemp("", "protodex-download-*.zip")
	if err != nil {
		return "", err
	}

	if _, err := io.Copy(tempFile, resp.Body); err != nil {
		return "", err
	}
	if err := tempFile.Close(); err != nil {
		return "", err
	}
	return tempFile.Name(), nil
}

func (f *Fetcher) markAsFetched() error {
	markerPath := filepath.Join(f.Dest, FetchedMarkerFile)
	markerFile, err := os.Create(markerPath)
	if err != nil {
		return fmt.Errorf("failed to create fetched marker: %w", err)
	}
	defer func() {
		if err := markerFile.Close(); err != nil {
			fmt.Printf("Warning: failed to close marker file: %v\n", err)
		}
	}()
	return nil
}

func extractZip(zipPath, destPath string) error {
	reader, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer func(reader *zip.ReadCloser) {
		if err := reader.Close(); err != nil {
			fmt.Printf("Warning: failed to close zip reader: %v\n", err)
		}
	}(reader)

	for _, file := range reader.File {
		if file.FileInfo().IsDir() {
			continue
		}

		rc, err := file.Open()
		if err != nil {
			return err
		}

		destFile := filepath.Join(destPath, file.Name)
		if err := os.MkdirAll(filepath.Dir(destFile), 0755); err != nil {
			if err := rc.Close(); err != nil {
				return err
			}
			return err
		}

		outFile, err := os.Create(destFile)
		if err != nil {
			if err := rc.Close(); err != nil {
				return err
			}
			return err
		}

		_, err = io.Copy(outFile, rc)

		if err := outFile.Close(); err != nil {
			return err
		}

		if err := rc.Close(); err != nil {
			return err
		}

		if err != nil {
			return err
		}
	}

	return nil
}
