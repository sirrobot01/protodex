package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/sirrobot01/protodex/internal/cli/style"
	"github.com/sirrobot01/protodex/internal/manager"
)

var validateCmd = &cobra.Command{
	Use:   "validate [dir]",
	Short: "Validate protobuf schemas",
	Long: `Validate protobuf schema files from a specified directory (default is the current directory).

If no directory is specified, the command will validate proto files in the current working directory.

Examples:
  protodex validate # Validate proto files in the current directory
  protodex validate ./dir # Validate proto files in the specified local directory`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {

		dir := "."
		if len(args) == 1 {
			dir = args[0]
		}

		// Resolve absolute path
		abs, err := filepath.Abs(dir)
		if err != nil {
			return fmt.Errorf("failed to resolve directory: %w", err)
		}
		// Check if directory exists
		if _, err := os.Stat(abs); os.IsNotExist(err) {
			return fmt.Errorf("directory does not exist: %s", abs)
		}
		dir = abs

		pm, err := manager.NewManager(dir)
		if err != nil {
			return err
		}

		protoFiles, err := pm.GetProtoFiles()
		if err != nil {
			return fmt.Errorf("failed to get proto files: %w", err)
		}
		if len(protoFiles) == 0 {
			return fmt.Errorf("no proto files found")
		}
		fmt.Printf("%s\n", style.FileCount(len(protoFiles), "proto file"))
		fmt.Printf("%s\n", style.Validate(fmt.Sprintf("Validating %d proto files", len(protoFiles))))

		if err := pm.Validate(protoFiles); err != nil {
			return fmt.Errorf("file validation failed: %w", err)
		}

		fmt.Printf("%s\n", style.Success("Validation successful"))
		return nil
	},
}
