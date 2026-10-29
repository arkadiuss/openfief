package ingest

import (
	"cmp"
	"slices"

	"github.com/arkadiuss/openfief/internal/ledger"
)

// MarketDataScope is what a sync asks a market-data connector to price
// (D-93): the instruments the owner holds and the model portfolio's
// instrument targets, as the ledger writes them, and the model portfolio's
// market indices, by the name the owner gave each (D-96). A connector adds
// the instruments its own config lists to watch.
type MarketDataScope struct {
	Instruments []ledger.Instrument
	Indices     []string
}

// NewMarketDataScope returns the scope of instruments and indices: each
// once, instruments in ISIN then venue symbol order, indices by name. An
// empty index name is left out.
func NewMarketDataScope(instruments []ledger.Instrument, indices []string) MarketDataScope {
	var out MarketDataScope
	if len(instruments) > 0 {
		out.Instruments = slices.Clone(instruments)
		slices.SortFunc(out.Instruments, func(a, b ledger.Instrument) int {
			return cmp.Or(cmp.Compare(a.ISIN.String(), b.ISIN.String()), cmp.Compare(a.Symbol, b.Symbol))
		})
		out.Instruments = slices.Compact(out.Instruments)
	}
	for _, ix := range indices {
		if ix != "" {
			out.Indices = append(out.Indices, ix)
		}
	}
	slices.Sort(out.Indices)
	out.Indices = slices.Compact(out.Indices)
	return out
}
