package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/arkadiuss/openfief/internal/ledger"
	"github.com/arkadiuss/openfief/internal/ledger/app"
	"github.com/arkadiuss/openfief/internal/portfolio"
)

// Store is the ledger's app.Store in PostgreSQL, in the ledger schema that
// Migrate creates. New returns one over a pool, where each call commits on
// its own; InTx passes one whose calls all run in one transaction.
type Store struct {
	db db
	tx pgx.Tx // nil outside InTx
}

// db is what Store runs its statements on: a *pgxpool.Pool or a pgx.Tx.
// Begin on a pgx.Tx starts a savepoint.
type db interface {
	Begin(ctx context.Context) (pgx.Tx, error)
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

var _ app.Store = (*Store)(nil)

// New returns the store over pool. Run Migrate first.
func New(pool *pgxpool.Pool) *Store {
	return &Store{db: pool}
}

// InTx runs fn with a Store whose calls all run in one database transaction,
// committed when fn returns nil and rolled back otherwise. Inside InTx it runs
// fn in the same transaction.
func (s *Store) InTx(ctx context.Context, fn func(tx app.Store) error) error {
	if s.tx != nil {
		return fn(s)
	}
	err := pgx.BeginFunc(ctx, s.db, func(tx pgx.Tx) error {
		return fn(&Store{db: tx, tx: tx})
	})
	if err != nil {
		return fmt.Errorf("ledger transaction: %w", err)
	}
	return nil
}

// ErrNoTx reports a call to WithTx on a Store that InTx did not pass in.
var ErrNoTx = errors.New("not in a ledger transaction")

// WithTx runs fn on the database transaction of a Store that InTx passed in,
// for work that must commit or roll back with the ledger's writes, such as
// inserting a River job with InsertTx (D-36). Outside InTx it fails with
// ErrNoTx.
func (s *Store) WithTx(fn func(tx pgx.Tx) error) error {
	if s.tx == nil {
		return ErrNoTx
	}
	return fn(s.tx)
}

// SaveSleeve creates s or replaces the sleeve with its ID. Its kind must be
// stored.
func (s *Store) SaveSleeve(ctx context.Context, sl portfolio.Sleeve) error {
	_, err := s.db.Exec(ctx, `
		INSERT INTO ledger.sleeves (id, name, kind, role, role_override) VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (id) DO UPDATE SET name = EXCLUDED.name, kind = EXCLUDED.kind, role = EXCLUDED.role,
			role_override = EXCLUDED.role_override`,
		string(sl.ID), sl.Name, string(sl.Kind), sl.Role.String(), sl.RoleOverride)
	if err != nil {
		return fmt.Errorf("saving sleeve %s: %w", sl.ID, err)
	}
	return nil
}

// Sleeves returns every sleeve, in ID order.
func (s *Store) Sleeves(ctx context.Context) ([]portfolio.Sleeve, error) {
	rows, err := s.db.Query(ctx, `
		SELECT id, name, kind, role, role_override FROM ledger.sleeves ORDER BY id COLLATE "C"`)
	if err != nil {
		return nil, fmt.Errorf("loading sleeves: %w", err)
	}
	out, err := pgx.CollectRows(rows, scanSleeve)
	if err != nil {
		return nil, fmt.Errorf("loading sleeves: %w", err)
	}
	return out, nil
}

// scanSleeve reads one sleeve row.
func scanSleeve(row pgx.CollectableRow) (portfolio.Sleeve, error) {
	var id, name, kind, role string
	var override bool
	if err := row.Scan(&id, &name, &kind, &role, &override); err != nil {
		return portfolio.Sleeve{}, fmt.Errorf("reading a sleeve: %w", err)
	}
	r, err := portfolio.ParseRole(role)
	if err != nil {
		return portfolio.Sleeve{}, fmt.Errorf("%w: sleeve %s: %w", ErrCorruptLog, id, err)
	}
	return portfolio.Sleeve{
		ID: portfolio.SleeveID(id), Name: name, Kind: portfolio.KindKey(kind), Role: r, RoleOverride: override,
	}, nil
}

// SaveAccount creates a or replaces the account with its ID. Its sleeve must
// be stored.
func (s *Store) SaveAccount(ctx context.Context, a ledger.Account) error {
	_, err := s.db.Exec(ctx, `
		INSERT INTO ledger.accounts (id, institution, name, sleeve_id) VALUES ($1, $2, $3, $4)
		ON CONFLICT (id) DO UPDATE SET institution = EXCLUDED.institution, name = EXCLUDED.name, sleeve_id = EXCLUDED.sleeve_id`,
		string(a.ID), a.Institution, a.Name, string(a.Sleeve))
	if err != nil {
		return fmt.Errorf("saving account %s: %w", a.ID, err)
	}
	return nil
}

// selectAccounts lists accounts; callers add WHERE and ORDER BY.
const selectAccounts = `SELECT id, institution, name, sleeve_id FROM ledger.accounts`

// Account returns the account with id, or an error wrapping app.ErrNotFound.
func (s *Store) Account(ctx context.Context, id ledger.AccountID) (ledger.Account, error) {
	rows, err := s.db.Query(ctx, selectAccounts+` WHERE id = $1`, string(id))
	if err != nil {
		return ledger.Account{}, fmt.Errorf("loading account %s: %w", id, err)
	}
	a, err := pgx.CollectExactlyOneRow(rows, scanAccount)
	if errors.Is(err, pgx.ErrNoRows) {
		return ledger.Account{}, fmt.Errorf("account %s: %w", id, app.ErrNotFound)
	}
	if err != nil {
		return ledger.Account{}, fmt.Errorf("loading account %s: %w", id, err)
	}
	return a, nil
}

// LockAccount locks the account's row until the transaction ends, with FOR
// NO KEY UPDATE: two seedings of the account wait for each other, while
// imports that only reference it don't. Outside InTx it fails with ErrNoTx;
// for an account that is not stored, with an error wrapping app.ErrNotFound.
func (s *Store) LockAccount(ctx context.Context, id ledger.AccountID) error {
	if s.tx == nil {
		return fmt.Errorf("locking account %s: %w", id, ErrNoTx)
	}
	tag, err := s.db.Exec(ctx, `SELECT 1 FROM ledger.accounts WHERE id = $1 FOR NO KEY UPDATE`, string(id))
	if err != nil {
		return fmt.Errorf("locking account %s: %w", id, err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("account %s: %w", id, app.ErrNotFound)
	}
	return nil
}

// Accounts returns every account, in ID order.
func (s *Store) Accounts(ctx context.Context) ([]ledger.Account, error) {
	rows, err := s.db.Query(ctx, selectAccounts+` ORDER BY id COLLATE "C"`)
	if err != nil {
		return nil, fmt.Errorf("loading accounts: %w", err)
	}
	out, err := pgx.CollectRows(rows, scanAccount)
	if err != nil {
		return nil, fmt.Errorf("loading accounts: %w", err)
	}
	return out, nil
}

// scanAccount reads one row of selectAccounts.
func scanAccount(row pgx.CollectableRow) (ledger.Account, error) {
	var id, institution, name, sleeve string
	if err := row.Scan(&id, &institution, &name, &sleeve); err != nil {
		return ledger.Account{}, fmt.Errorf("reading account: %w", err)
	}
	return ledger.Account{
		ID: ledger.AccountID(id), Institution: institution, Name: name, Sleeve: portfolio.SleeveID(sleeve),
	}, nil
}
