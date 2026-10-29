package postgres

import (
	"context"
	"fmt"
	"slices"

	"github.com/jackc/pgx/v5"

	"github.com/arkadiuss/openfief/internal/valuation"
	"github.com/arkadiuss/openfief/internal/valuation/app"
)

var _ app.RankingStore = (*Store)(nil)

// SaveSources records sources in valuation.sources, each kind and source
// once, with the institution it last named.
func (s *Store) SaveSources(ctx context.Context, sources []valuation.Source) error {
	err := s.inTx(ctx, nil, func(tx pgx.Tx) error {
		batch := &pgx.Batch{}
		for _, src := range sources {
			batch.Queue(`
				INSERT INTO valuation.sources (kind, source, institution) VALUES ($1, $2, $3)
				ON CONFLICT (kind, source) DO UPDATE SET institution = EXCLUDED.institution`,
				src.Kind.String(), src.Name, src.Institution)
		}
		return tx.SendBatch(ctx, batch).Close()
	})
	if err != nil {
		return fmt.Errorf("saving market-data sources: %w", err)
	}
	return nil
}

// SourceRanking returns the ranking of kind's sources: the owner's order,
// then every other source that delivered kind in its default place
// (valuation.RankSources).
func (s *Store) SourceRanking(ctx context.Context, kind valuation.MarketDataKind) (valuation.Ranking, error) {
	r, err := rankingIn(ctx, s.pool, kind)
	if err != nil {
		return valuation.Ranking{}, err
	}
	return r, nil
}

// querier runs queries: the pool or a transaction.
type querier interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
}

// rankingIn reads kind's ranking through q.
func rankingIn(ctx context.Context, q querier, kind valuation.MarketDataKind) (valuation.Ranking, error) {
	placed, err := placedIn(ctx, q, kind)
	if err != nil {
		return valuation.Ranking{}, err
	}
	rows, err := q.Query(ctx, `SELECT source, institution FROM valuation.sources WHERE kind = $1 ORDER BY source`, kind.String())
	if err != nil {
		return valuation.Ranking{}, fmt.Errorf("loading the sources of %s: %w", kind, err)
	}
	known, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (valuation.Source, error) {
		src := valuation.Source{Kind: kind}
		if scanErr := row.Scan(&src.Name, &src.Institution); scanErr != nil {
			return valuation.Source{}, fmt.Errorf("reading a source: %w", scanErr)
		}
		return src, nil
	})
	if err != nil {
		return valuation.Ranking{}, fmt.Errorf("loading the sources of %s: %w", kind, err)
	}
	return valuation.RankSources(kind, known, placed), nil
}

// placedIn reads the owner's order of kind's sources through q.
func placedIn(ctx context.Context, q querier, kind valuation.MarketDataKind) ([]string, error) {
	rows, err := q.Query(ctx, `SELECT source FROM valuation.source_rankings WHERE kind = $1 ORDER BY position`, kind.String())
	if err != nil {
		return nil, fmt.Errorf("loading the ranking of %s: %w", kind, err)
	}
	placed, err := pgx.CollectRows(rows, pgx.RowTo[string])
	if err != nil {
		return nil, fmt.Errorf("loading the ranking of %s: %w", kind, err)
	}
	return placed, nil
}

// rankOf returns the names of kind's sources, highest first.
func (s *Store) rankOf(ctx context.Context, kind valuation.MarketDataKind) ([]string, error) {
	r, err := rankingIn(ctx, s.pool, kind)
	if err != nil {
		return nil, err
	}
	return r.Names(), nil
}

// SaveSourceRanking replaces the owner's order of kind's sources with
// placed and, only when it differs from the stored one, enqueues revals in
// the same transaction (D-36). It reports whether it changed.
func (s *Store) SaveSourceRanking(ctx context.Context, kind valuation.MarketDataKind, placed []string, revals []valuation.Revaluation) (bool, error) {
	changed := false
	err := pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		if _, err := tx.Exec(ctx, `SELECT pg_advisory_xact_lock(hashtext('valuation.source_rankings'), hashtext($1))`, kind.String()); err != nil {
			return fmt.Errorf("locking the ranking: %w", err)
		}
		old, err := placedIn(ctx, tx, kind)
		if err != nil {
			return err
		}
		if slices.Equal(old, placed) {
			return nil
		}
		changed = true
		if _, err := tx.Exec(ctx, `DELETE FROM valuation.source_rankings WHERE kind = $1`, kind.String()); err != nil {
			return fmt.Errorf("removing the old ranking: %w", err)
		}
		batch := &pgx.Batch{}
		for i, name := range placed {
			batch.Queue(`INSERT INTO valuation.source_rankings (kind, position, source) VALUES ($1, $2, $3)`, kind.String(), i, name)
		}
		if err := tx.SendBatch(ctx, batch).Close(); err != nil {
			return fmt.Errorf("storing the ranking: %w", err)
		}
		return s.enqueueTx(ctx, tx, revals)
	})
	if err != nil {
		return false, fmt.Errorf("saving the ranking of %s: %w", kind, err)
	}
	return changed, nil
}
