//go:build !darwin && !linux

package scan

func isNetworkMount(path string) bool { return false }
