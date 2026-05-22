package scan

import (
	"os"
	"path/filepath"
	"time"
)

// sentinels are key project files whose mtime serves as a proxy for
// last-used time when atime is disabled. They are checked in order;
// the first existing sentinel wins.
var sentinels = []string{
	"package.json",
	filepath.Join(".git", "COMMIT_EDITMSG"),
	"go.mod",
	"Cargo.toml",
	"requirements.txt",
}

// lastUsedScore computes the staleness level and source for the directory at
// path using the fallback hierarchy from PRD §5.2.2:
//
//  1. atime if atime != mtime (indicates atime is active)
//  2. mtime of a sentinel file inside the directory
//  3. mtime of the directory itself
//  4. Unknown — never auto-selected for clean
//
// The returned StalenessLevel, source string, and time are frozen at scan
// time and stored in CandidateItem. The clean command must not re-derive them.
func lastUsedScore(path string, staleDays int) (StalenessLevel, string, time.Time) {
	info, err := os.Lstat(path)
	if err != nil {
		return StalenessUnknown, "unknown", time.Time{}
	}

	mtime := info.ModTime()
	atime := accessTime(info)

	// Step 1: atime is reliable only when it differs from mtime.
	// When atime == mtime the filesystem likely has noatime or relatime.
	if !atime.IsZero() && !atime.Equal(mtime) {
		return scoreFromAge(time.Since(atime), staleDays), "atime", atime
	}

	// Step 2: mtime of a sentinel file inside the directory.
	for _, sentinel := range sentinels {
		si, err := os.Lstat(filepath.Join(path, sentinel))
		if err != nil {
			continue
		}
		t := si.ModTime()
		return scoreFromAge(time.Since(t), staleDays), "sentinel_mtime", t
	}

	// Step 3: directory mtime itself (coarse — reflects installs, not project work).
	if !mtime.IsZero() {
		return scoreFromAge(time.Since(mtime), staleDays), "dir_mtime", mtime
	}

	return StalenessUnknown, "unknown", time.Time{}
}

// scoreFromAge maps an age duration to a StalenessLevel using DISKWHY_STALE_DAYS
// as the threshold between Stale and Unused.
func scoreFromAge(age time.Duration, staleDays int) StalenessLevel {
	days := int(age.Hours() / 24)
	switch {
	case days < 7:
		return StalenessActive
	case days < 30:
		return StalenessRecent
	case days < staleDays:
		return StalenessStale
	default:
		return StalenessUnused
	}
}

// oldestDays returns the age in days of t relative to now.
func oldestDays(t time.Time) int {
	if t.IsZero() {
		return 0
	}
	return int(time.Since(t).Hours() / 24)
}
