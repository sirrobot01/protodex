package cli

import (
	"fmt"
	"path/filepath"

	"github.com/sirrobot01/protodex/internal/manager"
	"github.com/spf13/cobra"
)

var depsCmd = &cobra.Command{
	Use:   "deps",
	Short: "Manage dependencies",
	Long:  `Manage project dependencies.`,
}

var resolveCmd = &cobra.Command{
	Use:   "resolve [dir]",
	Short: "Resolve and fetch dependencies",
	Long:  `Resolve and fetch all project dependencies.`,
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		dir := "."
		if len(args) == 1 {
			dir = args[0]
		}

		// Get absolute path
		abs, err := filepath.Abs(dir)
		if err != nil {
			return err
		}

		pm, err := manager.NewManager(abs)
		if err != nil {
			return err
		}
		fmt.Println("Resolving dependencies...")
		if err := pm.ResolveDependencies(); err != nil {
			return fmt.Errorf("failed to resolve dependencies: %w", err)
		}
		fmt.Println("Dependencies resolved successfully.")
		return nil
	},
}

var listCmd = &cobra.Command{
	Use:   "list [dir]",
	Short: "List current dependencies",
	Long:  `List all current project dependencies.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		dir := "."
		if len(args) == 1 {
			dir = args[0]
		}

		// Get absolute path
		abs, err := filepath.Abs(dir)
		if err != nil {
			return err
		}
		pm, err := manager.NewManager(abs)
		if err != nil {
			return err
		}
		deps := pm.Config().Dependencies
		if len(deps) == 0 {
			fmt.Println("No dependencies found.")
			return nil
		}
		fmt.Println("Current dependencies:")
		for _, dep := range deps {
			fmt.Printf("- %s (%s@%s)\n", dep.Name, dep.Source, dep.Version)
		}

		return nil
	},
}

var addCmd = &cobra.Command{
	Use:   "add [name] [source-url] --resolve [dir]",
	Short: "Add a new dependency",
	Long: `Add a new dependency from the specified source URL.

Source URL formats:
- GitHub: github://user/repo@[ref] (e.g., github://user/schemas@main)
- HTTP/HTTPS: Direct HTTP(S) URL to schema archive (e.g., https://example.com/schemas.zip)
- Protodex: protodex://package:version (e.g., protodex://user-service@v1.0.0)
- Local: file://path/to/local/dir or relative/absolute local path (e.g., ./local-schemas or /home/user/schemas)
`,
	Args: cobra.MinimumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		dir := "."
		resolve := false
		name := args[0]
		source := args[1]
		if len(args) > 2 {
			dir = args[2]
		}
		resolve, _ = cmd.Flags().GetBool("resolve")
		// Get absolute path
		abs, err := filepath.Abs(dir)
		if err != nil {
			return err
		}
		pm, err := manager.NewManager(abs)
		if err != nil {
			return err
		}
		fmt.Printf("Adding dependency from %s...\n", source)
		if err := pm.AddDependency(name, source, resolve); err != nil {
			return fmt.Errorf("failed to add dependency: %w", err)
		}
		fmt.Println("Dependency added successfully.")
		return nil
	},
}

func init() {
	addCmd.Flags().BoolP("resolve", "r", false, "Resolve the dependency immediately after adding")

	depsCmd.AddCommand(resolveCmd)
	depsCmd.AddCommand(listCmd)
	depsCmd.AddCommand(addCmd)
}
