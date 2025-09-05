package protoc

import (
	"archive/zip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

func (e *Executor) downloadProtoc() error {
	platform := getCurrentPlatform()
	if platform == "" {
		return fmt.Errorf("unsupported platform: %s/%s", runtime.GOOS, runtime.GOARCH)
	}

	binaryName := getBinaryName(platform)
	if binaryName == "" {
		return fmt.Errorf("failed to get binary name for platform: %s", platform)
	}

	// Create directory if it doesn't exist
	dir := filepath.Dir(e.protocPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Download and extract
	if err := downloadAndExtract(platform, e.protocPath, e.version); err != nil {
		return fmt.Errorf("failed to download protoc for %s: %w", platform, err)
	}

	// Make executable
	if err := os.Chmod(e.protocPath, 0755); err != nil {
		return fmt.Errorf("failed to make protoc executable: %w", err)
	}

	return nil
}

func getCurrentPlatform() string {
	arch := runtime.GOARCH

	// Normalize architecture
	switch arch {
	case "amd64":
		arch = "x86_64"
	case "arm64":
		arch = "aarch_64"
	default:
		return ""
	}

	// Map to protobuf platform names
	switch runtime.GOOS {
	case "linux":
		return fmt.Sprintf("linux-%s", arch)
	case "darwin":
		return fmt.Sprintf("osx-%s", arch)
	case "windows":
		if arch == "x86_64" {
			return "win64"
		}
		return ""
	default:
		return ""
	}
}

func getBinaryName(platform string) string {
	switch platform {
	case "linux-x86_64":
		return "protoc-linux-amd64"
	case "linux-aarch_64":
		return "protoc-linux-arm64"
	case "osx-x86_64":
		return "protoc-darwin-amd64"
	case "osx-aarch_64":
		return "protoc-darwin-arm64"
	case "win64":
		return "protoc-windows-amd64.exe"
	default:
		return ""
	}
}

func downloadAndExtract(platform, outputPath, version string) error {
	filename := fmt.Sprintf("protoc-%s-%s.zip", version, platform)
	url := fmt.Sprintf("https://github.com/protocolbuffers/protobuf/releases/download/v%s/%s",
		version, filename)

	// Download
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to download: %w", err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Printf("Warning: failed to close response body: %v\n", err)
		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed with status: %d", resp.StatusCode)
	}

	// Create temporary file
	tempFile, err := os.CreateTemp("", "protoc-*.zip")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	defer func(name string) {
		err := os.Remove(name)
		if err != nil {
			fmt.Printf("Warning: failed to remove temp file %s: %v\n", name, err)
		}
	}(tempFile.Name())
	defer func(tempFile *os.File) {
		err := tempFile.Close()
		if err != nil {
			fmt.Printf("Warning: failed to close temp file: %v\n", err)
		}
	}(tempFile)

	// Copy response to temp file
	_, err = io.Copy(tempFile, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to write download: %w", err)
	}

	// Extract protoc binary from zip
	if err := extractProtocFromZip(tempFile.Name(), outputPath); err != nil {
		return fmt.Errorf("failed to extract: %w", err)
	}

	return nil
}

func extractProtocFromZip(zipPath, outputPath string) error {
	// Open zip file
	reader, err := zip.OpenReader(zipPath)
	if err != nil {
		return fmt.Errorf("failed to open zip: %w", err)
	}
	defer func(reader *zip.ReadCloser) {
		err := reader.Close()
		if err != nil {
			fmt.Printf("Warning: failed to close zip reader: %v\n", err)
		}
	}(reader)

	// Find protoc binary in zip
	var protocFile *zip.File
	for _, file := range reader.File {
		if strings.HasSuffix(file.Name, "bin/protoc") || strings.HasSuffix(file.Name, "bin/protoc.exe") {
			protocFile = file
			break
		}
	}

	if protocFile == nil {
		return fmt.Errorf("protoc binary not found in zip")
	}

	// Extract protoc binary
	src, err := protocFile.Open()
	if err != nil {
		return fmt.Errorf("failed to open protoc in zip: %w", err)
	}
	defer func(src io.ReadCloser) {
		err := src.Close()
		if err != nil {
			fmt.Printf("Warning: failed to close protoc file: %v\n", err)
		}
	}(src)

	// Create output file
	dst, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer func(dst *os.File) {
		err := dst.Close()
		if err != nil {
			fmt.Printf("Warning: failed to close output file: %v\n", err)
		}
	}(dst)

	// Copy binary content
	_, err = io.Copy(dst, src)
	if err != nil {
		return fmt.Errorf("failed to copy protoc binary: %w", err)
	}

	return nil
}
