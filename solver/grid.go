package solver

import (
	"errors"
	"fmt"
)

var (
	ErrBadLength      = errors.New("grid must be exactly 81 characters")
	ErrBadChar        = errors.New("grid may contain only 1-9 for givens and 0 or . for blanks")
	ErrDuplicateGiven = errors.New("duplicate given in a row, column, or box")
)

// Grid holds digit values 0-9 in row-major order; 0 is an empty cell.
type Grid [81]byte

func Parse(s string) (Grid, error) {
	if len(s) != 81 {
		return Grid{}, fmt.Errorf("solver: got %d characters: %w", len(s), ErrBadLength)
	}
	var g Grid
	for i := 0; i < 81; i++ {
		d, err := parseCell(s[i])
		if err != nil {
			return Grid{}, fmt.Errorf("solver: position %d: %w", i, err)
		}
		g[i] = d
	}
	if err := checkDuplicateGivens(&g); err != nil {
		return Grid{}, err
	}
	return g, nil
}

func parseCell(c byte) (byte, error) {
	switch {
	case c >= '1' && c <= '9':
		return c - '0', nil
	case c == '0' || c == '.':
		return 0, nil
	default:
		return 0, fmt.Errorf("%q: %w", c, ErrBadChar)
	}
}

func checkDuplicateGivens(g *Grid) error {
	for u := range units {
		var seen [10]bool
		for _, i := range units[u] {
			d := g[i]
			if d == 0 {
				continue
			}
			if seen[d] {
				return fmt.Errorf("solver: digit %d appears twice in %s: %w", d, unitName(u), ErrDuplicateGiven)
			}
			seen[d] = true
		}
	}
	return nil
}

func (g Grid) String() string {
	var b [81]byte
	for i, d := range g {
		b[i] = '0' + d
	}
	return string(b[:])
}

func (g *Grid) complete() bool {
	for _, d := range g {
		if d == 0 {
			return false
		}
	}
	return true
}

// units enumerates the 27 houses in the ADR-0007 canonical order:
// rows 0-8, then columns 0-8, then boxes 0-8 row-major.
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

func cellUnitIndexes(i int) [3]int {
	r, c := i/9, i%9
	return [3]int{r, 9 + c, 18 + r/3*3 + c/3}
}

func unitName(u int) string {
	switch {
	case u < 9:
		return fmt.Sprintf("row %d", u)
	case u < 18:
		return fmt.Sprintf("column %d", u-9)
	default:
		return fmt.Sprintf("box %d", u-18)
	}
}
