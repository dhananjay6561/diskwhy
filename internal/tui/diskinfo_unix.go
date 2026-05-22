//go:build darwin || linux

package tui

import "syscall"

// DiskUsage returns total, used, and free byte counts for the filesystem
// containing path. Returns zeroes if the stat call fails.
func DiskUsage(path string) (total, used, free int64) {
	var stat syscall.Statfs_t
	if err := syscall.Statfs(path, &stat); err != nil {
		return 0, 0, 0
	}
	bsize := int64(stat.Bsize)
	total = int64(stat.Blocks) * bsize
	free = int64(stat.Bavail) * bsize
	if free < 0 {
		free = 0
	}
	if total < free {
		free = total
	}
	used = total - free
	return
}
