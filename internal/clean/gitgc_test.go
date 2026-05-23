package clean

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestGitObjectsBytes_nonexistent(t *testing.T) {
	n := gitObjectsBytes("/nonexistent/path/objects")
	if n != 0 {
		t.Errorf("nonexistent path should return 0, got %d", n)
	}
}

func TestGitObjectsBytes_emptyDir(t *testing.T) {
	dir := t.TempDir()
	n := gitObjectsBytes(dir)
	if n != 0 {
		t.Errorf("empty dir should return 0, got %d", n)
	}
}

func TestGitObjectsBytes_withFiles(t *testing.T) {
	dir := t.TempDir()
	sub := filepath.Join(dir, "sub")
	if err := os.Mkdir(sub, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sub, "file.bin"), make([]byte, 1024), 0644); err != nil {
		t.Fatal(err)
	}
	n := gitObjectsBytes(dir)
	if n != 1024 {
		t.Errorf("gitObjectsBytes = %d, want 1024", n)
	}
}

func TestRunGitGC_notARepo(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not in PATH")
	}
	dir := t.TempDir()
	_, err := runGitGC(context.Background(), dir, 5)
	if err == nil {
		t.Error("expected error for non-git directory")
	}
}

func TestRunGitGC_realRepo(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not in PATH")
	}
	dir := t.TempDir()
	init := exec.Command("git", "-C", dir, "init")
	if err := init.Run(); err != nil {
		t.Fatalf("git init: %v", err)
	}
	dotGit := filepath.Join(dir, ".git")
	freed, err := runGitGC(context.Background(), dotGit, 15)
	if err != nil {
		t.Errorf("runGitGC on empty repo: %v", err)
	}
	if freed < 0 {
		t.Errorf("freed bytes should be >= 0, got %d", freed)
	}
}
