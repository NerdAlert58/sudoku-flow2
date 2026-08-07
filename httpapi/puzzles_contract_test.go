package httpapi_test

// F-10 AC-5 — GET /v1/puzzles must be exactly the JSON marshal of
// catalog.Sections() wrapped as {sections:[...]} (contract C4 → C1).

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/NerdAlert58/sudoku-flow2/catalog"
	"github.com/NerdAlert58/sudoku-flow2/httpapi"
)

func TestPuzzlesContractMatchesCatalog(t *testing.T) {
	h := httpapi.New()
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/v1/puzzles", nil))
	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 (body %q)", rr.Code, rr.Body.String())
	}
	if ct := rr.Header().Get("Content-Type"); !strings.HasPrefix(ct, "application/json") {
		t.Errorf("Content-Type = %q, want application/json", ct)
	}
	want, err := json.Marshal(struct {
		Sections []catalog.Section `json:"sections"`
	}{catalog.Sections()})
	if err != nil {
		t.Fatal(err)
	}
	if got := trimmedBody(rr); got != string(want) {
		t.Errorf("body != marshal of catalog.Sections():\n got %.120s...\nwant %.120s...", got, want)
	}
	assertKeyOrder(t, rr.Body.Bytes(), []string{"sections"})

	var resp struct {
		Sections []catalog.Section `json:"sections"`
	}
	strictDecode(t, rr.Body.Bytes(), &resp)
	wantSections := []struct {
		name  string
		count int
	}{{"Original", 25}, {"Medium", 10}, {"Hard", 10}, {"Very Hard", 10}}
	if len(resp.Sections) != len(wantSections) {
		t.Fatalf("sections = %d, want 4", len(resp.Sections))
	}
	for i, ws := range wantSections {
		if resp.Sections[i].Name != ws.name || len(resp.Sections[i].Puzzles) != ws.count {
			t.Errorf("section %d = %q(%d), want %q(%d)", i,
				resp.Sections[i].Name, len(resp.Sections[i].Puzzles), ws.name, ws.count)
		}
	}
}
