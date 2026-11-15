package app

import (
	"context"
	"fmt"

	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
	"github.com/arkadiuss/openfief/internal/views"
)

// PeriodChange is how the portfolio's value changed over a period, with
// whether a revaluation that could change it is pending or running.
type PeriodChange struct {
	views.PeriodChange

	// Pending reports a revaluation of some account that is pending or
	// running: its stored valuations may change.
	Pending bool
}

// PeriodChange returns how the portfolio's value changed over the days from
// through to, both inclusive (D-61, D-77): its gross, net and liquidation
// values at the start and the end, the time-weighted return and the
// deposits and withdrawals in the period (views.ChangeOver). The start is
// the end of the day before from, or of from itself when from is the first
// day recorded or earlier, so the period never reaches before the record.
// A to before from is views.ErrInvalidRange, a range of more than
// views.MaxRangeDays too; a portfolio with nothing recorded, a to after the
// latest day every account is valued through, or a day in the period that
// lacks a valuation or an FX rate is views.ErrPeriodUnvalued.
func (v *Views) PeriodChange(ctx context.Context, from, to kernel.Date) (PeriodChange, error) {
	if err := views.ValidateRange(from, to); err != nil {
		return PeriodChange{}, fmt.Errorf("period change: %w", err)
	}
	accounts, err := v.Accounts(ctx)
	if err != nil {
		return PeriodChange{}, err
	}
	rec, err := v.recorded(ctx, accounts)
	if err != nil {
		return PeriodChange{}, err
	}
	opening, err := openingDay(&rec, from, to)
	if err != nil {
		return PeriodChange{}, err
	}
	links, err := v.ledger.TransferLinks(ctx)
	if err != nil {
		return PeriodChange{}, fmt.Errorf("loading transfer links: %w", err)
	}
	perf, err := v.performance(ctx, accounts, &rec, starts{def: rec.start}, opening, to, links)
	if err != nil {
		return PeriodChange{}, err
	}
	in := &views.PeriodInput{From: from, To: to, Days: perf.Series.Days, Reporting: v.reporting}
	if in.Opening, err = v.totalOn(ctx, accounts, opening); err != nil {
		return PeriodChange{}, err
	}
	if in.Closing, err = v.totalOn(ctx, accounts, to); err != nil {
		return PeriodChange{}, err
	}
	change, err := views.ChangeOver(in)
	if err != nil {
		return PeriodChange{}, fmt.Errorf("period change: %w", err)
	}
	return PeriodChange{PeriodChange: change, Pending: rec.pending}, nil
}

// openingDay returns the day whose end values start the period from..to over
// the record rec: the day before from, or the first day recorded when that
// is earlier. A period the record can't reach, because nothing is recorded,
// it ends after the day every account is valued through or before the first
// day recorded, is views.ErrPeriodUnvalued.
func openingDay(rec *record, from, to kernel.Date) (kernel.Date, error) {
	if len(rec.txs) == 0 || to.After(rec.through) {
		return kernel.Date{}, fmt.Errorf("period change: %w: %s is after the last day valued", views.ErrPeriodUnvalued, to)
	}
	opening := from.AddDays(-1)
	if opening.Before(rec.first) {
		opening = rec.first
	}
	if to.Before(opening) {
		return kernel.Date{}, fmt.Errorf("period change: %w: %s is before the first day recorded", views.ErrPeriodUnvalued, to)
	}
	return opening, nil
}

// totalOn returns the portfolio's total on day, or views.ErrPeriodUnvalued
// when no account has a valuation or cash that day.
func (v *Views) totalOn(ctx context.Context, accounts []ledger.Account, day kernel.Date) (views.DateTotal, error) {
	totals, err := v.totals(ctx, accounts, day, day)
	if err != nil {
		return views.DateTotal{}, err
	}
	for i := range totals {
		if totals[i].Date.Equal(day) {
			return totals[i], nil
		}
	}
	return views.DateTotal{}, fmt.Errorf("%w: nothing is valued on %s", views.ErrPeriodUnvalued, day)
}
