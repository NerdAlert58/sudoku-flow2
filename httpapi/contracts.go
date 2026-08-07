package httpapi

// Wire shapes for /v1 (contract C1). Field order is contract: encoding/json
// preserves struct order (AUDIT.md A5).

import (
	"github.com/NerdAlert58/sudoku-flow2/catalog"
	"github.com/NerdAlert58/sudoku-flow2/solver"
)

type healthResponse struct {
	Status     string `json:"status"`
	GoVersion  string `json:"goVersion"`
	APIVersion string `json:"apiVersion"`
}

type errorEnvelope struct {
	Error string `json:"error"`
	Code  string `json:"code"`
}

type solveRequest struct {
	Puzzle string `json:"puzzle"`
}

// solveResponse carries both domain outcomes of /v1/solve: 200 for
// solved/unsolvable/stalled and the full-shape 400 for invalid_input
// (ADR-0004). grade has no omitempty — the key is always present (AUDIT A5).
type solveResponse struct {
	APIVersion      string         `json:"apiVersion"`
	Input           string         `json:"input"`
	Status          string         `json:"status"`
	Solved          bool           `json:"solved"`
	Solution        string         `json:"solution"`
	Iterations      int            `json:"iterations"`
	EventCount      int            `json:"eventCount"`
	CandidateChecks int            `json:"candidateChecks"`
	SolveTimeMs     float64        `json:"solveTimeMs"`
	Grade           string         `json:"grade"`
	Events          []solver.Event `json:"events"`
}

type generateRequest struct {
	Difficulty string `json:"difficulty"`
}

type generateResponse struct {
	Puzzle     string `json:"puzzle"`
	Difficulty string `json:"difficulty"`
	Grade      string `json:"grade"`
}

type batchRequest struct {
	Puzzles []string `json:"puzzles"`
}

type batchItem struct {
	Puzzle           string  `json:"puzzle"`
	Solved           bool    `json:"solved"`
	SolveTimeMs      float64 `json:"solveTimeMs"`
	Iterations       int     `json:"iterations"`
	HardestTechnique string  `json:"hardestTechnique"`
}

type batchResponse struct {
	APIVersion  string      `json:"apiVersion"`
	Results     []batchItem `json:"results"`
	SolvedCount int         `json:"solvedCount"`
	Total       int         `json:"total"`
}

type puzzlesResponse struct {
	Sections []catalog.Section `json:"sections"`
}
