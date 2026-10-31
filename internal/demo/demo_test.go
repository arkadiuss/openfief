package demo_test

import (
	"bytes"
	"errors"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"github.com/arkadiuss/openfief/internal/demo"
	"github.com/arkadiuss/openfief/internal/ingest"
	"github.com/arkadiuss/openfief/internal/ingest/ibkrflex"
	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
	"github.com/arkadiuss/openfief/internal/valuation"
)

// testSeed and testToday fix the dataset most tests look at.
const (
	testSeed  = 20261028
	testToday = "2026-10-28"
)

func date(t *testing.T, s string) kernel.Date {
	t.Helper()
	d, err := kernel.ParseDate(s)
	if err != nil {
		t.Fatalf("ParseDate(%q): %v", s, err)
	}
	return d
}

func generate(t *testing.T, seed uint64, today string) demo.Dataset {
	t.Helper()
	ds, err := demo.Generate(seed, date(t, today))
	if err != nil {
		t.Fatalf("Generate(%d, %s): %v", seed, today, err)
	}
	return ds
}

func TestGenerateIsDeterministic(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		seedA    uint64
		seedB    uint64
		todayA   string
		todayB   string
		wantSame bool
	}{
		{name: "same seed and date give the same dataset", seedA: testSeed, seedB: testSeed, todayA: testToday, todayB: testToday, wantSame: true},
		{name: "another seed gives other prices", seedA: testSeed, seedB: testSeed + 1, todayA: testToday, todayB: testToday},
		{name: "another date gives another history", seedA: testSeed, seedB: testSeed, todayA: testToday, todayB: "2026-10-29"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			a, b := generate(t, tt.seedA, tt.todayA), generate(t, tt.seedB, tt.todayB)
			diff := cmp.Diff(a, b, cmp.AllowUnexported(demo.Instrument{}))
			if tt.wantSame && diff != "" {
				t.Errorf("datasets differ (-first +second):\n%s", diff)
			}
			if !tt.wantSame && cmp.Equal(a.Prices, b.Prices) {
				t.Errorf("prices are equal, want them to differ")
			}
		})
	}
}

func TestGenerateRefusesADateBeforeTheBonds(t *testing.T) {
	t.Parallel()
	_, err := demo.Generate(testSeed, date(t, "2025-12-31"))
	if !errors.Is(err, demo.ErrTooEarly) {
		t.Errorf("Generate before the last bond purchase + a year: err = %v, want ErrTooEarly", err)
	}
}

func TestSyntheticISIN(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		country string
		number  int
		want    string
		wantErr bool
	}{
		{name: "check digit makes it valid", country: "IE", number: 1, want: "IETESTD00014"},
		{name: "Polish bond", country: "PL", number: 9001, want: "PLTESTD90015"},
		{name: "number too large", country: "IE", number: 10_000, wantErr: true},
		{name: "negative number", country: "IE", number: -1, wantErr: true},
		{name: "bad country", country: "IRL", number: 1, wantErr: true},
		{name: "lower-case country has no valid ISIN", country: "ie", number: 1, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := demo.SyntheticISIN(tt.country, tt.number)
			if tt.wantErr {
				if !errors.Is(err, demo.ErrNotSynthetic) {
					t.Errorf("err = %v, want ErrNotSynthetic", err)
				}
				return
			}
			if err != nil || got.String() != tt.want || !demo.IsSynthetic(got) {
				t.Errorf("SyntheticISIN = %s, %v, synthetic %v; want %s", got, err, demo.IsSynthetic(got), tt.want)
			}
		})
	}
}

func TestIsSyntheticRefusesRealISINs(t *testing.T) {
	t.Parallel()
	// A well-known public ISIN shape (Apple's): valid, but not a demo one.
	public, err := ledger.ParseISIN("US0378331005")
	if err != nil {
		t.Fatalf("ParseISIN: %v", err)
	}
	if demo.IsSynthetic(public) || demo.IsSynthetic(ledger.ISIN{}) {
		t.Errorf("IsSynthetic accepts a real ISIN or the zero ISIN")
	}
}

// isinAttr finds the ISINs in a Flex statement.
var isinAttr = regexp.MustCompile(`isin="([^"]+)"`)

func TestDatasetIsSyntheticOnly(t *testing.T) {
	t.Parallel()
	ds := generate(t, testSeed, testToday)
	for _, s := range datasetISINs(t, &ds) {
		if s == "" {
			continue // an FX trade has no ISIN
		}
		isin, err := ledger.ParseISIN(s)
		if err != nil || !demo.IsSynthetic(isin) {
			t.Errorf("ISIN %q: valid %v, synthetic %v; want a valid synthetic ISIN", s, err == nil, demo.IsSynthetic(isin))
		}
	}
	for i := range ds.Accounts {
		if !strings.HasPrefix(string(ds.Accounts[i].ID), "demo-") {
			t.Errorf("account %s is not named as a demo account", ds.Accounts[i].ID)
		}
	}
}

// datasetISINs returns every ISIN in ds: its instruments', its bonds', and
// those in its statements and ledger transactions.
func datasetISINs(t *testing.T, ds *demo.Dataset) []string {
	t.Helper()
	isins := make([]string, 0, len(ds.Instruments)+len(ds.Bonds))
	for i := range ds.Instruments {
		isins = append(isins, ds.Instruments[i].Instrument.ISIN.String())
	}
	for i := range ds.Bonds {
		isins = append(isins, ds.Bonds[i].Instrument.ISIN.String())
	}
	for _, st := range append([][]byte{ds.IBKROpening}, ds.IBKRActivity...) {
		isins = append(isins, statementISINs(t, st)...)
	}
	for _, b := range []demo.Batch{ds.XTBOpening, ds.XTBActivity, ds.BondsOpening, ds.BondsActivity} {
		isins = append(isins, batchISINs(&b)...)
	}
	return isins
}

// accountAttr finds the IBKR account numbers in a Flex statement.
var accountAttr = regexp.MustCompile(`accountId="([^"]+)"`)

// statementISINs returns the ISINs in a Flex statement, after checking that
// it names no IBKR account but the placeholder.
func statementISINs(t *testing.T, st []byte) []string {
	t.Helper()
	accounts := accountAttr.FindAllSubmatch(st, -1)
	if len(accounts) == 0 {
		t.Errorf("statement without an accountId")
	}
	for _, a := range accounts {
		if string(a[1]) != demo.IBKRFlexAccount {
			t.Errorf("statement names IBKR account %q, want only %s", a[1], demo.IBKRFlexAccount)
		}
	}
	found := isinAttr.FindAllSubmatch(st, -1)
	out := make([]string, 0, len(found))
	for _, m := range found {
		out = append(out, string(m[1]))
	}
	return out
}

// batchISINs returns the ISINs of every holding posting in b.
func batchISINs(b *demo.Batch) []string {
	var out []string
	for i := range b.Transactions {
		for _, p := range b.Transactions[i].Postings {
			if p.Book.Kind == ledger.BookHolding {
				out = append(out, p.Book.Instrument.ISIN.String())
			}
		}
	}
	return out
}

func TestCPIReadingsArePublishedGUSValues(t *testing.T) {
	t.Parallel()
	got := map[string]string{}
	for _, r := range demo.CPIReadings() {
		got[time.Date(r.Year, r.Month, 1, 0, 0, 0, 0, time.UTC).Format("2006-01")] = r.Value.String()
	}
	// The readings the bond interest tables confirm (internal/jurisdiction/pl/official_test.go).
	want := map[string]string{"2024-09": "0.049", "2025-09": "0.029", "2024-11": "0.047", "2023-01": "0.166", "2026-09": "0.041"}
	for month, v := range want {
		if got[month] != v {
			t.Errorf("CPI %s = %q, want %q", month, got[month], v)
		}
	}
	if len(got) != 45 {
		t.Errorf("%d readings, want 45 (January 2023 to September 2026)", len(got))
	}
}

func TestDecideLoad(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		state   demo.DatabaseState
		reset   bool
		want    demo.LoadAction
		wantErr error
	}{
		{name: "empty database is loaded fresh", state: demo.DatabaseState{}, want: demo.LoadFresh},
		{name: "empty database with reset is loaded fresh", state: demo.DatabaseState{}, reset: true, want: demo.LoadFresh},
		{name: "demo database is loaded again", state: demo.DatabaseState{Marked: true, Rows: 500}, want: demo.LoadAgain},
		{name: "demo database with reset is reset", state: demo.DatabaseState{Marked: true, Rows: 500}, reset: true, want: demo.ResetAndLoad},
		{name: "unmarked database with data is refused", state: demo.DatabaseState{Rows: 1}, wantErr: demo.ErrNotDemoDatabase},
		{name: "unmarked database with data is refused even with reset", state: demo.DatabaseState{Rows: 1}, reset: true, wantErr: demo.ErrNotDemoDatabase},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := demo.DecideLoad(tt.state, tt.reset)
			if !errors.Is(err, tt.wantErr) || got != tt.want {
				t.Errorf("DecideLoad = %s, %v; want %s, %v", got, err, tt.want, tt.wantErr)
			}
		})
	}
}

func TestLoadActionString(t *testing.T) {
	t.Parallel()
	got := []string{demo.LoadFresh.String(), demo.LoadAgain.String(), demo.ResetAndLoad.String(), demo.LoadAction(0).String()}
	want := []string{"load-fresh", "load-again", "reset-and-load", "LoadAction(0)"}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("String (-want +got):\n%s", diff)
	}
}

// mapper is the IBKR connector's mapper for the demo account.
func mapper() ibkrflex.Mapper {
	return ibkrflex.Mapper{Accounts: map[string]ledger.AccountID{demo.IBKRFlexAccount: demo.IBKRAccount}}
}

// ledgerOf returns the transactions as the first version of each.
func ledgerOf(txs []ledger.Transaction) []ledger.VersionedTransaction {
	log := make([]ledger.VersionedTransaction, 0, len(txs))
	for i := range txs {
		log = append(log, ledger.VersionedTransaction{Transaction: txs[i], Version: 1})
	}
	return log
}

func TestIBKRStatementsMapWithOnlyTheIntendedProblems(t *testing.T) {
	t.Parallel()
	ds := generate(t, testSeed, testToday)
	opening, err := mapper().Map(bytes.NewReader(ds.IBKROpening), ingest.ModeOpening)
	if err != nil || len(opening.Problems) != 0 || len(opening.Transactions) == 0 {
		t.Fatalf("opening statement: err %v, %d transactions, problems %v; want transactions and no problems", err, len(opening.Transactions), opening.Problems)
	}
	if n := len(ds.IBKRActivity); n < 12 || n > 13 {
		t.Errorf("%d activity statements, want one per month of the year", n)
	}
	txs, assertions := opening.Transactions, opening.Assertions
	kinds := map[ledger.TransactionKind]int{}
	for i, st := range ds.IBKRActivity {
		m := mapActivity(t, i, st)
		for j := range m.Transactions {
			kinds[m.Transactions[j].Kind]++
		}
		txs, assertions = append(txs, m.Transactions...), append(assertions, m.Assertions...)
	}
	for _, k := range []ledger.TransactionKind{ledger.KindBuy, ledger.KindSell, ledger.KindDividend, ledger.KindFee, ledger.KindFXConversion, ledger.KindDeposit} {
		if kinds[k] == 0 {
			t.Errorf("no %s in the IBKR history", k)
		}
	}
	checkAssertions(t, "IBKR", txs, assertions)
}

func TestDirectAccountsBalance(t *testing.T) {
	t.Parallel()
	ds := generate(t, testSeed, testToday)
	tests := []struct {
		name              string
		opening, activity demo.Batch
	}{
		{name: "XTB", opening: ds.XTBOpening, activity: ds.XTBActivity},
		{name: "treasury bonds", opening: ds.BondsOpening, activity: ds.BondsActivity},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			for i := range tt.opening.Transactions {
				if !ledger.IsSeeding(&tt.opening.Transactions[i]) {
					t.Errorf("opening transaction %s does not seed", tt.opening.Transactions[i].Key)
				}
			}
			for i := range tt.activity.Transactions {
				if ledger.IsSeeding(&tt.activity.Transactions[i]) {
					t.Errorf("activity transaction %s seeds", tt.activity.Transactions[i].Key)
				}
			}
			txs := append(append([]ledger.Transaction(nil), tt.opening.Transactions...), tt.activity.Transactions...)
			checkAssertions(t, tt.name, txs, append(tt.opening.Assertions, tt.activity.Assertions...))
		})
	}
}

// mapActivity maps activity statement i, whose only problems must be the
// intended skipped summary rows.
func mapActivity(t *testing.T, i int, st []byte) ingest.Mapping {
	t.Helper()
	m, err := mapper().Map(bytes.NewReader(st), ingest.ModeTransactions)
	if err != nil {
		t.Fatalf("activity statement %d: %v", i, err)
	}
	for _, p := range m.Problems {
		if p.Kind != ingest.ProblemSkipped || !strings.Contains(p.Reason, `"SUMMARY"`) {
			t.Errorf("activity statement %d: unintended problem %s", i, p)
		}
	}
	return m
}

// checkAssertions checks every assertion against txs.
func checkAssertions(t *testing.T, name string, txs []ledger.Transaction, assertions []ledger.BalanceAssertion) {
	t.Helper()
	if len(assertions) == 0 {
		t.Errorf("%s: no balance assertions", name)
	}
	log := ledgerOf(txs)
	for i := range assertions {
		if err := ledger.CheckBalanceAssertion(log, &assertions[i]); err != nil {
			t.Errorf("%s: %v", name, err)
		}
	}
}

func TestBondCouponsFollowThePublishedRates(t *testing.T) {
	t.Parallel()
	ds := generate(t, testSeed, testToday)
	var got []string
	for i := range ds.BondsActivity.Transactions {
		tx := &ds.BondsActivity.Transactions[i]
		for _, p := range tx.Postings {
			if p.Book.Kind == ledger.BookCash {
				got = append(got, tx.Kind.String()+" "+tx.Date.String()+" "+p.Amount.String())
			}
		}
	}
	// COI0128 bought 2024-01-10: year 2 at November 2024 CPI 4.7% + margin 1.25% = 5.95 PLN
	// a bond, paid 2026-01-10 on 30 bonds. Year 1's coupon came before go-live.
	// It is interest, not a dividend (D-101).
	want := []string{"interest 2026-01-10 178.5 PLN"}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("coupons (-want +got):\n%s", diff)
	}
}

func TestPricesCoverTheHistoryOnBusinessDays(t *testing.T) {
	t.Parallel()
	ds := generate(t, testSeed, testToday)
	first := map[string]kernel.Date{}
	for _, p := range ds.Prices {
		if p.Close.Sign() <= 0 || p.Source != demo.PriceSource || isWeekend(p.Date) {
			t.Errorf("price %s on %s: %s from %s; want a positive demo close on a weekday", p.Instrument, p.Date, p.Close, p.Source)
		}
		if d, ok := first[p.Instrument]; !ok || p.Date.Before(d) {
			first[p.Instrument] = p.Date
		}
	}
	if len(first) != len(ds.Instruments) {
		t.Errorf("prices for %d instruments, want %d", len(first), len(ds.Instruments))
	}
	checkPricedAtGoLive(t, &ds, first)
}

// checkPricedAtGoLive checks that every instrument's first price, in first,
// is on or before go-live.
func checkPricedAtGoLive(t *testing.T, ds *demo.Dataset, first map[string]kernel.Date) {
	t.Helper()
	for in, d := range first {
		if _, err := valuation.PriceOnOrBefore(ds.Prices, in, ds.GoLive); err != nil || d.After(ds.GoLive) {
			t.Errorf("%s: first price %s, want one on or before go-live %s", in, d, ds.GoLive)
		}
	}
}

// isWeekend reports whether d is a Saturday or a Sunday.
func isWeekend(d kernel.Date) bool {
	wd := time.Date(d.Year(), d.Month(), d.Day(), 0, 0, 0, 0, time.UTC).Weekday()
	return wd == time.Saturday || wd == time.Sunday
}

// TestFXFromCoversEveryForeignLot checks that the dataset asks for FX rates
// from the open date of its earliest foreign-currency lot: a lot's cost
// converts at the tax rate of that date (D-20), so a rate source starting
// later leaves every net and liquidation value of the account missing.
func TestFXFromCoversEveryForeignLot(t *testing.T) {
	t.Parallel()
	ds := generate(t, testSeed, testToday)
	opening, err := mapper().Map(bytes.NewReader(ds.IBKROpening), ingest.ModeOpening)
	if err != nil {
		t.Fatalf("opening statement: %v", err)
	}
	if ds.FXFrom.After(ds.GoLive) {
		t.Errorf("FXFrom %s is after go-live %s", ds.FXFrom, ds.GoLive)
	}
	earliest := ds.GoLive
	for i := range opening.Transactions {
		tx := &opening.Transactions[i]
		if tx.Kind != ledger.KindOpeningLot {
			continue
		}
		if tx.LotOpened.Before(ds.FXFrom) {
			t.Errorf("%s: lot opened on %s, before FXFrom %s", tx.Key, tx.LotOpened, ds.FXFrom)
		}
		if tx.LotOpened.Before(earliest) {
			earliest = tx.LotOpened
		}
	}
	if !earliest.Equal(ds.FXFrom) {
		t.Errorf("FXFrom = %s, want the earliest lot's open date %s", ds.FXFrom, earliest)
	}
	if !earliest.Before(ds.GoLive) {
		t.Errorf("no IBKR lot opened before go-live: the test no longer covers the gap")
	}
}
