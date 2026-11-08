package pl_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/arkadiuss/openfief/internal/jurisdiction/pl"
)

// TestProvisionalPrefix builds the provisional ISIN's prefix of synthetic
// series codes: "PL", the code and "00" (D-214).
func TestProvisionalPrefix(t *testing.T) {
	t.Parallel()
	tests := []struct{ series, want string }{
		{series: "EDO0536", want: "PLEDO053600"},
		{series: "OTS0227", want: "PLOTS022700"},
	}
	for _, tt := range tests {
		t.Run(tt.series, func(t *testing.T) {
			t.Parallel()
			if got := pl.ProvisionalPrefix(series(t, tt.series)); got != tt.want {
				t.Errorf("ProvisionalPrefix(%s) = %q, want %q", tt.series, got, tt.want)
			}
		})
	}
}

// TestProvisionalSeries tells provisional ISINs from the registry's: a
// synthetic registry ISIN, a non-Polish one and text that only looks like
// a provisional ISIN are none (D-214).
func TestProvisionalSeries(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name, isin, symbol string
		want               string
		ok, provisional    bool
	}{
		{name: "an EDO series' provisional ISIN", isin: "PLEDO0536000", symbol: "EDO0536", want: "EDO0536", ok: true, provisional: true},
		{name: "provisional, but of another series than the symbol", isin: "PLEDO0536000", symbol: "EDO1136", want: "EDO0536", ok: true},
		{name: "a synthetic registry ISIN", isin: "PL9999000018", symbol: "EDO0536"},
		{name: "another country's code", isin: "DEEDO0536001", symbol: "EDO0536"},
		{name: "no filler after the series code", isin: "PLEDO0536110", symbol: "EDO0536"},
		{name: "an unknown bond type", isin: "PLXYZ0536000", symbol: "XYZ0536"},
		{name: "too short", isin: "PLEDO053600", symbol: "EDO0536"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, ok := pl.ProvisionalSeries(tt.isin)
			if ok != tt.ok {
				t.Fatalf("ProvisionalSeries(%q) ok = %t, want %t", tt.isin, ok, tt.ok)
			}
			if ok && got.String() != tt.want {
				t.Errorf("ProvisionalSeries(%q) = %s, want %s", tt.isin, got, tt.want)
			}
			if p := pl.IsProvisional(tt.isin, tt.symbol); p != tt.provisional {
				t.Errorf("IsProvisional(%q, %q) = %t, want %t", tt.isin, tt.symbol, p, tt.provisional)
			}
		})
	}
}

// TestPublishedISINOf looks series' registry ISINs up in synthetic
// published data, which may list a series' ISIN before its rates (D-214).
func TestPublishedISINOf(t *testing.T) {
	t.Parallel()
	p := pl.Published{
		Rates: []pl.PublishedRates{{Series: series(t, "EDO0536"), FirstPeriodRate: dec(t, "0.0611"), Margin: dec(t, "0.0222")}},
		ISINs: []pl.PublishedISIN{
			{Series: series(t, "EDO0536"), ISIN: "PL9999000018"}, {Series: series(t, "EDO1136"), ISIN: "PL9999000026"},
		},
	}
	tests := []struct {
		series, want string
		listed       bool
	}{
		{series: "EDO0536", want: "PL9999000018", listed: true},
		{series: "EDO1136", want: "PL9999000026", listed: true},
		{series: "COI1130"},
	}
	for _, tt := range tests {
		t.Run(tt.series, func(t *testing.T) {
			t.Parallel()
			got, err := p.ISINOf(series(t, tt.series))
			if (err == nil) != tt.listed {
				t.Fatalf("ISINOf(%s) error = %v, want listed %t", tt.series, err, tt.listed)
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("ISINOf(%s) (-want +got):\n%s", tt.series, diff)
			}
		})
	}
}
