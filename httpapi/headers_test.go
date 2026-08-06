package httpapi_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/NerdAlert58/sudoku-flow2/httpapi"
)

const wantCSP = "default-src 'self'; script-src 'self'; style-src 'self'; connect-src 'self'; img-src 'self'; object-src 'none'; base-uri 'self'; frame-ancestors 'none'"

func assertFrozenHeaders(t *testing.T, hdr http.Header) {
	t.Helper()
	if got := hdr.Get("Content-Security-Policy"); got != wantCSP {
		t.Errorf("Content-Security-Policy = %q, want %q", got, wantCSP)
	}
	if got := hdr.Get("Strict-Transport-Security"); got != "max-age=63072000" {
		t.Errorf("Strict-Transport-Security = %q, want %q", got, "max-age=63072000")
	}
	if got := hdr.Get("X-Frame-Options"); got != "DENY" {
		t.Errorf("X-Frame-Options = %q, want %q", got, "DENY")
	}
	if got := hdr.Get("X-Content-Type-Options"); got != "nosniff" {
		t.Errorf("X-Content-Type-Options = %q, want %q", got, "nosniff")
	}
	for key := range hdr {
		if strings.HasPrefix(key, "Access-Control-") {
			t.Errorf("CORS header %q must never be emitted", key)
		}
	}
}

func TestFrozenHeadersOnEveryRoute(t *testing.T) {
	h := httpapi.New()
	routes := []struct {
		name   string
		method string
		path   string
	}{
		{"health", http.MethodGet, "/v1/health"},
		{"index", http.MethodGet, "/"},
		{"asset", http.MethodGet, "/app.css"},
		{"v1 not found", http.MethodGet, "/v1/nope"},
		{"wrong method", http.MethodPost, "/v1/health"},
		{"static not found", http.MethodGet, "/no-such-page"},
	}
	for _, rt := range routes {
		for _, origin := range []string{"", "https://evil.example.com"} {
			name := rt.name
			if origin != "" {
				name += " with origin"
			}
			t.Run(name, func(t *testing.T) {
				req := httptest.NewRequest(rt.method, rt.path, nil)
				if origin != "" {
					req.Header.Set("Origin", origin)
				}
				rr := httptest.NewRecorder()
				h.ServeHTTP(rr, req)
				assertFrozenHeaders(t, rr.Header())
			})
		}
	}
}
