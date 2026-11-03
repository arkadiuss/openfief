package postgres_test

import (
	"slices"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/arkadiuss/openfief/internal/valuation"
	"github.com/arkadiuss/openfief/internal/valuation/postgres"
)

// TestOneSourcePerSeries checks D-94 in the store: every source's prices
// are stored, an instrument's prices come from one source only, chosen from
// the ranking when they are first needed (by default the first-party
// connector's) and recorded, so a day it lacks is never filled from
// another source; the other sources' prices come back as cross-checks; a
// new ranking changes no recorded source and revalues nothing, but chooses
// for an instrument first needed after it; and the owner's switch of one
// instrument's source revalues only when it changes.
func TestOneSourcePerSeries(t *testing.T) {
	t.Parallel()
	store, pool := newStore(t)
	ctx := t.Context()
	const etf, other = etfISIN + "/TSTB", "XX00000SYNT3/SYNT"
	price := func(instrument, on, closing, source string) valuation.Price {
		return valuation.Price{Instrument: instrument, Date: date(t, on), Close: money(t, closing, "USD"), Source: source}
	}
	eodhd2, ibkr2, ibkr3 := price(etf, "2026-03-02", "100", "prices-eu"), price(etf, "2026-03-02", "101", "ibkr-main"), price(etf, "2026-03-03", "102", "ibkr-main")
	if err := store.SavePrices(ctx, []valuation.Price{eodhd2, ibkr2, ibkr3}, nil); err != nil {
		t.Fatalf("SavePrices: %v", err)
	}
	saveSources(t, store)
	got, err := store.Prices(ctx, etf, date(t, "2026-03-31"))
	check(t, "prices: EODHD's only, IBKR's never fills its gap", []valuation.Price{eodhd2}, got, err)
	own, checks, err := store.CrossCheckedPrices(ctx)
	check(t, "cross-checked prices: the source's own", []valuation.Price{eodhd2}, own, err)
	check(t, "the other source's prices as cross-checks", valuation.OtherSourceChecks([]valuation.Price{ibkr2, ibkr3}), checks, err)

	changed, err := store.SaveSourceRanking(ctx, valuation.KindPrices, []string{"ibkr-main", "prices-eu"}, nil)
	if err != nil || !changed {
		t.Fatalf("SaveSourceRanking = %v, %v; want a change", changed, err)
	}
	got, err = store.Prices(ctx, etf, date(t, "2026-03-31"))
	check(t, "a new ranking keeps the recorded source", []valuation.Price{eodhd2}, got, err)
	checkPending(t, pool, nil, nil)
	if saveErr := store.SavePrices(ctx, []valuation.Price{price(other, "2026-03-02", "5", "prices-eu"), price(other, "2026-03-02", "6", "ibkr-main")}, nil); saveErr != nil {
		t.Fatalf("SavePrices: %v", saveErr)
	}
	got, err = store.Prices(ctx, other, date(t, "2026-03-31"))
	check(t, "an instrument first needed after the new ranking", []valuation.Price{price(other, "2026-03-02", "6", "ibkr-main")}, got, err)

	checkOwnerSwitch(t, store, pool, etf, []valuation.Price{ibkr2, ibkr3})
	checkRankingKept(t, store, pool)
}

// checkOwnerSwitch switches instrument's source to ibkr-main as the owner
// and checks that it revalues once, that the prices then come from it
// alone, and that the series lists it as the owner's choice.
func checkOwnerSwitch(t *testing.T, store *postgres.Store, pool *pgxpool.Pool, instrument string, want []valuation.Price) {
	t.Helper()
	ctx := t.Context()
	series := valuation.PriceSeries(instrument)
	ss := valuation.SeriesSource{Series: series, Source: "ibkr-main"}
	changed, err := store.SaveSeriesSource(ctx, &ss, []valuation.Revaluation{reval(t, "acc-1", "2026-03-02")})
	if err != nil || !changed {
		t.Fatalf("SaveSeriesSource = %v, %v; want a change", changed, err)
	}
	checkPending(t, pool, []string{"acc-1 2026-03-02 -"}, []string{"acc-1 available"})
	got, err := store.Prices(ctx, instrument, date(t, "2026-03-31"))
	check(t, "prices after the owner's switch", want, got, err)
	if changed, err = store.SaveSeriesSource(ctx, &ss, []valuation.Revaluation{reval(t, "acc-2", "2026-03-02")}); err != nil || changed {
		t.Errorf("the same source again = %v, %v; want no change", changed, err)
	}
	all, err := store.SeriesSources(ctx)
	if err != nil {
		t.Fatalf("SeriesSources: %v", err)
	}
	i := slices.IndexFunc(all, func(s valuation.SeriesSource) bool { return s.Series == series })
	wantSS := valuation.SeriesSource{Series: series, Source: "ibkr-main", ByOwner: true, Available: []string{"ibkr-main", "prices-eu"}}
	if i < 0 {
		t.Fatalf("SeriesSources = %v, want %v among them", all, wantSS)
	}
	check(t, "the owner's choice", wantSS, all[i], nil)
}

// saveSources records two price sources and NBP's FX source.
func saveSources(t *testing.T, store *postgres.Store) {
	t.Helper()
	if err := store.SaveSources(t.Context(), []valuation.Source{
		{Kind: valuation.KindPrices, Name: "ibkr-main", Institution: "ibkr"},
		{Kind: valuation.KindPrices, Name: "prices-eu", Institution: "eodhd"},
		{Kind: valuation.KindFXRates, Name: "nbp-fx", Institution: "nbp"},
	}); err != nil {
		t.Fatalf("SaveSources: %v", err)
	}
}

// checkRankingKept checks that saving the stored price ranking again
// changes nothing, and that a ranking naming a source twice is refused and
// leaves the stored one.
func checkRankingKept(t *testing.T, store *postgres.Store, pool *pgxpool.Pool) {
	t.Helper()
	ctx := t.Context()
	if changed, err := store.SaveSourceRanking(ctx, valuation.KindPrices, []string{"ibkr-main", "prices-eu"}, nil); err != nil || changed {
		t.Errorf("the same ranking again = %v, %v; want no change", changed, err)
	}
	if _, err := store.SaveSourceRanking(ctx, valuation.KindPrices, []string{"a", "a"}, nil); err == nil {
		t.Error("a source ranked twice was stored; the table must refuse it")
	}
	if got := textRows(t, pool, "SELECT source FROM valuation.source_rankings WHERE kind = 'prices' ORDER BY position"); !slices.Equal(got, []string{"ibkr-main", "prices-eu"}) {
		t.Errorf("stored ranking = %v; a refused ranking changed it", got)
	}
}

// TestRatesAndReadingsOneSource checks that FX rates and index readings of
// every source are stored and that reads return those of each series' one
// source: NBP's rate by default, and for each index kind its own source,
// the owner's first-ranked one when it has readings (D-94).
func TestRatesAndReadingsOneSource(t *testing.T) {
	t.Parallel()
	store, pool := newStore(t)
	ctx := t.Context()
	saveSources(t, store)
	usd := money(t, "1", "USD").Currency()
	rate := func(source, on, value string) valuation.FXRate {
		return valuation.FXRate{Currency: usd, Base: money(t, "0", "PLN").Currency(), Date: date(t, on), Rate: dec(t, value), Source: source}
	}
	ecbOnly := rate("ecb-fx", "2026-03-03", "4.0200")
	if err := store.SaveFXRates(ctx, []valuation.FXRate{rate("ecb-fx", "2026-03-02", "4.0100"), rate("nbp-fx", "2026-03-02", "4.0000"), ecbOnly}, nil); err != nil {
		t.Fatalf("SaveFXRates: %v", err)
	}
	rates, err := store.FXRates(ctx, usd, date(t, "2026-03-31"))
	check(t, "FX rates: NBP's only, by default", []valuation.FXRate{rate("nbp-fx", "2026-03-02", "4.0000")}, rates, err)

	reading := func(kind valuation.IndexKind, source, value string) valuation.IndexReading {
		return valuation.IndexReading{Kind: kind, Year: 2026, Month: time.January, Value: dec(t, value), Source: source}
	}
	cpiGUS, cpiBonds, refBonds := reading(valuation.CPI, "gus-copy", "0.05"), reading(valuation.CPI, "bonds", "0.049"), reading(valuation.ReferenceRate, "bonds", "0.0575")
	if saveErr := store.SaveIndexReadings(ctx, []valuation.IndexReading{cpiGUS, cpiBonds, refBonds}, nil); saveErr != nil {
		t.Fatalf("SaveIndexReadings: %v", saveErr)
	}
	if _, rankErr := store.SaveSourceRanking(ctx, valuation.KindIndexReadings, []string{"gus-copy"}, nil); rankErr != nil {
		t.Fatalf("SaveSourceRanking: %v", rankErr)
	}
	readings, err := store.IndexReadings(ctx)
	check(t, "index readings: each kind from its own source", []valuation.IndexReading{cpiGUS, refBonds}, readings, err)
	if n := countRows(t, pool, "SELECT count(*) FROM valuation.index_readings"); n != 3 {
		t.Errorf("stored %d readings, want every source's kept", n)
	}
	checkGapThresholds(t, store)
}

// checkGapThresholds checks that a kind without the owner's gap threshold
// has none stored, that the owner's is stored and replaced, and that the
// table refuses one out of range.
func checkGapThresholds(t *testing.T, store *postgres.Store) {
	t.Helper()
	ctx := t.Context()
	got, err := store.GapThresholds(ctx)
	check(t, "no gap threshold set", map[valuation.MarketDataKind]int{}, got.Days, err)
	for _, days := range []int{10, 14} {
		if saveErr := store.SaveGapThreshold(ctx, valuation.KindPrices, days); saveErr != nil {
			t.Fatalf("SaveGapThreshold(%d): %v", days, saveErr)
		}
	}
	if saveErr := store.SaveGapThreshold(ctx, valuation.KindFXRates, 0); saveErr == nil {
		t.Error("a gap threshold of 0 days was stored; the table must refuse it")
	}
	got, err = store.GapThresholds(ctx)
	check(t, "the owner's gap threshold", map[valuation.MarketDataKind]int{valuation.KindPrices: 14}, got.Days, err)
}
