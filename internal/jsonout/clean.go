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
	DryRun           bool        `json:"dry_run"`
	UseTrash         bool        `json:"use_trash"`
	Results          []CleanItem `json:"results"`
	DockerFreedBytes int64       `json:"docker_freed_bytes"`
	Summary          CleanSummary `json:"summary"`
}

// CleanItem records the outcome for one path.
type CleanItem struct {
	Path       string `json:"path"`
	Category   string `json:"category"`
	Outcome    string `json:"outcome"`
	BytesDelta int64  `json:"bytes_delta"`
	Error      string `json:"error,omitempty"`
}

// CleanSummary aggregates totals from a clean run.
type CleanSummary struct {
	TotalFreedBytes int64 `json:"total_freed_bytes"`
	ErrorCount      int   `json:"error_count"`
}

// WriteClean marshals a completed clean run to w as indented JSON.
func WriteClean(
	w io.Writer,
	results []clean.ItemResult,
	dockerFreedBytes int64,
	dryRun bool,
	useTrash bool,
) error {
	items := make([]CleanItem, 0, len(results))
	var totalFreed int64
	var errCount int

	for _, r := range results {
		ci := CleanItem{
			Path:       r.Path,
			Category:   r.Category,
			Outcome:    outcomeString(r.Outcome),
			BytesDelta: r.BytesDelta,
		}
		if r.Err != nil {
			ci.Error = r.Err.Error()
			errCount++
		}
		totalFreed += r.BytesDelta
		items = append(items, ci)
	}
	totalFreed += dockerFreedBytes

	out := CleanOutput{
		SchemaVersion:    1,
		CleanedAt:        time.Now().UTC(),
		DryRun:           dryRun,
		UseTrash:         useTrash,
		Results:          items,
		DockerFreedBytes: dockerFreedBytes,
		Summary: CleanSummary{
			TotalFreedBytes: totalFreed,
			ErrorCount:      errCount,
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
	case clean.OutcomeError:
		return "error"
	default:
		return "unknown"
	}
}
