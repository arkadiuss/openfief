package views

import (
	"cmp"
	"errors"
	"fmt"
	"slices"

	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/valuation"
)

// ErrInvalidRange reports a date range that ends before it starts or spans
// more than MaxRangeDays.
var ErrInvalidRange = errors.New("invalid date range")

// ErrNotReportingCurrency reports a value said to be in the reporting
// currency (D-62) that is in another currency.
var ErrNotReportingCurrency = errors.New("value is not in the reporting currency")

// MaxRangeDays is the longest date range, in days after its first date, that
// one valuation query may span: about ten years of daily values.
const MaxRangeDays = 3700

// DateTotal is the sum of the valuations dated on one date: in the reporting
// currency, and per original currency, in currency code order.
type DateTotal struct {
	Date       kernel.Date
	Reporting  kernel.Money
	ByCurrency []kernel.Money
}

// ValidateRange reports whether from..to, both inclusive, is a date range one
// valuation query may ask for.
func ValidateRange(from, to kernel.Date) error {
	if to.Before(from) {
		return fmt.Errorf("%w: %s is after %s", ErrInvalidRange, from, to)
	}
	if from.DaysUntil(to) > MaxRangeDays {
		return fmt.Errorf("%w: %s to %s is more than %d days", ErrInvalidRange, from, to, MaxRangeDays)
	}
	return nil
}

// TotalsByDate sums the valued holding valuations in vals per date, in date
// order: one total per date that has one. A holding valuation that lacks an
// input counts nowhere. Each one's converted value must be in the reporting
// currency; stored valuations convert to PLN (D-20), so only PLN reporting
// adds up, and any other is an error wrapping ErrNotReportingCurrency.
func TotalsByDate(vals []valuation.HoldingValuation, reporting kernel.Currency) ([]DateTotal, error) {
	byDate := map[kernel.Date][]kernel.Money{}
	sums := map[kernel.Date]kernel.Money{}
	for i := range vals {
		v := &vals[i]
		if !v.Valued() {
			continue
		}
		sum, err := addIn(reporting, sums[v.Date], v.PLN)
		if err != nil {
			return nil, fmt.Errorf("total on %s: %w", v.Date, err)
		}
		sums[v.Date] = sum
		byDate[v.Date] = append(byDate[v.Date], v.Value)
	}
	out := make([]DateTotal, 0, len(sums))
	for d, sum := range sums {
		out = append(out, DateTotal{Date: d, Reporting: sum, ByCurrency: sumPerCurrency(byDate[d])})
	}
	slices.SortFunc(out, func(a, b DateTotal) int { return a.Date.Compare(b.Date) })
	return out, nil
}

// addIn adds m to sum, both in currency c; a sum with no currency is
// nothing yet.
func addIn(c kernel.Currency, sum, m kernel.Money) (kernel.Money, error) {
	if !m.Currency().Equal(c) {
		return kernel.Money{}, fmt.Errorf("%w %s: %s", ErrNotReportingCurrency, c, m.Currency())
	}
	if sum.Currency().IsZero() {
		return m, nil
	}
	out, err := sum.Add(m)
	if err != nil {
		return kernel.Money{}, fmt.Errorf("adding %s: %w", c, err)
	}
	return out, nil
}

// sumPerCurrency sums ms per currency, in currency code order.
func sumPerCurrency(ms []kernel.Money) []kernel.Money {
	sums := map[kernel.Currency]kernel.Decimal{}
	for _, m := range ms {
		sums[m.Currency()] = sums[m.Currency()].Add(m.Amount())
	}
	out := make([]kernel.Money, 0, len(sums))
	for c, d := range sums {
		out = append(out, kernel.NewMoney(d, c))
	}
	slices.SortFunc(out, func(a, b kernel.Money) int { return cmp.Compare(a.Currency().String(), b.Currency().String()) })
	return out
}
