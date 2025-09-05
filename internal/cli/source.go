package cli

import (
	"fmt"
	"os"

	"github.com/sirrobot01/protodex/internal/manager/fetcher"
	"github.com/spf13/cobra"
)

var sourceCmd = &cobra.Command{
	Use:   "source [source URL]",
	Short: "Check if a source is valid",
	Long: `Check if a source is valid.

Sources:
  protodex:        protodex://package@version (e.g., protodex:user-service@v1.0.0)
  github:        github://user/repo@main (e.g., github://user/schemas@main)
  http/https:     Direct HTTP(S) URL to schema archive
Examples:
  protodex source protodex://user-service@v1.0.0
  protodex source github://user/schemas@main
  protodex source https://example.com/schemas.zip`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		source := args[0]
		dest, err := os.MkdirTemp("", "protodex-source-*")
		if err != nil {
			return fmt.Errorf("failed to create temp directory: %w", err)
		}
		// Ensure temp directory is cleaned up after check
		defer func(path string) {
			err := os.RemoveAll(path)
			if err != nil {
				fmt.Printf("Warning: failed to remove temp directory %s: %v\n", path, err)
			}
		}(dest)

		fch, err := fetcher.NewFetcherFromURL(source, dest)
		if err != nil {
			return fmt.Errorf("failed to initialize fetcher: %w", err)
		}

		if err := fch.Fetch(); err != nil {
			return fmt.Errorf("failed to fetch source: %w", err)
		}
		fmt.Printf("Source %s is valid and was fetched successfully.\n", source)

		return nil
	},
}
