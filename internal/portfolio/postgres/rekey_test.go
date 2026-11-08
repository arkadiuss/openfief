package postgres_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/arkadiuss/openfief/internal/ledger"
	"github.com/arkadiuss/openfief/internal/portfolio"
	"github.com/arkadiuss/openfief/internal/portfolio/postgres"
)

// TestRekeyInstrument re-keys the model's target of a provisional
// instrument to the registry's ISIN (D-214): renamed in place, or, when
// the model targets the registry's ISIN already, joined to it with both
// weights; running it again changes nothing.
func TestRekeyInstrument(t *testing.T) {
	t.Parallel()
	parse := func(isin, symbol string) ledger.Instrument {
		i, err := ledger.ParseISIN(isin)
		if err != nil {
			t.Fatalf("ParseISIN(%q): %v", isin, err)
		}
		return ledger.Instrument{ISIN: i, Symbol: symbol}
	}
	from, to := parse("PLEDO0536000", "EDO0536"), parse("PL9999000018", "EDO0536")
	tests := []struct {
		name        string
		model, want portfolio.Model
	}{
		{
			name: "renamed in place",
			model: portfolio.Model{Enabled: true, Targets: []portfolio.Target{
				target(t, portfolio.TargetInstrument, "IETEST000026/ETF1", "", "70"), target(t, portfolio.TargetInstrument, from.String(), "", "30"),
			}},
			want: portfolio.Model{Enabled: true, Targets: []portfolio.Target{
				target(t, portfolio.TargetInstrument, "IETEST000026/ETF1", "", "70"), target(t, portfolio.TargetInstrument, to.String(), "", "30"),
			}},
		},
		{
			name: "joined to the registry's ISIN",
			model: portfolio.Model{Targets: []portfolio.Target{
				target(t, portfolio.TargetInstrument, to.String(), "", "45.5"), target(t, portfolio.TargetInstrument, from.String(), "", "4.5"),
				target(t, portfolio.TargetIndex, "FAKE20", "PLN", "50"),
			}},
			want: portfolio.Model{Targets: []portfolio.Target{
				target(t, portfolio.TargetInstrument, to.String(), "", "50"), target(t, portfolio.TargetIndex, "FAKE20", "PLN", "50"),
			}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			store := postgres.New(newPool(t))
			if err := store.SaveModel(t.Context(), tt.model); err != nil {
				t.Fatalf("SaveModel: %v", err)
			}
			for range 2 {
				rekeyModel(t, store, from, to, tt.want)
			}
		})
	}
}

// rekeyModel re-keys the target of from in store's model, and fails t
// unless the model is want then.
func rekeyModel(t *testing.T, store *postgres.Store, from, to ledger.Instrument, want portfolio.Model) {
	t.Helper()
	if err := store.RekeyInstrument(t.Context(), from, to); err != nil {
		t.Fatalf("RekeyInstrument: %v", err)
	}
	got, err := store.Model(t.Context())
	if err != nil {
		t.Fatalf("Model: %v", err)
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("model after the re-key (-want +got):\n%s", diff)
	}
}
