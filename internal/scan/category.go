package scan

import (
	"io/fs"
	"path/filepath"
	"strings"
)

// Category names used in CandidateItem.Category.
const (
	CatNodeModules  = "node_modules"
	CatDocker       = "docker"
	CatBrewCache    = "brew_cache"
	CatXcodeDerived = "xcode_derived"
	CatPipCache     = "pip_cache"
	CatNpmCache     = "npm_cache"
	CatAptCache     = "apt_cache"
	CatSnap         = "snap"
	CatLogs         = "logs"
	CatJournald     = "journald"
	CatTrash        = "trash"
	CatGitObjects   = "git_objects"
	CatPycache      = "pycache"
	CatDownloads    = "downloads"
)

// knownPath associates a well-known filesystem path with a category.
type knownPath struct {
	Path     string
	Category string
}

// KnownCategoryPaths returns the fixed-location category paths for the given
// platform. These are checked directly during quick scan without traversal.
func KnownCategoryPaths(home, goos string) []knownPath {
	paths := []knownPath{
		// npm / yarn / pnpm cache — both platforms
		{filepath.Join(home, ".npm"), CatNpmCache},
		{filepath.Join(home, ".cache", "yarn"), CatNpmCache},
		{filepath.Join(home, ".pnpm-store"), CatNpmCache},
		// pip / uv cache — both platforms
		{filepath.Join(home, ".cache", "pip"), CatPipCache},
		{filepath.Join(home, ".cache", "uv"), CatPipCache},
	}

	switch goos {
	case "darwin":
		paths = append(paths,
			knownPath{filepath.Join(home, "Library", "Caches", "Homebrew"), CatBrewCache},
			knownPath{filepath.Join(home, "Library", "Developer", "Xcode", "DerivedData"), CatXcodeDerived},
			knownPath{filepath.Join(home, ".Trash"), CatTrash},
			knownPath{filepath.Join(home, "Downloads"), CatDownloads},
		)
	case "linux":
		paths = append(paths,
			knownPath{"/var/cache/apt", CatAptCache},
			knownPath{"/var/lib/snapd", CatSnap},
			knownPath{"/var/log/journal", CatJournald},
			knownPath{filepath.Join(home, ".local", "share", "Trash"), CatTrash},
		)
	}

	return paths
}

// categorize returns the category for a filesystem entry encountered during
// traversal, or an empty string if the entry should not be categorized.
//
// suppressMacCats suppresses macOS-specific categories when --path points to
// a non-home path (e.g. /var), per PRD §5.1.2.
func categorize(path string, entry fs.DirEntry, goos, home string, suppressMacCats bool) string {
	if !entry.IsDir() {
		return categorizeFile(path, entry, goos, home)
	}

	name := entry.Name()

	// Universal pattern matches applied regardless of --path.
	if name == "node_modules" {
		return CatNodeModules
	}
	if name == "__pycache__" {
		return CatPycache
	}
	// .git directories: report as git_objects so clean can run git gc.
	if name == ".git" {
		return CatGitObjects
	}

	// Known fixed-location paths — check by absolute path.
	// npm / yarn / pnpm / pip / uv — both platforms.
	switch path {
	case filepath.Join(home, ".npm"),
		filepath.Join(home, ".cache", "yarn"),
		filepath.Join(home, ".pnpm-store"):
		return CatNpmCache
	case filepath.Join(home, ".cache", "pip"),
		filepath.Join(home, ".cache", "uv"):
		return CatPipCache
	}

	// Platform-specific known paths.
	switch goos {
	case "darwin":
		if !suppressMacCats {
			switch path {
			case filepath.Join(home, "Library", "Caches", "Homebrew"):
				return CatBrewCache
			case filepath.Join(home, "Library", "Developer", "Xcode", "DerivedData"):
				return CatXcodeDerived
			case filepath.Join(home, ".Trash"):
				return CatTrash
			case filepath.Join(home, "Downloads"):
				return CatDownloads
			}
		}
	case "linux":
		switch path {
		case "/var/cache/apt":
			return CatAptCache
		case "/var/lib/snapd":
			return CatSnap
		case "/var/log/journal":
			return CatJournald
		case filepath.Join(home, ".local", "share", "Trash"):
			return CatTrash
		}
	}

	return ""
}

// categorizeFile returns the category for a non-directory entry, or empty string.
// Currently handles compressed log files.
func categorizeFile(path string, entry fs.DirEntry, goos, home string) string {
	name := entry.Name()
	if !strings.HasSuffix(name, ".gz") {
		return ""
	}
	dir := filepath.Dir(path)
	// /var/log/*.gz (both platforms)
	if dir == "/var/log" || strings.HasPrefix(dir, "/var/log/") {
		return CatLogs
	}
	// ~/Library/Logs/**/*.gz (macOS)
	if goos == "darwin" {
		logsRoot := filepath.Join(home, "Library", "Logs")
		if dir == logsRoot || strings.HasPrefix(dir, logsRoot+string(filepath.Separator)) {
			return CatLogs
		}
	}
	return ""
}

// ScanHeader returns the output header string for the given scan configuration.
// Format is "[OS / path]" when root is set, or the disk label for full scan.
func ScanHeader(root, goos string) string {
	osName := "Linux"
	if goos == "darwin" {
		osName = "macOS"
	}
	if root != "" {
		return "[" + osName + " / " + root + "]"
	}
	diskLabel := "/"
	if goos == "darwin" {
		diskLabel = "Macintosh HD"
	}
	return "[" + osName + " / " + diskLabel + "]"
}

// suppressMacCategories reports whether macOS-specific categories should be
// suppressed given the scan root and OS. Suppression applies when --path is
// set to a path outside the home directory on macOS.
func suppressMacCategories(root, home, goos string) bool {
	if root == "" || goos != "darwin" {
		return false
	}
	return !strings.HasPrefix(root, home)
}
