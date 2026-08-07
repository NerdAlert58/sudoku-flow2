package solver_test

import (
	"os"
	"strings"
	"testing"

	"github.com/NerdAlert58/sudoku-flow2/solver"
)

func veryHardGrids(tb testing.TB) []solver.Grid {
	tb.Helper()
	raw, err := os.ReadFile("../puzzles.txt")
	if err != nil {
		tb.Fatalf("read corpus: %v", err)
	}
	var sections [][]string
	for _, line := range strings.Split(string(raw), "\n") {
		line = strings.TrimSpace(line)
		switch {
		case line == "":
		case strings.HasPrefix(line, "#"):
			sections = append(sections, nil)
		default:
			if len(sections) == 0 {
				tb.Fatalf("corpus: puzzle line before first section header: %q", line)
			}
			sections[len(sections)-1] = append(sections[len(sections)-1], line)
		}
	}
	if len(sections) != 4 {
		tb.Fatalf("corpus has %d sections, want 4 (VERY HARD is the 4th)", len(sections))
	}
	seeds := sections[3]
	if len(seeds) != 10 {
		tb.Fatalf("VERY HARD section has %d seeds, want 10", len(seeds))
	}
	grids := make([]solver.Grid, len(seeds))
	for i, seed := range seeds {
		g, err := solver.Parse(seed)
		if err != nil {
			tb.Fatalf("Parse(%q): %v", seed, err)
		}
		grids[i] = g
	}
	return grids
}

func BenchmarkSolveSequential(b *testing.B) {
	grids := veryHardGrids(b)
	for b.Loop() {
		for _, g := range grids {
			if res := solver.Solve(g); res.Status != "solved" {
				b.Fatalf("Solve: Status = %q, want solved", res.Status)
			}
		}
	}
}

func BenchmarkSolveScanParallel(b *testing.B) {
	grids := veryHardGrids(b)
	for b.Loop() {
		for _, g := range grids {
			if res := solver.SolveScanParallel(g); res.Status != "solved" {
				b.Fatalf("SolveScanParallel: Status = %q, want solved", res.Status)
			}
		}
	}
}
