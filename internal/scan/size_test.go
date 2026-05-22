package scan

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestDirSize_BasicFiles(t *testing.T) {
	dir := t.TempDir()

	for i := 0; i < 3; i++ {
		f, err := os.CreateTemp(dir, "diskwhy_test_*.bin")
		if err != nil {
			t.Fatal(err)
		}
		if _, err := f.Write(make([]byte, 1024)); err != nil {
			t.Fatal(err)
		}
		f.Close()
	}

	size, err := dirSize(context.Background(), dir)
	if err != nil {
		t.Fatalf("dirSize: %v", err)
	}
	if size < 3*1024 {
		t.Errorf("dirSize = %d, want >= %d", size, 3*1024)
	}
}

func TestDirSize_Nested(t *testing.T) {
	dir := t.TempDir()
	sub := filepath.Join(dir, "sub")
	if err := os.Mkdir(sub, 0755); err != nil {
		t.Fatal(err)
	}

	for _, d := range []string{dir, sub} {
		f, _ := os.CreateTemp(d, "*.bin")
		f.Write(make([]byte, 512))
		f.Close()
	}

	size, err := dirSize(context.Background(), dir)
	if err != nil {
		t.Fatalf("dirSize: %v", err)
	}
	if size < 2*512 {
		t.Errorf("dirSize with nested = %d, want >= %d", size, 2*512)
	}
}

func TestDirSize_ContextCancel(t *testing.T) {
	dir := t.TempDir()
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	_, err := dirSize(ctx, dir)
	if err == nil {
		t.Error("expected error on cancelled context, got nil")
	}
}

func TestDirSize_Empty(t *testing.T) {
	dir := t.TempDir()
	size, err := dirSize(context.Background(), dir)
	if err != nil {
		t.Fatalf("dirSize on empty dir: %v", err)
	}
	if size != 0 {
		t.Errorf("empty dir size = %d, want 0", size)
	}
}
