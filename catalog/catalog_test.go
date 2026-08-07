package catalog_test

import (
	"bytes"
	"encoding/json"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/NerdAlert58/sudoku-flow2/catalog"
)

var puzzleRE = regexp.MustCompile(`^[0-9]{81}$`)

func TestSectionsNamesAndCounts(t *testing.T) {
	secs := catalog.Sections()
	want := []struct {
		name  string
		count int
	}{
		{"Original", 25},
		{"Medium", 10},
		{"Hard", 10},
		{"Very Hard", 10},
	}
	if len(secs) != len(want) {
		t.Fatalf("Sections() returned %d sections; want %d", len(secs), len(want))
	}
	for i, w := range want {
		if secs[i].Name != w.name {
			t.Errorf("section %d name = %q; want %q", i, secs[i].Name, w.name)
		}
		if len(secs[i].Puzzles) != w.count {
			t.Errorf("section %d (%s) has %d puzzles; want %d", i, w.name, len(secs[i].Puzzles), w.count)
		}
	}
}

func TestEveryPuzzleIs81Digits(t *testing.T) {
	for _, s := range catalog.Sections() {
		for i, p := range s.Puzzles {
			if !puzzleRE.MatchString(p) {
				t.Errorf("section %s puzzle %d = %q; want match for ^[0-9]{81}$", s.Name, i, p)
			}
		}
	}
}

func TestPuzzlesInFileOrder(t *testing.T) {
	src, err := os.ReadFile("../puzzles.txt")
	if err != nil {
		t.Fatalf("read repo-root puzzles.txt: %v", err)
	}
	var want []string
	for _, line := range strings.Split(string(src), "\n") {
		if puzzleRE.MatchString(line) {
			want = append(want, line)
		}
	}
	var got []string
	for _, s := range catalog.Sections() {
		got = append(got, s.Puzzles...)
	}
	if len(got) != len(want) {
		t.Fatalf("catalog serves %d puzzles; repo-root puzzles.txt has %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("puzzle %d out of file order: got %q; want %q", i, got[i], want[i])
		}
	}
}

func TestSectionJSONWireShape(t *testing.T) {
	b, err := json.Marshal(catalog.Section{Name: "Original", Puzzles: []string{"p"}})
	if err != nil {
		t.Fatalf("marshal Section: %v", err)
	}
	want := `{"name":"Original","puzzles":["p"]}`
	if string(b) != want {
		t.Fatalf("Section JSON = %s; want %s", b, want)
	}
}

func TestEmbeddedCopyMatchesRepoRoot(t *testing.T) {
	src, err := os.ReadFile("../puzzles.txt")
	if err != nil {
		t.Fatalf("read repo-root puzzles.txt: %v", err)
	}
	if !bytes.Equal(catalog.Raw(), src) {
		t.Fatal("catalog/puzzles.txt drifted from repo-root puzzles.txt; update the embedded copy in the same change")
	}
}

// AUDIT D3: section names are provenance labels, never solver-grade claims;
// no test in this repo may assert grade == section name.
func TestSectionNamesAreProvenanceLabelsOnly(t *testing.T) {
	canonical := map[string]bool{"Original": true, "Medium": true, "Hard": true, "Very Hard": true}
	for i, s := range catalog.Sections() {
		if !canonical[s.Name] {
			t.Errorf("section %d name %q is outside the canonical provenance set {Original, Medium, Hard, Very Hard}", i, s.Name)
		}
	}
}
