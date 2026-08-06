package httpapi_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"runtime"
	"strings"
	"testing"

	"github.com/NerdAlert58/sudoku-flow2/httpapi"
)

type envelope struct {
	Error string `json:"error"`
	Code  string `json:"code"`
}

func decodeEnvelope(t *testing.T, rr *httptest.ResponseRecorder, wantStatus int, wantCode string) {
	t.Helper()
	if rr.Code != wantStatus {
		t.Fatalf("status = %d, want %d (body %q)", rr.Code, wantStatus, rr.Body.String())
	}
	if ct := rr.Header().Get("Content-Type"); !strings.HasPrefix(ct, "application/json") {
		t.Errorf("Content-Type = %q, want application/json", ct)
	}
	dec := json.NewDecoder(bytes.NewReader(rr.Body.Bytes()))
	dec.DisallowUnknownFields()
	var env envelope
	if err := dec.Decode(&env); err != nil {
		t.Fatalf("envelope decode: %v (body %q)", err, rr.Body.String())
	}
	if env.Code != wantCode {
		t.Errorf("code = %q, want %q", env.Code, wantCode)
	}
	if env.Error == "" {
		t.Error("error message is empty")
	}
}

func TestHealthEndpoint(t *testing.T) {
	h := httpapi.New()
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/v1/health", nil))

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 (body %q)", rr.Code, rr.Body.String())
	}
	if ct := rr.Header().Get("Content-Type"); !strings.HasPrefix(ct, "application/json") {
		t.Errorf("Content-Type = %q, want application/json", ct)
	}
	want := `{"status":"ok","goVersion":"` + runtime.Version() + `","apiVersion":"1"}`
	if got := strings.TrimSuffix(rr.Body.String(), "\n"); got != want {
		t.Errorf("body = %q, want %q", got, want)
	}
}

func TestUnknownV1PathReturns404Envelope(t *testing.T) {
	h := httpapi.New()
	for _, path := range []string{"/v1/", "/v1/nope", "/v1/health/extra", "/v1/does/not/exist"} {
		t.Run(path, func(t *testing.T) {
			rr := httptest.NewRecorder()
			h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, path, nil))
			decodeEnvelope(t, rr, http.StatusNotFound, "not_found")
		})
	}
}

func TestHealthWrongMethodReturns405Envelope(t *testing.T) {
	h := httpapi.New()
	methods := []string{
		http.MethodPost,
		http.MethodPut,
		http.MethodDelete,
		http.MethodPatch,
		http.MethodHead,
		http.MethodOptions,
	}
	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			rr := httptest.NewRecorder()
			h.ServeHTTP(rr, httptest.NewRequest(method, "/v1/health", nil))
			decodeEnvelope(t, rr, http.StatusMethodNotAllowed, "method_not_allowed")
			if allow := rr.Header().Get("Allow"); allow != http.MethodGet {
				t.Errorf("Allow = %q, want %q", allow, http.MethodGet)
			}
		})
	}
}

func TestAccessLogLineOnHealth(t *testing.T) {
	var buf bytes.Buffer
	h := httpapi.NewWithLogWriter(&buf)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/v1/health", nil))

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rr.Code)
	}
	trimmed := strings.TrimRight(buf.String(), "\n")
	if trimmed == "" {
		t.Fatal("no access-log output")
	}
	lines := strings.Split(trimmed, "\n")
	if len(lines) != 1 {
		t.Fatalf("access-log lines = %d, want 1: %q", len(lines), buf.String())
	}
	var m map[string]any
	if err := json.Unmarshal([]byte(lines[0]), &m); err != nil {
		t.Fatalf("log line is not JSON: %v: %q", err, lines[0])
	}
	if m["method"] != http.MethodGet {
		t.Errorf("log method = %v, want %q", m["method"], http.MethodGet)
	}
	if m["path"] != "/v1/health" {
		t.Errorf("log path = %v, want %q", m["path"], "/v1/health")
	}
	if m["status"] != float64(200) {
		t.Errorf("log status = %v, want 200", m["status"])
	}
	if v, ok := m["duration"]; !ok || v == nil {
		t.Error("log duration field missing")
	}
}
