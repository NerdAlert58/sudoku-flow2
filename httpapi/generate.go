package httpapi

import (
	"context"
	crand "crypto/rand"
	"encoding/binary"
	"errors"
	"math/rand"
	"net/http"
	"time"

	"github.com/NerdAlert58/sudoku-flow2/generate"
)

func handleGenerate(w http.ResponseWriter, r *http.Request) {
	if !gatePostJSON(w, r) {
		return
	}
	var req generateRequest
	if !decodeBody(w, r, &req) {
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	puzzle, grade, err := generate.Generate(ctx, req.Difficulty, newRand())
	if err != nil {
		status, code := mapGenerateError(err)
		msg := "puzzle generation failed; try again"
		if code == "invalid_input" {
			msg = "unknown difficulty; expected easy, medium, hard, or expert"
		}
		writeEnvelope(w, status, code, msg)
		return
	}
	// difficulty == grade by construction (ADR-0009): both are the canonical
	// band Generate accepted, never a lowercase echo of the request.
	writeJSON(w, http.StatusOK, generateResponse{Puzzle: puzzle, Difficulty: grade, Grade: grade})
}

// mapGenerateError maps generate.Generate failures onto the envelope
// (ADR-0004/ADR-0009): unknown band → 400 invalid_input; everything else —
// budget exhaustion, a bare context deadline, any unexpected error — is an
// honest 500 generation_failed. errors.Is covers wrapped sentinels.
func mapGenerateError(err error) (int, string) {
	if errors.Is(err, generate.ErrUnknownBand) {
		return http.StatusBadRequest, "invalid_input"
	}
	return http.StatusInternalServerError, "generation_failed"
}

// newRand crypto-seeds a per-request math/rand source: shuffling needs no
// cryptographic stream, but concurrent requests must not share a seed.
// crypto/rand.Read cannot fail (Go 1.24+ aborts the program instead).
func newRand() *rand.Rand {
	var seed [8]byte
	crand.Read(seed[:])
	return rand.New(rand.NewSource(int64(binary.LittleEndian.Uint64(seed[:]))))
}
