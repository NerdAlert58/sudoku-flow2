package solver

import (
	"math/bits"
	"slices"
)

// Canonical wing scans (ADR-0007 extension, F-05): xy_wing and xyz_wing take
// pivots row-major and pincer pairs lexicographic over the pivot's row-major
// peer list; w_wing takes bivalue A row-major, matching non-peer B row-major
// after A, link digit ascending over the pair, strong-link units canonical
// 0-26. All three are constructive dilemmas over already-certain bounded
// disjunctions (AUDIT L1): every branch of the pivot/link disjunction forces
// the same digit out of the target cells.

func (s *solveState) detectXYWing() ([]Cell, []Elimination, bool) {
	for p := 0; p < 81; p++ {
		if pm := s.exactCandidates(p, 2); pm != 0 {
			if w, e, ok := s.xyWingAtPivot(p, pm); ok {
				return w, e, true
			}
		}
	}
	return nil, nil, false
}

func (s *solveState) xyWingAtPivot(p int, pm uint16) ([]Cell, []Elimination, bool) {
	peers := peerCells(p)
	for ai, a := range peers {
		ma := s.exactCandidates(a, 2)
		if ma == 0 {
			continue
		}
		for _, b := range peers[ai+1:] {
			z, ok := xyPincers(pm, ma, s.exactCandidates(b, 2))
			if !ok {
				continue
			}
			if e := s.elimsSeeingAll(z, []int{p, a, b}, a, b); len(e) != 0 {
				return sortedCells(p, a, b), e, true
			}
		}
	}
	return nil, nil, false
}

// xyPincers validates pivot {X,Y} with pincers {X,Z} and {Y,Z}: each pincer
// shares exactly one (distinct) pivot digit, and the pincers share exactly
// one digit Z outside the pivot.
func xyPincers(pm, ma, mb uint16) (int, bool) {
	if ma == 0 || mb == 0 || bits.OnesCount16(ma&pm) != 1 || bits.OnesCount16(mb&pm) != 1 {
		return 0, false
	}
	z := ma & mb
	if ma&pm == mb&pm || bits.OnesCount16(z) != 1 || z&pm != 0 {
		return 0, false
	}
	return bits.TrailingZeros16(z), true
}

func (s *solveState) detectXYZWing() ([]Cell, []Elimination, bool) {
	for p := 0; p < 81; p++ {
		if pm := s.exactCandidates(p, 3); pm != 0 {
			if w, e, ok := s.xyzWingAtPivot(p, pm); ok {
				return w, e, true
			}
		}
	}
	return nil, nil, false
}

func (s *solveState) xyzWingAtPivot(p int, pm uint16) ([]Cell, []Elimination, bool) {
	peers := peerCells(p)
	for ai, a := range peers {
		ma := s.exactCandidates(a, 2)
		if ma == 0 || ma&^pm != 0 {
			continue
		}
		for _, b := range peers[ai+1:] {
			mb := s.exactCandidates(b, 2)
			z := ma & mb
			if mb == 0 || mb&^pm != 0 || bits.OnesCount16(z) != 1 {
				continue
			}
			if e := s.elimsSeeingAll(bits.TrailingZeros16(z), []int{p, a, b}, p, a, b); len(e) != 0 {
				return sortedCells(p, a, b), e, true
			}
		}
	}
	return nil, nil, false
}

func (s *solveState) detectWWing() ([]Cell, []Elimination, bool) {
	for a := 0; a < 81; a++ {
		if ma := s.exactCandidates(a, 2); ma != 0 {
			if w, e, ok := s.wWingFromA(a, ma); ok {
				return w, e, true
			}
		}
	}
	return nil, nil, false
}

func (s *solveState) wWingFromA(a int, ma uint16) ([]Cell, []Elimination, bool) {
	lo := bits.TrailingZeros16(ma)
	hi := 15 - bits.LeadingZeros16(ma)
	for b := a + 1; b < 81; b++ {
		if sees(a, b) || s.exactCandidates(b, 2) != ma {
			continue
		}
		for _, x := range [2]int{lo, hi} {
			if w, e, ok := s.wWingLink(a, b, x, lo+hi-x); ok {
				return w, e, true
			}
		}
	}
	return nil, nil, false
}

// wWingLink hunts the strong link on link digit x: a unit whose only two
// x-candidates W1, W2 are distinct from A and B, one seeing A and the other
// seeing B; then the non-link digit y leaves every cell seeing both A and B.
func (s *solveState) wWingLink(a, b, x, y int) ([]Cell, []Elimination, bool) {
	e := s.elimsSeeingAll(y, []int{a, b}, a, b)
	if len(e) == 0 {
		return nil, nil, false
	}
	for u := 0; u < 27; u++ {
		spots := s.digitSpotsInUnit(u, x)
		if bits.OnesCount16(spots) != 2 {
			continue
		}
		w1 := units[u][bits.TrailingZeros16(spots)]
		w2 := units[u][15-bits.LeadingZeros16(spots)]
		if w1 == a || w1 == b || w2 == a || w2 == b || !linkPairing(w1, w2, a, b) {
			continue
		}
		return sortedCells(a, b, w1, w2), e, true
	}
	return nil, nil, false
}

func linkPairing(w1, w2, a, b int) bool {
	return (sees(w1, a) && sees(w2, b)) || (sees(w1, b) && sees(w2, a))
}

// exactCandidates returns the candidate mask of empty cell i when it holds
// exactly n candidates, else 0.
func (s *solveState) exactCandidates(i, n int) uint16 {
	if s.grid[i] != 0 {
		return 0
	}
	m := s.cellCandidates(i)
	if bits.OnesCount16(m) != n {
		return 0
	}
	return m
}

func sees(a, b int) bool {
	return a != b && (a/9 == b/9 || a%9 == b%9 || (a/27 == b/27 && a%9/3 == b%9/3))
}

// peerCells lists the 20 peers of cell i in row-major-ascending order.
func peerCells(i int) []int {
	ps := make([]int, 0, 20)
	for j := 0; j < 81; j++ {
		if sees(i, j) {
			ps = append(ps, j)
		}
	}
	return ps
}

// elimsSeeingAll collects digit-d eliminations from every cell outside
// exclude that sees all sight cells; the row-major scan births them sorted.
func (s *solveState) elimsSeeingAll(d int, exclude []int, sights ...int) []Elimination {
	var es []Elimination
	for i := 0; i < 81; i++ {
		if s.grid[i] != 0 || slices.Contains(exclude, i) || !seesAll(i, sights) {
			continue
		}
		if s.hasCandidate(i, d) {
			es = append(es, Elimination{Row: i / 9, Col: i % 9, Digit: d})
		}
	}
	return es
}

func seesAll(i int, sights []int) bool {
	for _, c := range sights {
		if !sees(i, c) {
			return false
		}
	}
	return true
}

func sortedCells(idx ...int) []Cell {
	slices.Sort(idx)
	cs := make([]Cell, len(idx))
	for j, i := range idx {
		cs[j] = Cell{Row: i / 9, Col: i % 9}
	}
	return cs
}
