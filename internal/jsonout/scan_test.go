package jsonout

import (
	"bytes"
	"encoding/json"
	"testing"
	"time"

	"github.com/dhananjay6561/diskwhy/internal/scan"
)

func TestWriteScan_Basic(t *testing.T) {
	result := &scan.Result{
		Items: []scan.CandidateItem{
			{
				Path:            "/home/user/.npm",
				Category:        scan.CatNpmCache,
				SizeBytes:       1 << 30,
				StalenessScore:  scan.StalenessUnused,
				StalenessSource: "dir_mtime",
				LastModified:    time.Now().Add(-100 * 24 * time.Hour),
				Count:           1,
				OldestDays:      100,
			},
		},
		SkippedCount: 2,
		Header:       "[Linux / /]",
		ScanMode:     "quick",
	}
	disk := DiskInfo{TotalBytes: 100 << 30, UsedBytes: 50 << 30, FreeBytes: 50 << 30}

	var buf bytes.Buffer
	if err := WriteScan(&buf, result, nil, disk, 150); err != nil {
		t.Fatalf("WriteScan: %v", err)
	}

	var out ScanOutput
	if err := json.Unmarshal(buf.Bytes(), &out); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if out.SchemaVersion != 1 {
		t.Errorf("schema_version = %d, want 1", out.SchemaVersion)
	}
	if out.ScanMode != "quick" {
		t.Errorf("scan_mode = %q, want %q", out.ScanMode, "quick")
	}
	if len(out.Items) != 1 {
		t.Fatalf("items len = %d, want 1", len(out.Items))
	}
	if out.Items[0].Category != scan.CatNpmCache {
		t.Errorf("item category = %q, want %q", out.Items[0].Category, scan.CatNpmCache)
	}
	if out.Items[0].Staleness != "unused" {
		t.Errorf("staleness = %q, want 'unused'", out.Items[0].Staleness)
	}
	if out.Summary.SafeToCleanBytes != 1<<30 {
		t.Errorf("safe_to_clean_bytes = %d, want %d", out.Summary.SafeToCleanBytes, 1<<30)
	}
	if out.Summary.SkippedCount != 2 {
		t.Errorf("skipped_count = %d, want 2", out.Summary.SkippedCount)
	}
	if out.Summary.ElapsedMs != 150 {
		t.Errorf("elapsed_ms = %d, want 150", out.Summary.ElapsedMs)
	}
	if out.Docker != nil {
		t.Error("docker should be nil when no docker result provided")
	}
	if out.Disk.TotalBytes != 100<<30 {
		t.Errorf("disk.total_bytes = %d, want %d", out.Disk.TotalBytes, 100<<30)
	}
}

func TestWriteScan_StaleCountsInSafe(t *testing.T) {
	result := &scan.Result{
		Items: []scan.CandidateItem{
			{SizeBytes: 1 << 30, StalenessScore: scan.StalenessStale},
			{SizeBytes: 2 << 30, StalenessScore: scan.StalenessUnused},
			{SizeBytes: 3 << 30, StalenessScore: scan.StalenessActive},
		},
	}

	var buf bytes.Buffer
	if err := WriteScan(&buf, result, nil, DiskInfo{}, 0); err != nil {
		t.Fatal(err)
	}

	var out ScanOutput
	json.Unmarshal(buf.Bytes(), &out)

	want := int64((1 + 2) << 30)
	if out.Summary.SafeToCleanBytes != want {
		t.Errorf("safe_to_clean_bytes = %d, want %d", out.Summary.SafeToCleanBytes, want)
	}
}

func TestWriteScan_EmptyItems(t *testing.T) {
	result := &scan.Result{ScanMode: "deep", Header: "[Linux / /]"}
	var buf bytes.Buffer
	if err := WriteScan(&buf, result, nil, DiskInfo{}, 0); err != nil {
		t.Fatalf("WriteScan empty: %v", err)
	}
	var out ScanOutput
	if err := json.Unmarshal(buf.Bytes(), &out); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(out.Items) != 0 {
		t.Errorf("expected 0 items, got %d", len(out.Items))
	}
}
