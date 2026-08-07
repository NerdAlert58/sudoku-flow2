package solver_test

import (
	"bytes"
	"testing"

	"github.com/NerdAlert58/sudoku-flow2/solver"
)

// AC-1 exit bar: with the full 13-technique ladder every corpus puzzle
// solves logic-only. (F-03's ORIGINAL-only singles test stays untouched;
// this test covers all four sections.)
func TestSolve_FullCorpus_AllSolved(t *testing.T) {
	for si, sec := range readCorpusSections(t) {
		for pi, seed := range sec {
			res := solver.Solve(mustParse(t, seed))
			if res.Status != "solved" {
				t.Errorf("section %d puzzle %d: Status = %q, want solved (grid %s)", si, pi, res.Status, seed)
				continue
			}
			assertRuleConformantSolution(t, seed, res.Solution.String())
		}
	}
}

// AC-6: determinism double-run over all 55 corpus seeds.
func TestSolve_FullCorpus_DeterminismDoubleRun(t *testing.T) {
	for si, sec := range readCorpusSections(t) {
		for pi, seed := range sec {
			first := solver.Solve(mustParse(t, seed))
			second := solver.Solve(mustParse(t, seed))
			if !bytes.Equal(eventsJSON(t, first.Events), eventsJSON(t, second.Events)) {
				t.Fatalf("section %d puzzle %d: events JSON differs between two consecutive in-process solves", si, pi)
			}
			if first.Iterations != second.Iterations ||
				first.EventCount != second.EventCount ||
				first.CandidateChecks != second.CandidateChecks {
				t.Fatalf("section %d puzzle %d: counters differ: run1 (%d %d %d) run2 (%d %d %d)",
					si, pi, first.Iterations, first.EventCount, first.CandidateChecks,
					second.Iterations, second.EventCount, second.CandidateChecks)
			}
		}
	}
}

// F-04 verifier debt: unproductive-event bugs must fail crisply, not hang.
// ADR-0007 fixes Iterations == EventCount for solved and EventCount+1
// otherwise, so Iterations <= EventCount+1 always; 2000 is a generous
// absolute ceiling (a full solve needs at most 81 placements plus bounded
// eliminations).
func TestSolve_IterationCap_Invariant(t *testing.T) {
	grids := []string{gridFinnedOnly, gridK2LateK3Early, beyondLadderGrid}
	for _, fx := range upperFixtures {
		grids = append(grids, fx.grid)
	}
	for _, sec := range readCorpusSections(t) {
		grids = append(grids, sec...)
	}
	for _, g := range grids {
		res := solver.Solve(mustParse(t, g))
		if res.Iterations > res.EventCount+1 {
			t.Errorf("grid %s: Iterations = %d > EventCount+1 = %d (unproductive pass)", g, res.Iterations, res.EventCount+1)
		}
		if res.Iterations > 2000 {
			t.Errorf("grid %s: Iterations = %d exceeds the absolute cap 2000", g, res.Iterations)
		}
	}
}
