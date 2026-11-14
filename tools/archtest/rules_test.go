package archtest_test

import (
	"os"
	"slices"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"golang.org/x/tools/txtar"

	"github.com/arkadiuss/openfief/tools/archtest"
)

// fixtureDigest is the one digest the fixtures' catalog pins (synthetic).
const fixtureDigest = "sha256:0000000000000000000000000000000000000000000000000000000000000001"

// loadFixture writes a txtar archive as a module example.com/fix in a
// temporary directory and loads it outside the workspace, offline.
func loadFixture(t *testing.T, archive string) *archtest.Repo {
	t.Helper()
	dir := t.TempDir()
	fsys, err := txtar.FS(txtar.Parse([]byte("-- go.mod --\nmodule example.com/fix\n\ngo 1.26.0\n" + archive)))
	if err != nil {
		t.Fatalf("fixture: %v", err)
	}
	if cerr := os.CopyFS(dir, fsys); cerr != nil {
		t.Fatalf("write fixture: %v", cerr)
	}
	repo, err := archtest.Load(&archtest.Config{
		Root:     dir,
		Module:   "example.com/fix",
		Patterns: []string{"./..."},
		Walk:     []string{dir},
		Env:      append(os.Environ(), "GOWORK=off", "GOFLAGS=-mod=mod", "GOPROXY=off"),
	})
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	return repo
}

func TestRules(t *testing.T) {
	t.Parallel()
	catalogTests := func(r *archtest.Repo) []archtest.Violation { return archtest.CatalogTests(r, []string{fixtureDigest}) }
	tests := []struct {
		name    string
		rule    func(*archtest.Repo) []archtest.Violation
		archive string
		want    []string
	}{
		{"context map: an allowed arrow", archtest.ContextMap, ledgerDomain + `
-- internal/valuation/v.go --
package valuation
import "example.com/fix/internal/ledger"
var _ = ledger.Account{}
`, nil},
		{"context map: an arrow the map doesn't name", archtest.ContextMap, `
-- internal/portfolio/p.go --
package portfolio
type Sleeve struct{}
-- internal/ingest/app/a.go --
package app
import "example.com/fix/internal/portfolio"
var _ = portfolio.Sleeve{}
`, []string{"internal/ingest/app -> internal/portfolio"}},
		{"context map: the kernel imports a context, an unknown context", archtest.ContextMap, ledgerDomain + `
-- internal/kernel/k.go --
package kernel
import "example.com/fix/internal/ledger"
var _ = ledger.Account{}
-- internal/stuff/s.go --
package stuff
`, []string{"internal/kernel -> internal/ledger", "internal/stuff"}},
		{"context map: the demo dataset reads domains and only cmd imports it", archtest.ContextMap, ledgerDomain + ledgerApp + `
-- internal/demo/d.go --
package demo
import (
	"example.com/fix/internal/ledger"
	"example.com/fix/internal/ledger/app"
)
var _, _ = ledger.Account{}, app.Use
-- internal/views/v.go --
package views
import "example.com/fix/internal/demo"
var _ = demo.Use
-- internal/demo/use.go --
package demo
var Use = 1
-- cmd/seed/main.go --
package main
import "example.com/fix/internal/demo"
func main() { _ = demo.Use }
`, []string{"internal/demo -> internal/ledger/app", "internal/views -> internal/demo"}},
		{"layering: domain, app, adapters", archtest.Layering, ledgerDomain + ledgerApp + ledgerPostgres + `
-- internal/ledger/l2.go --
package ledger
import "example.com/fix/internal/ledger/app"
var _ = app.Use
-- internal/ledger/app/a2.go --
package app
import "example.com/fix/internal/ledger/postgres"
var _ = postgres.Store{}
-- internal/valuation/postgres/p.go --
package postgres
import lp "example.com/fix/internal/ledger/postgres"
var _ = lp.Store{}
-- internal/ledger/postgres/pgtest/t.go --
package pgtest
import "example.com/fix/internal/ledger/postgres"
var _ = postgres.Store{}
-- tools/x/main.go --
package main
import "example.com/fix/internal/ledger/postgres"
func main() { _ = postgres.Store{} }
-- cmd/core/main.go --
package main
import "example.com/fix/internal/ledger/postgres"
func main() { _ = postgres.Store{} }
`, []string{
			"internal/ledger -> internal/ledger/app",
			"internal/ledger/app -> internal/ledger/postgres",
			"internal/valuation/postgres -> internal/ledger/postgres",
			"tools/x -> internal/ledger/postgres",
		}},
		{"plugins: only the SDK", archtest.Plugins, kernel + `
-- sdk/go/openfief/s.go --
package openfief
import "example.com/fix/internal/kernel"
var _ = kernel.Money{}
-- plugins/fx/main.go --
package main
import (
	"example.com/fix/internal/kernel"
	"example.com/fix/sdk/go/openfief"
)
func main() { _, _ = kernel.Money{}, openfief.Use }
-- sdk/go/openfief/use.go --
package openfief
var Use = 1
`, []string{"plugins/fx -> internal/kernel", "sdk/go/openfief -> internal/kernel"}},
		{"no network in transactions: a PostgreSQL adapter imports no client", archtest.NoNetworkInTransactions, `
-- internal/ledger/postgres/p.go --
package postgres
import "net/http"
var _ = http.StatusOK
-- internal/valuation/postgres/pgtest/t.go --
package pgtest
import "net/http"
var _ = http.StatusOK
-- internal/ingest/postgres/q.go --
package postgres
import "strings"
var _ = strings.ToLower
-- internal/views/ui/u.go --
package ui
import "net/http"
var _ = http.StatusOK
-- cmd/core/main.go --
package main
import "net/http"
func main() { _ = http.StatusOK }
`, []string{"internal/ledger/postgres -> net/http", "internal/valuation/postgres/pgtest -> net/http"}},
		{"banned packages, test-only ones included", archtest.Banned, `
-- internal/ledger/util/u.go --
package util
-- test/helpers/h_test.go --
package helpers
`, []string{"internal/ledger/util", "test/helpers"}},
		{"schemas: only the context's own", archtest.Schemas, `
-- internal/portfolio/postgres/migrations/0001.sql --
-- Accounts are named as valuation.tax_wrappers names them (a comment).
CREATE TABLE portfolio.sleeves (id text);
CREATE VIEW portfolio.wrapped AS SELECT * FROM valuation.tax_wrappers;
-- internal/ingest/postgres/s.go --
package postgres
const q = "SELECT id FROM ledger.accounts JOIN ingest.bindings USING (id)"
const msg = "ledger.Account not found"
-- internal/valuation/postgres/s.go --
package postgres
const jobs = "SELECT kind FROM river.river_job"
`, []string{"internal/ingest/postgres/s.go names ledger", "internal/portfolio/postgres/migrations/0001.sql names valuation"}},
		{"money: no floats in domain, app and contracts", archtest.Money, `
-- internal/ledger/l.go --
package ledger
func Ratio(a, b int) float64 { return float64(a) / float64(b) }
func Count(a int) int { return a }
-- internal/ledger/app/a.go --
package app
type Point struct{ X float32 }
-- internal/views/ui/chart.go --
package ui
type Point struct{ X float64 }
-- proto/fix/v1/fix.proto --
syntax = "proto3";
message Quote {
  string amount = 1;
  double price = 2;
}
`, []string{"internal/ledger Ratio", "internal/ledger/app Point", "proto/fix/v1/fix.proto price"}},
		{"pure domain: goroutines and package state", archtest.PureDomain, `
-- internal/ledger/l.go --
package ledger
import ("errors"; "regexp")
var ErrBroken = errors.New("broken")
var pattern = regexp.MustCompile("x")
var _ = pattern
func Run(f func()) { go f() }
-- internal/ledger/app/a.go --
package app
var cache = map[string]int{}
func Run(f func()) { go f(); _ = cache }
`, []string{"internal/ledger Run go", "internal/ledger var pattern"}},
		{"privacy: sensitive values and raw errors", archtest.Privacy, kernel + `
-- internal/ledger/postgres/p.go --
package postgres
import (
	"log/slog"
	"example.com/fix/internal/kernel"
)
type Store struct{ log *slog.Logger }
func (s Store) Amount(m kernel.Money) { s.log.Info("booked", "amount", m) }
func (s Store) Text(m kernel.Money) { s.log.Info("booked", slog.String("amount", m.String())) }
func (s Store) Raw(err error) { s.log.Error("failed", "error", err) }
func (s Store) Msg(err error) { slog.Warn("failed", slog.String("error", err.Error())) }
func (s Store) Safe(err error) { s.log.Error("failed", "error", kernel.LogSafe(err)) }
func (s Store) ID(id string) { s.log.Info("booked", "transaction", id) }
`, []string{
			"internal/ledger/postgres Store.Amount logs money",
			"internal/ledger/postgres Store.Msg logs an error",
			"internal/ledger/postgres Store.Raw logs an error",
			"internal/ledger/postgres Store.Text logs money",
		}},
		{"stable tests: no sleeps, no skips", archtest.StableTests, `
-- internal/ledger/l_test.go --
package ledger
import (
	clock "time"
	"testing"
)
func TestWait(t *testing.T) { clock.Sleep(clock.Second) }
func TestSkip(t *testing.T) { t.Run("x", func(tt *testing.T) { tt.Skipf("later") }) }
func helper(tb testing.TB) { tb.SkipNow() }
func TestParser(t *testing.T) { var p parser; p.Skip() }
`, []string{"internal/ledger/l_test.go TestSkip Skip", "internal/ledger/l_test.go TestWait time.Sleep", "internal/ledger/l_test.go helper Skip"}},
		{"catalog tests: pinned sizes, indexes and digests", catalogTests, catalogPkg + `
-- internal/plugins/catalog/c_test.go --
package catalog_test
import "example.com/fix/internal/plugins/catalog"
func TestEmpty(t *testing.T) {
	c, _ := catalog.Load()
	if len(c.Entries) != 0 { t.Error("not empty") }
}
func TestFirst(t *testing.T) {
	c, _ := catalog.Load()
	if c.Entries[0].Name != "nbp-fx" { t.Error("first") }
}
func TestDigest(t *testing.T) {
	want := "ghcr.io/example/x@sha256:0000000000000000000000000000000000000000000000000000000000000001"
	_ = want
}
func TestDerived(t *testing.T) {
	c, _ := catalog.Load()
	for i := range len(c.Entries) { _ = c.Entries[i] }
	if len(c.Entries) != len(c.Entries) { t.Error("never") }
}
`, []string{
			"internal/plugins/catalog/c_test.go TestDigest digest",
			"internal/plugins/catalog/c_test.go TestEmpty compares the number of catalog entries with a literal",
			"internal/plugins/catalog/c_test.go TestFirst picks a catalog entry by a literal index",
		}},
		{"catalog tests: a test of a function that reads the catalog", catalogTests, catalogPkg + `
-- cmd/op/main.go --
package main
import "example.com/fix/internal/plugins/catalog"
func trustPolicy() ([]string, error) { c, err := catalog.Load(); return []string{c.Entries[0].Name}, err }
func verifier() ([]string, error) { return trustPolicy() }
func main() { _, _ = verifier() }
-- cmd/op/main_test.go --
package main
import "example.com/fix/internal/plugins/catalog"
func TestPinned(t *testing.T) {
	got, _ := verifier()
	if len(got) != 1 { t.Error("pinned") }
}
func TestErrorOnly(t *testing.T) {
	if _, err := verifier(); err != nil { t.Error(err) }
}
func TestDerived(t *testing.T) {
	got, _ := trustPolicy()
	if len(got) != len(names(t)) { t.Error("derived") }
}
func names(t *testing.T) []string {
	c, _ := catalog.Load()
	var out []string
	for _, e := range c.Entries { out = append(out, e.Name) }
	return out
}
`, []string{"cmd/op/main_test.go TestPinned calls verifier"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var got []string
			for _, v := range tt.rule(loadFixture(t, tt.archive)) {
				got = append(got, v.Key)
			}
			slices.Sort(got)
			if diff := cmp.Diff(tt.want, got, cmpopts.EquateEmpty()); diff != "" {
				t.Errorf("violations (-want +got):\n%s", diff)
			}
		})
	}
}

// The fixtures' shared packages: a kernel with Money and LogSafe, a ledger
// domain and app, a ledger adapter, and a catalog.
const (
	kernel = `
-- internal/kernel/k.go --
package kernel
type Money struct{ amount string }
func (m Money) String() string { return m.amount }
func LogSafe(err error, known ...error) string { return "" }
`
	ledgerDomain = `
-- internal/ledger/l.go --
package ledger
type Account struct{}
`
	ledgerApp = `
-- internal/ledger/app/a.go --
package app
var Use = 1
`
	ledgerPostgres = `
-- internal/ledger/postgres/p.go --
package postgres
type Store struct{}
`
	catalogPkg = `
-- internal/plugins/catalog/c.go --
package catalog
type Entry struct{ Name string }
type Catalog struct{ Entries []Entry }
func Load() (Catalog, error) { return Catalog{}, nil }
`
)

func TestPlaceOf(t *testing.T) {
	t.Parallel()
	tests := []struct {
		path string
		want archtest.Place
	}{
		{"internal/kernel", archtest.Place{Context: "kernel", Layer: archtest.LayerKernel}},
		{"internal/ledger", archtest.Place{Context: "ledger", Layer: archtest.LayerDomain}},
		{"internal/jurisdiction/pl", archtest.Place{Context: "jurisdiction", Layer: archtest.LayerDomain}},
		{"internal/ledger/app", archtest.Place{Context: "ledger", Layer: archtest.LayerApp}},
		{"internal/plugins/cosign/cosigntest", archtest.Place{Context: "plugins", Layer: archtest.LayerAdapter, Adapter: "internal/plugins/cosign"}},
		{"cmd/core", archtest.Place{Layer: archtest.LayerOutside}},
		{"sdk/go/openfief", archtest.Place{Layer: archtest.LayerOutside}},
	}
	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			t.Parallel()
			if diff := cmp.Diff(tt.want, archtest.PlaceOf(tt.path)); diff != "" {
				t.Errorf("PlaceOf(%q) (-want +got):\n%s", tt.path, diff)
			}
		})
	}
}

func TestParseList(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		data    string
		want    []archtest.Entry
		wantErr error
	}{
		{"entries and comments", "# a comment\n\nlayering | a -> b | why\n", []archtest.Entry{{Rule: "layering", Key: "a -> b", Reason: "why", Line: 3}}, nil},
		{"no reason", "layering | a -> b | \n", nil, archtest.ErrList},
		{"unknown rule", "speed | a | why\n", nil, archtest.ErrList},
		{"listed twice", "money | a | why\nmoney | a | again\n", nil, archtest.ErrList},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := archtest.ParseList(tt.data)
			if diff := cmp.Diff(tt.wantErr, err, cmpopts.EquateErrors()); diff != "" {
				t.Fatalf("error (-want +got):\n%s", diff)
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("entries (-want +got):\n%s", diff)
			}
		})
	}
}

// TestCompare checks that the lists only shrink: a violation no list names
// is fresh, and a line that matches nothing is stale.
func TestCompare(t *testing.T) {
	t.Parallel()
	found := []archtest.Violation{
		{Rule: archtest.RuleLayering, Key: "a -> b", Detail: "listed in the baseline"},
		{Rule: archtest.RuleStableTests, Key: "x_test.go TestX Skip", Detail: "an exception"},
		{Rule: archtest.RuleMoney, Key: "internal/ledger Ratio", Detail: "new"},
	}
	baseline := []archtest.Entry{
		{Rule: archtest.RuleLayering, Key: "a -> b", Reason: "r", Line: 1},
		{Rule: archtest.RulePlugins, Key: "plugins/x -> internal/kernel", Reason: "fixed since", Line: 2},
	}
	exceptions := []archtest.Entry{{Rule: archtest.RuleStableTests, Key: "x_test.go TestX Skip", Reason: "r", Line: 1}}
	fresh, stale := archtest.Compare(found, baseline, exceptions)
	if diff := cmp.Diff(found[2:], fresh); diff != "" {
		t.Errorf("fresh (-want +got):\n%s", diff)
	}
	if diff := cmp.Diff(baseline[1:], stale); diff != "" {
		t.Errorf("stale (-want +got):\n%s", diff)
	}
	want := `[money] internal/ledger Ratio: new (docs/dev/code-organization.md#architecture-tests, "Money")`
	if got := found[2].String(); got != want {
		t.Errorf("String() = %q, want %q", got, want)
	}
}
