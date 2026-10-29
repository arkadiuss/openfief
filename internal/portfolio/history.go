package portfolio

import (
	"fmt"
	"slices"

	"github.com/arkadiuss/openfief/internal/kernel"
)

// targetHistory returns t's value in in's reporting currency per day, from
// its closes in in.Closes and, for a foreign currency, in.Rates, up to
// in.Through.
func targetHistory(t *Target, in *BacktestInput) (TargetHistory, error) {
	h := TargetHistory{Target: *t}
	closes, err := sortedCloses(t, in.Closes[t.Key()], in.Through)
	if err != nil {
		return TargetHistory{}, err
	}
	if len(closes) == 0 {
		h.Missing = MissingCloses
		return h, nil
	}
	currency := closes[0].Value.Currency()
	var rates []Rate
	if !currency.Equal(in.Reporting) {
		if rates, err = sortedRates(in.Rates, currency, in.Through); err != nil {
			return TargetHistory{}, err
		}
	}
	first, last := closes[0].Date, closes[len(closes)-1].Date
	if !currency.Equal(in.Reporting) {
		if len(rates) == 0 || rates[0].Date.After(last) {
			h.Missing = MissingRates
			return h, nil
		}
		first = laterDate(first, rates[0].Date)
	}
	h.Days = walk(closes, rates, first, last)
	return h, nil
}

// walk returns, for each day from first to last, the latest close times
// the latest rate on or before it, or the close alone when rates is nil.
// Both have one dated on or before first.
func walk(closes []Close, rates []Rate, first, last kernel.Date) []Point {
	days := make([]Point, 0, first.DaysUntil(last)+1)
	c, r := 0, 0
	for d := first; !d.After(last); d = d.AddDays(1) {
		for c+1 < len(closes) && !closes[c+1].Date.After(d) {
			c++
		}
		value := closes[c].Value.Amount()
		if rates != nil {
			for r+1 < len(rates) && !rates[r+1].Date.After(d) {
				r++
			}
			value = value.Mul(rates[r].Rate)
		}
		days = append(days, Point{Date: d, Value: value})
	}
	return days
}

// sortedCloses returns t's closes dated on or before through in date
// order, checking that each is positive, dated once, and in one currency:
// an index's own when t is an index.
func sortedCloses(t *Target, closes []Close, through kernel.Date) ([]Close, error) {
	out := make([]Close, 0, len(closes))
	for _, c := range closes {
		if !c.Date.After(through) {
			out = append(out, c)
		}
	}
	slices.SortFunc(out, func(a, b Close) int { return a.Date.Compare(b.Date) })
	for i, c := range out {
		switch {
		case c.Value.Sign() <= 0:
			return nil, fmt.Errorf("%w: %s on %s: %s is not positive", ErrInvalidHistory, t.Key(), c.Date, c.Value)
		case i > 0 && out[i-1].Date.Equal(c.Date):
			return nil, fmt.Errorf("%w: %s: two closes on %s", ErrInvalidHistory, t.Key(), c.Date)
		case !c.Value.Currency().Equal(out[0].Value.Currency()):
			return nil, fmt.Errorf("%w: %s: closes in %s and %s", ErrInvalidHistory, t.Key(), out[0].Value.Currency(), c.Value.Currency())
		case t.Kind == TargetIndex && !c.Value.Currency().Equal(t.Currency):
			return nil, fmt.Errorf("%w: %s: a level in %s, not the index's %s", ErrInvalidHistory, t.Key(), c.Value.Currency(), t.Currency)
		}
	}
	return out, nil
}

// sortedRates returns the rates of currency dated on or before through in
// date order, checking that each is positive and dated once.
func sortedRates(rates []Rate, currency kernel.Currency, through kernel.Date) ([]Rate, error) {
	var out []Rate
	for _, r := range rates {
		if r.Currency.Equal(currency) && !r.Date.After(through) {
			out = append(out, r)
		}
	}
	slices.SortFunc(out, func(a, b Rate) int { return a.Date.Compare(b.Date) })
	for i, r := range out {
		if r.Rate.Sign() <= 0 {
			return nil, fmt.Errorf("%w: %s rate on %s: %s is not positive", ErrInvalidHistory, currency, r.Date, r.Rate)
		}
		if i > 0 && out[i-1].Date.Equal(r.Date) {
			return nil, fmt.Errorf("%w: two %s rates on %s", ErrInvalidHistory, currency, r.Date)
		}
	}
	return out, nil
}

// laterDate returns the later of a and b.
func laterDate(a, b kernel.Date) kernel.Date {
	if b.After(a) {
		return b
	}
	return a
}
