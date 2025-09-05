package cli

import (
	"fmt"
	"os"

	"github.com/sirrobot01/protodex/internal/cli/style"
	"github.com/sirrobot01/protodex/internal/manager"
	"github.com/sirrobot01/protodex/internal/manager/fetcher"
	"github.com/spf13/cobra"
)

var generateCmd = &cobra.Command{
	Use:   "generate [language] [flags]",
	Short: "Generate code from a protobuf schema",
	Long: `Generate code from a protobuf schema in the specified language.
	
Supported languages: All languages supported by protoc.

Sources:
  local:          Local directory (default if no source specified)
  protodex:        protodex://package:version (e.g., protodex://user-service@v1.0.0)
  github:            github://user/repo@[ref] (e.g., github://user/schemas@main)
  http/https:     Direct HTTP(S) URL to schema archive

Examples:
  protodex generate go --output ./gen # Generate using project config
  protodex generate go ./dir -o ./gen # Generate Go code locally
  protodex generate go protodex://user-service@v1.0.0 -o ./gen
  protodex generate java github://user/schemas@main -o ./gen
  protodex generate ruby protodex://common-types@v2.1.0 -o ./gen
  protodex generate python https://example.com/schemas.zip -o ./gen`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		language := args[0]
		source := "."
		if len(args) > 1 {
			source = args[1]
		}
		outputDir, _ := cmd.Flags().GetString("output")
		fch, err := fetcher.NewFetcherFromURL(source, "")
		if err != nil {
			return fmt.Errorf("failed to initialize fetcher: %w", err)
		}

		// Create a temporary directory to fetch the source into
		// If source is local, use the provided path directly
		if fch.SourceType != fetcher.SourceLocal {
			tempDir, err := os.MkdirTemp("", "protodex-generate-*")
			if err != nil {
				return fmt.Errorf("failed to create temp directory: %w", err)
			}
			// Ensure temp directory is cleaned up after generation
			defer func(path string) {
				err := os.RemoveAll(path)
				if err != nil {
					fmt.Printf("Warning: failed to remove temp directory %s: %v\n", path, err)
				}
			}(tempDir)
			fch.Dest = tempDir
		}

		if err := fch.Fetch(); err != nil {
			return fmt.Errorf("failed to fetch source: %w", err)
		}

		pm, err := manager.NewManager(fch.Dest)
		if err != nil {
			return fmt.Errorf("failed to initialize project manager: %w", err)
		}

		// Let's get the proto files
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

		// If no language specified, use project config
		options := manager.LanguageConfig{
			Name:      language,
			OutputDir: outputDir,
		}

		if err := pm.Generate(protoFiles, options); err != nil {
			return fmt.Errorf("code generation failed: %w", err)
		}

		fmt.Printf("%s\n", style.Success("Code generation completed successfully"))
		return nil
	},
}

func init() {
	generateCmd.Flags().StringP("output", "o", "", "Output directory for generated code")
}
