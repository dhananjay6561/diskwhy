package jsonout

import (
	"encoding/json"
	"io"
	"time"

	"github.com/dhananjay6561/diskwhy/internal/docker"
	"github.com/dhananjay6561/diskwhy/internal/scan"
)

// ScanOutput is the schema_version:1 envelope for diskwhy scan --json.
type ScanOutput struct {
	SchemaVersion int         `json:"schema_version"`
	ScannedAt     time.Time   `json:"scanned_at"`
	ScanMode      string      `json:"scan_mode"`
	Header        string      `json:"header"`
	Disk          DiskInfo    `json:"disk"`
	Items         []ScanItem  `json:"items"`
	Docker        *DockerInfo `json:"docker,omitempty"`
	Summary       ScanSummary `json:"summary"`
}

// DiskInfo mirrors tui.DiskUsage output.
type DiskInfo struct {
	TotalBytes int64 `json:"total_bytes"`
	UsedBytes  int64 `json:"used_bytes"`
	FreeBytes  int64 `json:"free_bytes"`
}

// ScanItem is one categorised path from scan.CandidateItem.
type ScanItem struct {
	Path            string    `json:"path"`
	Category        string    `json:"category"`
	SizeBytes       int64     `json:"size_bytes"`
	Staleness       string    `json:"staleness"`
	StalenessSource string    `json:"staleness_source"`
	LastModified    time.Time `json:"last_modified"`
	Count           int       `json:"count"`
	OldestDays      int       `json:"oldest_days"`
}

// DockerInfo mirrors docker.Result.
type DockerInfo struct {
	SocketPath       string `json:"socket_path"`
	UnusedImageBytes int64  `json:"unused_image_bytes"`
	UsedImageBytes   int64  `json:"used_image_bytes"`
	VolumeBytes      int64  `json:"volume_bytes"`
	UnusedImageCount int    `json:"unused_image_count"`
	UsedImageCount   int    `json:"used_image_count"`
	VolumeCount      int    `json:"volume_count"`
}

// ScanSummary aggregates totals from a completed scan.
type ScanSummary struct {
	TotalFoundBytes  int64 `json:"total_found_bytes"`
	SafeToCleanBytes int64 `json:"safe_to_clean_bytes"`
	SkippedCount     int   `json:"skipped_count"`
	ElapsedMs        int64 `json:"elapsed_ms"`
}

// WriteScan marshals a completed scan and optional docker result to w as
// indented JSON. The caller is responsible for closing w.
func WriteScan(
	w io.Writer,
	result *scan.Result,
	dockerResult *docker.Result,
	disk DiskInfo,
	elapsedMs int64,
) error {
	items := make([]ScanItem, 0, len(result.Items))
	var totalBytes, safeBytes int64

	for _, it := range result.Items {
		items = append(items, ScanItem{
			Path:            it.Path,
			Category:        it.Category,
			SizeBytes:       it.SizeBytes,
			Staleness:       it.StalenessScore.String(),
			StalenessSource: it.StalenessSource,
			LastModified:    it.LastModified,
			Count:           it.Count,
			OldestDays:      it.OldestDays,
		})
		totalBytes += it.SizeBytes
		if it.StalenessScore == scan.StalenessStale || it.StalenessScore == scan.StalenessUnused {
			safeBytes += it.SizeBytes
		}
	}

	var dInfo *DockerInfo
	if dockerResult != nil && (dockerResult.UnusedImageBytes+dockerResult.UsedImageBytes+dockerResult.VolumeBytes) > 0 {
		dInfo = &DockerInfo{
			SocketPath:       dockerResult.SocketPath,
			UnusedImageBytes: dockerResult.UnusedImageBytes,
			UsedImageBytes:   dockerResult.UsedImageBytes,
			VolumeBytes:      dockerResult.VolumeBytes,
			UnusedImageCount: dockerResult.UnusedImageCount,
			UsedImageCount:   dockerResult.UsedImageCount,
			VolumeCount:      dockerResult.VolumeCount,
		}
		safeBytes += dockerResult.UnusedImageBytes
	}

	out := ScanOutput{
		SchemaVersion: 1,
		ScannedAt:     time.Now().UTC(),
		ScanMode:      result.ScanMode,
		Header:        result.Header,
		Disk:          disk,
		Items:         items,
		Docker:        dInfo,
		Summary: ScanSummary{
			TotalFoundBytes:  totalBytes,
			SafeToCleanBytes: safeBytes,
			SkippedCount:     result.SkippedCount,
			ElapsedMs:        elapsedMs,
		},
	}

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(out)
}
