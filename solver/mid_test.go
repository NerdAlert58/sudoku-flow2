package solver_test

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/NerdAlert58/sudoku-flow2/solver"
)

// Each fixture grid is a singles-exhausted state: no naked or hidden single fires,
// and the named technique's only firing instance is the cheapest legal move, so it
// must be Events[0]. States and expected sets were verified against an independent
// scratch reference implementation (F-04 test-author manifest).
const (
	gridPointing = "070020503039700206020030708050470062612953487047602035090000601260390804780060309"
	gridClaiming = "709005200062009005050002009971853624835426971246791583020917350097530002503260097"
	gridNaked    = "057180623800320957230750814795841362000263795623597481040632570302975140570418230"
	gridHidden   = "053871420000369857008425310386142795090783642000956183002534978040698231839217564"
	// gridPointing with r0c2=8 filled: creates a naked single (r0c3) while the
	// pointing pattern below stays live — the single must win the pass (AC-3).
	gridSingleVsPointing = "078020503039700206020030708050470062612953487047602035090000601260390804780060309"
)

var midFixtures = []struct {
	technique    string
	grid         string
	pattern      string
	witnesses    []solver.Cell
	eliminations []solver.Elimination
}{
	{
		technique:    "locked_candidates_pointing",
		grid:         gridPointing,
		pattern:      "digit 1 in box 6 confined to column 2 at r7c2,r8c2 -> eliminate 1 from r0c2,r2c2",
		witnesses:    []solver.Cell{{Row: 7, Col: 2}, {Row: 8, Col: 2}},
		eliminations: []solver.Elimination{{Row: 0, Col: 2, Digit: 1}, {Row: 2, Col: 2, Digit: 1}},
	},
	{
		technique:    "locked_candidates_claiming",
		grid:         gridClaiming,
		pattern:      "digit 4 in row 6 confined to box 6 at r6c0,r6c2 -> eliminate 4 from r7c0",
		witnesses:    []solver.Cell{{Row: 6, Col: 0}, {Row: 6, Col: 2}},
		eliminations: []solver.Elimination{{Row: 7, Col: 0, Digit: 4}},
	},
	{
		technique:    "naked_subset",
		grid:         gridNaked,
		pattern:      "naked pair {6,9} in column 2 at r2c2,r8c2 -> eliminate 6 from r1c2, 9 from r6c2",
		witnesses:    []solver.Cell{{Row: 2, Col: 2}, {Row: 8, Col: 2}},
		eliminations: []solver.Elimination{{Row: 1, Col: 2, Digit: 6}, {Row: 6, Col: 2, Digit: 9}},
	},
	{
		technique:    "hidden_subset",
		grid:         gridHidden,
		pattern:      "hidden pair {2,4} in column 0 at r1c0,r5c0 -> eliminate 1 from r1c0, 7 from r5c0",
		witnesses:    []solver.Cell{{Row: 1, Col: 0}, {Row: 5, Col: 0}},
		eliminations: []solver.Elimination{{Row: 1, Col: 0, Digit: 1}, {Row: 5, Col: 0, Digit: 7}},
	},
}

var midTechniqueStrings = map[string]bool{
	"locked_candidates_pointing": true,
	"locked_candidates_claiming": true,
	"naked_subset":               true,
	"hidden_subset":              true,
}

type namedGrid struct {
	name string
	grid string
}

func newFixtureGrids(t *testing.T) []namedGrid {
	t.Helper()
	grids := []namedGrid{
		{"single_vs_pointing", gridSingleVsPointing},
		{"medium_seed_0", mediumSeed(t)},
	}
	for _, fx := range midFixtures {
		grids = append(grids, namedGrid{fx.technique, fx.grid})
	}
	return grids
}

// gridHasCandidate recomputes candidacy from the raw grid string, independent
// of solver internals: the cell is empty and no row/col/box peer holds d.
func gridHasCandidate(grid string, r, c, d int) bool {
	if grid[r*9+c] != '0' {
		return false
	}
	digit := byte('0' + d)
	for k := 0; k < 9; k++ {
		if grid[r*9+k] == digit || grid[k*9+c] == digit {
			return false
		}
		if grid[(r/3*3+k/3)*9+c/3*3+k%3] == digit {
			return false
		}
	}
	return true
}

func witnessesSortedRowMajor(cells []solver.Cell) bool {
	for i := 1; i < len(cells); i++ {
		a, b := cells[i-1], cells[i]
		if a.Row*9+a.Col >= b.Row*9+b.Col {
			return false
		}
	}
	return true
}

func eliminationsSorted(es []solver.Elimination) bool {
	for i := 1; i < len(es); i++ {
		a, b := es[i-1], es[i]
		if a.Row*81*9+a.Col*9+a.Digit >= b.Row*81*9+b.Col*9+b.Digit {
			return false
		}
	}
	return true
}

func TestSolve_MidTechnique_FiresWithExpectedEvent(t *testing.T) {
	for _, fx := range midFixtures {
		t.Run(fx.technique, func(t *testing.T) {
			res := solver.Solve(mustParse(t, fx.grid))
			if len(res.Events) == 0 {
				t.Fatalf("no events (status %q): %s must fire as the cheapest legal move; pattern: %s", res.Status, fx.technique, fx.pattern)
			}
			ev := res.Events[0]
			if ev.Technique != fx.technique {
				t.Fatalf("Events[0].Technique = %q, want %q; pattern: %s", ev.Technique, fx.technique, fx.pattern)
			}
			if ev.Seq != 1 {
				t.Errorf("Events[0].Seq = %d, want 1", ev.Seq)
			}
			if ev.Placement != nil {
				t.Errorf("Placement = %+v, want nil: mid-ladder techniques only eliminate", ev.Placement)
			}
			if !reflect.DeepEqual(ev.WitnessCells, fx.witnesses) {
				t.Errorf("WitnessCells = %v, want %v; pattern: %s", ev.WitnessCells, fx.witnesses, fx.pattern)
			}
			if !reflect.DeepEqual(ev.Eliminations, fx.eliminations) {
				t.Errorf("Eliminations = %v, want %v; pattern: %s", ev.Eliminations, fx.eliminations, fx.pattern)
			}
			if ev.GridAfter != fx.grid {
				t.Errorf("GridAfter = %q, want the unchanged input grid: eliminations never place digits", ev.GridAfter)
			}
			for _, e := range fx.eliminations {
				if !gridHasCandidate(fx.grid, e.Row, e.Col, e.Digit) {
					t.Errorf("fixture defect: expected elimination %+v is not a live candidate of the input state", e)
				}
			}
		})
	}
}

func TestSolve_MidEvents_CanonicalSerialization(t *testing.T) {
	richEvents := 0
	for _, ng := range newFixtureGrids(t) {
		res := solver.Solve(mustParse(t, ng.grid))
		for _, ev := range res.Events {
			if !midTechniqueStrings[ev.Technique] {
				continue
			}
			if ev.Placement != nil {
				t.Errorf("%s event %d: Placement = %+v, want nil on %s", ng.name, ev.Seq, ev.Placement, ev.Technique)
			}
			if len(ev.Eliminations) == 0 {
				t.Errorf("%s event %d: Eliminations empty on %s; mid events must be productive", ng.name, ev.Seq, ev.Technique)
			}
			if !witnessesSortedRowMajor(ev.WitnessCells) {
				t.Errorf("%s event %d: WitnessCells %v not strictly sorted row-major (ADR-0007)", ng.name, ev.Seq, ev.WitnessCells)
			}
			if !eliminationsSorted(ev.Eliminations) {
				t.Errorf("%s event %d: Eliminations %v not strictly sorted row-major-then-digit (ADR-0007)", ng.name, ev.Seq, ev.Eliminations)
			}
			if len(ev.WitnessCells) >= 2 && len(ev.Eliminations) >= 2 {
				richEvents++
			}
		}
	}
	if richEvents == 0 {
		t.Fatal("no mid-ladder event carried >=2 witnesses and >=2 eliminations; the ordering assertions above are vacuous (the pointing fixture must produce one)")
	}
}

func TestSolve_MidLadderPuzzle_GradesMedium(t *testing.T) {
	cases := []namedGrid{
		{"medium_seed_0", mediumSeed(t)},
		{"pointing_fixture", gridPointing},
	}
	for _, c := range cases {
		res := solver.Solve(mustParse(t, c.grid))
		if res.Status != "solved" {
			t.Errorf("%s: Status = %q, want solved (verified solvable with techniques 1-6)", c.name, res.Status)
			continue
		}
		if res.Grade != "Medium" {
			t.Errorf("%s: Grade = %q, want Medium (hardest forced technique is mid-ladder)", c.name, res.Grade)
		}
		assertRuleConformantSolution(t, c.grid, res.Solution.String())
	}
}

func TestSolve_CheapestFirst_SingleBeforeMidLadder(t *testing.T) {
	res := solver.Solve(mustParse(t, gridSingleVsPointing))
	if len(res.Events) == 0 {
		t.Fatalf("no events (status %q): a naked single is available at r0c3", res.Status)
	}
	if got := res.Events[0].Technique; got != "naked_single" {
		t.Fatalf("Events[0].Technique = %q, want naked_single: the live pointing pattern (digit 1, box 6, column 2) must not preempt an available single", got)
	}
}

func TestSolve_NewFixtures_DeterminismDoubleRun(t *testing.T) {
	for _, ng := range newFixtureGrids(t) {
		first := solver.Solve(mustParse(t, ng.grid))
		second := solver.Solve(mustParse(t, ng.grid))
		if !bytes.Equal(eventsJSON(t, first.Events), eventsJSON(t, second.Events)) {
			t.Errorf("%s: events JSON differs between two consecutive in-process solves", ng.name)
		}
		if first.Iterations != second.Iterations ||
			first.EventCount != second.EventCount ||
			first.CandidateChecks != second.CandidateChecks {
			t.Errorf("%s: counters differ between runs: (%d %d %d) vs (%d %d %d)",
				ng.name, first.Iterations, first.EventCount, first.CandidateChecks,
				second.Iterations, second.EventCount, second.CandidateChecks)
		}
	}
}

func TestSolve_Corpus_CapSixDeterminismAndSolvedCount(t *testing.T) {
	total, solved, originalSolved := 0, 0, 0
	for si, sec := range readCorpusSections(t) {
		for pi, seed := range sec {
			total++
			first := solver.Solve(mustParse(t, seed))
			second := solver.Solve(mustParse(t, seed))
			if !bytes.Equal(eventsJSON(t, first.Events), eventsJSON(t, second.Events)) ||
				first.Iterations != second.Iterations ||
				first.EventCount != second.EventCount ||
				first.CandidateChecks != second.CandidateChecks {
				t.Fatalf("section %d puzzle %d: double-run mismatch", si, pi)
			}
			if first.Status != "solved" {
				continue
			}
			solved++
			if si == 0 {
				originalSolved++
			}
		}
	}
	if originalSolved != 25 {
		t.Errorf("ORIGINAL puzzles solved = %d, want 25", originalSolved)
	}
	t.Logf("corpus puzzles solved with the ladder capped at technique 6: %d/%d", solved, total)
}
