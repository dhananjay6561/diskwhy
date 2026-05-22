package jsonout

import (
	"bytes"
	"encoding/json"
	"errors"
	"testing"

	"github.com/dhananjay6561/diskwhy/internal/clean"
	"github.com/dhananjay6561/diskwhy/internal/scan"
)

func TestWriteClean_Basic(t *testing.T) {
	results := []clean.ItemResult{
		{Path: "/home/user/.npm", Category: scan.CatNpmCache, Outcome: clean.OutcomeDeleted, BytesDelta: 1 << 30},
		{Path: "/home/user/.ssh", Category: scan.CatNodeModules, Outcome: clean.OutcomeSkipped},
		{Path: "/home/user/repo/.git", Category: scan.CatGitObjects, Outcome: clean.OutcomeGCRun},
	}

	var buf bytes.Buffer
	if err := WriteClean(&buf, results, 500<<20, false, false); err != nil {
		t.Fatalf("WriteClean: %v", err)
	}

	var out CleanOutput
	if err := json.Unmarshal(buf.Bytes(), &out); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if out.SchemaVersion != 1 {
		t.Errorf("schema_version = %d, want 1", out.SchemaVersion)
	}
	if out.DryRun {
		t.Error("dry_run should be false")
	}
	if out.UseTrash {
		t.Error("use_trash should be false")
	}
	if len(out.Results) != 3 {
		t.Fatalf("results len = %d, want 3", len(out.Results))
	}
	if out.Results[0].Outcome != "deleted" {
		t.Errorf("outcome[0] = %q, want 'deleted'", out.Results[0].Outcome)
	}
	if out.Results[1].Outcome != "skipped" {
		t.Errorf("outcome[1] = %q, want 'skipped'", out.Results[1].Outcome)
	}
	if out.Results[2].Outcome != "gc_run" {
		t.Errorf("outcome[2] = %q, want 'gc_run'", out.Results[2].Outcome)
	}
	if out.DockerFreedBytes != 500<<20 {
		t.Errorf("docker_freed_bytes = %d, want %d", out.DockerFreedBytes, 500<<20)
	}
	want := int64(1<<30) + int64(500<<20)
	if out.Summary.TotalFreedBytes != want {
		t.Errorf("total_freed_bytes = %d, want %d", out.Summary.TotalFreedBytes, want)
	}
}

func TestWriteClean_AllOutcomes(t *testing.T) {
	outcomes := []struct {
		o    clean.Outcome
		want string
	}{
		{clean.OutcomeDryRun, "dry_run"},
		{clean.OutcomeSkipped, "skipped"},
		{clean.OutcomeTrashed, "trashed"},
		{clean.OutcomeDeleted, "deleted"},
		{clean.OutcomeGCRun, "gc_run"},
		{clean.OutcomeError, "error"},
	}
	for _, c := range outcomes {
		if got := outcomeString(c.o); got != c.want {
			t.Errorf("outcomeString(%d) = %q, want %q", c.o, got, c.want)
		}
	}
	// unknown sentinel
	if got := outcomeString(clean.Outcome(999)); got != "unknown" {
		t.Errorf("outcomeString(999) = %q, want 'unknown'", got)
	}
}

func TestWriteClean_DryRun(t *testing.T) {
	var buf bytes.Buffer
	if err := WriteClean(&buf, nil, 0, true, false); err != nil {
		t.Fatal(err)
	}
	var out CleanOutput
	json.Unmarshal(buf.Bytes(), &out)
	if !out.DryRun {
		t.Error("dry_run should be true")
	}
	if len(out.Results) != 0 {
		t.Errorf("nil results should produce empty array, got %d items", len(out.Results))
	}
}

func TestWriteClean_ErrorItem(t *testing.T) {
	results := []clean.ItemResult{
		{
			Path:     "/some/path",
			Category: scan.CatLogs,
			Outcome:  clean.OutcomeError,
			Err:      errors.New("permission denied"),
		},
	}
	var buf bytes.Buffer
	if err := WriteClean(&buf, results, 0, false, false); err != nil {
		t.Fatalf("WriteClean with error item: %v", err)
	}
	var out CleanOutput
	json.Unmarshal(buf.Bytes(), &out)
	if out.Results[0].Error == "" {
		t.Error("error field should be non-empty for OutcomeError items")
	}
	if out.Summary.ErrorCount != 1 {
		t.Errorf("error_count = %d, want 1", out.Summary.ErrorCount)
	}
}
