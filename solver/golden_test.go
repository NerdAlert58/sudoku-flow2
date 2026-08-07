package solver_test

import (
	"bytes"
	"os"
	"testing"

	"github.com/NerdAlert58/sudoku-flow2/solver"
)

// ORIGINAL #1 solves with singles alone (frozen corpus invariant), so ladder
// extensions must never alter its event log: the golden file mechanically
// anchors the canonical scan order (F-04 AC-4).
func TestSolve_GoldenEventLog_Original1(t *testing.T) {
	golden, err := os.ReadFile("testdata/mid/golden_original1_events.json")
	if err != nil {
		t.Fatalf("read golden: %v", err)
	}
	res := solver.Solve(mustParse(t, originalSeeds(t)[0]))
	if res.Status != "solved" {
		t.Fatalf("Status = %q, want solved", res.Status)
	}
	got := eventsJSON(t, res.Events)
	if !bytes.Equal(got, golden) {
		t.Fatalf("ORIGINAL #1 event log deviates from golden (%d bytes, want %d): canonical scan order or singles behavior changed", len(got), len(golden))
	}
}
