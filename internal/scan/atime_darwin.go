//go:build darwin

package scan

import (
	"os"
	"syscall"
	"time"
)

// accessTime extracts atime from a FileInfo obtained via os.Lstat.
// On APFS with noatime, atime == mtime; callers must check before using.
func accessTime(info os.FileInfo) time.Time {
	stat, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		return time.Time{}
	}
	return time.Unix(int64(stat.Atimespec.Sec), int64(stat.Atimespec.Nsec))
}
