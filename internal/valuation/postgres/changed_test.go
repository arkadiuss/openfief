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
