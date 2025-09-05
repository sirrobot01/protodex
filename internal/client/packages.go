package client

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func (c *HTTPClient) ListPackages() ([]*Package, error) {
	url := fmt.Sprintf("%s/api/packages", c.baseURL)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			c.logger.Error().Err(err).Msg("failed to close response body")
		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to list packages: %s", resp.Status)
	}

	var packages []*Package
	if err := json.NewDecoder(resp.Body).Decode(&packages); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return packages, nil
}

func (c *HTTPClient) GetPackage(name string) (*Package, error) {
	url := fmt.Sprintf("%s/api/packages/%s", c.baseURL, name)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("package not found: %s", resp.Status)
	}

	var pkg Package
	if err := json.NewDecoder(resp.Body).Decode(&pkg); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &pkg, nil
}

func (c *HTTPClient) CreatePackage(name, description string, tags []string) (*Package, error) {
	createReq := map[string]interface{}{
		"name":        name,
		"description": description,
		"tags":        tags,
	}

	jsonData, err := json.Marshal(createReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/api/packages", c.baseURL)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to create package: %s - %s", resp.Status, string(body))
	}

	var pkg Package
	if err := json.NewDecoder(resp.Body).Decode(&pkg); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &pkg, nil
}

func (c *HTTPClient) SearchPackages(query string, tags []string) ([]*Package, error) {
	url := fmt.Sprintf("%s/api/packages/search?q=%s", c.baseURL, query)
	for _, tag := range tags {
		url += fmt.Sprintf("&tags=%s", tag)
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("search failed: %s", resp.Status)
	}

	var packages []*Package
	if err := json.NewDecoder(resp.Body).Decode(&packages); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return packages, nil
}

func (c *HTTPClient) PushVersion(packageName, version string, zipData []byte) (*Version, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Add version field
	err := writer.WriteField("version", version)
	if err != nil {
		return nil, err
	}

	// Add zip file
	part, err := writer.CreateFormFile("zip", fmt.Sprintf("%s-%s.zip", packageName, version))
	if err != nil {
		return nil, fmt.Errorf("failed to create form file: %w", err)
	}

	if _, err := part.Write(zipData); err != nil {
		return nil, fmt.Errorf("failed to write zip data: %w", err)
	}

	writer.Close()

	url := fmt.Sprintf("%s/api/packages/%s/versions", c.baseURL, packageName)
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := c.do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("push failed: %s - %s", resp.Status, string(body))
	}

	var ver Version
	if err := json.NewDecoder(resp.Body).Decode(&ver); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &ver, nil
}

func (c *HTTPClient) PullVersion(packageName, version, outputDir string) error {
	url := fmt.Sprintf("%s/api/packages/%s/versions/%s/files", c.baseURL, packageName, version)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("pull failed: %s - %s", resp.Status, string(body))
	}

	// Create output directory
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Read response content
	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	// Check if response is a ZIP archive
	contentType := resp.Header.Get("Content-Type")
	if contentType == "application/zip" || strings.HasSuffix(resp.Header.Get("Content-Disposition"), ".zip") {
		// Extract ZIP archive
		return c.extractZipArchive(content, outputDir)
	}

	// Handle single file response (fallback)
	filename := fmt.Sprintf("%s-%s.proto", packageName, version)
	if disposition := resp.Header.Get("Content-Disposition"); disposition != "" {
		if parts := strings.Split(disposition, "filename="); len(parts) > 1 {
			filename = strings.Trim(parts[1], "\"")
		}
	}

	outputFile := filepath.Join(outputDir, filename)
	if err := os.WriteFile(outputFile, content, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

func (c *HTTPClient) extractZipArchive(content []byte, outputDir string) error {
	reader, err := zip.NewReader(bytes.NewReader(content), int64(len(content)))
	if err != nil {
		return fmt.Errorf("failed to create zip reader: %w", err)
	}

	for _, file := range reader.File {
		// Extract file path
		filePath := filepath.Join(outputDir, file.Name)

		// Create directory if needed
		if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
			return fmt.Errorf("failed to create directory for %s: %w", filePath, err)
		}

		// Skip directories
		if file.FileInfo().IsDir() {
			continue
		}

		// Extract file
		rc, err := file.Open()
		if err != nil {
			return fmt.Errorf("failed to open zip entry %s: %w", file.Name, err)
		}

		fileContent, err := io.ReadAll(rc)
		rc.Close()
		if err != nil {
			return fmt.Errorf("failed to read zip entry %s: %w", file.Name, err)
		}

		if err := os.WriteFile(filePath, fileContent, 0644); err != nil {
			return fmt.Errorf("failed to write file %s: %w", filePath, err)
		}
	}

	return nil
}

func (c *HTTPClient) ListVersions(packageName string) ([]*Version, error) {
	url := fmt.Sprintf("%s/api/packages/%s/versions", c.baseURL, packageName)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to list versions: %s", resp.Status)
	}

	var versions []*Version
	if err := json.NewDecoder(resp.Body).Decode(&versions); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return versions, nil
}

func (c *HTTPClient) ViewSchema(packageName, version string) (*SchemaView, error) {
	url := fmt.Sprintf("%s/api/packages/%s/versions/%s/schema", c.baseURL, packageName, version)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("view schema failed: %s - %s", resp.Status, string(body))
	}

	var schema SchemaView
	if err := json.NewDecoder(resp.Body).Decode(&schema); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &schema, nil
}

func (c *HTTPClient) GenerateCode(packageName, version, language, outputDir string, options GenerateOptions) (*GenerateResult, error) {
	generateReq := map[string]interface{}{
		"language":     language,
		"output_dir":   outputDir,
		"package_name": options.PackageName,
		"module_path":  options.ModulePath,
	}

	jsonData, err := json.Marshal(generateReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/api/packages/%s/versions/%s/generate", c.baseURL, packageName, version)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("generate failed: %s - %s", resp.Status, string(body))
	}

	var result GenerateResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

func ParsePackageRef(ref string) (pkg, version string, err error) {
	parts := strings.Split(ref, ":")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("expected format package:version")
	}
	return parts[0], parts[1], nil
}
