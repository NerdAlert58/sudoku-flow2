package httpapi_test

// F-10 AC-6 — the route×edge matrix. ADR-0004 (frozen 7-code envelope),
// ADR-0005 (in-handler method dispatch, 405+Allow), AUDIT A7 (byte cap:
// Content-Length fast path AND lazy chunked enforcement), AUDIT S2 (no
// Access-Control-* ever). Every response here also passes the frozen-header
// and no-internals floors.

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/NerdAlert58/sudoku-flow2/httpapi"
)

var postRoutes = []string{"/v1/solve", "/v1/generate", "/v1/validate-batch"}

// readTracker pins "rejected before the body is read" (AUDIT A7, PRD "413
// before reading" / "415 before body read").
type readTracker struct{ reads int }

func (r *readTracker) Read([]byte) (int, error) {
	r.reads++
	return 0, io.EOF
}

func edgeRequest(method, path string, body *readTracker, contentType string) *http.Request {
	var req *http.Request
	if body != nil {
		req = httptest.NewRequest(method, path, body)
	} else {
		req = httptest.NewRequest(method, path, nil)
	}
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	req.Header.Set("Origin", "https://evil.example.com")
	return req
}

func assertErrorFloor(t *testing.T, rr *httptest.ResponseRecorder) {
	t.Helper()
	assertFrozenHeaders(t, rr.Header())
	assertNoInternals(t, rr.Body.String())
}

func TestEdgeMatrixContentType415(t *testing.T) {
	h := httpapi.New()
	for _, route := range postRoutes {
		t.Run(route+" wrong type", func(t *testing.T) {
			body := &readTracker{}
			rr := httptest.NewRecorder()
			h.ServeHTTP(rr, edgeRequest(http.MethodPost, route, body, "text/plain"))
			decodeEnvelope(t, rr, http.StatusUnsupportedMediaType, "unsupported_media_type")
			assertErrorFloor(t, rr)
			if body.reads != 0 {
				t.Errorf("body read %d times before 415; want 0", body.reads)
			}
		})
		t.Run(route+" missing type", func(t *testing.T) {
			rr := httptest.NewRecorder()
			h.ServeHTTP(rr, edgeRequest(http.MethodPost, route, nil, ""))
			decodeEnvelope(t, rr, http.StatusUnsupportedMediaType, "unsupported_media_type")
			assertErrorFloor(t, rr)
		})
		t.Run(route+" charset parameter accepted", func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, route, strings.NewReader("{"))
			req.Header.Set("Content-Type", "application/json; charset=utf-8")
			rr := httptest.NewRecorder()
			h.ServeHTTP(rr, req)
			decodeEnvelope(t, rr, http.StatusBadRequest, "invalid_input")
		})
	}
}

func TestEdgeMatrixMalformedJSON400(t *testing.T) {
	h := httpapi.New()
	for _, route := range postRoutes {
		for _, body := range []string{"{", "not json at all", `[]`} {
			t.Run(route+" "+body, func(t *testing.T) {
				rr := postJSON(t, h, route, body)
				decodeEnvelope(t, rr, http.StatusBadRequest, "invalid_input")
				assertErrorFloor(t, rr)
			})
		}
	}
}

func TestEdgeMatrixMethodNotAllowed405(t *testing.T) {
	h := httpapi.New()
	wrongMethods := map[string][]string{
		"/v1/solve":          {http.MethodGet, http.MethodPut, http.MethodDelete, http.MethodPatch, http.MethodHead, http.MethodOptions},
		"/v1/generate":       {http.MethodGet, http.MethodPut, http.MethodDelete, http.MethodPatch, http.MethodHead, http.MethodOptions},
		"/v1/validate-batch": {http.MethodGet, http.MethodPut, http.MethodDelete, http.MethodPatch, http.MethodHead, http.MethodOptions},
		"/v1/puzzles":        {http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodPatch, http.MethodHead, http.MethodOptions},
	}
	allow := map[string]string{
		"/v1/solve": http.MethodPost, "/v1/generate": http.MethodPost,
		"/v1/validate-batch": http.MethodPost, "/v1/puzzles": http.MethodGet,
	}
	for route, methods := range wrongMethods {
		for _, method := range methods {
			t.Run(method+" "+route, func(t *testing.T) {
				rr := httptest.NewRecorder()
				h.ServeHTTP(rr, edgeRequest(method, route, nil, ""))
				decodeEnvelope(t, rr, http.StatusMethodNotAllowed, "method_not_allowed")
				if got := rr.Header().Get("Allow"); got != allow[route] {
					t.Errorf("Allow = %q, want %q", got, allow[route])
				}
				assertErrorFloor(t, rr)
			})
		}
	}
	t.Run("method dispatch precedes content-type gate", func(t *testing.T) {
		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, edgeRequest(http.MethodDelete, "/v1/solve", nil, "text/plain"))
		decodeEnvelope(t, rr, http.StatusMethodNotAllowed, "method_not_allowed")
	})
}

func TestEdgeMatrixUnknownV1Subpaths404(t *testing.T) {
	h := httpapi.New()
	paths := []string{"/v1/solve/extra", "/v1/generate/x", "/v1/validate-batch/x", "/v1/puzzles/x"}
	for _, path := range paths {
		t.Run(path, func(t *testing.T) {
			rr := httptest.NewRecorder()
			h.ServeHTTP(rr, edgeRequest(http.MethodPost, path, nil, "application/json"))
			decodeEnvelope(t, rr, http.StatusNotFound, "not_found")
			assertErrorFloor(t, rr)
		})
	}
}

func TestEdgeMatrixByteCap413FastPath(t *testing.T) {
	h := httpapi.New()
	for _, route := range postRoutes {
		t.Run(route, func(t *testing.T) {
			body := &readTracker{}
			req := edgeRequest(http.MethodPost, route, body, "application/json")
			req.ContentLength = oneMiB + 1
			rr := httptest.NewRecorder()
			h.ServeHTTP(rr, req)
			decodeEnvelope(t, rr, http.StatusRequestEntityTooLarge, "payload_too_large")
			assertErrorFloor(t, rr)
			if body.reads != 0 {
				t.Errorf("declared-oversize body read %d times; Content-Length fast path must 413 without reading (AUDIT A7)", body.reads)
			}
		})
	}
}

func TestEdgeMatrixByteCap413Chunked(t *testing.T) {
	h := httpapi.New()
	oversized := `{"pad":"` + strings.Repeat("0", oneMiB) + `"}`
	for _, route := range postRoutes {
		t.Run(route, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, route, strings.NewReader(oversized))
			req.ContentLength = -1
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Origin", "https://evil.example.com")
			rr := httptest.NewRecorder()
			h.ServeHTTP(rr, req)
			decodeEnvelope(t, rr, http.StatusRequestEntityTooLarge, "payload_too_large")
			assertErrorFloor(t, rr)
		})
	}
}
