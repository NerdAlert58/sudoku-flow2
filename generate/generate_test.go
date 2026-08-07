package generate_test

import (
	"context"
	"errors"
	"math/rand"
	"testing"
	"time"

	"github.com/NerdAlert58/sudoku-flow2/generate"
	"github.com/NerdAlert58/sudoku-flow2/oracle"
	"github.com/NerdAlert58/sudoku-flow2/solver"
)

// AC-4: the C3 tuple is exactly (puzzle, grade, err) — no counter or attempt
// data can appear in any return value, enforced at compile time.
// AC-5 needs no test here: the determinism suites are solver-side
// (solver/determinism_test.go) and this piece never touches them.
var _ func(ctx context.Context, band string, rng *rand.Rand) (puzzle string, grade string, err error) = generate.Generate

// bands pins the PRD request->grade mapping: easy->Easy, medium->Medium,
// hard->Hard, expert->Expert. Request bands are the lowercase enum only.
var bands = []struct{ band, grade string }{
	{"easy", "Easy"},
	{"medium", "Medium"},
	{"hard", "Hard"},
	{"expert", "Expert"},
}

// Committed AC-1 seed scheme (EVAL "UC-3 Generate"): 1000*(bandIndex+1)+i,
// i in 0..24 — easy 1000..1024, medium 2000..2024, hard 3000..3024,
// expert 4000..4024.
func ac1Seed(bandIdx, i int) int64 { return int64(1000*(bandIdx+1) + i) }

const perCallDeadline = 5 * time.Second

func generateSeeded(t *testing.T, band string, seed int64) (string, string, error) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), perCallDeadline)
	defer cancel()
	return generate.Generate(ctx, band, rand.New(rand.NewSource(seed)))
}

// AC-1 + AC-2 (eval row "UC-3 Generate"): 25 seeded generations per band,
// 100 total. Each puzzle parses, is oracle-unique, ladder-solves to the
// oracle solution with at least one event (a complete grid is not a
// generated puzzle), and carries the requested band's exact grade literal
// both in Generate's return and end-to-end through solver.Solve — the
// generation-independent band-literal assertion from the F-05 intake note.
// AC-2's explicit mismatch counter must be zero per band.
func TestGenerate_SeededMatrix(t *testing.T) {
	for bi, b := range bands {
		t.Run(b.band, func(t *testing.T) {
			t.Parallel()
			mismatches := 0
			for i := 0; i < 25; i++ {
				mismatches += checkGeneration(t, b.band, b.grade, ac1Seed(bi, i))
			}
			if mismatches != 0 {
				t.Errorf("AC-2: %d grade mismatches over 25 %s generations, want 0", mismatches, b.band)
			}
		})
	}
}

// checkGeneration runs one seeded generation, asserts the AC-1 properties,
// and returns the number of grade mismatches it observed (the AC-2 counter).
func checkGeneration(t *testing.T, band, wantGrade string, seed int64) int {
	t.Helper()
	puzzle, grade, err := generateSeeded(t, band, seed)
	if err != nil {
		t.Errorf("seed %d: Generate(%q) error: %v", seed, band, err)
		return 0
	}
	mismatches := 0
	if grade != wantGrade {
		mismatches++
		t.Errorf("seed %d: grade = %q, want %q", seed, grade, wantGrade)
	}
	g, perr := solver.Parse(puzzle)
	if perr != nil {
		t.Errorf("seed %d: puzzle %q does not parse: %v", seed, puzzle, perr)
		return mismatches
	}
	sol, count := oracle.Solve(g)
	if count != 1 {
		t.Errorf("seed %d: oracle count = %d, want 1 (puzzle %s)", seed, count, puzzle)
	}
	res := solver.Solve(g)
	if res.Status != "solved" {
		t.Errorf("seed %d: ladder status = %q, want solved (puzzle %s)", seed, res.Status, puzzle)
		return mismatches
	}
	if res.EventCount == 0 {
		t.Errorf("seed %d: zero-event solve — the generated puzzle has no blanks (puzzle %s)", seed, puzzle)
	}
	if res.Solution != sol {
		t.Errorf("seed %d: ladder solution %s != oracle solution %s", seed, res.Solution.String(), sol.String())
	}
	if res.Grade != wantGrade {
		mismatches++
		t.Errorf("seed %d: end-to-end solver grade = %q, want %q (puzzle %s)", seed, res.Grade, wantGrade, puzzle)
	}
	return mismatches
}

// AC-3: an already-expired deadline surfaces as an error wrapping BOTH
// generate.ErrBudgetExhausted and the context's error — never a puzzle,
// however fast one could have been produced — with empty puzzle/grade and a
// prompt return.
func TestGenerate_ExpiredDeadline(t *testing.T) {
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(-time.Millisecond))
	defer cancel()
	start := time.Now()
	puzzle, grade, err := generate.Generate(ctx, "expert", rand.New(rand.NewSource(1)))
	elapsed := time.Since(start)
	if err == nil {
		t.Fatalf("expired deadline: got (%q, %q, nil), want error", puzzle, grade)
	}
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("error %v does not wrap context.DeadlineExceeded", err)
	}
	if !errors.Is(err, generate.ErrBudgetExhausted) {
		t.Errorf("error %v does not wrap generate.ErrBudgetExhausted", err)
	}
	if puzzle != "" || grade != "" {
		t.Errorf("error return carried (%q, %q), want empty strings", puzzle, grade)
	}
	if elapsed > 2*time.Second {
		t.Errorf("returned in %v, want a prompt return on a dead context", elapsed)
	}
}

// AC-3: a canceled context is the same contract — error wrapping
// ErrBudgetExhausted and context.Canceled, empty puzzle/grade.
func TestGenerate_CanceledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	puzzle, grade, err := generate.Generate(ctx, "easy", rand.New(rand.NewSource(2)))
	if err == nil {
		t.Fatalf("canceled context: got (%q, %q, nil), want error", puzzle, grade)
	}
	if !errors.Is(err, context.Canceled) {
		t.Errorf("error %v does not wrap context.Canceled", err)
	}
	if !errors.Is(err, generate.ErrBudgetExhausted) {
		t.Errorf("error %v does not wrap generate.ErrBudgetExhausted", err)
	}
	if puzzle != "" || grade != "" {
		t.Errorf("error return carried (%q, %q), want empty strings", puzzle, grade)
	}
}

// AC-4: any band outside the exact lowercase enum — including canonical
// grade casing like "Easy" — is errors.Is(err, generate.ErrUnknownBand)
// with empty puzzle/grade.
func TestGenerate_UnknownBand(t *testing.T) {
	for _, band := range []string{"", "Easy", "EXPERT", "very hard", "banana"} {
		ctx, cancel := context.WithTimeout(context.Background(), perCallDeadline)
		puzzle, grade, err := generate.Generate(ctx, band, rand.New(rand.NewSource(3)))
		cancel()
		if !errors.Is(err, generate.ErrUnknownBand) {
			t.Errorf("band %q: err = %v, want errors.Is(err, generate.ErrUnknownBand)", band, err)
		}
		if puzzle != "" || grade != "" {
			t.Errorf("band %q: got (%q, %q), want empty strings", band, puzzle, grade)
		}
	}
}
