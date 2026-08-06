package solver_test

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/NerdAlert58/sudoku-flow2/solver"
)

func TestSolve_ConcurrentCountersMatchSerial(t *testing.T) {
	seeds := originalSeeds(t)
	type baseline struct {
		iterations, eventCount, candidateChecks int
		events                                  []byte
	}
	serial := make([]baseline, len(seeds))
	for i, seed := range seeds {
		res := solver.Solve(mustParse(t, seed))
		serial[i] = baseline{res.Iterations, res.EventCount, res.CandidateChecks, eventsJSON(t, res.Events)}
	}
	t.Run("parallel", func(t *testing.T) {
		for i, seed := range seeds {
			t.Run(fmt.Sprintf("seed%02d", i), func(t *testing.T) {
				t.Parallel()
				res := solver.Solve(mustParse(t, seed))
				want := serial[i]
				if res.Iterations != want.iterations ||
					res.EventCount != want.eventCount ||
					res.CandidateChecks != want.candidateChecks {
					t.Fatalf("concurrent counters (iterations=%d eventCount=%d candidateChecks=%d) != serial baseline (iterations=%d eventCount=%d candidateChecks=%d)",
						res.Iterations, res.EventCount, res.CandidateChecks,
						want.iterations, want.eventCount, want.candidateChecks)
				}
				if !bytes.Equal(eventsJSON(t, res.Events), want.events) {
					t.Fatal("concurrent events JSON != serial baseline")
				}
			})
		}
	})
}
