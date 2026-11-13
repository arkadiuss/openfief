package portfolio_test

import (
	"errors"
	"testing"

	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/portfolio"
)

// driftModel is a model of two instruments and an index, weights summing to
// 100%.
func driftModel(t *testing.T) *portfolio.Model {
	t.Helper()
	return &portfolio.Model{Enabled: true, Targets: []portfolio.Target{
		{Kind: portfolio.TargetInstrument, Name: etfA, Weight: mustDec(t, "60")},
		{Kind: portfolio.TargetInstrument, Name: etfB, Weight: mustDec(t, "30")},
		{Kind: portfolio.TargetIndex, Name: index1, Currency: mustCurrency(t, "PLN"), Weight: mustDec(t, "10")},
	}}
}

// values parses the values held, by instrument.
func values(t *testing.T, held map[string]string) map[string]kernel.Decimal {
	t.Helper()
	out := map[string]kernel.Decimal{}
	for k, v := range held {
		out[k] = mustDec(t, v)
	}
	return out
}

// wantFigure is what a target's figure shows: its current weight and drift,
// both empty for a target that is not held.
type wantFigure struct{ current, drift string }

// checkFigure fails the test unless f is want.
func checkFigure(t *testing.T, i int, f *portfolio.Figure, want wantFigure) {
	t.Helper()
	if f.Held != (want.current != "") {
		t.Errorf("figure %d: Held = %v, want %v", i, f.Held, want.current != "")
	}
	if want.current == "" {
		return
	}
	if w := mustDec(t, want.current); !f.Current.Equal(w) {
		t.Errorf("figure %d: Current = %s, want %s", i, f.Current, w)
	}
	if w := mustDec(t, want.drift); !f.Drift.Equal(w) {
		t.Errorf("figure %d: Drift = %s, want %s", i, f.Drift, w)
	}
}

func TestDriftOf(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		held      map[string]string
		total     string
		want      []wantFigure
		wantOther string
	}{
		{
			name: "held above and below the targets", held: map[string]string{etfA: "6310", etfB: "2000"}, total: "10000",
			want: []wantFigure{{"63.1", "3.1"}, {"20", "-10"}, {}}, wantOther: "16.9",
		},
		{
			name: "an instrument not held is zero", held: map[string]string{etfA: "5000"}, total: "10000",
			want: []wantFigure{{"50", "-10"}, {"0", "-30"}, {}}, wantOther: "50",
		},
		{
			name: "weights round half-even to two places", held: map[string]string{etfA: "1", etfB: "1"}, total: "3",
			want: []wantFigure{{"33.33", "-26.67"}, {"33.33", "3.33"}, {}}, wantOther: "33.34",
		},
		{
			name: "an instrument outside the model counts as other", held: map[string]string{etfA: "6000", "XX0000000CC3/ETFC": "1000"}, total: "10000",
			want: []wantFigure{{"60", "0"}, {"0", "-30"}, {}}, wantOther: "40",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := portfolio.DriftOf(driftModel(t), values(t, tt.held), mustDec(t, tt.total))
			if err != nil {
				t.Fatalf("DriftOf: %v", err)
			}
			if len(got.Figures) != len(tt.want) {
				t.Fatalf("got %d figures, want %d", len(got.Figures), len(tt.want))
			}
			for i := range got.Figures {
				checkFigure(t, i, &got.Figures[i], tt.want[i])
			}
			if want := mustDec(t, tt.wantOther); !got.Other.Equal(want) {
				t.Errorf("Other = %s, want %s", got.Other, want)
			}
		})
	}
}

func TestDriftOfRefusals(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		held  map[string]string
		total string
		want  error
	}{
		{name: "a portfolio worth nothing", total: "0", want: portfolio.ErrNoPortfolioValue},
		{name: "a portfolio worth less than nothing", total: "-5", want: portfolio.ErrNoPortfolioValue},
		{name: "a negative holding", held: map[string]string{etfA: "-1"}, total: "100", want: portfolio.ErrInvalidModel},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := portfolio.DriftOf(driftModel(t), values(t, tt.held), mustDec(t, tt.total))
			if !errors.Is(err, tt.want) {
				t.Errorf("DriftOf error = %v, want %v", err, tt.want)
			}
		})
	}
}
