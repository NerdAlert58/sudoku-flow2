package catalog

import (
	"fmt"
	"strings"
	"testing"
)

const validPuzzle = "700605000000000030509300024002000000401907052000501000004050000310492000007003000"

func synth(counts []int) []string {
	headers := []string{
		"# === ORIGINAL (unlabeled) ===",
		"# === MEDIUM ===",
		"# === HARD ===",
		"# === VERY HARD ===",
		"# === EXTRA ===",
	}
	var lines []string
	for i, n := range counts {
		lines = append(lines, headers[i])
		for j := 0; j < n; j++ {
			lines = append(lines, validPuzzle)
		}
		lines = append(lines, "")
	}
	return lines
}

func blob(lines []string) []byte {
	return []byte(strings.Join(lines, "\n"))
}

func TestParseCatalogRejectsMalformed(t *testing.T) {
	fiveSections := synth([]int{25, 10, 10, 10, 10})

	shortLine := synth([]int{25, 10, 10, 10})
	shortLine[1] = validPuzzle[:80]

	badChar := synth([]int{25, 10, 10, 10})
	badChar[1] = "x" + validPuzzle[1:]

	orphanPuzzle := append([]string{validPuzzle}, synth([]int{25, 10, 10, 10})...)

	cases := []struct {
		name string
		data []byte
	}{
		{"five sections", blob(fiveSections)},
		{"80-char puzzle line", blob(shortLine)},
		{"non-digit character", blob(badChar)},
		{"empty input", []byte{}},
		{"puzzle before first header", blob(orphanPuzzle)},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := parseCatalog(tc.data); err == nil {
				t.Fatalf("parseCatalog accepted %s; want non-nil error", tc.name)
			}
		})
	}
}

func TestMustParsePanicsOnMalformed(t *testing.T) {
	if got := len(mustParse(blob(synth([]int{25, 10, 10, 10})))); got != 4 {
		t.Fatalf("mustParse on valid fixture returned %d sections; want 4", got)
	}
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("mustParse on malformed input did not panic")
		}
		if fmt.Sprint(r) == "" {
			t.Fatal("mustParse panicked with an empty message")
		}
	}()
	mustParse([]byte("not-a-catalog"))
}

func TestSectionsEmbeddedDataParses(t *testing.T) {
	if got := len(Sections()); got != 4 {
		t.Fatalf("Sections() on embedded data returned %d sections; want 4", got)
	}
}
