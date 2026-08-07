package httpapi_test

// F-10 AC-4 — POST /v1/validate-batch. ADR-0014 pins every per-item field
// value; AUDIT A7 pins the two caps (byte cap at the edge, item cap
// post-parse pre-solve). Runs under -race per CONTEXT.md test discipline —
// the full-corpus batch is the goroutine-per-puzzle race vehicle.

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/NerdAlert58/sudoku-flow2/httpapi"
	"github.com/NerdAlert58/sudoku-flow2/solver"
)

func TestBatchContractFullCorpusInOrder(t *testing.T) {
	h := httpapi.New()
	puzzles := corpus(t)
	rr := postJSON(t, h, "/v1/validate-batch", batchReqBody(t, puzzles))
	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 (body %q)", rr.Code, rr.Body.String())
	}
	assertKeyOrder(t, rr.Body.Bytes(), batchFieldOrder)
	var resp batchResponse
	strictDecode(t, rr.Body.Bytes(), &resp)
	if resp.APIVersion != "1" {
		t.Errorf("apiVersion = %q, want %q", resp.APIVersion, "1")
	}
	if resp.Total != 55 || resp.SolvedCount != 55 {
		t.Errorf("solvedCount/total = %d/%d, want 55/55", resp.SolvedCount, resp.Total)
	}
	if len(resp.Results) != 55 {
		t.Fatalf("len(results) = %d, want 55", len(resp.Results))
	}
	for i, item := range resp.Results {
		want := solver.Solve(mustParseGrid(t, puzzles[i]))
		if item.Puzzle != puzzles[i] {
			t.Errorf("result %d: puzzle echo = %q, want %q (in-order)", i, item.Puzzle, puzzles[i])
		}
		if !item.Solved {
			t.Errorf("result %d: solved = false, want true", i)
		}
		if item.SolveTimeMs <= 0 {
			t.Errorf("result %d: solveTimeMs = %v, want > 0", i, item.SolveTimeMs)
		}
		if item.Iterations != want.Iterations {
			t.Errorf("result %d: iterations = %d, want %d (independent per-solve counters)", i, item.Iterations, want.Iterations)
		}
		if hardest := hardestFired(t, want.Events); item.HardestTechnique != hardest {
			t.Errorf("result %d: hardestTechnique = %q, want %q", i, item.HardestTechnique, hardest)
		}
	}
}

func TestBatchContractItemFieldOrder(t *testing.T) {
	h := httpapi.New()
	rr := postJSON(t, h, "/v1/validate-batch", batchReqBody(t, corpus(t)[:1]))
	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 (body %q)", rr.Code, rr.Body.String())
	}
	var wrap struct {
		Results []json.RawMessage `json:"results"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &wrap); err != nil {
		t.Fatal(err)
	}
	if len(wrap.Results) != 1 {
		t.Fatalf("len(results) = %d, want 1", len(wrap.Results))
	}
	assertKeyOrder(t, wrap.Results[0], batchItemFieldOrder)
}

func TestBatchContractADR0014ItemValues(t *testing.T) {
	h := httpapi.New()
	all := corpus(t)
	valid, inner := all[0], all[1]
	wrapped := "  " + inner + "\r\n"
	stalled := strings.Repeat("0", 81)
	complete := oracleSolution(t, valid)
	dup := "11" + strings.Repeat("0", 79)

	validRes := solver.Solve(mustParseGrid(t, valid))
	innerRes := solver.Solve(mustParseGrid(t, inner))
	stalledRes := solver.Solve(mustParseGrid(t, stalled))
	if stalledRes.Status != "stalled" {
		t.Fatalf("fixture drift: all-zeros grid status = %q, want stalled", stalledRes.Status)
	}

	items := []string{valid, wrapped, "nonsense", dup, stalled, complete}
	rr := postJSON(t, h, "/v1/validate-batch", batchReqBody(t, items))
	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 (body %q)", rr.Code, rr.Body.String())
	}
	var resp batchResponse
	strictDecode(t, rr.Body.Bytes(), &resp)
	if resp.Total != 6 || resp.SolvedCount != 3 {
		t.Errorf("solvedCount/total = %d/%d, want 3/6", resp.SolvedCount, resp.Total)
	}
	if len(resp.Results) != 6 {
		t.Fatalf("len(results) = %d, want 6", len(resp.Results))
	}

	checks := []struct {
		name      string
		want      batchItem
		attempted bool
	}{
		{"valid", batchItem{Puzzle: valid, Solved: true, Iterations: validRes.Iterations, HardestTechnique: hardestFired(t, validRes.Events)}, true},
		{"whitespace CRLF wrapped", batchItem{Puzzle: wrapped, Solved: true, Iterations: innerRes.Iterations, HardestTechnique: hardestFired(t, innerRes.Events)}, true},
		{"malformed short line", batchItem{Puzzle: "nonsense"}, false},
		{"malformed duplicate givens", batchItem{Puzzle: dup}, false},
		{"stalled attempted item", batchItem{Puzzle: stalled, Iterations: stalledRes.Iterations, HardestTechnique: hardestFired(t, stalledRes.Events)}, true},
		{"complete grid", batchItem{Puzzle: complete, Solved: true, Iterations: 0, HardestTechnique: ""}, true},
	}
	for i, c := range checks {
		got := resp.Results[i]
		t.Run(c.name, func(t *testing.T) {
			if c.attempted {
				if got.SolveTimeMs <= 0 {
					t.Errorf("solveTimeMs = %v, want > 0 for an attempted item", got.SolveTimeMs)
				}
				got.SolveTimeMs = 0
			}
			if got != c.want {
				t.Errorf("item:\n got %+v\nwant %+v (ADR-0014; malformed items must be exactly {raw,false,0,0,\"\"})", got, c.want)
			}
		})
	}
}

func TestBatchContractEmptyList(t *testing.T) {
	h := httpapi.New()
	rr := postJSON(t, h, "/v1/validate-batch", `{"puzzles":[]}`)
	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 (body %q)", rr.Code, rr.Body.String())
	}
	var resp batchResponse
	strictDecode(t, rr.Body.Bytes(), &resp)
	if resp.Total != 0 || resp.SolvedCount != 0 || len(resp.Results) != 0 {
		t.Errorf("empty batch: %+v, want zero counts and no results", resp)
	}
	var m map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &m); err != nil {
		t.Fatal(err)
	}
	if m["results"] == nil {
		t.Error(`results is null; want the JSON array [] (ADR-0004 array convention)`)
	}
}

func TestBatchContractItemCap(t *testing.T) {
	h := httpapi.New()
	p := corpus(t)[0]

	over := make([]string, 257)
	for i := range over {
		over[i] = p
	}
	rr := postJSON(t, h, "/v1/validate-batch", batchReqBody(t, over))
	decodeEnvelope(t, rr, http.StatusRequestEntityTooLarge, "payload_too_large")
	assertNoInternals(t, rr.Body.String())

	atCap := over[:256]
	rr = postJSON(t, h, "/v1/validate-batch", batchReqBody(t, atCap))
	if rr.Code != http.StatusOK {
		t.Fatalf("256 items: status = %d, want 200 (cap is inclusive; body %q)", rr.Code, rr.Body.String())
	}
	var resp batchResponse
	strictDecode(t, rr.Body.Bytes(), &resp)
	if resp.Total != 256 || resp.SolvedCount != 256 {
		t.Errorf("solvedCount/total = %d/%d, want 256/256", resp.SolvedCount, resp.Total)
	}
}

func TestBatchContractNonStringItemIsMalformedJSON(t *testing.T) {
	h := httpapi.New()
	rr := postJSON(t, h, "/v1/validate-batch", `{"puzzles":[123]}`)
	decodeEnvelope(t, rr, http.StatusBadRequest, "invalid_input")
}
