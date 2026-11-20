package answer

import (
	"go/parser"
	"go/token"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"testing"
)

// TestReadsOnlyThroughPublicAPI checks D-6 at the source: the answerers
// import no OpenFief package but the shared kernel and the SDK's generated
// read API contract, so what a card shows, for the page or for MCP, is what
// the read API shows the caller's token, never the core's own contexts or
// adapters, and it knows nothing of the UI that draws it.
func TestReadsOnlyThroughPublicAPI(t *testing.T) {
	t.Parallel()
	files, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatal(err)
	}
	allowed := []string{
		"github.com/arkadiuss/openfief/internal/kernel",
		"github.com/arkadiuss/openfief/sdk/go/gen/openfief/views/v1",
	}
	for _, file := range files {
		for _, imp := range openfiefImports(t, file) {
			if !slices.ContainsFunc(allowed, func(a string) bool { return imp == a || strings.HasPrefix(imp, a+"/") }) {
				t.Errorf("%s imports %s: the answerers read only through the public read API (D-6)", file, imp)
			}
		}
	}
}

// openfiefImports returns the OpenFief packages file imports.
func openfiefImports(t *testing.T, file string) []string {
	t.Helper()
	f, err := parser.ParseFile(token.NewFileSet(), file, nil, parser.ImportsOnly)
	if err != nil {
		t.Fatal(err)
	}
	var out []string
	for _, spec := range f.Imports {
		path, err := strconv.Unquote(spec.Path.Value)
		if err != nil {
			t.Fatal(err)
		}
		if strings.HasPrefix(path, "github.com/arkadiuss/openfief/") {
			out = append(out, path)
		}
	}
	return out
}
