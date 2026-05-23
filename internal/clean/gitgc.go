package clean

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

// runGitGC runs "git gc --prune=now" in the repository that owns dotGitPath.
// dotGitPath must point to a .git directory; gc runs in its parent.
//
// Space freed is measured as the delta of .git/objects byte-size before and
// after GC (PRD §5.5.6). If the delta is zero or negative — the repo was
// already tidy — 0 is returned with no error.
//
// A hard timeout (default 30 s) is applied. --aggressive is never used.
func runGitGC(ctx context.Context, dotGitPath string, timeoutSecs int) (int64, error) {
	repoRoot := filepath.Dir(dotGitPath)
	objectsPath := filepath.Join(dotGitPath, "objects")

	before := gitObjectsBytes(objectsPath)

	gcCtx, cancel := context.WithTimeout(ctx, time.Duration(timeoutSecs)*time.Second)
	defer cancel()

	cmd := exec.CommandContext(gcCtx, "git", "-C", repoRoot, "gc", "--prune=now", "--quiet")
	out, err := cmd.CombinedOutput()
	if err != nil {
		if gcCtx.Err() != nil {
			return 0, fmt.Errorf("git gc timed out after %ds in %q\nFix: repository may be very large; run git gc manually", timeoutSecs, repoRoot)
		}
		return 0, fmt.Errorf("git gc failed in %q: %s\nFix: check that git is installed and the repository is not corrupt", repoRoot, out)
	}

	after := gitObjectsBytes(objectsPath)
	freed := before - after
	if freed < 0 {
		freed = 0
	}
	return freed, nil
}

// gitObjectsBytes returns the total byte-size of all entries under path using
// os.ReadDir so it is consistent with the scan engine's size measurement.
func gitObjectsBytes(path string) int64 {
	entries, err := os.ReadDir(path)
	if err != nil {
		return 0
	}
	var total int64
	for _, e := range entries {
		info, err := e.Info()
		if err != nil {
			continue
		}
		if e.IsDir() {
			total += gitObjectsBytes(filepath.Join(path, e.Name()))
		} else {
			total += info.Size()
		}
	}
	return total
}
