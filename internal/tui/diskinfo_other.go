//go:build !darwin && !linux

package tui

func DiskUsage(_ string) (total, used, free int64) { return 0, 0, 0 }
