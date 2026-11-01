package ingest_test

import (
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/arkadiuss/openfief/internal/ingest"
	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
)

// currency parses a currency code, failing the test on a bad one.
func currency(t *testing.T, code string) kernel.Currency {
	t.Helper()
	c, err := kernel.ParseCurrency(code)
	if err != nil {
		t.Fatal(err)
	}
	return c
}

// TestListingOf checks which trading currency a ticker's confirmation
// settles (D-159): the confirmed listing's, else the exchange's its market
// names, with a price in pence converted to pounds.
func TestListingOf(t *testing.T) {
	t.Parallel()
	confirmed, other := isin(t, isinA), isin(t, "XX0000000002")
	tests := []struct {
		name      string
		country   string
		proposed  ledger.ISIN
		currency  string
		wantCode  string
		wantMinor bool
	}{
		{name: "the confirmed listing's currency", country: "GB", proposed: confirmed, currency: "USD", wantCode: "USD"},
		{name: "a listing quoted in pence", country: "GB", proposed: confirmed, currency: "GBX", wantCode: "GBP", wantMinor: true},
		{name: "a listing that names no currency takes the exchange's", country: "PL", proposed: confirmed, wantCode: "PLN"},
		{name: "a proposal of another ISIN doesn't count", country: "DE", proposed: other, currency: "USD", wantCode: "EUR"},
		{name: "London's exchange quotes in pence", country: "GB", wantCode: "GBP", wantMinor: true},
		{name: "a market of no known exchange", country: "", wantCode: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ticker := ingest.Ticker{Symbol: "TESTA.X", Country: tt.country}
			proposal := ingest.ISINProposal{Symbol: "TESTA.X", ISIN: tt.proposed, Currency: tt.currency}
			got := ingest.ListingOf(&ticker, &proposal, confirmed)
			if got.ISIN != confirmed || got.Currency.String() != tt.wantCode || got.Minor != tt.wantMinor {
				t.Errorf("ListingOf = %+v, want %s minor=%v", got, tt.wantCode, tt.wantMinor)
			}
		})
	}
}

// TestOpeningLotTradingCurrency checks that an opening lot priced by its
// open price costs its quantity times that price in the listing's trading
// currency, rounded half to even to its minor unit, a price in pence
// converted to pounds (D-159).
func TestOpeningLotTradingCurrency(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		listing  func(t *testing.T) ingest.Listing
		price    kernel.Decimal
		want     string
		wantCode string
		wantErr  error
	}{
		{name: "in the listing's currency", listing: listingIn("EUR", false), price: kernel.NewDecimal(1234567, -4), want: "864.2", wantCode: "EUR"},
		{name: "pence converted to pounds", listing: listingIn("GBP", true), price: kernel.NewDecimal(41255, -1), want: "288.78", wantCode: "GBP"},
		{name: "a currency without a minor unit", listing: listingIn("JPY", false), price: kernel.NewDecimal(15005, -1), want: "10504", wantCode: "JPY"},
		{name: "no known trading currency", listing: listingIn("", false), price: kernel.NewDecimal(1, 0), wantErr: ingest.ErrUnknownTradingCurrency},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			e := entryCase(t, ledger.KindOpeningLot, false, "2025-06-02", false)
			e.Amount, e.OpenPrice, e.Quantity = kernel.Money{}, tt.price, kernel.NewQuantity(kernel.NewDecimal(7, 0))
			got, err := e.Transaction(tt.listing(t))
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("Transaction() error = %v, want %v", err, tt.wantErr)
			}
			if err != nil {
				return
			}
			if diff := cmp.Diff(tt.want+" "+tt.wantCode, costOf(&got)); diff != "" {
				t.Errorf("cost (-want +got):\n%s", diff)
			}
		})
	}
}

// costOf returns an opening lot's cost, its positive posting, as
// "<amount> <currency>".
func costOf(tx *ledger.Transaction) string {
	for _, p := range tx.Postings {
		if p.Amount.Sign() > 0 {
			return p.Amount.Amount().String() + " " + p.Amount.Currency().String()
		}
	}
	return ""
}

// listingIn returns a listing of isinA trading in code, its minor unit
// when minor; an empty code is none.
func listingIn(code string, minor bool) func(t *testing.T) ingest.Listing {
	return func(t *testing.T) ingest.Listing {
		t.Helper()
		l := ingest.Listing{ISIN: isin(t, isinA), Minor: minor}
		if code != "" {
			l.Currency = currency(t, code)
		}
		return l
	}
}

func TestParseTradingCurrency(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		code, want string
		minor, ok  bool
	}{
		{code: "PLN", want: "PLN", ok: true},
		{code: "GBX", want: "GBP", minor: true, ok: true},
		{code: "GBp", want: "GBP", minor: true, ok: true},
		{code: "ZAc", want: "ZAR", minor: true, ok: true},
		{code: "", ok: false},
		{code: "pln", ok: false},
	} {
		c, minor, ok := ingest.ParseTradingCurrency(tt.code)
		if c.String() != tt.want || minor != tt.minor || ok != tt.ok {
			t.Errorf("ParseTradingCurrency(%q) = %s, %v, %v; want %s, %v, %v", tt.code, c, minor, ok, tt.want, tt.minor, tt.ok)
		}
	}
}
