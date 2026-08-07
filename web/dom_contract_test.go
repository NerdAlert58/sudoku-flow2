package web

import (
	"regexp"
	"strings"
	"testing"
)

// DOM-contract layer for F-11 AC-1 (EVAL.md "UI (SPA)" layer (a)): static
// assertions over the embedded assets. The behavioral layer is the operator
// visual smoke (AC-2); this file pins CSP safety, the element hooks for the
// 11-item checklist, and the frozen design tokens.

func mustAsset(t *testing.T, name string) string {
	t.Helper()
	b, err := FS.ReadFile(name)
	if err != nil {
		t.Fatalf("FS.ReadFile(%q): %v", name, err)
	}
	return string(b)
}

func htmlTags(doc string) []string {
	return regexp.MustCompile(`<[a-zA-Z][^>]*>`).FindAllString(doc, -1)
}

func tagWith(tags []string, needles ...string) string {
	for _, tag := range tags {
		ok := true
		for _, n := range needles {
			if !strings.Contains(tag, n) {
				ok = false
				break
			}
		}
		if ok {
			return tag
		}
	}
	return ""
}

func normalizeRef(s string) string {
	return strings.TrimPrefix(strings.TrimPrefix(s, "./"), "/")
}

var ladderTechniques = []string{
	"naked_single",
	"hidden_single",
	"locked_candidates_pointing",
	"locked_candidates_claiming",
	"naked_subset",
	"hidden_subset",
	"x_wing",
	"swordfish",
	"jellyfish",
	"xy_wing",
	"xyz_wing",
	"w_wing",
	"simple_colouring",
}

// --- CSP safety (AUDIT.md S1) ---

func TestIndexCSPSafe(t *testing.T) {
	doc := mustAsset(t, "index.html")

	for _, m := range regexp.MustCompile(`(?is)<script\b([^>]*)>(.*?)</script>`).FindAllStringSubmatch(doc, -1) {
		if !regexp.MustCompile(`(?i)\bsrc\s*=`).MatchString(m[1]) {
			t.Errorf("script tag without src: %q", m[0])
		}
		if strings.TrimSpace(m[2]) != "" {
			t.Errorf("inline script body: %q", strings.TrimSpace(m[2]))
		}
	}
	if regexp.MustCompile(`(?i)<style\b`).MatchString(doc) {
		t.Error("inline <style> tag present")
	}
	styleAttr := regexp.MustCompile(`(?i)[\s"']style\s*=`)
	onHandler := regexp.MustCompile(`(?i)\son[a-z]+\s*=`)
	for _, tag := range htmlTags(doc) {
		if styleAttr.MatchString(tag) {
			t.Errorf("style attribute in tag %q", tag)
		}
		if onHandler.MatchString(tag) {
			t.Errorf("inline event handler in tag %q", tag)
		}
	}
}

func TestAppJSCSPSafe(t *testing.T) {
	js := mustAsset(t, "app.js")
	for _, banned := range []string{"innerHTML", "insertAdjacentHTML", "document.write"} {
		if strings.Contains(js, banned) {
			t.Errorf("app.js contains forbidden %s", banned)
		}
	}
	if regexp.MustCompile(`\beval\s*\(`).MatchString(js) {
		t.Error("app.js contains eval(")
	}
	if regexp.MustCompile(`\bnew\s+Function\b`).MatchString(js) {
		t.Error("app.js contains new Function")
	}
}

func TestIndexReferencesExactlyAppAssets(t *testing.T) {
	doc := mustAsset(t, "index.html")
	tags := htmlTags(doc)

	hrefRE := regexp.MustCompile(`(?i)\bhref\s*=\s*"([^"]*)"`)
	var sheets []string
	for _, tag := range tags {
		if !strings.HasPrefix(strings.ToLower(tag), "<link") ||
			!regexp.MustCompile(`(?i)\brel\s*=\s*"stylesheet"`).MatchString(tag) {
			continue
		}
		m := hrefRE.FindStringSubmatch(tag)
		if m == nil {
			t.Errorf("stylesheet link without href: %q", tag)
			continue
		}
		sheets = append(sheets, normalizeRef(m[1]))
	}
	if len(sheets) != 1 || sheets[0] != "app.css" {
		t.Errorf("stylesheet links = %v, want exactly [app.css]", sheets)
	}

	srcRE := regexp.MustCompile(`(?i)\bsrc\s*=\s*"([^"]*)"`)
	var scripts []string
	for _, tag := range tags {
		if !strings.HasPrefix(strings.ToLower(tag), "<script") {
			continue
		}
		m := srcRE.FindStringSubmatch(tag)
		if m == nil {
			t.Errorf("script tag without src: %q", tag)
			continue
		}
		scripts = append(scripts, normalizeRef(m[1]))
	}
	if len(scripts) != 1 || scripts[0] != "app.js" {
		t.Errorf("script srcs = %v, want exactly [app.js]", scripts)
	}
}

// --- Element hooks: the 11-item checklist (EVAL.md row "UI (SPA)") ---

func TestChecklist01GridCells(t *testing.T) {
	doc := mustAsset(t, "index.html")
	js := mustAsset(t, "app.js")
	if tagWith(htmlTags(doc), `id="grid"`) == "" {
		t.Error(`index.html: no element with id="grid"`)
	}
	combined := strings.ToLower(doc + js)
	for _, marker := range []string{"inputmode", "numeric", "maxlength"} {
		if !strings.Contains(combined, marker) {
			t.Errorf("cell input marker %q missing from index.html+app.js", marker)
		}
	}
	for _, fragment := range []string{"Row ", ", Column "} {
		if !strings.Contains(js, fragment) {
			t.Errorf("app.js: aria-label template fragment %q missing", fragment)
		}
	}
}

func TestChecklist02SeedSelect(t *testing.T) {
	doc := mustAsset(t, "index.html")
	if tagWith(htmlTags(doc), "<select", `id="seed-select"`) == "" {
		t.Error(`index.html: no <select> with id="seed-select"`)
	}
	if !strings.Contains(mustAsset(t, "app.js"), "/v1/puzzles") {
		t.Error("app.js: no /v1/puzzles reference to feed the seed dropdown")
	}
}

func TestChecklist03EntryControls(t *testing.T) {
	tags := htmlTags(mustAsset(t, "index.html"))
	if tagWith(tags, "<input", `id="paste-input"`) == "" &&
		tagWith(tags, "<textarea", `id="paste-input"`) == "" {
		t.Error(`index.html: no native <input>/<textarea> with id="paste-input"`)
	}
	for _, id := range []string{"clear-btn", "solve-btn"} {
		if tagWith(tags, "<button", `id="`+id+`"`) == "" {
			t.Errorf("index.html: no native <button> with id=%q", id)
		}
	}
}

func TestChecklist04StatusRegion(t *testing.T) {
	tags := htmlTags(mustAsset(t, "index.html"))
	if tagWith(tags, `id="status"`, `aria-live="polite"`) == "" {
		t.Error(`index.html: no element carrying both id="status" and aria-live="polite"`)
	}
}

func TestChecklist05StatsAndTechniqueStrings(t *testing.T) {
	if tagWith(htmlTags(mustAsset(t, "index.html")), `id="stats"`) == "" {
		t.Error(`index.html: no element with id="stats"`)
	}
	js := mustAsset(t, "app.js")
	for _, tech := range ladderTechniques {
		if !strings.Contains(js, tech) {
			t.Errorf("app.js: frozen technique string %q missing", tech)
		}
	}
}

func TestChecklist06TransportControls(t *testing.T) {
	tags := htmlTags(mustAsset(t, "index.html"))
	for _, id := range []string{"btn-first", "btn-prev", "btn-play", "btn-next", "btn-last"} {
		if tagWith(tags, "<button", `id="`+id+`"`) == "" {
			t.Errorf("index.html: no native <button> with id=%q", id)
		}
	}
	for _, id := range []string{"step-pos", "step-desc"} {
		if tagWith(tags, `id="`+id+`"`) == "" {
			t.Errorf("index.html: no element with id=%q", id)
		}
	}
}

func TestChecklist07EventLog(t *testing.T) {
	tags := htmlTags(mustAsset(t, "index.html"))
	if tagWith(tags, "<ul", `id="event-log"`) == "" &&
		tagWith(tags, "<ol", `id="event-log"`) == "" {
		t.Error(`index.html: no <ul>/<ol> list with id="event-log"`)
	}
	if !regexp.MustCompile(`addEventListener\(\s*['"]click['"]`).MatchString(mustAsset(t, "app.js")) {
		t.Error(`app.js: no addEventListener("click", ...) wiring`)
	}
}

func TestChecklist08ExplainPanel(t *testing.T) {
	doc := mustAsset(t, "index.html")
	tags := htmlTags(doc)
	for _, id := range []string{"explain", "band-chip"} {
		if tagWith(tags, `id="`+id+`"`) == "" {
			t.Errorf("index.html: no element with id=%q", id)
		}
	}
	m := regexp.MustCompile(`(?i)\bid="hint"[^>]*>([^<]*)`).FindStringSubmatch(doc)
	if m == nil || len(strings.TrimSpace(m[1])) < 10 {
		t.Error(`index.html: no element id="hint" with static pre-solve instructional text as its direct content`)
	}
}

func TestChecklist09DesignTokensAndGridRecipe(t *testing.T) {
	css := mustAsset(t, "app.css")
	for _, class := range []string{".placement", ".witness", ".elimination"} {
		if !strings.Contains(css, class) {
			t.Errorf("app.css: highlight class %s missing", class)
		}
	}
	for _, token := range []string{
		"--bg", "--surface", "--border", "--text", "--text-muted",
		"--accent", "--accent-hover",
		"--space-1", "--space-2", "--space-3", "--space-4", "--space-5",
	} {
		if !regexp.MustCompile(regexp.QuoteMeta(token) + `\s*:`).MatchString(css) {
			t.Errorf("app.css: frozen token %s not declared", token)
		}
	}
	if !regexp.MustCompile(`--accent\s*:\s*#2563eb`).MatchString(css) {
		t.Error("app.css: --accent must be exactly #2563eb")
	}
	if !strings.Contains(css, "grid-template-columns") {
		t.Error("app.css: no grid-template-columns (gap-as-gridlines recipe)")
	}
	if !strings.Contains(css, "3px") {
		t.Error("app.css: no 3px value (box-boundary gridline width)")
	}
}

func TestChecklist10FetchEndpoints(t *testing.T) {
	js := mustAsset(t, "app.js")
	if !strings.Contains(js, "fetch(") {
		t.Error("app.js: no fetch( call")
	}
	for _, path := range []string{"/v1/puzzles", "/v1/solve"} {
		if !strings.Contains(js, path) {
			t.Errorf("app.js: endpoint %s missing", path)
		}
	}
}

func TestChecklist11NoClientSideSolving(t *testing.T) {
	js := mustAsset(t, "app.js")
	if regexp.MustCompile(`function\s+solve\s*\(`).MatchString(js) {
		t.Error("app.js: defines function solve( — client-side solving forbidden (AC-5)")
	}
	if regexp.MustCompile(`(?i)backtrack`).MatchString(js) {
		t.Error("app.js: contains backtrack — client-side solving forbidden (AC-5)")
	}
}

func TestAssetSizesAreRealImplementation(t *testing.T) {
	if n := len(mustAsset(t, "app.js")); n <= 3*1024 {
		t.Errorf("app.js is %d bytes, want > 3072 (stub-sized)", n)
	}
	if n := len(mustAsset(t, "index.html")); n <= 1024 {
		t.Errorf("index.html is %d bytes, want > 1024 (stub-sized)", n)
	}
}
