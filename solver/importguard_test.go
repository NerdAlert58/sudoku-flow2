package solver_test

import (
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"strings"
	"testing"
)

const modulePath = "github.com/NerdAlert58/sudoku-flow2"

// shippedImports parses every non-test .go file in the module and returns
// module-root-relative file path -> import paths. The guard is source-level
// (go/parser, ImportsOnly), so it holds regardless of build tags or whether
// the package currently compiles.
func shippedImports(t *testing.T) map[string][]string {
	t.Helper()
	const root = ".."
	out := map[string][]string{}
	fset := token.NewFileSet()
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if path == root {
				return nil
			}
			if name := d.Name(); strings.HasPrefix(name, ".") || name == "testdata" || name == "node_modules" {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(d.Name(), ".go") || strings.HasSuffix(d.Name(), "_test.go") {
			return nil
		}
		f, perr := parser.ParseFile(fset, path, nil, parser.ImportsOnly)
		if perr != nil {
			return perr
		}
		rel, rerr := filepath.Rel(root, path)
		if rerr != nil {
			return rerr
		}
		var imps []string
		for _, imp := range f.Imports {
			imps = append(imps, strings.Trim(imp.Path.Value, `"`))
		}
		out[filepath.ToSlash(rel)] = imps
		return nil
	})
	if err != nil {
		t.Fatalf("walk module source: %v", err)
	}
	if len(out) == 0 {
		t.Fatal("no non-test .go files found; the containment guard has nothing to check")
	}
	return out
}

// AC-5(b), eval row "Solve-path containment": no non-test package in the
// module imports the test-only oracle (ADR-0002, ARCHITECTURE C6). This keeps
// the replay proof non-circular — oracle stays unreachable from shipped code.
// (AC-5(a), solver-imports-stdlib-only, lives in imports_test.go since F-03.)
func TestContainment_NoShippedCodeImportsOracle(t *testing.T) {
	for file, imps := range shippedImports(t) {
		for _, imp := range imps {
			if imp == modulePath+"/oracle" {
				t.Errorf("%s imports %s: shipped code must never import the oracle", file, imp)
			}
		}
	}
}

// AC-5(c), the ARCHITECTURE C3 sealing claim: generate is importable by
// httpapi alone. The generate package does not exist yet (F-08); until it
// lands this passes vacuously over an importer set that must already be empty,
// then enforces the seal the moment any file imports it.
func TestContainment_GenerateImportedOnlyByHTTPAPI(t *testing.T) {
	for file, imps := range shippedImports(t) {
		for _, imp := range imps {
			if imp == modulePath+"/generate" && !strings.HasPrefix(file, "httpapi/") {
				t.Errorf("%s imports %s: only httpapi may import generate", file, imp)
			}
		}
	}
}
