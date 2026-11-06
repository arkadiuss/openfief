package valuation_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/arkadiuss/openfief/internal/jurisdiction"
	"github.com/arkadiuss/openfief/internal/jurisdiction/pl"
	"github.com/arkadiuss/openfief/internal/valuation"
)

// TestUnclassified checks which holdings the owner is asked to classify for
// OKI's exemptions (D-132, D-148): each instrument priced at market and
// valued, once, in ISIN order; not a bond, a row with a missing input or an
// instrument the owner classified, and nothing in a wrapper without a
// yearly tax.
func TestUnclassified(t *testing.T) {
	t.Parallel()
	const fund = "PLTEST000010"
	row := func(isin string, basis valuation.Basis, missing valuation.MissingReason) valuation.HoldingValuation {
		return valuation.HoldingValuation{
			Instrument: instrument(t, isin, "SYNTH"), Date: date(t, "2027-03-02"), Basis: basis,
			Value: money(t, "100", "PLN"), Reporting: money(t, "100", "PLN"), Missing: missing,
		}
	}
	rows := []valuation.HoldingValuation{
		row(fund, valuation.Market, 0), row(etfISIN, valuation.Market, 0), row(fund, valuation.Market, 0),
		row("PLEDO0536000", valuation.GrossAccrued, 0), row("PLTEST000028", valuation.Market, valuation.MissingPrice),
	}
	tests := []struct {
		name    string
		wrapper jurisdiction.Wrapper
		owner   map[string]jurisdiction.ExemptionClass
		want    []string
	}{
		{name: "OKI", wrapper: "oki", want: []string{etfISIN, fund}},
		{name: "OKI, one classified by the owner", wrapper: "oki", owner: map[string]jurisdiction.ExemptionClass{fund: pl.Investment}, want: []string{etfISIN}},
		{name: "taxable has no yearly tax", wrapper: "taxable"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := valuation.Unclassified(pl.NewPack(), tt.wrapper, rows, tt.owner)
			if diff := cmp.Diff(tt.want, got, cmpopts.EquateEmpty()); diff != "" {
				t.Errorf("(-want +got):\n%s", diff)
			}
		})
	}
}
