package protoc

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

type Executor struct {
	version    string
	protocPath string
	depsPath   string
}

// NewExecutor creates a new Executor instance.
// depsPath is the path to the directory containing dependencies (typically located in the ~/.protodex/deps directory) like google well-known types.
func NewExecutor(protocPath, version, depsPath string) *Executor {
	return &Executor{
		protocPath: protocPath,
		depsPath:   depsPath,
		version:    version,
	}
}

func (e *Executor) ensureProtoc() error {
	if e.protocPath == "" {
		return fmt.Errorf("protoc binary path is not set")
	}

	// LookPath checks existence AND executability
	if _, err := exec.LookPath(e.protocPath); err != nil {
		// Try to download if it doesn't exist or isn't executable
		fmt.Printf("protoc binary not found in %s, downloading...", filepath.Dir(e.protocPath))
		if downloadErr := e.downloadProtoc(); downloadErr != nil {
			return fmt.Errorf("protoc not found and download failed: %w", downloadErr)
		}

		// Verify after download
		if _, err := exec.LookPath(e.protocPath); err != nil {
			return fmt.Errorf("protoc still not executable after download: %w", err)
		}
	}

	return nil
}

func (e *Executor) Run(protoFiles []string, args ...string) error {
	if err := e.ensureProtoc(); err != nil {
		return err
	}

	// Build protoc args
	var protocArgs []string
	// Import path if set
	if e.depsPath != "" {
		protocArgs = append(protocArgs, fmt.Sprintf("--proto_path=%s", e.depsPath))
	}

	// Add user args and proto file
	protocArgs = append(protocArgs, args...)
	protocArgs = append(protocArgs, protoFiles...)

	cmd := exec.Command(e.protocPath, protocArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}
