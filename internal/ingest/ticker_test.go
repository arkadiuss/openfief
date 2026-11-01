package ingest_test

import (
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/arkadiuss/openfief/internal/ingest"
	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
)

// Synthetic ISINs with valid check digits.
const (
	isinA = "PLTESTA00002"
)

// pln returns an amount in PLN.
func pln(t *testing.T, s string) kernel.Money {
	t.Helper()
	c, err := kernel.ParseCurrency("PLN")
	if err != nil {
		t.Fatal(err)
	}
	m, err := kernel.ParseMoney(s, c)
	if err != nil {
		t.Fatal(err)
	}
	return m
}

// isin parses a synthetic ISIN.
func isin(t *testing.T, s string) ledger.ISIN {
	t.Helper()
	i, err := ledger.ParseISIN(s)
	if err != nil {
		t.Fatal(err)
	}
	return i
}

// entry returns a ticker entry of kind for symbol in account acc-1, with
// external ID id.
func entry(t *testing.T, symbol, id string, kind ledger.TransactionKind) ingest.TickerEntry {
	t.Helper()
	day, err := kernel.ParseDate("2026-01-05")
	if err != nil {
		t.Fatal(err)
	}
	e := ingest.TickerEntry{
		Symbol: symbol, Kind: kind, Amount: pln(t, "400"), Ref: "Cash Operations#7[ID=" + id + "]",
		Header: ledger.TransactionHeader{Key: ledger.TransactionKey{Source: "xtb:acc-1", ExternalID: id}, Account: "acc-1", Date: day},
	}
	if kind != ledger.KindDividend {
		e.Quantity = kernel.NewQuantity(kernel.NewDecimal(4, 0))
	}
	return e
}

// entryOutcome is what a ticker entry's transaction is compared by.
type entryOutcome struct {
	Kind       ledger.TransactionKind
	Key        ledger.TransactionKey
	Account    ledger.AccountID
	Instrument string
}

// outcomeOf returns what tx is compared by; none for no transaction.
func outcomeOf(tx *ledger.Transaction) entryOutcome {
	if len(tx.Postings) == 0 {
		return entryOutcome{}
	}
	return entryOutcome{Kind: tx.Kind, Key: tx.Key, Account: tx.Account, Instrument: tx.Postings[0].Book.Instrument.String()}
}

// wantOutcome is the outcome of an entry of kind keyed key, booked in
// acc-1 with TESTA.PL's synthetic ISIN; none for no kind.
func wantOutcome(kind ledger.TransactionKind, key ledger.TransactionKey) entryOutcome {
	if kind == 0 {
		return entryOutcome{}
	}
	return entryOutcome{Kind: kind, Key: key, Account: "acc-1", Instrument: isinA + "/TESTA.PL"}
}

// entryCase returns a ticker entry of kind for TESTA.PL, with one unit when
// quantity, a lot open date when opened isn't empty and a charge of 1 PLN
// when charge.
func entryCase(t *testing.T, kind ledger.TransactionKind, quantity bool, opened string, charge bool) ingest.TickerEntry {
	t.Helper()
	e := entry(t, "TESTA.PL", "9000000002", kind)
	if quantity {
		e.Quantity = kernel.NewQuantity(kernel.NewDecimal(1, 0))
	}
	if opened != "" {
		day, err := kernel.ParseDate(opened)
		if err != nil {
			t.Fatal(err)
		}
		e.Opened = day
	}
	if charge {
		e.Charge = pln(t, "1")
	}
	return e
}

func TestTickerEntryTransaction(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		kind     ledger.TransactionKind
		quantity bool
		opened   string
		charge   bool
		wantKind ledger.TransactionKind
		wantErr  error
	}{
		{name: "buy", kind: ledger.KindBuy, wantKind: ledger.KindBuy},
		{name: "opening lot of a position held since its open date", kind: ledger.KindOpeningLot, opened: "2025-06-02", wantKind: ledger.KindOpeningLot},
		{name: "opening lot without an open date", kind: ledger.KindOpeningLot, wantErr: ingest.ErrInvalidTickerEntry},
		{name: "opening lot opened after its go-live", kind: ledger.KindOpeningLot, opened: "2026-01-06", wantErr: ingest.ErrInvalidTickerEntry},
		{name: "opening lot with a charge", kind: ledger.KindOpeningLot, opened: "2025-06-02", charge: true, wantErr: ingest.ErrInvalidTickerEntry},
		{name: "buy with a lot open date", kind: ledger.KindBuy, opened: "2025-06-02", wantErr: ingest.ErrInvalidTickerEntry},
		{name: "sell", kind: ledger.KindSell, wantKind: ledger.KindSell},
		{name: "dividend", kind: ledger.KindDividend, wantKind: ledger.KindDividend},
		{name: "dividend with units", kind: ledger.KindDividend, quantity: true, wantErr: ingest.ErrInvalidTickerEntry},
		{name: "deposit", kind: ledger.KindDeposit, wantErr: ingest.ErrInvalidTickerEntry},
		{name: "no kind", kind: 0, wantErr: ingest.ErrInvalidTickerEntry},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			e := entryCase(t, tt.kind, tt.quantity, tt.opened, tt.charge)
			got, err := e.Transaction(isin(t, isinA))
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("Transaction() error = %v, want %v", err, tt.wantErr)
			}
			if diff := cmp.Diff(wantOutcome(tt.wantKind, e.Header.Key), outcomeOf(&got)); diff != "" {
				t.Errorf("Transaction() (-want +got):\n%s", diff)
			}
		})
	}
}

func TestBookTickerEntries(t *testing.T) {
	t.Parallel()
	buyA := entry(t, "TESTA.PL", "9000000002", ledger.KindBuy)
	buyB := entry(t, "TESTB.UK", "9000000003", ledger.KindBuy)
	m := ingest.Mapping{TickerEntries: []ingest.TickerEntry{buyA, buyB}}
	got, waiting := ingest.BookTickerEntries(&m, map[string]ledger.ISIN{"TESTA.PL": isin(t, isinA)})
	if len(got.Transactions) != 1 || got.Transactions[0].Key != buyA.Header.Key {
		t.Errorf("booked %v, want only %s", got.Transactions, buyA.Header.Key)
	}
	if diff := cmp.Diff([]ingest.TickerEntry{buyB}, waiting); diff != "" {
		t.Errorf("waiting (-want +got):\n%s", diff)
	}
	wantProblems := []ingest.Problem{ingest.Unsupported(buyB.Ref, "waits for the owner to confirm the ISIN of TESTB.UK")}
	if diff := cmp.Diff(wantProblems, got.Problems); diff != "" {
		t.Errorf("problems (-want +got):\n%s", diff)
	}
	if len(got.TickerEntries) != 0 || len(m.Transactions) != 0 {
		t.Error("BookTickerEntries left ticker entries in the result or changed its input")
	}
}

func TestTickerValidate(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		ticker  ingest.Ticker
		wantErr error
	}{
		{name: "xStation symbol", ticker: ingest.Ticker{Symbol: "TESTA.PL", Code: "TESTA", Country: "PL", Name: "Test Alpha SA"}},
		{name: "no country", ticker: ingest.Ticker{Symbol: "TESTX.ZZ", Code: "TESTX"}},
		{name: "no symbol", ticker: ingest.Ticker{Code: "TESTA"}, wantErr: ingest.ErrInvalidTicker},
		{name: "space in symbol", ticker: ingest.Ticker{Symbol: "TEST A"}, wantErr: ingest.ErrInvalidTicker},
		{name: "lower-case country", ticker: ingest.Ticker{Symbol: "TESTA.PL", Country: "pl"}, wantErr: ingest.ErrInvalidTicker},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if err := tt.ticker.Validate(); !errors.Is(err, tt.wantErr) {
				t.Errorf("Validate() = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

// TestResolveBrokerAccountsTickerEntries checks that ticker entries named
// by a mapped broker account get its OpenFief account, as transactions do,
// and those of an unmapped one are left out.
func TestResolveBrokerAccountsTickerEntries(t *testing.T) {
	t.Parallel()
	e := entry(t, "TESTA.PL", "9000000002", ledger.KindBuy)
	e.Header.Account, e.Header.Key.Source = "", "xtb"
	m := ingest.Mapping{ByBrokerAccount: []ingest.BrokerEntries{
		{Account: "0000000", TickerEntries: []ingest.TickerEntry{e}},
		{Account: "0000001", TickerEntries: []ingest.TickerEntry{e}},
	}}
	known := []ingest.BrokerAccount{{Instance: "xtb", ID: "0000000", Account: "acc-1"}}
	r := ingest.ResolveBrokerAccounts(&m, known, ingest.ModeTransactions)
	want := e
	want.Header.Account, want.Header.Key.Source = "acc-1", "xtb:acc-1"
	if diff := cmp.Diff([]ingest.TickerEntry{want}, r.Mapping.TickerEntries); diff != "" {
		t.Errorf("ticker entries (-want +got):\n%s", diff)
	}
	if diff := cmp.Diff([]ingest.BrokerAccountID{"0000001"}, r.Unmapped); diff != "" {
		t.Errorf("unmapped (-want +got):\n%s", diff)
	}
}

// TestResolveBrokerAccountsReseedingTickerEntries checks that an opening
// batch of a broker account already seeded leaves out its opening-lot
// ticker entries, as it does its seeding transactions, and keeps the
// others (D-136).
func TestResolveBrokerAccountsReseedingTickerEntries(t *testing.T) {
	t.Parallel()
	buy := entry(t, "TESTA.PL", "9000000002", ledger.KindBuy)
	lot := entry(t, "TESTA.PL", "open-position:1", ledger.KindOpeningLot)
	for _, e := range []*ingest.TickerEntry{&buy, &lot} {
		e.Header.Account, e.Header.Key.Source = "", "xtb"
	}
	m := ingest.Mapping{ByBrokerAccount: []ingest.BrokerEntries{{Account: "0000000", TickerEntries: []ingest.TickerEntry{buy, lot}}}}
	tests := []struct {
		name   string
		seeded bool
		want   []string
	}{
		{name: "not seeded yet", want: []string{"xtb:acc-1/9000000002", "xtb:acc-1/open-position:1"}},
		{name: "already seeded", seeded: true, want: []string{"xtb:acc-1/9000000002"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			known := []ingest.BrokerAccount{{Instance: "xtb", ID: "0000000", Account: "acc-1", Seeded: tt.seeded}}
			r := ingest.ResolveBrokerAccounts(&m, known, ingest.ModeOpening)
			var got []string
			for _, e := range r.Mapping.TickerEntries {
				got = append(got, e.Header.Key.String())
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("ticker entries (-want +got):\n%s", diff)
			}
		})
	}
}
