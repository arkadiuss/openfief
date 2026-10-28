package pl_test

import (
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/arkadiuss/openfief/internal/jurisdiction/pl"
)

func TestPeriodRate(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		series string
		first  string
		margin string
		idx    map[int]string
		period int
		want   string
	}{
		{name: "period 1 is the fixed first-period rate", series: "EDO0334", first: "0.06", margin: "0.02", period: 1, want: "0.06"},
		{name: "EDO period 2 = CPI 4.9% + margin 2% = 6.9%", series: "EDO0334", first: "0.06", margin: "0.02", idx: map[int]string{2: "0.049"}, period: 2, want: "0.069"},
		{name: "deflation -1% counts as 0, rate = margin 2%", series: "EDO0334", first: "0.06", margin: "0.02", idx: map[int]string{2: "-0.01"}, period: 2, want: "0.02"},
		{name: "ROR period 2 = reference rate 5.75% + margin 0", series: "ROR0126", first: "0.0575", margin: "0", idx: map[int]string{2: "0.0575"}, period: 2, want: "0.0575"},
		{name: "DOR period 3 = reference 5% + margin 0.15%", series: "DOR0227", first: "0.059", margin: "0.0015", idx: map[int]string{3: "0.05"}, period: 3, want: "0.0515"},
		{name: "negative NBP reference rate counts as 0, rate = margin", series: "DOR0227", first: "0.059", margin: "0.0015", idx: map[int]string{3: "-0.001"}, period: 3, want: "0.0015"},
		{name: "TOS is fixed for its whole term", series: "TOS0227", first: "0.0595", margin: "0", period: 3, want: "0.0595"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			terms := pl.Terms{Series: series(t, tt.series), FirstPeriodRate: dec(t, tt.first), Margin: dec(t, tt.margin), EarlyRedemptionFee: plnMoney(t, "0")}
			start := date(t, "2025-03-15")
			got, err := terms.PeriodRate(indices(t, tt.idx), tt.period, start)
			if err != nil {
				t.Fatalf("PeriodRate: %v", err)
			}
			if diff := cmp.Diff(dec(t, tt.want), got); diff != "" {
				t.Errorf("(-want +got):\n%s", diff)
			}
		})
	}
}

func TestPeriodRateNeedsTheIndexReading(t *testing.T) {
	t.Parallel()
	terms := pl.Terms{Series: series(t, "EDO0334"), FirstPeriodRate: dec(t, "0.06"), Margin: dec(t, "0.02")}
	start := date(t, "2025-03-15")
	_, err := terms.PeriodRate(pl.Indices{3: dec(t, "0.03")}, 2, start)
	var mie *pl.MissingIndexError
	if !errors.As(err, &mie) {
		t.Fatalf("error = %v, want *MissingIndexError", err)
	}
	if diff := cmp.Diff(pl.MissingIndexError{Series: terms.Series, Period: 2, Start: start}, *mie); diff != "" {
		t.Errorf("MissingIndexError (-want +got):\n%s", diff)
	}
	if got, want := mie.Error(), "no index reading for EDO0334 interest period 2 starting 2025-03-15"; got != want {
		t.Errorf("Error() = %q, want %q", got, want)
	}
}

func TestNominal(t *testing.T) {
	t.Parallel()
	if diff := cmp.Diff(plnMoney(t, "100"), pl.Nominal()); diff != "" {
		t.Errorf("(-want +got):\n%s", diff)
	}
}
