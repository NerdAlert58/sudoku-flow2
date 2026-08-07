package oracle

import (
	"math/bits"

	"github.com/NerdAlert58/sudoku-flow2/solver"
)

// Structural witness checks, one per elimination technique (ADR-0013):
// each verifies that the named pattern holds in the shadow state and that
// every stated elimination is a target the pattern justifies. They are
// independent re-derivations from the witness cells — never calls into the
// solver's detectors. All 13 ladder techniques are covered even though the
// corpus exercises a subset; F-07's curated fixtures consume the rest.
var witnessChecks = map[string]func(*verifier, []int, []solver.Elimination) bool{
	"locked_candidates_pointing": (*verifier).pointingHolds,
	"locked_candidates_claiming": (*verifier).claimingHolds,
	"naked_subset":               (*verifier).nakedSubsetHolds,
	"hidden_subset":              (*verifier).hiddenSubsetHolds,
	"x_wing":                     fishCheck(2),
	"swordfish":                  fishCheck(3),
	"jellyfish":                  fishCheck(4),
	"xy_wing":                    (*verifier).xyWingHolds,
	"xyz_wing":                   (*verifier).xyzWingHolds,
	"w_wing":                     (*verifier).wWingHolds,
	"simple_colouring":           (*verifier).colouringHolds,
}

// pointingHolds: the witnesses are exactly the live d-candidates of one box,
// confined to one line; every elimination is d on that line outside the box.
func (v *verifier) pointingHolds(w []int, elims []solver.Elimination) bool {
	d, ok := soleDigit(elims)
	if !ok || len(w) == 0 {
		return false
	}
	box := 18 + boxOf(w[0])
	if !v.spotsAre(box, d, w) {
		return false
	}
	for _, u := range commonUnits(w) {
		if u < 18 && elimsOutside(elims, u, box) {
			return true
		}
	}
	return false
}

// claimingHolds: the witnesses are exactly the live d-candidates of one
// line, confined to one box; every elimination is d in that box off the
// line.
func (v *verifier) claimingHolds(w []int, elims []solver.Elimination) bool {
	d, ok := soleDigit(elims)
	if !ok || len(w) == 0 {
		return false
	}
	box := 18 + boxOf(w[0])
	if _, inBox := slotMask(box, w); !inBox {
		return false
	}
	for _, u := range commonUnits(w) {
		if u < 18 && v.spotsAre(u, d, w) && elimsOutside(elims, box, u) {
			return true
		}
	}
	return false
}

// nakedSubsetHolds: k blank witness cells sharing a unit, candidate union
// exactly k digits; every elimination strips a union digit from a
// non-witness cell that shares a unit with all k.
func (v *verifier) nakedSubsetHolds(w []int, elims []solver.Elimination) bool {
	k := len(w)
	if k < 2 || k > 4 {
		return false
	}
	var union uint16
	for _, i := range w {
		if v.grid[i] != 0 {
			return false
		}
		union |= v.cands[i]
	}
	shared := commonUnits(w)
	if bits.OnesCount16(union) != k || len(shared) == 0 {
		return false
	}
	return nakedElimsJustified(elims, w, union, shared)
}

func nakedElimsJustified(elims []solver.Elimination, w []int, union uint16, shared []int) bool {
	for _, e := range elims {
		i := e.Row*9 + e.Col
		if union&(1<<e.Digit) == 0 || contains(w, i) || !inAnyUnit(i, shared) {
			return false
		}
	}
	return true
}

// hiddenSubsetHolds: within a shared unit, exactly k digits are confined to
// the k witness cells and collectively need all of them; every elimination
// strips a non-subset digit from a witness cell.
func (v *verifier) hiddenSubsetHolds(w []int, elims []solver.Elimination) bool {
	if len(w) < 2 || len(w) > 4 {
		return false
	}
	for _, u := range commonUnits(w) {
		if v.hiddenSubsetInUnit(u, w, elims) {
			return true
		}
	}
	return false
}

func (v *verifier) hiddenSubsetInUnit(u int, w []int, elims []solver.Elimination) bool {
	wm, ok := slotMask(u, w)
	if !ok {
		return false
	}
	confined, cover := v.confinedDigits(u, wm)
	if bits.OnesCount16(confined) != len(w) || cover != wm {
		return false
	}
	for _, e := range elims {
		if !contains(w, e.Row*9+e.Col) || confined&(1<<e.Digit) != 0 {
			return false
		}
	}
	return true
}

// confinedDigits returns the digit mask of digits whose live spots in u are
// non-empty and fall entirely within wm, plus the union of those spots.
func (v *verifier) confinedDigits(u int, wm uint16) (digits, cover uint16) {
	for d := 1; d <= 9; d++ {
		spots := v.digitSpots(u, d)
		if spots != 0 && spots&^wm == 0 {
			digits |= 1 << d
			cover |= spots
		}
	}
	return digits, cover
}

func witnessIndexes(cells []solver.Cell) ([]int, bool) {
	idx := make([]int, len(cells))
	var seen [81]bool
	for j, c := range cells {
		i := c.Row*9 + c.Col
		if seen[i] {
			return nil, false
		}
		seen[i] = true
		idx[j] = i
	}
	return idx, true
}

// soleDigit returns the one digit every elimination states, if shared.
func soleDigit(elims []solver.Elimination) (int, bool) {
	d := elims[0].Digit
	for _, e := range elims {
		if e.Digit != d {
			return 0, false
		}
	}
	return d, true
}

// spotsAre reports whether unit u's live d-candidates are exactly cells.
func (v *verifier) spotsAre(u, d int, cells []int) bool {
	want, ok := slotMask(u, cells)
	return ok && v.digitSpots(u, d) == want
}

// slotMask maps cells to their slot mask within unit u; ok is false when a
// cell is outside the unit.
func slotMask(u int, cells []int) (uint16, bool) {
	var m uint16
	for _, i := range cells {
		k := slotIn(u, i)
		if k < 0 {
			return 0, false
		}
		m |= 1 << k
	}
	return m, true
}

// slotIn returns i's slot within unit u, or -1.
func slotIn(u, i int) int {
	for k, c := range units[u] {
		if c == i {
			return k
		}
	}
	return -1
}

// commonUnits lists every unit containing all the given cells.
func commonUnits(cells []int) []int {
	var us []int
	for u := 0; u < 27; u++ {
		if _, ok := slotMask(u, cells); ok {
			us = append(us, u)
		}
	}
	return us
}

func inAnyUnit(i int, us []int) bool {
	for _, u := range us {
		if slotIn(u, i) >= 0 {
			return true
		}
	}
	return false
}

func contains(cells []int, i int) bool {
	for _, c := range cells {
		if c == i {
			return true
		}
	}
	return false
}

// elimsOutside: every elimination cell lies in unit `in` and outside unit
// `notIn`.
func elimsOutside(elims []solver.Elimination, in, notIn int) bool {
	for _, e := range elims {
		i := e.Row*9 + e.Col
		if slotIn(in, i) < 0 || slotIn(notIn, i) >= 0 {
			return false
		}
	}
	return true
}
