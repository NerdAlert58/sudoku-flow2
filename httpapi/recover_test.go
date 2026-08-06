package httpapi

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestPanicRecoveryWritesEnvelopeAndOneLogLine(t *testing.T) {
	var logBuf bytes.Buffer
	h := chain(&logBuf, http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		panic("boom")
	}))
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/v1/boom", nil))

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500 (body %q)", rr.Code, rr.Body.String())
	}
	dec := json.NewDecoder(bytes.NewReader(rr.Body.Bytes()))
	dec.DisallowUnknownFields()
	var env struct {
		Error string `json:"error"`
		Code  string `json:"code"`
	}
	if err := dec.Decode(&env); err != nil {
		t.Fatalf("envelope decode: %v (body %q)", err, rr.Body.String())
	}
	if env.Code != "internal_error" {
		t.Errorf("code = %q, want %q", env.Code, "internal_error")
	}
	if env.Error == "" {
		t.Error("error message is empty")
	}

	hdr := rr.Header()
	wantHeaders := map[string]string{
		"Content-Security-Policy":   "default-src 'self'; script-src 'self'; style-src 'self'; connect-src 'self'; img-src 'self'; object-src 'none'; base-uri 'self'; frame-ancestors 'none'",
		"Strict-Transport-Security": "max-age=63072000",
		"X-Frame-Options":           "DENY",
		"X-Content-Type-Options":    "nosniff",
	}
	for k, want := range wantHeaders {
		if got := hdr.Get(k); got != want {
			t.Errorf("%s = %q, want %q", k, got, want)
		}
	}

	trimmed := strings.TrimRight(logBuf.String(), "\n")
	if trimmed == "" {
		t.Fatal("no access-log output")
	}
	lines := strings.Split(trimmed, "\n")
	if len(lines) != 1 {
		t.Fatalf("access-log lines = %d, want 1: %q", len(lines), logBuf.String())
	}
	var m map[string]any
	if err := json.Unmarshal([]byte(lines[0]), &m); err != nil {
		t.Fatalf("log line is not JSON: %v: %q", err, lines[0])
	}
	if m["method"] != http.MethodGet {
		t.Errorf("log method = %v, want %q", m["method"], http.MethodGet)
	}
	if m["path"] != "/v1/boom" {
		t.Errorf("log path = %v, want %q", m["path"], "/v1/boom")
	}
	if m["status"] != float64(500) {
		t.Errorf("log status = %v, want 500", m["status"])
	}
	if v, ok := m["duration"]; !ok || v == nil {
		t.Error("log duration field missing")
	}
}
