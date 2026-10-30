package demo_test

import (
	"testing"

	"github.com/arkadiuss/openfief/internal/demo"
	"github.com/arkadiuss/openfief/internal/valuation"
)

// TestCrossCheckPricesFlagOne checks that the dataset's IBKR mark prices,
// compared with its prices as the core compares them (D-111), give exactly
// one flagged price check at the default 1% (D-110), on the latest
// statement's day, and match everywhere else, so the demo's price checks
// page shows a flagged check.
func TestCrossCheckPricesFlagOne(t *testing.T) {
	t.Parallel()
	ds := generate(t, testSeed, testToday)
	if len(ds.CrossCheckPrices) == 0 {
		t.Fatal("the dataset has no cross-check prices")
	}
	checks, err := valuation.CheckPrices(ds.Prices, ds.CrossCheckPrices, valuation.Thresholds{}, nil)
	if err != nil {
		t.Fatalf("CheckPrices: %v", err)
	}
	if len(checks) != len(ds.CrossCheckPrices) {
		t.Errorf("%d checks of %d cross-check prices; want every one compared", len(checks), len(ds.CrossCheckPrices))
	}
	var flagged []valuation.PriceCheck
	latest := ds.CrossCheckPrices[len(ds.CrossCheckPrices)-1].Date
	for i := range checks {
		c := &checks[i]
		if c.CrossCheck.Source != demo.IBKRInstitution || c.Price.Source != demo.PriceSource {
			t.Errorf("check of %s on %s from %s against %s; want IBKR's mark against the demo's price",
				c.CrossCheck.Instrument, c.CrossCheck.Date, c.CrossCheck.Source, c.Price.Source)
		}
		if c.Status == valuation.PriceFlagged {
			flagged = append(flagged, *c)
		}
	}
	if len(flagged) != 1 || !flagged[0].CrossCheck.Date.Equal(latest) {
		t.Errorf("flagged checks = %v, want one, on the latest statement's day %s", flagged, latest)
	}
}
