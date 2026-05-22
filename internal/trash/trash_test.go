package trash

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestUniqueDestReturnsOriginalWhenFree(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	dest := filepath.Join(dir, "sample.txt")

	got := uniqueDest(dest)
	if got != dest {
		t.Fatalf("uniqueDest() = %q, want %q", got, dest)
	}
}

func TestUniqueDestAddsCounterWhenOccupied(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	dest := filepath.Join(dir, "sample.txt")
	if err := os.WriteFile(dest, []byte("x"), 0600); err != nil {
		t.Fatalf("seed occupied file: %v", err)
	}

	got := uniqueDest(dest)
	want := filepath.Join(dir, "sample 1.txt")
	if got != want {
		t.Fatalf("uniqueDest() = %q, want %q", got, want)
	}
}

func TestUniqueDestBeyondThousandCollisions(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	dest := filepath.Join(dir, "__pycache__")

	if err := os.Mkdir(dest, 0700); err != nil {
		t.Fatalf("seed occupied dir: %v", err)
	}

	for i := 1; i <= 1000; i++ {
		p := filepath.Join(dir, fmt.Sprintf("__pycache__ %d", i))
		if err := os.Mkdir(p, 0700); err != nil {
			t.Fatalf("seed collision %d: %v", i, err)
		}
	}

	got := uniqueDest(dest)
	want := filepath.Join(dir, "__pycache__ 1001")
	if got != want {
		t.Fatalf("uniqueDest() = %q, want %q", got, want)
	}
}
