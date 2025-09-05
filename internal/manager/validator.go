package manager

import (
	"fmt"
	"os"
	"strings"
)

func (m *Manager) Validate(protoFiles []string) error {
	// Check if all files exist and are .proto files
	for _, file := range protoFiles {
		if !strings.HasSuffix(file, ".proto") {
			return fmt.Errorf("file is not a .proto file: %s", file)
		}
		if _, err := os.Stat(file); os.IsNotExist(err) {
			return fmt.Errorf("file does not exist: %s", file)
		}
	}

	return m.validateWithProtoc(protoFiles)
}

func (m *Manager) validateWithProtoc(protoFiles []string) error {
	if err := m.ResolveDependencies(); err != nil {
		return fmt.Errorf("failed to get import paths: %w", err)
	}

	// Build protoc args for validation
	var args []string
	// Use descriptor_set_out to validate without generating files
	args = append(args, "--descriptor_set_out=/dev/null")
	args = append(args, fmt.Sprintf("--proto_path=%s", m.projectPath)) // Project directory

	// Run protoc validation
	if err := m.executor.Run(protoFiles, args...); err != nil {
		return fmt.Errorf("validation failed for %s: %w", protoFiles, err)
	}

	return nil
}
