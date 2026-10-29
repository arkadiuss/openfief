package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/arkadiuss/openfief/internal/ingest"
	"github.com/arkadiuss/openfief/internal/ingest/app"
	"github.com/arkadiuss/openfief/internal/ledger"
)

// ErrBatchNotOpen reports a batch outcome for a batch that is not the
// instance's open one.
var ErrBatchNotOpen = errors.New("import batch is not open")

// ErrUnknownMode reports a stored mode that is not an ingest mode.
var ErrUnknownMode = errors.New("unknown ingest mode")

// Store is app.SyncStore, app.BindingStore, app.AccountBindings,
// app.BatchRecorder and app.BrokerAccountStore in PostgreSQL, in the ingest schema that Migrate
// creates.
type Store struct {
	pool *pgxpool.Pool
}

var (
	_ app.SyncStore       = (*Store)(nil)
	_ app.BindingStore    = (*Store)(nil)
	_ app.BatchRecorder   = (*Store)(nil)
	_ app.AccountBindings = (*Store)(nil)
)

// New returns the store over pool. Run Migrate first.
func New(pool *pgxpool.Pool) *Store {
	return &Store{pool: pool}
}

// ClaimSync claims instance until until, unless an unexpired claim holds
// it at now, and returns its state; a new instance starts with none. The
// instance counts as seeded only while every broker account mapped for it
// is seeded too, so a broker account the owner maps later is seeded by the
// next sync (D-72, D-23).
func (s *Store) ClaimSync(ctx context.Context, instance string, now, until time.Time) (app.SyncClaim, error) {
	rows, err := s.pool.Query(ctx, `
		INSERT INTO ingest.connector_instances AS i (instance, lease_until) VALUES ($1, $3)
		ON CONFLICT (instance) DO UPDATE SET lease_until = EXCLUDED.lease_until
			WHERE i.lease_until IS NULL OR i.lease_until <= $2
		RETURNING watermark,
			seeded AND NOT EXISTS (SELECT FROM ingest.broker_accounts b
				WHERE b.instance = i.instance AND b.account_id IS NOT NULL AND NOT b.seeded),
			failures, open_batch,
			(SELECT b.mode FROM ingest.import_batches b WHERE b.id = i.open_batch)`,
		instance, now, until)
	if err != nil {
		return app.SyncClaim{}, fmt.Errorf("claiming %s: %w", instance, err)
	}
	claim, err := pgx.CollectExactlyOneRow(rows, scanClaim)
	if errors.Is(err, pgx.ErrNoRows) {
		return app.SyncClaim{}, fmt.Errorf("%w: %s", app.ErrSyncRunning, instance)
	}
	if err != nil {
		return app.SyncClaim{}, fmt.Errorf("claiming %s: %w", instance, err)
	}
	return claim, nil
}

// scanClaim reads one row of ClaimSync's RETURNING.
func scanClaim(row pgx.CollectableRow) (app.SyncClaim, error) {
	var (
		c          app.SyncClaim
		open, mode *string
	)
	if err := row.Scan(&c.Watermark, &c.Seeded, &c.Failures, &open, &mode); err != nil {
		return app.SyncClaim{}, fmt.Errorf("scanning: %w", err)
	}
	if open != nil && mode != nil {
		m, err := parseMode(*mode)
		if err != nil {
			return app.SyncClaim{}, err
		}
		c.Open = &app.OpenBatch{ID: ingest.BatchID(*open), Mode: m}
	}
	return c, nil
}

// parseMode parses a mode as ingest.Mode.String writes it.
func parseMode(s string) (ingest.Mode, error) {
	for _, m := range []ingest.Mode{ingest.ModeOpening, ingest.ModeTransactions} {
		if m.String() == s {
			return m, nil
		}
	}
	return 0, fmt.Errorf("%w: %q", ErrUnknownMode, s)
}

// StartBatch records b as instance's open sync batch, abandoning the open
// batch it replaces, in one database transaction.
func (s *Store) StartBatch(ctx context.Context, instance string, b app.OpenBatch, at time.Time) error {
	err := pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		if _, err := tx.Exec(ctx, `
			UPDATE ingest.import_batches SET state = 'abandoned'
			WHERE state = 'open' AND id = (SELECT open_batch FROM ingest.connector_instances WHERE instance = $1)`,
			instance); err != nil {
			return fmt.Errorf("abandoning the open batch: %w", err)
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO ingest.import_batches (id, instance, source, mode, started_at) VALUES ($1, $2, $3, $4, $5)`,
			string(b.ID), instance, app.SourceSync, b.Mode.String(), at); err != nil {
			return fmt.Errorf("inserting: %w", err)
		}
		if _, err := tx.Exec(ctx, `UPDATE ingest.connector_instances SET open_batch = $2 WHERE instance = $1`,
			instance, string(b.ID)); err != nil {
			return fmt.Errorf("opening: %w", err)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("starting batch %s of %s: %w", b.ID, instance, err)
	}
	return nil
}

// CompleteSync records b as completed and stores the instance's watermark,
// seeded flag, next sync and released claim, in one database transaction.
func (s *Store) CompleteSync(ctx context.Context, instance string, b *ingest.Batch, watermark string, next, at time.Time) error {
	err := pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		tag, err := tx.Exec(ctx, `
			UPDATE ingest.import_batches SET state = 'completed', attempts = attempts + 1, last_attempt_at = $3,
				finished_at = $3, institution = $4, transactions = $5, assertions = $6, new = $7, new_versions = $8,
				unchanged = $9, failed_assertions = $10, problems = $11
			WHERE id = $1 AND instance = $2 AND state = 'open'`,
			string(b.ID), instance, at, string(b.Institution), b.Transactions, b.Assertions, b.Imported.New,
			b.Imported.NewVersions, b.Imported.Unchanged, b.Imported.FailedAssertions, len(b.Problems))
		if err != nil {
			return fmt.Errorf("completing: %w", err)
		}
		if tag.RowsAffected() == 0 {
			return ErrBatchNotOpen
		}
		if _, err := tx.Exec(ctx, `
			UPDATE ingest.connector_instances SET watermark = $2, seeded = seeded OR $3, failures = 0,
				next_sync_at = $4, lease_until = NULL, open_batch = NULL
			WHERE instance = $1`,
			instance, watermark, b.Mode == ingest.ModeOpening, next); err != nil {
			return fmt.Errorf("storing the watermark: %w", err)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("completing batch %s of %s: %w", b.ID, instance, err)
	}
	return nil
}

// FailSync counts a failed attempt of the open batch id and stores the
// instance's failures, next sync and released claim, in one database
// transaction; the batch stays open.
func (s *Store) FailSync(ctx context.Context, instance string, id ingest.BatchID, failures int, next, at time.Time) error {
	err := pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		tag, err := tx.Exec(ctx, `
			UPDATE ingest.import_batches SET attempts = attempts + 1, last_attempt_at = $3
			WHERE id = $1 AND instance = $2 AND state = 'open'`,
			string(id), instance, at)
		if err != nil {
			return fmt.Errorf("counting the attempt: %w", err)
		}
		if tag.RowsAffected() == 0 {
			return ErrBatchNotOpen
		}
		if _, err := tx.Exec(ctx, `
			UPDATE ingest.connector_instances SET failures = $2, next_sync_at = $3, lease_until = NULL
			WHERE instance = $1`,
			instance, failures, next); err != nil {
			return fmt.Errorf("scheduling the retry: %w", err)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("failing batch %s of %s: %w", id, instance, err)
	}
	return nil
}

// RecordBatch records b, a completed batch of instance that is not a sync,
// such as an upload or a manual entry, in the batch history: started and
// finished now, after one attempt, with its counts. Its source must be one
// the history knows (sync, file, upload, entry), and recording an ID again
// is an error.
func (s *Store) RecordBatch(ctx context.Context, instance string, b *ingest.Batch) error {
	if _, err := s.pool.Exec(ctx, `
		INSERT INTO ingest.import_batches (id, instance, source, mode, started_at, state, attempts, last_attempt_at,
			institution, transactions, assertions, new, new_versions, unchanged, failed_assertions, problems, finished_at)
		VALUES ($1, $2, $3, $4, now(), 'completed', 1, now(), $5, $6, $7, $8, $9, $10, $11, $12, now())`,
		string(b.ID), instance, b.Source, b.Mode.String(), string(b.Institution), b.Transactions, b.Assertions,
		b.Imported.New, b.Imported.NewVersions, b.Imported.Unchanged, b.Imported.FailedAssertions, len(b.Problems)); err != nil {
		return fmt.Errorf("recording batch %s of %s: %w", b.ID, instance, err)
	}
	return nil
}

// SyncStates returns every instance's next sync, the zero time for one
// whose first attempt has not ended.
func (s *Store) SyncStates(ctx context.Context) ([]app.SyncState, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT instance, next_sync_at FROM ingest.connector_instances ORDER BY instance`)
	if err != nil {
		return nil, fmt.Errorf("loading sync states: %w", err)
	}
	states, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (app.SyncState, error) {
		var (
			st   app.SyncState
			next *time.Time
		)
		if scanErr := row.Scan(&st.Instance, &next); scanErr != nil {
			return app.SyncState{}, fmt.Errorf("scanning: %w", scanErr)
		}
		if next != nil {
			st.NextSync = *next
		}
		return st, nil
	})
	if err != nil {
		return nil, fmt.Errorf("loading sync states: %w", err)
	}
	return states, nil
}

// Bind stores a connector binding of instance to account; an existing one
// keeps its time.
func (s *Store) Bind(ctx context.Context, instance string, account ledger.AccountID, at time.Time) error {
	if _, err := s.pool.Exec(ctx, `
		INSERT INTO ingest.connector_bindings (instance, account_id, bound_at) VALUES ($1, $2, $3)
		ON CONFLICT (instance, account_id) DO NOTHING`, instance, string(account), at); err != nil {
		return fmt.Errorf("binding %s to %s: %w", instance, account, err)
	}
	return nil
}

// BoundAccounts returns the accounts instance is bound to, in ID order.
func (s *Store) BoundAccounts(ctx context.Context, instance string) ([]ledger.AccountID, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT account_id FROM ingest.connector_bindings WHERE instance = $1 ORDER BY account_id`, instance)
	if err != nil {
		return nil, fmt.Errorf("loading the bindings of %s: %w", instance, err)
	}
	ids, err := pgx.CollectRows(rows, pgx.RowTo[ledger.AccountID])
	if err != nil {
		return nil, fmt.Errorf("loading the bindings of %s: %w", instance, err)
	}
	return ids, nil
}

// BoundInstances returns the connector instances bound to account, in name
// order.
func (s *Store) BoundInstances(ctx context.Context, account ledger.AccountID) ([]string, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT instance FROM ingest.connector_bindings WHERE account_id = $1 ORDER BY instance`, string(account))
	if err != nil {
		return nil, fmt.Errorf("loading the bindings of account %s: %w", account, err)
	}
	names, err := pgx.CollectRows(rows, pgx.RowTo[string])
	if err != nil {
		return nil, fmt.Errorf("loading the bindings of account %s: %w", account, err)
	}
	return names, nil
}
