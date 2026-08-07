package httpapi

import (
	"net/http"
	"time"

	"github.com/NerdAlert58/sudoku-flow2/solver"
)

func handleSolve(w http.ResponseWriter, r *http.Request) {
	if !gatePostJSON(w, r) {
		return
	}
	var req solveRequest
	if !decodeBody(w, r, &req) {
		return
	}
	g, err := solver.Parse(req.Puzzle)
	if err != nil {
		// Domain-invalid puzzles get the FULL solve shape at 400, not the
		// envelope, with the raw input echoed byte-for-byte (ADR-0004).
		writeJSON(w, http.StatusBadRequest, invalidSolveResponse(req.Puzzle))
		return
	}
	start := time.Now()
	res := solver.Solve(g)
	elapsed := time.Since(start)
	writeJSON(w, http.StatusOK, solveResponseFrom(req.Puzzle, res, msOf(elapsed)))
}

// msOf is the ADR-0006 metric: float64 milliseconds of the solver.Solve wall
// clock only, measured in the handler, excluded from byte-identity claims.
func msOf(d time.Duration) float64 {
	return d.Seconds() * 1e3
}

func invalidSolveResponse(input string) solveResponse {
	return solveResponse{
		APIVersion: "1",
		Input:      input,
		Status:     "invalid_input",
		Events:     []solver.Event{},
	}
}

func solveResponseFrom(input string, res solver.SolveResult, ms float64) solveResponse {
	events := res.Events
	if events == nil {
		events = []solver.Event{}
	}
	return solveResponse{
		APIVersion:      "1",
		Input:           input,
		Status:          res.Status,
		Solved:          res.Status == "solved",
		Solution:        res.Solution.String(),
		Iterations:      res.Iterations,
		EventCount:      res.EventCount,
		CandidateChecks: res.CandidateChecks,
		SolveTimeMs:     ms,
		Grade:           res.Grade,
		Events:          events,
	}
}
