package demo

import (
	"time"

	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/valuation"
)

const (
	// ppm is one in millionths.
	ppm = 1_000_000
	// minCents keeps a price path from reaching zero.
	minCents = 100
	// centsExp is the decimal exponent of an amount in hundredths.
	centsExp = -2
)

// pricePath is one instrument's synthetic closes, in hundredths of its
// currency, on business days.
type pricePath map[kernel.Date]int64

// isBusinessDay reports whether d is Monday to Friday. Holidays are
// ignored: the demo's markets never close.
func isBusinessDay(d kernel.Date) bool {
	switch time.Date(d.Year(), d.Month(), d.Day(), 0, 0, 0, 0, time.UTC).Weekday() {
	case time.Saturday, time.Sunday:
		return false
	case time.Monday, time.Tuesday, time.Wednesday, time.Thursday, time.Friday:
		return true
	}
	return true
}

// walk returns in's closes on every business day from first to last: a
// random walk with drift, in integer arithmetic so that every close is an
// exact number of hundredths. Each day's return is driftPPM plus a uniform
// draw in [-volPPM, volPPM], in millionths.
func walk(in *Instrument, r *rng, first, last kernel.Date) pricePath {
	path := pricePath{}
	cents := in.startCents
	for d := first; !d.After(last); d = d.AddDays(1) {
		if !isBusinessDay(d) {
			continue
		}
		path[d] = cents
		step := in.driftPPM + r.between(-in.volPPM, in.volPPM)
		cents = max(cents+cents*step/ppm, minCents)
	}
	return path
}

// cents returns an amount in hundredths as a decimal.
func cents(n int64) kernel.Decimal { return kernel.NewDecimal(n, centsExp) }

// prices returns path as end-of-day prices of in (D-21), in date order from
// first to last.
func prices(in *Instrument, path pricePath, first, last kernel.Date) []valuation.Price {
	var out []valuation.Price
	for d := first; !d.After(last); d = d.AddDays(1) {
		if c, ok := path[d]; ok {
			out = append(out, valuation.Price{
				Instrument: in.Instrument.String(), Date: d, Close: kernel.NewMoney(cents(c), in.Currency), Source: PriceSource,
			})
		}
	}
	return out
}
