package solver

import (
	"math/bits"
	"slices"
)

// Canonical fish scan (ADR-0007 extension, F-05): orientation outermost —
// rows-base then cols-base — then digits ascending, then base-line
// k-combinations lexicographic over ascending line indexes. A base line holds
// >=2 candidates of the digit; the union of candidate cross-lines over the k
// base lines must be EXACTLY k. That single count is also the finned/sashimi
// exclusion (AUDIT L1, AC-4): any fin pushes the union past k and the
// combination is skipped.

func (s *solveState) detectXWing() ([]Cell, []Elimination, bool)     { return s.detectFish(2) }
func (s *solveState) detectSwordfish() ([]Cell, []Elimination, bool) { return s.detectFish(3) }
func (s *solveState) detectJellyfish() ([]Cell, []Elimination, bool) { return s.detectFish(4) }

func (s *solveState) detectFish(k int) ([]Cell, []Elimination, bool) {
	for _, rowsBase := range [2]bool{true, false} {
		for d := 1; d <= 9; d++ {
			if w, e, ok := s.fishForDigit(k, d, rowsBase); ok {
				return w, e, true
			}
		}
	}
	return nil, nil, false
}

func (s *solveState) fishForDigit(k, d int, rowsBase bool) (w []Cell, e []Elimination, ok bool) {
	unitBase := 0
	if !rowsBase {
		unitBase = 9
	}
	var lines []int
	var spots []uint16
	for l := 0; l < 9; l++ {
		if m := s.digitSpotsInUnit(unitBase+l, d); bits.OnesCount16(m) >= 2 {
			lines = append(lines, l)
			spots = append(spots, m)
		}
	}
	if len(lines) < k {
		return nil, nil, false
	}
	ok = forEachCombo(len(lines), k, func(combo []int) bool {
		var cover, baseSet uint16
		for _, c := range combo {
			cover |= spots[c]
			baseSet |= 1 << lines[c]
		}
		if bits.OnesCount16(cover) != k {
			return false
		}
		if e = s.fishElims(d, rowsBase, baseSet, cover); len(e) == 0 {
			return false
		}
		w = fishWitnesses(rowsBase, lines, spots, combo)
		return true
	})
	return w, e, ok
}

func (s *solveState) fishElims(d int, rowsBase bool, baseSet, cover uint16) []Elimination {
	var es []Elimination
	for r := 0; r < 9; r++ {
		for c := 0; c < 9; c++ {
			base, cross := r, c
			if !rowsBase {
				base, cross = c, r
			}
			if baseSet&(1<<base) != 0 || cover&(1<<cross) == 0 {
				continue
			}
			if i := r*9 + c; s.grid[i] == 0 && s.hasCandidate(i, d) {
				es = append(es, Elimination{Row: r, Col: c, Digit: d})
			}
		}
	}
	return es
}

func fishWitnesses(rowsBase bool, lines []int, spots []uint16, combo []int) []Cell {
	var cs []Cell
	for _, c := range combo {
		for cross := 0; cross < 9; cross++ {
			if spots[c]&(1<<cross) == 0 {
				continue
			}
			r, col := lines[c], cross
			if !rowsBase {
				r, col = cross, lines[c]
			}
			cs = append(cs, Cell{Row: r, Col: col})
		}
	}
	slices.SortFunc(cs, func(a, b Cell) int { return (a.Row*9 + a.Col) - (b.Row*9 + b.Col) })
	return cs
}
