package postgres_test

import (
	"testing"

	"github.com/arkadiuss/openfief/internal/jurisdiction/pl"
	"github.com/arkadiuss/openfief/internal/valuation"
)

// TestSavePricesReportsOnlyWhatWasNewOrDifferent checks the count the price
// upsert gives back (D-211): a price that is new or that differs from the
// stored one counts, and one sent again unchanged, as a sync that re-sends the
// same prices does, counts for none, however its amount is written. The stored
// row is replaced by the different one.
func TestSavePricesReportsOnlyWhatWasNewOrDifferent(t *testing.T) {
	t.Parallel()
	store, pool := newStore(t)
	in := syntheticInputs(t)
	changed := in.prices[0]
	changed.Close = money(t, "999", "USD")
	sameAmount := in.prices[0]
	sameAmount.Close = money(t, "101.370", "USD")
	otherSource := in.prices[0]
	otherSource.Source = "other"
	steps := []struct {
		name   string
		prices []valuation.Price
		want   int
	}{
		{"both new", in.prices, 2},
		{"the same again", in.prices, 0},
		{"the same with the amount written another way", []valuation.Price{sameAmount}, 0},
		{"one differs, one is the same", []valuation.Price{changed, in.prices[1]}, 1},
		{"the one that differed, again", []valuation.Price{changed}, 0},
		{"another source's price of the same day", []valuation.Price{otherSource}, 1},
	}
	for _, s := range steps {
		got, err := store.SavePricesChanged(t.Context(), s.prices, nil)
		if err != nil || got != s.want {
			t.Errorf("%s: SavePricesChanged = %d, %v; want %d", s.name, got, err, s.want)
		}
	}
	if n := countRows(t, pool, "SELECT count(*) FROM valuation.prices WHERE close = 999"); n != 1 {
		t.Errorf("rows holding the changed close = %d, want the stored row replaced", n)
	}
}

// TestSaveFXRatesReportsOnlyWhatWasNewOrDifferent is the same for FX rates.
func TestSaveFXRatesReportsOnlyWhatWasNewOrDifferent(t *testing.T) {
	t.Parallel()
	store, _ := newStore(t)
	in := syntheticInputs(t)
	differs := in.rates[0]
	differs.Rate = dec(t, "4.5")
	steps := []struct {
		name  string
		rates []valuation.FXRate
		want  int
	}{
		{"new", in.rates, 1},
		{"the same again", in.rates, 0},
		{"a different rate", []valuation.FXRate{differs}, 1},
		{"that one again", []valuation.FXRate{differs}, 0},
	}
	for _, s := range steps {
		got, err := store.SaveFXRatesChanged(t.Context(), s.rates, nil)
		if err != nil || got != s.want {
			t.Errorf("%s: SaveFXRatesChanged = %d, %v; want %d", s.name, got, err, s.want)
		}
	}
}

// TestSaveIndexLevelsReportsOnlyWhatWasNewOrDifferent is the same for market
// index levels (D-211): a level that is new, or that differs in its value or in
// the instance that delivered it, counts; the same level sent again, however
// its value is written, counts for none, and writes nothing.
func TestSaveIndexLevelsReportsOnlyWhatWasNewOrDifferent(t *testing.T) {
	t.Parallel()
	store, pool := newStore(t)
	level := func(index, on, value, source string) valuation.IndexLevel {
		return valuation.IndexLevel{Index: index, Date: date(t, on), Level: dec(t, value), Source: source}
	}
	steps := []struct {
		name   string
		levels []valuation.IndexLevel
		want   int
	}{
		{"two new", []valuation.IndexLevel{level("FAKE20", "2025-01-02", "1500", "eodhd"), level("FAKE20", "2025-01-03", "1522.60", "eodhd")}, 2},
		{"the same again", []valuation.IndexLevel{level("FAKE20", "2025-01-02", "1500", "eodhd"), level("FAKE20", "2025-01-03", "1522.60", "eodhd")}, 0},
		{"the same with the value written another way", []valuation.IndexLevel{level("FAKE20", "2025-01-02", "1500.00", "eodhd")}, 0},
		{"a different value, and one that is the same", []valuation.IndexLevel{level("FAKE20", "2025-01-02", "1501", "eodhd"), level("FAKE20", "2025-01-03", "1522.6", "eodhd")}, 1},
		{"the same value from another instance", []valuation.IndexLevel{level("FAKE20", "2025-01-02", "1501", "eodhd-2")}, 1},
		{"that again", []valuation.IndexLevel{level("FAKE20", "2025-01-02", "1501", "eodhd-2")}, 0},
		{"another index of the same day", []valuation.IndexLevel{level("OTHER", "2025-01-02", "99.5", "eodhd")}, 1},
		{"nothing sent", nil, 0},
	}
	for _, s := range steps {
		got, err := store.SaveIndexLevelsChanged(t.Context(), s.levels)
		if err != nil || got != s.want {
			t.Errorf("%s: SaveIndexLevelsChanged = %d, %v; want %d", s.name, got, err, s.want)
		}
	}
	if n := countRows(t, pool, "SELECT count(*) FROM valuation.index_levels"); n != 3 {
		t.Errorf("rows = %d, want one per index and day", n)
	}
}

// TestSaveIndexReadingsReportsOnlyWhatWasNewOrDifferent is the same for index
// readings, which also queue the revaluations they cause.
func TestSaveIndexReadingsReportsOnlyWhatWasNewOrDifferent(t *testing.T) {
	t.Parallel()
	store, _ := newStore(t)
	in := syntheticInputs(t)
	differs := in.readings[0]
	differs.Value = dec(t, "0.031")
	otherSource := in.readings[0]
	otherSource.Source = "other"
	steps := []struct {
		name     string
		readings []valuation.IndexReading
		want     int
	}{
		{"new", in.readings, 1},
		{"the same again", in.readings, 0},
		{"a different value", []valuation.IndexReading{differs}, 1},
		{"that one again", []valuation.IndexReading{differs}, 0},
		{"the first value from another instance", []valuation.IndexReading{otherSource}, 1},
	}
	for _, s := range steps {
		got, err := store.SaveIndexReadingsChanged(t.Context(), s.readings, nil)
		if err != nil || got != s.want {
			t.Errorf("%s: SaveIndexReadingsChanged = %d, %v; want %d", s.name, got, err, s.want)
		}
	}
}

// TestJobsFailedAtLeastCountsTheJobsThatKeepFailing checks the count the audit
// outbox's attention item reads (D-211, D-196): jobs of a kind that have failed
// at least some attempts and are still waiting count, a job that has not
// failed that often, one of another kind and one that finished count for none.
func TestJobsFailedAtLeastCountsTheJobsThatKeepFailing(t *testing.T) {
	t.Parallel()
	store, pool := newStore(t)
	ctx := t.Context()
	insert := func(kind, state string, failures int) {
		t.Helper()
		_, err := pool.Exec(ctx, `
			INSERT INTO river.river_job (state, kind, args, queue, max_attempts, attempt, errors, finalized_at)
			SELECT $1::river.river_job_state, $2, '{}', 'audit', 32767, $3::int,
			       ARRAY(SELECT '{"at":"2026-03-02T10:00:00Z","attempt":1,"error":"x","trace":""}'::jsonb FROM generate_series(1, $3::int)),
			       CASE WHEN $1 IN ('completed', 'discarded', 'cancelled') THEN now() END`, state, kind, failures)
		if err != nil {
			t.Fatalf("inserting a %s job of %s: %v", state, kind, err)
		}
	}
	const kind = "identity.audit_entry"
	insert(kind, "retryable", 5)
	insert(kind, "retryable", 9)
	insert(kind, "available", 5)
	insert(kind, "retryable", 4)
	insert(kind, "completed", 6)
	insert(kind, "discarded", 7)
	insert("valuation.revalue_account", "retryable", 8)
	for failures, want := range map[int]int{1: 4, 5: 3, 6: 1, 9: 1, 10: 0} {
		got, err := store.JobsFailedAtLeast(ctx, kind, failures)
		if err != nil || got != want {
			t.Errorf("JobsFailedAtLeast(%d) = %d, %v; want %d", failures, got, err, want)
		}
	}
}

// TestSaveBondTermsReportsOnlyWhatWasNewOrDifferent is the same for bond terms:
// the same terms from the same source, sent again, change nothing; other terms,
// or the same from another instance, do.
func TestSaveBondTermsReportsOnlyWhatWasNewOrDifferent(t *testing.T) {
	t.Parallel()
	store, _ := newStore(t)
	in := syntheticInputs(t)
	later := in.terms
	later.Margin = dec(t, "0.03")
	steps := []struct {
		name   string
		terms  *pl.Terms
		source string
		want   bool
	}{
		{"new", &in.terms, "treasury-bonds", true},
		{"the same again", &in.terms, "treasury-bonds", false},
		{"the same terms from another instance", &in.terms, "treasury-b", true},
		{"other terms", &later, "treasury-b", true},
		{"those again", &later, "treasury-b", false},
	}
	for _, s := range steps {
		got, err := store.SaveBondTermsChanged(t.Context(), in.bond, s.terms, s.source, nil)
		if err != nil || got != s.want {
			t.Errorf("%s: SaveBondTermsChanged = %v, %v; want %v", s.name, got, err, s.want)
		}
	}
}
