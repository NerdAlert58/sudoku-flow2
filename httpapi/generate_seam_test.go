package httpapi

// F-10 AC-3 (error half) — SEAM PIN, white-box by design. The wire cannot
// inject generator failures (no RNG/clock seam crosses the frozen contract),
// so the builder MUST expose exactly this in-package seam:
//
//	func mapGenerateError(err error) (int, string)
//
// used by the /v1/generate handler to map generate.Generate errors onto the
// envelope. Pins (ADR-0004/ADR-0009): generate.ErrUnknownBand → 400
// "invalid_input"; generate.ErrBudgetExhausted (wrapped or not) → 500
// "generation_failed"; any other generate error (e.g. the 5s context
// deadline surfacing directly) → 500 "generation_failed". errors.Is
// semantics — wrapped sentinels must map identically. This file cannot
// compile until F-08 exports both sentinels and the builder adds the seam;
// that is the intended RED state.

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/NerdAlert58/sudoku-flow2/generate"
)

func TestMapGenerateErrorSeam(t *testing.T) {
	cases := []struct {
		name       string
		err        error
		wantStatus int
		wantCode   string
	}{
		{"unknown band", generate.ErrUnknownBand, http.StatusBadRequest, "invalid_input"},
		{"unknown band wrapped", fmt.Errorf("generate: %w", generate.ErrUnknownBand), http.StatusBadRequest, "invalid_input"},
		{"budget exhausted", generate.ErrBudgetExhausted, http.StatusInternalServerError, "generation_failed"},
		{"budget exhausted wrapped", fmt.Errorf("generate: band expert: %w", generate.ErrBudgetExhausted), http.StatusInternalServerError, "generation_failed"},
		{"bare context deadline", context.DeadlineExceeded, http.StatusInternalServerError, "generation_failed"},
		{"any other error", errors.New("filler wedged"), http.StatusInternalServerError, "generation_failed"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			status, code := mapGenerateError(c.err)
			if status != c.wantStatus || code != c.wantCode {
				t.Errorf("mapGenerateError(%v) = (%d, %q), want (%d, %q)",
					c.err, status, code, c.wantStatus, c.wantCode)
			}
		})
	}
}
