package ingest_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/arkadiuss/openfief/internal/ingest"
	"github.com/arkadiuss/openfief/internal/ledger"
)

// scopeInstrument is a synthetic instrument.
func scopeInstrument(t *testing.T, isin, symbol string) ledger.Instrument {
	t.Helper()
	i, err := ledger.ParseISIN(isin)
	if err != nil {
		t.Fatalf("ParseISIN(%q): %v", isin, err)
	}
	return ledger.Instrument{ISIN: i, Symbol: symbol}
}

func TestNewMarketDataScope(t *testing.T) {
	t.Parallel()
	a := scopeInstrument(t, "IETEST000026", "TSTB")
	aOther := scopeInstrument(t, "IETEST000026", "TSTB2")
	b := scopeInstrument(t, "XX00000SYNT3", "SYNT")
	tests := []struct {
		name        string
		instruments []ledger.Instrument
		indices     []string
		want        ingest.MarketDataScope
	}{
		{name: "nothing held and no model"},
		{
			name:        "each instrument once, in ISIN then symbol order",
			instruments: []ledger.Instrument{b, aOther, a, b},
			want:        ingest.MarketDataScope{Instruments: []ledger.Instrument{a, aOther, b}},
		},
		{
			name:    "each index once, by name, none empty",
			indices: []string{"WIG20", "", "FAKE20", "WIG20"},
			want:    ingest.MarketDataScope{Indices: []string{"FAKE20", "WIG20"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := ingest.NewMarketDataScope(tt.instruments, tt.indices)
			if diff := cmp.Diff(tt.want, got, cmp.Comparer(func(x, y ledger.ISIN) bool { return x.Equal(y) })); diff != "" {
				t.Errorf("NewMarketDataScope (-want +got):\n%s", diff)
			}
		})
	}
}
