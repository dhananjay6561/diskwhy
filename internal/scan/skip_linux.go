//go:build linux

package scan

import "syscall"

// Filesystem type magic numbers per linux/magic.h.
const (
	nfsMagic   = 0x6969
	smbMagic   = 0x517b
	cifs2Magic = 0xff534d42
	fuseMagic  = 0x65735546
)

func isNetworkMount(path string) bool {
	var stat syscall.Statfs_t
	if err := syscall.Statfs(path, &stat); err != nil {
		return false
	}
	switch int64(stat.Type) {
	case nfsMagic, smbMagic, cifs2Magic, fuseMagic:
		return true
	}
	return false
}
