package fetcher

import (
	"fmt"
	"os"
)

func (f *Fetcher) downloadAndExtract() error {
	zipPath, err := f.download(f.Source)
	if err != nil {
		return fmt.Errorf("failed to download from URL: %w", err)
	}
	defer func() {
		if err := os.Remove(zipPath); err != nil {
			fmt.Printf("Warning: failed to remove temp file: %v\n", err)
		}
	}()
	// Extract zip
	if err := extractZip(zipPath, f.Dest); err != nil {
		return fmt.Errorf("failed to extract zip: %w", err)
	}

	return nil
}
