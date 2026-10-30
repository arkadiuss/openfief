package valuation_test

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/arkadiuss/openfief/internal/valuation"
)

func TestMarketDataKindNames(t *testing.T) {
	t.Parallel()
	for _, k := range valuation.MarketDataKinds() {
		got, err := valuation.ParseMarketDataKind(k.String())
		if err != nil || got != k {
			t.Errorf("ParseMarketDataKind(%q) = %v, %v; want %v", k, got, err, k)
		}
	}
	if _, err := valuation.ParseMarketDataKind("levels"); err == nil {
		t.Error("ParseMarketDataKind(levels) succeeded, want ErrUnknownMarketDataKind")
	}
	if got := valuation.MarketDataKind(9).String(); got != "MarketDataKind(9)" {
		t.Errorf("String of an unknown kind = %q", got)
	}
}

func TestRankSources(t *testing.T) {
	t.Parallel()
	known := []valuation.Source{
		{Kind: valuation.KindPrices, Name: "ibkr-main", Institution: "ibkr"},
		{Kind: valuation.KindPrices, Name: "zz-prices", Institution: "vendor"},
		{Kind: valuation.KindPrices, Name: "prices-eu", Institution: "eodhd"},
		{Kind: valuation.KindFXRates, Name: "nbp-fx", Institution: "nbp"},
		{Kind: valuation.KindFXRates, Name: "ecb-fx", Institution: "ecb"},
		{Kind: valuation.KindIndexReadings, Name: "bonds", Institution: "treasury-bonds"},
		{Kind: valuation.KindIndexReadings, Name: "gus", Institution: "gus"},
	}
	tests := []struct {
		name   string
		kind   valuation.MarketDataKind
		placed []string
		want   []valuation.RankedSource
	}{
		{
			name: "prices by default: EODHD first, then by name",
			kind: valuation.KindPrices,
			want: []valuation.RankedSource{{Name: "prices-eu", Institution: "eodhd"}, {Name: "ibkr-main", Institution: "ibkr"}, {Name: "zz-prices", Institution: "vendor"}},
		},
		{
			name: "FX rates by default: NBP first", kind: valuation.KindFXRates,
			want: []valuation.RankedSource{{Name: "nbp-fx", Institution: "nbp"}, {Name: "ecb-fx", Institution: "ecb"}},
		},
		{
			name: "index readings by default: the treasury bond connector first", kind: valuation.KindIndexReadings,
			want: []valuation.RankedSource{{Name: "bonds", Institution: "treasury-bonds"}, {Name: "gus", Institution: "gus"}},
		},
		{
			name: "the owner's order first, the rest by default after it", kind: valuation.KindPrices,
			placed: []string{"zz-prices", "gone"},
			want: []valuation.RankedSource{
				{Name: "zz-prices", Institution: "vendor", Placed: true},
				{Name: "gone", Placed: true},
				{Name: "prices-eu", Institution: "eodhd"},
				{Name: "ibkr-main", Institution: "ibkr"},
			},
		},
		{name: "no source yet", kind: valuation.KindPrices + 5},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := valuation.RankSources(tt.kind, known, tt.placed)
			if diff := cmp.Diff(valuation.Ranking{Kind: tt.kind, Sources: tt.want}, got); diff != "" {
				t.Errorf("RankSources (-want +got):\n%s", diff)
			}
		})
	}
}

func TestValidateRanking(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		kind    valuation.MarketDataKind
		sources []string
		wantErr error
	}{
		{name: "a ranking", kind: valuation.KindPrices, sources: []string{"eodhd", "ibkr-main"}},
		{name: "empty: every source in its default place", kind: valuation.KindFXRates},
		{name: "a source twice", kind: valuation.KindPrices, sources: []string{"eodhd", "eodhd"}, wantErr: valuation.ErrInvalidRanking},
		{name: "an empty name", kind: valuation.KindPrices, sources: []string{""}, wantErr: valuation.ErrInvalidRanking},
		{name: "a name too long", kind: valuation.KindPrices, sources: []string{string(make([]byte, 65))}, wantErr: valuation.ErrInvalidRanking},
		{name: "an unknown kind", kind: 0, wantErr: valuation.ErrUnknownMarketDataKind},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if diff := cmp.Diff(tt.wantErr, valuation.ValidateRanking(tt.kind, tt.sources), cmpopts.EquateErrors()); diff != "" {
				t.Errorf("ValidateRanking (-want +got):\n%s", diff)
			}
		})
	}
}

// rankedPrice is a synthetic close of instrument on day from source.
func rankedPrice(t *testing.T, day, source, amount string) valuation.Price {
	t.Helper()
	return valuation.Price{Instrument: "XX00000SYNT3/SYNT", Date: date(t, day), Close: money(t, amount, "EUR"), Source: source}
}

func TestRankedPrices(t *testing.T) {
	t.Parallel()
	eodhd2 := rankedPrice(t, "2026-03-02", "eodhd", "100")
	ibkr2 := rankedPrice(t, "2026-03-02", "ibkr-main", "101")
	ibkr3 := rankedPrice(t, "2026-03-03", "ibkr-main", "102")
	other3 := rankedPrice(t, "2026-03-03", "other", "103")
	eodhd4 := rankedPrice(t, "2026-03-04", "eodhd", "104")
	all := []valuation.Price{ibkr2, eodhd2, ibkr3, other3, eodhd4}
	tests := []struct {
		name        string
		rank        []string
		wantWinners []valuation.Price
		wantLower   []valuation.Price
	}{
		{
			name: "the highest-ranked source with a value wins each date", rank: []string{"eodhd", "ibkr-main"},
			wantWinners: []valuation.Price{eodhd2, ibkr3, eodhd4}, wantLower: []valuation.Price{ibkr2, other3},
		},
		{
			name: "a new ranking moves the winners", rank: []string{"ibkr-main", "eodhd"},
			wantWinners: []valuation.Price{ibkr2, ibkr3, eodhd4}, wantLower: []valuation.Price{eodhd2, other3},
		},
		{
			name: "sources the ranking doesn't name rank after it, by name", rank: []string{"zz"},
			wantWinners: []valuation.Price{eodhd2, ibkr3, eodhd4}, wantLower: []valuation.Price{ibkr2, other3},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := valuation.RankedPrices(all, tt.rank)
			if diff := cmp.Diff(tt.wantWinners, got.Winners); diff != "" {
				t.Errorf("winners (-want +got):\n%s", diff)
			}
			if diff := cmp.Diff(tt.wantLower, got.Lower); diff != "" {
				t.Errorf("lower-ranked (-want +got):\n%s", diff)
			}
		})
	}
	checks := valuation.LowerRankedChecks([]valuation.Price{ibkr2})
	want := []valuation.CrossCheckPrice{{Instrument: ibkr2.Instrument, Date: ibkr2.Date, Close: ibkr2.Close, Source: "ibkr-main"}}
	if diff := cmp.Diff(want, checks); diff != "" {
		t.Errorf("LowerRankedChecks (-want +got):\n%s", diff)
	}
}

func TestRankedFXRatesAndReadings(t *testing.T) {
	t.Parallel()
	usd := cur(t, "USD")
	rate := func(day, source, value string) valuation.FXRate {
		return valuation.FXRate{Currency: usd, Date: date(t, day), Rate: dec(t, value), Source: source}
	}
	nbp, ecb, ecbOnly := rate("2026-03-02", "nbp-fx", "4.0000"), rate("2026-03-02", "ecb-fx", "4.0100"), rate("2026-03-03", "ecb-fx", "4.0200")
	gotRates := valuation.RankedFXRates([]valuation.FXRate{ecb, nbp, ecbOnly}, []string{"nbp-fx", "ecb-fx"})
	if diff := cmp.Diff([]valuation.FXRate{nbp, ecbOnly}, gotRates); diff != "" {
		t.Errorf("RankedFXRates (-want +got):\n%s", diff)
	}
	reading := func(source, value string) valuation.IndexReading {
		return valuation.IndexReading{Kind: valuation.CPI, Year: 2026, Month: time.January, Value: dec(t, value), Source: source}
	}
	bonds, gus := reading("bonds", "0.049"), reading("gus", "0.05")
	gotReadings := valuation.RankedIndexReadings([]valuation.IndexReading{gus, bonds}, []string{"bonds"})
	if diff := cmp.Diff([]valuation.IndexReading{bonds}, gotReadings); diff != "" {
		t.Errorf("RankedIndexReadings (-want +got):\n%s", diff)
	}
	if got := valuation.RankedFXRates(nil, nil); len(got) != 0 {
		t.Errorf("RankedFXRates of nothing = %v", got)
	}
}
