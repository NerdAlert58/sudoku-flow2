package solver_test

import (
	"encoding/json"
	"testing"

	"github.com/NerdAlert58/sudoku-flow2/solver"
)

var (
	_ func(string) (solver.Grid, error)    = solver.Parse
	_ func(solver.Grid) solver.SolveResult = solver.Solve
	_ error                                = solver.ErrBadLength
	_ error                                = solver.ErrBadChar
	_ error                                = solver.ErrDuplicateGiven
)

var _ = func() {
	var g solver.Grid
	var _ string = g.String()
	var r solver.SolveResult
	var _ string = r.Status
	var _ solver.Grid = r.Solution
	var _ []solver.Event = r.Events
	var _ int = r.Iterations
	var _ int = r.EventCount
	var _ int = r.CandidateChecks
	var _ string = r.Grade
	var _ solver.Event = solver.Event{
		Seq:          1,
		Technique:    "naked_single",
		WitnessCells: []solver.Cell{{Row: 0, Col: 0}},
		Placement:    &solver.Placement{Row: 0, Col: 0, Digit: 9},
		Eliminations: []solver.Elimination{{Row: 0, Col: 0, Digit: 9}},
		GridAfter:    "",
	}
}

func TestEvent_JSONShape_FrozenTags(t *testing.T) {
	res := solver.Solve(mustParse(t, originalSeeds(t)[0]))
	if len(res.Events) == 0 {
		t.Fatal("no events to inspect")
	}
	var evs []map[string]json.RawMessage
	if err := json.Unmarshal(eventsJSON(t, res.Events[:1]), &evs); err != nil {
		t.Fatalf("unmarshal events: %v", err)
	}
	ev := evs[0]
	for _, key := range []string{"seq", "technique", "witnessCells", "placement", "gridAfter"} {
		if _, ok := ev[key]; !ok {
			t.Errorf("event JSON missing key %q", key)
		}
	}
	if _, ok := ev["eliminations"]; ok {
		t.Error(`event JSON carries "eliminations" on a singles event; want omitempty to drop it`)
	}
	if len(ev) != 5 {
		t.Errorf("event JSON has %d keys, want exactly 5: seq, technique, witnessCells, placement, gridAfter", len(ev))
	}
	var witness []map[string]int
	if err := json.Unmarshal(ev["witnessCells"], &witness); err != nil {
		t.Fatalf("witnessCells: %v", err)
	}
	if len(witness) == 0 {
		t.Fatal("witnessCells empty")
	}
	for _, c := range witness {
		if _, ok := c["row"]; !ok {
			t.Error(`witness cell missing "row"`)
		}
		if _, ok := c["col"]; !ok {
			t.Error(`witness cell missing "col"`)
		}
		if len(c) != 2 {
			t.Errorf("witness cell has %d keys, want exactly 2: row, col", len(c))
		}
	}
	var placement map[string]int
	if err := json.Unmarshal(ev["placement"], &placement); err != nil {
		t.Fatalf("placement: %v", err)
	}
	for _, key := range []string{"row", "col", "digit"} {
		if _, ok := placement[key]; !ok {
			t.Errorf("placement missing key %q", key)
		}
	}
	if len(placement) != 3 {
		t.Errorf("placement has %d keys, want exactly 3: row, col, digit", len(placement))
	}
}
