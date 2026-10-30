package postgres_test

import (
	"testing"

	"github.com/arkadiuss/openfief/internal/valuation"
)

func TestCrossCheckPricesRoundTrip(t *testing.T) {
	t.Parallel()
	store, pool := newStore(t)
	ctx := t.Context()
	const etf, other = etfISIN + "/TSTB", "PLTEST000010/TSTA"
	crossCheck := func(in, on, closing, code, source string) valuation.CrossCheckPrice {
		return valuation.CrossCheckPrice{Instrument: in, Date: date(t, on), Close: money(t, closing, code), Source: source}
	}
	price := func(in, on, closing, code string) valuation.Price {
		return valuation.Price{Instrument: in, Date: date(t, on), Close: money(t, closing, code), Source: "eodhd"}
	}
	prices := []valuation.Price{
		price(etf, "2026-03-05", "99", "USD"), price(etf, "2026-03-06", "100", "USD"),
		price(other, "2026-03-06", "50", "EUR"),
	}
	if err := store.SavePrices(ctx, prices, nil); err != nil {
		t.Fatalf("SavePrices: %v", err)
	}
	first := []valuation.CrossCheckPrice{
		crossCheck(etf, "2026-03-06", "100.4", "USD", "ibkr-b"), crossCheck(etf, "2026-03-06", "100.1", "USD", "ibkr-a"),
		crossCheck(etf, "2026-03-04", "98", "USD", "ibkr-a"),
	}
	if err := store.SaveCrossCheckPrices(ctx, first); err != nil {
		t.Fatalf("SaveCrossCheckPrices: %v", err)
	}
	// Saving one again for the same instrument, date and source replaces
	// it, as a retried batch does.
	again := crossCheck(etf, "2026-03-06", "100.2", "USD", "ibkr-a")
	if err := store.SaveCrossCheckPrices(ctx, []valuation.CrossCheckPrice{again}); err != nil {
		t.Fatalf("SaveCrossCheckPrices again: %v", err)
	}
	gotPrices, gotChecks, err := store.CrossCheckedPrices(ctx)
	check(t, "cross-checked prices: every date of the instrument", prices[:2], gotPrices, err)
	check(t, "cross-check prices", []valuation.CrossCheckPrice{first[2], again, first[0]}, gotChecks, err)

	// Cross-check prices change no valuation input, so they enqueue nothing,
	// and the valuation prices are untouched.
	checkPending(t, pool, nil, nil)
	stored, err := store.Prices(ctx, etf, date(t, "2026-03-31"))
	check(t, "valuation prices", prices[:2], stored, err)

	if err := store.SaveCrossCheckPrices(ctx, []valuation.CrossCheckPrice{crossCheck(etf, "2026-03-09", "0", "USD", "ibkr-a")}); err == nil {
		t.Error("saving a zero cross-check price succeeded; the table check must refuse it")
	}
}
