package scan

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"sync/atomic"
)

// Config drives a single scan run.
type Config struct {
	Root      string // empty means scan from home; set by --path
	Deep      bool
	StaleDays int
	Workers   int
}

// Result is returned by Scan.
type Result struct {
	Items        []CandidateItem
	SkippedCount int    // paths skipped due to permission errors
	Header       string // e.g. "[macOS / Macintosh HD]"
	ScanMode     string // "quick" | "deep" | "path"
	ScanPath     string // empty for full scan, set to --path value
}

// Scan performs a scan and returns all discovered CandidateItems.
// The scan honours ctx cancellation at directory-entry granularity.
func Scan(ctx context.Context, cfg Config) (*Result, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("locate home directory: %w", err)
	}

	goos := runtime.GOOS
	suppress := suppressMacCategories(cfg.Root, home, goos)

	w := &walker{
		sem:             make(chan struct{}, cfg.Workers),
		home:            home,
		goos:            goos,
		staleDays:       cfg.StaleDays,
		suppressMacCats: suppress,
	}

	var (
		scanMode string
		scanPath string
		roots    []string
	)

	switch {
	case cfg.Root != "":
		scanMode = "path"
		scanPath = cfg.Root
		abs, err := filepath.Abs(cfg.Root)
		if err != nil {
			return nil, fmt.Errorf("resolve scan path %q: %w", cfg.Root, err)
		}
		roots = []string{abs}
		w.explicitRoot = true

	case cfg.Deep:
		scanMode = "deep"
		roots = deepScanRoots(home, goos)

	default:
		scanMode = "quick"
		items, skipped, err := w.quickScan(ctx, home)
		if err != nil {
			return nil, err
		}
		return &Result{
			Items:        items,
			SkippedCount: skipped,
			Header:       ScanHeader("", goos),
			ScanMode:     scanMode,
		}, nil
	}

	items, skipped, err := w.deepScan(ctx, roots)
	if err != nil {
		return nil, err
	}
	return &Result{
		Items:        items,
		SkippedCount: skipped,
		Header:       ScanHeader(scanPath, goos),
		ScanMode:     scanMode,
		ScanPath:     scanPath,
	}, nil
}

// deepScanRoots returns the set of root paths to traverse for a full deep scan.
func deepScanRoots(home, goos string) []string {
	roots := []string{home}
	if goos == "linux" {
		// Check system-level paths that live outside home.
		for _, p := range []string{"/var/cache/apt", "/var/lib/snapd", "/var/log"} {
			if _, err := os.Lstat(p); err == nil {
				roots = append(roots, p)
			}
		}
	}
	return roots
}

// walker holds the shared state for a traversal.
type walker struct {
	sem             chan struct{} // semaphore: capacity = workers
	home            string
	goos            string
	staleDays       int
	suppressMacCats bool
	// explicitRoot is set when the caller provided a specific Root path.
	// In that mode, volatile-path skip checks (alwaysSkip, network mounts)
	// are bypassed so the scanner fully traverses the requested directory.
	// The blocklist (isBlocklistedHome) is always enforced regardless.
	explicitRoot bool
}

// quickScan checks each well-known category path directly without traversal.
// This is the default scan mode; target: < 1 second.
func (w *walker) quickScan(ctx context.Context, home string) ([]CandidateItem, int, error) {
	paths := KnownCategoryPaths(home, w.goos)

	type sizeResult struct {
		item CandidateItem
		err  error
	}

	ch := make(chan sizeResult, len(paths))
	var wg sync.WaitGroup

	for _, kp := range paths {
		if isBlocklistedHome(kp.Path, home) {
			continue
		}
		info, err := os.Lstat(kp.Path)
		if err != nil || !info.IsDir() {
			continue // path doesn't exist on this system
		}
		wg.Add(1)
		go func(kp knownPath) {
			defer wg.Done()
			sz, err := dirSize(ctx, kp.Path)
			if err != nil && ctx.Err() != nil {
				ch <- sizeResult{err: ctx.Err()}
				return
			}
			if sz == 0 {
				return
			}
			score, source, lastMod := lastUsedScore(kp.Path, w.staleDays)
			ch <- sizeResult{item: CandidateItem{
				Path:            kp.Path,
				SizeBytes:       sz,
				Category:        kp.Category,
				StalenessScore:  score,
				StalenessSource: source,
				LastModified:    lastMod,
				Count:           1,
				OldestDays:      oldestDays(lastMod),
			}}
		}(kp)
	}

	go func() {
		wg.Wait()
		close(ch)
	}()

	var items []CandidateItem
	for r := range ch {
		if r.err != nil {
			return items, 0, r.err
		}
		if r.item.Path != "" {
			items = append(items, r.item)
		}
	}
	return items, 0, nil
}

// deepScan traverses roots using a bounded worker pool and returns all
// categorised items. The pool uses a semaphore with an inline fallback:
// if all worker slots are occupied, subdirectories are processed on the
// calling goroutine rather than blocking for a free slot. This prevents
// the classic all-workers-waiting-for-a-slot deadlock.
func (w *walker) deepScan(ctx context.Context, roots []string) ([]CandidateItem, int, error) {
	resultsCh := make(chan CandidateItem, 500)
	var skipped int64

	var wg sync.WaitGroup
	for _, root := range roots {
		if !w.explicitRoot && shouldSkip(root, w.home) {
			continue
		}
		wg.Add(1)
		w.sem <- struct{}{} // acquire initial slot
		go func(r string) {
			defer func() { <-w.sem }()
			w.walk(ctx, &wg, r, resultsCh, &skipped)
		}(root)
	}

	// Close resultsCh once all walkers finish so the collector below can drain.
	go func() {
		wg.Wait()
		close(resultsCh)
	}()

	var items []CandidateItem
	for item := range resultsCh {
		items = append(items, item)
		if ctx.Err() != nil {
			break
		}
	}

	// Drain any remaining items after context cancellation.
	for range resultsCh {
	}

	if ctx.Err() != nil {
		return items, int(atomic.LoadInt64(&skipped)), ctx.Err()
	}
	return items, int(atomic.LoadInt64(&skipped)), nil
}

// walk processes a single directory. It sends categorised CandidateItems to
// resultsCh and spawns goroutines for uncategorised subdirectories. When the
// semaphore is saturated, subdirectories are processed inline on this goroutine
// instead of blocking for a free slot.
func (w *walker) walk(
	ctx context.Context,
	wg *sync.WaitGroup,
	path string,
	resultsCh chan<- CandidateItem,
	skipped *int64,
) {
	defer wg.Done()

	if ctx.Err() != nil {
		return
	}

	entries, err := os.ReadDir(path)
	if err != nil {
		if os.IsPermission(err) {
			atomic.AddInt64(skipped, 1)
		}
		return
	}

	for _, entry := range entries {
		select {
		case <-ctx.Done():
			return
		default:
		}

		entryPath := filepath.Join(path, entry.Name())

		// os.ReadDir uses Lstat: symlinks appear with ModeSymlink set.
		// Never follow symlinks during traversal.
		if entry.Type()&os.ModeSymlink != 0 {
			continue
		}

		if isBlocklistedHome(entryPath, w.home) {
			continue
		}

		cat := categorize(entryPath, entry, w.goos, w.home, w.suppressMacCats)
		if cat != "" {
			sz, err := dirSize(ctx, entryPath)
			if err != nil && ctx.Err() != nil {
				return
			}
			if sz == 0 && !entry.IsDir() {
				info, err2 := entry.Info()
				if err2 == nil {
					sz = info.Size()
				}
			}
			score, source, lastMod := lastUsedScore(entryPath, w.staleDays)
			select {
			case resultsCh <- CandidateItem{
				Path:            entryPath,
				SizeBytes:       sz,
				Category:        cat,
				StalenessScore:  score,
				StalenessSource: source,
				LastModified:    lastMod,
				Count:           1,
				OldestDays:      oldestDays(lastMod),
			}:
			case <-ctx.Done():
				return
			}
			// Do not recurse into a categorised directory.
			continue
		}

		if !entry.IsDir() {
			continue
		}

		if !w.explicitRoot && shouldSkip(entryPath, w.home) {
			continue
		}

		// Spawn a goroutine if a worker slot is available; otherwise process
		// inline to avoid blocking (deadlock prevention).
		wg.Add(1)
		select {
		case w.sem <- struct{}{}:
			go func(p string) {
				defer func() { <-w.sem }()
				w.walk(ctx, wg, p, resultsCh, skipped)
			}(entryPath)
		default:
			// All worker slots occupied: process inline on this goroutine.
			w.walk(ctx, wg, entryPath, resultsCh, skipped)
		}
	}
}
