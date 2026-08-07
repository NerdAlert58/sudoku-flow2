package oracle

import (
	"math/bits"

	"github.com/NerdAlert58/sudoku-flow2/solver"
)

func fishCheck(k int) func(*verifier, []int, []solver.Elimination) bool {
	return func(v *verifier, w []int, elims []solver.Elimination) bool {
		return v.fishHolds(k, w, elims)
	}
}

// fishHolds: the witnesses are exactly the live d-candidates of k base
// lines (each holding at least two) whose cross-lines cover exactly k
// lines; every elimination is d on a cover line off the base lines. The
// exact-k cover is also the finned/sashimi exclusion.
func (v *verifier) fishHolds(k int, w []int, elims []solver.Elimination) bool {
	d, ok := soleDigit(elims)
	if !ok || len(w) == 0 {
		return false
	}
	return v.fishOriented(k, d, true, w, elims) || v.fishOriented(k, d, false, w, elims)
}

func (v *verifier) fishOriented(k, d int, rowsBase bool, w []int, elims []solver.Elimination) bool {
	baseSet, cover := fishAxes(w, rowsBase)
	if bits.OnesCount16(baseSet) != k || bits.OnesCount16(cover) != k {
		return false
	}
	for b := 0; b < 9; b++ {
		if baseSet&(1<<b) != 0 && !v.baseLineExact(b, d, rowsBase, w) {
			return false
		}
	}
	for _, e := range elims {
		base, cross := axes(e.Row*9+e.Col, rowsBase)
		if baseSet&(1<<base) != 0 || cover&(1<<cross) == 0 {
			return false
		}
	}
	return true
}

func axes(i int, rowsBase bool) (base, cross int) {
	if rowsBase {
		return i / 9, i % 9
	}
	return i % 9, i / 9
}

func fishAxes(w []int, rowsBase bool) (baseSet, cover uint16) {
	for _, i := range w {
		base, cross := axes(i, rowsBase)
		baseSet |= 1 << base
		cover |= 1 << cross
	}
	return baseSet, cover
}

// baseLineExact: the witnesses on base line b are exactly its live
// d-candidates, at least two of them.
func (v *verifier) baseLineExact(b, d int, rowsBase bool, w []int) bool {
	u := b
	if !rowsBase {
		u = 9 + b
	}
	var mask uint16
	for _, i := range w {
		if base, cross := axes(i, rowsBase); base == b {
			mask |= 1 << cross
		}
	}
	return bits.OnesCount16(mask) >= 2 && v.digitSpots(u, d) == mask
}

// xyWingHolds: one witness is a bivalue pivot seen by the two bivalue
// pincers; each pincer shares one distinct pivot digit and the pincers
// share exactly the eliminated digit z outside the pivot; every elimination
// is z at a non-witness cell seeing both pincers.
func (v *verifier) xyWingHolds(w []int, elims []solver.Elimination) bool {
	z, ok := soleDigit(elims)
	if !ok || len(w) != 3 {
		return false
	}
	for pi := range w {
		p, a, b := w[pi], w[(pi+1)%3], w[(pi+2)%3]
		if v.xyShape(p, a, b, z) && elimsSee(elims, w, a, b) {
			return true
		}
	}
	return false
}

func (v *verifier) xyShape(p, a, b, z int) bool {
	pm, ma, mb := v.exactCandidates(p, 2), v.exactCandidates(a, 2), v.exactCandidates(b, 2)
	if pm == 0 || ma == 0 || mb == 0 || !sees(a, p) || !sees(b, p) {
		return false
	}
	if bits.OnesCount16(ma&pm) != 1 || bits.OnesCount16(mb&pm) != 1 || ma&pm == mb&pm {
		return false
	}
	return ma&mb == 1<<z && pm&(1<<z) == 0
}

// xyzWingHolds: a trivalue pivot seen by two bivalue pincers whose masks
// sit inside the pivot's and share exactly the eliminated digit; every
// elimination sees all three witnesses.
func (v *verifier) xyzWingHolds(w []int, elims []solver.Elimination) bool {
	z, ok := soleDigit(elims)
	if !ok || len(w) != 3 {
		return false
	}
	for pi := range w {
		p, a, b := w[pi], w[(pi+1)%3], w[(pi+2)%3]
		if v.xyzShape(p, a, b, z) && elimsSee(elims, w, p, a, b) {
			return true
		}
	}
	return false
}

func (v *verifier) xyzShape(p, a, b, z int) bool {
	pm, ma, mb := v.exactCandidates(p, 3), v.exactCandidates(a, 2), v.exactCandidates(b, 2)
	if pm == 0 || ma == 0 || mb == 0 || !sees(a, p) || !sees(b, p) {
		return false
	}
	return ma&^pm == 0 && mb&^pm == 0 && ma&mb == 1<<z
}

// wWingHolds: two same-pair bivalues A and B that do not see each other,
// plus a strong link on the pair's other digit through the two remaining
// witnesses (one seeing A, the other B); every elimination is the pair
// digit y at a cell outside {A,B} seeing both.
func (v *verifier) wWingHolds(w []int, elims []solver.Elimination) bool {
	y, ok := soleDigit(elims)
	if !ok || len(w) != 4 {
		return false
	}
	for ai := 0; ai < 4; ai++ {
		for bi := ai + 1; bi < 4; bi++ {
			a, b := w[ai], w[bi]
			w1, w2 := otherTwo(w, ai, bi)
			if v.wShape(a, b, w1, w2, y) && elimsSee(elims, []int{a, b}, a, b) {
				return true
			}
		}
	}
	return false
}

func otherTwo(w []int, ai, bi int) (int, int) {
	rest := make([]int, 0, 2)
	for j, c := range w {
		if j != ai && j != bi {
			rest = append(rest, c)
		}
	}
	return rest[0], rest[1]
}

func (v *verifier) wShape(a, b, w1, w2, y int) bool {
	ma := v.exactCandidates(a, 2)
	if ma == 0 || v.exactCandidates(b, 2) != ma || sees(a, b) || ma&(1<<y) == 0 {
		return false
	}
	x := bits.TrailingZeros16(ma &^ (1 << y))
	return pairSight(w1, w2, a, b) && v.strongLink(w1, w2, x)
}

// pairSight: one link cell sees a and the other sees b.
func pairSight(w1, w2, a, b int) bool {
	return (sees(w1, a) && sees(w2, b)) || (sees(w1, b) && sees(w2, a))
}

// strongLink: some unit holding both cells has exactly them as its live
// d-candidates.
func (v *verifier) strongLink(a, b, d int) bool {
	for u := 0; u < 27; u++ {
		if slotIn(u, a) >= 0 && slotIn(u, b) >= 0 && v.spotsAre(u, d, []int{a, b}) {
			return true
		}
	}
	return false
}

// elimsSee: every elimination cell is outside the excluded set and sees all
// the sight cells.
func elimsSee(elims []solver.Elimination, exclude []int, sights ...int) bool {
	for _, e := range elims {
		i := e.Row*9 + e.Col
		if contains(exclude, i) || !seesAll(i, sights) {
			return false
		}
	}
	return true
}

func seesAll(i int, sights []int) bool {
	for _, c := range sights {
		if !sees(i, c) {
			return false
		}
	}
	return true
}
