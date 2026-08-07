package solver_test

import (
	"errors"
	"testing"

	"github.com/NerdAlert58/sudoku-flow2/solver"
)

const completeGrid = "123456789456789123789123456214365897365897214897214365531642978642978531978531642"

// Parses clean (no unit has a duplicate given), but (0,0) sees 1-8 via its row and
// column, making it a naked single 9; placing it strips box-mate (2,2) — whose row
// and column peers already cover 1-8 — to zero candidates (ADR-0008 top-of-pass check).
const zeroCandidateGrid = "000007800" +
	"000000000" +
	"000123456" +
	"107000000" +
	"208000000" +
	"300000000" +
	"400000000" +
	"500000000" +
	"600000000"

func TestSolve_CompleteGrid(t *testing.T) {
	res := solver.Solve(mustParse(t, completeGrid))
	if res.Status != "solved" {
		t.Fatalf("Status = %q, want solved", res.Status)
	}
	if res.Grade != "Easy" {
		t.Fatalf("Grade = %q, want Easy (ADR-0014 floor band)", res.Grade)
	}
	if res.Iterations != 0 || res.EventCount != 0 || res.CandidateChecks != 0 {
		t.Fatalf("counters = (iterations=%d eventCount=%d candidateChecks=%d), want (0 0 0) per ADR-0014", res.Iterations, res.EventCount, res.CandidateChecks)
	}
	if len(res.Events) != 0 {
		t.Fatalf("len(Events) = %d, want 0", len(res.Events))
	}
	if got := res.Solution.String(); got != completeGrid {
		t.Fatalf("Solution.String() = %q, want the input grid", got)
	}
}

func TestParse_CompleteGridWithDuplicate(t *testing.T) {
	b := []byte(completeGrid)
	b[80] = '4'
	_, err := solver.Parse(string(b))
	if err == nil {
		t.Fatal("Parse accepted a complete grid with a duplicate; want ErrDuplicateGiven")
	}
	if !errors.Is(err, solver.ErrDuplicateGiven) {
		t.Fatalf("Parse error = %v, want errors.Is(err, ErrDuplicateGiven)", err)
	}
}

func TestSolve_ZeroCandidate_Unsolvable(t *testing.T) {
	res := solver.Solve(mustParse(t, zeroCandidateGrid))
	if res.Status != "unsolvable" {
		t.Fatalf("Status = %q, want unsolvable (zero-candidate cell, ADR-0008)", res.Status)
	}
	if res.Grade != "" {
		t.Fatalf("Grade = %q, want empty for unsolvable", res.Grade)
	}
	if res.Iterations != res.EventCount+1 {
		t.Fatalf("Iterations = %d, EventCount = %d; ADR-0007 requires Iterations == EventCount+1 for unsolvable", res.Iterations, res.EventCount)
	}
}
