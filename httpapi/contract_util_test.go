package httpapi_test

// F-10 shared helpers. Wire truth: ARCHITECTURE.md §C1, ADR-0004/0006/0007/
// 0009/0014, AUDIT.md A5/A7, PRD §API contract. Field-order slices below ARE
// the frozen contract (encoding/json preserves struct order, AUDIT A5).

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/NerdAlert58/sudoku-flow2/catalog"
	"github.com/NerdAlert58/sudoku-flow2/oracle"
	"github.com/NerdAlert58/sudoku-flow2/solver"
)

const oneMiB = 1 << 20

var solveFieldOrder = []string{
	"apiVersion", "input", "status", "solved", "solution",
	"iterations", "eventCount", "candidateChecks", "solveTimeMs", "grade", "events",
}

var batchFieldOrder = []string{"apiVersion", "results", "solvedCount", "total"}

var batchItemFieldOrder = []string{"puzzle", "solved", "solveTimeMs", "iterations", "hardestTechnique"}

var generateFieldOrder = []string{"puzzle", "difficulty", "grade"}

// ladderOrder mirrors PRD §Domain context (the frozen 13-technique ladder);
// index = ladder position for ADR-0014 hardestTechnique expectations.
var ladderOrder = []string{
	"naked_single", "hidden_single",
	"locked_candidates_pointing", "locked_candidates_claiming",
	"naked_subset", "hidden_subset",
	"x_wing", "swordfish", "jellyfish", "xy_wing",
	"xyz_wing", "w_wing", "simple_colouring",
}

// bandOf pins PRD "inputs map easy→Easy, medium→Medium, hard→Hard,
// expert→Expert grades".
var bandOf = map[string]string{
	"easy": "Easy", "medium": "Medium", "hard": "Hard", "expert": "Expert",
}

type solveResponse struct {
	APIVersion      string         `json:"apiVersion"`
	Input           string         `json:"input"`
	Status          string         `json:"status"`
	Solved          bool           `json:"solved"`
	Solution        string         `json:"solution"`
	Iterations      int            `json:"iterations"`
	EventCount      int            `json:"eventCount"`
	CandidateChecks int            `json:"candidateChecks"`
	SolveTimeMs     float64        `json:"solveTimeMs"`
	Grade           string         `json:"grade"`
	Events          []solver.Event `json:"events"`
}

type batchItem struct {
	Puzzle           string  `json:"puzzle"`
	Solved           bool    `json:"solved"`
	SolveTimeMs      float64 `json:"solveTimeMs"`
	Iterations       int     `json:"iterations"`
	HardestTechnique string  `json:"hardestTechnique"`
}

type batchResponse struct {
	APIVersion  string      `json:"apiVersion"`
	Results     []batchItem `json:"results"`
	SolvedCount int         `json:"solvedCount"`
	Total       int         `json:"total"`
}

type generateResponse struct {
	Puzzle     string `json:"puzzle"`
	Difficulty string `json:"difficulty"`
	Grade      string `json:"grade"`
}

func postJSON(t *testing.T, h http.Handler, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	return rr
}

func trimmedBody(rr *httptest.ResponseRecorder) string {
	return strings.TrimSuffix(rr.Body.String(), "\n")
}

func strictDecode(t *testing.T, data []byte, v any) {
	t.Helper()
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	if err := dec.Decode(v); err != nil {
		t.Fatalf("strict decode into %T: %v (body %q)", v, err, data)
	}
}

// jsonKeys returns the top-level object keys of obj in wire order.
func jsonKeys(t *testing.T, obj []byte) []string {
	t.Helper()
	dec := json.NewDecoder(bytes.NewReader(obj))
	if tok, err := dec.Token(); err != nil || tok != json.Delim('{') {
		t.Fatalf("not a JSON object: tok=%v err=%v (body %q)", tok, err, obj)
	}
	var keys []string
	for dec.More() {
		tok, err := dec.Token()
		if err != nil {
			t.Fatalf("key token: %v", err)
		}
		key, ok := tok.(string)
		if !ok {
			t.Fatalf("key token %v is not a string", tok)
		}
		keys = append(keys, key)
		var raw json.RawMessage
		if err := dec.Decode(&raw); err != nil {
			t.Fatalf("value of %q: %v", key, err)
		}
	}
	return keys
}

func assertKeyOrder(t *testing.T, obj []byte, want []string) {
	t.Helper()
	got := jsonKeys(t, obj)
	if fmt.Sprint(got) != fmt.Sprint(want) {
		t.Errorf("field order = %v, want %v", got, want)
	}
}

// assertSolveShape pins the exact frozen key set+order and that events is a
// JSON array, never null (ADR-0004 events:[]).
func assertSolveShape(t *testing.T, body []byte) {
	t.Helper()
	assertKeyOrder(t, body, solveFieldOrder)
	var m map[string]any
	if err := json.Unmarshal(body, &m); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if m["events"] == nil {
		t.Error(`events is null; contract requires a JSON array (ADR-0004)`)
	}
}

func corpus(t *testing.T) []string {
	t.Helper()
	var all []string
	for _, sec := range catalog.Sections() {
		all = append(all, sec.Puzzles...)
	}
	if len(all) != 55 {
		t.Fatalf("corpus has %d puzzles, want 55", len(all))
	}
	return all
}

func mustParseGrid(t *testing.T, p string) solver.Grid {
	t.Helper()
	g, err := solver.Parse(p)
	if err != nil {
		t.Fatalf("fixture %q does not parse: %v", p, err)
	}
	return g
}

func oracleSolution(t *testing.T, p string) string {
	t.Helper()
	sol, count := oracle.Solve(mustParseGrid(t, p))
	if count != 1 {
		t.Fatalf("fixture %q has %d oracle solutions, want 1", p, count)
	}
	return sol.String()
}

// hardestFired computes the ADR-0014 hardestTechnique expectation: the event
// string of the highest-ladder-position technique that fired, "" when none.
func hardestFired(t *testing.T, events []solver.Event) string {
	t.Helper()
	best := -1
	for _, ev := range events {
		pos := -1
		for i, name := range ladderOrder {
			if name == ev.Technique {
				pos = i
				break
			}
		}
		if pos < 0 {
			t.Fatalf("event technique %q is not in the frozen ladder", ev.Technique)
		}
		if pos > best {
			best = pos
		}
	}
	if best < 0 {
		return ""
	}
	return ladderOrder[best]
}

// assertNoInternals is the SECURITY.md F-9 hygiene floor for every error body.
func assertNoInternals(t *testing.T, body string) {
	t.Helper()
	for _, leak := range []string{"runtime", "goroutine", ".go:", "panic"} {
		if strings.Contains(body, leak) {
			t.Errorf("error body leaks internals (%q): %q", leak, body)
		}
	}
}

func solveReqBody(p string) string {
	return fmt.Sprintf(`{"puzzle":%q}`, p)
}

func batchReqBody(t *testing.T, puzzles []string) string {
	t.Helper()
	b, err := json.Marshal(struct {
		Puzzles []string `json:"puzzles"`
	}{puzzles})
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}
