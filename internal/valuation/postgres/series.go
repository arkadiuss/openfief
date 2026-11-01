package postgres

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"slices"

	"github.com/jackc/pgx/v5"

	"github.com/arkadiuss/openfief/internal/valuation"
)

// valueTable is the table holding one kind's values and its column of the
// series' key.
type valueTable struct{ table, key string }

// valueTableOf returns the table of kind's values, and false for an
// unknown kind.
func valueTableOf(kind valuation.MarketDataKind) (valueTable, bool) {
	switch kind {
	case valuation.KindPrices:
		return valueTable{"valuation.prices", "instrument"}, true
	case valuation.KindFXRates:
		return valueTable{"valuation.fx_rates", "currency"}, true
	case valuation.KindIndexReadings:
		return valueTable{"valuation.index_readings", "kind"}, true
	}
	return valueTable{}, false
}

// AvailableSources returns the sources with a value of series, by name.
func (s *Store) AvailableSources(ctx context.Context, series valuation.Series) ([]string, error) {
	t, ok := valueTableOf(series.Kind)
	if !ok {
		return nil, fmt.Errorf("sources of %s: %w", series, valuation.ErrUnknownMarketDataKind)
	}
	rows, err := s.pool.Query(ctx, "SELECT DISTINCT source FROM "+t.table+" WHERE "+t.key+" = $1 ORDER BY source", series.Key)
	if err != nil {
		return nil, fmt.Errorf("loading the sources of %s: %w", series, err)
	}
	out, err := pgx.CollectRows(rows, pgx.RowTo[string])
	if err != nil {
		return nil, fmt.Errorf("loading the sources of %s: %w", series, err)
	}
	return out, nil
}

// SourceOf returns the source series takes its values from (D-94). A
// series without one gets it now, when it first needs a value: the source
// with a value of it that its kind's ranking places highest
// (valuation.ChooseSource), recorded so it never changes unless the owner
// changes it. It returns false when no source has a value of series yet.
// Two callers choosing at once record one choice: the first.
func (s *Store) SourceOf(ctx context.Context, series valuation.Series) (valuation.SeriesSource, bool, error) {
	if ss, ok, err := s.recordedSource(ctx, series); err != nil || ok {
		return ss, ok, err
	}
	available, err := s.AvailableSources(ctx, series)
	if err != nil {
		return valuation.SeriesSource{}, false, err
	}
	rank, err := s.rankOf(ctx, series.Kind)
	if err != nil {
		return valuation.SeriesSource{}, false, err
	}
	chosen, ok := valuation.ChooseSource(rank, available)
	if !ok {
		return valuation.SeriesSource{}, false, nil
	}
	if _, err := s.pool.Exec(ctx, `
		INSERT INTO valuation.series_sources (kind, key, source) VALUES ($1, $2, $3)
		ON CONFLICT (kind, key) DO NOTHING`, series.Kind.String(), series.Key, chosen); err != nil {
		return valuation.SeriesSource{}, false, fmt.Errorf("recording the source of %s: %w", series, err)
	}
	return s.recordedSource(ctx, series)
}

// recordedSource returns series' recorded source, and false when it has
// none.
func (s *Store) recordedSource(ctx context.Context, series valuation.Series) (valuation.SeriesSource, bool, error) {
	ss := valuation.SeriesSource{Series: series}
	err := s.pool.QueryRow(ctx, `SELECT source, by_owner FROM valuation.series_sources WHERE kind = $1 AND key = $2`,
		series.Kind.String(), series.Key).Scan(&ss.Source, &ss.ByOwner)
	if errors.Is(err, pgx.ErrNoRows) {
		return valuation.SeriesSource{}, false, nil
	}
	if err != nil {
		return valuation.SeriesSource{}, false, fmt.Errorf("loading the source of %s: %w", series, err)
	}
	return ss, true, nil
}

// SeriesSources returns every series with a recorded source, in kind then
// key order, each with the sources that have a value of it.
func (s *Store) SeriesSources(ctx context.Context) ([]valuation.SeriesSource, error) {
	rows, err := s.pool.Query(ctx, `SELECT kind, key, source, by_owner FROM valuation.series_sources`)
	if err != nil {
		return nil, fmt.Errorf("loading series sources: %w", err)
	}
	out, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (valuation.SeriesSource, error) {
		var kind string
		var ss valuation.SeriesSource
		if scanErr := row.Scan(&kind, &ss.Series.Key, &ss.Source, &ss.ByOwner); scanErr != nil {
			return valuation.SeriesSource{}, fmt.Errorf("reading a series source: %w", scanErr)
		}
		k, parseErr := valuation.ParseMarketDataKind(kind)
		if parseErr != nil {
			return valuation.SeriesSource{}, fmt.Errorf("%w: series source: %w", ErrCorrupt, parseErr)
		}
		ss.Series.Kind = k
		return ss, nil
	})
	if err != nil {
		return nil, fmt.Errorf("loading series sources: %w", err)
	}
	available, err := s.allAvailable(ctx)
	if err != nil {
		return nil, err
	}
	for i := range out {
		out[i].Available = available[out[i].Series]
	}
	slices.SortFunc(out, func(a, b valuation.SeriesSource) int {
		return cmp.Or(cmp.Compare(a.Series.Kind, b.Series.Kind), cmp.Compare(a.Series.Key, b.Series.Key))
	})
	return out, nil
}

// allAvailable returns the sources with a value of each series, by name.
func (s *Store) allAvailable(ctx context.Context) (map[valuation.Series][]string, error) {
	out := map[valuation.Series][]string{}
	for _, kind := range valuation.MarketDataKinds() {
		t, _ := valueTableOf(kind)
		rows, err := s.pool.Query(ctx, "SELECT DISTINCT "+t.key+", source FROM "+t.table+" ORDER BY 1, 2")
		if err != nil {
			return nil, fmt.Errorf("loading the sources of %s: %w", kind, err)
		}
		var key, source string
		_, err = pgx.ForEachRow(rows, []any{&key, &source}, func() error {
			series := valuation.Series{Kind: kind, Key: key}
			out[series] = append(out[series], source)
			return nil
		})
		if err != nil {
			return nil, fmt.Errorf("loading the sources of %s: %w", kind, err)
		}
	}
	return out, nil
}

// SaveSeriesSource records ss.Source as the owner's choice of ss.Series'
// source and, only when the source changed, enqueues revals in the same
// transaction (D-36). It reports whether the source changed.
func (s *Store) SaveSeriesSource(ctx context.Context, ss *valuation.SeriesSource, revals []valuation.Revaluation) (bool, error) {
	changed := false
	err := pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		var old string
		err := tx.QueryRow(ctx, `SELECT source FROM valuation.series_sources WHERE kind = $1 AND key = $2 FOR UPDATE`,
			ss.Series.Kind.String(), ss.Series.Key).Scan(&old)
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("loading the old source: %w", err)
		}
		changed = old != ss.Source
		if _, err := tx.Exec(ctx, `
			INSERT INTO valuation.series_sources (kind, key, source, by_owner) VALUES ($1, $2, $3, true)
			ON CONFLICT (kind, key) DO UPDATE SET source = EXCLUDED.source, by_owner = true, chosen_at = now()`,
			ss.Series.Kind.String(), ss.Series.Key, ss.Source); err != nil {
			return fmt.Errorf("storing the source: %w", err)
		}
		if !changed {
			return nil
		}
		return s.enqueueTx(ctx, tx, revals)
	})
	if err != nil {
		return false, fmt.Errorf("saving the source of %s: %w", ss.Series, err)
	}
	return changed, nil
}

// GapThresholds returns the owner's gap threshold of each kind that has
// one (D-94).
func (s *Store) GapThresholds(ctx context.Context) (valuation.GapThresholds, error) {
	rows, err := s.pool.Query(ctx, `SELECT kind, days FROM valuation.gap_thresholds`)
	if err != nil {
		return valuation.GapThresholds{}, fmt.Errorf("loading gap thresholds: %w", err)
	}
	out := valuation.GapThresholds{Days: map[valuation.MarketDataKind]int{}}
	var kind string
	var days int32
	_, err = pgx.ForEachRow(rows, []any{&kind, &days}, func() error {
		k, parseErr := valuation.ParseMarketDataKind(kind)
		if parseErr != nil {
			return fmt.Errorf("%w: gap threshold: %w", ErrCorrupt, parseErr)
		}
		out.Days[k] = int(days)
		return nil
	})
	if err != nil {
		return valuation.GapThresholds{}, fmt.Errorf("loading gap thresholds: %w", err)
	}
	return out, nil
}

// SaveGapThreshold sets kind's gap threshold to days.
func (s *Store) SaveGapThreshold(ctx context.Context, kind valuation.MarketDataKind, days int) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO valuation.gap_thresholds (kind, days) VALUES ($1, $2)
		ON CONFLICT (kind) DO UPDATE SET days = EXCLUDED.days, set_at = now()`, kind.String(), days)
	if err != nil {
		return fmt.Errorf("saving the gap threshold of %s: %w", kind, err)
	}
	return nil
}
