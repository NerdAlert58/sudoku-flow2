package solver_test

import (
	"go/parser"
	"go/token"
	"os"
	"strings"
	"testing"
)

func TestSolverPackage_ImportsStdlibOnly(t *testing.T) {
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatalf("read package dir: %v", err)
	}
	var sourceFiles []string
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		sourceFiles = append(sourceFiles, name)
	}
	if len(sourceFiles) == 0 {
		t.Fatal("no non-test .go files in the solver package; the import guard has nothing to check")
	}
	fset := token.NewFileSet()
	for _, name := range sourceFiles {
		f, err := parser.ParseFile(fset, name, nil, parser.ImportsOnly)
		if err != nil {
			t.Fatalf("parse %s: %v", name, err)
		}
		for _, imp := range f.Imports {
			path := strings.Trim(imp.Path.Value, `"`)
			// Stdlib import paths never contain a dot in the first segment;
			// module-path and external imports always do (ADR-0002).
			if first, _, _ := strings.Cut(path, "/"); strings.Contains(first, ".") {
				t.Errorf("%s imports %q: solver must import only the Go standard library", name, path)
			}
		}
	}
}
