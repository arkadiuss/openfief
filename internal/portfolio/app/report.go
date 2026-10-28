package app

import (
	"context"
	"fmt"

	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/portfolio"
)

// TargetReport is one target's own history in the reporting currency: its
// first and last day, why it has none, and its max drawdown over it (D-54).
type TargetReport struct {
	Target      portfolio.Target
	First, Last kernel.Date
	Missing     portfolio.MissingData
	Drawdown    portfolio.Drawdown
}

// Report is the model portfolio as the read API serves it while it is on:
// the model, its backtest from the first day every target has data (D-53),
// the model's max drawdown over the whole backtest and each target's over
// its own history (D-54), in the reporting currency.
type Report struct {
	Model     portfolio.Model
	Backtest  portfolio.BacktestSeries
	Drawdown  portfolio.Drawdown
	Targets   []TargetReport
	Reporting kernel.Currency
}

// Report returns the model portfolio's report from the end-of-day history
// stored through through, and false when the model portfolio is off: then
// nothing about it is read or returned (D-52).
func (m *Models) Report(ctx context.Context, through kernel.Date) (Report, bool, error) {
	model, err := m.store.Model(ctx)
	if err != nil {
		return Report{}, false, fmt.Errorf("loading the model portfolio: %w", err)
	}
	if !model.Enabled {
		return Report{}, false, nil
	}
	in, err := m.backtestInput(ctx, &model, through)
	if err != nil {
		return Report{}, false, err
	}
	series, err := portfolio.Backtest(in)
	if err != nil {
		return Report{}, false, fmt.Errorf("backtesting the model portfolio: %w", err)
	}
	out := Report{Model: model, Backtest: series, Reporting: m.reporting, Targets: make([]TargetReport, 0, len(series.Targets))}
	if out.Drawdown, err = portfolio.MaxDrawdown(series.Days); err != nil {
		return Report{}, false, fmt.Errorf("the model's drawdown: %w", err)
	}
	for i := range series.Targets {
		h := &series.Targets[i]
		tr := TargetReport{Target: h.Target, Missing: h.Missing}
		if len(h.Days) > 0 {
			tr.First, tr.Last = h.Days[0].Date, h.Days[len(h.Days)-1].Date
		}
		if tr.Drawdown, err = portfolio.MaxDrawdown(h.Days); err != nil {
			return Report{}, false, fmt.Errorf("drawdown of %s: %w", h.Target.Key(), err)
		}
		out.Targets = append(out.Targets, tr)
	}
	return out, true, nil
}

// backtestInput loads each target's closes or levels and the FX rates of
// every currency they are in, through through.
func (m *Models) backtestInput(ctx context.Context, model *portfolio.Model, through kernel.Date) (*portfolio.BacktestInput, error) {
	in := &portfolio.BacktestInput{
		Targets: model.Targets, Closes: map[string][]portfolio.Close{}, Reporting: m.reporting, Through: through,
	}
	currencies := map[kernel.Currency]bool{}
	for i := range model.Targets {
		t := &model.Targets[i]
		closes, err := m.closes(ctx, t, through)
		if err != nil {
			return nil, err
		}
		in.Closes[t.Key()] = closes
		for _, c := range closes {
			currencies[c.Value.Currency()] = true
		}
	}
	for c := range currencies {
		if c.Equal(m.reporting) {
			continue
		}
		rates, err := m.market.FXRates(ctx, c, through)
		if err != nil {
			return nil, fmt.Errorf("loading FX rates: %w", err)
		}
		for _, r := range rates {
			in.Rates = append(in.Rates, portfolio.Rate{Currency: r.Currency, Date: r.Date, Rate: r.Rate})
		}
	}
	return in, nil
}

// closes loads t's end-of-day history through through: an instrument's
// prices, or an index's levels in its currency.
func (m *Models) closes(ctx context.Context, t *portfolio.Target, through kernel.Date) ([]portfolio.Close, error) {
	switch t.Kind {
	case portfolio.TargetInstrument:
		prices, err := m.market.Prices(ctx, t.Name, through)
		if err != nil {
			return nil, fmt.Errorf("loading prices: %w", err)
		}
		out := make([]portfolio.Close, 0, len(prices))
		for _, p := range prices {
			out = append(out, portfolio.Close{Date: p.Date, Value: p.Close})
		}
		return out, nil
	case portfolio.TargetIndex:
		levels, err := m.market.IndexLevels(ctx, t.Name, through)
		if err != nil {
			return nil, fmt.Errorf("loading index levels: %w", err)
		}
		out := make([]portfolio.Close, 0, len(levels))
		for _, l := range levels {
			out = append(out, portfolio.Close{Date: l.Date, Value: kernel.NewMoney(l.Level, t.Currency)})
		}
		return out, nil
	}
	return nil, fmt.Errorf("%w: target %s", portfolio.ErrInvalidModel, t.Key())
}
