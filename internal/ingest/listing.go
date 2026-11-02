package ingest

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
)

// ErrUnknownTradingCurrency reports an opening lot priced in its
// instrument's trading currency (D-159) whose confirmed ticker's listing
// names no currency and whose market names no exchange OpenFief knows.
var ErrUnknownTradingCurrency = errors.New("the instrument's trading currency is unknown")

// TradingCurrencyChoiceError reports a ticker's confirmation that picks no
// trading currency, or one its exchange doesn't quote in, when the listing
// names none and the exchange quotes in more than one (D-170): Choices are
// the codes to pick from. With no Choices, it picked one when there is
// none to pick.
type TradingCurrencyChoiceError struct {
	Symbol  string
	Picked  string
	Choices []string
}

// Error names the choices, or says there are none.
func (e *TradingCurrencyChoiceError) Error() string {
	if len(e.Choices) == 0 {
		return fmt.Sprintf("%s's trading currency is settled by its listing or exchange, %q can't be picked", e.Symbol, e.Picked)
	}
	return fmt.Sprintf("pick %s's trading currency, one of %s: its listing names none and its exchange quotes in each",
		e.Symbol, strings.Join(e.Choices, ", "))
}

// Listing is what the owner's confirmation of a ticker settles (D-116,
// D-159, D-170): the ISIN every entry of the ticker is booked with, and
// the trading currency its prices are quoted in, the currency of the
// confirmed listing or, when it names none, of the exchange the ticker's
// market suffix names, or the one the owner picked when that exchange
// quotes in more than one. Minor says prices are quoted in the currency's
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

// ExchangeCurrencies returns the trading currency codes the exchange of a
// market in country, an ISO 3166-1 alpha-2 code a broker's ticker suffix
// names, quotes in (D-159, D-170), as ParseTradingCurrency reads them, or
// none for a country OpenFief knows no exchange of: London quotes in
// pence, pounds and dollars.
func ExchangeCurrencies(country string) []string {
	switch country {
	case "DE", "FR", "NL", "BE", "PT", "IE", "IT", "ES", "FI", "AT":
		return []string{"EUR"}
	case "GB":
		return []string{"GBX", "GBP", "USD"}
	}
	if code := map[string]string{
		"PL": "PLN", "US": "USD", "CH": "CHF", "CZ": "CZK", "DK": "DKK", "SE": "SEK", "NO": "NOK",
	}[country]; code != "" {
		return []string{code}
	}
	return nil
}

// settledCode returns the trading currency code the confirmation of isin
// for t settles without the owner (D-159): proposal's, the market-data
// connector's, when the owner confirmed its ISIN and it names one, else the
// only one t's exchange quotes in, else "".
func settledCode(t *Ticker, proposal *ISINProposal, isin ledger.ISIN) string {
	if proposal.ISIN == isin {
		if _, _, ok := ParseTradingCurrency(proposal.Currency); ok {
			return proposal.Currency
		}
	}
	if codes := ExchangeCurrencies(t.Country); len(codes) == 1 {
		return codes[0]
	}
	return ""
}

// CurrencyChoices returns the trading currency codes the owner picks one
// of when confirming isin for ticker t (D-170): those of the exchange t's
// market names, when the listing names no currency and that exchange quotes
// in more than one; none when the confirmation settles it (settledCode) or
// the exchange is unknown.
func CurrencyChoices(t *Ticker, proposal *ISINProposal, isin ledger.ISIN) []string {
	if settledCode(t, proposal, isin) != "" {
		return nil
	}
	if codes := ExchangeCurrencies(t.Country); len(codes) > 1 {
		return codes
	}
	return nil
}

// ListingOf returns the listing the owner's confirmation of isin for ticker
// t settles (D-159, D-170): the trading currency of proposal, the
// market-data connector's, when the owner confirmed its ISIN and it names
// one, else the only one the exchange t's market names quotes in, else
// chosen, the code the owner picked, when it is one of CurrencyChoices,
// else none. No currency is assumed.
func ListingOf(t *Ticker, proposal *ISINProposal, isin ledger.ISIN, chosen string) Listing {
	out := Listing{ISIN: isin}
	code := settledCode(t, proposal, isin)
	if code == "" && slices.Contains(CurrencyChoices(t, proposal, isin), chosen) {
		code = chosen
	}
	if c, minor, ok := ParseTradingCurrency(code); ok {
		out.Currency, out.Minor = c, minor
	}
	return out
}

// ConfirmListing returns the listing the owner's confirmation of isin for
// ticker t with the trading currency code chosen, "" for none, settles
// (ListingOf). When the confirmation must pick one (CurrencyChoices), a
// chosen code that is none of them fails with a *TradingCurrencyChoiceError
// naming them, so the ticker's entries keep waiting (D-170); otherwise
// chosen must be "" or the code the confirmation settles.
func ConfirmListing(t *Ticker, proposal *ISINProposal, isin ledger.ISIN, chosen string) (Listing, error) {
	choices := CurrencyChoices(t, proposal, isin)
	switch {
	case len(choices) > 0 && !slices.Contains(choices, chosen):
		return Listing{}, &TradingCurrencyChoiceError{Symbol: t.Symbol, Picked: chosen, Choices: choices}
	case len(choices) == 0 && chosen != "" && chosen != settledCode(t, proposal, isin):
		return Listing{}, &TradingCurrencyChoiceError{Symbol: t.Symbol, Picked: chosen}
	}
	return ListingOf(t, proposal, isin, chosen), nil
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
