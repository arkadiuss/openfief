package postgres_test

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"github.com/arkadiuss/openfief/internal/jurisdiction/pl"
	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
	"github.com/arkadiuss/openfief/internal/valuation"
	"github.com/arkadiuss/openfief/internal/valuation/postgres"
)

// inputs are synthetic inputs of every kind.
type inputs struct {
	prices   []valuation.Price
	rates    []valuation.FXRate
	readings []valuation.IndexReading
	bond     ledger.ISIN
	terms    pl.Terms
}

func syntheticInputs(t *testing.T) inputs {
	t.Helper()
	etf := instrument(t, etfISIN, "SYNTH").String()
	usd, err := kernel.ParseCurrency("USD")
	if err != nil {
		t.Fatal(err)
	}
	series, err := pl.ParseSeries("EDO0336")
	if err != nil {
		t.Fatal(err)
	}
	return inputs{
		prices: []valuation.Price{
			{Instrument: etf, Date: date(t, "2026-03-02"), Close: money(t, "101.37", "USD"), Source: "vendor"},
			{Instrument: etf, Date: date(t, "2026-03-03"), Close: money(t, "102.5", "USD"), Source: "vendor"},
		},
		rates:    []valuation.FXRate{{Currency: usd, Date: date(t, "2026-03-02"), Rate: dec(t, "4.0123"), Source: "nbp-a"}},
		readings: []valuation.IndexReading{{Kind: valuation.CPI, Year: 2026, Month: time.January, Value: dec(t, "0.029"), Source: "gus"}},
		bond:     instrument(t, bondISIN, "EDO0336").ISIN,
		terms:    pl.Terms{Series: series, FirstPeriodRate: dec(t, "0.0560"), Margin: dec(t, "0.02"), EarlyRedemptionFee: money(t, "3.00", "PLN")},
	}
}

// saveInputs stores in, writing the second price twice to show that an
// upsert replaces it.
func saveInputs(t *testing.T, store *postgres.Store, in *inputs) {
	t.Helper()
	ctx := t.Context()
	first := in.prices[1]
	first.Close = money(t, "1", "USD")
	err := store.SavePrices(ctx, []valuation.Price{in.prices[0], first}, nil)
	if err == nil {
		err = store.SavePrices(ctx, in.prices[1:], nil)
	}
	if err == nil {
		err = store.SaveFXRates(ctx, in.rates, nil)
	}
	if err == nil {
		err = store.SaveIndexReadings(ctx, in.readings, nil)
	}
	if err == nil {
		err = store.SaveBondTerms(ctx, in.bond, &in.terms, nil)
	}
	if err == nil {
		err = store.SaveTaxWrapper(ctx, acc2, pl.IKE, nil)
	}
	if err != nil {
		t.Fatalf("saving inputs: %v", err)
	}
}

func TestInputsRoundTrip(t *testing.T) {
	t.Parallel()
	store, _ := newStore(t)
	ctx := t.Context()
	in := syntheticInputs(t)
	saveInputs(t, store, &in)

	gotPrices, err := store.Prices(ctx, in.prices[0].Instrument, date(t, "2026-03-03"))
	check(t, "prices", in.prices, gotPrices, err)
	gotPrices, err = store.Prices(ctx, in.prices[0].Instrument, date(t, "2026-03-02"))
	check(t, "prices through 2026-03-02", in.prices[:1], gotPrices, err)
	gotRates, err := store.FXRates(ctx, in.rates[0].Currency, date(t, "2026-03-09"))
	check(t, "FX rates", in.rates, gotRates, err)
	gotReadings, err := store.IndexReadings(ctx)
	check(t, "index readings", in.readings, gotReadings, err)
	gotTerms, ok, err := store.BondTerms(ctx, in.bond)
	check(t, "bond terms", in.terms, gotTerms, err)
	_, notBond, err := store.BondTerms(ctx, instrument(t, etfISIN, "SYNTH").ISIN)
	if !ok || notBond || err != nil {
		t.Errorf("BondTerms found: bond %v, ETF %v, %v", ok, notBond, err)
	}
	w1, err1 := store.TaxWrapper(ctx, acc1)
	w2, err2 := store.TaxWrapper(ctx, acc2)
	if w1 != pl.Taxable || w2 != pl.IKE || err1 != nil || err2 != nil {
		t.Errorf("TaxWrapper = %v, %v (%v, %v); want taxable and IKE", w1, w2, err1, err2)
	}
}

func check[T any](t *testing.T, what string, want, got T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("%s: %v", what, err)
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("%s (-want +got):\n%s", what, diff)
	}
}

func TestReplaceValuations(t *testing.T) {
	t.Parallel()
	store, _ := newStore(t)
	ctx := t.Context()
	etf := instrument(t, etfISIN, "SYNTH")
	row := func(on string) valuation.HoldingValuation {
		return valuation.HoldingValuation{
			Account: acc1, Instrument: etf, Date: date(t, on), Basis: valuation.Market, Quantity: kernel.NewQuantity(dec(t, "12.5")),
			Value: money(t, "1267.125", "USD"), PLN: money(t, "5084.09", "PLN"), Source: "vendor",
			PriceDate: date(t, "2026-03-02"), RateDate: date(t, "2026-03-02"),
		}
	}
	gap := valuation.HoldingValuation{
		Account: acc1, Instrument: etf, Date: date(t, "2026-03-04"), Basis: valuation.Market, Quantity: kernel.NewQuantity(dec(t, "12.5")),
		Source: "vendor", PriceDate: date(t, "2026-03-02"), Missing: valuation.MissingFXRate, Detail: "no FX rate: USD on or before 2026-03-04",
	}
	all := []valuation.HoldingValuation{row("2026-03-02"), row("2026-03-03"), gap}
	err := store.ReplaceValuations(ctx, acc1, date(t, "2026-03-02"), all, date(t, "2026-03-03"))
	if err != nil {
		t.Fatalf("ReplaceValuations: %v", err)
	}
	got, err := store.Valuations(ctx, acc1, date(t, "2026-03-01"), date(t, "2026-03-31"))
	check(t, "valuations", all, got, err)
	// Revaluing from 2026-03-03 replaces that day on and keeps the day before.
	err = store.ReplaceValuations(ctx, acc1, date(t, "2026-03-03"), []valuation.HoldingValuation{row("2026-03-03")}, date(t, "2026-03-05"))
	if err != nil {
		t.Fatalf("ReplaceValuations again: %v", err)
	}
	got, err = store.Valuations(ctx, acc1, date(t, "2026-03-01"), date(t, "2026-03-31"))
	check(t, "valuations after revaluing from 2026-03-03", all[:2], got, err)
	through, ok, err := store.ValuedThrough(ctx, acc1)
	check(t, "valued-through date", date(t, "2026-03-05"), through, err)
	if _, never, err := store.ValuedThrough(ctx, ledger.AccountID("acc-9")); !ok || never || err != nil {
		t.Errorf("ValuedThrough found: %v, unknown account %v, %v", ok, never, err)
	}
}
