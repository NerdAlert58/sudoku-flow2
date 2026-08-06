package solver_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/NerdAlert58/sudoku-flow2/solver"
)

const seedZeroBlanks = "700605000000000030509300024002000000401907052000501000004050000310492000007003000"

func gridString(cells map[int]byte) string {
	b := []byte(strings.Repeat("0", 81))
	for i, d := range cells {
		b[i] = d
	}
	return string(b)
}

func gridWithChar(c byte) string {
	b := []byte(strings.Repeat("0", 81))
	b[40] = c
	return string(b)
}

func TestParse_Grammar(t *testing.T) {
	valid := []struct {
		name, in, want string
	}{
		{"zero blanks", seedZeroBlanks, seedZeroBlanks},
		{"dot blanks", strings.ReplaceAll(seedZeroBlanks, "0", "."), seedZeroBlanks},
		{"all blank zeros", strings.Repeat("0", 81), strings.Repeat("0", 81)},
		{"all blank dots", strings.Repeat(".", 81), strings.Repeat("0", 81)},
		{"complete grid", completeGrid, completeGrid},
	}
	for _, tc := range valid {
		t.Run("valid/"+tc.name, func(t *testing.T) {
			g, err := solver.Parse(tc.in)
			if err != nil {
				t.Fatalf("Parse(%q) error: %v", tc.in, err)
			}
			if got := g.String(); got != tc.want {
				t.Fatalf("String() = %q, want %q", got, tc.want)
			}
		})
	}

	invalid := []struct {
		name string
		in   string
		want error
	}{
		{"length 0", "", solver.ErrBadLength},
		{"length 80", strings.Repeat("0", 80), solver.ErrBadLength},
		{"length 82", strings.Repeat("0", 82), solver.ErrBadLength},
		{"letter", gridWithChar('a'), solver.ErrBadChar},
		{"space", gridWithChar(' '), solver.ErrBadChar},
		{"dash", gridWithChar('-'), solver.ErrBadChar},
		{"newline", gridWithChar('\n'), solver.ErrBadChar},
		{"duplicate in row", gridString(map[int]byte{0: '5', 5: '5'}), solver.ErrDuplicateGiven},
		{"duplicate in column", gridString(map[int]byte{0: '5', 36: '5'}), solver.ErrDuplicateGiven},
		{"duplicate in box", gridString(map[int]byte{0: '5', 10: '5'}), solver.ErrDuplicateGiven},
	}
	for _, tc := range invalid {
		t.Run("invalid/"+tc.name, func(t *testing.T) {
			_, err := solver.Parse(tc.in)
			if err == nil {
				t.Fatalf("Parse(%q): want error %v, got nil", tc.in, tc.want)
			}
			if !errors.Is(err, tc.want) {
				t.Fatalf("Parse(%q) error = %v, want errors.Is(err, %v)", tc.in, err, tc.want)
			}
		})
	}
}
