package oracle_test

import (
	"strings"
	"testing"

	"github.com/NerdAlert58/sudoku-flow2/oracle"
	"github.com/NerdAlert58/sudoku-flow2/solver"
)

// Fixture provenance: completeGrid copies solver/edge_test.go's constant;
// twoSolutionGrid copies solver/loop_test.go's beyondLadderGrid — four empty
// cells forming a {2,5} rectangle at r4/r5 x c1/c7, exactly two completions
// (re-verified by an independent brute force for this piece).
const (
	completeGrid    = "123456789456789123789123456214365897365897214897214365531642978642978531978531642"
	twoSolutionGrid = "762314589819576234534928176476152398103869407908437601681293745247685913395741862"
)

func mustParse(t *testing.T, s string) solver.Grid {
	t.Helper()
	g, err := solver.Parse(s)
	if err != nil {
		t.Fatalf("Parse(%q): %v", s, err)
	}
	return g
}

// AC-1 tiny fixture: an 81-given valid grid is its own unique solution.
func TestSolve_CompleteGrid_CountOneIdentity(t *testing.T) {
	g := mustParse(t, completeGrid)
	sol, count := oracle.Solve(g)
	if count != 1 {
		t.Fatalf("count = %d, want 1", count)
	}
	if sol != g {
		t.Fatalf("solution = %s, want the input grid", sol.String())
	}
}

// AC-1 tiny fixture, hand-checkable: blanking r0c0 of completeGrid leaves row 0
// holding 2-9, so the oracle must restore digit 1 and report uniqueness.
func TestSolve_NearlyCompleteGrid_HandCheckable(t *testing.T) {
	sol, count := oracle.Solve(mustParse(t, "0"+completeGrid[1:]))
	if count != 1 {
		t.Fatalf("count = %d, want 1", count)
	}
	if got := sol.String(); got != completeGrid {
		t.Fatalf("solution = %s, want %s", got, completeGrid)
	}
}

func TestSolve_TwoSolutionGrid_CountTwo(t *testing.T) {
	if _, count := oracle.Solve(mustParse(t, twoSolutionGrid)); count != 2 {
		t.Fatalf("count = %d, want 2", count)
	}
}

// The count is capped at 2: even the empty grid (astronomically many
// completions) reports exactly 2 — the cap keeps every oracle call cheap.
func TestSolve_EmptyGrid_CountCappedAtTwo(t *testing.T) {
	if _, count := oracle.Solve(mustParse(t, strings.Repeat("0", 81))); count != 2 {
		t.Fatalf("count = %d, want 2 (capped)", count)
	}
}
