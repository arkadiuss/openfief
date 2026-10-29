package ui

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"go/parser"
	"go/token"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// TestReadsOnlyThroughPublicAPI checks D-6 at the source: the UI's code
// imports no OpenFief package but the shared kernel and the SDK's
// generated clients, so it can reach the record only through the public
// read API, never the core's own contexts or adapters. The AuthService,
// ModelPortfolioService and BondLotService clients are the core's contracts
// for view plugins: they log the owner in and hand the model portfolio
// (D-52) and manual bond lots (D-9, D-49) to the core.
func TestReadsOnlyThroughPublicAPI(t *testing.T) {
	t.Parallel()
	files, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatal(err)
	}
	allowed := []string{
		"github.com/arkadiuss/openfief/internal/kernel",
		"github.com/arkadiuss/openfief/sdk/go/gen/openfief/views/v1",
		"github.com/arkadiuss/openfief/sdk/go/gen/openfief/identity/v1",
		"github.com/arkadiuss/openfief/sdk/go/gen/openfief/portfolio/v1",
		"github.com/arkadiuss/openfief/sdk/go/gen/openfief/ingest/v1",
		"github.com/arkadiuss/openfief/sdk/go/openfief",
	}
	fset := token.NewFileSet()
	for _, file := range files {
		if strings.HasSuffix(file, "_test.go") {
			continue
		}
		f, err := parser.ParseFile(fset, file, nil, parser.ImportsOnly)
		if err != nil {
			t.Fatal(err)
		}
		for _, imp := range f.Imports {
			path, err := strconv.Unquote(imp.Path.Value)
			if err != nil {
				t.Fatal(err)
			}
			if strings.HasPrefix(path, "github.com/arkadiuss/openfief/") && !allowedImport(path, allowed) {
				t.Errorf("%s imports %s: the UI reads only through the public read API (D-6)", file, path)
			}
		}
	}
}

// allowedImport reports whether path is one of allowed or below one.
func allowedImport(path string, allowed []string) bool {
	for _, a := range allowed {
		if path == a || strings.HasPrefix(path, a+"/") {
			return true
		}
	}
	return false
}

// TestNewNeedsClients checks that the UI refuses to start without a client
// for either of the core's services.
func TestNewNeedsClients(t *testing.T) {
	t.Parallel()
	if _, err := New(&Config{}); !errors.Is(err, ErrMissingDependency) {
		t.Errorf("New(Config{}) error = %v, want ErrMissingDependency", err)
	}
}

// TestVendoredLibraries checks that the vendored htmx and Chart.js are
// byte for byte the pinned npm releases: htmx.org 2.0.11's dist/htmx.min.js
// and chart.js 4.5.1's dist/chart.umd.min.js.
func TestVendoredLibraries(t *testing.T) {
	t.Parallel()
	want := map[string]string{
		"web/static/vendor/htmx-2.0.11.min.js":     "d6fdc75f204e6bdefa99b69bf1e6d4ac69b8a364f77929f45c13476b4000f717",
		"web/static/vendor/chart-4.5.1.umd.min.js": "48444a82d4edcb5bec0f1965faacdde18d9c17db3063d042abada2f705c9f54a",
	}
	for path, sum := range want {
		data, err := webFS.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		got := sha256.Sum256(data)
		if hex.EncodeToString(got[:]) != sum {
			t.Errorf("%s: sha256 %x, want %s", path, got, sum)
		}
	}
}
