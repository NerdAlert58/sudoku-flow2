package solver_test

import (
	"testing"

	"github.com/NerdAlert58/sudoku-flow2/solver"
)

func TestSolve_LoopInvariants_SolvedSeeds(t *testing.T) {
	for i, seed := range originalSeeds(t) {
		res := solver.Solve(mustParse(t, seed))
		if res.Status != "solved" {
			t.Fatalf("seed %d: Status = %q, want solved", i, res.Status)
		}
		if res.Iterations != res.EventCount {
			t.Fatalf("seed %d: Iterations = %d, EventCount = %d; ADR-0007 requires equality for solved", i, res.Iterations, res.EventCount)
		}
		if res.EventCount != len(res.Events) {
			t.Fatalf("seed %d: EventCount = %d, len(Events) = %d; want equal", i, res.EventCount, len(res.Events))
		}
		if len(res.Events) == 0 {
			t.Fatalf("seed %d: no events for an incomplete solved seed", i)
		}
		for j, ev := range res.Events {
			if ev.Seq != j+1 {
				t.Fatalf("seed %d event index %d: Seq = %d, want %d", i, j, ev.Seq, j+1)
			}
			if ev.Placement == nil {
				t.Fatalf("seed %d event %d: Placement = nil; singles events must place", i, j)
			}
			if len(ev.Eliminations) != 0 {
				t.Fatalf("seed %d event %d: Eliminations non-empty on a singles event", i, j)
			}
			if len(ev.WitnessCells) == 0 {
				t.Fatalf("seed %d event %d: WitnessCells empty", i, j)
			}
			if len(ev.GridAfter) != 81 {
				t.Fatalf("seed %d event %d: len(GridAfter) = %d, want 81", i, j, len(ev.GridAfter))
			}
		}
		if last := res.Events[len(res.Events)-1]; last.GridAfter != res.Solution.String() {
			t.Fatalf("seed %d: last GridAfter = %q, want Solution.String() = %q", i, last.GridAfter, res.Solution.String())
		}
	}
}

// Permanent stall anchor (re-anchored by F-05 as sanctioned): four empty
// cells forming a two-solution {2,5} rectangle at r4/r5 x c1/c7. No ladder
// technique 1-13 fires (the x-wing shape on digits 2 and 5 has zero
// elimination targets, so the productive-event rule skips it), and no
// logic-only technique ever can - both completions are valid, so nothing is
// deducible. Verified against the F-05 scratch reference: full-ladder stall
// with zero events, >=2 brute-force solutions.
const beyondLadderGrid = "762314589819576234534928176476152398103869407908437601681293745247685913395741862"

func TestSolve_Stalled_BeyondLadder(t *testing.T) {
	res := solver.Solve(mustParse(t, beyondLadderGrid))
	if res.Status != "stalled" {
		t.Fatalf("Status = %q, want stalled (HARD seed needs beyond-current-ladder techniques)", res.Status)
	}
	if res.Grade != "" {
		t.Fatalf("Grade = %q, want empty for stalled", res.Grade)
	}
	if res.Iterations != res.EventCount+1 {
		t.Fatalf("Iterations = %d, EventCount = %d; ADR-0007 requires Iterations == EventCount+1 for stalled", res.Iterations, res.EventCount)
	}
}
