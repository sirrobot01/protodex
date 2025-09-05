package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/sirrobot01/protodex/internal/cli/style"
	"github.com/sirrobot01/protodex/internal/config"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Show your current configuration located at $HOME/.protodex/config.yaml",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := config.Get()

		fmt.Printf("%s %s\n\n", style.Subtle("Configuration file:"), style.Bold(cfg.ConfigPath))

		fmt.Printf("%s %s(%s)\n", style.Subtle("protoc_bin:"), style.Bold(cfg.Protoc.Bin), style.Subtle(cfg.Protoc.Version))

		fmt.Printf("%s %s\n", style.Subtle("registry:"), style.Bold(cfg.Registry))
		token := cfg.HashedToken
		if token != "" {
			token = token[:8] + "..." // Mask token
		}
		fmt.Printf("%s %s\n", style.Subtle("token:"), style.Bold(token))
		database := cfg.Database
		fmt.Printf("%s %s(%s)\n", style.Subtle("database:"), style.Bold(database.DSN), database.Type)
		return nil
	},
}
