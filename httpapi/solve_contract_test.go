package httpapi_test

// F-10 AC-1/AC-2/AC-6(solve-shape)/AC-8 — POST /v1/solve wire contract.
// ADR-0004 (400 carries the FULL solve shape; raw input echo), ADR-0006
// (solveTimeMs float, outside byte-identity), ADR-0014 (complete-grid edge).

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strings"
	"testing"

	"github.com/NerdAlert58/sudoku-flow2/catalog"
	"github.com/NerdAlert58/sudoku-flow2/httpapi"
)

func TestSolveContractCorpusOracleEqual(t *testing.T) {
	h := httpapi.New()
	for i, p := range corpus(t) {
		t.Run(fmt.Sprintf("puzzle_%02d", i), func(t *testing.T) {
			rr := postJSON(t, h, "/v1/solve", solveReqBody(p))
			if rr.Code != http.StatusOK {
				t.Fatalf("status = %d, want 200 (body %q)", rr.Code, rr.Body.String())
			}
			if ct := rr.Header().Get("Content-Type"); !strings.HasPrefix(ct, "application/json") {
				t.Errorf("Content-Type = %q, want application/json", ct)
			}
			assertSolveShape(t, rr.Body.Bytes())
			var resp solveResponse
			strictDecode(t, rr.Body.Bytes(), &resp)
			if resp.APIVersion != "1" {
				t.Errorf("apiVersion = %q, want %q", resp.APIVersion, "1")
			}
			if resp.Input != p {
				t.Errorf("input echo = %q, want %q", resp.Input, p)
			}
			if resp.Status != "solved" || !resp.Solved {
				t.Fatalf("status/solved = %q/%v, want solved/true", resp.Status, resp.Solved)
			}
			if want := oracleSolution(t, p); resp.Solution != want {
				t.Errorf("solution = %q, want oracle %q", resp.Solution, want)
			}
			if resp.Grade == "" {
				t.Error("grade is empty for a solved puzzle")
			}
			if resp.Iterations != resp.EventCount {
				t.Errorf("iterations %d != eventCount %d (ADR-0007 solved invariant)", resp.Iterations, resp.EventCount)
			}
			if len(resp.Events) != resp.EventCount {
				t.Errorf("len(events) %d != eventCount %d", len(resp.Events), resp.EventCount)
			}
			if resp.CandidateChecks <= 0 {
				t.Errorf("candidateChecks = %d, want > 0", resp.CandidateChecks)
			}
			if resp.SolveTimeMs <= 0 {
				t.Errorf("solveTimeMs = %v, want > 0", resp.SolveTimeMs)
			}
			if n := len(resp.Events); n > 0 && resp.Events[n-1].GridAfter != resp.Solution {
				t.Error("last event gridAfter != solution")
			}
		})
	}
}

func TestSolveContractGoldenSolvedBody(t *testing.T) {
	h := httpapi.New()
	p := catalog.Sections()[3].Puzzles[0]
	sol := oracleSolution(t, p)
	rr := postJSON(t, h, "/v1/solve", solveReqBody(p))
	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 (body %q)", rr.Code, rr.Body.String())
	}
	body := trimmedBody(rr)
	wantPrefix := `{"apiVersion":"1","input":"` + p + `","status":"solved","solved":true,"solution":"` + sol + `","iterations":`
	if !strings.HasPrefix(body, wantPrefix) {
		t.Errorf("raw body prefix mismatch\n got: %.200s\nwant: %.200s", body, wantPrefix)
	}
	assertKeyOrder(t, []byte(body), solveFieldOrder)
	assertEventShapes(t, []byte(body))
}

// assertEventShapes pins the C1 event object orders on one golden response:
// placement events {seq,technique,witnessCells,placement,gridAfter},
// elimination events {seq,technique,witnessCells,eliminations,gridAfter},
// cells {row,col}, placements/eliminations {row,col,digit}.
func assertEventShapes(t *testing.T, body []byte) {
	t.Helper()
	var wrap struct {
		Events []json.RawMessage `json:"events"`
	}
	if err := json.Unmarshal(body, &wrap); err != nil {
		t.Fatal(err)
	}
	var sawPlacement, sawElimination bool
	for _, raw := range wrap.Events {
		var ev struct {
			WitnessCells []json.RawMessage `json:"witnessCells"`
			Placement    json.RawMessage   `json:"placement"`
			Eliminations []json.RawMessage `json:"eliminations"`
		}
		if err := json.Unmarshal(raw, &ev); err != nil {
			t.Fatal(err)
		}
		if len(ev.WitnessCells) > 0 {
			assertKeyOrder(t, ev.WitnessCells[0], []string{"row", "col"})
		}
		switch {
		case ev.Placement != nil && !sawPlacement:
			sawPlacement = true
			assertKeyOrder(t, raw, []string{"seq", "technique", "witnessCells", "placement", "gridAfter"})
			assertKeyOrder(t, ev.Placement, []string{"row", "col", "digit"})
		case ev.Eliminations != nil && !sawElimination:
			sawElimination = true
			assertKeyOrder(t, raw, []string{"seq", "technique", "witnessCells", "eliminations", "gridAfter"})
			assertKeyOrder(t, ev.Eliminations[0], []string{"row", "col", "digit"})
		}
	}
	if !sawPlacement || !sawElimination {
		t.Fatalf("golden fixture lacks event variety: placement=%v elimination=%v", sawPlacement, sawElimination)
	}
}

func TestSolveContractDeterministicDoublePost(t *testing.T) {
	h := httpapi.New()
	p := catalog.Sections()[3].Puzzles[1]
	var bodies [2]map[string]any
	for i := range bodies {
		rr := postJSON(t, h, "/v1/solve", solveReqBody(p))
		if rr.Code != http.StatusOK {
			t.Fatalf("POST %d: status = %d, want 200", i, rr.Code)
		}
		if err := json.Unmarshal(rr.Body.Bytes(), &bodies[i]); err != nil {
			t.Fatal(err)
		}
		ms, ok := bodies[i]["solveTimeMs"].(float64)
		if !ok || ms <= 0 {
			t.Errorf("POST %d: solveTimeMs = %v, want positive float", i, bodies[i]["solveTimeMs"])
		}
		delete(bodies[i], "solveTimeMs")
	}
	if !reflect.DeepEqual(bodies[0], bodies[1]) {
		t.Error("double-POST responses differ beyond solveTimeMs (ADR-0006 violation)")
	}
}

func TestSolveContractInvalidGoldenBody(t *testing.T) {
	h := httpapi.New()
	raw := strings.Repeat("0", 80) + "x"
	rr := postJSON(t, h, "/v1/solve", solveReqBody(raw))
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400 (body %q)", rr.Code, rr.Body.String())
	}
	want := `{"apiVersion":"1","input":"` + raw + `","status":"invalid_input","solved":false,` +
		`"solution":"","iterations":0,"eventCount":0,"candidateChecks":0,"solveTimeMs":0,` +
		`"grade":"","events":[]}`
	if got := trimmedBody(rr); got != want {
		t.Errorf("golden 400 body:\n got %s\nwant %s", got, want)
	}
	assertFrozenHeaders(t, rr.Header())
}

func TestSolveContractDomainInvalid400(t *testing.T) {
	h := httpapi.New()
	complete := oracleSolution(t, catalog.Sections()[0].Puzzles[0])
	dup := []byte(complete)
	dup[0] = complete[1]
	cases := []struct {
		name, puzzle string
	}{
		{"too short", strings.Repeat("0", 80)},
		{"too long", strings.Repeat("0", 82)},
		{"empty", ""},
		{"bad char", strings.Repeat("0", 40) + "a" + strings.Repeat("0", 40)},
		{"duplicate givens", "11" + strings.Repeat("0", 79)},
		{"complete but duplicate", string(dup)},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			rr := postJSON(t, h, "/v1/solve", solveReqBody(c.puzzle))
			if rr.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, want 400 (body %q)", rr.Code, rr.Body.String())
			}
			assertSolveShape(t, rr.Body.Bytes())
			var resp solveResponse
			strictDecode(t, rr.Body.Bytes(), &resp)
			zeroed := resp.Status == "invalid_input" && !resp.Solved && resp.Solution == "" &&
				resp.Iterations == 0 && resp.EventCount == 0 && resp.CandidateChecks == 0 &&
				resp.SolveTimeMs == 0 && resp.Grade == "" && len(resp.Events) == 0
			if !zeroed {
				t.Errorf("400 body not the zeroed invalid_input shape: %+v", resp)
			}
			if resp.Input != c.puzzle {
				t.Errorf("input echo = %q, want raw %q", resp.Input, c.puzzle)
			}
			assertNoInternals(t, rr.Body.String())
		})
	}
}

func TestSolveContractDotBlankInput(t *testing.T) {
	h := httpapi.New()
	p := catalog.Sections()[0].Puzzles[1]
	dotted := strings.ReplaceAll(p, "0", ".")
	rr := postJSON(t, h, "/v1/solve", solveReqBody(dotted))
	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 (body %q)", rr.Code, rr.Body.String())
	}
	var resp solveResponse
	strictDecode(t, rr.Body.Bytes(), &resp)
	if resp.Status != "solved" {
		t.Fatalf("status = %q, want solved", resp.Status)
	}
	if resp.Input != dotted {
		t.Errorf("input echo = %q, want the raw dotted string %q", resp.Input, dotted)
	}
	if want := oracleSolution(t, p); resp.Solution != want {
		t.Errorf("solution = %q, want %q", resp.Solution, want)
	}
}

func TestSolveContractCompleteGrid(t *testing.T) {
	h := httpapi.New()
	complete := oracleSolution(t, catalog.Sections()[0].Puzzles[0])
	rr := postJSON(t, h, "/v1/solve", solveReqBody(complete))
	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 (body %q)", rr.Code, rr.Body.String())
	}
	assertSolveShape(t, rr.Body.Bytes())
	var resp solveResponse
	strictDecode(t, rr.Body.Bytes(), &resp)
	adr14 := resp.Status == "solved" && resp.Solved && resp.Solution == complete &&
		resp.Grade == "Easy" && resp.Iterations == 0 && resp.EventCount == 0 &&
		resp.CandidateChecks == 0 && len(resp.Events) == 0
	if !adr14 {
		t.Errorf("complete-grid response violates ADR-0014: %+v", resp)
	}
	if resp.SolveTimeMs < 0 {
		t.Errorf("solveTimeMs = %v, want >= 0", resp.SolveTimeMs)
	}
}
