package valuation_test

import (
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/valuation"
)

func TestSeries(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		series     valuation.Series
		wantString string
		wantErr    error
	}{
		{name: "an instrument's prices", series: valuation.PriceSeries("XX00000SYNT3/SYNT"), wantString: "prices/XX00000SYNT3/SYNT"},
		{name: "a currency's FX rates", series: valuation.FXSeries(cur(t, "USD")), wantString: "fx_rates/USD"},
		{name: "CPI readings", series: valuation.IndexSeries(valuation.CPI), wantString: "index_readings/cpi"},
		{
			name: "the NBP reference rate's readings", series: valuation.IndexSeries(valuation.ReferenceRate),
			wantString: "index_readings/reference_rate",
		},
		{name: "PLN has no rate", series: valuation.FXSeries(valuation.PLN()), wantString: "fx_rates/PLN", wantErr: valuation.ErrInvalidSeries},
		{
			name: "not a currency", series: valuation.Series{Kind: valuation.KindFXRates, Key: "dollars"},
			wantString: "fx_rates/dollars", wantErr: valuation.ErrInvalidSeries,
		},
		{
			name: "not an index kind", series: valuation.Series{Kind: valuation.KindIndexReadings, Key: "wig20"},
			wantString: "index_readings/wig20", wantErr: valuation.ErrInvalidSeries,
		},
		{name: "no key", series: valuation.PriceSeries(""), wantString: "prices/", wantErr: valuation.ErrInvalidSeries},
		{name: "an instrument with a venue", series: valuation.PriceSeries("XX00000SYNT3/SYNT"), wantString: "prices/XX00000SYNT3/SYNT"},
		{
			name: "a key too long", series: valuation.PriceSeries(strings.Repeat("x", 201)),
			wantString: "prices/" + strings.Repeat("x", 201), wantErr: valuation.ErrInvalidSeries,
		},
		{
			name: "an unknown kind", series: valuation.Series{Kind: 9, Key: "x"},
			wantString: "MarketDataKind(9)/x", wantErr: valuation.ErrInvalidSeries,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := tt.series.String(); got != tt.wantString {
				t.Errorf("String = %q, want %q", got, tt.wantString)
			}
			if diff := cmp.Diff(tt.wantErr, tt.series.Validate(), cmpopts.EquateErrors()); diff != "" {
				t.Errorf("Validate (-want +got):\n%s", diff)
			}
		})
	}
}

func TestChooseSource(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		rank      []string
		available []string
		want      string
		wantOK    bool
	}{
		{
			name: "the highest-ranked source with a value", rank: []string{"eodhd", "vendor-b"},
			available: []string{"vendor-b", "eodhd"}, want: "eodhd", wantOK: true,
		},
		{
			name: "a higher-ranked source without a value is passed over", rank: []string{"eodhd", "vendor-b"},
			available: []string{"vendor-b"}, want: "vendor-b", wantOK: true,
		},
		{
			name: "a ranked source before an unranked one", rank: []string{"vendor-b"},
			available: []string{"aa-vendor", "vendor-b"}, want: "vendor-b", wantOK: true,
		},
		{name: "unranked sources by name", available: []string{"zz", "aa"}, want: "aa", wantOK: true},
		{name: "no source has a value", rank: []string{"eodhd"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, ok := valuation.ChooseSource(tt.rank, tt.available)
			if got != tt.want || ok != tt.wantOK {
				t.Errorf("ChooseSource = %q, %v; want %q, %v", got, ok, tt.want, tt.wantOK)
			}
		})
	}
}

// sourcedPrice is a synthetic close of a synthetic instrument on day from
// source.
func sourcedPrice(t *testing.T, day, source, amount string) valuation.Price {
	t.Helper()
	return valuation.Price{Instrument: "XX00000SYNT3/SYNT", Date: date(t, day), Close: money(t, amount, "EUR"), Source: source}
}

func TestPricesBySource(t *testing.T) {
	t.Parallel()
	eodhd2 := sourcedPrice(t, "2026-03-02", "eodhd", "100")
	other2 := sourcedPrice(t, "2026-03-02", "vendor-b", "101")
	other3 := sourcedPrice(t, "2026-03-03", "vendor-b", "102")
	eodhd4 := sourcedPrice(t, "2026-03-04", "eodhd", "104")
	all := []valuation.Price{eodhd2, other2, other3, eodhd4}
	tests := []struct {
		name       string
		own        string
		wantOwn    []valuation.Price
		wantOthers []valuation.Price
	}{
		{
			name: "only the source's own prices, the other source's on every date", own: "eodhd",
			wantOwn: []valuation.Price{eodhd2, eodhd4}, wantOthers: []valuation.Price{other2, other3},
		},
		{
			name: "the owner's choice of another source", own: "vendor-b",
			wantOwn: []valuation.Price{other2, other3}, wantOthers: []valuation.Price{eodhd2, eodhd4},
		},
		{name: "a source without prices", own: "gone", wantOthers: all},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := valuation.PricesBySource(all, tt.own)
			if diff := cmp.Diff(tt.wantOwn, got.Own); diff != "" {
				t.Errorf("own (-want +got):\n%s", diff)
			}
			if diff := cmp.Diff(tt.wantOthers, got.Others); diff != "" {
				t.Errorf("others (-want +got):\n%s", diff)
			}
		})
	}
}

// TestMissingDayCarriesForward shows the one-source rule end to end in the
// domain: a day the source lacks takes its own latest earlier price, never
// the other source's price of that day.
func TestMissingDayCarriesForward(t *testing.T) {
	t.Parallel()
	eodhd2 := sourcedPrice(t, "2026-03-02", "eodhd", "100")
	other3 := sourcedPrice(t, "2026-03-03", "vendor-b", "102")
	own := valuation.PricesBySource([]valuation.Price{eodhd2, other3}, "eodhd").Own
	tests := []struct {
		name string
		day  string
		want valuation.Price
	}{
		{name: "a day with the source's price", day: "2026-03-02", want: eodhd2},
		{name: "a day only another source priced takes the latest earlier own price", day: "2026-03-03", want: eodhd2},
		{name: "a later day without any price", day: "2026-03-05", want: eodhd2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := valuation.PriceOnOrBefore(own, eodhd2.Instrument, date(t, tt.day))
			if err != nil {
				t.Fatalf("PriceOnOrBefore: %v", err)
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("price (-want +got):\n%s", diff)
			}
		})
	}
}

func TestFXRatesAndReadingsBySource(t *testing.T) {
	t.Parallel()
	nbp, ecb := rate(t, "USD", "2026-03-02", "4.0000"), rate(t, "USD", "2026-03-03", "4.0100")
	ecb.Source = "ecb-fx"
	gotRates := valuation.FXRatesBySource([]valuation.FXRate{ecb, nbp}, "nbp-a")
	if diff := cmp.Diff(valuation.BySource[valuation.FXRate]{Own: []valuation.FXRate{nbp}, Others: []valuation.FXRate{ecb}}, gotRates); diff != "" {
		t.Errorf("FXRatesBySource (-want +got):\n%s", diff)
	}
	reading := func(source, value string) valuation.IndexReading {
		return valuation.IndexReading{Kind: valuation.CPI, Year: 2026, Month: time.January, Value: dec(t, value), Source: source}
	}
	bonds, gus := reading("bonds", "0.049"), reading("gus", "0.05")
	gotReadings := valuation.IndexReadingsBySource([]valuation.IndexReading{gus, bonds}, "bonds")
	want := valuation.BySource[valuation.IndexReading]{Own: []valuation.IndexReading{bonds}, Others: []valuation.IndexReading{gus}}
	if diff := cmp.Diff(want, gotReadings); diff != "" {
		t.Errorf("IndexReadingsBySource (-want +got):\n%s", diff)
	}
	if got := valuation.FXRatesBySource(nil, "nbp-a"); len(got.Own)+len(got.Others) != 0 {
		t.Errorf("FXRatesBySource of nothing = %v", got)
	}
}

func TestOtherSourceChecks(t *testing.T) {
	t.Parallel()
	other := sourcedPrice(t, "2026-03-02", "vendor-b", "101")
	want := []valuation.CrossCheckPrice{{Instrument: other.Instrument, Date: other.Date, Close: other.Close, Source: "vendor-b"}}
	if diff := cmp.Diff(want, valuation.OtherSourceChecks([]valuation.Price{other})); diff != "" {
		t.Errorf("OtherSourceChecks (-want +got):\n%s", diff)
	}
}

func TestGapThresholds(t *testing.T) {
	t.Parallel()
	owner := valuation.GapThresholds{Days: map[valuation.MarketDataKind]int{valuation.KindFXRates: 3}}
	tests := []struct {
		name string
		kind valuation.MarketDataKind
		want int
	}{
		{name: "prices by default", kind: valuation.KindPrices, want: 7},
		{name: "the owner's threshold", kind: valuation.KindFXRates, want: 3},
		{name: "index readings by default", kind: valuation.KindIndexReadings, want: 92},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := owner.For(tt.kind); got != tt.want {
				t.Errorf("For(%s) = %d, want %d", tt.kind, got, tt.want)
			}
		})
	}
}

func TestValidateGapThreshold(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		kind    valuation.MarketDataKind
		days    int
		wantErr error
	}{
		{name: "one day", kind: valuation.KindPrices, days: 1},
		{name: "ten years", kind: valuation.KindIndexReadings, days: 3660},
		{name: "zero", kind: valuation.KindPrices, wantErr: valuation.ErrInvalidGapThreshold},
		{name: "too long", kind: valuation.KindFXRates, days: 3661, wantErr: valuation.ErrInvalidGapThreshold},
		{name: "an unknown kind", days: 5, wantErr: valuation.ErrUnknownMarketDataKind},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if diff := cmp.Diff(tt.wantErr, valuation.ValidateGapThreshold(tt.kind, tt.days), cmpopts.EquateErrors()); diff != "" {
				t.Errorf("ValidateGapThreshold (-want +got):\n%s", diff)
			}
		})
	}
}

func TestGapOf(t *testing.T) {
	t.Parallel()
	series := valuation.PriceSeries("XX00000SYNT3/SYNT")
	tests := []struct {
		name       string
		latest     kernel.Date
		day        string
		threshold  int
		wantGap    bool
		wantDays   int
		wantDetail string
	}{
		{
			name: "a weekend is no gap", latest: date(t, "2026-03-06"), day: "2026-03-09", threshold: 7, wantDays: 3,
			wantDetail: "no value from eodhd since 2026-03-06",
		},
		{
			name: "exactly the threshold is no gap yet", latest: date(t, "2026-03-02"), day: "2026-03-09", threshold: 7, wantDays: 7,
			wantDetail: "no value from eodhd since 2026-03-02",
		},
		{
			name: "longer than the threshold", latest: date(t, "2026-03-01"), day: "2026-03-09", threshold: 7, wantGap: true, wantDays: 8,
			wantDetail: "no value from eodhd since 2026-03-01",
		},
		{
			name: "no value at all", day: "2026-03-09", threshold: 7, wantGap: true, wantDays: -1,
			wantDetail: "no value from eodhd yet",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			g, gap := valuation.GapOf(series, "eodhd", tt.latest, date(t, tt.day), tt.threshold)
			if gap != tt.wantGap || g.Days() != tt.wantDays || g.Detail() != tt.wantDetail {
				t.Errorf("GapOf = %v, %d days, %q; want %v, %d, %q", gap, g.Days(), g.Detail(), tt.wantGap, tt.wantDays, tt.wantDetail)
			}
		})
	}
}
