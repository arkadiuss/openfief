package xtb_test

import (
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"github.com/arkadiuss/openfief/internal/ingest"
	"github.com/arkadiuss/openfief/internal/ingest/xtb"
	"github.com/arkadiuss/openfief/internal/ingest/xtb/xtbtest"
)

// updateFlag rewrites the golden files:
// go test ./internal/ingest/xtb -run TestMapGolden -update.
const updateFlag = "update"

func TestMain(m *testing.M) {
	flag.Bool(updateFlag, false, "rewrite golden files")
	os.Exit(m.Run())
}

// The made-up comments of the fixture's deposit and withdrawal: they name a
// sending bank and the XTB account, as R-24 says real ones do.
const (
	depositComment    = "Przelew z Bank Testowy S.A. 11 2222 3333"
	withdrawalComment = "Wyplata na rachunek 0000000 w Bank Testowy"
)

// at returns a UTC date-time in 2025 or 2026.
func at(year int, month time.Month, day, hour, minute int) time.Time {
	return time.Date(year, month, day, hour, minute, 0, 0, time.UTC)
}

// fixture is a synthetic export of account 0000000 in PLN with every kind
// of row R-24 lists, CFD rows, and amounts with float artifacts.
func fixture() *xtbtest.Export {
	return &xtbtest.Export{
		Account: xtbtest.Account, Currency: "PLN", From: at(2025, 11, 1, 0, 0), To: at(2026, 6, 30, 0, 0),
		Cash: [][]xtbtest.Cell{
			xtbtest.CashRow("Stock purchase", "Test CFD", "TESTCFD", "CFD", at(2026, 6, 2, 9, 0), "-100", "9000000011", "OPEN BUY 1/1 @ 100"),
			xtbtest.CashRow("Sec fee", "", "", "", at(2026, 6, 1, 9, 0), "-0.01", "9000000010", ""),
			xtbtest.CashRow("Withdrawal", "", "", "", at(2026, 5, 1, 12, 0), "-1000", "9000000009", withdrawalComment),
			xtbtest.CashRow("Stock sell", "Test Alpha SA", "TESTA.PL", "STOCK", at(2026, 4, 1, 14, 5), "399.99000000000001", "9000000006", "CLOSE BUY 3/10 @ 133.33"),
			xtbtest.CashRow("Withholding tax", "Test Beta PLC", "TESTB.UK", "ETF", at(2026, 3, 2, 8, 0), "-1.8799999999999999", "9000000005", "TESTB.UK GBP WHT 15%"),
			xtbtest.CashRow("Dividend", "Test Beta PLC", "TESTB.UK", "ETF", at(2026, 3, 2, 8, 0), "12.5", "9000000004", "TESTB.UK GBP 0.5000/ SHR"),
			xtbtest.CashRow("Free funds interest tax", "", "", "", at(2026, 2, 1, 6, 0), "-0.23", "9000000008", "Free-funds Interest Tax 2026-01"),
			xtbtest.CashRow("Free funds interest", "", "", "", at(2026, 2, 1, 6, 0), "1.23", "9000000007", "Free-funds Interest 2026-01"),
			xtbtest.CashRow("Stock purchase", "Test Beta PLC", "TESTB.UK", "ETF", at(2025, 11, 5, 23, 30), "-500", "9000000003", "OPEN BUY 5/5 @ 20.00"),
			xtbtest.CashRow("Stock purchase", "Test Alpha SA", "TESTA.PL", "STOCK", at(2025, 11, 4, 10, 0), "-1234.5000000000002", "9000000002", "OPEN BUY 10/10 @ 123.45"),
			xtbtest.CashRow("Deposit", "", "", "", at(2025, 11, 3, 8, 0), "10000", "9000000001", depositComment),
		},
		Closed: [][]xtbtest.Cell{
			xtbtest.ClosedRow(&xtbtest.Position{
				Name: "Test Alpha SA", Ticker: "TESTA.PL", Category: "STOCK", Volume: "3", OpenPrice: "123.45", Opened: at(2025, 11, 4, 10, 0),
				ClosePrice: "133.33", Closed: at(2026, 4, 1, 14, 5), Profit: "29.64", PositionID: "9000000101",
			}),
			xtbtest.ClosedRow(&xtbtest.Position{
				Name: "Test Index", Ticker: "TESTCFD", Category: "IND", Volume: "1", OpenPrice: "100", Opened: at(2026, 1, 5, 10, 0),
				ClosePrice: "90", Closed: at(2026, 1, 6, 10, 0), Profit: "-10", PositionID: "9000000102",
			}),
		},
		Open: [][]xtbtest.Cell{
			xtbtest.OpenRow(&xtbtest.Position{Name: "Test Alpha SA", Ticker: "TESTA.PL", Category: "STOCK", Volume: "7", OpenPrice: "123.45", Opened: at(2025, 11, 4, 10, 0)}),
			xtbtest.OpenRow(&xtbtest.Position{Name: "Test Gold", Ticker: "TESTCMD", Category: "CMD", Volume: "0.1", OpenPrice: "2000", Opened: at(2026, 6, 1, 10, 0)}),
		},
	}
}

// mapExport maps e in transactions mode under its xStation file name.
func mapExport(t *testing.T, e *xtbtest.Export) ingest.Mapping {
	t.Helper()
	m, err := xtb.Mapper{}.Map(e.Bytes(), e.FileName(), ingest.ModeTransactions)
	if err != nil {
		t.Fatalf("Map: %v", err)
	}
	return m
}

func TestMapGolden(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		export func() *xtbtest.Export
	}{
		{name: "export.transactions.golden", export: fixture},
		{name: "inline.transactions.golden", export: func() *xtbtest.Export {
			e := fixture()
			e.Inline, e.NoTotals = true, true
			return e
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			m := mapExport(t, tt.export())
			checkGolden(t, tt.name, render(&m))
		})
	}
}

// checkGolden compares got with testdata/name, or rewrites it with -update.
func checkGolden(t *testing.T, name, got string) {
	t.Helper()
	if f := flag.Lookup(updateFlag); f != nil && f.Value.String() == "true" {
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

// render writes a mapping as stable text.
func render(m *ingest.Mapping) string {
	var b strings.Builder
	fmt.Fprintf(&b, "broker accounts %v\n", m.BrokerAccounts)
	for _, tk := range m.Tickers {
		fmt.Fprintf(&b, "ticker %s code=%s country=%s name=%q\n", tk.Symbol, tk.Code, tk.Country, tk.Name)
	}
	for i := range m.ByBrokerAccount {
		e := &m.ByBrokerAccount[i]
		for j := range e.Transactions {
			tx := &e.Transactions[j]
			fmt.Fprintf(&b, "%s tx %s %s %s account=%q %q\n", e.Account, tx.Key, tx.Kind, tx.Date, tx.Account, tx.Description)
			for _, p := range tx.Postings {
				fmt.Fprintf(&b, "  %s %s\n", p.Book, p.Amount)
			}
		}
		for j := range e.TickerEntries {
			te := &e.TickerEntries[j]
			fmt.Fprintf(&b, "%s ticker-entry %s %s %s %s units=%s amount=%s charge=%s %q ref=%s\n",
				e.Account, te.Header.Key, te.Kind, te.Header.Date, te.Symbol, te.Quantity, te.Amount, te.Charge, te.Description, te.Ref)
		}
	}
	for _, p := range m.Problems {
		fmt.Fprintf(&b, "problem %s\n", p)
	}
	return b.String()
}

// TestMapNeverKeepsTransferComments checks that nothing of a deposit's or
// withdrawal's comment, which name the sending bank and the XTB account,
// reaches the mapping, and that the account number is only the broker
// account the mapping is named by.
func TestMapNeverKeepsTransferComments(t *testing.T) {
	t.Parallel()
	m := mapExport(t, fixture())
	all := fmt.Sprintf("%+v", m)
	for _, secret := range []string{"Bank Testowy", "Przelew", "Wyplata", "11 2222 3333"} {
		if strings.Contains(all, secret) {
			t.Errorf("the mapping holds %q from a transfer's comment", secret)
		}
	}
	m.BrokerAccounts = nil
	for i := range m.ByBrokerAccount {
		m.ByBrokerAccount[i].Account = ""
	}
	if s := fmt.Sprintf("%+v", m); regexp.MustCompile(`(^|[^0-9])` + regexp.QuoteMeta(xtbtest.Account) + `([^0-9]|$)`).MatchString(s) {
		t.Errorf("the account number appears outside the broker account: %s", s)
	}
}

func TestMapRefuses(t *testing.T) {
	t.Parallel()
	good := fixture()
	other := fixture()
	other.Account = "0000001"
	tests := []struct {
		name     string
		data     []byte
		fileName string
		mode     ingest.Mode
		want     error
	}{
		{name: "opening mode", data: good.Bytes(), fileName: good.FileName(), mode: ingest.ModeOpening, want: xtb.ErrOpeningMode},
		{name: "unknown mode", data: good.Bytes(), fileName: good.FileName(), mode: ingest.Mode(0), want: xtb.ErrUnknownMode},
		{name: "not a zip", data: []byte("Type,Amount\n"), fileName: "x.xlsx", mode: ingest.ModeTransactions, want: xtb.ErrNotWorkbook},
		{name: "too large", data: make([]byte, xtb.MaxFileBytes+1), fileName: "x.xlsx", mode: ingest.ModeTransactions, want: xtb.ErrTooLarge},
		{name: "no cash sheet", data: xtbtest.Workbook(xtbtest.Sheet{Name: "Sheet1"}), fileName: "x.xlsx", mode: ingest.ModeTransactions, want: xtb.ErrLayout},
		{
			name: "file name of another account", data: other.Bytes(), fileName: good.FileName(),
			mode: ingest.ModeTransactions, want: xtb.ErrLayout,
		},
		{
			name: "two currencies", data: good.Bytes(), fileName: strings.Replace(good.FileName(), "PLN", "EUR", 1),
			mode: ingest.ModeTransactions, want: xtb.ErrCurrency,
		},
		{
			name: "unknown currency", data: noSummary(), fileName: strings.Replace(good.FileName(), "PLN", "XAU", 1),
			mode: ingest.ModeTransactions, want: xtb.ErrCurrency,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := xtb.Mapper{}.Map(tt.data, tt.fileName, tt.mode)
			if !errors.Is(err, tt.want) {
				t.Errorf("Map() error = %v, want %v", err, tt.want)
			}
		})
	}
}

// noSummary is an export with only its Cash Operations sheet, so only the
// file name can name the account's currency.
func noSummary() []byte {
	e := fixture()
	sheet := xtbtest.Sheet{Name: "Cash Operations", Rows: [][]xtbtest.Cell{
		{xtbtest.Text("Account number"), xtbtest.Text(e.Account)},
		{},
		{},
		{},
		{xtbtest.Text("Type"), xtbtest.Text("Ticker"), xtbtest.Text("Time"), xtbtest.Text("Amount"), xtbtest.Text("ID")},
	}}
	return xtbtest.Workbook(sheet)
}

// TestMapCurrencyFromSummary checks that an export renamed by the owner
// still maps, its currency taken from Open Positions' summary.
func TestMapCurrencyFromSummary(t *testing.T) {
	t.Parallel()
	e := fixture()
	m, err := xtb.Mapper{}.Map(e.Bytes(), "renamed.xlsx", ingest.ModeTransactions)
	if err != nil {
		t.Fatalf("Map: %v", err)
	}
	want := mapExport(t, e)
	if diff := cmp.Diff(render(&want), render(&m)); diff != "" {
		t.Errorf("renamed export maps otherwise (-want +got):\n%s", diff)
	}
}

// TestMapAmounts checks rounding to the minor unit: a float artifact is
// rounded away, a real fraction of a grosz is refused as an ingest problem.
func TestMapAmounts(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		amount string
		want   string
	}{
		{name: "artifact above", amount: "250.00000000000003", want: "250 PLN"},
		{name: "artifact below", amount: "249.99999999999997", want: "250 PLN"},
		{name: "exponent", amount: "2.5E2", want: "250 PLN"},
		{name: "fraction of a grosz", amount: "250.005", want: "problem"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			e := fixture()
			e.Cash = [][]xtbtest.Cell{xtbtest.CashRow("Deposit", "", "", "", at(2026, 1, 2, 10, 0), tt.amount, "9000000001", "")}
			m := mapExport(t, e)
			got := "problem"
			if len(m.ByBrokerAccount) == 1 && len(m.ByBrokerAccount[0].Transactions) == 1 {
				got = m.ByBrokerAccount[0].Transactions[0].Postings[0].Amount.String()
			}
			if got != tt.want {
				t.Errorf("deposit of %s = %s, want %s", tt.amount, got, tt.want)
			}
		})
	}
}
