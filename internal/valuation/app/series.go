package app

import (
	"context"
	"fmt"
	"slices"

	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
	"github.com/arkadiuss/openfief/internal/valuation"
)

// SeriesSources returns, to the owner, every series that has a source, in
// kind then key order, each with the sources that have a value of it
// (D-94).
func (r *Rankings) SeriesSources(ctx context.Context, origin, sessionToken string) ([]valuation.SeriesSource, error) {
	if err := r.checkOwner(ctx, origin, sessionToken); err != nil {
		return nil, err
	}
	out, err := r.store.SeriesSources(ctx)
	if err != nil {
		return nil, fmt.Errorf("loading the series' sources: %w", err)
	}
	return out, nil
}

// SetSeriesSource makes source the source of series, for the owner (D-94):
// from then on the series takes every value from it, and the other
// sources' values only cross-check it. A series breaking
// valuation.Series.Validate, or a source without any value of the series,
// is an error wrapping ErrInvalidInput, and nothing changes. When the
// source changed, the accounts it affects are revalued from their first
// transaction in the same transaction (D-36): for prices, every account
// that held the instrument; for an FX rate or an index, every account.
func (r *Rankings) SetSeriesSource(ctx context.Context, origin, sessionToken string, series valuation.Series, source string) error {
	if err := r.checkOwner(ctx, origin, sessionToken); err != nil {
		return err
	}
	if err := series.Validate(); err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidInput, err)
	}
	available, err := r.store.AvailableSources(ctx, series)
	if err != nil {
		return fmt.Errorf("setting the source of %s: %w", series, err)
	}
	if !slices.Contains(available, source) {
		return fmt.Errorf("%w: %q has no value of %s", ErrInvalidInput, source, series)
	}
	var revals []valuation.Revaluation
	if series.Kind == valuation.KindPrices {
		revals, err = r.planner.ForInstrumentFromStart(ctx, series.Key)
	} else {
		revals, err = r.planner.FromStart(ctx, "")
	}
	if err != nil {
		return fmt.Errorf("setting the source of %s: %w", series, err)
	}
	if _, err := r.store.SaveSeriesSource(ctx, &valuation.SeriesSource{Series: series, Source: source, ByOwner: true}, revals); err != nil {
		return fmt.Errorf("setting the source of %s: %w", series, err)
	}
	return nil
}

// SetGapThreshold makes days kind's gap threshold, for the owner (D-94): a
// series of kind whose source has had no value for longer raises an
// attention item. It applies from the next revaluation, the end-of-day one
// at the latest, so nothing is revalued now. A threshold breaking
// valuation.ValidateGapThreshold is an error wrapping ErrInvalidInput.
func (r *Rankings) SetGapThreshold(ctx context.Context, origin, sessionToken string, kind valuation.MarketDataKind, days int) error {
	if err := r.checkOwner(ctx, origin, sessionToken); err != nil {
		return err
	}
	if err := valuation.ValidateGapThreshold(kind, days); err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidInput, err)
	}
	if err := r.store.SaveGapThreshold(ctx, kind, days); err != nil {
		return fmt.Errorf("setting the gap threshold of %s: %w", kind, err)
	}
	return nil
}

// ForInstrumentFromStart returns a revaluation from its first transaction
// of every account whose log has a holding posting of instrument
// (ledger.Instrument.String()).
func (p *Planner) ForInstrumentFromStart(ctx context.Context, instrument string) ([]valuation.Revaluation, error) {
	logs, err := p.logs(ctx)
	if err != nil {
		return nil, err
	}
	var out []valuation.Revaluation
	for _, a := range logs {
		first, ok := firstDate(a.log)
		if ok && held(a.log, func(in ledger.Instrument) bool { return in.String() == instrument }) {
			out = append(out, valuation.Revaluation{Account: a.id, From: first})
		}
	}
	return out, nil
}

// GapKind is the kind of the attention item that tells the owner a series'
// source has had no value for longer than its gap threshold (D-94, D-130);
// its subject is the account and the series (gapSubject).
const GapKind = "market-data-gap"

// gapSubject is the subject of account's gap item for series.
func gapSubject(account string, series valuation.Series) string {
	return account + "/" + series.String()
}

// latestOnOrBefore returns the latest of dates on or before day, zero when
// none is.
func latestOnOrBefore(dates []kernel.Date, day kernel.Date) kernel.Date {
	var out kernel.Date
	for _, d := range dates {
		if !d.After(day) && (out.IsZero() || d.After(out)) {
			out = d
		}
	}
	return out
}
