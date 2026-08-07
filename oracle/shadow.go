package oracle

import (
	"math/bits"

	"github.com/NerdAlert58/sudoku-flow2/solver"
)

// The board geometry is recomputed here from scratch: the verifier may never
// read solver internals (ADR-0013 non-circularity), only its exported event
// log. Unit order matches the ADR-0007 canon: rows 0-8, columns 0-8, boxes
// 0-8 row-major.
var units = buildUnits()

func buildUnits() [27][9]int {
	var u [27][9]int
	for i := 0; i < 9; i++ {
		for k := 0; k < 9; k++ {
			u[i][k] = i*9 + k
			u[9+i][k] = k*9 + i
			u[18+i][k] = (i/3*3+k/3)*9 + i%3*3 + k%3
		}
	}
	return u
}

func cellUnits(i int) [3]int {
	r, c := i/9, i%9
	return [3]int{r, 9 + c, 18 + r/3*3 + c/3}
}

func boxOf(i int) int {
	return i/27*3 + i%9/3
}

func sees(a, b int) bool {
	return a != b && (a/9 == b/9 || a%9 == b%9 || boxOf(a) == boxOf(b))
}

// shadow is the verifier's own candidate state: peer elimination from the
// input grid, then only event-driven updates (placements strip peers,
// validated eliminations remove exactly the stated candidates).
type shadow struct {
	grid  solver.Grid
	cands [81]uint16
}

func newShadow(g solver.Grid) *shadow {
	s := &shadow{grid: g}
	for i, d := range g {
		if d == 0 {
			s.cands[i] = allDigits &^ s.peerDigits(i)
		}
	}
	return s
}

func (s *shadow) peerDigits(i int) uint16 {
	var m uint16
	for _, u := range cellUnits(i) {
		for _, j := range units[u] {
			m |= 1 << s.grid[j]
		}
	}
	return m
}

func (s *shadow) has(i, d int) bool {
	return s.cands[i]&(1<<d) != 0
}

func (s *shadow) place(i, d int) {
	s.grid[i] = byte(d)
	s.cands[i] = 0
	for _, u := range cellUnits(i) {
		for _, j := range units[u] {
			s.cands[j] &^= 1 << d
		}
	}
}

// digitSpots returns the slot mask of unit u's live d-candidates.
func (s *shadow) digitSpots(u, d int) uint16 {
	var m uint16
	for k, i := range units[u] {
		if s.grid[i] == 0 && s.has(i, d) {
			m |= 1 << k
		}
	}
	return m
}

// singleAvailable is the verifier's independent naked/hidden single
// detection (ADR-0013): the cheapest-first scheduling check rejects any
// elimination event fired while one of these was live.
func (s *shadow) singleAvailable() bool {
	for i, m := range s.cands {
		if s.grid[i] == 0 && bits.OnesCount16(m) == 1 {
			return true
		}
	}
	for u := 0; u < 27; u++ {
		for d := 1; d <= 9; d++ {
			if bits.OnesCount16(s.digitSpots(u, d)) == 1 {
				return true
			}
		}
	}
	return false
}

// hiddenSingleAt reports whether cell i is the sole live d-candidate of one
// of its three units.
func (s *shadow) hiddenSingleAt(i, d int) bool {
	if !s.has(i, d) {
		return false
	}
	for _, u := range cellUnits(i) {
		if bits.OnesCount16(s.digitSpots(u, d)) == 1 {
			return true
		}
	}
	return false
}

// exactCandidates returns blank cell i's candidate mask when it holds
// exactly n candidates, else 0.
func (s *shadow) exactCandidates(i, n int) uint16 {
	if s.grid[i] != 0 || bits.OnesCount16(s.cands[i]) != n {
		return 0
	}
	return s.cands[i]
}
