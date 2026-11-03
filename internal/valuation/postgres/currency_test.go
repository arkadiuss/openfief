package postgres_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/valuation"
)

// TestStoredCurrenciesAreNamed checks D-180 in the store: a valuation stored
// in EUR, as a pack reporting in EUR stores it, reads back in EUR for its
// gross, net, liquidation, cost basis and sale cost; rates of one currency
// into two bases are both kept, each naming its base, and a rate quoted in
// its own currency is refused; and no constraint of the valuation schema
// names PLN.
func TestStoredCurrenciesAreNamed(t *testing.T) {
	t.Parallel()
	store, pool := newStore(t)
	ctx := t.Context()
	etf := instrument(t, etfISIN, "SYNTH")
	eur := valuation.HoldingValuation{
		Account: acc1, Instrument: etf, Date: date(t, "2026-03-02"), Basis: valuation.Market, Quantity: kernel.NewQuantity(dec(t, "2")),
		Value: money(t, "200", "USD"), Reporting: money(t, "180.00", "EUR"), Source: "vendor",
		PriceDate: date(t, "2026-03-02"), RateDate: date(t, "2026-03-02"),
		Net: money(t, "175.00", "EUR"), Liquidation: money(t, "174.00", "EUR"),
		TaxCost: money(t, "170.00", "EUR"), SaleCost: money(t, "1.00", "EUR"),
	}
	if err := store.ReplaceValuations(ctx, acc1, eur.Date, []valuation.HoldingValuation{eur}, eur.Date); err != nil {
		t.Fatalf("ReplaceValuations: %v", err)
	}
	got, err := store.Valuations(ctx, acc1, eur.Date, eur.Date)
	check(t, "valuations in EUR", []valuation.HoldingValuation{eur}, got, err)

	usd := money(t, "0", "USD").Currency()
	rates := []valuation.FXRate{
		{Currency: usd, Base: money(t, "0", "PLN").Currency(), Date: date(t, "2026-03-02"), Rate: dec(t, "4"), Source: "fx-a"},
		{Currency: usd, Base: money(t, "0", "EUR").Currency(), Date: date(t, "2026-03-02"), Rate: dec(t, "0.9"), Source: "fx-a"},
	}
	if err := store.SaveFXRates(ctx, rates, nil); err != nil {
		t.Fatalf("SaveFXRates: %v", err)
	}
	stored := textRows(t, pool, `SELECT currency || '/' || base || ' ' || rate::text FROM valuation.fx_rates ORDER BY base`)
	if diff := cmp.Diff([]string{"USD/EUR 0.9", "USD/PLN 4"}, stored); diff != "" {
		t.Errorf("stored rates (-want +got):\n%s", diff)
	}
	own := rates[0]
	own.Base = usd
	if err := store.SaveFXRates(ctx, []valuation.FXRate{own}, nil); err == nil {
		t.Error("SaveFXRates of a rate in its own currency: no error")
	}
	named := textRows(t, pool, `
		SELECT conname FROM pg_constraint
		WHERE connamespace = 'valuation'::regnamespace AND pg_get_constraintdef(oid) LIKE '%PLN%' ORDER BY conname`)
	if len(named) != 0 {
		t.Errorf("constraints of the valuation schema naming PLN: %v", named)
	}
}
