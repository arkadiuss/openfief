package postgres_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/arkadiuss/openfief/internal/valuation"
)

// TestPriceHistory checks that an ISIN's price history holds every listing
// and every source in the range, in listing, date and source order, and
// nothing of another ISIN or outside the range.
func TestPriceHistory(t *testing.T) {
	t.Parallel()
	store, _ := newStore(t)
	etf, otherListing := instrument(t, etfISIN, "SYNTH").String(), instrument(t, etfISIN, "SYNX").String()
	bond := instrument(t, bondISIN, "EDO0336").String()
	price := func(inst, day, closing, currency, source string) valuation.Price {
		return valuation.Price{Instrument: inst, Date: date(t, day), Close: money(t, closing, currency), Source: source}
	}
	all := []valuation.Price{
		price(etf, "2026-03-01", "100", "USD", "vendor"),
		price(etf, "2026-03-02", "101.37", "USD", "vendor"),
		price(etf, "2026-03-02", "101.40", "USD", "other-vendor"),
		price(etf, "2026-03-03", "102.5", "USD", "vendor"),
		price(otherListing, "2026-03-02", "93.10", "EUR", "vendor"),
		price(bond, "2026-03-02", "100", "PLN", "vendor"),
	}
	if err := store.SavePrices(t.Context(), all, nil); err != nil {
		t.Fatalf("SavePrices: %v", err)
	}
	got, err := store.PriceHistory(t.Context(), etfISIN, date(t, "2026-03-02"), date(t, "2026-03-03"))
	if err != nil {
		t.Fatalf("PriceHistory: %v", err)
	}
	want := []valuation.Price{all[2], all[1], all[3], all[4]}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("PriceHistory (-want +got):\n%s", diff)
	}
	none, err := store.PriceHistory(t.Context(), "XX0000000001", date(t, "2026-03-01"), date(t, "2026-03-03"))
	if err != nil || len(none) != 0 {
		t.Errorf("PriceHistory of an unpriced ISIN = %v, %v", none, err)
	}
	latest, err := store.LatestPrices(t.Context(), etfISIN)
	if err != nil {
		t.Fatalf("LatestPrices: %v", err)
	}
	if diff := cmp.Diff([]valuation.Price{all[2], all[3], all[4]}, latest); diff != "" {
		t.Errorf("LatestPrices (-want +got):\n%s", diff)
	}
}
