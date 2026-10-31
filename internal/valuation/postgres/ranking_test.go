package postgres_test

import (
	"slices"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/arkadiuss/openfief/internal/valuation"
	"github.com/arkadiuss/openfief/internal/valuation/postgres"
)

// TestRankedPrices checks that every source's prices are stored, that
// reads return the highest-ranked source's price of each date, by default
// the first-party connector's, that the owner's ranking replaces it and
// revalues only when it changes, and that lower-ranked prices come back as
// cross-checks (D-94, D-22).
func TestRankedPrices(t *testing.T) {
	t.Parallel()
	store, pool := newStore(t)
	ctx := t.Context()
	const etf = etfISIN + "/TSTB"
	price := func(on, closing, source string) valuation.Price {
		return valuation.Price{Instrument: etf, Date: date(t, on), Close: money(t, closing, "USD"), Source: source}
	}
	eodhd2, ibkr2, ibkr3 := price("2026-03-02", "100", "prices-eu"), price("2026-03-02", "101", "ibkr-main"), price("2026-03-03", "102", "ibkr-main")
	if err := store.SavePrices(ctx, []valuation.Price{eodhd2, ibkr2, ibkr3}, nil); err != nil {
		t.Fatalf("SavePrices: %v", err)
	}
	saveSources(t, store)
	got, err := store.Prices(ctx, etf, date(t, "2026-03-31"))
	check(t, "prices by default: EODHD's wins, IBKR's fills its gap", []valuation.Price{eodhd2, ibkr3}, got, err)
	winners, checks, err := store.CrossCheckedPrices(ctx)
	check(t, "cross-checked prices: every date of the instrument", []valuation.Price{eodhd2, ibkr3}, winners, err)
	check(t, "lower-ranked prices as cross-checks", valuation.LowerRankedChecks([]valuation.Price{ibkr2}), checks, err)

	changed, err := store.SaveSourceRanking(ctx, valuation.KindPrices, []string{"ibkr-main", "prices-eu"}, []valuation.Revaluation{reval(t, "acc-1", "2026-03-02")})
	if err != nil || !changed {
		t.Fatalf("SaveSourceRanking = %v, %v; want a change", changed, err)
	}
	checkPending(t, pool, []string{"acc-1 2026-03-02 -"}, []string{"acc-1 available"})
	got, err = store.Prices(ctx, etf, date(t, "2026-03-31"))
	check(t, "prices with IBKR ranked first", []valuation.Price{ibkr2, ibkr3}, got, err)
	ranking, err := store.SourceRanking(ctx, valuation.KindPrices)
	check(t, "the ranking", []string{"ibkr-main", "prices-eu"}, ranking.Names(), err)
	checkRankingKept(t, store, pool)
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

// TestRankedRatesAndReadings checks that FX rates and index readings of
// every source are stored and that reads return the highest-ranked one:
// NBP's rate by default, and the owner's first source's reading (D-94).
func TestRankedRatesAndReadings(t *testing.T) {
	t.Parallel()
	store, pool := newStore(t)
	ctx := t.Context()
	saveSources(t, store)
	usd := money(t, "1", "USD").Currency()
	rate := func(source, value string) valuation.FXRate {
		return valuation.FXRate{Currency: usd, Date: date(t, "2026-03-02"), Rate: dec(t, value), Source: source}
	}
	if err := store.SaveFXRates(ctx, []valuation.FXRate{rate("ecb-fx", "4.0100"), rate("nbp-fx", "4.0000")}, nil); err != nil {
		t.Fatalf("SaveFXRates: %v", err)
	}
	rates, err := store.FXRates(ctx, usd, date(t, "2026-03-31"))
	check(t, "FX rates: NBP's wins by default", []valuation.FXRate{rate("nbp-fx", "4.0000")}, rates, err)

	reading := func(source, value string) valuation.IndexReading {
		return valuation.IndexReading{Kind: valuation.CPI, Year: 2026, Month: time.January, Value: dec(t, value), Source: source}
	}
	if saveErr := store.SaveIndexReadings(ctx, []valuation.IndexReading{reading("gus-copy", "0.05"), reading("bonds", "0.049")}, nil); saveErr != nil {
		t.Fatalf("SaveIndexReadings: %v", saveErr)
	}
	if _, rankErr := store.SaveSourceRanking(ctx, valuation.KindIndexReadings, []string{"gus-copy"}, nil); rankErr != nil {
		t.Fatalf("SaveSourceRanking: %v", rankErr)
	}
	readings, err := store.IndexReadings(ctx)
	check(t, "index readings: the owner's first source wins", []valuation.IndexReading{reading("gus-copy", "0.05")}, readings, err)
	if n := countRows(t, pool, "SELECT count(*) FROM valuation.index_readings"); n != 2 {
		t.Errorf("stored %d readings, want both sources' kept", n)
	}
}
