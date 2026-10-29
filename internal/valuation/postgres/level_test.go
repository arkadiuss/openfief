package postgres_test

import (
	"testing"

	"github.com/arkadiuss/openfief/internal/valuation"
)

func TestIndexLevelsRoundTrip(t *testing.T) {
	t.Parallel()
	store, pool := newStore(t)
	ctx := t.Context()
	level := func(index, on, value, source string) valuation.IndexLevel {
		return valuation.IndexLevel{Index: index, Date: date(t, on), Level: dec(t, value), Source: source}
	}
	first := []valuation.IndexLevel{
		level("FAKE20", "2025-01-03", "1522.60", "eodhd"), level("FAKE20", "2025-01-02", "1500", "eodhd"),
		level("OTHER", "2025-01-02", "99.5", "eodhd"),
	}
	if err := store.SaveIndexLevels(ctx, first); err != nil {
		t.Fatalf("SaveIndexLevels: %v", err)
	}
	// Saving a level again replaces it, as a retried batch does.
	if err := store.SaveIndexLevels(ctx, []valuation.IndexLevel{level("FAKE20", "2025-01-03", "1530.1", "eodhd-2")}); err != nil {
		t.Fatalf("SaveIndexLevels again: %v", err)
	}
	got, err := store.IndexLevels(ctx, "FAKE20", date(t, "2025-01-31"))
	want := []valuation.IndexLevel{level("FAKE20", "2025-01-02", "1500", "eodhd"), level("FAKE20", "2025-01-03", "1530.1", "eodhd-2")}
	check(t, "levels of FAKE20", want, got, err)
	got, err = store.IndexLevels(ctx, "FAKE20", date(t, "2025-01-02"))
	check(t, "levels through 2025-01-02", want[:1], got, err)
	if got, err = store.IndexLevels(ctx, "NONE", date(t, "2025-01-31")); err != nil || len(got) != 0 {
		t.Errorf("levels of an unknown index = %v, %v; want none", got, err)
	}
	checkPending(t, pool, nil, nil)
	if err := store.SaveIndexLevels(ctx, []valuation.IndexLevel{level("FAKE20", "2025-01-06", "0", "eodhd")}); err == nil {
		t.Error("saving a zero level succeeded; the table check must refuse it")
	}
}
