package clean

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/dhananjay6561/diskwhy/internal/scan"
)

func TestRun_DryRun(t *testing.T) {
	dir := t.TempDir()
	nm := filepath.Join(dir, "node_modules")
	if err := os.MkdirAll(nm, 0755); err != nil {
		t.Fatal(err)
	}
	f, _ := os.CreateTemp(nm, "*.js")
	f.Write(make([]byte, 512))
	f.Close()

	items := []scan.CandidateItem{
		{
			Path:           nm,
			Category:       scan.CatNodeModules,
			SizeBytes:      512,
			StalenessScore: scan.StalenessUnused,
		},
	}

	cfg := Config{
		Categories: []string{scan.CatNodeModules},
		DryRun:     true,
		Home:       dir,
	}
	results := Run(context.Background(), cfg, items)

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Outcome != OutcomeDryRun {
		t.Errorf("outcome = %v, want OutcomeDryRun", results[0].Outcome)
	}
	// File must still exist in dry-run mode.
	if _, err := os.Lstat(nm); os.IsNotExist(err) {
		t.Error("dry-run should not delete the directory")
	}
}

func TestRun_Delete(t *testing.T) {
	dir := t.TempDir()
	nm := filepath.Join(dir, "node_modules")
	if err := os.MkdirAll(nm, 0755); err != nil {
		t.Fatal(err)
	}

	items := []scan.CandidateItem{
		{
			Path:           nm,
			Category:       scan.CatNodeModules,
			SizeBytes:      0,
			StalenessScore: scan.StalenessUnused,
		},
	}

	cfg := Config{
		Categories: []string{scan.CatNodeModules},
		DryRun:     false,
		UseTrash:   false,
		Home:       dir,
	}
	results := Run(context.Background(), cfg, items)

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Outcome != OutcomeDeleted {
		t.Errorf("outcome = %v, want OutcomeDeleted", results[0].Outcome)
	}
	if _, err := os.Lstat(nm); !os.IsNotExist(err) {
		t.Error("directory should be deleted after clean")
	}
}

func TestRun_SkipActiveItem(t *testing.T) {
	dir := t.TempDir()
	nm := filepath.Join(dir, "node_modules")
	os.MkdirAll(nm, 0755)

	items := []scan.CandidateItem{
		{
			Path:           nm,
			Category:       scan.CatNodeModules,
			SizeBytes:      0,
			StalenessScore: scan.StalenessActive,
		},
	}

	cfg := Config{
		Categories: []string{scan.CatNodeModules},
		DryRun:     false,
		Home:       dir,
	}
	results := Run(context.Background(), cfg, items)

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Outcome != OutcomeSkipped {
		t.Errorf("active item outcome = %v, want OutcomeSkipped", results[0].Outcome)
	}
}

func TestRun_CategoryFilter(t *testing.T) {
	dir := t.TempDir()
	items := []scan.CandidateItem{
		{Path: filepath.Join(dir, "a"), Category: scan.CatNodeModules, StalenessScore: scan.StalenessUnused},
		{Path: filepath.Join(dir, "b"), Category: scan.CatPycache, StalenessScore: scan.StalenessUnused},
	}

	cfg := Config{
		Categories: []string{scan.CatPycache},
		DryRun:     true,
		Home:       dir,
	}
	results := Run(context.Background(), cfg, items)

	// Only pycache item should produce a result.
	if len(results) != 1 {
		t.Errorf("expected 1 result (pycache only), got %d", len(results))
	}
}

func TestRun_ContextCancel(t *testing.T) {
	dir := t.TempDir()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	items := []scan.CandidateItem{
		{Path: filepath.Join(dir, "a"), Category: scan.CatPycache, StalenessScore: scan.StalenessUnused},
		{Path: filepath.Join(dir, "b"), Category: scan.CatPycache, StalenessScore: scan.StalenessUnused},
	}

	cfg := Config{Categories: []string{scan.CatPycache}, DryRun: true, Home: dir}
	results := Run(ctx, cfg, items)
	// Should stop early due to context cancellation.
	if len(results) > 0 {
		t.Errorf("cancelled context should produce 0 results, got %d", len(results))
	}
}

func TestRun_CacheAlwaysDeleted(t *testing.T) {
	dir := t.TempDir()
	pc := filepath.Join(dir, "__pycache__")
	os.MkdirAll(pc, 0755)

	items := []scan.CandidateItem{
		{
			Path:           pc,
			Category:       scan.CatPycache,
			SizeBytes:      0,
			StalenessScore: scan.StalenessActive, // active, but cache so safe
		},
	}
	cfg := Config{Categories: []string{scan.CatPycache}, DryRun: false, Home: dir}
	results := Run(context.Background(), cfg, items)

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Outcome != OutcomeDeleted {
		t.Errorf("cache outcome = %v, want OutcomeDeleted", results[0].Outcome)
	}
}
