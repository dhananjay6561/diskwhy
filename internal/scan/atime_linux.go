//go:build linux

package scan

import (
	"os"
	"syscall"
	"time"
)

// accessTime extracts atime from a FileInfo obtained via os.Lstat.
// On Linux with relatime or noatime mount options, atime == mtime;
// callers must check before using.
func accessTime(info os.FileInfo) time.Time {
	stat, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		return time.Time{}
	}
	return time.Unix(stat.Atim.Sec, stat.Atim.Nsec)
}
