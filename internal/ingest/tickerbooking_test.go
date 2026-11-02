package ingest_test

import (
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/arkadiuss/openfief/internal/ingest"
	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
)

// tbDec parses a synthetic decimal.
func tbDec(t *testing.T, s string) kernel.Decimal {
	t.Helper()
	d, err := kernel.ParseDecimal(s)
	if err != nil {
		t.Fatalf("ParseDecimal(%q): %v", s, err)
	}
	return d
}

// holdingOf is account acc-1's ticker assertion of units of symbol at the
// end of 2026-06-30.
func holdingOf(t *testing.T, symbol, units string) ingest.TickerAssertion {
	t.Helper()
	return ingest.TickerAssertion{
		Symbol: symbol, Account: "acc-1", Date: goLiveDate(t, "2026-06-30"), Quantity: kernel.NewQuantity(tbDec(t, units)),
		Ref: "Open Positions#10",
	}
}

// listingOf is the listing of isin in PLN.
func listingOf(t *testing.T, code string) ingest.Listing {
	t.Helper()
	return ingest.Listing{ISIN: isin(t, code), Currency: currency(t, "PLN")}
}

func TestTickerAssertion(t *testing.T) {
	t.Parallel()
	a := holdingOf(t, "TESTA.PL", "7")
	in := ledger.Instrument{ISIN: isin(t, "PLTESTA00002"), Symbol: "TESTA.PL"}
	want := ledger.HoldingAssertion("acc-1", goLiveDate(t, "2026-06-30"), in, kernel.NewQuantity(tbDec(t, "7")))
	if diff := cmp.Diff(want, a.Assertion(listingOf(t, "PLTESTA00002"))); diff != "" {
		t.Errorf("Assertion (-want +got):\n%s", diff)
	}
	tests := []struct {
		name string
		edit func(*ingest.TickerAssertion)
		want error
	}{
		{name: "valid", edit: func(*ingest.TickerAssertion) {}},
		{name: "none held is asserted too", edit: func(a *ingest.TickerAssertion) { a.Quantity = kernel.Quantity{} }},
		{name: "no symbol", edit: func(a *ingest.TickerAssertion) { a.Symbol = "" }, want: ingest.ErrInvalidTickerAssertion},
		{name: "no date", edit: func(a *ingest.TickerAssertion) { a.Date = kernel.Date{} }, want: ingest.ErrInvalidTickerAssertion},
		{name: "a negative quantity", edit: func(a *ingest.TickerAssertion) { a.Quantity = a.Quantity.Neg() }, want: ingest.ErrInvalidTickerAssertion},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			a := holdingOf(t, "TESTA.PL", "7")
			tt.edit(&a)
			if err := a.Validate(); !errors.Is(err, tt.want) {
				t.Errorf("Validate() = %v, want %v", err, tt.want)
			}
		})
	}
}

// TestBookTickerAssertions checks D-155's split: a holding of a confirmed
// ticker becomes a balance assertion of its instrument, one of a ticker
// not confirmed yet waits.
func TestBookTickerAssertions(t *testing.T) {
	t.Parallel()
	m := ingest.Mapping{TickerAssertions: []ingest.TickerAssertion{holdingOf(t, "TESTA.PL", "7"), holdingOf(t, "TESTB.UK", "20")}}
	confirmed := map[string]ingest.Listing{"TESTA.PL": listingOf(t, "PLTESTA00002")}
	got, waiting := ingest.BookTickerAssertions(&m, confirmed)
	want := []ledger.BalanceAssertion{m.TickerAssertions[0].Assertion(confirmed["TESTA.PL"])}
	if diff := cmp.Diff(want, got.Assertions); diff != "" {
		t.Errorf("assertions (-want +got):\n%s", diff)
	}
	if diff := cmp.Diff([]ingest.TickerAssertion{m.TickerAssertions[1]}, waiting); diff != "" {
		t.Errorf("waiting (-want +got):\n%s", diff)
	}
	if got.TickerAssertions != nil {
		t.Errorf("the booked mapping keeps ticker assertions %v", got.TickerAssertions)
	}
}

// buyOf is account acc-1's buy of 4 units of symbol for 400 PLN, keyed id.
func buyOf(t *testing.T, symbol, id string) ingest.TickerEntry {
	t.Helper()
	return ingest.TickerEntry{
		Symbol: symbol, Kind: ledger.KindBuy, Quantity: kernel.NewQuantity(tbDec(t, "4")),
		Amount: kernel.NewMoney(tbDec(t, "400"), currency(t, "PLN")), Description: "buy " + symbol,
		Header: ledger.TransactionHeader{Key: ledger.TransactionKey{Source: "xtb:acc-1", ExternalID: id}, Account: "acc-1", Date: goLiveDate(t, "2026-01-05")},
	}
}

// TestBookedTickerRows checks what a delivery keeps of its booked rows for
// a correction (D-153): the entries of confirmed tickers that book, and
// the holdings of confirmed tickers.
func TestBookedTickerRows(t *testing.T) {
	t.Parallel()
	invalid := buyOf(t, "TESTA.PL", "9000000009")
	invalid.Quantity = kernel.Quantity{}
	m := ingest.Mapping{
		TickerEntries:    []ingest.TickerEntry{buyOf(t, "TESTA.PL", "9000000002"), buyOf(t, "TESTB.UK", "9000000003"), invalid},
		TickerAssertions: []ingest.TickerAssertion{holdingOf(t, "TESTA.PL", "7"), holdingOf(t, "TESTB.UK", "20")},
	}
	got := ingest.BookedTickerRows("xtb-main", &m, map[string]ingest.Listing{"TESTA.PL": listingOf(t, "PLTESTA00002")})
	want := ingest.TickerRows{
		Instance: "xtb-main", Entries: []ingest.TickerEntry{m.TickerEntries[0]}, Assertions: []ingest.TickerAssertion{m.TickerAssertions[0]},
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("BookedTickerRows (-want +got):\n%s", diff)
	}
	if diff := cmp.Diff([]ledger.AccountID{"acc-1"}, got.Accounts()); diff != "" {
		t.Errorf("Accounts (-want +got):\n%s", diff)
	}
}

// TestTickerBookingLedger checks what a confirmation and a correction book
// (D-116, D-153, D-155): each row's transaction with the listing, keyed in
// its instance, each holding asserted of the listing's instrument, and for
// a correction to another ISIN the holdings of the old instrument
// replaced, so the corrected transactions supersede the old ones by key.
func TestTickerBookingLedger(t *testing.T) {
	t.Parallel()
	rows := []ingest.TickerRows{
		{Instance: "xtb-main", Entries: []ingest.TickerEntry{buyOf(t, "TESTA.PL", "9000000002")}, Assertions: []ingest.TickerAssertion{holdingOf(t, "TESTA.PL", "4")}},
		{Instance: "xtb-ike", Entries: []ingest.TickerEntry{buyOf(t, "TESTA.PL", "9000000005")}},
	}
	old, corrected := listingOf(t, "PLTESTA00002"), listingOf(t, "PLTESTC00008")
	tests := []struct {
		name             string
		booking          ingest.TickerBooking
		wantISIN         string
		wantReplacedISIN string
	}{
		{name: "a confirmation", booking: ingest.TickerBooking{Listing: old, Rows: rows}, wantISIN: "PLTESTA00002"},
		{name: "a correction to another ISIN", booking: ingest.TickerBooking{Listing: corrected, Previous: old, Rows: rows}, wantISIN: "PLTESTC00008", wantReplacedISIN: "PLTESTA00002"},
		{name: "a correction of the currency only", booking: ingest.TickerBooking{Listing: old, Previous: old, Rows: rows}, wantISIN: "PLTESTA00002"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := tt.booking.Ledger()
			if err != nil {
				t.Fatalf("Ledger: %v", err)
			}
			wantKeys := []string{
				"plugin:xtb-main:xtb:acc-1/9000000002 holding:" + tt.wantISIN + "/TESTA.PL",
				"plugin:xtb-ike:xtb:acc-1/9000000005 holding:" + tt.wantISIN + "/TESTA.PL",
			}
			if diff := cmp.Diff(wantKeys, bookedKeys(got.Transactions)); diff != "" {
				t.Errorf("transactions (-want +got):\n%s", diff)
			}
			in := ledger.Instrument{ISIN: isin(t, tt.wantISIN), Symbol: "TESTA.PL"}
			want := []ledger.BalanceAssertion{ledger.HoldingAssertion("acc-1", goLiveDate(t, "2026-06-30"), in, kernel.NewQuantity(tbDec(t, "4")))}
			if diff := cmp.Diff(want, got.Asserted); diff != "" {
				t.Errorf("asserted (-want +got):\n%s", diff)
			}
			if diff := cmp.Diff(replacedOf(t, tt.wantReplacedISIN), got.Replaced, cmpopts.EquateEmpty()); diff != "" {
				t.Errorf("replaced (-want +got):\n%s", diff)
			}
		})
	}
	bad := ingest.TickerBooking{Listing: old, Rows: []ingest.TickerRows{{Instance: "xtb-main", Entries: []ingest.TickerEntry{{Symbol: "TESTA.PL"}}}}}
	if _, err := bad.Ledger(); !errors.Is(err, ingest.ErrInvalidTickerEntry) {
		t.Errorf("Ledger of an entry that books nothing = %v, want ErrInvalidTickerEntry", err)
	}
}

// bookedKeys names each of txs by its key and its first ledger account.
func bookedKeys(txs []ledger.Transaction) []string {
	keys := make([]string, 0, len(txs))
	for i := range txs {
		keys = append(keys, txs[i].Key.String()+" "+txs[i].Postings[0].Book.String())
	}
	return keys
}

// replacedOf is the holding of 4 units of TESTA.PL as code at the end of
// 2026-06-30 that a correction replaces, or none for "".
func replacedOf(t *testing.T, code string) []ledger.BalanceAssertion {
	t.Helper()
	if code == "" {
		return nil
	}
	was := ledger.Instrument{ISIN: isin(t, code), Symbol: "TESTA.PL"}
	return []ledger.BalanceAssertion{ledger.HoldingAssertion("acc-1", goLiveDate(t, "2026-06-30"), was, kernel.NewQuantity(tbDec(t, "4")))}
}

func TestConfirmedTickerCurrencyChoices(t *testing.T) {
	t.Parallel()
	london := ingest.ConfirmedTicker{Ticker: ingest.Ticker{Symbol: "TESTB.UK", Country: "GB"}}
	warsaw := ingest.ConfirmedTicker{Ticker: ingest.Ticker{Symbol: "TESTA.PL", Country: "PL"}}
	if diff := cmp.Diff([]string{"GBX", "GBP", "USD"}, london.CurrencyChoices()); diff != "" {
		t.Errorf("London's choices (-want +got):\n%s", diff)
	}
	if got := warsaw.CurrencyChoices(); got != nil {
		t.Errorf("Warsaw's choices = %v, want none", got)
	}
}
