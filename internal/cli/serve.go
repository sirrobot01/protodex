package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/sirrobot01/protodex/internal/config"
	"github.com/spf13/cobra"

	"github.com/sirrobot01/protodex/internal/cli/style"
	"github.com/sirrobot01/protodex/internal/server"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the protodex server",
	Long:  `Start the protodex server with both API and web interface.`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		port, _ := cmd.Flags().GetInt("port")
		dataDir, _ := cmd.Flags().GetString("data-dir")

		if dataDir == "" {
			// Set default data dir to ~/.protodex/data
			dataDir = filepath.Join(config.GetProtodexPath(), "data")
			// Create the directory if it doesn't exist
			if err := os.MkdirAll(dataDir, 0755); err != nil {
				return fmt.Errorf("failed to create data directory: %w", err)
			}
			fmt.Printf("%s %s\n", style.Subtle("Using default data directory:"), style.Bold(dataDir))
		}

		// Start API server
		server := server.New(dataDir, port)

		fmt.Printf("%s\n", style.Info(fmt.Sprintf("Starting protodex server on port %d", port)))
		fmt.Printf("%s %s\n", style.Subtle("API:"), style.Bold(fmt.Sprintf("http://localhost:%d/api", port)))
		fmt.Printf("%s %s\n", style.Subtle("Web UI:"), style.Bold(fmt.Sprintf("http://localhost:%d", port)))

		return server.Start(port)
	},
}

func init() {
	serveCmd.Flags().IntP("port", "p", 8080, "Port to run the server on")
	serveCmd.Flags().StringP("data-dir", "d", "", "Directory to store data")
}
