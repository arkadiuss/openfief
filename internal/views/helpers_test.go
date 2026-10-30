package views_test

import (
	"testing"

	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
	"github.com/arkadiuss/openfief/internal/valuation"
)

// Synthetic ISINs with valid check digits; no real instrument is implied.
const (
	isinPL = "PLTEST000010"
	isinIE = "IETEST000026"
)

// pln is the reporting currency of the tests.
func pln(t *testing.T) kernel.Currency {
	t.Helper()
	c, err := kernel.ParseCurrency("PLN")
	if err != nil {
		t.Fatalf("ParseCurrency(PLN): %v", err)
	}
	return c
}

func money(t *testing.T, amount, code string) kernel.Money {
	t.Helper()
	c, err := kernel.ParseCurrency(code)
	if err != nil {
		t.Fatalf("ParseCurrency(%q): %v", code, err)
	}
	m, err := kernel.ParseMoney(amount, c)
	if err != nil {
		t.Fatalf("ParseMoney(%q): %v", amount, err)
	}
	return m
}

func date(t *testing.T, s string) kernel.Date {
	t.Helper()
	d, err := kernel.ParseDate(s)
	if err != nil {
		t.Fatalf("ParseDate(%q): %v", s, err)
	}
	return d
}

func instrument(t *testing.T, isin, symbol string) ledger.Instrument {
	t.Helper()
	i, err := ledger.ParseISIN(isin)
	if err != nil {
		t.Fatalf("ParseISIN(%q): %v", isin, err)
	}
	return ledger.Instrument{ISIN: i, Symbol: symbol}
}

// val is a holding valuation in account acc-1 of the instrument isin/symbol:
// value in its currency, pln its value in PLN; a value of "missing" lacks a
// price.
type val struct {
	isin, symbol, day, value, currency, pln string
}

func (v *val) build(t *testing.T) valuation.HoldingValuation {
	t.Helper()
	if v.value == "missing" {
		return valuation.HoldingValuation{
			Account: "acc-1", Instrument: instrument(t, v.isin, v.symbol), Date: date(t, v.day),
			Basis: valuation.Market, Missing: valuation.MissingPrice,
		}
	}
	return valuation.HoldingValuation{
		Account: "acc-1", Instrument: instrument(t, v.isin, v.symbol), Date: date(t, v.day),
		Basis: valuation.Market, Source: "test", Value: money(t, v.value, v.currency), PLN: money(t, v.pln, "PLN"),
		Net: money(t, v.pln, "PLN"), Liquidation: money(t, v.pln, "PLN"),
	}
}

func vals(t *testing.T, vs []val) []valuation.HoldingValuation {
	t.Helper()
	out := make([]valuation.HoldingValuation, 0, len(vs))
	for i := range vs {
		out = append(out, vs[i].build(t))
	}
	return out
}
