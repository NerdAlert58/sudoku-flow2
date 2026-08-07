// Package oracle is the test-only trust machinery of ADR-0002/ADR-0013: a
// brute-force ground-truth solver plus the replay verifier. Shipped code
// must never import it (solver/importguard_test.go makes the ban mechanical),
// which keeps every proof anchored here non-circular.
package oracle

import (
	"math/bits"

	"github.com/NerdAlert58/sudoku-flow2/solver"
)

const allDigits uint16 = 0b1111111110

// Solve brute-forces g by bitmask backtracking and counts solutions, capped
// at 2 so every call stays cheap. When the count is 1 the returned grid is
// the unique solution; a complete valid grid is its own solution with count
// 1. Cell choice (fewest candidates, lowest index on ties) and digit order
// (ascending) are fixed, so the returned grid is deterministic.
func Solve(g solver.Grid) (solver.Grid, int) {
	b := &bruteForce{grid: g}
	for i, d := range g {
		if d != 0 {
			b.mark(i, int(d))
		}
	}
	b.search()
	if b.count == 0 {
		return g, 0
	}
	return b.solution, b.count
}

type bruteForce struct {
	grid     solver.Grid
	rows     [9]uint16
	cols     [9]uint16
	boxes    [9]uint16
	solution solver.Grid
	count    int
}

func (b *bruteForce) mark(i, d int) {
	b.rows[i/9] |= 1 << d
	b.cols[i%9] |= 1 << d
	b.boxes[boxOf(i)] |= 1 << d
}

func (b *bruteForce) unmark(i, d int) {
	b.rows[i/9] &^= 1 << d
	b.cols[i%9] &^= 1 << d
	b.boxes[boxOf(i)] &^= 1 << d
}

func (b *bruteForce) search() {
	i, m := b.pickCell()
	if i < 0 {
		if b.count == 0 {
			b.solution = b.grid
		}
		b.count++
		return
	}
	for d := 1; d <= 9 && b.count < 2; d++ {
		if m&(1<<d) == 0 {
			continue
		}
		b.grid[i] = byte(d)
		b.mark(i, d)
		b.search()
		b.unmark(i, d)
		b.grid[i] = 0
	}
}

// pickCell returns the blank cell with the fewest candidates and its
// candidate mask, or -1 when the grid is complete. A zero-candidate cell is
// picked first and prunes the branch (its digit loop tries nothing).
func (b *bruteForce) pickCell() (int, uint16) {
	best, bestMask, bestN := -1, uint16(0), 10
	for i, d := range b.grid {
		if d != 0 {
			continue
		}
		m := allDigits &^ (b.rows[i/9] | b.cols[i%9] | b.boxes[boxOf(i)])
		if n := bits.OnesCount16(m); n < bestN {
			best, bestMask, bestN = i, m, n
		}
	}
	return best, bestMask
}
