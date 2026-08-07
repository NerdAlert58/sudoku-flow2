package solver_test

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/NerdAlert58/sudoku-flow2/solver"
)

// AC-2: the ladder-cap harness itself is under test here.

func TestSolveCapped_RegistryCoverage(t *testing.T) {
	if n := solver.LadderSize(); n != 13 {
		t.Fatalf("LadderSize() = %d, want 13", n)
	}
	if len(perTechnique) != solver.LadderSize() {
		t.Fatalf("perTechnique covers %d positions, want %d", len(perTechnique), solver.LadderSize())
	}
	for i, tc := range perTechnique {
		if tc.pos != i+1 {
			t.Errorf("perTechnique[%d].pos = %d, want %d", i, tc.pos, i+1)
		}
	}
}

func TestSolveCapped_CapZero_StallsEveryIncompleteState(t *testing.T) {
	grids := []string{originalSeeds(t)[0]}
	for _, tc := range perTechnique {
		grids = append(grids, tc.fires)
	}
	for _, grid := range grids {
		res := solver.SolveCapped(mustParse(t, grid), 0)
		if res.Status != "stalled" {
			t.Errorf("%.12s...: Status = %q, want stalled with every technique disabled", grid, res.Status)
		}
		if res.EventCount != 0 || len(res.Events) != 0 {
			t.Errorf("%.12s...: EventCount = %d, want 0", grid, res.EventCount)
		}
		if res.Iterations != 1 {
			t.Errorf("%.12s...: Iterations = %d, want 1 (one pass, nothing fires)", grid, res.Iterations)
		}
		if res.Solution.String() != grid {
			t.Errorf("%.12s...: Solution mutated with every technique disabled", grid)
		}
	}
}

// A complete grid solves at cap 0 exactly as Solve does (ADR-0014: zero
// passes run, so the cap never matters).
func TestSolveCapped_CapZero_CompleteGridMatchesSolve(t *testing.T) {
	g := mustParse(t, completeGrid)
	want := resultJSON(t, solver.Solve(g))
	got := resultJSON(t, solver.SolveCapped(g, 0))
	if !bytes.Equal(want, got) {
		t.Fatalf("SolveCapped(complete, 0) = %s, want Solve's result %s", got, want)
	}
}

// Capping at the full ladder reproduces Solve byte-for-byte on five corpus
// seeds spanning all four sections.
func TestSolveCapped_FullCap_ByteIdenticalToSolve(t *testing.T) {
	secs := readCorpusSections(t)
	if len(secs) < 4 {
		t.Fatalf("corpus has %d sections, want 4", len(secs))
	}
	seeds := []string{secs[0][0], secs[1][0], secs[2][0], secs[3][0], secs[3][9]}
	for i, seed := range seeds {
		g := mustParse(t, seed)
		want := resultJSON(t, solver.Solve(g))
		got := resultJSON(t, solver.SolveCapped(g, solver.LadderSize()))
		if !bytes.Equal(want, got) {
			t.Errorf("seed %d: SolveCapped(., %d) differs from Solve", i, solver.LadderSize())
		}
	}
}

func resultJSON(t *testing.T, res solver.SolveResult) []byte {
	t.Helper()
	b, err := json.Marshal(res)
	if err != nil {
		t.Fatalf("marshal SolveResult: %v", err)
	}
	return b
}
