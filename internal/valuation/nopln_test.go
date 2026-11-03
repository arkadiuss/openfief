package valuation_test

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestNoAssumedPLN checks D-180 in the code: no valuation, view or model
// portfolio code, its queries and codecs included, spells out PLN. Every
// currency a stored valuation or rate names comes from the jurisdiction
// pack, and only the Polish pack and the NBP adapter, which aren't scanned,
// may name PLN. Comments are not code and are left out.
func TestNoAssumedPLN(t *testing.T) {
	t.Parallel()
	dirs := []string{
		".", "app", "postgres", filepath.Join("..", "views"), filepath.Join("..", "views", "app"),
		filepath.Join("..", "portfolio"), filepath.Join("..", "portfolio", "app"),
	}
	for _, dir := range dirs {
		for _, found := range assumedPLN(t, dir) {
			t.Errorf("assumes PLN: %s", found)
		}
	}
}

// assumedPLN returns, as file:line: text, the lines of code of the
// non-test Go files in dir that spell out PLN, comments left out.
func assumedPLN(t *testing.T, dir string) []string {
	t.Helper()
	files := os.DirFS(dir)
	names, err := fs.Glob(files, "*.go")
	if err != nil {
		t.Fatalf("listing %s: %v", dir, err)
	}
	var out []string
	for _, name := range names {
		if strings.HasSuffix(name, "_test.go") {
			continue
		}
		data, err := fs.ReadFile(files, name)
		if err != nil {
			t.Fatalf("reading %s: %v", name, err)
		}
		for i, line := range strings.Split(string(data), "\n") {
			if code, _, _ := strings.Cut(line, "//"); strings.Contains(code, "PLN") {
				out = append(out, fmt.Sprintf("%s:%d: %s", filepath.Join(dir, name), i+1, strings.TrimSpace(line)))
			}
		}
	}
	return out
}
