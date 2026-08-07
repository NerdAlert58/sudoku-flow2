package solver_test

import (
	"reflect"
	"testing"

	"github.com/NerdAlert58/sudoku-flow2/solver"
)

// Each fixture grid is a singles-and-mid-exhausted state: no technique at
// ladder positions 1-6 fires, no upper technique cheaper than the named one
// fires, and the named technique's first canonical instance (per the F-05
// pinned enumeration orders, see the fixture patterns) must be Events[0].
// States, witnesses, and eliminations were verified against an independent
// scratch reference implementing all 13 techniques; every expected
// elimination was proven sound by exact brute force (placing the eliminated
// digit admits no solution).
//
// Pinned enumeration orders (ADR-0007 extensions; binding on the builder):
//   - Fish (x_wing k=2, swordfish k=3, jellyfish k=4): orientation outermost
//     (rows-base first, then cols-base), digits ascending, base-line
//     k-combinations lexicographic over ascending line indexes. A base line
//     holds >=2 candidates of the digit; PLAIN fish only - the union of
//     candidate cross-lines over the k base lines must be EXACTLY k (any fin
//     breaks this and the combination is skipped, AC-4). Witnesses: every
//     candidate cell in the base lines. Eliminations: the digit from cover
//     lines outside the base lines.
//   - xy_wing: pivot (bivalue) cells row-major; pincer pairs (A,B)
//     lexicographic over the pivot's row-major-ascending peer list; pincers
//     bivalue, each sharing a different pivot digit, sharing digit Z with
//     each other and not with the pivot; eliminate Z from cells seeing both
//     pincers. Witnesses: pivot + both pincers.
//   - xyz_wing: pivot (trivalue) row-major; pincer pairs lexicographic over
//     row-major peers; pincers bivalue subsets of the pivot sharing exactly
//     one digit Z; eliminate Z from cells seeing pivot and both pincers.
//     Witnesses: pivot + both pincers.
//   - w_wing: bivalue cell A row-major, then matching bivalue B (identical
//     candidate pair {X,Y}, NOT a peer of A) row-major after A, then link
//     digit X ascending over the pair, then strong-link units canonical 0-26
//     (a unit whose only two X-candidates W1, W2 are distinct from A and B,
//     one seeing A and the other seeing B); eliminate the non-link digit Y
//     from every cell other than A and B seeing both A and B. Witnesses:
//     A, B, W1, W2.
//   - simple_colouring: digits ascending; conjugate graph per digit (units
//     with exactly two candidate cells are edges); components seeded from
//     the lowest uncoloured row-major linked cell; BFS FIFO, neighbours in
//     ascending cell order, seed = colour 0, first assignment wins. Per
//     component WRAP is checked before TRAP; wrap checks colour 0 before
//     colour 1. Wrap: a colour class with two cells sharing a unit is false;
//     eliminate the digit from every cell of that colour. Trap: every
//     candidate cell outside the component seeing both colours loses the
//     digit, all targets in ONE event. Witnesses: every coloured cell of the
//     component (both colours).
//
// All witnesses serialize row-major; eliminations row-major-then-digit.
const (
	gridXWing     = "000029560030501800000060000000000008083000100000030700308000010000080004000007680"
	gridSwordfish = "274650803961783542358204076006005200405120637032006050849567321527000060613942785"
	gridJellyfish = "007000000301009724000000900006001200000000000102805403723008105000302000809057302"
	gridXYWing    = "542063010003800245810542060386050000274380150951070080008025034430608500025030090"
	gridXYZWing   = "700006000325791684000000300030008000080010000000004000800000436073600020000083000"
	gridWWing     = "274318965310960020960420001590740083407800500800500700183659000659274008742183659"
	gridSCTrap    = "164007080935082017287001000356128070412709850879054001608013705501076308703805100"
	gridSCWrap    = "906347502537802904402050030328574691654291378179683425891705240763428159245000000"
)

var upperFixtures = []struct {
	technique    string
	grid         string
	pattern      string
	witnesses    []solver.Cell
	eliminations []solver.Elimination
}{
	{
		technique: "x_wing",
		grid:      gridXWing,
		pattern:   "digit 3 rows-base x-wing, base rows {0,8} cover cols {3,8} -> eliminate 3 from r2c3,r2c8,r7c3 (unique instance)",
		witnesses: []solver.Cell{{Row: 0, Col: 3}, {Row: 0, Col: 8}, {Row: 8, Col: 3}, {Row: 8, Col: 8}},
		eliminations: []solver.Elimination{
			{Row: 2, Col: 3, Digit: 3}, {Row: 2, Col: 8, Digit: 3}, {Row: 7, Col: 3, Digit: 3},
		},
	},
	{
		technique: "swordfish",
		grid:      gridSwordfish,
		pattern:   "digit 9 rows-base swordfish, base rows {2,5,7} cover cols {4,6,8} -> eliminate 9 from r3c4,r3c8; 2 instances exist, rows-base enumerates first",
		witnesses: []solver.Cell{
			{Row: 2, Col: 4}, {Row: 2, Col: 6}, {Row: 5, Col: 4}, {Row: 5, Col: 6},
			{Row: 5, Col: 8}, {Row: 7, Col: 6}, {Row: 7, Col: 8},
		},
		eliminations: []solver.Elimination{
			{Row: 3, Col: 4, Digit: 9}, {Row: 3, Col: 8, Digit: 9},
		},
	},
	{
		technique: "jellyfish",
		grid:      gridJellyfish,
		pattern:   "digit 6 rows-base jellyfish, base rows {1,5,6,8} cover cols {1,3,4,7} -> 14 eliminations; the cols-base complement is the 2nd instance, rows-base enumerates first",
		witnesses: []solver.Cell{
			{Row: 1, Col: 1}, {Row: 1, Col: 3}, {Row: 1, Col: 4}, {Row: 5, Col: 4},
			{Row: 5, Col: 7}, {Row: 6, Col: 3}, {Row: 6, Col: 4}, {Row: 6, Col: 7},
			{Row: 8, Col: 1}, {Row: 8, Col: 3}, {Row: 8, Col: 7},
		},
		eliminations: []solver.Elimination{
			{Row: 0, Col: 1, Digit: 6}, {Row: 0, Col: 3, Digit: 6}, {Row: 0, Col: 4, Digit: 6},
			{Row: 0, Col: 7, Digit: 6}, {Row: 2, Col: 1, Digit: 6}, {Row: 2, Col: 3, Digit: 6},
			{Row: 2, Col: 4, Digit: 6}, {Row: 2, Col: 7, Digit: 6}, {Row: 4, Col: 3, Digit: 6},
			{Row: 4, Col: 4, Digit: 6}, {Row: 4, Col: 7, Digit: 6}, {Row: 7, Col: 1, Digit: 6},
			{Row: 7, Col: 4, Digit: 6}, {Row: 7, Col: 7, Digit: 6},
		},
	},
	{
		technique: "xy_wing",
		grid:      gridXYWing,
		pattern:   "pivot r6c1, pincers r6c6 and r7c2, Z=7 -> eliminate 7 from r6c0,r7c7,r7c8 (unique instance)",
		witnesses: []solver.Cell{{Row: 6, Col: 1}, {Row: 6, Col: 6}, {Row: 7, Col: 2}},
		eliminations: []solver.Elimination{
			{Row: 6, Col: 0, Digit: 7}, {Row: 7, Col: 7, Digit: 7}, {Row: 7, Col: 8, Digit: 7},
		},
	},
	{
		technique:    "xyz_wing",
		grid:         gridXYZWing,
		pattern:      "pivot r2c4 (trivalue), pincers r2c5 and r7c4, Z=5 -> eliminate 5 from r0c4 (unique instance)",
		witnesses:    []solver.Cell{{Row: 2, Col: 4}, {Row: 2, Col: 5}, {Row: 7, Col: 4}},
		eliminations: []solver.Elimination{{Row: 0, Col: 4, Digit: 5}},
	},
	{
		technique: "w_wing",
		grid:      gridWWing,
		pattern:   "bivalue pair A=r1c8, B=r6c7 both {4,7}, strong link on 4 in row 5 at W1=r5c7 (sees B), W2=r5c8 (sees A) -> eliminate 7 from r2c7,r6c8; 3 instances exist, this is first canonical",
		witnesses: []solver.Cell{{Row: 1, Col: 8}, {Row: 5, Col: 7}, {Row: 5, Col: 8}, {Row: 6, Col: 7}},
		eliminations: []solver.Elimination{
			{Row: 2, Col: 7, Digit: 7}, {Row: 6, Col: 8, Digit: 7},
		},
	},
	{
		technique: "simple_colouring",
		grid:      gridSCTrap,
		pattern:   "colour TRAP on digit 3: conjugate component seeded at r2c7 = {r2c7,r4c4,r4c8,r5c3,r5c7}; r2c4 sees both colours -> loses 3 (only colouring instance)",
		witnesses: []solver.Cell{
			{Row: 2, Col: 7}, {Row: 4, Col: 4}, {Row: 4, Col: 8}, {Row: 5, Col: 3}, {Row: 5, Col: 7},
		},
		eliminations: []solver.Elimination{{Row: 2, Col: 4, Digit: 3}},
	},
	{
		technique: "simple_colouring",
		grid:      gridSCWrap,
		pattern:   "colour WRAP on digit 6: component seeded at r1c4 = {r1c4,r1c7,r2c5,r2c8,r8c5,r8c7}; colour-0 class {r1c4,r2c8,r8c5,r8c7} holds r8c5+r8c7 in row 8 -> whole class loses 6",
		witnesses: []solver.Cell{
			{Row: 1, Col: 4}, {Row: 1, Col: 7}, {Row: 2, Col: 5}, {Row: 2, Col: 8},
			{Row: 8, Col: 5}, {Row: 8, Col: 7},
		},
		eliminations: []solver.Elimination{
			{Row: 1, Col: 4, Digit: 6}, {Row: 2, Col: 8, Digit: 6},
			{Row: 8, Col: 5, Digit: 6}, {Row: 8, Col: 7, Digit: 6},
		},
	},
}

func TestSolve_UpperTechnique_FiresWithExpectedEvent(t *testing.T) {
	for _, fx := range upperFixtures {
		t.Run(fx.technique+"_"+fx.grid[:8], func(t *testing.T) {
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
				t.Errorf("Placement = %+v, want nil: upper-ladder techniques only eliminate", ev.Placement)
			}
			if !reflect.DeepEqual(ev.WitnessCells, fx.witnesses) {
				t.Errorf("WitnessCells = %v, want %v; pattern: %s", ev.WitnessCells, fx.witnesses, fx.pattern)
			}
			if !reflect.DeepEqual(ev.Eliminations, fx.eliminations) {
				t.Errorf("Eliminations = %v, want %v; pattern: %s", ev.Eliminations, fx.eliminations, fx.pattern)
			}
			if !witnessesSortedRowMajor(ev.WitnessCells) {
				t.Errorf("WitnessCells %v not strictly sorted row-major (ADR-0007)", ev.WitnessCells)
			}
			if !eliminationsSorted(ev.Eliminations) {
				t.Errorf("Eliminations %v not strictly sorted row-major-then-digit (ADR-0007)", ev.Eliminations)
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

// AC-2 fish clause: base and cover sets are structurally exact - k distinct
// base rows, k distinct cover columns (rows-base fixtures), witnesses cover
// every base-row candidate, and each elimination sits in a cover column
// outside the base rows.
func TestSolve_UpperFish_StructuralPattern(t *testing.T) {
	fishSizes := map[string]int{"x_wing": 2, "swordfish": 3, "jellyfish": 4}
	tested := 0
	for _, fx := range upperFixtures {
		k, isFish := fishSizes[fx.technique]
		if !isFish {
			continue
		}
		tested++
		baseRows, coverCols := map[int]bool{}, map[int]bool{}
		for _, w := range fx.witnesses {
			baseRows[w.Row] = true
			coverCols[w.Col] = true
		}
		if len(baseRows) != k || len(coverCols) != k {
			t.Errorf("%s: witnesses span %d rows x %d cols, want %d x %d", fx.technique, len(baseRows), len(coverCols), k, k)
		}
		if min, max := k*2, k*k; len(fx.witnesses) < min || len(fx.witnesses) > max {
			t.Errorf("%s: %d witness cells, want %d..%d base cells", fx.technique, len(fx.witnesses), min, max)
		}
		digit := fx.eliminations[0].Digit
		for _, e := range fx.eliminations {
			if e.Digit != digit {
				t.Errorf("%s: elimination digit %d differs from fish digit %d", fx.technique, e.Digit, digit)
			}
			if baseRows[e.Row] {
				t.Errorf("%s: elimination %+v sits in a base row", fx.technique, e)
			}
			if !coverCols[e.Col] {
				t.Errorf("%s: elimination %+v outside the cover columns", fx.technique, e)
			}
		}
	}
	if tested != 3 {
		t.Fatalf("structural fish check covered %d fixtures, want 3", tested)
	}
}

// AC-4: this state's ONLY x-wing-shaped patterns are finned (digit 7, rows
// 7+8: r8 has spots {c1,c6}, r7 has {c1,c6} plus fin r7c8 - a fin-blind
// detector would eliminate 7 from r1c6,r4c6; further finned shapes exist on
// digit 8 and in columns). Verified by the scratch reference: NO technique
// at any ladder position fires here, so the solve must stall with zero
// events - any fish event means finned/sashimi matching leaked in.
const gridFinnedOnly = "713865942596241000824739651187953264659124000432678195361487529205396010908512036"

func TestSolve_FinnedFishOnly_NeverEmitsFishEvent(t *testing.T) {
	res := solver.Solve(mustParse(t, gridFinnedOnly))
	for _, ev := range res.Events {
		switch ev.Technique {
		case "x_wing", "swordfish", "jellyfish":
			t.Fatalf("event %d: %s fired on a finned-only pattern (plain fish require cover union == k)", ev.Seq, ev.Technique)
		}
	}
	if res.Status != "stalled" {
		t.Fatalf("Status = %q, want stalled: no ladder technique fires on this state", res.Status)
	}
	if len(res.Events) != 0 {
		t.Fatalf("EventCount = %d, want 0: no ladder technique fires on this state", len(res.Events))
	}
}

// AC-3 structural half: the wrap event's shape is the positive-form
// justification - the witness set is the whole conjugate component, the
// eliminated cells are a subset of it (the false colour class), and that
// class contains two cells sharing a unit (the wrap evidence). The
// no-assume-propagate-revert code shape is the builder's burden; F-06 replay
// is the mechanical backstop.
func TestSolve_ColourWrap_PositiveFormEventShape(t *testing.T) {
	res := solver.Solve(mustParse(t, gridSCWrap))
	if len(res.Events) == 0 {
		t.Fatalf("no events (status %q): the colour wrap is the only legal move on this state", res.Status)
	}
	ev := res.Events[0]
	if ev.Technique != "simple_colouring" {
		t.Fatalf("Events[0].Technique = %q, want simple_colouring", ev.Technique)
	}
	witness := map[solver.Cell]bool{}
	for _, w := range ev.WitnessCells {
		witness[w] = true
	}
	sameUnit := func(a, b solver.Elimination) bool {
		return a.Row == b.Row || a.Col == b.Col || (a.Row/3 == b.Row/3 && a.Col/3 == b.Col/3)
	}
	evidence := false
	for i, e := range ev.Eliminations {
		if !witness[solver.Cell{Row: e.Row, Col: e.Col}] {
			t.Errorf("wrap elimination %+v is not a coloured chain cell", e)
		}
		for _, f := range ev.Eliminations[i+1:] {
			if sameUnit(e, f) {
				evidence = true
			}
		}
	}
	if !evidence {
		t.Error("no two eliminated cells share a unit: the wrap's shared-unit evidence is missing from the event")
	}
	if len(ev.Eliminations) >= len(ev.WitnessCells) {
		t.Errorf("%d eliminations >= %d witnesses: the surviving colour class must stay in the witness set", len(ev.Eliminations), len(ev.WitnessCells))
	}
}

// AC-5: grading bands per the frozen table - hardest fired in positions 7-10
// grades Hard, 11-13 grades Expert. Both grids solve completely; ceilings
// verified with the scratch reference (xy_wing fixture: hardest = xy_wing;
// w_wing fixture: hardest = w_wing).
func TestSolve_UpperLadder_GradingBands(t *testing.T) {
	cases := []struct {
		name, grid, grade string
	}{
		{"hard_ceiling_xy_wing", gridXYWing, "Hard"},
		{"expert_ceiling_w_wing", gridWWing, "Expert"},
	}
	for _, c := range cases {
		res := solver.Solve(mustParse(t, c.grid))
		if res.Status != "solved" {
			t.Errorf("%s: Status = %q, want solved (verified solvable by the full ladder)", c.name, res.Status)
			continue
		}
		if res.Grade != c.grade {
			t.Errorf("%s: Grade = %q, want %q", c.name, res.Grade, c.grade)
		}
		assertRuleConformantSolution(t, c.grid, res.Solution.String())
	}
}

// F-04 verifier debt (mutation M4): pins naked-subset enumeration with k
// ascending OUTERMOST. A naked triple {r6c4,r6c6,r6c8} lives in row 6
// (unit 6) with eliminations, and a naked pair {r1c7,r8c7} of {4,7} lives in
// column 7 (unit 16). Units-outermost enumeration would fire the earlier
// unit's triple; the frozen k-outermost order fires the pair. No single,
// pointing, or claiming fires on this state (scratch-reference verified).
const gridK2LateK3Early = "068027091190685302020019080912736050876541923435892167250004000040008200681203009"

func TestSolve_NakedSubset_KAscendingOutermost(t *testing.T) {
	res := solver.Solve(mustParse(t, gridK2LateK3Early))
	if len(res.Events) == 0 {
		t.Fatalf("no events (status %q): the naked pair in column 7 must fire", res.Status)
	}
	ev := res.Events[0]
	if ev.Technique != "naked_subset" {
		t.Fatalf("Events[0].Technique = %q, want naked_subset", ev.Technique)
	}
	wantW := []solver.Cell{{Row: 1, Col: 7}, {Row: 8, Col: 7}}
	wantE := []solver.Elimination{{Row: 6, Col: 7, Digit: 7}, {Row: 7, Col: 7, Digit: 7}}
	if !reflect.DeepEqual(ev.WitnessCells, wantW) {
		t.Errorf("WitnessCells = %v, want the k=2 pair %v (k ascending outermost), not the earlier-unit k=3 triple", ev.WitnessCells, wantW)
	}
	if !reflect.DeepEqual(ev.Eliminations, wantE) {
		t.Errorf("Eliminations = %v, want %v", ev.Eliminations, wantE)
	}
}

// F-05 verifier debt (mutation M2a): pins WRAP checked before TRAP within one
// component. Digit 8 has exactly ONE colouring component on this state (so it
// is trivially the first canonical instance): colour 0 = {r0c0,r2c3,r8c4},
// colour 1 = {r0c4,r2c1,r7c0,r7c3}. It carries BOTH conclusions at once -
// wrap: r7c0+r7c3 (colour 1) share row 7, so the whole colour-1 class is
// false; trap: the outside cell r8c1 holds candidate 8 and sees both colours
// (r8c4 in row 8, r2c1 in column 1, r7c0 in box 6). The frozen priority emits
// the wrap; a trap-first mutant emits [r8c1#8] instead. No cheaper ladder
// technique fires here, and each wrap elimination was proven sound by exact
// brute force (scratch-reference verified; the state is solvable).
const gridSCWrapTrap = "060503102152900008304021569436198005798254613215000984603479051000015006501602097"

func TestSolve_ColourWrapPriority_WrapBeatsTrapOnSameComponent(t *testing.T) {
	res := solver.Solve(mustParse(t, gridSCWrapTrap))
	if len(res.Events) == 0 {
		t.Fatalf("no events (status %q): the digit-8 colouring component must fire", res.Status)
	}
	ev := res.Events[0]
	if ev.Technique != "simple_colouring" {
		t.Fatalf("Events[0].Technique = %q, want simple_colouring", ev.Technique)
	}
	wantW := []solver.Cell{
		{Row: 0, Col: 0}, {Row: 0, Col: 4}, {Row: 2, Col: 1}, {Row: 2, Col: 3},
		{Row: 7, Col: 0}, {Row: 7, Col: 3}, {Row: 8, Col: 4},
	}
	wantE := []solver.Elimination{
		{Row: 0, Col: 4, Digit: 8}, {Row: 2, Col: 1, Digit: 8},
		{Row: 7, Col: 0, Digit: 8}, {Row: 7, Col: 3, Digit: 8},
	}
	if !reflect.DeepEqual(ev.WitnessCells, wantW) {
		t.Errorf("WitnessCells = %v, want the whole component %v", ev.WitnessCells, wantW)
	}
	if !reflect.DeepEqual(ev.Eliminations, wantE) {
		t.Errorf("Eliminations = %v, want the false colour-1 class %v (wrap before trap: a trap-first mutant emits the single trap target r8c1 instead)", ev.Eliminations, wantE)
	}
	for _, e := range wantE {
		if !gridHasCandidate(gridSCWrapTrap, e.Row, e.Col, e.Digit) {
			t.Errorf("fixture defect: expected elimination %+v is not a live candidate of the input state", e)
		}
	}
	if !gridHasCandidate(gridSCWrapTrap, 8, 1, 8) {
		t.Error("fixture defect: the trap target r8c1 lost candidate 8, so the state no longer discriminates wrap-vs-trap priority")
	}
}
