package dependency

import (
	"archive/zip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

const googleProtocVersion = "v32.0"

func (dc *Resolver) resolveGoogleWellKnownProtos(dep Config) error {
	targetDir := filepath.Join(dc.outputPath, dep.Name)

	if dc.isCached(targetDir) {
		return nil
	}

	fmt.Printf("Downloading %s from GitHub...\n", dep.Name)

	version := dep.Version
	if version == "" {
		version = googleProtocVersion
	}
	versionWithoutPrefix := strings.TrimPrefix(version, "v")

	url := fmt.Sprintf(
		"https://github.com/protocolbuffers/protobuf/releases/download/%s/protoc-%s-linux-x86_64.zip",
		version, versionWithoutPrefix,
	)

	fmt.Printf("Downloading from %s...\n", url)
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to download includes: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download includes: HTTP %d", resp.StatusCode)
	}

	// Save to temp file
	tempFile, err := os.CreateTemp("", "protoc-includes-*.zip")
	if err != nil {
		return err
	}
	defer os.Remove(tempFile.Name())

	if _, err := io.Copy(tempFile, resp.Body); err != nil {
		return err
	}
	tempFile.Close()

	if err := extractIncludes(tempFile.Name(), targetDir); err != nil {
		return err
	}

	// Create marker file to indicate completion
	markerFile := filepath.Join(targetDir, ".protodex-complete")
	if err := os.WriteFile(markerFile, []byte(googleProtocVersion), 0644); err != nil {
		return err
	}

	fmt.Printf("Protobuf includes cached to %s\n", targetDir)
	return nil
}

func extractIncludes(zipPath, targetDir string) error {
	reader, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer reader.Close()

	extracted := 0
	for _, file := range reader.File {
		// Only extract include files
		if !strings.HasPrefix(file.Name, "include/") {
			continue
		}

		// Skip directories
		if file.FileInfo().IsDir() {
			continue
		}

		// Remove "include/" prefix to get relative path
		relativePath := strings.TrimPrefix(file.Name, "include/")
		relativePath = strings.TrimPrefix(relativePath, "google/protobuf/")
		targetPath := filepath.Join(targetDir, relativePath)

		// Create directory if needed
		if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
			return err
		}
		if err := extractFile(file, targetPath); err != nil {
			return err
		}

		extracted++
	}

	fmt.Printf("Extracted %d include files\n", extracted)
	return nil
}

func extractFile(file *zip.File, targetPath string) error {
	rc, err := file.Open()
	if err != nil {
		return err
	}
	defer rc.Close()

	outFile, err := os.Create(targetPath)
	if err != nil {
		return err
	}
	defer outFile.Close()

	_, err = io.Copy(outFile, rc)
	return err
}
