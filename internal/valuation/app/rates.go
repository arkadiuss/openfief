// Package app holds valuation's use cases and the ports they consume.
// Adapters such as internal/valuation/nbp implement the ports; only cmd
// wires them together.
package app

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"slices"

	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/valuation"
)

// ErrInvalidRange reports a date range that ends before it starts, or a
// span limit below one day.
var ErrInvalidRange = errors.New("invalid date range")

// RateProvider is the swappable FX rate provider (D-20): it returns the
// rates to PLN it published for each business day in [from, to], both
// inclusive. NBP table A is the default. A range with no published table,
// such as a weekend or a holiday, is an empty result, not an error. A
// provider may cap how many days one call covers; FetchRates splits a longer
// range for it.
type RateProvider interface {
	Rates(ctx context.Context, from, to kernel.Date) ([]valuation.FXRate, error)
}

// DateRange is the days from From to To, both inclusive.
type DateRange struct {
	From kernel.Date
	To   kernel.Date
}

// SplitRange cuts the days of [from, to], both inclusive, into consecutive
// ranges of at most maxDays days each, in date order: every range but the
// last is exactly maxDays days long. The result depends only on the
// arguments, so the same range always produces the same requests.
func SplitRange(from, to kernel.Date, maxDays int) ([]DateRange, error) {
	if maxDays < 1 {
		return nil, fmt.Errorf("%w: span limit %d days", ErrInvalidRange, maxDays)
	}
	if to.Before(from) {
		return nil, fmt.Errorf("%w: %s is after %s", ErrInvalidRange, from, to)
	}
	var out []DateRange
	for start := from; !start.After(to); start = start.AddDays(maxDays) {
		end := start.AddDays(maxDays - 1)
		if end.After(to) {
			end = to
		}
		out = append(out, DateRange{From: start, To: end})
	}
	return out, nil
}

// FetchRates returns every rate p published in [from, to], both inclusive,
// asking for at most maxDays days per call (SplitRange), and sorted by date,
// then currency. It stops at the first failing call and returns its error.
func FetchRates(ctx context.Context, p RateProvider, from, to kernel.Date, maxDays int) ([]valuation.FXRate, error) {
	ranges, err := SplitRange(from, to, maxDays)
	if err != nil {
		return nil, err
	}
	var rates []valuation.FXRate
	for _, r := range ranges {
		got, err := p.Rates(ctx, r.From, r.To)
		if err != nil {
			return nil, fmt.Errorf("fetch FX rates %s to %s: %w", r.From, r.To, err)
		}
		rates = append(rates, got...)
	}
	slices.SortStableFunc(rates, compareRates)
	return rates, nil
}

func compareRates(a, b valuation.FXRate) int {
	if c := a.Date.Compare(b.Date); c != 0 {
		return c
	}
	return cmp.Compare(a.Currency.String(), b.Currency.String())
}
