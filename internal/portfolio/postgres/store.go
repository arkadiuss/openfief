package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/portfolio"
)

// ErrCorrupt reports a stored row the portfolio context can't read back.
var ErrCorrupt = errors.New("corrupt portfolio row")

// Store keeps the model portfolio in the portfolio schema. It implements
// portfolio/app.ModelStore.
type Store struct {
	pool *pgxpool.Pool
}

// New returns the store over pool. Run Migrate first.
func New(pool *pgxpool.Pool) *Store { return &Store{pool: pool} }

// Model returns the stored model portfolio, or the zero Model, off with no
// targets, when none was ever saved (D-52).
func (s *Store) Model(ctx context.Context) (portfolio.Model, error) {
	var m portfolio.Model
	err := s.pool.QueryRow(ctx, "SELECT enabled FROM portfolio.model").Scan(&m.Enabled)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return portfolio.Model{}, fmt.Errorf("loading the model portfolio: %w", err)
	}
	rows, err := s.pool.Query(ctx, `
		SELECT kind, name, coalesce(currency, ''), weight::text FROM portfolio.model_targets ORDER BY position`)
	if err != nil {
		return portfolio.Model{}, fmt.Errorf("loading the model's targets: %w", err)
	}
	if m.Targets, err = pgx.CollectRows(rows, scanTarget); err != nil {
		return portfolio.Model{}, fmt.Errorf("loading the model's targets: %w", err)
	}
	if len(m.Targets) == 0 {
		m.Targets = nil
	}
	return m, nil
}

// scanTarget reads one target row.
func scanTarget(row pgx.CollectableRow) (portfolio.Target, error) {
	var kind, name, currency, weight string
	if err := row.Scan(&kind, &name, &currency, &weight); err != nil {
		return portfolio.Target{}, fmt.Errorf("reading a target: %w", err)
	}
	t := portfolio.Target{Name: name}
	var errs [3]error
	t.Kind, errs[0] = portfolio.ParseTargetKind(kind)
	t.Weight, errs[1] = kernel.ParseDecimal(weight)
	if currency != "" {
		t.Currency, errs[2] = kernel.ParseCurrency(currency)
	}
	if err := errors.Join(errs[:]...); err != nil {
		return portfolio.Target{}, fmt.Errorf("%w: target %s: %w", ErrCorrupt, name, err)
	}
	return t, nil
}

// SaveModel replaces the stored model portfolio, whether it is on and all
// its targets, in one database transaction. It stores what it is given; the
// portfolio context checks the model first (app.CheckModel).
func (s *Store) SaveModel(ctx context.Context, m portfolio.Model) error {
	err := pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		if _, err := tx.Exec(ctx, `
			INSERT INTO portfolio.model (singleton, enabled) VALUES (true, $1)
			ON CONFLICT (singleton) DO UPDATE SET enabled = EXCLUDED.enabled, updated_at = now()`, m.Enabled); err != nil {
			return fmt.Errorf("storing whether it is on: %w", err)
		}
		if _, err := tx.Exec(ctx, "DELETE FROM portfolio.model_targets"); err != nil {
			return fmt.Errorf("removing the old targets: %w", err)
		}
		batch := &pgx.Batch{}
		for i := range m.Targets {
			t := &m.Targets[i]
			var currency *string
			if !t.Currency.IsZero() {
				c := t.Currency.String()
				currency = &c
			}
			batch.Queue(`INSERT INTO portfolio.model_targets (position, kind, name, currency, weight) VALUES ($1, $2, $3, $4, $5)`,
				i, t.Kind.String(), t.Name, currency, t.Weight.String())
		}
		return tx.SendBatch(ctx, batch).Close()
	})
	if err != nil {
		return fmt.Errorf("saving the model portfolio: %w", err)
	}
	return nil
}
