package portfolio

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/arkadiuss/openfief/internal/kernel"
)

// ErrInvalidHistory reports end-of-day history a backtest can't use: two
// closes or two rates of one thing on one date, closes of one target in
// two currencies, or a close, level or rate that is not positive.
var ErrInvalidHistory = errors.New("invalid end-of-day history")

const (
	// BacktestPlaces is how many decimal places the backtest's index, a
	// rebased series and a drawdown's depth keep, rounded with
	// BacktestRounding, as the time-weighted return does (views.TWRPlaces).
	BacktestPlaces int32 = 12
	// BacktestRounding is half-even, so rounding errors don't lean one way
	// over thousands of days.
	BacktestRounding = kernel.RoundHalfEven
	// unitPlaces is how many decimal places the units of each target the
	// backtest holds keep: more than the index, so a target priced in the
	// hundred thousands still keeps the index exact to BacktestPlaces.
	unitPlaces int32 = 18
)

// indexBase is the backtest's index on its first day, and a rebased
// series' on its first.
const indexBase = 100

// Close is one end-of-day value of a target in its own currency: an
// instrument's closing price or an index's closing level.
type Close struct {
	Date  kernel.Date
	Value kernel.Money
}

// Rate is the value in the reporting currency of one unit of Currency on
// Date. The rate for a day is the latest one dated on or before it.
type Rate struct {
	Currency kernel.Currency
	Date     kernel.Date
	Rate     kernel.Decimal
}

// Point is a series' value at the end of one day.
type Point struct {
	Date  kernel.Date
	Value kernel.Decimal
}

// MissingData says why a target has no history in the reporting currency.
// The zero value means it has some.
type MissingData int

const (
	// MissingCloses is a target with no close or level on or before the
	// last day asked for.
	MissingCloses MissingData = iota + 1
	// MissingRates is a target whose currency has no FX rate into the
	// reporting currency on or before its last close.
	MissingRates
)

// String returns "closes" or "rates".
func (m MissingData) String() string {
	switch m {
	case MissingCloses:
		return "closes"
	case MissingRates:
		return "rates"
	}
	return "MissingData(" + strconv.Itoa(int(m)) + ")"
}

// BacktestInput is what a backtest is computed from.
type BacktestInput struct {
	Targets []Target
	// Closes holds each target's closes or levels by Target.Key, any order.
	Closes map[string][]Close
	// Rates into Reporting, any order, of every currency a target is in.
	Rates     []Rate
	Reporting kernel.Currency
	// Through is the last day the backtest may reach.
	Through kernel.Date
}

// TargetHistory is one target's value in the reporting currency at the end
// of each day of its own history: from the first day it has a close and,
// in a foreign currency, a rate, to its last close on or before Through. A
// close and a rate hold until the next one.
type TargetHistory struct {
	Target  Target
	Days    []Point
	Missing MissingData
}

// BacktestSeries is the model portfolio's backtest (D-53): its index per
// day, 100 on the first day every target has data, and every target's own
// history. Days is empty when some target has no history or the targets'
// histories don't overlap.
type BacktestSeries struct {
	Days    []Point
	Targets []TargetHistory
}

// Backtest computes the model portfolio's value from end-of-day history
// (D-53, glossary: Backtest). Each target's value is its latest close
// converted at the latest rate into the reporting currency (D-61, D-62).
// The backtest starts on the first day every target has data, at 100, held
// at the target weights; it ends on the last day every target has data, at
// Through at the latest. At the end of each year's last day it rebalances
// to the target weights. Units of each target keep unitPlaces decimal
// places and the index BacktestPlaces, both rounded half-even.
func Backtest(in *BacktestInput) (BacktestSeries, error) {
	out := BacktestSeries{Targets: make([]TargetHistory, 0, len(in.Targets))}
	for i := range in.Targets {
		h, err := targetHistory(&in.Targets[i], in)
		if err != nil {
			return BacktestSeries{}, err
		}
		out.Targets = append(out.Targets, h)
	}
	common, ok := commonSpan(out.Targets)
	if !ok {
		return out, nil
	}
	days, err := simulate(out.Targets, common.first, common.last)
	if err != nil {
		return BacktestSeries{}, err
	}
	out.Days = days
	return out, nil
}

// span is the first and the last day of a series.
type span struct {
	first, last kernel.Date
}

// commonSpan returns the first and the last day every one of histories
// covers, and false when some has no days, there are none, or they don't
// overlap.
func commonSpan(histories []TargetHistory) (span, bool) {
	var common span
	for i := range histories {
		days := histories[i].Days
		if len(days) == 0 {
			return span{}, false
		}
		first, last := days[0].Date, days[len(days)-1].Date
		if i == 0 || first.After(common.first) {
			common.first = first
		}
		if i == 0 || last.Before(common.last) {
			common.last = last
		}
	}
	return common, len(histories) > 0 && !common.first.After(common.last)
}

// simulate runs the backtest over histories from start to end, both days
// every history covers.
func simulate(histories []TargetHistory, start, end kernel.Date) ([]Point, error) {
	units, err := rebalance(histories, start, kernel.NewDecimal(indexBase, 0))
	if err != nil {
		return nil, err
	}
	days := make([]Point, 0, start.DaysUntil(end)+1)
	for d := start; !d.After(end); d = d.AddDays(1) {
		value := kernel.Decimal{}
		for i := range histories {
			value = value.Add(units[i].Mul(valueOn(&histories[i], d)))
		}
		if value, err = value.Round(BacktestPlaces, BacktestRounding); err != nil {
			return nil, fmt.Errorf("backtest: %w", err)
		}
		days = append(days, Point{Date: d, Value: value})
		if d.Month() == time.December && d.Day() == 31 && d.Before(end) {
			if units, err = rebalance(histories, d, value); err != nil {
				return nil, err
			}
		}
	}
	return days, nil
}

// rebalance returns the units of each target that hold value at the target
// weights at the end of day d: weight / 100 × value / the target's value.
func rebalance(histories []TargetHistory, d kernel.Date, value kernel.Decimal) ([]kernel.Decimal, error) {
	units := make([]kernel.Decimal, len(histories))
	for i := range histories {
		h := &histories[i]
		u, err := h.Target.Weight.Mul(value).Quo(valueOn(h, d).Mul(kernel.NewDecimal(indexBase, 0)), unitPlaces, BacktestRounding)
		if err != nil {
			return nil, fmt.Errorf("backtest: rebalancing %s on %s: %w", h.Target.Key(), d, err)
		}
		units[i] = u
	}
	return units, nil
}

// valueOn returns h's value on d, a day its history covers.
func valueOn(h *TargetHistory, d kernel.Date) kernel.Decimal {
	return h.Days[h.Days[0].Date.DaysUntil(d)].Value
}
