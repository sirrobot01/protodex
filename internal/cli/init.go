package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/sirrobot01/protodex/internal/cli/style"
	"github.com/sirrobot01/protodex/internal/manager"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new protodex project",
	Long: `Initialize a new protodex project by creating a protodex.yaml configuration file.

This creates a project-level configuration file that can be used to manage schemas,
specify file patterns, and configure code generation settings.

Examples:
  protodex init
  protodex init user-service`,
	Aliases: []string{"initialize"},
	Args:    cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		projectName := ""
		if len(args) > 1 {
			projectName = args[0]
		}
		projectDir, _ := cmd.Flags().GetString("dir")
		description, _ := cmd.Flags().GetString("description")

		if projectDir == "" {
			// Get current directory
			cwd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("failed to get current directory: %w", err)
			}
			projectDir = cwd
		}

		if projectName == "" {
			projectName = filepath.Base(projectDir)
		}

		pm, err := manager.NewManager(projectDir)
		if err != nil {
			return fmt.Errorf("failed to initialize project manager: %w", err)
		}

		configFile := pm.ConfigFile()
		if _, err := os.Stat(configFile); err == nil {
			// Check if it's a valid protodex project
			if _, err := manager.CheckConfig(configFile); err == nil {
				return fmt.Errorf("a protodex project already exists in %s: %w", projectDir, err)
			}
			// The file exists but is not a valid config, we can overwrite it
		}

		if err := pm.CreateDefaultConfig(projectName, description); err != nil {
			return fmt.Errorf("failed to save project config: %w", err)
		}

		fmt.Printf("%s\n", style.Success(fmt.Sprintf("Initialized protodex project in %s", projectDir)))
		fmt.Printf("  %s %s\n", style.Subtle("Configuration file created:"), style.Bold(pm.ConfigFile()))
		fmt.Printf("  %s %s\n", style.Subtle("Package name:"), style.Bold(projectName))
		return nil
	},
}

func init() {
	initCmd.Flags().StringP("dir", "D", ".", "Directory to initialize the project in")
	initCmd.Flags().StringP("description", "d", "", "Package description")
}
