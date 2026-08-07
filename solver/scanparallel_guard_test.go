package solver_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"strings"
	"testing"
)

// ADR-0015: SolveScanParallel is reachable only by explicit call from the
// committed benchmark inside the solver package; any reference from non-test
// code outside solver breaches solve-path containment.
func TestSolveScanParallel_NoReferenceOutsideSolverPackage(t *testing.T) {
	root := ".."
	solverDir := filepath.Join(root, "solver")
	fset := token.NewFileSet()
	scanned := 0
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			if path != root && strings.HasPrefix(d.Name(), ".") {
				return filepath.SkipDir
			}
			if path == solverDir || d.Name() == "node_modules" || d.Name() == "testdata" {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(d.Name(), ".go") || strings.HasSuffix(d.Name(), "_test.go") {
			return nil
		}
		f, err := parser.ParseFile(fset, path, nil, parser.SkipObjectResolution)
		if err != nil {
			return err
		}
		scanned++
		ast.Inspect(f, func(n ast.Node) bool {
			if id, ok := n.(*ast.Ident); ok && id.Name == "SolveScanParallel" {
				t.Errorf("%s: references SolveScanParallel; the scan-parallel variant must stay off every serving path (ADR-0015)", fset.Position(id.Pos()))
			}
			return true
		})
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if scanned == 0 {
		t.Fatal("no non-test .go files found outside the solver package; the containment guard has nothing to check")
	}
}
