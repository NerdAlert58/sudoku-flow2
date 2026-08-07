package generate_test

import (
	"testing"

	"github.com/NerdAlert58/sudoku-flow2/oracle"
	"github.com/NerdAlert58/sudoku-flow2/solver"
)

// Committed AC-6 seed scheme (EVAL "UC-2 Replay proof", generated slice):
// 1000*(bandIndex+1)+500+i, i in 0..4 — easy 1500..1504, medium 2500..2504,
// hard 3500..3504, expert 4500..4504. Disjoint from the AC-1 ranges.
func ac6Seed(bandIdx, i int) int64 { return int64(1000*(bandIdx+1) + 500 + i) }

// AC-6 (eval row "UC-2 Replay proof", generated slice): 20 seeded
// generations, 5 per band. Every generated puzzle's solve replays clean
// through oracle.ReplayVerify — the verifier is fail-fast per event, so nil
// means every event passed every ADR-0013 check and the final grid equals
// the oracle's unique solution.
func TestGenerate_ReplaySlice(t *testing.T) {
	for bi, b := range bands {
		t.Run(b.band, func(t *testing.T) {
			t.Parallel()
			for i := 0; i < 5; i++ {
				seed := ac6Seed(bi, i)
				puzzle, _, err := generateSeeded(t, b.band, seed)
				if err != nil {
					t.Errorf("seed %d: Generate(%q) error: %v", seed, b.band, err)
					continue
				}
				g, perr := solver.Parse(puzzle)
				if perr != nil {
					t.Errorf("seed %d: puzzle %q does not parse: %v", seed, puzzle, perr)
					continue
				}
				res := solver.Solve(g)
				if res.Status != "solved" {
					t.Errorf("seed %d: ladder status = %q, want solved (puzzle %s)", seed, res.Status, puzzle)
					continue
				}
				if verr := oracle.ReplayVerify(g, res); verr != nil {
					t.Errorf("seed %d: replay verification failed: %v (puzzle %s)", seed, verr, puzzle)
				}
			}
		})
	}
}
