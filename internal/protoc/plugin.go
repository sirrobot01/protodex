package protoc

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

type CustomPlugin struct {
	Name      string            // Plugin name (e.g., "twirp")
	Command   string            // Plugin command (e.g., "protoc-gen-twirp")
	OutputDir string            // Output directory for this plugin. e.g twirp_out=./gen/go, this defaults to outputDir if empty
	Options   map[string]string // Plugin options
	Required  bool
}

type PluginManager struct{}

type Info struct {
	Name        string    // Plugin name, e.g., protoc-gen-go, protoc-gen-go-grpc
	CheckCmd    *exec.Cmd // Command to check if installed
	DownloadCmd *exec.Cmd // Command to download/install if applicable
	Filename    string    // Expected filename after download
}

func NewPluginManager() (*PluginManager, error) {
	return &PluginManager{}, nil
}

func (pm *PluginManager) Process(language, outputDir string, customPlugins []CustomPlugin) ([]string, error) {
	// Build the list of protoc args for the specified language and custom plugins
	// First let's try to validate the language plugin
	if err := pm.ensureLanguagePlugins(language); err != nil {
		return nil, fmt.Errorf("failed to ensure language plugins for %s: %w", language, err)
	}

	// Build args
	var args []string
	args = append(args, fmt.Sprintf("--%s_out=%s", language, outputDir))

	// Let's validate and add custom plugins
	for _, plugin := range customPlugins {
		if err := pm.ensureCustomPlugin(plugin); err != nil {
			if plugin.Required {
				return nil, fmt.Errorf("failed to ensure custom plugin %s: %w", plugin.Name, err)
			}
			fmt.Printf("Warning: Skipping optional plugin %s: %v\n", plugin.Name, err)
			continue
		}

		if plugin.OutputDir == "" {
			plugin.OutputDir = outputDir
		}
		args = append(args, fmt.Sprintf("--%s_out=%s", plugin.Name, plugin.OutputDir))
		for k, v := range plugin.Options {
			args = append(args, fmt.Sprintf("--%s=%s", k, v))
		}
	}

	return args, nil
}

// EnsureLanguagePlugins ensures required plugins for the specified language are available
// This the basic plugin for generating code for the language, not including external plugins like grpc etc.
// For example, Go needs protoc-gen-go
func (pm *PluginManager) ensureLanguagePlugins(language string) error {
	switch strings.ToLower(language) {
	case "go":
		return pm.ensureGoPlugins()
	case "dart":
		return pm.ensureDartPlugins()
	case "rust":
		return pm.ensureRustPlugins()
	case "swift":
		return pm.ensureSwiftPlugins()
	case "ts":
		return pm.ensureTypeScriptPlugins()
	case "js":
		return pm.ensureJavaScriptPlugins()
	default:
		return nil // Languages with built-in support (C++, Java, Python, C#, PHP, Ruby, Objective-C, Kotlin)
	}
}

func (pm *PluginManager) ensureCustomPlugin(plugin CustomPlugin) error {
	// Validate the custom plugin is available
	if _, err := exec.LookPath(plugin.Command); err != nil {
		return fmt.Errorf("plugin command '%s' not found in PATH: %w", plugin.Command, err)
	}
	return nil
}

func (pm *PluginManager) ensureSwiftPlugins() error {
	plugin := Info{
		Name:        "protoc-gen-swift",
		Filename:    "protoc-gen-swift",
		DownloadCmd: exec.Command("brew", "install", "swift-protobuf"),
	}
	if err := pm.ensurePlugin(plugin); err != nil {
		return fmt.Errorf("failed to ensure %s: %w", plugin.Name, err)
	}
	return nil
}

func (pm *PluginManager) ensureTypeScriptPlugins() error {
	// Check for protoc-gen-ts (ts-proto)
	plugin := Info{
		Name:        "protoc-gen-ts",
		Filename:    "protoc-gen-ts",
		DownloadCmd: exec.Command("npm", "install", "-g", "ts-proto"),
	}
	if err := pm.ensurePlugin(plugin); err != nil {
		return fmt.Errorf("failed to ensure %s: %w", plugin.Name, err)
	}
	return nil
}

func (pm *PluginManager) ensureJavaScriptPlugins() error {
	// For modern JS, we can use protoc-gen-es (from Buf)
	plugin := Info{
		Name:        "protoc-gen-es",
		Filename:    "protoc-gen-es",
		DownloadCmd: exec.Command("npm", "install", "-g", "@bufbuild/protoc-gen-es"),
	}
	if err := pm.ensurePlugin(plugin); err != nil {
		return fmt.Errorf("failed to ensure %s: %w", plugin.Name, err)
	}
	return nil
}

func (pm *PluginManager) ensureDartPlugins() error {
	plugin := Info{
		Name:        "protoc-gen-dart",
		Filename:    "protoc-gen-dart",
		DownloadCmd: exec.Command("dart", "pub", "global", "activate", "protoc_plugin"),
	}
	if err := pm.ensurePlugin(plugin); err != nil {
		return fmt.Errorf("failed to ensure %s: %w", plugin.Name, err)
	}
	return nil
}

func (pm *PluginManager) ensureGoPlugins() error {
	plugin := Info{
		Name:        "protoc-gen-go",
		Filename:    "protoc-gen-go",
		DownloadCmd: exec.Command("go", "install", "google.golang.org/protobuf/cmd/protoc-gen-go@latest"),
	}

	// ensure protoc-gen-go exists
	if err := pm.ensurePlugin(plugin); err != nil {
		return fmt.Errorf("failed to ensure %s: %w", plugin.Name, err)
	}
	return nil
}

func (pm *PluginManager) ensureRustPlugins() error {
	plugin := Info{
		Name:        "protoc-gen-rs",
		Filename:    "protoc-gen-rs",
		DownloadCmd: exec.Command("cargo", "install", "protobuf-codegen"),
	}
	if err := pm.ensurePlugin(plugin); err != nil {
		return fmt.Errorf("failed to ensure %s: %w", plugin.Name, err)
	}
	return nil
}

func (pm *PluginManager) ensurePlugin(plugin Info) error {
	// Check if plugin already in PATH
	if _, err := exec.LookPath(plugin.Name); err == nil {
		return nil
	}

	// Only run CheckCmd if LookPath fails and you need version validation
	if plugin.CheckCmd != nil {
		plugin.CheckCmd.Env = os.Environ()
		plugin.CheckCmd.Stdout = nil // Suppress output for checks
		plugin.CheckCmd.Stderr = nil
		if err := plugin.CheckCmd.Run(); err == nil {
			return nil
		}
	}
	fmt.Printf("Downloading %s...\n", plugin.Name)
	if err := pm.installPlugin(plugin); err != nil {
		return fmt.Errorf("failed to download and install %s: %w", plugin.Name, err)
	}
	return nil
}

func (pm *PluginManager) installPlugin(plugin Info) error {
	if plugin.DownloadCmd != nil {
		// Use the provided command to install
		fmt.Printf("Installing %s using command: %s\n", plugin.Name, strings.Join(plugin.DownloadCmd.Args, " "))
		plugin.DownloadCmd.Env = os.Environ()
		plugin.DownloadCmd.Stdout = os.Stdout
		plugin.DownloadCmd.Stderr = os.Stderr
		if err := plugin.DownloadCmd.Run(); err != nil {
			return fmt.Errorf("failed to run install command for %s: %w", plugin.Name, err)
		}
		return nil
	}
	return fmt.Errorf("no download method for plugin %s", plugin.Name)
}
