package manager

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewManager(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "protodex-manager-test-")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create a test config file
	configPath := filepath.Join(tmpDir, "protodex.yaml")
	configContent := `
package:
  name: "test-package"
  description: "Test package for manager tests"

files:
  base_dir: "."

deps:
  - name: "common-proto"
    source: "protodex:common-proto:v1.0.0"
    version: "v1.0.0"

gen:
  languages:
    - name: "go"
      output_dir: "./gen/go"
    - name: "python"
      output_dir: "./gen/python"
`
	err = os.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(t, err)

	manager, err := NewManager(tmpDir)
	require.NoError(t, err)
	require.NotNil(t, manager)

	config := manager.Config()
	assert.Equal(t, "test-package", config.Package.Name)
	assert.Equal(t, "Test package for manager tests", config.Package.Description)
	assert.Len(t, config.Dependencies, 1)
	assert.Equal(t, "common-proto", config.Dependencies[0].Name)
}

func TestCreateDefaultConfig(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "protodex-manager-test-")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	manager, err := NewManager(tmpDir)
	require.NoError(t, err)

	err = manager.CreateDefaultConfig("test-package", "A test package")
	require.NoError(t, err)

	// Check that config file was created
	configPath := filepath.Join(tmpDir, "protodex.yaml")
	assert.FileExists(t, configPath)

	// Reload manager to test the created config
	manager, err = NewManager(tmpDir)
	require.NoError(t, err)

	config := manager.Config()
	assert.Equal(t, "test-package", config.Package.Name)
	assert.Equal(t, "A test package", config.Package.Description)
	assert.NotNil(t, config.Generation)
}

func TestGetProtoFiles(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "protodex-manager-test-")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create test proto files
	protoDir := filepath.Join(tmpDir, "proto")
	err = os.MkdirAll(protoDir, 0755)
	require.NoError(t, err)

	testFiles := []string{"user.proto", "order.proto", "common.proto"}
	for _, file := range testFiles {
		filePath := filepath.Join(protoDir, file)
		err = os.WriteFile(filePath, []byte("syntax = \"proto3\";"), 0644)
		require.NoError(t, err)
	}

	// Create config that includes these files
	configPath := filepath.Join(tmpDir, "protodex.yaml")
	configContent := `
package:
  name: "test-package"

files:
`
	err = os.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(t, err)

	manager, err := NewManager(tmpDir)
	require.NoError(t, err)

	files, err := manager.GetProtoFiles()
	require.NoError(t, err)

	assert.Len(t, files, 3)

	// Check that all expected files are found
	fileNames := make([]string, len(files))
	for i, file := range files {
		fileNames[i] = filepath.Base(file)
	}

	for _, expectedFile := range testFiles {
		assert.Contains(t, fileNames, expectedFile)
	}
}

func TestGetLanguage(t *testing.T) {
	config := &ProjectConfig{
		Generation: GenerationConfig{
			Languages: []LanguageConfig{
				{
					Name:      "go",
					OutputDir: "./gen/go",
					Options:   map[string]string{"optimize_for": "speed"},
				},
				{
					Name:      "python",
					OutputDir: "./gen/python",
					Options:   map[string]string{"package_name": "test_package"},
				},
			},
		},
	}

	// Test existing language
	goConfig := config.GetLanguage("go")
	require.NotNil(t, goConfig)
	assert.Equal(t, "go", goConfig.Name)
	assert.Equal(t, "./gen/go", goConfig.OutputDir)
	assert.Equal(t, "speed", goConfig.Options["optimize_for"])

	// Test non-existent language
	rustConfig := config.GetLanguage("rust")
	assert.Nil(t, rustConfig)
}

func TestConfigFile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "protodex-manager-test-")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	manager, err := NewManager(tmpDir)
	require.NoError(t, err)

	expectedPath := filepath.Join(tmpDir, "protodex.yaml")
	assert.Equal(t, expectedPath, manager.ConfigFile())
}

func TestSaveConfig(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "protodex-manager-test-")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	manager, err := NewManager(tmpDir)
	require.NoError(t, err)

	// Create a test config
	config := &ProjectConfig{
		Package: PackageConfig{
			Name:        "test-package",
			Description: "Test package",
		},
		Generation: GenerationConfig{
			Languages: []LanguageConfig{
				{
					Name:      "go",
					OutputDir: "./gen/go",
				},
			},
		},
	}

	// Save the config
	err = manager.Save(config)
	require.NoError(t, err)

	// Verify file was created
	configPath := filepath.Join(tmpDir, "protodex.yaml")
	assert.FileExists(t, configPath)

	// Reload and verify
	newManager, err := NewManager(tmpDir)
	require.NoError(t, err)

	loadedConfig := newManager.Config()
	assert.Equal(t, "test-package", loadedConfig.Package.Name)
	assert.Equal(t, "Test package", loadedConfig.Package.Description)
}
