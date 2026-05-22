package scan

import (
	"io/fs"
	"path/filepath"
	"testing"
)

// dirEntry is a minimal fs.DirEntry for table-driven tests.
type dirEntry struct {
	name  string
	isDir bool
}

func (d dirEntry) Name() string               { return d.name }
func (d dirEntry) IsDir() bool                { return d.isDir }
func (d dirEntry) Type() fs.FileMode          { return 0 }
func (d dirEntry) Info() (fs.FileInfo, error) { return nil, nil }

func TestCategorize_UniversalPatterns(t *testing.T) {
	home := "/home/user"
	cases := []struct {
		path  string
		entry fs.DirEntry
		goos  string
		want  string
	}{
		{
			"/home/user/project/node_modules",
			dirEntry{"node_modules", true},
			"linux", CatNodeModules,
		},
		{
			"/home/user/project/__pycache__",
			dirEntry{"__pycache__", true},
			"linux", CatPycache,
		},
		{
			"/home/user/project/.git",
			dirEntry{".git", true},
			"linux", CatGitObjects,
		},
		// non-matching dir
		{
			"/home/user/project/src",
			dirEntry{"src", true},
			"linux", "",
		},
	}
	for _, c := range cases {
		got := categorize(c.path, c.entry, c.goos, home, false)
		if got != c.want {
			t.Errorf("categorize(%q) = %q, want %q", c.path, got, c.want)
		}
	}
}

func TestCategorize_KnownLinuxPaths(t *testing.T) {
	home := "/home/user"
	cases := []struct {
		path string
		want string
	}{
		{"/var/cache/apt", CatAptCache},
		{"/var/lib/snapd", CatSnap},
		{"/var/log/journal", CatJournald},
		{filepath.Join(home, ".local", "share", "Trash"), CatTrash},
		{filepath.Join(home, ".npm"), CatNpmCache},
		{filepath.Join(home, ".cache", "yarn"), CatNpmCache},
		{filepath.Join(home, ".pnpm-store"), CatNpmCache},
		{filepath.Join(home, ".cache", "pip"), CatPipCache},
		{filepath.Join(home, ".cache", "uv"), CatPipCache},
	}
	for _, c := range cases {
		name := filepath.Base(c.path)
		entry := dirEntry{name, true}
		got := categorize(c.path, entry, "linux", home, false)
		if got != c.want {
			t.Errorf("categorize(%q, linux) = %q, want %q", c.path, got, c.want)
		}
	}
}

func TestCategorize_KnownDarwinPaths(t *testing.T) {
	home := "/Users/user"
	cases := []struct {
		path string
		want string
	}{
		{filepath.Join(home, "Library", "Caches", "Homebrew"), CatBrewCache},
		{filepath.Join(home, "Library", "Developer", "Xcode", "DerivedData"), CatXcodeDerived},
		{filepath.Join(home, ".Trash"), CatTrash},
		{filepath.Join(home, "Downloads"), CatDownloads},
		{filepath.Join(home, ".npm"), CatNpmCache},
		{filepath.Join(home, ".cache", "pip"), CatPipCache},
	}
	for _, c := range cases {
		name := filepath.Base(c.path)
		entry := dirEntry{name, true}
		got := categorize(c.path, entry, "darwin", home, false)
		if got != c.want {
			t.Errorf("categorize(%q, darwin) = %q, want %q", c.path, got, c.want)
		}
	}
}

func TestCategorize_SuppressMacCats(t *testing.T) {
	home := "/Users/user"
	path := filepath.Join(home, "Library", "Caches", "Homebrew")
	entry := dirEntry{"Homebrew", true}
	got := categorize(path, entry, "darwin", home, true)
	if got != "" {
		t.Errorf("with suppressMacCats=true, brew cache should not be categorized, got %q", got)
	}
}

func TestCategorizeFile_GzLogs(t *testing.T) {
	home := "/home/user"
	cases := []struct {
		path string
		goos string
		want string
	}{
		{"/var/log/syslog.gz", "linux", CatLogs},
		{"/var/log/apt/history.gz", "linux", CatLogs},
		{filepath.Join(home, "Library", "Logs", "crash.gz"), "darwin", CatLogs},
		{filepath.Join(home, "file.gz"), "linux", ""},
		{"/var/log/syslog", "linux", ""},
	}
	for _, c := range cases {
		name := filepath.Base(c.path)
		entry := dirEntry{name, false}
		got := categorize(c.path, entry, c.goos, home, false)
		if got != c.want {
			t.Errorf("categorize(%q, %s) = %q, want %q", c.path, c.goos, got, c.want)
		}
	}
}

func TestScanHeader(t *testing.T) {
	cases := []struct {
		root string
		goos string
		want string
	}{
		{"", "darwin", "[macOS / Macintosh HD]"},
		{"", "linux", "[Linux / /]"},
		{"/var", "linux", "[Linux / /var]"},
		{"/Users/dj", "darwin", "[macOS / /Users/dj]"},
	}
	for _, c := range cases {
		got := ScanHeader(c.root, c.goos)
		if got != c.want {
			t.Errorf("ScanHeader(%q, %q) = %q, want %q", c.root, c.goos, got, c.want)
		}
	}
}

func TestSuppressMacCategories(t *testing.T) {
	home := "/Users/user"
	cases := []struct {
		root string
		goos string
		want bool
	}{
		{"", "darwin", false},
		{home, "darwin", false},
		{filepath.Join(home, "projects"), "darwin", false},
		{"/var", "darwin", true},
		{"/opt", "darwin", true},
		{"", "linux", false},
		{"/var", "linux", false},
	}
	for _, c := range cases {
		got := suppressMacCategories(c.root, home, c.goos)
		if got != c.want {
			t.Errorf("suppressMacCategories(%q, %q) = %v, want %v", c.root, c.goos, got, c.want)
		}
	}
}

func TestKnownCategoryPaths_Linux(t *testing.T) {
	home := "/home/user"
	paths := KnownCategoryPaths(home, "linux")
	m := make(map[string]string)
	for _, p := range paths {
		m[p.Path] = p.Category
	}
	if m["/var/cache/apt"] != CatAptCache {
		t.Error("missing apt cache path for linux")
	}
	if m[filepath.Join(home, ".npm")] != CatNpmCache {
		t.Error("missing npm path for linux")
	}
}

func TestKnownCategoryPaths_Darwin(t *testing.T) {
	home := "/Users/user"
	paths := KnownCategoryPaths(home, "darwin")
	m := make(map[string]string)
	for _, p := range paths {
		m[p.Path] = p.Category
	}
	if m[filepath.Join(home, "Library", "Caches", "Homebrew")] != CatBrewCache {
		t.Error("missing brew cache path for darwin")
	}
	if m[filepath.Join(home, ".Trash")] != CatTrash {
		t.Error("missing trash path for darwin")
	}
}
