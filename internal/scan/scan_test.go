package scan

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestDeepScan_NodeModules(t *testing.T) {
	root := t.TempDir()

	nm := filepath.Join(root, "project", "node_modules", "some-pkg")
	if err := os.MkdirAll(nm, 0755); err != nil {
		t.Fatal(err)
	}
	f, _ := os.CreateTemp(nm, "*.js")
	f.Write(make([]byte, 512))
	f.Close()

	cfg := Config{Root: root, Deep: true, StaleDays: 90, Workers: 2}
	result, err := Scan(context.Background(), cfg)
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}

	var found bool
	for _, item := range result.Items {
		if item.Category == CatNodeModules {
			found = true
			if item.SizeBytes == 0 {
				t.Error("node_modules SizeBytes should be > 0")
			}
			if item.StalenessScore == StalenessUnknown {
				t.Error("staleness should not be unknown for fresh dir")
			}
		}
	}
	if !found {
		t.Error("expected node_modules category in scan results")
	}
}

func TestDeepScan_Pycache(t *testing.T) {
	root := t.TempDir()

	pc := filepath.Join(root, "src", "__pycache__")
	if err := os.MkdirAll(pc, 0755); err != nil {
		t.Fatal(err)
	}
	f, _ := os.CreateTemp(pc, "*.pyc")
	f.Write(make([]byte, 256))
	f.Close()

	cfg := Config{Root: root, Deep: true, StaleDays: 90, Workers: 2}
	result, err := Scan(context.Background(), cfg)
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}

	var found bool
	for _, item := range result.Items {
		if item.Category == CatPycache {
			found = true
		}
	}
	if !found {
		t.Error("expected __pycache__ in scan results")
	}
}

func TestDeepScan_GitObjects(t *testing.T) {
	root := t.TempDir()

	gitDir := filepath.Join(root, "repo", ".git")
	if err := os.MkdirAll(gitDir, 0755); err != nil {
		t.Fatal(err)
	}
	f, _ := os.CreateTemp(gitDir, "*.pack")
	f.Write(make([]byte, 1024))
	f.Close()

	cfg := Config{Root: root, Deep: true, StaleDays: 90, Workers: 2}
	result, err := Scan(context.Background(), cfg)
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}

	var found bool
	for _, item := range result.Items {
		if item.Category == CatGitObjects {
			found = true
		}
	}
	if !found {
		t.Error("expected git_objects in scan results")
	}
}

func TestScan_ScanMode_Path(t *testing.T) {
	root := t.TempDir()
	result, err := Scan(context.Background(), Config{Root: root, Deep: false, StaleDays: 90, Workers: 2})
	if err != nil {
		t.Fatal(err)
	}
	// When Root is set, mode is always "path".
	if result.ScanMode != "path" {
		t.Errorf("scan mode with Root = %q, want 'path'", result.ScanMode)
	}
}

func TestScan_ScanMode_Quick(t *testing.T) {
	result, err := Scan(context.Background(), Config{Root: "", Deep: false, StaleDays: 90, Workers: 2})
	if err != nil {
		t.Fatal(err)
	}
	if result.ScanMode != "quick" {
		t.Errorf("quick scan mode = %q, want 'quick'", result.ScanMode)
	}
}

func TestScan_ContextCancel(t *testing.T) {
	root := t.TempDir()
	// Create enough depth to ensure the walker has something to cancel on.
	for i := 0; i < 5; i++ {
		d := filepath.Join(root, "a", "b", "c", "d", "e")
		os.MkdirAll(d, 0755)
		f, _ := os.CreateTemp(d, "*.bin")
		f.Close()
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := Scan(ctx, Config{Root: root, Deep: true, StaleDays: 90, Workers: 2})
	if err == nil {
		t.Error("expected error from cancelled context")
	}
}

func TestScan_Header(t *testing.T) {
	root := t.TempDir()
	result, err := Scan(context.Background(), Config{Root: root, Deep: false, StaleDays: 90, Workers: 2})
	if err != nil {
		t.Fatal(err)
	}
	if result.Header == "" {
		t.Error("header should not be empty")
	}
}

func TestShouldSkip_Blocklisted(t *testing.T) {
	home := t.TempDir()
	if !shouldSkip("/proc", home) {
		t.Error("/proc should be skipped")
	}
	if !shouldSkip("/sys", home) {
		t.Error("/sys should be skipped")
	}
}

func TestDeepScanRoots(t *testing.T) {
	home := t.TempDir()
	roots := deepScanRoots(home, runtime.GOOS)
	if len(roots) == 0 {
		t.Error("deepScanRoots should return at least the home dir")
	}
	// Home should always be in the roots.
	found := false
	for _, r := range roots {
		if r == home {
			found = true
		}
	}
	if !found {
		t.Errorf("home dir %q not found in deepScanRoots: %v", home, roots)
	}
}

func TestScan_StalenessScoreFrozen(t *testing.T) {
	root := t.TempDir()
	nm := filepath.Join(root, "node_modules")
	os.MkdirAll(nm, 0755)

	result, err := Scan(context.Background(), Config{Root: root, Deep: true, StaleDays: 90, Workers: 2})
	if err != nil {
		t.Fatal(err)
	}
	for _, item := range result.Items {
		// Verify the staleness source field is set for all items.
		if item.StalenessSource == "" {
			t.Errorf("item %q has empty StalenessSource", item.Path)
		}
	}
}
