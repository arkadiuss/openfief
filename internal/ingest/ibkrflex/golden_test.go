package ibkrflex_test

import (
	"bytes"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/arkadiuss/openfief/internal/ingest"
	"github.com/arkadiuss/openfief/internal/ingest/ibkrflex"
	"github.com/arkadiuss/openfief/internal/ledger"
)

// updateFlag rewrites the golden files:
// go test ./internal/ingest/ibkrflex -run TestMapGolden -update.
const updateFlag = "update"

func TestMain(m *testing.M) {
	flag.Bool(updateFlag, false, "rewrite golden files")
	os.Exit(m.Run())
}

// updating reports whether the test run rewrites golden files.
func updating() bool {
	f := flag.Lookup(updateFlag)
	return f != nil && f.Value.String() == "true"
}

// testMapper feeds the synthetic IBKR account U0000000 into ledger account
// acc-ibkr; U0000001 feeds nothing.
func testMapper() ibkrflex.Mapper {
	return ibkrflex.Mapper{Accounts: map[string]ledger.AccountID{"U0000000": "acc-ibkr"}}
}

// mapFixture maps the synthetic fixture testdata/name in mode.
func mapFixture(t *testing.T, name string, mode ingest.Mode) ingest.Mapping {
	t.Helper()
	data, err := fs.ReadFile(os.DirFS("testdata"), name)
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	m, err := testMapper().Map(bytes.NewReader(data), mode)
	if err != nil {
		t.Fatalf("Map(%s, %s): %v", name, mode, err)
	}
	return m
}

// checkGolden compares got with testdata/name, or rewrites it with -update.
func checkGolden(t *testing.T, name, got string) {
	t.Helper()
	if updating() {
		if err := os.WriteFile(filepath.Join("testdata", name), []byte(got), 0o600); err != nil {
			t.Fatalf("write golden: %v", err)
		}
	}
	want, err := fs.ReadFile(os.DirFS("testdata"), name)
	if err != nil {
		t.Fatalf("read golden (run with -update to create it): %v", err)
	}
	if diff := cmp.Diff(string(want), got); diff != "" {
		t.Errorf("mapping differs from testdata/%s (-want +got):\n%s", name, diff)
	}
}

func TestMapGolden(t *testing.T) {
	t.Parallel()
	tests := []struct {
		fixture string
		mode    ingest.Mode
		golden  string
	}{
		{fixture: "opening.xml", mode: ingest.ModeOpening, golden: "opening.opening.golden"},
		{fixture: "opening.xml", mode: ingest.ModeTransactions, golden: "opening.transactions.golden"},
		{fixture: "activity.xml", mode: ingest.ModeTransactions, golden: "activity.transactions.golden"},
		{fixture: "activity.xml", mode: ingest.ModeOpening, golden: "activity.opening.golden"},
	}
	for _, tt := range tests {
		t.Run(tt.golden, func(t *testing.T) {
			t.Parallel()
			m := mapFixture(t, tt.fixture, tt.mode)
			checkGolden(t, tt.golden, render(&m))
		})
	}
}

// TestMapValidTransactions checks that every mapped transaction is a valid,
// balanced ledger transaction and every assertion a valid one.
func TestMapValidTransactions(t *testing.T) {
	t.Parallel()
	for _, fixture := range []string{"opening.xml", "activity.xml"} {
		for _, mode := range []ingest.Mode{ingest.ModeOpening, ingest.ModeTransactions} {
			m := mapFixture(t, fixture, mode)
			checkValid(t, fixture+" "+mode.String(), &m)
		}
	}
}

// checkValid fails the test for each invalid transaction or assertion in m.
func checkValid(t *testing.T, what string, m *ingest.Mapping) {
	t.Helper()
	for i := range m.Transactions {
		if err := m.Transactions[i].Validate(); err != nil {
			t.Errorf("%s: transaction %s: %v", what, m.Transactions[i].Key, err)
		}
	}
	for i := range m.Assertions {
		if err := m.Assertions[i].Validate(); err != nil {
			t.Errorf("%s: assertion %d: %v", what, i, err)
		}
	}
}

// render writes a mapping as stable text: transactions with their postings,
// then assertions, then problems, each in mapping order.
func render(m *ingest.Mapping) string {
	var b strings.Builder
	for i := range m.Transactions {
		t := &m.Transactions[i]
		fmt.Fprintf(&b, "tx %s %s %s account=%s", t.Key, t.Kind, t.Date, t.Account)
		if !t.LotOpened.IsZero() {
			fmt.Fprintf(&b, " lot-opened=%s", t.LotOpened)
		}
		fmt.Fprintf(&b, " %q\n", t.Description)
		for _, p := range t.Postings {
			fmt.Fprintf(&b, "  %s %s", p.Book, p.Amount)
			if !p.Quantity.IsZero() {
				fmt.Fprintf(&b, " units=%s", p.Quantity)
			}
			b.WriteString("\n")
		}
	}
	for _, a := range m.Assertions {
		fmt.Fprintf(&b, "assert %s %s %s %s\n", a.Account, a.Date, a.Book, a.Balance)
	}
	for _, p := range m.Problems {
		fmt.Fprintf(&b, "problem %s\n", p)
	}
	return b.String()
}
