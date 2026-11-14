package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/portfolio"
	"github.com/arkadiuss/openfief/internal/portfolio/app"
)

// ErrCorrupt reports a stored row the portfolio context can't read back.
var ErrCorrupt = errors.New("corrupt portfolio row")

// Store keeps the model portfolio, the sleeve kinds, the sleeves and which
// account is in which sleeve in the portfolio schema. It implements
// portfolio/app.ModelStore and portfolio/app.RegistryStore. New returns
// one over a pool, where each call commits on its own; InTx passes one
// whose calls all run in one transaction.
type Store struct {
	db DB
	tx pgx.Tx // nil outside InTx
}

// DB is what Store runs its statements on: a *pgxpool.Pool, a pgx.Tx, or
// anything that answers as one, such as the core's pool that joins the
// transaction a context carries. Begin on a pgx.Tx starts a savepoint.
type DB interface {
	Begin(ctx context.Context) (pgx.Tx, error)
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

// New returns the store over pool. Run Migrate first.
func New(pool DB) *Store { return &Store{db: pool} }

// InTx runs fn with a Store whose calls all run in one database
// transaction, committed when fn returns nil and rolled back otherwise.
// Inside InTx it runs fn in the same transaction.
func (s *Store) InTx(ctx context.Context, fn func(tx app.RegistryStore) error) error {
	if s.tx != nil {
		return fn(s)
	}
	err := pgx.BeginFunc(ctx, s.db, func(tx pgx.Tx) error {
		return fn(&Store{db: tx, tx: tx})
	})
	if err != nil {
		return fmt.Errorf("portfolio transaction: %w", err)
	}
	return nil
}

// Model returns the stored model portfolio, or the zero Model, off with no
// targets, when none was ever saved (D-52).
func (s *Store) Model(ctx context.Context) (portfolio.Model, error) {
	var m portfolio.Model
	err := s.db.QueryRow(ctx, "SELECT enabled FROM portfolio.model").Scan(&m.Enabled)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return portfolio.Model{}, fmt.Errorf("loading the model portfolio: %w", err)
	}
	rows, err := s.db.Query(ctx, `
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
	err := pgx.BeginFunc(ctx, s.db, func(tx pgx.Tx) error {
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
