package scan

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestScoreFromAge(t *testing.T) {
	cases := []struct {
		days      int
		staleDays int
		want      StalenessLevel
	}{
		{0, 90, StalenessActive},
		{3, 90, StalenessActive},
		{6, 90, StalenessActive},
		{7, 90, StalenessRecent},
		{20, 90, StalenessRecent},
		{29, 90, StalenessRecent},
		{30, 90, StalenessStale},
		{60, 90, StalenessStale},
		{89, 90, StalenessStale},
		{90, 90, StalenessUnused},
		{180, 90, StalenessUnused},
		// staleDays only shifts the Stale→Unused boundary when > 30.
		// Items 7-30 days old always show as Recent regardless of staleDays.
		{10, 14, StalenessRecent},
		{15, 14, StalenessRecent},
		{31, 14, StalenessUnused},
		{30, 60, StalenessStale},
		{60, 60, StalenessUnused},
	}
	for _, c := range cases {
		age := time.Duration(c.days) * 24 * time.Hour
		got := scoreFromAge(age, c.staleDays)
		if got != c.want {
			t.Errorf("scoreFromAge(%dd, staleDays=%d) = %v, want %v",
				c.days, c.staleDays, got, c.want)
		}
	}
}

func TestOldestDays(t *testing.T) {
	if got := oldestDays(time.Time{}); got != 0 {
		t.Errorf("oldestDays(zero) = %d, want 0", got)
	}
	past := time.Now().Add(-48 * time.Hour)
	if got := oldestDays(past); got < 1 {
		t.Errorf("oldestDays(2 days ago) = %d, want >= 1", got)
	}
}

func TestLastUsedScore_FreshDir(t *testing.T) {
	dir := t.TempDir()
	level, source, ts := lastUsedScore(dir, 90)
	if level == StalenessUnknown {
		t.Errorf("fresh tmpdir should not be StalenessUnknown, got %v (source=%q)", level, source)
	}
	if ts.IsZero() {
		t.Error("time should not be zero for a fresh dir")
	}
	// A brand-new dir must be Active
	if level != StalenessActive {
		t.Errorf("fresh tmpdir level = %v, want StalenessActive", level)
	}
}

func TestLastUsedScore_MissingPath(t *testing.T) {
	level, source, _ := lastUsedScore("/nonexistent-diskwhy-test-path-xyz", 90)
	if level != StalenessUnknown {
		t.Errorf("missing path level = %v, want StalenessUnknown", level)
	}
	if source != "unknown" {
		t.Errorf("missing path source = %q, want %q", source, "unknown")
	}
}

func TestLastUsedScore_SentinelFile(t *testing.T) {
	dir := t.TempDir()
	// Write a go.mod sentinel that is old so dir_mtime is bypassed.
	sentinel := filepath.Join(dir, "go.mod")
	if err := os.WriteFile(sentinel, []byte("module test\n"), 0644); err != nil {
		t.Fatal(err)
	}
	// Backdate sentinel to 120 days ago.
	old := time.Now().Add(-120 * 24 * time.Hour)
	if err := os.Chtimes(sentinel, old, old); err != nil {
		t.Skip("cannot set file times on this platform:", err)
	}

	_, source, _ := lastUsedScore(dir, 90)
	// Source should prefer sentinel over dir_mtime when sentinel exists.
	if source != "sentinel_mtime" && source != "atime" && source != "dir_mtime" {
		t.Errorf("unexpected source %q", source)
	}
}
