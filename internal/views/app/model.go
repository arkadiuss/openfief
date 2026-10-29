package app

import (
	"context"
	"errors"
	"fmt"

	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/portfolio"
	portfolioapp "github.com/arkadiuss/openfief/internal/portfolio/app"
	"github.com/arkadiuss/openfief/internal/views"
)

// ErrModelOff reports a query for the model portfolio while it is off, or
// in a core that serves none: then the read API says nothing about it
// (D-52).
var ErrModelOff = errors.New("the model portfolio is off")

// ModelReports reports the model portfolio. *portfolio/app.Models
// implements it.
type ModelReports interface {
	// Report returns the model portfolio's report from the history stored
	// through through, and false when it is off.
	Report(ctx context.Context, through kernel.Date) (portfolioapp.Report, bool, error)
}

// Option configures New.
type Option func(*Views)

// WithModel serves the model portfolio from models. Without it the model
// portfolio is always off.
func WithModel(models ModelReports) Option {
	return func(v *Views) { v.models = models }
}

// ModelPortfolio is the model portfolio as the read API serves it: its
// report, and its backtest from From to To rebased to 100 on its first day
// (D-53).
type ModelPortfolio struct {
	From, To kernel.Date
	Report   portfolioapp.Report
	Line     []portfolio.Point
}

// endOfHistory is the through date of a report when nothing is recorded:
// the last day there is, so every stored history counts.
func endOfHistory() (kernel.Date, error) {
	d, err := kernel.ParseDate("9999-12-31")
	if err != nil {
		return kernel.Date{}, fmt.Errorf("model portfolio: %w", err)
	}
	return d, nil
}

// ModelPortfolio returns the model portfolio, or an error wrapping
// ErrModelOff while it is off (D-52). from and to resolve as Performance
// does: a zero from is go-live, a from before go-live starts there, and a
// zero to is the latest valued-through date of any account, so the
// model's line and the time-weighted return cover the same days, both 100
// on the first (D-53). With nothing recorded, a zero from is the
// backtest's first day and a zero to all stored history. The drawdowns
// cover the whole history through to (D-54).
func (v *Views) ModelPortfolio(ctx context.Context, from, to kernel.Date) (ModelPortfolio, error) {
	if v.models == nil {
		return ModelPortfolio{}, ErrModelOff
	}
	r, err := v.modelRange(ctx, from, to)
	if err != nil {
		return ModelPortfolio{}, err
	}
	report, on, err := v.models.Report(ctx, r.through)
	if err != nil {
		return ModelPortfolio{}, fmt.Errorf("model portfolio: %w", err)
	}
	if !on {
		return ModelPortfolio{}, ErrModelOff
	}
	out := ModelPortfolio{From: r.from, To: r.to, Report: report}
	if out.Line, err = portfolio.Since(report.Backtest.Days, r.from); err != nil {
		return ModelPortfolio{}, fmt.Errorf("model portfolio: %w", err)
	}
	if len(out.Line) > 0 {
		out.From, out.To = out.Line[0].Date, out.Line[len(out.Line)-1].Date
	}
	return out, nil
}

// modelDays is the model's line's range as resolved, and the date its
// report runs through.
type modelDays struct {
	from, to, through kernel.Date
}

// modelRange resolves from and to as ModelPortfolio says.
func (v *Views) modelRange(ctx context.Context, from, to kernel.Date) (modelDays, error) {
	accounts, err := v.Accounts(ctx)
	if err != nil {
		return modelDays{}, err
	}
	rec, err := v.recorded(ctx, accounts)
	if err != nil {
		return modelDays{}, err
	}
	if len(rec.txs) > 0 {
		from = later(from, rec.goLive)
		if to.IsZero() {
			to = later(rec.latest, from)
		}
	}
	if to.IsZero() {
		through, err := endOfHistory()
		return modelDays{from: from, through: through}, err
	}
	if to.Before(from) {
		return modelDays{}, fmt.Errorf("model portfolio: %w: %s is before %s", views.ErrInvalidRange, to, from)
	}
	return modelDays{from: from, to: to, through: to}, nil
}
