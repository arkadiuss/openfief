package valuation_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/arkadiuss/openfief/internal/jurisdiction/pl"
	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
	"github.com/arkadiuss/openfief/internal/valuation"
)

// eurPack is the Polish pack's rules reporting in EUR: a fake pack whose
// reporting currency isn't PLN, to show that valuation takes the currency
// from the pack (D-180).
type eurPack struct {
	*pl.Pack
}

func (eurPack) ReportingCurrency() kernel.Currency {
	c, err := kernel.ParseCurrency("EUR")
	if err != nil {
		return kernel.Currency{}
	}
	return c
}

// TestValueMarketHoldingInPackCurrency checks that a holding at market is
// stored in its pack's reporting currency, converted with the rate into
// that currency only (D-180): the Polish pack's PLN with the rate into PLN,
// a pack reporting in EUR with the rate into EUR, and with no rate into
// the pack's currency the FX rate is missing, though one into another
// currency exists.
func TestValueMarketHoldingInPackCurrency(t *testing.T) {
	t.Parallel()
	etf := instrument(t, etfISIN, "SYNTH")
	prices := []valuation.Price{{Instrument: etf.String(), Date: date(t, "2026-03-06"), Close: money(t, "100", "USD"), Source: "vendor"}}
	intoEUR := rate(t, "USD", "2026-03-06", "0.9000")
	intoEUR.Base = cur(t, "EUR")
	intoPLN := rate(t, "USD", "2026-03-06", "4.0000")
	tests := []struct {
		name    string
		pack    *valuation.AccountTerms
		rates   []valuation.FXRate
		want    string
		missing valuation.MissingReason
	}{
		{name: "the Polish pack: into PLN", pack: taxableTerms(), rates: []valuation.FXRate{intoEUR, intoPLN}, want: "800 PLN"},
		{
			name: "a pack reporting in EUR: into EUR", pack: &valuation.AccountTerms{Pack: eurPack{pl.NewPack()}, Wrapper: "taxable"},
			rates: []valuation.FXRate{intoPLN, intoEUR}, want: "180 EUR",
		},
		{
			name: "no rate into the pack's currency", pack: &valuation.AccountTerms{Pack: eurPack{pl.NewPack()}, Wrapper: "taxable"},
			rates: []valuation.FXRate{intoPLN}, missing: valuation.MissingFXRate,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			h := holding(t, etf, "2026-03-06", "2")
			m := valuation.MarketHolding{Prices: prices, Rates: tt.rates, Lots: []ledger.Lot{marketLot(t, etf, "2026-03-02", "2", "190", "USD")}, Terms: tt.pack}
			got, err := valuation.ValueMarketHolding(&h, &m)
			if err != nil {
				t.Fatalf("ValueMarketHolding: %v", err)
			}
			if diff := cmp.Diff([]any{tt.missing, tt.want}, []any{got.Missing, reportingOf(&got)}); diff != "" {
				t.Errorf("missing input and stored value (-want +got):\n%s", diff)
			}
		})
	}
}

// reportingOf returns v's value in its reporting currency, or "" when it
// has none.
func reportingOf(v *valuation.HoldingValuation) string {
	if !v.Valued() {
		return ""
	}
	return v.Reporting.String()
}
