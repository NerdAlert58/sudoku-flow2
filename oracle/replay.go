package oracle

import (
	"errors"
	"fmt"

	"github.com/NerdAlert58/sudoku-flow2/solver"
)

// The pinned rejection surface (ADR-0013): each ReplayVerify error wraps at
// most one sentinel, and every per-event error names the offending event as
// "seq N".
var (
	ErrSingleAvailable         = errors.New("a naked or hidden single is available")
	ErrPlacementNotOracle      = errors.New("placement disagrees with the oracle solution")
	ErrEliminationNotCandidate = errors.New("elimination targets a non-candidate")
	ErrEliminationIsTruth      = errors.New("elimination targets the oracle solution value")
)

type verifier struct {
	*shadow
	sol    solver.Grid
	unique bool
}

// ReplayVerify replays every event of res against a shadow candidate state
// derived from input alone (ADR-0013). Events are checked in order,
// fail-fast: the first violating event's error is returned. When the oracle
// count is 1 and res is solved, the grid after the last event must equal
// the oracle's unique solution; oracle-anchored checks are skipped when the
// count is not 1 (no unique truth to compare against).
func ReplayVerify(input solver.Grid, res solver.SolveResult) error {
	sol, count := Solve(input)
	v := &verifier{shadow: newShadow(input), sol: sol, unique: count == 1}
	for i := range res.Events {
		if err := v.checkEvent(&res.Events[i]); err != nil {
			return err
		}
	}
	if res.Status == "solved" && v.unique && v.grid != sol {
		return fmt.Errorf("oracle: final grid %s does not equal the oracle solution %s",
			v.grid.String(), sol.String())
	}
	return nil
}

func (v *verifier) checkEvent(ev *solver.Event) error {
	if err := checkEventShape(ev); err != nil {
		return err
	}
	if ev.Placement != nil {
		return v.checkPlacement(ev)
	}
	return v.checkElimination(ev)
}

// checkEventShape rejects malformed events before any indexed access.
func checkEventShape(ev *solver.Event) error {
	if (ev.Placement != nil) == (len(ev.Eliminations) > 0) {
		return fmt.Errorf("oracle: seq %d: event must carry exactly one of placement or eliminations", ev.Seq)
	}
	for _, c := range ev.WitnessCells {
		if !inRange(c.Row, c.Col, 1) {
			return fmt.Errorf("oracle: seq %d: witness cell r%dc%d out of range", ev.Seq, c.Row, c.Col)
		}
	}
	if p := ev.Placement; p != nil && !inRange(p.Row, p.Col, p.Digit) {
		return fmt.Errorf("oracle: seq %d: placement out of range", ev.Seq)
	}
	for _, e := range ev.Eliminations {
		if !inRange(e.Row, e.Col, e.Digit) {
			return fmt.Errorf("oracle: seq %d: elimination out of range", ev.Seq)
		}
	}
	return nil
}

func inRange(r, c, d int) bool {
	return r >= 0 && r <= 8 && c >= 0 && c <= 8 && d >= 1 && d <= 9
}

// Placement intra-order (pinned): oracle-equality before the recomputed
// single condition, then witness shape, then gridAfter.
func (v *verifier) checkPlacement(ev *solver.Event) error {
	p := ev.Placement
	i := p.Row*9 + p.Col
	if v.unique && byte(p.Digit) != v.sol[i] {
		return fmt.Errorf("oracle: seq %d: placed %d at r%dc%d, oracle value is %d: %w",
			ev.Seq, p.Digit, p.Row, p.Col, v.sol[i], ErrPlacementNotOracle)
	}
	if err := v.checkSingleCondition(ev, i); err != nil {
		return err
	}
	if len(ev.WitnessCells) != 1 || ev.WitnessCells[0] != (solver.Cell{Row: p.Row, Col: p.Col}) {
		return fmt.Errorf("oracle: seq %d: %s witness must be exactly the placed cell", ev.Seq, ev.Technique)
	}
	v.place(i, p.Digit)
	if got := v.grid.String(); got != ev.GridAfter {
		return fmt.Errorf("oracle: seq %d: gridAfter %q is not the prior grid plus the placement (%q)",
			ev.Seq, ev.GridAfter, got)
	}
	return nil
}

// checkSingleCondition recomputes the NAMED single's condition from the
// shadow state; a placement by any other technique name is malformed.
func (v *verifier) checkSingleCondition(ev *solver.Event, i int) error {
	p := ev.Placement
	switch ev.Technique {
	case "naked_single":
		if v.cands[i] != 1<<p.Digit {
			return fmt.Errorf("oracle: seq %d: r%dc%d is not a naked single for %d in the shadow state",
				ev.Seq, p.Row, p.Col, p.Digit)
		}
	case "hidden_single":
		if !v.hiddenSingleAt(i, p.Digit) {
			return fmt.Errorf("oracle: seq %d: r%dc%d is not a hidden single for %d in the shadow state",
				ev.Seq, p.Row, p.Col, p.Digit)
		}
	default:
		return fmt.Errorf("oracle: seq %d: placement technique %q is not a single", ev.Seq, ev.Technique)
	}
	return nil
}

// Elimination intra-order (pinned): liveness, oracle-truth, witness
// structure. The scheduling check runs first — cheapest-first discipline
// (AUDIT L4/L5) is violated the moment the event fires, whatever it states.
func (v *verifier) checkElimination(ev *solver.Event) error {
	if v.singleAvailable() {
		return fmt.Errorf("oracle: seq %d: %s fired while %w", ev.Seq, ev.Technique, ErrSingleAvailable)
	}
	for _, e := range ev.Eliminations {
		if !v.has(e.Row*9+e.Col, e.Digit) {
			return fmt.Errorf("oracle: seq %d: %d at r%dc%d: %w",
				ev.Seq, e.Digit, e.Row, e.Col, ErrEliminationNotCandidate)
		}
	}
	if err := v.checkNotTruth(ev); err != nil {
		return err
	}
	if err := v.checkWitness(ev); err != nil {
		return err
	}
	if ev.GridAfter != v.grid.String() {
		return fmt.Errorf("oracle: seq %d: elimination gridAfter must leave the digits unchanged", ev.Seq)
	}
	for _, e := range ev.Eliminations {
		v.cands[e.Row*9+e.Col] &^= 1 << e.Digit
	}
	return nil
}

func (v *verifier) checkNotTruth(ev *solver.Event) error {
	if !v.unique {
		return nil
	}
	for _, e := range ev.Eliminations {
		if byte(e.Digit) == v.sol[e.Row*9+e.Col] {
			return fmt.Errorf("oracle: seq %d: %d at r%dc%d: %w",
				ev.Seq, e.Digit, e.Row, e.Col, ErrEliminationIsTruth)
		}
	}
	return nil
}

func (v *verifier) checkWitness(ev *solver.Event) error {
	check, known := witnessChecks[ev.Technique]
	if !known {
		return fmt.Errorf("oracle: seq %d: unknown elimination technique %q", ev.Seq, ev.Technique)
	}
	w, distinct := witnessIndexes(ev.WitnessCells)
	if !distinct || !check(v, w, ev.Eliminations) {
		return fmt.Errorf("oracle: seq %d: %s witness pattern does not hold in the shadow state",
			ev.Seq, ev.Technique)
	}
	return nil
}
