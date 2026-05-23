package scan

import (
	"path/filepath"
	"testing"
)

func TestIsBlocklisted(t *testing.T) {
	cases := []struct {
		path string
		want bool
	}{
		{"/proc", true},
		{"/proc/net/tcp", true},
		{"/sys", true},
		{"/sys/kernel/mm", true},
		{"/dev", true},
		{"/dev/null", true},
		{"/System", true},
		{"/System/Library", true},
		{"/bin", true},
		{"/sbin", true},
		{"/usr/bin", true},
		{"/usr/bin/git", true},
		{"/usr/sbin", true},
		{"/etc", true},
		{"/etc/hosts", true},
		{"/boot", true},
		{"/lib", true},
		{"/lib64", true},
		{"/usr/lib", true},
		{"/.Spotlight-V100", true},
		// not blocklisted
		{"/home", false},
		{"/home/user", false},
		{"/usr/local", false},
		{"/usr/local/bin", false},
		{"/var", false},
		{"/opt", false},
		{"/tmp", false},
	}
	for _, c := range cases {
		got := isBlocklisted(c.path)
		if got != c.want {
			t.Errorf("isBlocklisted(%q) = %v, want %v", c.path, got, c.want)
		}
	}
}

func TestIsBlocklistedHome(t *testing.T) {
	home := "/home/user"
	cases := []struct {
		path string
		want bool
	}{
		// system blocklist inherited
		{"/proc", true},
		{"/etc/hosts", true},
		// home-relative credential paths
		{filepath.Join(home, ".ssh"), true},
		{filepath.Join(home, ".ssh", "id_rsa"), true},
		{filepath.Join(home, ".gnupg"), true},
		{filepath.Join(home, ".gnupg", "private-keys-v1.d"), true},
		// safe paths not blocked
		{filepath.Join(home, ".config"), false},
		{filepath.Join(home, "Downloads"), false},
		{filepath.Join(home, ".cache"), false},
		{"/home/other/.ssh", false}, // different user's ssh, not blocked by home check
	}
	for _, c := range cases {
		got := isBlocklistedHome(c.path, home)
		if got != c.want {
			t.Errorf("isBlocklistedHome(%q) = %v, want %v", c.path, got, c.want)
		}
	}
}

func TestIsBlocklistedExport(t *testing.T) {
	if !IsBlocklisted("/proc/net") {
		t.Error("exported IsBlocklisted should block /proc/net")
	}
	if IsBlocklisted("/home/user/Downloads") {
		t.Error("exported IsBlocklisted should not block ~/Downloads")
	}
}

func TestIsBlocklistedHomeExport(t *testing.T) {
	if got := IsBlocklistedHome("/proc", "/home/u"); !got {
		t.Error("exported IsBlocklistedHome should block /proc")
	}
	if got := IsBlocklistedHome("/home/u/Downloads", "/home/u"); got {
		t.Error("exported IsBlocklistedHome should not block ~/Downloads")
	}
}
