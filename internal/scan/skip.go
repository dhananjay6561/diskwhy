package scan

import (
	"path/filepath"
	"runtime"
	"strings"
)

// alwaysSkip lists paths that are never traversed, per PRD §5.1.3.
// These are virtual filesystems and ephemeral directories with no persistent
// disk usage of interest to the user.
var alwaysSkip = []string{
	"/proc",
	"/sys",
	"/dev",
	"/run",
	"/tmp",
	"/.Spotlight-V100",
}

// darwinSkip lists macOS paths that are always skipped.
var darwinSkip = []string{
	"/System",
	"/private/var/vm",
}

// shouldSkip reports whether the scanner should skip traversal of path entirely.
// path must be absolute and cleaned.
func shouldSkip(path, home string) bool {
	// Always-skip paths (PRD §5.1.3).
	for _, s := range alwaysSkip {
		if path == s || strings.HasPrefix(path, s+string(filepath.Separator)) {
			return true
		}
	}
	// macOS-only always-skip paths.
	if runtime.GOOS == "darwin" {
		for _, s := range darwinSkip {
			if path == s || strings.HasPrefix(path, s+string(filepath.Separator)) {
				return true
			}
		}
	}
	// Blocklisted paths are also never traversed.
	if isBlocklistedHome(path, home) {
		return true
	}
	// Network mounts have unreliable mtime and are slow to traverse.
	if isNetworkMount(path) {
		return true
	}
	return false
}
