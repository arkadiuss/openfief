package views

import (
	"cmp"
	"errors"
	"fmt"
	"slices"

	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
	"github.com/arkadiuss/openfief/internal/valuation"
)

// ErrInvalidRange reports a date range that ends before it starts or spans
// more than MaxRangeDays.
var ErrInvalidRange = errors.New("invalid date range")

// ErrNotPLN reports a value said to be in PLN that is in another currency.
var ErrNotPLN = errors.New("value is not in PLN")

// MaxRangeDays is the longest date range, in days after its first date, that
// one valuation query may span: about ten years of daily values.
const MaxRangeDays = 3700

// HoldingValuation is a stored valuation of one holding in one account
// (D-36): its value on a date with the basis and source it was computed on,
// in the instrument's own currency and converted to PLN.
type HoldingValuation struct {
	Account    ledger.AccountID
	Instrument ledger.Instrument
	Date       kernel.Date
	Basis      valuation.Basis
	Source     string
	// Value is in the currency the holding is valued in, such as a price's
	// currency; it is never converted.
	Value kernel.Money
	// PLN is Value converted to PLN.
	PLN kernel.Money
}

// DateTotal is the sum of the valuations dated on one date: in PLN, and per
// original currency, in currency code order.
type DateTotal struct {
	Date       kernel.Date
	PLN        kernel.Money
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

// TotalsByDate sums vals per date, in date order: one total per date that has
// a valuation. A PLN value in another currency is an error.
func TotalsByDate(vals []HoldingValuation) ([]DateTotal, error) {
	byDate := map[kernel.Date][]kernel.Money{}
	pln := map[kernel.Date]kernel.Money{}
	for i := range vals {
		v := &vals[i]
		sum, err := addPLN(pln[v.Date], v.PLN)
		if err != nil {
			return nil, fmt.Errorf("total on %s: %w", v.Date, err)
		}
		pln[v.Date] = sum
		byDate[v.Date] = append(byDate[v.Date], v.Value)
	}
	out := make([]DateTotal, 0, len(pln))
	for d, sum := range pln {
		out = append(out, DateTotal{Date: d, PLN: sum, ByCurrency: sumPerCurrency(byDate[d])})
	}
	slices.SortFunc(out, func(a, b DateTotal) int { return a.Date.Compare(b.Date) })
	return out, nil
}

// addPLN adds m to sum, both in PLN; a zero sum is no PLN amount yet.
func addPLN(sum, m kernel.Money) (kernel.Money, error) {
	if !m.Currency().Equal(valuation.PLN()) {
		return kernel.Money{}, fmt.Errorf("%w: %s", ErrNotPLN, m.Currency())
	}
	if sum.Currency().IsZero() {
		return m, nil
	}
	out, err := sum.Add(m)
	if err != nil {
		return kernel.Money{}, fmt.Errorf("adding PLN: %w", err)
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
