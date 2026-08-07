package generate

import (
	"math/rand"

	"github.com/NerdAlert58/sudoku-flow2/solver"
)

// filler builds a complete random grid by backtracking over cells 0..80 with
// an rng-shuffled digit order per cell.
type filler struct {
	grid              solver.Grid
	rows, cols, boxes [9]uint16
	rng               *rand.Rand
}

func fill(rng *rand.Rand) solver.Grid {
	f := &filler{rng: rng}
	f.next(0)
	return f.grid
}

func (f *filler) next(i int) bool {
	if i == 81 {
		return true
	}
	r, c, b := i/9, i%9, boxOf(i)
	digits := [9]byte{1, 2, 3, 4, 5, 6, 7, 8, 9}
	f.rng.Shuffle(9, func(x, y int) { digits[x], digits[y] = digits[y], digits[x] })
	for _, d := range digits {
		bit := uint16(1) << d
		if (f.rows[r]|f.cols[c]|f.boxes[b])&bit != 0 {
			continue
		}
		f.grid[i] = d
		f.rows[r] |= bit
		f.cols[c] |= bit
		f.boxes[b] |= bit
		if f.next(i + 1) {
			return true
		}
		f.grid[i] = 0
		f.rows[r] &^= bit
		f.cols[c] &^= bit
		f.boxes[b] &^= bit
	}
	return false
}

func boxOf(i int) int { return i/27*3 + i%9/3 }
