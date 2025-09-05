package config

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"gopkg.in/yaml.v3"
)

type DatabaseType string

const (
	DatabaseTypeSQLite   DatabaseType = "sqlite"
	DatabaseTypePostgres DatabaseType = "postgres"
	DefaultRegistryURL                = "http://localhost:8080"
	DefaultProtocVersion              = "32.0"
)

type DatabaseConfig struct {
	Type DatabaseType `yaml:"type"`
	DSN  string       `yaml:"dsn"`
}

type ProtocConfig struct {
	Bin     string `yaml:"bin"`
	Version string `yaml:"version"`
}

type Config struct {
	Protoc      ProtocConfig   `yaml:"protoc"`
	LogLevel    string         `yaml:"log_level"`
	Registry    string         `yaml:"registry"`
	HashedToken string         `yaml:"hashed_token"`
	Database    DatabaseConfig `yaml:"database"`
	ConfigPath  string         `yaml:"config_path"`
}

var (
	globalConfig *Config
	once         sync.Once
	configPath   string
)

func SetConfigPath(path string) {
	configPath = path
}

// Get returns the global config instance
func Get() *Config {
	once.Do(func() {
		defaultProtocPath := filepath.Join(GetProtodexPath(), "bin", "protoc")

		globalConfig = &Config{
			Protoc: ProtocConfig{
				Bin:     defaultProtocPath,
				Version: DefaultProtocVersion,
			},
			LogLevel: "info",
			Registry: DefaultRegistryURL,
			Database: DatabaseConfig{
				Type: DatabaseTypeSQLite,
				DSN:  "protodex.db",
			},
		}
		_configPath := getConfigPath(configPath)
		if _, err := os.Stat(_configPath); !os.IsNotExist(err) {
			data, err := os.ReadFile(_configPath)
			if err == nil {
				if err := yaml.Unmarshal(data, globalConfig); err == nil {
					globalConfig.ConfigPath = _configPath
					return
				}
			}
		}
	})
	return globalConfig
}

func getConfigPath(override string) string {
	if override != "" {
		return override
	}

	return filepath.Join(GetProtodexPath(), "config.yaml")
}

func (c *Config) Save() error {
	if err := os.MkdirAll(filepath.Dir(c.ConfigPath), 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(c.ConfigPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}
	return nil
}

func (c *Config) GetRegistryURL(flagRegistry string) string {
	if flagRegistry != "" {
		return flagRegistry
	}
	if url := os.Getenv("PROTODEX_REGISTRY_URL"); url != "" {
		return url
	}
	if c.Registry == "" {
		return getDefaultRegistryURL()
	}
	return c.Registry
}

func (c *Config) GetToken(flagToken string) string {
	if flagToken != "" {
		return flagToken
	}
	return c.HashedToken
}

func getDefaultRegistryURL() string {
	if url := os.Getenv("PROTODEX_REGISTRY_URL"); url != "" {
		return url
	}
	return DefaultRegistryURL
}

func GetProtodexPath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(homeDir, ".protodex")
}
