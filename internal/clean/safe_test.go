package clean

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/dhananjay6561/diskwhy/internal/scan"
)

func TestIsSafeToDelete(t *testing.T) {
	home := "/home/user"
	cases := []struct {
		name string
		item scan.CandidateItem
		want bool
	}{
		{
			"cache always safe regardless of staleness",
			scan.CandidateItem{
				Path: filepath.Join(home, ".cache", "pip"), Category: scan.CatPipCache,
				StalenessScore: scan.StalenessActive,
			},
			true,
		},
		{
			"brew cache always safe",
			scan.CandidateItem{
				Path: filepath.Join(home, "Library", "Caches", "Homebrew"), Category: scan.CatBrewCache,
				StalenessScore: scan.StalenessActive,
			},
			true,
		},
		{
			"pycache always safe",
			scan.CandidateItem{
				Path: filepath.Join(home, "project", "__pycache__"), Category: scan.CatPycache,
				StalenessScore: scan.StalenessActive,
			},
			true,
		},
		{
			"node_modules active not safe",
			scan.CandidateItem{
				Path: filepath.Join(home, "project", "node_modules"), Category: scan.CatNodeModules,
				StalenessScore: scan.StalenessActive,
			},
			false,
		},
		{
			"node_modules stale safe",
			scan.CandidateItem{
				Path: filepath.Join(home, "project", "node_modules"), Category: scan.CatNodeModules,
				StalenessScore: scan.StalenessStale,
			},
			true,
		},
		{
			"node_modules unused safe",
			scan.CandidateItem{
				Path: filepath.Join(home, "project", "node_modules"), Category: scan.CatNodeModules,
				StalenessScore: scan.StalenessUnused,
			},
			true,
		},
		{
			"node_modules recent not safe",
			scan.CandidateItem{
				Path: filepath.Join(home, "project", "node_modules"), Category: scan.CatNodeModules,
				StalenessScore: scan.StalenessRecent,
			},
			false,
		},
		{
			"blocklisted path always blocked",
			scan.CandidateItem{
				Path: "/proc/net", Category: scan.CatLogs,
				StalenessScore: scan.StalenessUnused,
			},
			false,
		},
		{
			"ssh dir always blocked",
			scan.CandidateItem{
				Path: filepath.Join(home, ".ssh"), Category: scan.CatNodeModules,
				StalenessScore: scan.StalenessUnused,
			},
			false,
		},
		{
			"gnupg dir always blocked",
			scan.CandidateItem{
				Path: filepath.Join(home, ".gnupg"), Category: scan.CatPycache,
				StalenessScore: scan.StalenessUnused,
			},
			false,
		},
		{
			"git objects not deletable (handled by git gc)",
			scan.CandidateItem{
				Path: filepath.Join(home, "project", ".git"), Category: scan.CatGitObjects,
				StalenessScore: scan.StalenessUnused,
			},
			false,
		},
		{
			"docker not deletable (handled by SDK)",
			scan.CandidateItem{
				Path: "/var/lib/docker", Category: scan.CatDocker,
				StalenessScore: scan.StalenessUnused,
			},
			false,
		},
		{
			"downloads unused safe",
			scan.CandidateItem{
				Path: filepath.Join(home, "Downloads"), Category: scan.CatDownloads,
				StalenessScore: scan.StalenessUnused,
			},
			true,
		},
		{
			"downloads active not safe",
			scan.CandidateItem{
				Path: filepath.Join(home, "Downloads"), Category: scan.CatDownloads,
				StalenessScore: scan.StalenessActive,
			},
			false,
		},
		{
			"trash always safe",
			scan.CandidateItem{
				Path: filepath.Join(home, ".local", "share", "Trash"), Category: scan.CatTrash,
				StalenessScore: scan.StalenessActive,
			},
			true,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := isSafeToDelete(c.item, home)
			if got != c.want {
				t.Errorf("isSafeToDelete = %v, want %v", got, c.want)
			}
		})
	}
}

func TestSafeRemove_PermanentDelete(t *testing.T) {
	dir := t.TempDir()
	f, err := os.CreateTemp(dir, "diskwhy_test_*.bin")
	if err != nil {
		t.Fatal(err)
	}
	path := f.Name()
	f.Close()

	if err := safeRemove(path, false); err != nil {
		t.Fatalf("safeRemove: %v", err)
	}
	if _, err := os.Lstat(path); !os.IsNotExist(err) {
		t.Error("file should not exist after safeRemove with permanent delete")
	}
}

func TestSafeRemove_Directory(t *testing.T) {
	dir := t.TempDir()
	sub := filepath.Join(dir, "to_delete")
	if err := os.Mkdir(sub, 0755); err != nil {
		t.Fatal(err)
	}
	nested, _ := os.CreateTemp(sub, "*.bin")
	nested.Close()

	if err := safeRemove(sub, false); err != nil {
		t.Fatalf("safeRemove directory: %v", err)
	}
	if _, err := os.Lstat(sub); !os.IsNotExist(err) {
		t.Error("directory should not exist after safeRemove")
	}
}
