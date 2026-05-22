//go:build darwin

package trash

import (
	"fmt"
	"os"
	"path/filepath"
)

func platformAvailable() bool { return true }

func platformMove(abs string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("locate home directory: %w", err)
	}
	trashDir := filepath.Join(home, ".Trash")
	if err := os.MkdirAll(trashDir, 0700); err != nil {
		return fmt.Errorf("create ~/.Trash: %w", err)
	}
	dest := uniqueDest(filepath.Join(trashDir, filepath.Base(abs)))
	if err := os.Rename(abs, dest); err != nil {
		return fmt.Errorf("move %q to trash: %w", abs, err)
	}
	return nil
}
