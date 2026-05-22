//go:build !darwin && !linux

package trash

import "fmt"

func platformAvailable() bool { return false }

func platformMove(abs string) error {
	return fmt.Errorf("trash is not supported on this platform: move items manually or use --no-trash to permanently delete")
}
