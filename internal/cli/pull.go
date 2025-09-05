package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/sirrobot01/protodex/internal/cli/style"
	"github.com/sirrobot01/protodex/internal/client"
)

var pullCmd = &cobra.Command{
	Use:   "pull [package:version] [output-path]",
	Short: "Pull protobuf schema(s) from the registry",
	Long: `Pull protobuf schema(s) from the registry. If no output path is specified, saves to current directory.

When run within a protodex project, this command can also add the package as a dependency to your protodex.yaml.

Examples:
  protodex pull user-service:v1.0.0
  protodex pull user-service:latest ./schemas
`,
	Args: cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		packageRef := args[0]

		outputPath := "."
		if len(args) == 2 {
			outputPath = args[1]
		}

		// Parse package reference
		pkg, version, err := client.ParsePackageRef(packageRef)
		if err != nil {
			return fmt.Errorf("invalid package reference: %w", err)
		}

		c, err := client.New()
		if err != nil {
			return fmt.Errorf("failed to create client: %w", err)
		}

		absPath, err := filepath.Abs(outputPath)
		if err != nil {
			return fmt.Errorf("failed to get absolute path for %s: %w", outputPath, err)
		}

		fmt.Printf("%s\n", style.Download(fmt.Sprintf("Pulling %s to %s", packageRef, absPath)))

		// Create output directory if it doesn't exist
		if err := os.MkdirAll(absPath, 0755); err != nil {
			return fmt.Errorf("failed to create output directory: %w", err)
		}

		// Show spinner during pull
		spinner := style.NewSpinner()
		done := spinner.Start(fmt.Sprintf("Downloading %s", packageRef))

		// Pull the package (downloads and extracts zip)
		err = c.PullVersion(pkg, version, absPath)
		done <- true

		if err != nil {
			return fmt.Errorf("pull failed: %w", err)
		}

		fmt.Printf("\r%s\n", style.Success(fmt.Sprintf("Successfully pulled and extracted %s to %s", packageRef, absPath)))
		return nil
	},
}
