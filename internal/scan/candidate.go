package scan

import "time"

// StalenessLevel is the frozen freshness classification assigned at scan time.
// The clean command receives this value from CandidateItem and never re-derives
// it from disk.
type StalenessLevel int

const (
	StalenessUnknown StalenessLevel = iota // no reliable signal
	StalenessActive                        // < 7 days
	StalenessRecent                        // 7-30 days
	StalenessStale                         // 30-DISKWHY_STALE_DAYS days
	StalenessUnused                        // > DISKWHY_STALE_DAYS days
)

func (s StalenessLevel) String() string {
	switch s {
	case StalenessActive:
		return "active"
	case StalenessRecent:
		return "recent"
	case StalenessStale:
		return "stale"
	case StalenessUnused:
		return "unused"
	default:
		return "unknown"
	}
}

// CandidateItem is the contract between the scan phase and the clean phase.
// StalenessScore is computed once at scan time. The clean command must use
// this frozen value and never re-read mtime from disk to re-derive it.
type CandidateItem struct {
	Path            string
	SizeBytes       int64
	Category        string
	StalenessScore  StalenessLevel
	StalenessSource string    // "atime" | "sentinel_mtime" | "dir_mtime" | "unknown"
	LastModified    time.Time
	Count           int // number of matched sub-items (repos, projects, etc.)
	OldestDays      int // age of the oldest signal in days
}
