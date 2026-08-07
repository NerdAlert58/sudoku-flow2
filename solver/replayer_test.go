package solver_test

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/NerdAlert58/sudoku-flow2/oracle"
	"github.com/NerdAlert58/sudoku-flow2/solver"
)

// AC-3 — the pinned rejection contract of oracle.ReplayVerify (ADR-0013):
// events are checked in order and the FIRST violation is returned, naming its
// seq ("seq N"), before any whole-result check (final grid == oracle solution).
// Elimination-event checks run in the order candidate-liveness, oracle-truth,
// witness structure; placement-event checks run oracle-equality first. The
// four errors.Is sentinels below are the pinned rejection surface; each case
// here violates exactly one rule, so the sentinel identity is deterministic.
var replaySentinels = []error{
	oracle.ErrSingleAvailable,
	oracle.ErrPlacementNotOracle,
	oracle.ErrEliminationNotCandidate,
	oracle.ErrEliminationIsTruth,
}

func assertRejection(t *testing.T, err, want error, seq int) {
	t.Helper()
	if err == nil {
		t.Fatal("ReplayVerify = nil, want a rejection")
	}
	if !errors.Is(err, want) {
		t.Fatalf("ReplayVerify = %v, want errors.Is(err, %v)", err, want)
	}
	for _, other := range replaySentinels {
		if other != want && errors.Is(err, other) {
			t.Fatalf("ReplayVerify = %v matches both %v and %v; sentinels must be distinct", err, want, other)
		}
	}
	if !strings.Contains(err.Error(), fmt.Sprintf("seq %d", seq)) {
		t.Fatalf("ReplayVerify = %v: the error must name the offending event as %q", err, fmt.Sprintf("seq %d", seq))
	}
}

func unitCellAt(u, k int) int {
	switch {
	case u < 9:
		return u*9 + k
	case u < 18:
		return k*9 + (u - 9)
	default:
		b := u - 18
		return (b/3*3+k/3)*9 + b%3*3 + k%3
	}
}

// candidateCount recomputes a cell's live-candidate count from the raw grid
// string — peer elimination only, independent of solver internals.
func candidateCount(grid string, i int) int {
	n := 0
	for d := 1; d <= 9; d++ {
		if gridHasCandidate(grid, i/9, i%9, d) {
			n++
		}
	}
	return n
}

// singleAvailable reports whether a naked or hidden single exists in the
// peer-elimination candidate state of the raw grid — the same availability
// notion the verifier's scheduling check must implement.
func singleAvailable(grid string) bool {
	for i := 0; i < 81; i++ {
		if grid[i] == '0' && candidateCount(grid, i) == 1 {
			return true
		}
	}
	for u := 0; u < 27; u++ {
		for d := 1; d <= 9; d++ {
			n := 0
			for k := 0; k < 9; k++ {
				i := unitCellAt(u, k)
				if gridHasCandidate(grid, i/9, i%9, d) {
					n++
				}
			}
			if n == 1 {
				return true
			}
		}
	}
	return false
}

// assertPointingPremise verifies the fixture fact the tampered events lean on:
// digit 1's candidates in box 6 are exactly r7c2 and r8c2 (confined to column
// 2), so a locked_candidates_pointing event with those witnesses structurally
// holds in the input state.
func assertPointingPremise(t *testing.T, grid string) {
	t.Helper()
	for _, i := range []int{54, 55, 56, 63, 64, 65, 72, 73, 74} {
		has := gridHasCandidate(grid, i/9, i%9, 1)
		want := i == 65 || i == 74
		if has != want {
			t.Fatalf("premise: box 6 digit-1 candidate at r%dc%d = %v, want %v", i/9, i%9, has, want)
		}
	}
}

var pointingWitnesses = []solver.Cell{{Row: 7, Col: 2}, {Row: 8, Col: 2}}

// AC-3 scheduling check (AUDIT.md L4/L5 cheapest-first discipline): a
// structurally genuine pointing elimination is rejected when a single is
// available. gridSingleVsPointing (mid_test.go) holds a naked single at r0c3
// while the box-6 pointing pattern stays live; eliminating the live non-truth
// candidate 1 at r2c2 violates only the singles-availability rule, so the
// sentinel is order-independent.
func TestReplayVerify_EliminationWhileSingleAvailable(t *testing.T) {
	g := mustParse(t, gridSingleVsPointing)
	res := solver.Solve(g)
	if res.Status != "solved" {
		t.Fatalf("premise: Status = %q, want solved", res.Status)
	}
	if candidateCount(gridSingleVsPointing, 3) != 1 {
		t.Fatal("premise: r0c3 is not a naked single")
	}
	assertPointingPremise(t, gridSingleVsPointing)
	if !gridHasCandidate(gridSingleVsPointing, 2, 2, 1) {
		t.Fatal("premise: candidate 1 at r2c2 is not live")
	}
	if res.Solution[2*9+2] == 1 {
		t.Fatal("premise: 1 is the true value of r2c2; the elimination would be unsound")
	}
	tampered := res
	tampered.Events = []solver.Event{{
		Seq:          1,
		Technique:    "locked_candidates_pointing",
		WitnessCells: pointingWitnesses,
		Eliminations: []solver.Elimination{{Row: 2, Col: 2, Digit: 1}},
		GridAfter:    g.String(),
	}}
	tampered.EventCount, tampered.Iterations = 1, 1
	assertRejection(t, oracle.ReplayVerify(g, tampered), oracle.ErrSingleAvailable, 1)
}

// AC-3 fabricated elimination: r3c2 sits on the pattern's column outside box 6
// but holds no live candidate 1 (peer r4c1 = 1), so the event eliminates a
// non-candidate. gridPointing is singles-exhausted, so the scheduling rule
// stays silent and liveness is the violated rule.
func TestReplayVerify_EliminationOfNonCandidate(t *testing.T) {
	g := mustParse(t, gridPointing)
	res := solver.Solve(g)
	if res.Status != "solved" {
		t.Fatalf("premise: Status = %q, want solved", res.Status)
	}
	if singleAvailable(gridPointing) {
		t.Fatal("premise: gridPointing must be singles-exhausted")
	}
	assertPointingPremise(t, gridPointing)
	if gridPointing[3*9+2] != '0' || gridHasCandidate(gridPointing, 3, 2, 1) {
		t.Fatal("premise: r3c2 must be blank with no live candidate 1")
	}
	tampered := res
	tampered.Events = []solver.Event{{
		Seq:          1,
		Technique:    "locked_candidates_pointing",
		WitnessCells: pointingWitnesses,
		Eliminations: []solver.Elimination{{Row: 3, Col: 2, Digit: 1}},
		GridAfter:    g.String(),
	}}
	tampered.EventCount, tampered.Iterations = 1, 1
	assertRejection(t, oracle.ReplayVerify(g, tampered), oracle.ErrEliminationNotCandidate, 1)
}

// AC-3 eliminate-the-truth: the true value of r0c2 is a live candidate (the
// truth always survives peer elimination) and the event eliminates it. It is
// live, so liveness passes; it equals the oracle's value, so the truth rule is
// the violated one under the pinned check order.
func TestReplayVerify_EliminationOfOracleValue(t *testing.T) {
	g := mustParse(t, gridPointing)
	res := solver.Solve(g)
	if res.Status != "solved" {
		t.Fatalf("premise: Status = %q, want solved", res.Status)
	}
	if singleAvailable(gridPointing) {
		t.Fatal("premise: gridPointing must be singles-exhausted")
	}
	truth := int(res.Solution[2])
	if !gridHasCandidate(gridPointing, 0, 2, truth) {
		t.Fatal("premise: the true value of r0c2 must be a live candidate")
	}
	tampered := res
	tampered.Events = []solver.Event{{
		Seq:          1,
		Technique:    "locked_candidates_pointing",
		WitnessCells: pointingWitnesses,
		Eliminations: []solver.Elimination{{Row: 0, Col: 2, Digit: truth}},
		GridAfter:    g.String(),
	}}
	tampered.EventCount, tampered.Iterations = 1, 1
	assertRejection(t, oracle.ReplayVerify(g, tampered), oracle.ErrEliminationIsTruth, 1)
}

// AC-3 tampered placement: the first single of ORIGINAL seed 0, placed with a
// digit that disagrees with the oracle solution. gridAfter is kept consistent
// with the stated (wrong) placement so oracle-equality is the violated rule.
func TestReplayVerify_PlacementDisagreesWithOracle(t *testing.T) {
	seed := originalSeeds(t)[0]
	g := mustParse(t, seed)
	res := solver.Solve(g)
	if res.Status != "solved" {
		t.Fatalf("premise: Status = %q, want solved", res.Status)
	}
	if len(res.Events) == 0 || res.Events[0].Placement == nil {
		t.Fatal("premise: the first event of an ORIGINAL solve must be a single placement")
	}
	p := *res.Events[0].Placement
	wrong := p.Digit%9 + 1
	after := []byte(g.String())
	after[p.Row*9+p.Col] = byte('0' + wrong)
	tampered := res
	tampered.Events = []solver.Event{{
		Seq:          1,
		Technique:    res.Events[0].Technique,
		WitnessCells: []solver.Cell{{Row: p.Row, Col: p.Col}},
		Placement:    &solver.Placement{Row: p.Row, Col: p.Col, Digit: wrong},
		GridAfter:    string(after),
	}}
	tampered.EventCount, tampered.Iterations = 1, 1
	assertRejection(t, oracle.ReplayVerify(g, tampered), oracle.ErrPlacementNotOracle, 1)
}

// AC-3 witness-structure arm: each subtest builds an event that passes every
// check ahead of checkWitness — scheduling (singles-exhausted state), liveness
// (live candidates), oracle-truth (non-truth targets) — so the witness
// structure is the only violated rule. Witness-structure errors are
// descriptive by contract (ADR-0013): "seq N" named, no sentinel wrapped.
// This is the committed falsifier for a verifier that skips checkWitness.
func assertWitnessRejection(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		t.Fatal("ReplayVerify = nil, want a witness-structure rejection")
	}
	if !strings.Contains(err.Error(), "seq 1") {
		t.Fatalf("ReplayVerify = %v: the error must name the offending event as %q", err, "seq 1")
	}
	for _, s := range replaySentinels {
		if errors.Is(err, s) {
			t.Fatalf("ReplayVerify = %v wraps sentinel %v; witness-structure errors carry no sentinel", err, s)
		}
	}
}

func TestReplayVerify_WitnessStructureTamper(t *testing.T) {
	// Witnesses kept genuine (digit 1 in box 6 confined to column 2), the
	// elimination tampered to r0c0: live candidate 1, non-truth, but off the
	// pointing column — the pattern does not justify the target.
	t.Run("pointing_elimination_off_line", func(t *testing.T) {
		g := mustParse(t, gridPointing)
		res := solver.Solve(g)
		if res.Status != "solved" {
			t.Fatalf("premise: Status = %q, want solved", res.Status)
		}
		if singleAvailable(gridPointing) {
			t.Fatal("premise: gridPointing must be singles-exhausted")
		}
		assertPointingPremise(t, gridPointing)
		if !gridHasCandidate(gridPointing, 0, 0, 1) {
			t.Fatal("premise: candidate 1 at r0c0 is not live")
		}
		if res.Solution[0] == 1 {
			t.Fatal("premise: 1 is the true value of r0c0; the truth rule would fire instead")
		}
		tampered := res
		tampered.Events = []solver.Event{{
			Seq:          1,
			Technique:    "locked_candidates_pointing",
			WitnessCells: pointingWitnesses,
			Eliminations: []solver.Elimination{{Row: 0, Col: 0, Digit: 1}},
			GridAfter:    g.String(),
		}}
		tampered.EventCount, tampered.Iterations = 1, 1
		assertWitnessRejection(t, oracle.ReplayVerify(g, tampered))
	})

	// Eliminations kept genuine (the solver's own x_wing targets), the last
	// witness tampered r8c8 -> r8c7: the cells then span three columns, so no
	// 2x2 base/cover geometry holds in either fish orientation.
	t.Run("x_wing_witnesses_break_geometry", func(t *testing.T) {
		g := mustParse(t, gridXWing)
		res := solver.Solve(g)
		if len(res.Events) == 0 || res.Events[0].Technique != "x_wing" {
			t.Fatalf("premise: Events[0] must be the x_wing event (status %q, %d events)", res.Status, len(res.Events))
		}
		if singleAvailable(gridXWing) {
			t.Fatal("premise: gridXWing must be singles-exhausted")
		}
		ev := res.Events[0]
		if len(ev.WitnessCells) != 4 || ev.WitnessCells[3] != (solver.Cell{Row: 8, Col: 8}) {
			t.Fatalf("premise: witnesses = %v, want the pinned base ending at r8c8", ev.WitnessCells)
		}
		sol, count := oracle.Solve(g)
		for _, e := range ev.Eliminations {
			if !gridHasCandidate(gridXWing, e.Row, e.Col, e.Digit) {
				t.Fatalf("premise: elimination %+v is not a live candidate", e)
			}
			if count == 1 && int(sol[e.Row*9+e.Col]) == e.Digit {
				t.Fatalf("premise: elimination %+v targets the oracle truth; the truth rule would fire instead", e)
			}
		}
		witnesses := append([]solver.Cell(nil), ev.WitnessCells...)
		witnesses[3] = solver.Cell{Row: 8, Col: 7}
		tampered := res
		tampered.Events = []solver.Event{{
			Seq:          1,
			Technique:    "x_wing",
			WitnessCells: witnesses,
			Eliminations: ev.Eliminations,
			GridAfter:    g.String(),
		}}
		tampered.EventCount, tampered.Iterations = 1, 1
		assertWitnessRejection(t, oracle.ReplayVerify(g, tampered))
	})
}
