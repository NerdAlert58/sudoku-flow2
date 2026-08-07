package solver_test

import (
	"testing"

	"github.com/NerdAlert58/sudoku-flow2/oracle"
	"github.com/NerdAlert58/sudoku-flow2/solver"
)

// AC-1: on the 25 ORIGINAL seeds the brute-force oracle and the ladder solver
// agree cell-for-cell. (MEDIUM+ solution agreement rides AC-2's final-grid
// check in replay_test.go.)
func TestOracleSolve_OriginalSeeds_AgreeWithSolver(t *testing.T) {
	for i, seed := range originalSeeds(t) {
		g := mustParse(t, seed)
		res := solver.Solve(g)
		if res.Status != "solved" {
			t.Fatalf("seed %d: Status = %q, want solved", i, res.Status)
		}
		sol, count := oracle.Solve(g)
		if count != 1 {
			t.Errorf("seed %d: oracle count = %d, want 1", i, count)
		}
		if sol != res.Solution {
			t.Errorf("seed %d: oracle solution %s != solver solution %s", i, sol.String(), res.Solution.String())
		}
	}
}

// AC-1: every corpus puzzle has exactly one solution (2-capped count).
func TestOracleSolve_FullCorpus_UniqueSolution(t *testing.T) {
	for si, sec := range readCorpusSections(t) {
		for pi, seed := range sec {
			if _, count := oracle.Solve(mustParse(t, seed)); count != 1 {
				t.Errorf("section %d puzzle %d (%s): oracle count = %d, want 1", si, pi, seed, count)
			}
		}
	}
}
