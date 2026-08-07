// Package generate produces difficulty-graded sudoku puzzles: randomized
// full-grid fill, clue removal under a capped uniqueness counter, grading via
// solver.Solve, and a grade-targeted accept/retry loop under the caller's
// context deadline (ADR-0009). The backtracking here is the only shipped
// backtracking and never escapes this package (ADR-0002).
package generate

import (
	"context"
	"errors"
	"fmt"
	"math/rand"

	"github.com/NerdAlert58/sudoku-flow2/solver"
)

var (
	ErrUnknownBand     = errors.New("unknown difficulty band")
	ErrBudgetExhausted = errors.New("generation budget exhausted")
)

// grades maps the request enum (C3: exact lowercase only) to the solver's
// grade literals. Success returns the mapped literal, so a grade/band
// mismatch is unrepresentable (AC-2).
var grades = map[string]string{
	"easy":   "Easy",
	"medium": "Medium",
	"hard":   "Hard",
	"expert": "Expert",
}

// maxGivens: grading starts only once a dig is at or below this many givens,
// so an accepted easy puzzle is a real puzzle, not a near-full grid.
const maxGivens = 45

func Generate(ctx context.Context, band string, rng *rand.Rand) (string, string, error) {
	want, ok := grades[band]
	if !ok {
		return "", "", fmt.Errorf("generate: band %q: %w", band, ErrUnknownBand)
	}
	for {
		if err := ctx.Err(); err != nil {
			return "", "", fmt.Errorf("generate: %w: %w", ErrBudgetExhausted, err)
		}
		if puzzle, ok := dig(ctx, rng, fill(rng), want); ok {
			return puzzle, want, nil
		}
	}
}

// dig removes clues from a full grid in rng order, keeping only removals that
// preserve uniqueness. At or below maxGivens givens it grades after every
// removal and returns the first state whose ladder grade equals want; every
// accepted puzzle is therefore unique-by-construction and ladder-solvable
// with at least one blank. ok=false when the dig bottoms out minimal (or the
// context dies) without a match; the caller retries with a fresh fill.
func dig(ctx context.Context, rng *rand.Rand, g solver.Grid, want string) (string, bool) {
	givens := 81
	for _, i := range rng.Perm(81) {
		if ctx.Err() != nil {
			return "", false
		}
		d := g[i]
		g[i] = 0
		if countUpTo2(g) != 1 {
			g[i] = d
			continue
		}
		givens--
		if givens > maxGivens {
			continue
		}
		if res := solver.Solve(g); res.Status == "solved" && res.Grade == want {
			return g.String(), true
		}
	}
	return "", false
}
