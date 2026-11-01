package valuation_test

import (
	"testing"

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
