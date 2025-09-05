package protoc

import (
	"fmt"
)

type GenerateOptions struct {
	ProjectPath   string
	Options       map[string]string
	CustomPlugins []CustomPlugin // Custom plugins from project config
}

func (e *Executor) GenerateCode(language string, protoFiles []string, outputDir string, options GenerateOptions) error {
	if err := e.ensureProtoc(); err != nil {
		return err
	}

	args := make([]string, 0)

	// Ensure plugins are available
	pluginManager, err := NewPluginManager()
	if err != nil {
		return fmt.Errorf("failed to create plugin manager: %w", err)
	}

	pluginArgs, err := pluginManager.Process(language, outputDir, options.CustomPlugins)
	if err != nil {
		return fmt.Errorf("failed to process plugins: %w", err)
	}
	args = append(args, pluginArgs...)
	args = append(args, fmt.Sprintf("--proto_path=%s", options.ProjectPath))

	// Add any additional options
	for k, v := range options.Options {
		args = append(args, fmt.Sprintf("--%s=%s", k, v))
	}
	return e.Run(protoFiles, args...)
}
