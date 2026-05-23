package scan

import (
	"path/filepath"
	"strings"
)

// systemBlocked is the hardcoded path blocklist from PRD §6.1.
// It is compiled into the binary as a constant and cannot be overridden
// at runtime via config, flags, or environment variables.
var systemBlocked = []string{
	"/proc",
	"/sys",
	"/dev",
	"/System",
	"/private/var/vm",
	"/bin",
	"/sbin",
	"/usr/bin",
	"/usr/sbin",
	"/etc",
	"/boot",
	"/lib",
	"/lib64",
	"/usr/lib",
	"/.Spotlight-V100",
}

// isBlocklisted reports whether path is on or under a blocklisted prefix.
// path must be absolute and cleaned.
func isBlocklisted(path string) bool {
	for _, b := range systemBlocked {
		if path == b || strings.HasPrefix(path, b+string(filepath.Separator)) {
			return true
		}
	}
	return false
}

// isBlocklistedHome also checks home-relative credential paths (~/.ssh, ~/.gnupg)
// which cannot be resolved to absolute paths without knowing the home directory.
func isBlocklistedHome(path, home string) bool {
	if isBlocklisted(path) {
		return true
	}
	for _, rel := range []string{".ssh", ".gnupg"} {
		b := filepath.Join(home, rel)
		if path == b || strings.HasPrefix(path, b+string(filepath.Separator)) {
			return true
		}
	}
	return false
}

// IsBlocklistedHome is the exported variant used by the clean package.
func IsBlocklistedHome(path, home string) bool { return isBlocklistedHome(path, home) }

// IsBlocklisted is the exported absolute-path-only variant used by safeRemove
// for its independent Layer 2 syscall-site re-validation (PRD §6.2).
func IsBlocklisted(path string) bool { return isBlocklisted(path) }
