package solver_test

import (
	"testing"

	"github.com/NerdAlert58/sudoku-flow2/oracle"
	"github.com/NerdAlert58/sudoku-flow2/solver"
)

// AC-2 corpus driver — eval row "UC-2 Replay proof": every event of every
// corpus solve passes the ADR-0013 verifier. The eval row's 20-puzzle
// generated slice is F-08 AC-6's burden, not exercised here.
func TestReplayVerify_FullCorpus_AllEventsPass(t *testing.T) {
	for si, sec := range readCorpusSections(t) {
		for pi, seed := range sec {
			g := mustParse(t, seed)
			res := solver.Solve(g)
			if res.Status != "solved" {
				t.Fatalf("section %d puzzle %d: Status = %q, want solved", si, pi, res.Status)
			}
			if err := oracle.ReplayVerify(g, res); err != nil {
				t.Errorf("section %d puzzle %d (%s): ReplayVerify: %v", si, pi, seed, err)
			}
		}
	}
}

// AC-4 oracle-side: ReplayVerify is pure — a second run over the same result
// agrees with the first (nil both times, all 55). The solver-side byte-compare
// determinism tests (determinism_test.go, upper_corpus_test.go) stand untouched.
func TestReplayVerify_FullCorpus_SecondRunConsistent(t *testing.T) {
	for si, sec := range readCorpusSections(t) {
		for pi, seed := range sec {
			g := mustParse(t, seed)
			res := solver.Solve(g)
			first := oracle.ReplayVerify(g, res)
			second := oracle.ReplayVerify(g, res)
			if first != nil || second != nil {
				t.Fatalf("section %d puzzle %d: ReplayVerify runs = (%v, %v), want (nil, nil)", si, pi, first, second)
			}
		}
	}
}

// ADR-0014 boundary: a complete valid grid solves with zero events; its replay
// proof is vacuously valid (count==1, final grid == oracle solution).
func TestReplayVerify_CompleteGrid_ZeroEvents(t *testing.T) {
	g := mustParse(t, completeGrid)
	res := solver.Solve(g)
	if err := oracle.ReplayVerify(g, res); err != nil {
		t.Fatalf("ReplayVerify on the zero-event complete-grid solve: %v", err)
	}
}
