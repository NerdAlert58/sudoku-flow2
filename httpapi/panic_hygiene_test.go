package httpapi

// F-10 AC-7 (SECURITY.md F-9) — white-box extension of the F-01 recover test.
// The 500 envelope message is pinned to the fixed generic string the F-01
// recoverPanic implementation writes; no panic value, stack frame, or file
// path may reach any wire body.

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

const wantPanicBody = `{"error":"internal server error","code":"internal_error"}`

func TestPanicHygieneFixedGenericMessage(t *testing.T) {
	panics := []struct {
		name  string
		value any
	}{
		{"string with path and line", "exploded at /Users/nerd/Git/sudoku-flow2/httpapi/routes.go:42"},
		{"error value", errors.New("goroutine 7 deadlocked in runtime.gopark")},
		{"struct value", struct{ Secret string }{"panic: internal state"}},
		{"nil map write runtime error", nil},
	}
	for _, p := range panics {
		t.Run(p.name, func(t *testing.T) {
			var logBuf bytes.Buffer
			h := chain(&logBuf, http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
				if p.value == nil {
					var m map[string]int
					m["boom"] = 1
				}
				panic(p.value)
			}))
			rr := httptest.NewRecorder()
			h.ServeHTTP(rr, httptest.NewRequest(http.MethodPost, "/v1/solve", strings.NewReader("{}")))

			if rr.Code != http.StatusInternalServerError {
				t.Fatalf("status = %d, want 500 (body %q)", rr.Code, rr.Body.String())
			}
			body := strings.TrimSuffix(rr.Body.String(), "\n")
			if body != wantPanicBody {
				t.Errorf("500 body = %q, want the fixed generic envelope %q", body, wantPanicBody)
			}
			for _, leak := range []string{"runtime", "goroutine", ".go:", "panic", "Secret", "deadlocked", "exploded"} {
				if strings.Contains(body, leak) {
					t.Errorf("500 body leaks internals (%q): %q", leak, body)
				}
			}
			if ct := rr.Header().Get("Content-Type"); !strings.HasPrefix(ct, "application/json") {
				t.Errorf("Content-Type = %q, want application/json", ct)
			}
		})
	}
}
