package httpapi_test

// F-10 pin (verifier RUBRIC_GAP-1) — POST /v1/solve wire behavior for the two
// non-solved terminal statuses. ADR-0009/D-009: stalled and unsolvable are
// HTTP 200, not errors. ADR-0007: iterations == eventCount+1 for both.

import (
	"net/http"
	"testing"

	"github.com/NerdAlert58/sudoku-flow2/httpapi"
	"github.com/NerdAlert58/sudoku-flow2/solver"
)

// Copied from solver/loop_test.go beyondLadderGrid (external test packages
// are unimportable): permanent full-ladder stall with zero events.
const beyondLadderGrid = "762314589819576234534928176476152398103869407908437601681293745247685913395741862"

// Copied from solver/edge_test.go zeroCandidateGrid: placing the naked
// single at (0,0) strips (2,2) to zero candidates (ADR-0008).
const zeroCandidateGrid = "000007800" +
	"000000000" +
	"000123456" +
	"107000000" +
	"208000000" +
	"300000000" +
	"400000000" +
	"500000000" +
	"600000000"

func TestSolveContractStalled200(t *testing.T) {
	isolated := solver.Solve(mustParseGrid(t, beyondLadderGrid))
	if isolated.Status != "stalled" {
		t.Fatalf("fixture drift: isolated status = %q, want stalled", isolated.Status)
	}
	if isolated.Solution.String() != beyondLadderGrid {
		t.Fatalf("fixture drift: isolated solution = %q, want the input (zero-event stall)", isolated.Solution.String())
	}
	rr := postJSON(t, httpapi.New(), "/v1/solve", solveReqBody(beyondLadderGrid))
	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 per ADR-0009 (body %q)", rr.Code, rr.Body.String())
	}
	assertFrozenHeaders(t, rr.Header())
	assertSolveShape(t, rr.Body.Bytes())
	var resp solveResponse
	strictDecode(t, rr.Body.Bytes(), &resp)
	if resp.Status != "stalled" || resp.Solved {
		t.Errorf("status/solved = %q/%v, want stalled/false", resp.Status, resp.Solved)
	}
	if resp.Solution != isolated.Solution.String() {
		t.Errorf("solution = %q, want isolated solver partial grid %q", resp.Solution, isolated.Solution.String())
	}
	if resp.Grade != "" {
		t.Errorf("grade = %q, want empty for stalled", resp.Grade)
	}
	if len(resp.Events) != 0 {
		t.Errorf("len(events) = %d, want 0 for this zero-event stall", len(resp.Events))
	}
	if resp.Iterations != resp.EventCount+1 {
		t.Errorf("iterations %d != eventCount %d + 1 (ADR-0007 stalled invariant)", resp.Iterations, resp.EventCount)
	}
	if resp.SolveTimeMs <= 0 {
		t.Errorf("solveTimeMs = %v, want > 0", resp.SolveTimeMs)
	}
}

func TestSolveContractUnsolvable200(t *testing.T) {
	isolated := solver.Solve(mustParseGrid(t, zeroCandidateGrid))
	if isolated.Status != "unsolvable" {
		t.Fatalf("fixture drift: isolated status = %q, want unsolvable", isolated.Status)
	}
	if isolated.Solution.String() == zeroCandidateGrid {
		t.Fatal("fixture drift: isolated solution equals the input; want the partial grid with placements made")
	}
	rr := postJSON(t, httpapi.New(), "/v1/solve", solveReqBody(zeroCandidateGrid))
	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 per ADR-0009 (body %q)", rr.Code, rr.Body.String())
	}
	assertFrozenHeaders(t, rr.Header())
	assertSolveShape(t, rr.Body.Bytes())
	var resp solveResponse
	strictDecode(t, rr.Body.Bytes(), &resp)
	if resp.Status != "unsolvable" || resp.Solved {
		t.Errorf("status/solved = %q/%v, want unsolvable/false", resp.Status, resp.Solved)
	}
	if resp.Solution != isolated.Solution.String() {
		t.Errorf("solution = %q, want isolated solver partial grid %q", resp.Solution, isolated.Solution.String())
	}
	if resp.Grade != "" {
		t.Errorf("grade = %q, want empty for unsolvable", resp.Grade)
	}
	if resp.Iterations != resp.EventCount+1 {
		t.Errorf("iterations %d != eventCount %d + 1 (ADR-0007 unsolvable invariant)", resp.Iterations, resp.EventCount)
	}
	if resp.SolveTimeMs <= 0 {
		t.Errorf("solveTimeMs = %v, want > 0", resp.SolveTimeMs)
	}
}
