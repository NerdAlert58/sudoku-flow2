package httpapi_test

import (
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"

	"github.com/NerdAlert58/sudoku-flow2/httpapi"
)

var (
	scriptTagRe    = regexp.MustCompile(`(?i)<script\b[^>]*>`)
	styleTagRe     = regexp.MustCompile(`(?i)<style\b`)
	styleAttrRe    = regexp.MustCompile(`(?i)\bstyle\s*=`)
	eventHandlerRe = regexp.MustCompile(`(?i)\bon[a-z]+\s*=`)
)

func TestIndexPageServesExternalAssetsOnly(t *testing.T) {
	h := httpapi.New()
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/", nil))

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rr.Code)
	}
	if ct := rr.Header().Get("Content-Type"); !strings.HasPrefix(ct, "text/html") {
		t.Errorf("Content-Type = %q, want text/html", ct)
	}
	body := rr.Body.String()
	if !strings.Contains(body, "app.css") {
		t.Error("index page has no reference to app.css")
	}
	if !strings.Contains(body, "app.js") {
		t.Error("index page has no reference to app.js")
	}
	for _, tag := range scriptTagRe.FindAllString(body, -1) {
		if !strings.Contains(strings.ToLower(tag), "src") {
			t.Errorf("inline <script> tag: %q", tag)
		}
	}
	if m := styleTagRe.FindString(body); m != "" {
		t.Error("inline <style> tag present")
	}
	if m := styleAttrRe.FindString(body); m != "" {
		t.Errorf("inline style attribute: %q", m)
	}
	if m := eventHandlerRe.FindString(body); m != "" {
		t.Errorf("inline event handler attribute: %q", m)
	}
}

func TestStaticAssetsServed(t *testing.T) {
	h := httpapi.New()
	cases := []struct {
		path     string
		wantType string
	}{
		{"/app.css", "text/css"},
		{"/app.js", "javascript"},
	}
	for _, c := range cases {
		t.Run(c.path, func(t *testing.T) {
			rr := httptest.NewRecorder()
			h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, c.path, nil))
			if rr.Code != http.StatusOK {
				t.Fatalf("status = %d, want 200", rr.Code)
			}
			if ct := rr.Header().Get("Content-Type"); !strings.Contains(ct, c.wantType) {
				t.Errorf("Content-Type = %q, want containing %q", ct, c.wantType)
			}
			if rr.Body.Len() == 0 {
				t.Error("empty body")
			}
		})
	}
}
