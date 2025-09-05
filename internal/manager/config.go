package manager

import (
	"fmt"

	"github.com/sirrobot01/protodex/internal/manager/dependency"
	"github.com/sirrobot01/protodex/internal/manager/fetcher"
)

type ProjectConfig struct {
	Package      PackageConfig       `yaml:"package"`
	Files        FileConfig          `yaml:"files"`
	Generation   GenerationConfig    `yaml:"gen"`
	Dependencies []dependency.Config `yaml:"deps"`
	Plugins      []PluginConfig      `yaml:"plugins"` // Global plugins for all languages
}

func (pc ProjectConfig) GetLanguage(lang string) *LanguageConfig {
	for _, l := range pc.Generation.Languages {
		if l.Name == lang {
			return &l
		}
	}
	return nil
}

// GetAllPlugins returns all plugins for a language (global + language-specific)
func (pc ProjectConfig) GetAllPlugins(lang string) []PluginConfig {
	var allPlugins []PluginConfig

	// Add global plugins first
	allPlugins = append(allPlugins, pc.Plugins...)

	// Add language-specific plugins
	if langConfig := pc.GetLanguage(lang); langConfig != nil {
		allPlugins = append(allPlugins, langConfig.Plugins...)
	}

	return allPlugins
}

type PackageConfig struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
}

type FileConfig struct {
	Exclude []string `yaml:"exclude"`
	BaseDir string   `yaml:"base_dir"`
}

type GenerationConfig struct {
	Languages []LanguageConfig `yaml:"languages"`
}

type PluginConfig struct {
	Name      string            `yaml:"name"`       // Plugin name (e.g., "twirp", "grpc-gateway")
	Command   string            `yaml:"command"`    // Plugin command (e.g., "protoc-gen-twirp")
	Options   map[string]string `yaml:"options"`    // Plugin-specific options
	OutputDir string            `yaml:"output_dir"` // Output directory for this plugin. e.g twirp_out=./gen/go
	Required  bool              `yaml:"required"`
}

type LanguageConfig struct {
	Name      string            `yaml:"name"`
	OutputDir string            `yaml:"output_dir"`
	Options   map[string]string `yaml:"options"` // Base language options
	Plugins   []PluginConfig    `yaml:"plugins"` // Language-specific plugins
}

const ProjectConfigFileName = "protodex.yaml"

func (m *Manager) CreateDefaultConfig(packageName, description string) error {
	config := NewDefaultConfig(packageName, description)
	// Save the default config to file
	if err := m.Save(config); err != nil {
		return fmt.Errorf("failed to save default config: %w", err)
	}
	return nil
}

func NewDefaultConfig(packageName, description string) *ProjectConfig {
	return &ProjectConfig{
		Package: PackageConfig{
			Name:        packageName,
			Description: description,
		},
		Files: FileConfig{
			Exclude: []string{},
			BaseDir: ".",
		},
		Generation: GenerationConfig{
			Languages: []LanguageConfig{
				{
					Name:      "go",
					OutputDir: "./gen/go",
				},
			},
		},
		Dependencies: []dependency.Config{
			{
				Name: "google/protobuf",
				Type: fetcher.SourceGoogleWellKnown,
			},
		},
	}
}
