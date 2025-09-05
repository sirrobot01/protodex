package cli

import (
	"archive/zip"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/sirrobot01/protodex/internal/cli/style"
	"github.com/sirrobot01/protodex/internal/client"
	"github.com/sirrobot01/protodex/internal/manager"
)

var pushCmd = &cobra.Command{
	Use:   "push version",
	Short: "Push protobuf schema(s) to the registry",
	Long: `Push one or more protobuf schemas to the registry with the specified package name and version.

If no proto files are specified, the command will look for a protodex.yaml project config file
and use the files specified there.

Examples:
  protodex push v1.0.0 # Push a single file to version v1.0.0
  protodex push v1.0.0 ./dir # Push all proto files in the specified directory to version v1.0.0
`,
	Args: cobra.MaximumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		// First argument is always the version
		version := args[0]
		dir := "."
		if len(args) == 2 {
			dir = args[1]
		}

		// Resolve absolute path
		projectDir, err := filepath.Abs(dir)
		if err != nil {
			return fmt.Errorf("failed to resolve project directory: %w", err)
		}

		// Check if project directory exists
		if _, err := os.Stat(projectDir); os.IsNotExist(err) {
			return fmt.Errorf("project directory does not exist: %s", projectDir)
		}
		pm, err := manager.NewManager(projectDir)

		if err != nil {
			return fmt.Errorf("failed to initialize project manager: %w", err)
		}
		protoFiles, err := pm.GetProtoFiles() // Local source only for push
		if err != nil {
			return fmt.Errorf("failed to get proto files: %w", err)
		}
		if len(protoFiles) == 0 {
			return fmt.Errorf("no proto files found")
		}
		fmt.Printf("Found %s\n", style.FileCount(len(protoFiles), "proto file"))
		fmt.Printf("%s\n", style.Validate(fmt.Sprintf("Validating %d proto files", len(protoFiles))))

		if err := pm.Validate(protoFiles); err != nil {
			return fmt.Errorf("file validation failed: %w", err)
		}
		fmt.Printf("%s\n", style.Success("Validation successful"))

		// Get package name from config
		packageName := pm.Config().Package.Name
		if packageName == "" {
			return fmt.Errorf("package name not found in config - please set 'package.name' in protodex.yaml")
		}

		// Prepare files to include in zip
		allFiles := make([]string, 0, len(protoFiles)+2)
		allFiles = append(allFiles, protoFiles...)

		// Add protodex.yaml if it exists
		configFile := filepath.Join(projectDir, "protodex.yaml")
		if _, err := os.Stat(configFile); err == nil {
			allFiles = append(allFiles, configFile)
		}

		// Add README.md if it exists
		readmeFile := filepath.Join(projectDir, "README.md")
		if _, err := os.Stat(readmeFile); err == nil {
			allFiles = append(allFiles, readmeFile)
		}

		fmt.Printf("%s\n", style.Upload(fmt.Sprintf("Bundling %d files into zip archive", len(allFiles))))

		// Create zip archive
		zipData, err := createProjectZip(allFiles, projectDir)
		if err != nil {
			return fmt.Errorf("failed to create zip archive: %w", err)
		}

		// Initialize client and push to registry
		c, err := client.New()
		if err != nil {
			return fmt.Errorf("failed to initialize client: %w", err)
		}

		// Show spinner during push
		spinner := style.NewSpinner()
		done := spinner.Start(fmt.Sprintf("Pushing %s:%s", packageName, version))

		pushedVersion, err := c.PushVersion(packageName, version, zipData)
		done <- true

		if err != nil {
			return fmt.Errorf("failed to push to registry: %w", err)
		}

		fmt.Printf("\r%s\n", style.Success(fmt.Sprintf("Successfully pushed %s to registry", style.Version(packageName, pushedVersion.Version))))
		fmt.Printf("%s %s\n", style.Subtle("Version ID:"), style.Bold(pushedVersion.ID))
		fmt.Printf("%s %s\n", style.Subtle("Created at:"), style.Bold(pushedVersion.CreatedAt.Format("2006-01-02 15:04:05")))
		return nil
	},
}

// createProjectZip creates a zip archive containing all project files with proper directory structure
func createProjectZip(filePaths []string, projectDir string) ([]byte, error) {
	var buf bytes.Buffer
	zipWriter := zip.NewWriter(&buf)

	for _, filePath := range filePaths {
		// Get relative path from project directory
		relPath, err := filepath.Rel(projectDir, filePath)
		if err != nil {
			return nil, fmt.Errorf("failed to get relative path for %s: %w", filePath, err)
		}

		// Skip files outside project directory
		if strings.HasPrefix(relPath, "..") {
			continue
		}

		// Read file content
		content, err := os.ReadFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("failed to read file %s: %w", filePath, err)
		}

		// Create entry in zip with relative path to maintain structure
		f, err := zipWriter.Create(relPath)
		if err != nil {
			return nil, fmt.Errorf("failed to create zip entry for %s: %w", relPath, err)
		}

		// Write file content to zip
		if _, err := f.Write(content); err != nil {
			return nil, fmt.Errorf("failed to write content to zip entry for %s: %w", relPath, err)
		}
	}

	// Close zip writer
	if err := zipWriter.Close(); err != nil {
		return nil, fmt.Errorf("failed to close zip writer: %w", err)
	}

	return buf.Bytes(), nil
}
