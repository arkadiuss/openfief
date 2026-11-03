package postgres_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/arkadiuss/openfief/internal/valuation"
)

// TestStoredTermsAndLevelsNameTheirInstance checks D-179's record in the
// store: bond series terms are stored with the connector instance that
// delivered them, terms delivered again by another instance replace them
// with that instance, and neither terms nor an index level can be stored
// without one.
func TestStoredTermsAndLevelsNameTheirInstance(t *testing.T) {
	t.Parallel()
	store, pool := newStore(t)
	ctx := t.Context()
	in := syntheticInputs(t)
	if err := store.SaveBondTerms(ctx, in.bond, &in.terms, "bonds-a", nil); err != nil {
		t.Fatalf("SaveBondTerms: %v", err)
	}
	sources := func() []string {
		return textRows(t, pool, `SELECT isin || ' ' || source FROM valuation.bond_terms ORDER BY isin`)
	}
	if diff := cmp.Diff([]string{in.bond.String() + " bonds-a"}, sources()); diff != "" {
		t.Errorf("bond terms' sources (-want +got):\n%s", diff)
	}
	if err := store.SaveBondTerms(ctx, in.bond, &in.terms, "bonds-b", nil); err != nil {
		t.Fatalf("SaveBondTerms again: %v", err)
	}
	if diff := cmp.Diff([]string{in.bond.String() + " bonds-b"}, sources()); diff != "" {
		t.Errorf("bond terms' sources after another instance (-want +got):\n%s", diff)
	}
	if err := store.SaveBondTerms(ctx, in.bond, &in.terms, "", nil); err == nil {
		t.Error("SaveBondTerms without a source: no error")
	}
	level := valuation.IndexLevel{Index: "FAKE20", Date: date(t, "2026-03-02"), Level: dec(t, "1500")}
	if err := store.SaveIndexLevels(ctx, []valuation.IndexLevel{level}); err == nil {
		t.Error("SaveIndexLevels without a source: no error")
	}
}

// TestStoredReadingsNameTheirInstance checks D-186's record in the store:
// each index reading is stored with the connector instance that delivered
// it, and none can be stored without one.
func TestStoredReadingsNameTheirInstance(t *testing.T) {
	t.Parallel()
	store, pool := newStore(t)
	ctx := t.Context()
	reading := valuation.IndexReading{Kind: valuation.CPI, Year: 2026, Month: 2, Value: dec(t, "0.049"), Source: "bonds-a"}
	if err := store.SaveIndexReadings(ctx, []valuation.IndexReading{reading}, nil); err != nil {
		t.Fatalf("SaveIndexReadings: %v", err)
	}
	got := textRows(t, pool, `SELECT kind || ' ' || year || '-' || month || ' ' || source FROM valuation.index_readings`)
	if diff := cmp.Diff([]string{"cpi 2026-2 bonds-a"}, got); diff != "" {
		t.Errorf("index readings' sources (-want +got):\n%s", diff)
	}
	reading.Source = ""
	if err := store.SaveIndexReadings(ctx, []valuation.IndexReading{reading}, nil); err == nil {
		t.Error("SaveIndexReadings without a source: no error")
	}
}
