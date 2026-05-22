package scan

import (
	"context"
	"os"
	"path/filepath"
)

// dirSize returns the total byte count of all files under path.
// It uses os.ReadDir (Lstat semantics) throughout: symlinks are counted
// by their own size on disk, never by their target's size.
// It returns the partial total accumulated before ctx cancellation.
func dirSize(ctx context.Context, path string) (int64, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return 0, err
	}
	var total int64
	for _, entry := range entries {
		select {
		case <-ctx.Done():
			return total, ctx.Err()
		default:
		}
		info, err := entry.Info() // Lstat semantics; never follows symlinks
		if err != nil {
			continue
		}
		if entry.Type()&os.ModeSymlink != 0 {
			total += info.Size()
			continue
		}
		if entry.IsDir() {
			sub, err := dirSize(ctx, filepath.Join(path, entry.Name()))
			if err == nil {
				total += sub
			}
			// On error (permission denied, etc.) skip and continue
		} else {
			total += info.Size()
		}
	}
	return total, nil
}
