package demo

import (
	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/valuation"
)

// flaggedMarkPermille is how far above the demo's price the one flagged
// cross-check price lies, in thousandths: 3%, beyond the 1% threshold
// (D-110).
const flaggedMarkPermille = 1030

// permille is one in thousandths.
const permille = 1000

// crossCheckPrices returns IBKR's synthetic mark prices (D-22): for every
// monthly statement, each instrument held at its end marked at the demo's
// latest close on or before the statement's last day, so a check of a
// month that ends on a weekend compares with Friday's close (D-111). The
// latest statement's first instrument is marked 3% above it, so the demo
// shows a flagged price check (docs/dev/demo.md). They come from the IBKR
// institution and value nothing.
func crossCheckPrices(m *market, h *ibkrHistory) []valuation.CrossCheckPrice {
	var out []valuation.CrossCheckPrice
	for i := range h.statements {
		st := &h.statements[i]
		for j, in := range sortedInstruments(st.end.units) {
			c, ok := closeOnOrBefore(m.paths[in], st.to)
			if !ok {
				continue
			}
			if i == len(h.statements)-1 && j == 0 {
				c = c * flaggedMarkPermille / permille
			}
			out = append(out, valuation.CrossCheckPrice{
				Instrument: in.Instrument.String(), Date: st.to, Close: kernel.NewMoney(cents(c), in.Currency), Source: IBKRInstitution,
			})
		}
	}
	return out
}

// closeOnOrBefore returns path's latest close dated on or before day,
// looking back a week at most.
func closeOnOrBefore(path pricePath, day kernel.Date) (int64, bool) {
	const week = 7
	for d := day; !d.Before(day.AddDays(-week)); d = d.AddDays(-1) {
		if c, ok := path[d]; ok {
			return c, true
		}
	}
	return 0, false
}
