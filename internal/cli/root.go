package cli

import (
	"github.com/sirrobot01/protodex/internal/config"
	"github.com/spf13/cobra"
)

var (
	version       = "0.0.0-dev" // Default version, overridden at build time
	protocVersion = config.DefaultProtocVersion
)

var rootCmd = &cobra.Command{
	Use:     "protodex",
	Short:   "A Git-like protobuf schema registry",
	Version: version,
	Long:    `Protodex is a lightweight, self-hosted protobuf schema registry that provides Git-like operations for managing, versioning, and distributing protocol buffer schemas.`,
	PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
		configPath, _ := cmd.Flags().GetString("config")
		config.SetConfigPath(configPath)
		config.Get()
		return nil
	},
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Add flags
	rootCmd.PersistentFlags().StringP("log-level", "l", "info", "Logging level (debug, info, warn, error)")

	// Set version format
	rootCmd.SetVersionTemplate(`
protodex version: {{.Version}}
protoc version: ` + protocVersion + `
`)

	// Add subcommands
	rootCmd.AddCommand(serveCmd)
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(validateCmd)
	rootCmd.AddCommand(generateCmd)
	rootCmd.AddCommand(configCmd)
	rootCmd.AddCommand(pullCmd)
	rootCmd.AddCommand(pushCmd)
	rootCmd.AddCommand(loginCmd)
	rootCmd.AddCommand(logoutCmd)
	rootCmd.AddCommand(sourceCmd)
	rootCmd.AddCommand(depsCmd)
}
