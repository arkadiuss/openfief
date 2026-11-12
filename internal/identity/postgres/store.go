package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/arkadiuss/openfief/internal/identity"
	"github.com/arkadiuss/openfief/internal/identity/app"
)

// Store is the identity context's app.Store in PostgreSQL, in the identity
// schema that Migrate creates.
type Store struct {
	pool *pgxpool.Pool
}

var _ app.Store = (*Store)(nil)

// New returns the store over pool. Run Migrate first.
func New(pool *pgxpool.Pool) *Store {
	return &Store{pool: pool}
}

// querier is what a pool and a transaction share.
type querier interface {
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

// executor runs statements in the transaction ctx carries or, without one,
// on the pool: it holds the one to use, and promotes its methods.
type executor struct {
	querier
}

// txKey is the context key of the transaction Atomically opened.
type txKey struct{}

// Atomically runs fn in one database transaction, committed when fn returns
// nil and rolled back on any error. The ctx fn receives carries the
// transaction, so every store method called with it joins it. Inside a
// transaction already, it opens a savepoint instead, so a nested call rolls
// back alone when its error is handled and commits with the outer one.
func (s *Store) Atomically(ctx context.Context, fn func(ctx context.Context) error) error {
	var begin interface {
		Begin(ctx context.Context) (pgx.Tx, error)
	} = s.pool
	if tx, ok := ctx.Value(txKey{}).(pgx.Tx); ok {
		begin = tx
	}
	err := pgx.BeginFunc(ctx, begin, func(tx pgx.Tx) error {
		return fn(context.WithValue(ctx, txKey{}, tx))
	})
	if err != nil {
		return fmt.Errorf("in a transaction: %w", err)
	}
	return nil
}

// SaveScriptToken stores a new token with its key's hash.
func (s *Store) SaveScriptToken(ctx context.Context, t *identity.ScriptToken, keyHash []byte) error {
	_, err := s.db(ctx).Exec(ctx, `
		INSERT INTO identity.script_tokens (id, name, scope, key_sha256, created_at, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6)`,
		string(t.ID), t.Name, t.Scope.String(), keyHash, t.CreatedAt, t.ExpiresAt)
	if err != nil {
		return fmt.Errorf("saving script token %s: %w", t.ID, err)
	}
	return nil
}

// rowColumns are the columns scanToken reads, in its order.
const rowColumns = `id, name, scope, created_at, expires_at, revoked_at`

// ScriptToken returns the token id and its key's hash, or an error wrapping
// app.ErrUnknownToken.
func (s *Store) ScriptToken(ctx context.Context, id identity.ScriptTokenID) (identity.ScriptToken, []byte, error) {
	var hash []byte
	tok, err := scanToken(s.db(ctx).QueryRow(ctx, `SELECT `+rowColumns+`, key_sha256 FROM identity.script_tokens WHERE id = $1`, string(id)), &hash)
	if errors.Is(err, pgx.ErrNoRows) {
		return identity.ScriptToken{}, nil, fmt.Errorf("%w: %s", app.ErrUnknownToken, id)
	}
	if err != nil {
		return identity.ScriptToken{}, nil, fmt.Errorf("loading script token %s: %w", id, err)
	}
	return tok, hash, nil
}

// RevokeScriptToken marks the token id revoked at at, keeping an earlier
// revocation's time, or fails with an error wrapping app.ErrUnknownToken.
func (s *Store) RevokeScriptToken(ctx context.Context, id identity.ScriptTokenID, at time.Time) error {
	tag, err := s.db(ctx).Exec(ctx, `
		UPDATE identity.script_tokens SET revoked_at = COALESCE(revoked_at, GREATEST($2, created_at)) WHERE id = $1`,
		string(id), at)
	if err != nil {
		return fmt.Errorf("revoking script token %s: %w", id, err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("%w: %s", app.ErrUnknownToken, id)
	}
	return nil
}

// ScriptTokens returns every token, in creation then ID order.
func (s *Store) ScriptTokens(ctx context.Context) ([]identity.ScriptToken, error) {
	rows, err := s.db(ctx).Query(ctx, `SELECT `+rowColumns+` FROM identity.script_tokens ORDER BY created_at, id`)
	if err != nil {
		return nil, fmt.Errorf("listing script tokens: %w", err)
	}
	toks, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (identity.ScriptToken, error) {
		return scanToken(row)
	})
	if err != nil {
		return nil, fmt.Errorf("listing script tokens: %w", err)
	}
	return toks, nil
}

// db returns the transaction ctx carries, which Atomically put there, or the
// pool: every method of the store runs inside the caller's transaction when
// there is one, so a state change and its audit entry commit together.
func (s *Store) db(ctx context.Context) executor {
	if tx, ok := ctx.Value(txKey{}).(pgx.Tx); ok {
		return executor{tx}
	}
	return executor{s.pool}
}

// scanToken reads a row of rowColumns, then into extra any further
// columns.
func scanToken(row pgx.Row, extra ...any) (identity.ScriptToken, error) {
	var (
		id, name, scope string
		t               identity.ScriptToken
		revoked         *time.Time
	)
	dest := append([]any{&id, &name, &scope, &t.CreatedAt, &t.ExpiresAt, &revoked}, extra...)
	if err := row.Scan(dest...); err != nil {
		return identity.ScriptToken{}, fmt.Errorf("scanning a script token: %w", err)
	}
	sc, err := identity.ParseScope(scope)
	if err != nil {
		return identity.ScriptToken{}, fmt.Errorf("script token %s: %w", id, err)
	}
	t.ID, t.Name, t.Scope = identity.ScriptTokenID(id), name, sc
	t.CreatedAt, t.ExpiresAt = t.CreatedAt.UTC(), t.ExpiresAt.UTC()
	if revoked != nil {
		t.RevokedAt = revoked.UTC()
	}
	return t, nil
}
