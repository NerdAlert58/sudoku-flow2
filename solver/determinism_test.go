package solver_test

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/NerdAlert58/sudoku-flow2/solver"
)

func eventsJSON(t *testing.T, evs []solver.Event) []byte {
	t.Helper()
	b, err := json.Marshal(evs)
	if err != nil {
		t.Fatalf("marshal events: %v", err)
	}
	return b
}

func TestSolve_Determinism_DoubleRun(t *testing.T) {
	for i, seed := range originalSeeds(t) {
		first := solver.Solve(mustParse(t, seed))
		second := solver.Solve(mustParse(t, seed))
		if !bytes.Equal(eventsJSON(t, first.Events), eventsJSON(t, second.Events)) {
			t.Fatalf("seed %d: events JSON differs between two consecutive in-process solves", i)
		}
		if first.Iterations != second.Iterations ||
			first.EventCount != second.EventCount ||
			first.CandidateChecks != second.CandidateChecks {
			t.Fatalf("seed %d: counters differ: run1 (iterations=%d eventCount=%d candidateChecks=%d) run2 (iterations=%d eventCount=%d candidateChecks=%d)",
				i, first.Iterations, first.EventCount, first.CandidateChecks,
				second.Iterations, second.EventCount, second.CandidateChecks)
		}
	}
}
