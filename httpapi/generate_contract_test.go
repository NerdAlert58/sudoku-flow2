package httpapi_test

// F-10 AC-3 (wire half) — POST /v1/generate. The wire carries no RNG seed, so
// these tests assert shape + grade discipline over real generations (3 per
// band); the seeded 25/band matrix is F-08's package-level suite, and the
// error→status seam is pinned white-box in generate_seam_test.go.
//
// Pinned interpretation (flagged for the builder): ADR-0009 says "grade can
// never differ from difficulty" and PRD §Domain context maps inputs
// easy→Easy…; together they force the response `difficulty` field to be the
// canonical capitalized band, byte-equal to `grade` — not a lowercase echo.

import (
	"net/http"
	"testing"

	"github.com/NerdAlert58/sudoku-flow2/httpapi"
	"github.com/NerdAlert58/sudoku-flow2/solver"
)

func TestGenerateContractShapeAndGrade(t *testing.T) {
	h := httpapi.New()
	for _, req := range []string{"easy", "medium", "hard", "expert"} {
		for run := 0; run < 3; run++ {
			t.Run(req, func(t *testing.T) {
				rr := postJSON(t, h, "/v1/generate", `{"difficulty":"`+req+`"}`)
				if rr.Code != http.StatusOK {
					t.Fatalf("status = %d, want 200 (body %q)", rr.Code, rr.Body.String())
				}
				assertKeyOrder(t, rr.Body.Bytes(), generateFieldOrder)
				var resp generateResponse
				strictDecode(t, rr.Body.Bytes(), &resp)
				if resp.Grade != bandOf[req] {
					t.Errorf("grade = %q, want %q for difficulty %q", resp.Grade, bandOf[req], req)
				}
				if resp.Difficulty != resp.Grade {
					t.Errorf("difficulty = %q, grade = %q; ADR-0009 pins them byte-equal", resp.Difficulty, resp.Grade)
				}
				if len(resp.Puzzle) != 81 {
					t.Fatalf("puzzle length = %d, want 81", len(resp.Puzzle))
				}
				for i := 0; i < len(resp.Puzzle); i++ {
					if c := resp.Puzzle[i]; c < '0' || c > '9' {
						t.Fatalf("puzzle[%d] = %q, want 0-9 (corpus/gridAfter digit convention)", i, c)
					}
				}
				oracleSolution(t, resp.Puzzle)
				res := solver.Solve(mustParseGrid(t, resp.Puzzle))
				if res.Status != "solved" {
					t.Errorf("ladder status = %q, want solved", res.Status)
				}
				if res.Grade != resp.Grade {
					t.Errorf("solver grade = %q, response grade = %q; must match by construction", res.Grade, resp.Grade)
				}
			})
		}
	}
}

func TestGenerateContractUnknownDifficulty400(t *testing.T) {
	h := httpapi.New()
	cases := []struct {
		name, body string
	}{
		{"unknown word", `{"difficulty":"brutal"}`},
		{"empty string", `{"difficulty":""}`},
		{"missing key", `{}`},
		{"wrong case Easy", `{"difficulty":"Easy"}`},
		{"wrong case EASY", `{"difficulty":"EASY"}`},
		{"padded", `{"difficulty":" easy"}`},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			rr := postJSON(t, h, "/v1/generate", c.body)
			decodeEnvelope(t, rr, http.StatusBadRequest, "invalid_input")
			assertNoInternals(t, rr.Body.String())
			assertFrozenHeaders(t, rr.Header())
		})
	}
}
