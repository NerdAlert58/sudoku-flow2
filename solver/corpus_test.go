package solver_test

import (
	"os"
	"strings"
	"testing"

	"github.com/NerdAlert58/sudoku-flow2/solver"
)

func readCorpusSections(t *testing.T) [][]string {
	t.Helper()
	raw, err := os.ReadFile("../puzzles.txt")
	if err != nil {
		t.Fatalf("read corpus: %v", err)
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
				t.Fatalf("corpus: puzzle line before first section header: %q", line)
			}
			sections[len(sections)-1] = append(sections[len(sections)-1], line)
		}
	}
	return sections
}

func originalSeeds(t *testing.T) []string {
	t.Helper()
	sections := readCorpusSections(t)
	if len(sections) == 0 {
		t.Fatal("corpus: no sections found")
	}
	if len(sections[0]) != 25 {
		t.Fatalf("corpus: first (ORIGINAL) section has %d puzzles, want 25", len(sections[0]))
	}
	return sections[0]
}

func mediumSeed(t *testing.T) string {
	t.Helper()
	sections := readCorpusSections(t)
	if len(sections) < 2 || len(sections[1]) == 0 {
		t.Fatal("corpus: no MEDIUM section")
	}
	return sections[1][0]
}

func mustParse(t *testing.T, s string) solver.Grid {
	t.Helper()
	g, err := solver.Parse(s)
	if err != nil {
		t.Fatalf("Parse(%q): %v", s, err)
	}
	return g
}

// Independent rule-conformance check: never trusts solver internals (AUDIT.md L4 spirit;
// the oracle cross-check arrives in F-06).
func assertRuleConformantSolution(t *testing.T, seed, sol string) {
	t.Helper()
	if len(sol) != 81 {
		t.Fatalf("solution length = %d, want 81", len(sol))
	}
	for i := 0; i < 81; i++ {
		if sol[i] < '1' || sol[i] > '9' {
			t.Fatalf("solution cell %d = %q, want 1-9", i, sol[i])
		}
		if seed[i] >= '1' && seed[i] <= '9' && sol[i] != seed[i] {
			t.Fatalf("solution cell %d = %q, but given was %q", i, sol[i], seed[i])
		}
	}
	unit := func(kind string, u int, cells [9]int) {
		var seen [10]bool
		for _, i := range cells {
			d := int(sol[i] - '0')
			if seen[d] {
				t.Fatalf("%s %d: digit %d appears twice", kind, u, d)
			}
			seen[d] = true
		}
	}
	for u := 0; u < 9; u++ {
		var row, col, box [9]int
		for k := 0; k < 9; k++ {
			row[k] = u*9 + k
			col[k] = k*9 + u
			box[k] = (u/3*3+k/3)*9 + u%3*3 + k%3
		}
		unit("row", u, row)
		unit("col", u, col)
		unit("box", u, box)
	}
}

func TestSolve_OriginalCorpus_SinglesOnly(t *testing.T) {
	for i, seed := range originalSeeds(t) {
		res := solver.Solve(mustParse(t, seed))
		if res.Status != "solved" {
			t.Fatalf("seed %d: Status = %q, want %q", i, res.Status, "solved")
		}
		if res.Grade != "Easy" {
			t.Fatalf("seed %d: Grade = %q, want %q (singles-only tier)", i, res.Grade, "Easy")
		}
		assertRuleConformantSolution(t, seed, res.Solution.String())
		for _, ev := range res.Events {
			if ev.Technique != "naked_single" && ev.Technique != "hidden_single" {
				t.Fatalf("seed %d event %d: Technique = %q, want naked_single or hidden_single", i, ev.Seq, ev.Technique)
			}
		}
	}
}
