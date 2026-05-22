package clean

// THIS IS THE ONLY FILE IN THE REPOSITORY THAT CALLS os.Remove OR os.RemoveAll.
// All deletions — permanent or trash-based — funnel through safeRemove.

import (
	"fmt"
	"os"

	"github.com/dhananjay6561/diskwhy/internal/scan"
	"github.com/dhananjay6561/diskwhy/internal/trash"
)

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

// cacheCategories are always safe to delete — their contents are regenerable.
var cacheCategories = map[string]bool{
	scan.CatBrewCache:    true,
	scan.CatXcodeDerived: true,
	scan.CatPipCache:     true,
	scan.CatNpmCache:     true,
	scan.CatAptCache:     true,
	scan.CatPycache:      true,
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

// safeRemove deletes path. When useTrash is true the item is moved to the OS
// trash instead of being permanently deleted.
func safeRemove(path string, useTrash bool) error {
	if useTrash {
		if err := trash.Move(path); err != nil {
			return fmt.Errorf("move %q to trash: %w\nFix: run without --trash to delete permanently", path, err)
		}
		return nil
	}
	if err := os.RemoveAll(path); err != nil {
		return fmt.Errorf("delete %q: %w\nFix: check file permissions or run with sudo", path, err)
	}
	return nil
}
