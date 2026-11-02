package ingest_test

import (
	"errors"
	"strings"
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
// settles (D-159, D-170): the confirmed listing's, else the only one its
// market's exchange quotes in, else the one the owner picked from that
// exchange's; never an assumed one, and a price in pence converted to
// pounds.
func TestListingOf(t *testing.T) {
	t.Parallel()
	confirmed, other := isin(t, isinA), isin(t, "XX0000000002")
	tests := []struct {
		name      string
		country   string
		proposed  ledger.ISIN
		currency  string
		chosen    string
		wantCode  string
		wantMinor bool
	}{
		{name: "the confirmed listing's currency", country: "GB", proposed: confirmed, currency: "USD", wantCode: "USD"},
		{name: "a listing quoted in pence", country: "GB", proposed: confirmed, currency: "GBX", wantCode: "GBP", wantMinor: true},
		{name: "the listing's currency wins over a pick", country: "GB", proposed: confirmed, currency: "USD", chosen: "GBX", wantCode: "USD"},
		{name: "a listing that names no currency takes the exchange's", country: "PL", proposed: confirmed, wantCode: "PLN"},
		{name: "a proposal of another ISIN doesn't count", country: "DE", proposed: other, currency: "USD", wantCode: "EUR"},
		{name: "London assumes no currency", country: "GB", wantCode: ""},
		{name: "London in the pence the owner picked", country: "GB", chosen: "GBX", wantCode: "GBP", wantMinor: true},
		{name: "London in the pounds the owner picked", country: "GB", chosen: "GBP", wantCode: "GBP"},
		{name: "London in the dollars the owner picked for another ISIN", country: "GB", proposed: other, currency: "GBX", chosen: "USD", wantCode: "USD"},
		{name: "a pick London doesn't quote in", country: "GB", chosen: "EUR", wantCode: ""},
		{name: "a pick on a one-currency exchange doesn't count", country: "PL", chosen: "EUR", wantCode: "PLN"},
		{name: "a market of no known exchange", country: "", chosen: "PLN", wantCode: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ticker := ingest.Ticker{Symbol: "TESTA.X", Country: tt.country}
			proposal := ingest.ISINProposal{Symbol: "TESTA.X", ISIN: tt.proposed, Currency: tt.currency}
			got := ingest.ListingOf(&ticker, &proposal, confirmed, tt.chosen)
			if got.ISIN != confirmed || got.Currency.String() != tt.wantCode || got.Minor != tt.wantMinor {
				t.Errorf("ListingOf = %+v, want %s minor=%v", got, tt.wantCode, tt.wantMinor)
			}
		})
	}
}

// TestConfirmListing checks that confirming a ticker whose listing names no
// trading currency, on an exchange that quotes in more than one, needs the
// owner's pick of one of them, and is refused naming them otherwise, while
// a confirmation the listing or a one-currency exchange settles takes no
// pick (D-170).
func TestConfirmListing(t *testing.T) {
	t.Parallel()
	confirmed, other := isin(t, isinA), isin(t, "XX0000000002")
	pickLondon := "pick one of GBX, GBP, USD"
	tests := []struct {
		name     string
		country  string
		proposed ledger.ISIN
		currency string
		chosen   string
		want     string
	}{
		{name: "London without a pick is refused", country: "GB", want: pickLondon},
		{name: "London with a currency it doesn't quote in is refused", country: "GB", chosen: "EUR", want: pickLondon},
		{name: "London with a lower-case pick is refused", country: "GB", chosen: "gbx", want: pickLondon},
		{name: "London with pence", country: "GB", chosen: "GBX", want: "GBP"},
		{name: "London with dollars", country: "GB", chosen: "USD", want: "USD"},
		{name: "a typed ISIN on London needs a pick", country: "GB", proposed: other, currency: "GBX", want: pickLondon},
		{name: "a listing that names its currency takes no pick", country: "GB", proposed: confirmed, currency: "GBX", want: "GBP"},
		{name: "repeating the listing's currency is no pick", country: "GB", proposed: confirmed, currency: "GBX", chosen: "GBX", want: "GBP"},
		{name: "another currency than the listing's is refused", country: "GB", proposed: confirmed, currency: "GBX", chosen: "USD", want: "nothing to pick"},
		{name: "a one-currency exchange takes no pick", country: "PL", want: "PLN"},
		{name: "a pick on a one-currency exchange is refused", country: "PL", chosen: "EUR", want: "nothing to pick"},
		{name: "a market of no known exchange takes no pick", country: "", want: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ticker := ingest.Ticker{Symbol: "TESTB.UK", Country: tt.country}
			proposal := ingest.ISINProposal{Symbol: "TESTB.UK", ISIN: tt.proposed, Currency: tt.currency}
			got, err := ingest.ConfirmListing(&ticker, &proposal, confirmed, tt.chosen)
			if got.ISIN != confirmed && err == nil {
				t.Errorf("ConfirmListing ISIN = %v, want %v", got.ISIN, confirmed)
			}
			if diff := cmp.Diff(tt.want, confirmOutcome(got, err)); diff != "" {
				t.Errorf("ConfirmListing (-want +got):\n%s", diff)
			}
		})
	}
}

// confirmOutcome describes what ConfirmListing returned: the listing's
// currency, "pick one of" the choices a refusal names in its message and
// Choices, "nothing to pick" for a refused pick, or the unexpected error.
func confirmOutcome(l ingest.Listing, err error) string {
	choice, ok := errors.AsType[*ingest.TradingCurrencyChoiceError](err)
	switch {
	case err == nil:
		return l.Currency.String()
	case !ok:
		return "unexpected: " + err.Error()
	case len(choice.Choices) == 0:
		return "nothing to pick"
	case !strings.Contains(err.Error(), strings.Join(choice.Choices, ", ")):
		return "choices not named: " + err.Error()
	}
	return "pick one of " + strings.Join(choice.Choices, ", ")
}

// TestExchangeCurrencies checks the trading currencies each exchange of a
// market country quotes in, and the choices a ticker waiting there offers
// the owner: only an exchange that quotes in more than one offers any
// (D-159, D-170).
func TestExchangeCurrencies(t *testing.T) {
	t.Parallel()
	tests := []struct {
		country     string
		want        []string
		wantChoices []string
	}{
		{country: "GB", want: []string{"GBX", "GBP", "USD"}, wantChoices: []string{"GBX", "GBP", "USD"}},
		{country: "PL", want: []string{"PLN"}},
		{country: "US", want: []string{"USD"}},
		{country: "DE", want: []string{"EUR"}},
		{country: "JP"},
		{country: "none"},
	}
	for _, tt := range tests {
		t.Run(tt.country, func(t *testing.T) {
			t.Parallel()
			if diff := cmp.Diff(tt.want, ingest.ExchangeCurrencies(tt.country)); diff != "" {
				t.Errorf("ExchangeCurrencies(%q) (-want +got):\n%s", tt.country, diff)
			}
			p := ingest.PendingTicker{Ticker: ingest.Ticker{Symbol: "TESTB.X", Country: tt.country}}
			if diff := cmp.Diff(tt.wantChoices, p.CurrencyChoices()); diff != "" {
				t.Errorf("PendingTicker.CurrencyChoices (-want +got):\n%s", diff)
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
