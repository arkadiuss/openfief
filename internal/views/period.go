package views

import (
	"errors"
	"fmt"

	"github.com/arkadiuss/openfief/internal/kernel"
)

// ErrPeriodUnvalued reports a period the portfolio is not valued for: its
// last day is after the day every account is valued through, a day it needs
// has no valuation or no FX rate, or nothing is recorded in it.
var ErrPeriodUnvalued = errors.New("the portfolio is not valued for the period")

// percentPlaces is how many decimal places a change or a return keeps when
// it is given in percent, rounded half-even.
const percentPlaces int32 = 2

// percentScale turns a fraction into percent.
const percentScale = 100

// percentOf is part as a percentage of whole, to percentPlaces, half-even;
// ok is false when whole is not positive, since a change from nothing has no
// percentage.
func percentOf(part, whole kernel.Decimal) (kernel.Decimal, bool, error) {
	if whole.Sign() <= 0 {
		return kernel.Decimal{}, false, nil
	}
	p, err := part.Mul(kernel.NewDecimal(percentScale, 0)).Quo(whole, percentPlaces, kernel.RoundHalfEven)
	if err != nil {
		return kernel.Decimal{}, false, fmt.Errorf("percentage: %w", err)
	}
	return p, true, nil
}

// ValueChange is how one of the three values (D-77) changed over a period:
// the value at the end of the opening day and at the end of the last day,
// the difference, and the difference in percent of the start.
type ValueChange struct {
	Start, End, Change kernel.Money
	// Known is false when a valued holding or cash lacks this value at
	// either end, so no other field is set.
	Known bool
	// Percent is Change as a percentage of Start; HasPercent is false when
	// Start is not positive.
	Percent    kernel.Decimal
	HasPercent bool
}

// changeOf returns the change from start to end.
func changeOf(start, end kernel.Money) (ValueChange, error) {
	diff, err := end.Sub(start)
	if err != nil {
		return ValueChange{}, fmt.Errorf("change: %w", err)
	}
	out := ValueChange{Start: start, End: end, Change: diff, Known: true}
	if out.Percent, out.HasPercent, err = percentOf(diff.Amount(), start.Amount()); err != nil {
		return ValueChange{}, err
	}
	return out, nil
}

// changeOfKnown is changeOf, or the unknown change when a value is missing: a
// value lacking at either end has no change.
func changeOfKnown(missing bool, start, end kernel.Money) (ValueChange, error) {
	if missing {
		return ValueChange{}, nil
	}
	return changeOf(start, end)
}

// PeriodChange is how the portfolio's value changed over a period of days
// (D-61, D-77): the three values at the start and the end, the
// time-weighted return, and the deposits and withdrawals in the period.
type PeriodChange struct {
	// From and To are the period as asked, both inclusive; Opening is the day
	// whose end values are the start.
	From, To, Opening kernel.Date
	Gross             ValueChange
	// Net and Liquidation are not Known when a valued holding or cash lacks
	// that value on the opening day or on To.
	Net, Liquidation ValueChange
	// Return is the time-weighted return over the period in percent, from
	// the end of the opening day.
	Return kernel.Decimal
	// Deposits and Withdrawals are the money paid in and out on the days
	// after Opening through To, by day: a day's deposits and withdrawals net
	// against each other, so a day counts as one or the other. Both are
	// non-negative; NetDeposits is Deposits minus Withdrawals.
	Deposits, Withdrawals, NetDeposits kernel.Money
}

// PeriodInput is what a period's change is computed from: the totals on the
// opening day and on the last day, and the performance series from the
// opening day through the last (Performance).
type PeriodInput struct {
	From, To         kernel.Date
	Opening, Closing DateTotal
	// Days is the performance series, one day after another, the first dated
	// Opening.Date and the last Closing.Date.
	Days      []PerformanceDay
	Reporting kernel.Currency
}

// check returns ErrPeriodUnvalued unless the series runs from the opening
// day to the period's last, which the closing total is dated.
func (in *PeriodInput) check() error {
	n := len(in.Days)
	if n == 0 || !in.Days[0].Date.Equal(in.Opening.Date) || !in.Days[n-1].Date.Equal(in.Closing.Date) || !in.Closing.Date.Equal(in.To) {
		return fmt.Errorf("%w: %s to %s", ErrPeriodUnvalued, in.From, in.To)
	}
	return nil
}

// ChangeOver computes how the portfolio's value changed over the period of
// in: each of the three values at the end of the opening day and at the end
// of the last day with their difference and its percentage of the start;
// the time-weighted return, the series' last index over its first; and the
// deposits and withdrawals on the days after the opening day, which the
// start already holds the earlier ones of. A series that does not run from
// the opening day to the last is ErrPeriodUnvalued, so is a total that has
// no date of its own.
func ChangeOver(in *PeriodInput) (PeriodChange, error) {
	if err := in.check(); err != nil {
		return PeriodChange{}, err
	}
	out := PeriodChange{From: in.From, To: in.To, Opening: in.Opening.Date}
	var err error
	if out.Gross, err = changeOf(in.Opening.Reporting, in.Closing.Reporting); err != nil {
		return PeriodChange{}, fmt.Errorf("gross: %w", err)
	}
	if out.Net, err = changeOfKnown(in.Opening.NetMissing || in.Closing.NetMissing, in.Opening.Net, in.Closing.Net); err != nil {
		return PeriodChange{}, fmt.Errorf("net: %w", err)
	}
	liquidationMissing := in.Opening.LiquidationMissing || in.Closing.LiquidationMissing
	if out.Liquidation, err = changeOfKnown(liquidationMissing, in.Opening.Liquidation, in.Closing.Liquidation); err != nil {
		return PeriodChange{}, fmt.Errorf("liquidation: %w", err)
	}
	if out.Return, err = in.timeWeightedReturn(); err != nil {
		return PeriodChange{}, err
	}
	paid, err := flowsOver(in.Days[1:], in.Reporting)
	if err != nil {
		return PeriodChange{}, err
	}
	out.Deposits, out.Withdrawals, out.NetDeposits = paid.Deposits, paid.Withdrawals, paid.Net
	return out, nil
}

// timeWeightedReturn is the series' last index over its first, in percent.
func (in *PeriodInput) timeWeightedReturn() (kernel.Decimal, error) {
	first, last := in.Days[0].Index, in.Days[len(in.Days)-1].Index
	ret, ok, err := percentOf(last.Sub(first), first)
	if err != nil {
		return kernel.Decimal{}, fmt.Errorf("return: %w", err)
	}
	if !ok {
		return kernel.Decimal{}, fmt.Errorf("%w: the series starts at no index", ErrPeriodUnvalued)
	}
	return ret, nil
}

// paidFlows are the money paid in and out over some days, and the net.
type paidFlows struct {
	Deposits, Withdrawals, Net kernel.Money
}

// flowsOver returns the money paid in and out on days, by day, and the net.
func flowsOver(days []PerformanceDay, reporting kernel.Currency) (paidFlows, error) {
	var in, out []kernel.Money
	for i := range days {
		switch f := days[i].Flow; {
		case f.Sign() > 0:
			in = append(in, f)
		case f.Sign() < 0:
			out = append(out, f.Neg())
		}
	}
	var (
		paid paidFlows
		err  error
	)
	if paid.Deposits, err = sumIn(reporting, in...); err != nil {
		return paidFlows{}, fmt.Errorf("deposits: %w", err)
	}
	if paid.Withdrawals, err = sumIn(reporting, out...); err != nil {
		return paidFlows{}, fmt.Errorf("withdrawals: %w", err)
	}
	if paid.Net, err = paid.Deposits.Sub(paid.Withdrawals); err != nil {
		return paidFlows{}, fmt.Errorf("net deposits: %w", err)
	}
	return paid, nil
}
