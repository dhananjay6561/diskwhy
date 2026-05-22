//go:build linux

package trash

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

func platformAvailable() bool { return true }

func platformMove(abs string) error {
	trashDir, err := xdgTrashDir()
	if err != nil {
		return err
	}
	filesDir := filepath.Join(trashDir, "files")
	infoDir := filepath.Join(trashDir, "info")
	for _, d := range []string{filesDir, infoDir} {
		if err := os.MkdirAll(d, 0700); err != nil {
			return fmt.Errorf("create trash directory %q: %w", d, err)
		}
	}
	dest := uniqueDest(filepath.Join(filesDir, filepath.Base(abs)))
	trashName := filepath.Base(dest)
	if err := os.Rename(abs, dest); err != nil {
		return fmt.Errorf("move %q to trash: %w", abs, err)
	}
	infoContent := fmt.Sprintf("[Trash Info]\nPath=%s\nDeletionDate=%s\n",
		abs, time.Now().Format("2006-01-02T15:04:05"))
	infoPath := filepath.Join(infoDir, trashName+".trashinfo")
	if err := os.WriteFile(infoPath, []byte(infoContent), 0600); err != nil {
		// trashinfo write failure is non-fatal: the file was already moved
		fmt.Fprintf(os.Stderr, "warning: could not write trashinfo for %q: %v\n", trashName, err)
	}
	return nil
}

func xdgTrashDir() (string, error) {
	if xdgData := os.Getenv("XDG_DATA_HOME"); xdgData != "" {
		return filepath.Join(xdgData, "Trash"), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("locate home directory: %w", err)
	}
	return filepath.Join(home, ".local", "share", "Trash"), nil
}
