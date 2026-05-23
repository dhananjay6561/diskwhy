package clean

// Outcome is one of the states a clean operation can produce per PRD §5.5.1.
type Outcome int

const (
	OutcomeDryRun  Outcome = iota // dry-run mode; item not touched
	OutcomeSkipped                // blocked by safety check or user choice
	OutcomeTrashed                // moved to OS trash
	OutcomeDeleted                // permanently removed
	OutcomeGCRun                  // git gc completed in repository
	OutcomePartial                // deletion started but cancelled mid-directory (SIGTERM)
	OutcomeError                  // operation failed
)

// ItemResult records what happened to one item during a clean run.
type ItemResult struct {
	Path         string
	Category     string
	Outcome      Outcome
	BytesDelta   int64 // bytes freed; 0 when moved to trash or after git gc
	FilesRemoved int   // set for OutcomePartial: files removed before cancellation
	FilesTotal   int   // set for OutcomePartial: total files counted before deletion started
	Err          error
}
