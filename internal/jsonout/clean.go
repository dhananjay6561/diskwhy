package jsonout

import (
	"encoding/json"
	"io"
	"time"

	"github.com/dhananjay6561/diskwhy/internal/clean"
)

// CleanOutput is the schema_version:1 envelope for diskwhy clean --json.
type CleanOutput struct {
	SchemaVersion    int         `json:"schema_version"`
	CleanedAt        time.Time   `json:"cleaned_at"`
	Mode             string      `json:"mode"`    // "trash" | "delete"
	DryRun           bool        `json:"dry_run"`
	Results          []CleanItem `json:"items"`
	DockerFreedBytes int64       `json:"docker_freed_bytes"`
	Summary          CleanSummary `json:"summary"`
}

// CleanItem records the outcome for one path.
type CleanItem struct {
	Path         string `json:"path"`
	Category     string `json:"category"`
	Status       string `json:"status"` // trashed | deleted | gc_run | skipped | partial | error | dry_run
	BytesDelta   int64  `json:"bytes_freed"`
	FilesRemoved *int   `json:"files_removed,omitempty"` // populated for status=partial
	FilesTotal   *int   `json:"files_total,omitempty"`   // populated for status=partial
	Error        string `json:"error,omitempty"`
}

// CleanSummary aggregates totals from a clean run.
type CleanSummary struct {
	TotalItems      int   `json:"total_items"`
	Succeeded       int   `json:"succeeded"`
	Partial         int   `json:"partial"`
	Failed          int   `json:"failed"`
	Skipped         int   `json:"skipped"`
	TotalFreedBytes int64 `json:"bytes_freed"`
	PartialFailure  bool  `json:"partial_failure"` // true when any item is partial or error
}

// WriteClean marshals a completed clean run to w as indented JSON.
func WriteClean(
	w io.Writer,
	results []clean.ItemResult,
	dockerFreedBytes int64,
	dryRun bool,
	useTrash bool,
) error {
	mode := "delete"
	if useTrash {
		mode = "trash"
	}

	items := make([]CleanItem, 0, len(results))
	var totalFreed int64
	var succeeded, partial, failed, skipped int

	for _, r := range results {
		ci := CleanItem{
			Path:       r.Path,
			Category:   r.Category,
			Status:     outcomeString(r.Outcome),
			BytesDelta: r.BytesDelta,
		}
		if r.Err != nil {
			ci.Error = r.Err.Error()
		}
		if r.Outcome == clean.OutcomePartial {
			ci.FilesRemoved = &r.FilesRemoved
			ci.FilesTotal = &r.FilesTotal
		}

		switch r.Outcome {
		case clean.OutcomeDeleted, clean.OutcomeTrashed, clean.OutcomeGCRun:
			succeeded++
			totalFreed += r.BytesDelta
		case clean.OutcomePartial:
			partial++
		case clean.OutcomeError:
			failed++
		case clean.OutcomeSkipped, clean.OutcomeDryRun:
			skipped++
		}
		items = append(items, ci)
	}
	totalFreed += dockerFreedBytes

	out := CleanOutput{
		SchemaVersion:    1,
		CleanedAt:        time.Now().UTC(),
		Mode:             mode,
		DryRun:           dryRun,
		Results:          items,
		DockerFreedBytes: dockerFreedBytes,
		Summary: CleanSummary{
			TotalItems:      len(results),
			Succeeded:       succeeded,
			Partial:         partial,
			Failed:          failed,
			Skipped:         skipped,
			TotalFreedBytes: totalFreed,
			PartialFailure:  partial > 0 || failed > 0,
		},
	}

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(out)
}

func outcomeString(o clean.Outcome) string {
	switch o {
	case clean.OutcomeDryRun:
		return "dry_run"
	case clean.OutcomeSkipped:
		return "skipped"
	case clean.OutcomeTrashed:
		return "trashed"
	case clean.OutcomeDeleted:
		return "deleted"
	case clean.OutcomeGCRun:
		return "gc_run"
	case clean.OutcomePartial:
		return "partial"
	case clean.OutcomeError:
		return "error"
	default:
		return "unknown"
	}
}
