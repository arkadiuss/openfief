package ingest

import (
	"errors"

	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
)

// ErrUnknownTradingCurrency reports an opening lot priced in its
// instrument's trading currency (D-159) whose confirmed ticker's listing
// names no currency and whose market names no exchange OpenFief knows.
var ErrUnknownTradingCurrency = errors.New("the instrument's trading currency is unknown")

// Listing is what the owner's confirmation of a ticker settles (D-116,
// D-159): the ISIN every entry of the ticker is booked with, and the
// trading currency its prices are quoted in, the currency of the
// confirmed listing or, when it names none, of the exchange the ticker's
// market suffix names. Minor says prices are quoted in the currency's
// minor unit, such as London's pence, and are converted to the major
// unit. A zero Currency is no known trading currency.
type Listing struct {
	ISIN     ledger.ISIN
	Currency kernel.Currency
	Minor    bool
}

// minorUnitOf returns the currency whose minor unit, a hundredth of it, a
// listing's trading currency code names, and whether it names one: pence
// sterling, South African cents and Israeli agorot.
func minorUnitOf(code string) (string, bool) {
	switch code {
	case "GBX", "GBp":
		return "GBP", true
	case "ZAc", "ZAC":
		return "ZAR", true
	case "ILA":
		return "ILS", true
	}
	return "", false
}

// ParseTradingCurrency reads the currency a listing quotes its prices in,
// as a market-data source names it (D-159): an ISO 4217 code, or the code
// of a minor unit such as "GBX" for pence sterling, which is its currency
// with Minor set. Anything else, an empty code included, is no currency.
func ParseTradingCurrency(code string) (kernel.Currency, bool, bool) {
	if major, ok := minorUnitOf(code); ok {
		c, err := kernel.ParseCurrency(major)
		return c, true, err == nil
	}
	c, err := kernel.ParseCurrency(code)
	if err != nil {
		return kernel.Currency{}, false, false
	}
	return c, false, true
}

// exchangeCurrencyCode returns the trading currency of the exchange of a
// market country a broker's ticker suffix names (D-159), what its prices
// are quoted in, or "" for none: London quotes shares in pence.
func exchangeCurrencyCode(country string) string {
	switch country {
	case "DE", "FR", "NL", "BE", "PT", "IE", "IT", "ES", "FI", "AT":
		return "EUR"
	case "GB":
		return "GBX"
	}
	return map[string]string{
		"PL": "PLN", "US": "USD", "CH": "CHF", "CZ": "CZK", "DK": "DKK", "SE": "SEK", "NO": "NOK",
	}[country]
}

// ExchangeCurrency returns the trading currency of the exchange of a
// market in country, an ISO 3166-1 alpha-2 code, and whether its prices are
// quoted in the minor unit; false for a country OpenFief knows no exchange
// of.
func ExchangeCurrency(country string) (kernel.Currency, bool, bool) {
	code := exchangeCurrencyCode(country)
	if code == "" {
		return kernel.Currency{}, false, false
	}
	return ParseTradingCurrency(code)
}

// ListingOf returns the listing the owner's confirmation of isin for ticker
// t settles (D-159): the trading currency of proposal, the market-data
// connector's, when the owner confirmed its ISIN and it names a currency,
// else that of the exchange t's market names (ExchangeCurrency), else
// none.
func ListingOf(t *Ticker, proposal *ISINProposal, isin ledger.ISIN) Listing {
	out := Listing{ISIN: isin}
	if proposal.ISIN == isin {
		if c, minor, ok := ParseTradingCurrency(proposal.Currency); ok {
			out.Currency, out.Minor = c, minor
			return out
		}
	}
	if c, minor, ok := ExchangeCurrency(t.Country); ok {
		out.Currency, out.Minor = c, minor
	}
	return out
}

// Price returns price, quoted on l, in l's currency's major unit: a price
// in the minor unit divided by 100, exactly.
func (l Listing) Price(price kernel.Decimal) kernel.Decimal {
	if l.Minor {
		return price.Mul(kernel.NewDecimal(1, -minorUnitPlaces))
	}
	return price
}

// minorUnitPlaces are the decimal places of a minor unit that is a
// hundredth of its currency.
const minorUnitPlaces = 2

// CurrencyPlaces returns the decimal places of c's minor unit (ISO 4217):
// none for a currency without one, such as JPY, and two for the others a
// broker trades in.
func CurrencyPlaces(c kernel.Currency) int32 {
	switch c.String() {
	case "JPY", "KRW", "ISK", "CLP", "VND":
		return 0
	}
	return minorUnitPlaces
}
