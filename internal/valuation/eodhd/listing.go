package eodhd

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
)

// penceCode is the code EODHD gives prices quoted in pence sterling, such as
// most shares on the LSE. It is not an ISO 4217 code.
const penceCode = "GBX"

var (
	// ErrNotListed reports an ISIN with no listing on a supported exchange,
	// or none on the exchange asked for.
	ErrNotListed = errors.New("not listed on a supported exchange")
	// ErrAmbiguous reports an ISIN listed on more than one supported
	// exchange when no exchange was asked for.
	ErrAmbiguous = errors.New("listed on more than one supported exchange")
	// ErrQuoteCurrency reports a listing currency that is neither an
	// ISO 4217 code nor pence sterling.
	ErrQuoteCurrency = errors.New("unknown quote currency")
)

// SupportedExchange reports whether code is the EODHD code of an exchange
// this adapter takes prices from, one of the venues D-51 names: the Warsaw
// Stock Exchange (WAR), the London Stock Exchange (LSE), Xetra (XETRA), and
// Euronext's markets, which EODHD codes one per city: Paris (PA),
// Amsterdam (AS), Brussels (BR), Lisbon (LS), Dublin (IR), Milan (MI) and
// Oslo (OL).
func SupportedExchange(code string) bool {
	switch code {
	case "WAR", "LSE", "XETRA", "PA", "AS", "BR", "LS", "IR", "MI", "OL":
		return true
	}
	return false
}

// Listing is one instrument on one exchange, as EODHD's search finds it.
type Listing struct {
	ISIN ledger.ISIN
	// Code is the instrument's ticker on the exchange, such as "SYNTH".
	Code string
	// Exchange is EODHD's exchange code, such as "XETRA".
	Exchange string
	// Currency is the currency EODHD quotes the listing in, such as "EUR",
	// or "GBX" for pence sterling.
	Currency string
}

// Ticker is the listing's EODHD ticker, such as "SYNTH.XETRA".
func (l Listing) Ticker() string { return l.Code + "." + l.Exchange }

// Pick returns the one listing of isin on a supported exchange, on exchange
// when that is not empty. Listings of other ISINs, which EODHD's fuzzy
// search also returns, are ignored, and so are repeats of one ticker. No
// listing is ErrNotListed; several, with no exchange asked for, is
// ErrAmbiguous, naming their tickers.
func Pick(listings []Listing, isin ledger.ISIN, exchange string) (Listing, error) {
	found := candidates(listings, isin, exchange)
	switch len(found) {
	case 0:
		if exchange != "" {
			return Listing{}, fmt.Errorf("%w: %s on %s", ErrNotListed, isin, exchange)
		}
		return Listing{}, fmt.Errorf("%w: %s", ErrNotListed, isin)
	case 1:
		return found[0], nil
	}
	tickers := make([]string, len(found))
	for i, f := range found {
		tickers[i] = f.Ticker()
	}
	return Listing{}, fmt.Errorf("%w: %s as %s; name one exchange", ErrAmbiguous, isin, strings.Join(tickers, ", "))
}

// candidates are the listings of isin on a supported exchange, on exchange
// when that is not empty, each ticker once.
func candidates(listings []Listing, isin ledger.ISIN, exchange string) []Listing {
	var found []Listing
	for _, l := range listings {
		if l.ISIN != isin || !SupportedExchange(l.Exchange) || (exchange != "" && l.Exchange != exchange) {
			continue
		}
		if !slices.ContainsFunc(found, func(f Listing) bool { return f.Ticker() == l.Ticker() }) {
			found = append(found, l)
		}
	}
	return found
}

// Quote returns a close of the listing as money in its currency. A close in
// pence sterling becomes pounds, divided by 100 exactly.
func (l Listing) Quote(closing kernel.Decimal) (kernel.Money, error) {
	code, amount := l.Currency, closing
	if code == penceCode {
		code, amount = "GBP", closing.Mul(kernel.NewDecimal(1, -2))
	}
	c, err := kernel.ParseCurrency(code)
	if err != nil {
		return kernel.Money{}, fmt.Errorf("%w: %s: %w", ErrQuoteCurrency, l.Ticker(), err)
	}
	return kernel.NewMoney(amount, c), nil
}
