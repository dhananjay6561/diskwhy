//go:build darwin

package scan

import (
	"strings"
	"syscall"
)

func isNetworkMount(path string) bool {
	var stat syscall.Statfs_t
	if err := syscall.Statfs(path, &stat); err != nil {
		return false
	}
	name := int8SliceToString(stat.Fstypename[:])
	lower := strings.ToLower(strings.TrimRight(name, "\x00"))
	return lower == "nfs" || lower == "smbfs" || lower == "cifs" ||
		strings.HasPrefix(lower, "fuse") || lower == "afpfs"
}

func int8SliceToString(s []int8) string {
	b := make([]byte, 0, len(s))
	for _, v := range s {
		if v == 0 {
			break
		}
		b = append(b, byte(v))
	}
	return string(b)
}
