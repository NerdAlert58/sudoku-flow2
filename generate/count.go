package generate

import (
	"math/bits"

	"github.com/NerdAlert58/sudoku-flow2/solver"
)

// counter counts solutions by backtracking with a fewest-candidates cell
// choice, capped at 2 — just enough to decide uniqueness. The count reaches
// callers only as dig's keep/restore decision, never any return value
// (ADR-0002 blinded surface).
type counter struct {
	grid              solver.Grid
	rows, cols, boxes [9]uint16
	count             int
}

func countUpTo2(g solver.Grid) int {
	c := &counter{grid: g}
	for i, d := range g {
		if d == 0 {
			continue
		}
		bit := uint16(1) << d
		c.rows[i/9] |= bit
		c.cols[i%9] |= bit
		c.boxes[boxOf(i)] |= bit
	}
	c.search()
	return c.count
}

func (c *counter) search() {
	i, mask := c.bestCell()
	if i < 0 {
		c.count++
		return
	}
	r, col, b := i/9, i%9, boxOf(i)
	for d := 1; d <= 9 && c.count < 2; d++ {
		bit := uint16(1) << d
		if mask&bit == 0 {
			continue
		}
		c.grid[i] = byte(d)
		c.rows[r] |= bit
		c.cols[col] |= bit
		c.boxes[b] |= bit
		c.search()
		c.grid[i] = 0
		c.rows[r] &^= bit
		c.cols[col] &^= bit
		c.boxes[b] &^= bit
	}
}

// bestCell returns the empty cell with the fewest legal digits and its
// candidate mask, or (-1, 0) on a complete grid. A zero-candidate cell is
// returned immediately: the branch is dead and search's loop places nothing.
func (c *counter) bestCell() (int, uint16) {
	best, bestMask, bestN := -1, uint16(0), 10
	for i, d := range c.grid {
		if d != 0 {
			continue
		}
		mask := ^(c.rows[i/9] | c.cols[i%9] | c.boxes[boxOf(i)]) & 0x3fe
		if n := bits.OnesCount16(mask); n < bestN {
			best, bestMask, bestN = i, mask, n
			if n <= 1 {
				break
			}
		}
	}
	return best, bestMask
}
