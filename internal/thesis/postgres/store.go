package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/arkadiuss/openfief/internal/thesis"
	"github.com/arkadiuss/openfief/internal/thesis/app"
)

// ErrCorrupt reports a stored row the thesis context can't read back.
var ErrCorrupt = errors.New("corrupt thesis row")

var _ app.Store = (*Store)(nil)

// Store keeps the draft theses in the thesis schema. It implements app.Store.
type Store struct {
	pool DB
}

// DB is what Store runs its statements on: a *pgxpool.Pool, or anything that
// answers as one, such as the core's pool that joins the transaction a context
// carries.
type DB interface {
	Begin(ctx context.Context) (pgx.Tx, error)
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

// Joiner is a DB that can carry a transaction in a context, so that what runs
// with that context joins it: the store hands the context of a change's
// transaction to the hook that runs before its commit, which enqueues the
// change's audit entry in it (D-211).
type Joiner interface {
	Join(ctx context.Context, tx pgx.Tx) context.Context
}

// New returns the store over pool. Run Migrate first.
func New(pool DB) *Store { return &Store{pool: pool} }

// uniqueViolation is the SQLSTATE of a primary key that is taken.
const uniqueViolation = "23505"

// Save stores draft as version base+1 of thesis id, within the transaction
// that runs within before it commits. A create (base 0) finds the id taken as
// thesis.ErrStale; an edit whose base is not the latest version is
// thesis.ErrStale too, and one of a thesis with no version thesis.ErrNotFound.
func (s *Store) Save(ctx context.Context, id string, base int, at time.Time, d *thesis.Draft, within func(context.Context) error) error {
	err := pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		if base > 0 {
			if err := latestIs(ctx, tx, id, base); err != nil {
				return err
			}
		}
		_, err := tx.Exec(ctx, `
			INSERT INTO thesis.thesis (thesis_id, version, saved_at, statement, horizon, instruments, invalidation, sources)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
			id, base+1, at, d.Statement, string(d.Horizon), d.Instruments, d.Invalidation, d.Sources)
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok && pgErr.Code == uniqueViolation {
			return thesis.ErrStale
		}
		if err != nil {
			return fmt.Errorf("writing a version: %w", err)
		}
		if within == nil {
			return nil
		}
		if j, ok := s.pool.(Joiner); ok {
			return within(j.Join(ctx, tx))
		}
		return within(ctx)
	})
	if err != nil {
		return fmt.Errorf("saving a draft thesis in one transaction: %w", err)
	}
	return nil
}

// latestIs checks, in tx, that base is the latest version of thesis id.
func latestIs(ctx context.Context, tx pgx.Tx, id string, base int) error {
	var latest *int32
	err := tx.QueryRow(ctx, "SELECT max(version) FROM thesis.thesis WHERE thesis_id = $1", id).Scan(&latest)
	if err != nil {
		return fmt.Errorf("reading the latest version: %w", err)
	}
	switch {
	case latest == nil:
		return thesis.ErrNotFound
	case int(*latest) != base:
		return thesis.ErrStale
	}
	return nil
}

// versionColumns are the columns of a version, in the order scanVersion reads.
const versionColumns = "version, saved_at, statement, horizon, instruments, invalidation, sources"

// scanVersion reads one version.
func scanVersion(row pgx.Row) (thesis.Version, error) {
	var (
		v       thesis.Version
		version int32
		horizon string
	)
	err := row.Scan(&version, &v.SavedAt, &v.Draft.Statement, &horizon, &v.Draft.Instruments, &v.Draft.Invalidation, &v.Draft.Sources)
	if err != nil {
		return thesis.Version{}, fmt.Errorf("scanning a version: %w", err)
	}
	h, err := thesis.ParseHorizon(horizon)
	if err != nil {
		return thesis.Version{}, fmt.Errorf("%w: %w", ErrCorrupt, err)
	}
	v.Number, v.Draft.Horizon, v.SavedAt = int(version), h, v.SavedAt.UTC()
	return v, nil
}

// Get returns the thesis id with every version, oldest first, or an error
// wrapping thesis.ErrNotFound.
func (s *Store) Get(ctx context.Context, id string) (thesis.Thesis, error) {
	rows, err := s.pool.Query(ctx, "SELECT "+versionColumns+" FROM thesis.thesis WHERE thesis_id = $1 ORDER BY version", id)
	if err != nil {
		return thesis.Thesis{}, fmt.Errorf("reading a draft thesis: %w", err)
	}
	versions, err := pgx.CollectRows(rows, func(r pgx.CollectableRow) (thesis.Version, error) { return scanVersion(r) })
	if err != nil {
		return thesis.Thesis{}, fmt.Errorf("reading a draft thesis: %w", err)
	}
	if len(versions) == 0 {
		return thesis.Thesis{}, thesis.ErrNotFound
	}
	return thesis.Thesis{ID: id, Versions: versions}, nil
}

// List returns every thesis as its latest version, with when its first one
// was saved, the most recently changed first.
func (s *Store) List(ctx context.Context) ([]thesis.Summary, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT t.thesis_id, f.saved_at, t.version, t.saved_at, t.statement, t.horizon, t.instruments, t.invalidation, t.sources
		FROM thesis.thesis t
		JOIN thesis.thesis f ON f.thesis_id = t.thesis_id AND f.version = 1
		WHERE t.version = (SELECT max(version) FROM thesis.thesis m WHERE m.thesis_id = t.thesis_id)
		ORDER BY t.saved_at DESC, t.thesis_id`)
	if err != nil {
		return nil, fmt.Errorf("listing the draft theses: %w", err)
	}
	out, err := pgx.CollectRows(rows, func(r pgx.CollectableRow) (thesis.Summary, error) {
		var (
			sum     thesis.Summary
			version int32
			horizon string
		)
		scanErr := r.Scan(&sum.ID, &sum.CreatedAt, &version, &sum.UpdatedAt, &sum.Draft.Statement, &horizon,
			&sum.Draft.Instruments, &sum.Draft.Invalidation, &sum.Draft.Sources)
		if scanErr != nil {
			return thesis.Summary{}, fmt.Errorf("scanning a thesis: %w", scanErr)
		}
		h, parseErr := thesis.ParseHorizon(horizon)
		if parseErr != nil {
			return thesis.Summary{}, fmt.Errorf("%w: %w", ErrCorrupt, parseErr)
		}
		sum.Version, sum.Draft.Horizon = int(version), h
		sum.CreatedAt, sum.UpdatedAt = sum.CreatedAt.UTC(), sum.UpdatedAt.UTC()
		return sum, nil
	})
	if err != nil {
		return nil, fmt.Errorf("listing the draft theses: %w", err)
	}
	return out, nil
}
