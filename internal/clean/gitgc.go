package clean

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"time"
)

// runGitGC runs "git gc --prune=now" in the repository that owns dotGitPath.
// dotGitPath must point to a .git directory; gc runs in its parent.
// A hard 30-second timeout is applied regardless of the context deadline.
// --aggressive is never used per PRD §5.6.
func runGitGC(ctx context.Context, dotGitPath string, timeoutSecs int) error {
	repoRoot := filepath.Dir(dotGitPath)

	gcCtx, cancel := context.WithTimeout(ctx, time.Duration(timeoutSecs)*time.Second)
	defer cancel()

	cmd := exec.CommandContext(gcCtx, "git", "-C", repoRoot, "gc", "--prune=now", "--quiet")
	out, err := cmd.CombinedOutput()
	if err != nil {
		if gcCtx.Err() != nil {
			return fmt.Errorf("git gc timed out after %ds in %q\nFix: the repository may be very large; run git gc manually", timeoutSecs, repoRoot)
		}
		return fmt.Errorf("git gc failed in %q: %s\nFix: check that git is installed and the repository is not corrupt", repoRoot, out)
	}
	return nil
}
