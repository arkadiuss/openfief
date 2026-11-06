package app

import (
	"cmp"
	"context"
	"fmt"
	"slices"

	"github.com/arkadiuss/openfief/internal/jurisdiction/pl"
	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
	"github.com/arkadiuss/openfief/internal/valuation"
)

// AttentionRaiser raises and clears attention items for the owner (D-130),
// the shape of the attention store's port, such as a gap in a series'
// market data (raiseGaps). Both are idempotent: raising an item that is
// open, or clearing one that isn't, changes nothing.
type AttentionRaiser interface {
	Raise(ctx context.Context, kind, subject, detail string) error
	Clear(ctx context.Context, kind, subject string) error
}

// WithAttention makes r raise and clear the attention items its
// revaluations find through a, and returns r.
func (r *Revaluer) WithAttention(a AttentionRaiser) *Revaluer {
	r.attention = a
	return r
}

// raiseGaps raises an attention item for each series a holding of account
// needs on today, the last day valued, whose source has had no value for
// longer than its kind's gap threshold, and clears it for every other
// series the revaluation read, so an item resolves when values resume or
// no holding needs the series any more (D-94, D-130). Its subject names
// the account, so one account's revaluation never clears another's item.
func (r *Revaluer) raiseGaps(ctx context.Context, in *inputCache, account ledger.AccountID, rows []valuation.HoldingValuation, today kernel.Date) error {
	if r.attention == nil {
		return nil
	}
	gaps, err := r.gaps(ctx, in, NeededSeries(rows, today, in.reporting, in.indexation), today)
	if err != nil {
		return err
	}
	gapped := map[valuation.Series]bool{}
	for i := range gaps {
		g := &gaps[i]
		gapped[g.Series] = true
		if err := r.attention.Raise(ctx, GapKind, gapSubject(string(account), g.Series), g.Detail()); err != nil {
			return fmt.Errorf("raising an attention item for %s: %w", g.Series, err)
		}
	}
	for _, series := range in.read() {
		if gapped[series] {
			continue
		}
		if err := r.attention.Clear(ctx, GapKind, gapSubject(string(account), series)); err != nil {
			return fmt.Errorf("clearing the attention item for %s: %w", series, err)
		}
	}
	return nil
}

// gaps returns the gaps of needed on today: each series whose source's
// latest value the revaluation loaded is older than its kind's threshold.
// A series with no source yet, because no source has a value of it, has no
// gap: its holdings are stored with the missing input.
func (r *Revaluer) gaps(ctx context.Context, in *inputCache, needed []valuation.Series, today kernel.Date) ([]valuation.SeriesGap, error) {
	thresholds, err := r.store.GapThresholds(ctx)
	if err != nil {
		return nil, fmt.Errorf("loading gap thresholds: %w", err)
	}
	var out []valuation.SeriesGap
	for _, series := range needed {
		src, ok, err := r.store.SourceOf(ctx, series)
		if err != nil {
			return nil, fmt.Errorf("loading the source of %s: %w", series, err)
		}
		if !ok {
			continue
		}
		if g, gap := valuation.GapOf(series, src.Source, latestOnOrBefore(in.dates(series), today), today, thresholds.For(series.Kind)); gap {
			out = append(out, g)
		}
	}
	return out, nil
}

// NeededSeries returns the series the holdings among rows valued on day
// take their values from, in order of first need without repeats: a
// holding at market needs its instrument's prices and, unless its price is
// in reporting, the pack's currency its valuation is stored in (D-180),
// its currency's FX rates; a bond needs the period index its type
// is indexed to, which indexation returns for its ISIN (pl.Fixed for none).
// A holding with no price yet needs no FX rate, since its currency is
// unknown.
func NeededSeries(rows []valuation.HoldingValuation, day kernel.Date, reporting kernel.Currency, indexation func(ledger.ISIN) pl.Indexation) []valuation.Series {
	var out []valuation.Series
	for i := range rows {
		v := &rows[i]
		if !v.Date.Equal(day) {
			continue
		}
		for _, s := range seriesOf(v, reporting, indexation) {
			if !slices.Contains(out, s) {
				out = append(out, s)
			}
		}
	}
	return out
}

// seriesOf returns the series the holding valued in v needs (NeededSeries).
func seriesOf(v *valuation.HoldingValuation, reporting kernel.Currency, indexation func(ledger.ISIN) pl.Indexation) []valuation.Series {
	switch v.Basis {
	case valuation.Market:
		out := []valuation.Series{valuation.PriceSeries(v.Instrument.String())}
		if c := v.Value.Currency(); !c.IsZero() && !c.Equal(reporting) {
			out = append(out, valuation.FXSeries(c))
		}
		return out
	case valuation.GrossAccrued:
		switch indexation(v.Instrument.ISIN) {
		case pl.Inflation:
			return []valuation.Series{valuation.IndexSeries(valuation.CPI)}
		case pl.ReferenceRate:
			return []valuation.Series{valuation.IndexSeries(valuation.ReferenceRate)}
		case pl.Fixed:
		}
	}
	return nil
}

// indexation returns the period index the bond with isin is indexed to,
// from its series terms the revaluation loaded; pl.Fixed when it has none.
func (c *inputCache) indexation(isin ledger.ISIN) pl.Indexation {
	t := c.terms[isin]
	if t == nil {
		return pl.Fixed
	}
	st, err := t.Series.Type.Structure()
	if err != nil {
		return pl.Fixed
	}
	return st.Indexation
}

// dates returns the dates of the values of series the revaluation loaded:
// a reading is dated on its month's first day.
func (c *inputCache) dates(series valuation.Series) []kernel.Date {
	switch series.Kind {
	case valuation.KindPrices:
		return datesOf(c.pricesOf[series.Key], func(p *valuation.Price) kernel.Date { return p.Date })
	case valuation.KindFXRates:
		cur, err := kernel.ParseCurrency(series.Key)
		if err != nil {
			return nil
		}
		return datesOf(c.ratesOf[cur], func(r *valuation.FXRate) kernel.Date { return r.Date })
	case valuation.KindIndexReadings:
		var out []kernel.Date
		for i := range c.readings {
			if r := &c.readings[i]; r.Kind.String() == series.Key {
				if d, err := kernel.NewDate(r.Year, r.Month, 1); err == nil {
					out = append(out, d)
				}
			}
		}
		return out
	}
	return nil
}

// datesOf returns the date of each of values.
func datesOf[T any](values []T, date func(*T) kernel.Date) []kernel.Date {
	out := make([]kernel.Date, 0, len(values))
	for i := range values {
		out = append(out, date(&values[i]))
	}
	return out
}

// read returns every series whose values the revaluation loaded, in kind
// then key order.
func (c *inputCache) read() []valuation.Series {
	var out []valuation.Series
	for instrument := range c.pricesOf {
		out = append(out, valuation.PriceSeries(instrument))
	}
	for cur := range c.ratesOf {
		out = append(out, valuation.FXSeries(cur))
	}
	if c.haveIndexR {
		out = append(out, valuation.IndexSeries(valuation.CPI), valuation.IndexSeries(valuation.ReferenceRate))
	}
	slices.SortFunc(out, func(a, b valuation.Series) int {
		return cmp.Or(cmp.Compare(a.Kind, b.Kind), cmp.Compare(a.Key, b.Key))
	})
	return out
}
