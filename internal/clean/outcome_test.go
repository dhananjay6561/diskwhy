package clean

import "testing"

func TestOutcomeConstants(t *testing.T) {
	// Ensure constants are distinct and the 6-state machine is complete.
	outcomes := []Outcome{
		OutcomeDryRun,
		OutcomeSkipped,
		OutcomeTrashed,
		OutcomeDeleted,
		OutcomeGCRun,
		OutcomeError,
	}
	seen := make(map[Outcome]bool, len(outcomes))
	for _, o := range outcomes {
		if seen[o] {
			t.Errorf("duplicate outcome value %d", o)
		}
		seen[o] = true
	}
	if len(seen) != 6 {
		t.Errorf("expected 6 distinct outcomes, got %d", len(seen))
	}
}
