package solver

import (
	"math/bits"
	"slices"
)

// Canonical scan orders (ADR-0007, conventions recorded for F-04): pointing
// scans boxes 0-8 then digits ascending; claiming scans rows 0-8 then columns
// 0-8 then digits ascending; subsets scan k=2,3,4 outermost, then units in
// canonical order, then combinations lexicographic. Witnesses and
// eliminations are collected in unit-slot order, which is row-major for every
// unit kind, so they are born sorted.

func (s *solveState) detectPointing() ([]Cell, []Elimination, bool) {
	for u := 18; u < 27; u++ {
		for d := 1; d <= 9; d++ {
			spots := s.digitSpotsInUnit(u, d)
			line := confinedLine(u, spots)
			if line < 0 {
				continue
			}
			elims := s.elimsOutside(line, u, d)
			if len(elims) == 0 {
				continue
			}
			return unitCells(u, spots), elims, true
		}
	}
	return nil, nil, false
}

func (s *solveState) detectClaiming() ([]Cell, []Elimination, bool) {
	for u := 0; u < 18; u++ {
		for d := 1; d <= 9; d++ {
			spots := s.digitSpotsInUnit(u, d)
			box := confinedBox(u, spots)
			if box < 0 {
				continue
			}
			elims := s.elimsOutside(box, u, d)
			if len(elims) == 0 {
				continue
			}
			return unitCells(u, spots), elims, true
		}
	}
	return nil, nil, false
}

// confinedLine returns the unit index of the single row (preferred, per the
// rows-before-columns unit order) or column holding every set slot of the box
// unit, or -1.
func confinedLine(boxUnit int, spots uint16) int {
	row, col := -1, -1
	first := true
	for k := 0; k < 9; k++ {
		if spots&(1<<k) == 0 {
			continue
		}
		i := units[boxUnit][k]
		if first {
			row, col, first = i/9, i%9, false
			continue
		}
		if i/9 != row {
			row = -1
		}
		if i%9 != col {
			col = -1
		}
	}
	switch {
	case first:
		return -1
	case row >= 0:
		return row
	case col >= 0:
		return 9 + col
	}
	return -1
}

func confinedBox(lineUnit int, spots uint16) int {
	box := -1
	for k := 0; k < 9; k++ {
		if spots&(1<<k) == 0 {
			continue
		}
		b := cellUnitIndexes(units[lineUnit][k])[2]
		if box >= 0 && b != box {
			return -1
		}
		box = b
	}
	return box
}

func (s *solveState) elimsOutside(scanUnit, skipUnit, d int) []Elimination {
	var es []Elimination
	for _, i := range units[scanUnit] {
		if s.grid[i] != 0 || inUnit(i, skipUnit) || !s.hasCandidate(i, d) {
			continue
		}
		es = append(es, Elimination{Row: i / 9, Col: i % 9, Digit: d})
	}
	return es
}

func inUnit(i, u int) bool {
	ui := cellUnitIndexes(i)
	return ui[0] == u || ui[1] == u || ui[2] == u
}

func (s *solveState) detectNakedSubset() ([]Cell, []Elimination, bool) {
	for k := 2; k <= 4; k++ {
		for u := 0; u < 27; u++ {
			if w, e, ok := s.nakedSubsetInUnit(u, k); ok {
				return w, e, true
			}
		}
	}
	return nil, nil, false
}

func (s *solveState) nakedSubsetInUnit(u, k int) (w []Cell, e []Elimination, ok bool) {
	var slots []int
	for slot, i := range units[u] {
		if s.grid[i] == 0 {
			slots = append(slots, slot)
		}
	}
	// A subset spanning every empty cell has no elimination targets.
	if len(slots) <= k {
		return nil, nil, false
	}
	masks := make([]uint16, len(slots))
	for j, slot := range slots {
		masks[j] = s.cellCandidates(units[u][slot])
	}
	ok = forEachCombo(len(slots), k, func(combo []int) bool {
		var union, comboSlots uint16
		for _, c := range combo {
			union |= masks[c]
			comboSlots |= 1 << slots[c]
		}
		if bits.OnesCount16(union) != k {
			return false
		}
		if e = nakedElims(u, slots, masks, combo, union); len(e) == 0 {
			return false
		}
		w = unitCells(u, comboSlots)
		return true
	})
	return w, e, ok
}

func nakedElims(u int, slots []int, masks []uint16, combo []int, union uint16) []Elimination {
	var es []Elimination
	for j, slot := range slots {
		if slices.Contains(combo, j) {
			continue
		}
		es = appendDigitElims(es, units[u][slot], masks[j]&union)
	}
	return es
}

func appendDigitElims(es []Elimination, cell int, digits uint16) []Elimination {
	for d := 1; d <= 9; d++ {
		if digits&(1<<d) != 0 {
			es = append(es, Elimination{Row: cell / 9, Col: cell % 9, Digit: d})
		}
	}
	return es
}

func (s *solveState) detectHiddenSubset() ([]Cell, []Elimination, bool) {
	for k := 2; k <= 4; k++ {
		for u := 0; u < 27; u++ {
			if w, e, ok := s.hiddenSubsetInUnit(u, k); ok {
				return w, e, true
			}
		}
	}
	return nil, nil, false
}

func (s *solveState) hiddenSubsetInUnit(u, k int) (w []Cell, e []Elimination, ok bool) {
	var digits []int
	var spots []uint16
	for d := 1; d <= 9; d++ {
		if m := s.digitSpotsInUnit(u, d); m != 0 {
			digits = append(digits, d)
			spots = append(spots, m)
		}
	}
	// With only k live digits left there is no other digit to eliminate.
	if len(digits) <= k {
		return nil, nil, false
	}
	ok = forEachCombo(len(digits), k, func(combo []int) bool {
		var cells, comboDigits uint16
		for _, c := range combo {
			cells |= spots[c]
			comboDigits |= 1 << digits[c]
		}
		if bits.OnesCount16(cells) != k {
			return false
		}
		if e = hiddenElims(u, cells, comboDigits, spots, digits); len(e) == 0 {
			return false
		}
		w = unitCells(u, cells)
		return true
	})
	return w, e, ok
}

func hiddenElims(u int, cells, comboDigits uint16, spots []uint16, digits []int) []Elimination {
	var es []Elimination
	for slot := 0; slot < 9; slot++ {
		if cells&(1<<slot) == 0 {
			continue
		}
		i := units[u][slot]
		for j, d := range digits {
			if comboDigits&(1<<d) == 0 && spots[j]&(1<<slot) != 0 {
				es = append(es, Elimination{Row: i / 9, Col: i % 9, Digit: d})
			}
		}
	}
	return es
}

func (s *solveState) digitSpotsInUnit(u, d int) uint16 {
	var m uint16
	for k, i := range units[u] {
		if s.grid[i] == 0 && s.hasCandidate(i, d) {
			m |= 1 << k
		}
	}
	return m
}

func (s *solveState) cellCandidates(i int) uint16 {
	var m uint16
	for d := 1; d <= 9; d++ {
		if s.hasCandidate(i, d) {
			m |= 1 << d
		}
	}
	return m
}

func unitCells(u int, slots uint16) []Cell {
	var cs []Cell
	for k := 0; k < 9; k++ {
		if slots&(1<<k) != 0 {
			i := units[u][k]
			cs = append(cs, Cell{Row: i / 9, Col: i % 9})
		}
	}
	return cs
}

// forEachCombo enumerates k-combinations of {0..n-1} in lexicographic order
// (ADR-0007), invoking fn until it reports a firing instance. Requires n > k.
func forEachCombo(n, k int, fn func([]int) bool) bool {
	combo := make([]int, k)
	for i := range combo {
		combo[i] = i
	}
	for {
		if fn(combo) {
			return true
		}
		i := k - 1
		for i >= 0 && combo[i] == n-k+i {
			i--
		}
		if i < 0 {
			return false
		}
		combo[i]++
		for j := i + 1; j < k; j++ {
			combo[j] = combo[j-1] + 1
		}
	}
}
