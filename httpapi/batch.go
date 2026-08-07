package httpapi

import (
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/NerdAlert58/sudoku-flow2/solver"
)

// maxBatchItems is the goroutine-count bound: over-cap 413s after parse and
// before any solving (AUDIT.md A7); the cap is inclusive.
const maxBatchItems = 256

func handleValidateBatch(w http.ResponseWriter, r *http.Request) {
	if !gatePostJSON(w, r) {
		return
	}
	var req batchRequest
	if !decodeBody(w, r, &req) {
		return
	}
	if len(req.Puzzles) > maxBatchItems {
		writeEnvelope(w, http.StatusRequestEntityTooLarge, "payload_too_large", "batch exceeds 256 puzzles")
		return
	}
	// Goroutine-per-puzzle into a pre-sized slice: each goroutine owns index
	// i alone, and solver counters are per-solve state (ADR-0007), so the
	// fan-out is race-free with no shared mutable state.
	results := make([]batchItem, len(req.Puzzles))
	var wg sync.WaitGroup
	for i, raw := range req.Puzzles {
		wg.Go(func() { results[i] = solveBatchItem(raw) })
	}
	wg.Wait()
	solved := 0
	for _, item := range results {
		if item.Solved {
			solved++
		}
	}
	writeJSON(w, http.StatusOK, batchResponse{
		APIVersion:  "1",
		Results:     results,
		SolvedCount: solved,
		Total:       len(req.Puzzles),
	})
}

// solveBatchItem produces the ADR-0014 item: puzzle echoes the raw string
// byte-for-byte, trimming applies to parsing only, and a malformed line is
// exactly {raw, false, 0, 0, ""}.
func solveBatchItem(raw string) batchItem {
	g, err := solver.Parse(strings.TrimSpace(raw))
	if err != nil {
		return batchItem{Puzzle: raw}
	}
	start := time.Now()
	res := solver.Solve(g)
	elapsed := time.Since(start)
	return batchItem{
		Puzzle:           raw,
		Solved:           res.Status == "solved",
		SolveTimeMs:      msOf(elapsed),
		Iterations:       res.Iterations,
		HardestTechnique: hardestTechnique(res.Events),
	}
}

// ladderRank mirrors the frozen 13-technique ladder (PRD §Domain context);
// the value is the ladder position for the ADR-0014 hardestTechnique rule.
// The solver's registry is unexported by design — this duplicate is contract
// data, pinned against drift by the F-10 batch corpus test.
var ladderRank = map[string]int{
	"naked_single":               1,
	"hidden_single":              2,
	"locked_candidates_pointing": 3,
	"locked_candidates_claiming": 4,
	"naked_subset":               5,
	"hidden_subset":              6,
	"x_wing":                     7,
	"swordfish":                  8,
	"jellyfish":                  9,
	"xy_wing":                    10,
	"xyz_wing":                   11,
	"w_wing":                     12,
	"simple_colouring":           13,
}

// hardestTechnique is the highest-ladder-position technique that fired during
// the attempt, regardless of final status; "" when none fired (ADR-0014).
func hardestTechnique(events []solver.Event) string {
	best, name := 0, ""
	for _, ev := range events {
		if rank := ladderRank[ev.Technique]; rank > best {
			best, name = rank, ev.Technique
		}
	}
	return name
}
