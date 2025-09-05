package fetcher

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func (f *Fetcher) downloadAndExtractGithub(githubURL, branch string) error {
	fmt.Printf("Downloading %s from GitHub...\n", githubURL)

	// Stip protocol if present
	githubURL = strings.TrimPrefix(githubURL, "https://")
	githubURL = strings.TrimPrefix(githubURL, "http://")
	githubURL = strings.TrimPrefix(githubURL, "git@")
	githubURL = strings.TrimSuffix(githubURL, ".git")

	// Handle SSH style URL
	if strings.HasPrefix(githubURL, "github.com:") {
		githubURL = strings.Replace(githubURL, "github.com:", "github.com/", 1)
	}

	// Ensure it starts with github.com
	if !strings.HasPrefix(githubURL, "github.com/") {
		return fmt.Errorf("invalid GitHub URL: %s", githubURL)
	}

	// Parse GitHub URL
	// Expected format: github.com/owner/repo or github.com/owner/repo/subdir
	parts := strings.Split(strings.TrimPrefix(githubURL, "github.com/"), "/")
	if len(parts) < 2 {
		return fmt.Errorf("invalid GitHub source format: %s", githubURL)
	}

	owner := parts[0]
	repo := parts[1]
	subDir := ""
	if len(parts) > 2 {
		subDir = strings.Join(parts[2:], "/")
	}

	version := branch
	if version == "" {
		version = "main" // default branch
	}

	// Remove 'v' prefix if present for the download URL
	downloadVersion := strings.TrimPrefix(version, "v")

	// Try release download first, then fallback to archive
	var url string

	if strings.HasPrefix(version, "v") {
		// Try as a release
		url = fmt.Sprintf("https://github.com/%s/%s/archive/refs/tags/%s.zip", owner, repo, version)
	} else {
		// Try as a branch/commit
		url = fmt.Sprintf("https://github.com/%s/%s/archive/refs/heads/%s.zip", owner, repo, downloadVersion)
	}

	fmt.Printf("Downloading from %s...\n", url)

	zipPath, err := f.download(url)
	if err != nil {
		return err
	}

	defer func() {
		if err := os.Remove(zipPath); err != nil {
			fmt.Printf("Warning: failed to remove temp file: %v\n", err)
		}
	}()

	// Extract zip
	if err := extractGitHubZip(zipPath, f.Dest, subDir); err != nil {
		return err
	}
	return nil
}

func extractGitHubZip(zipPath, targetDir, subDir string) error {
	reader, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer func(reader *zip.ReadCloser) {
		err := reader.Close()
		if err != nil {
			fmt.Printf("Warning: failed to close zip reader: %v\n", err)
		}
	}(reader)

	// GitHub archives have a top-level directory like "repo-version/"
	var rootPrefix string
	for _, file := range reader.File {
		if !file.FileInfo().IsDir() {
			parts := strings.Split(file.Name, "/")
			if len(parts) > 0 {
				rootPrefix = parts[0] + "/"
				break
			}
		}
	}

	extracted := 0
	for _, file := range reader.File {
		// Skip directories
		if file.FileInfo().IsDir() {
			continue
		}

		// Remove root prefix
		relativePath := strings.TrimPrefix(file.Name, rootPrefix)

		// If subDir specified, only extract from that subdirectory
		if subDir != "" {
			if !strings.HasPrefix(relativePath, subDir+"/") {
				continue
			}
			// Remove subDir prefix
			relativePath = strings.TrimPrefix(relativePath, subDir+"/")
		}

		// Skip if empty path
		if relativePath == "" {
			continue
		}

		targetPath := filepath.Join(targetDir, relativePath)

		// Create directory if needed
		if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
			return err
		}

		// Extract file
		if err := extractFile(file, targetPath); err != nil {
			return fmt.Errorf("failed to extract %s: %w", relativePath, err)
		}
		extracted++
	}

	fmt.Printf("Extracted %d files from GitHub repository\n", extracted)
	return nil
}

func extractFile(file *zip.File, targetPath string) error {
	rc, err := file.Open()
	if err != nil {
		return err
	}
	defer func(rc io.ReadCloser) {
		err := rc.Close()
		if err != nil {
			fmt.Printf("Warning: failed to close file reader: %v\n", err)
		}
	}(rc)

	outFile, err := os.Create(targetPath)
	if err != nil {
		return err
	}
	defer func(outFile *os.File) {
		err := outFile.Close()
		if err != nil {
			fmt.Printf("Warning: failed to close output file: %v\n", err)
		}
	}(outFile)

	_, err = io.Copy(outFile, rc)
	return err
}
