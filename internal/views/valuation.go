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

// DateTotal is the sum of the valued holdings and the cash on one date
// (D-106): its three values (D-77) in the reporting currency, and every
// amount per original currency, in currency code order.
type DateTotal struct {
	Date kernel.Date
	// Reporting is the gross total, cash included.
	Reporting  kernel.Money
	ByCurrency []kernel.Money
	// Net and Liquidation are the totals of those values; NetMissing and
	// LiquidationMissing report that some valued holding or converted cash
	// lacks it, so the total is not given.
	Net, Liquidation               kernel.Money
	NetMissing, LiquidationMissing bool
	// Cash is the cash in the totals, in the reporting currency.
	Cash kernel.Money
	// CashUnconverted reports cash with no FX rate on or before the date,
	// in no total but ByCurrency.
	CashUnconverted bool
}

// dayTotal accumulates one date's total.
type dayTotal struct {
	DateTotal

	originals []kernel.Money
}

// TotalsByDate sums the valued holding valuations in vals and the cash in
// cash per date, in date order: one total per date that has either. A
// holding valuation that lacks an input counts nowhere; one that lacks its
// net or liquidation value leaves that total missing. Each value must be
// in the reporting currency; stored valuations convert to PLN (D-20), so
// only PLN reporting adds up, and any other is an error wrapping
// ErrNotReportingCurrency.
func TotalsByDate(vals []valuation.HoldingValuation, cash []DayCash, reporting kernel.Currency) ([]DateTotal, error) {
	days := map[kernel.Date]*dayTotal{}
	day := func(d kernel.Date) *dayTotal {
		if t, ok := days[d]; ok {
			return t
		}
		zero := kernel.NewMoney(kernel.Decimal{}, reporting)
		t := &dayTotal{DateTotal: DateTotal{Date: d, Reporting: zero, Net: zero, Liquidation: zero, Cash: zero}}
		days[d] = t
		return t
	}
	for i := range vals {
		if v := &vals[i]; v.Valued() {
			if err := day(v.Date).addHolding(v, reporting); err != nil {
				return nil, fmt.Errorf("total on %s: %w", v.Date, err)
			}
		}
	}
	for i := range cash {
		if err := day(cash[i].Date).addCash(&cash[i], reporting); err != nil {
			return nil, fmt.Errorf("total on %s: %w", cash[i].Date, err)
		}
	}
	out := make([]DateTotal, 0, len(days))
	for _, t := range days {
		t.ByCurrency = sumPerCurrency(t.originals)
		out = append(out, t.DateTotal)
	}
	slices.SortFunc(out, func(a, b DateTotal) int { return a.Date.Compare(b.Date) })
	return out, nil
}

// addHolding adds a valued holding valuation.
func (t *dayTotal) addHolding(v *valuation.HoldingValuation, reporting kernel.Currency) error {
	t.originals = append(t.originals, v.Value)
	return t.add(reporting, v.PLN, v.Net, v.Liquidation, !v.NetValued(), !v.LiquidationValued())
}

// addCash adds one account's cash in one currency.
func (t *dayTotal) addCash(c *DayCash, reporting kernel.Currency) error {
	t.originals = append(t.originals, c.Amount)
	if c.Unconverted {
		t.CashUnconverted = true
		return nil
	}
	var err error
	if t.Cash, err = addIn(reporting, t.Cash, c.Values.Gross); err != nil {
		return err
	}
	return t.add(reporting, c.Values.Gross, c.Values.Net, c.Values.Liquidation, false, c.Values.LiquidationUnknown)
}

// add adds the three values, each unless it is missing.
func (t *dayTotal) add(reporting kernel.Currency, gross, net, liquidation kernel.Money, noNet, noLiquidation bool) error {
	var err error
	if t.Reporting, err = addIn(reporting, t.Reporting, gross); err != nil {
		return err
	}
	if noNet {
		t.NetMissing = true
	} else if t.Net, err = addIn(reporting, t.Net, net); err != nil {
		return err
	}
	if noLiquidation {
		t.LiquidationMissing = true
	} else if t.Liquidation, err = addIn(reporting, t.Liquidation, liquidation); err != nil {
		return err
	}
	return nil
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
