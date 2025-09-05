package manager

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/rs/zerolog"
	"github.com/sirrobot01/protodex/internal/config"
	"github.com/sirrobot01/protodex/internal/manager/fetcher"
	"gopkg.in/yaml.v3"

	"github.com/sirrobot01/protodex/internal/logger"
	"github.com/sirrobot01/protodex/internal/manager/dependency"
	"github.com/sirrobot01/protodex/internal/protoc"
)

type Manager struct {
	projectPath string
	configFile  string
	config      *ProjectConfig
	logger      zerolog.Logger
	executor    *protoc.Executor
	resolver    *dependency.Resolver
}

func NewManager(projectPath string) (*Manager, error) {
	resolver, err := dependency.NewResolver()
	if err != nil {
		return nil, fmt.Errorf("failed to create dependency cache: %w", err)
	}
	cfg := config.Get()
	exec := protoc.NewExecutor(cfg.Protoc.Bin, cfg.Protoc.Version, resolver.GetDependencyPath())

	m := &Manager{
		projectPath: projectPath,
		configFile:  filepath.Join(projectPath, ProjectConfigFileName),
		resolver:    resolver,
		logger:      logger.Get(),
		executor:    exec,
	}
	projectConfig, err := m.load()
	if err != nil {
		return nil, fmt.Errorf("failed to load project config: %w", err)
	}
	m.config = projectConfig

	return m, nil
}

func (m *Manager) Config() *ProjectConfig {
	return m.config
}

func (m *Manager) ClearCache() error {
	return m.resolver.Clear()
}

func (m *Manager) ListCachedDependencies() ([]string, error) {
	return m.resolver.ListCached()
}

func (m *Manager) AddDependency(name, source string, resolve bool) error {
	// Parse the source to get dependency config

	sourceInfo, err := fetcher.ParseSource(source)
	if err != nil {
		return fmt.Errorf("failed to parse dependency source: %w", err)
	}
	// Check if dependency already exists
	for _, dep := range m.config.Dependencies {
		if dep.Name == name {
			return fmt.Errorf("dependency %s already exists", dep.Name)
		}
	}
	depConfig := dependency.Config{
		Name:    name,
		Source:  sourceInfo.Source,
		Type:    sourceInfo.Type,
		Version: sourceInfo.Version,
	}
	m.config.Dependencies = append(m.config.Dependencies, depConfig)
	// Save the config
	if err := m.Save(m.config); err != nil {
		return fmt.Errorf("failed to save project config: %w", err)
	}
	if resolve {
		// Resolve dependencies from cache/download
		if err := m.resolver.ResolveDependencies([]dependency.Config{depConfig}); err != nil {
			return fmt.Errorf("failed to resolve dependency %s: %w", name, err)
		}
	}
	return nil
}

func (m *Manager) ResolveDependencies() error {
	if len(m.config.Dependencies) == 0 {
		return nil
	}
	// Resolve dependencies from cache/download
	return m.resolver.ResolveDependencies(m.config.Dependencies)
}

func (m *Manager) Generate(protoFiles []string, options LanguageConfig) error {
	// Get all the files
	// At this stage, we've already validated the files
	if len(protoFiles) == 0 {
		return fmt.Errorf("no proto files to generate")
	}
	cfg := m.Config()
	// Check if user sets an option from config, override config with cli
	opt := cfg.GetLanguage(options.Name)
	if opt != nil {
		// Language is set in config, use cli options
		if options.OutputDir != "" {
			opt.OutputDir = options.OutputDir
		}
		if len(options.Plugins) > 0 {
			opt.Plugins = options.Plugins
		}
		if len(options.Options) > 0 {
			opt.Options = options.Options
		}
	}

	if opt == nil {
		return fmt.Errorf("no language configured")
	}
	if opt.OutputDir == "" {
		return fmt.Errorf("output directory is required for language %s", opt.Name)
	}
	if opt.Name == "" {
		return fmt.Errorf("language name is required")
	}

	// create directory if not exists
	if err := os.MkdirAll(opt.OutputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory %s: %w", opt.OutputDir, err)
	}

	customPlugins := make([]protoc.CustomPlugin, 0)
	for _, plugin := range m.config.GetAllPlugins(opt.Name) {
		customPlugins = append(customPlugins, protoc.CustomPlugin{
			Name:      plugin.Name,
			Command:   plugin.Command,
			OutputDir: plugin.OutputDir,
			Options:   plugin.Options,
			Required:  plugin.Required,
		})
	}
	generateOptions := protoc.GenerateOptions{
		Options:       opt.Options,
		CustomPlugins: customPlugins,
		ProjectPath:   m.projectPath,
	}

	err := m.executor.GenerateCode(opt.Name, protoFiles, opt.OutputDir, generateOptions)
	if err != nil {
		return fmt.Errorf("failed to generate code for language %s: %w", opt.Name, err)
	}
	return nil
}

func (m *Manager) load() (*ProjectConfig, error) {
	projectName := filepath.Base(m.projectPath)
	projectDescription := fmt.Sprintf("Protodex project %s", projectName)
	projectCfg := NewDefaultConfig(projectName, projectDescription)

	if _, err := os.Stat(m.configFile); os.IsNotExist(err) {
		return projectCfg, nil
	}

	data, err := os.ReadFile(m.configFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read project config: %w", err)
	}

	if err := yaml.Unmarshal(data, projectCfg); err != nil {
		return nil, fmt.Errorf("failed to parse project config: %w", err)
	}

	if projectCfg.Files.BaseDir == "" {
		projectCfg.Files.BaseDir = "."
	}

	return projectCfg, nil
}

func (m *Manager) Save(config *ProjectConfig) error {
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal project config: %w", err)
	}

	if err := os.WriteFile(m.configFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write project config: %w", err)
	}

	return nil
}

func (m *Manager) ConfigFile() string {
	return m.configFile
}

func (m *Manager) GetProtoFiles() ([]string, error) {
	baseDir := filepath.Join(m.projectPath, m.config.Files.BaseDir)

	var files []string

	err := filepath.WalkDir(baseDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}

		// Fast suffix check first (most efficient)
		if !strings.HasSuffix(d.Name(), ".proto") {
			return nil
		}

		// Only check excludes if we have a .proto file
		if m.shouldExclude(path, baseDir) {
			return nil
		}

		files = append(files, path)
		return nil
	})

	return files, err
}

func (m *Manager) shouldExclude(path, baseDir string) bool {
	for _, pattern := range m.config.Files.Exclude {
		var fullPattern string
		if filepath.IsAbs(pattern) {
			fullPattern = pattern
		} else {
			fullPattern = filepath.Join(baseDir, pattern)
		}

		if matched, _ := filepath.Match(fullPattern, path); matched {
			return true
		}
	}
	return false
}

func CheckConfig(configFile string) (*ProjectConfig, error) {
	data, err := os.ReadFile(configFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read project config: %w", err)
	}
	cfg := &ProjectConfig{}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("failed to parse project config: %w", err)
	}
	return cfg, nil
}
