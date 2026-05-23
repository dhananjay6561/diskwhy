package clean

// THIS IS THE ONLY FILE IN THE REPOSITORY THAT CALLS os.Remove OR os.RemoveAll.
// All deletions — permanent or trash-based — funnel through safeRemove.

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/dhananjay6561/diskwhy/internal/scan"
	"github.com/dhananjay6561/diskwhy/internal/trash"
)

// ErrCancelled is returned by safeRemove when context cancellation stops a
// deletion mid-directory, leaving the path in a partially deleted state.
var ErrCancelled = errors.New("cancelled")

// deletableByCategory lists categories that can be removed via safeRemove.
// CatGitObjects and CatDocker are excluded — they are handled by dedicated
// code paths (git gc and docker SDK respectively).
var deletableByCategory = map[string]bool{
	scan.CatNodeModules:  true,
	scan.CatBrewCache:    true,
	scan.CatXcodeDerived: true,
	scan.CatPipCache:     true,
	scan.CatNpmCache:     true,
	scan.CatAptCache:     true,
	scan.CatLogs:         true,
	scan.CatJournald:     true,
	scan.CatTrash:        true,
	scan.CatPycache:      true,
	scan.CatDownloads:    true,
}

// cacheCategories are always safe to delete — their contents are regenerable
// or already discarded (trash).
var cacheCategories = map[string]bool{
	scan.CatBrewCache:    true,
	scan.CatXcodeDerived: true,
	scan.CatPipCache:     true,
	scan.CatNpmCache:     true,
	scan.CatAptCache:     true,
	scan.CatPycache:      true,
	scan.CatTrash:        true,
}

// isSafeToDelete returns true when item may be passed to safeRemove.
// It checks the blocklist, category allow-list, and (for non-cache items)
// requires that staleness is Stale or Unused so recent data is never touched.
func isSafeToDelete(item scan.CandidateItem, home string) bool {
	if scan.IsBlocklistedHome(item.Path, home) {
		return false
	}
	if !deletableByCategory[item.Category] {
		return false
	}
	if cacheCategories[item.Category] {
		return true
	}
	return item.StalenessScore == scan.StalenessStale || item.StalenessScore == scan.StalenessUnused
}

// safeRemove is the only function permitted to call os.Remove.
// Layer 2 (PRD §6.2): independently re-validates path against the blocklist
// and resolves symlinks before any syscall fires. Then deletes using a
// context-aware recursive loop so SIGTERM cancellation granularity is one file.
//
// Returns (filesRemoved, filesTotal, err). When err == ErrCancelled the
// deletion is partial: filesRemoved entries were removed before cancellation.
// useTrash moves the item to the OS trash instead of permanent deletion.
func safeRemove(ctx context.Context, path string, useTrash bool) (filesRemoved, filesTotal int, err error) {
	if !filepath.IsAbs(path) {
		return 0, 0, fmt.Errorf("safeRemove: refusing relative path %q", path)
	}
	// Layer 2: independent blocklist re-check (separate from isSafeToDelete).
	if scan.IsBlocklisted(path) {
		return 0, 0, fmt.Errorf("safeRemove: %q is in the system blocklist", path)
	}
	// Resolve symlink target and re-check blocklist to close the TOCTOU window.
	if target, err2 := filepath.EvalSymlinks(path); err2 == nil && target != path {
		if scan.IsBlocklisted(target) {
			return 0, 0, fmt.Errorf("safeRemove: symlink %q resolves to blocked path %q", path, target)
		}
	}

	if useTrash {
		if err := trash.Move(path); err != nil {
			return 0, 1, fmt.Errorf("move %q to trash: %w\nFix: run without --trash to delete permanently", path, err)
		}
		return 1, 1, nil
	}

	filesTotal = countEntries(path)
	filesRemoved, err = removeWithContext(ctx, path)
	return filesRemoved, filesTotal, err
}

// removeWithContext deletes path recursively, checking ctx.Done() between
// directory entries. The minimum cancellation unit is a single file.
// Returns (removed, ErrCancelled) when interrupted mid-directory.
func removeWithContext(ctx context.Context, path string) (int, error) {
	info, err := os.Lstat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, nil
		}
		return 0, err
	}
	if !info.IsDir() {
		if err := os.Remove(path); err != nil {
			return 0, err
		}
		return 1, nil
	}
	entries, err := os.ReadDir(path)
	if err != nil {
		return 0, err
	}
	removed := 0
	for _, e := range entries {
		select {
		case <-ctx.Done():
			return removed, ErrCancelled
		default:
		}
		n, ferr := removeWithContext(ctx, filepath.Join(path, e.Name()))
		removed += n
		if ferr != nil {
			return removed, ferr
		}
	}
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return removed, err
	}
	return removed + 1, nil
}

// countEntries counts total filesystem entries under path (including path itself).
// Used to compute filesTotal before deletion starts so partial reports are accurate.
func countEntries(path string) int {
	info, err := os.Lstat(path)
	if err != nil {
		return 1
	}
	if !info.IsDir() {
		return 1
	}
	entries, err := os.ReadDir(path)
	if err != nil {
		return 1
	}
	total := 1
	for _, e := range entries {
		total += countEntries(filepath.Join(path, e.Name()))
	}
	return total
}
