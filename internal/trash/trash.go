package trash

import (
	"fmt"
	"os"
	"path/filepath"
)

// Move moves the item at path to the OS trash.
// On macOS it moves to ~/.Trash with collision handling.
// On Linux it follows the XDG Trash specification (FreeDesktop.org).
// It returns an error on unsupported platforms.
func Move(path string) error {
	abs, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("resolve path %q: %w", path, err)
	}
	return platformMove(abs)
}

// Available reports whether the current platform supports trash operations.
func Available() bool {
	return platformAvailable()
}

// uniqueDest returns a collision-free destination path. If dest already exists
// it appends an incrementing counter before the extension until it finds a free
// slot.
func uniqueDest(dest string) string {
	if _, err := os.Lstat(dest); os.IsNotExist(err) {
		return dest
	}
	ext := filepath.Ext(dest)
	base := dest[:len(dest)-len(ext)]
	for i := 1; ; i++ {
		candidate := fmt.Sprintf("%s %d%s", base, i, ext)
		if _, err := os.Lstat(candidate); os.IsNotExist(err) {
			return candidate
		}
	}
}
