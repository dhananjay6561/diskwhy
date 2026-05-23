package clean

import (
	"context"
	"errors"

	"github.com/dhananjay6561/diskwhy/internal/scan"
)

// Config controls a clean run.
type Config struct {
	Categories     []string // category names to clean; empty means nothing
	DryRun         bool
	UseTrash       bool
	GitTimeoutSecs int
	Home           string
}

// Run processes items one at a time, respecting context cancellation between
// each item. Items whose category is not in cfg.Categories are silently skipped.
func Run(ctx context.Context, cfg Config, items []scan.CandidateItem) []ItemResult {
	allowed := make(map[string]bool, len(cfg.Categories))
	for _, c := range cfg.Categories {
		allowed[c] = true
	}

	results := make([]ItemResult, 0, len(items))

	for _, item := range items {
		if ctx.Err() != nil {
			break
		}
		if !allowed[item.Category] {
			continue
		}

		if item.Category == scan.CatGitObjects {
			results = append(results, handleGitGC(ctx, item, cfg))
			continue
		}

		results = append(results, handleFileItem(ctx, item, cfg))
	}

	return results
}

func handleGitGC(ctx context.Context, item scan.CandidateItem, cfg Config) ItemResult {
	if cfg.DryRun {
		return ItemResult{Path: item.Path, Category: item.Category, Outcome: OutcomeDryRun}
	}
	if !isSafeToDelete(item, cfg.Home) {
		return ItemResult{Path: item.Path, Category: item.Category, Outcome: OutcomeSkipped}
	}
	secs := cfg.GitTimeoutSecs
	if secs <= 0 {
		secs = 30
	}
	freed, err := runGitGC(ctx, item.Path, secs)
	if err != nil {
		return ItemResult{Path: item.Path, Category: item.Category, Outcome: OutcomeError, Err: err}
	}
	return ItemResult{Path: item.Path, Category: item.Category, Outcome: OutcomeGCRun, BytesDelta: freed}
}

func handleFileItem(ctx context.Context, item scan.CandidateItem, cfg Config) ItemResult {
	if cfg.DryRun {
		return ItemResult{Path: item.Path, Category: item.Category, Outcome: OutcomeDryRun, BytesDelta: item.SizeBytes}
	}
	if !isSafeToDelete(item, cfg.Home) {
		return ItemResult{Path: item.Path, Category: item.Category, Outcome: OutcomeSkipped}
	}
	removed, total, err := safeRemove(ctx, item.Path, cfg.UseTrash)
	if err != nil {
		if errors.Is(err, ErrCancelled) {
			return ItemResult{
				Path:         item.Path,
				Category:     item.Category,
				Outcome:      OutcomePartial,
				FilesRemoved: removed,
				FilesTotal:   total,
				Err:          err,
			}
		}
		return ItemResult{Path: item.Path, Category: item.Category, Outcome: OutcomeError, Err: err}
	}
	outcome := OutcomeDeleted
	if cfg.UseTrash {
		outcome = OutcomeTrashed
	}
	return ItemResult{Path: item.Path, Category: item.Category, Outcome: outcome, BytesDelta: item.SizeBytes}
}
