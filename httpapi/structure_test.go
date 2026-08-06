package httpapi_test

import (
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func repoRoot(t *testing.T) string {
	t.Helper()
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	return filepath.Dir(filepath.Dir(thisFile))
}

func TestNoInternalDirectory(t *testing.T) {
	root := repoRoot(t)
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			return nil
		}
		if path != root && strings.HasPrefix(d.Name(), ".") {
			return filepath.SkipDir
		}
		if d.Name() == "internal" {
			t.Errorf("internal/ directory exists at %s", path)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestGoModHasZeroRequires(t *testing.T) {
	data, err := os.ReadFile(filepath.Join(repoRoot(t), "go.mod"))
	if err != nil {
		t.Fatalf("go.mod: %v", err)
	}
	for _, line := range strings.Split(string(data), "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), "require") {
			t.Errorf("go.mod contains a require line: %q", line)
		}
	}
}

func TestGoSumAbsent(t *testing.T) {
	_, err := os.Stat(filepath.Join(repoRoot(t), "go.sum"))
	if err == nil {
		t.Fatal("go.sum exists; module must be stdlib-only")
	}
	if !os.IsNotExist(err) {
		t.Fatal(err)
	}
}
