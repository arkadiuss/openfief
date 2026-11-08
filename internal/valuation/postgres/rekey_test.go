package postgres_test

import (
	"slices"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// rekeyRows returns every row the re-key moves, as "table key detail", in
// order.
func rekeyRows(t *testing.T, pool *pgxpool.Pool) []string {
	t.Helper()
	rows, err := pool.Query(t.Context(), `
		SELECT 'bond_terms ' || isin || ' ' || source FROM valuation.bond_terms
		UNION ALL SELECT 'exemption_classes ' || instrument || ' ' || class FROM valuation.exemption_classes
		UNION ALL SELECT 'prices ' || instrument || ' ' || date || ' ' || close::text FROM valuation.prices
		UNION ALL SELECT 'cross_check_prices ' || instrument || ' ' || date FROM valuation.cross_check_prices
		UNION ALL SELECT 'price_check_thresholds ' || instrument || ' ' || threshold::text FROM valuation.price_check_thresholds
		UNION ALL SELECT 'series_sources ' || key || ' ' || source FROM valuation.series_sources`)
	if err != nil {
		t.Fatalf("query: %v", err)
	}
	got, err := pgx.CollectRows(rows, pgx.RowTo[string])
	if err != nil {
		t.Fatalf("collect: %v", err)
	}
	slices.Sort(got)
	return got
}

// TestRekeyInstrument re-keys a treasury bond series from its provisional
// ISIN to the registry's (D-214): its terms, exemption class, prices,
// cross-check prices, threshold and prices source move, rows the registry's
// ISIN already has win over the provisional one's, other instruments stay,
// and running it again changes nothing.
func TestRekeyInstrument(t *testing.T) {
	t.Parallel()
	store, pool := newStore(t)
	ctx := t.Context()
	from, to := instrument(t, "PLEDO0536000", "EDO0536"), instrument(t, "PL9999000018", "EDO0536")
	for _, q := range []string{
		`INSERT INTO valuation.bond_terms (isin, series, first_period_rate, margin, early_redemption_fee, source)
			VALUES ('PLEDO0536000', 'EDO0536', 0.06, 0.02, 2, 'bonds-old'), ('PLTEST000010', 'EDO1136', 0.06, 0.02, 2, 'bonds-other')`,
		`INSERT INTO valuation.exemption_classes (instrument, class) VALUES ('PLEDO0536000', 'savings')`,
		`INSERT INTO valuation.prices (instrument, date, close, currency, source) VALUES
			('PLEDO0536000/EDO0536', '2026-05-04', 100, 'PLN', 'md'), ('PLEDO0536000/EDO0536', '2026-05-05', 101, 'PLN', 'md'),
			('PL9999000018/EDO0536', '2026-05-05', 102, 'PLN', 'md')`,
		`INSERT INTO valuation.cross_check_prices (instrument, date, close, currency, source) VALUES ('PLEDO0536000/EDO0536', '2026-05-04', 100, 'PLN', 'xc')`,
		`INSERT INTO valuation.price_check_thresholds (instrument, threshold) VALUES ('PLEDO0536000/EDO0536', 0.02)`,
		`INSERT INTO valuation.series_sources (kind, key, source) VALUES ('prices', 'PLEDO0536000/EDO0536', 'md')`,
	} {
		if _, err := pool.Exec(ctx, q); err != nil {
			t.Fatalf("setup: %v", err)
		}
	}
	want := []string{
		"bond_terms PL9999000018 bonds-old", "bond_terms PLTEST000010 bonds-other",
		"cross_check_prices PL9999000018/EDO0536 2026-05-04",
		"exemption_classes PL9999000018 savings",
		"price_check_thresholds PL9999000018/EDO0536 0.02",
		"prices PL9999000018/EDO0536 2026-05-04 100", "prices PL9999000018/EDO0536 2026-05-05 102",
		"series_sources PL9999000018/EDO0536 md",
	}
	for range 2 {
		if err := store.RekeyInstrument(ctx, from, to); err != nil {
			t.Fatalf("RekeyInstrument: %v", err)
		}
		if diff := cmp.Diff(want, rekeyRows(t, pool)); diff != "" {
			t.Errorf("rows after the re-key (-want +got):\n%s", diff)
		}
	}
	if _, err := pool.Exec(ctx, `INSERT INTO valuation.bond_terms (isin, series, first_period_rate, margin, early_redemption_fee, source)
		VALUES ('PLEDO0536000', 'EDO0536', 0.05, 0.01, 1, 'bonds-again')`); err != nil {
		t.Fatalf("setup: %v", err)
	}
	if err := store.RekeyInstrument(ctx, from, to); err != nil {
		t.Fatalf("RekeyInstrument: %v", err)
	}
	if diff := cmp.Diff(want, rekeyRows(t, pool)); diff != "" {
		t.Errorf("rows after the registry's terms won (-want +got):\n%s", diff)
	}
}
