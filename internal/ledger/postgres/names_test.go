package postgres_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/arkadiuss/openfief/internal/ledger"
)

// TestInstrumentNames checks that the store keeps the first name a
// connector delivers for an ISIN and every owner's rename, in order, and
// that its rows are append-only (D-197).
func TestInstrumentNames(t *testing.T) {
	t.Parallel()
	store, pool := seededStore(t)
	ctx := t.Context()
	a, b := instrument(t, isinA, "TSTA").ISIN, instrument(t, isinB, "TSTB").ISIN
	appends := [][]ledger.InstrumentName{
		{{ISIN: a, Name: "Synthetic A", Source: "xtb-main"}, {ISIN: b, Name: "Synthetic B", Source: "ibkr-main"}},
		{{ISIN: a, Name: "SYNTHETIC A ETF", Source: "ibkr-main"}},
		{{ISIN: a, Name: "My A", ByOwner: true}},
	}
	for _, names := range appends {
		if err := store.AppendInstrumentNames(ctx, names); err != nil {
			t.Fatalf("AppendInstrumentNames: %v", err)
		}
	}
	got, err := store.InstrumentNames(ctx)
	if err != nil {
		t.Fatalf("InstrumentNames: %v", err)
	}
	want := []ledger.InstrumentName{appends[0][0], appends[0][1], appends[2][0]}
	if diff := cmp.Diff(want, got, cmp.Comparer(ledger.ISIN.Equal)); diff != "" {
		t.Errorf("names (-want +got):\n%s", diff)
	}
	for _, sql := range []string{
		"UPDATE ledger.instrument_names SET name = 'x'",
		"DELETE FROM ledger.instrument_names",
		"TRUNCATE ledger.instrument_names",
	} {
		if _, err := pool.Exec(ctx, sql); sqlState(err) != appendOnlyViolation {
			t.Errorf("%s: error %v, want SQLSTATE %s", sql, err, appendOnlyViolation)
		}
	}
}
