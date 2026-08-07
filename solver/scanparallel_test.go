package solver_test

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/NerdAlert58/sudoku-flow2/solver"
)

func scanParallelEquivalenceGrids(t *testing.T) []namedGrid {
	t.Helper()
	var grids []namedGrid
	total := 0
	for si, sec := range readCorpusSections(t) {
		for pi, seed := range sec {
			grids = append(grids, namedGrid{fmt.Sprintf("corpus_s%d_p%02d", si, pi), seed})
			total++
		}
	}
	if total != 55 {
		t.Fatalf("corpus has %d seeds, want 55", total)
	}
	for _, fx := range upperFixtures {
		grids = append(grids, namedGrid{"fixture_" + fx.technique + "_" + fx.grid[:8], fx.grid})
	}
	grids = append(grids,
		namedGrid{"fixture_finned_only", gridFinnedOnly},
		namedGrid{"fixture_k2_late_k3_early", gridK2LateK3Early},
		namedGrid{"fixture_beyond_ladder", beyondLadderGrid},
	)
	return grids
}

func TestSolveScanParallel_EquivalentToSolve_AllButCandidateChecks(t *testing.T) {
	for _, ng := range scanParallelEquivalenceGrids(t) {
		g := mustParse(t, ng.grid)
		want := solver.Solve(g)
		got := solver.SolveScanParallel(g)
		if got.Status != want.Status {
			t.Errorf("%s: Status = %q, want %q", ng.name, got.Status, want.Status)
		}
		if got.Solution.String() != want.Solution.String() {
			t.Errorf("%s: Solution = %s, want %s", ng.name, got.Solution.String(), want.Solution.String())
		}
		if !bytes.Equal(eventsJSON(t, got.Events), eventsJSON(t, want.Events)) {
			t.Errorf("%s: events JSON differs from sequential Solve", ng.name)
		}
		if got.Iterations != want.Iterations {
			t.Errorf("%s: Iterations = %d, want %d", ng.name, got.Iterations, want.Iterations)
		}
		if got.EventCount != want.EventCount {
			t.Errorf("%s: EventCount = %d, want %d", ng.name, got.EventCount, want.EventCount)
		}
		if got.Grade != want.Grade {
			t.Errorf("%s: Grade = %q, want %q", ng.name, got.Grade, want.Grade)
		}
	}
}
