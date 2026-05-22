//go:build !darwin && !linux

package scan

import (
	"os"
	"time"
)

// accessTime returns zero on unsupported platforms; callers treat this as
// atime unavailable and fall through to the next staleness signal.
func accessTime(_ os.FileInfo) time.Time { return time.Time{} }
