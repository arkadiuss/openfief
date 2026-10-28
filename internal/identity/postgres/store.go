package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
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

// SaveScriptToken stores a new token with its key's hash.
func (s *Store) SaveScriptToken(ctx context.Context, t *identity.ScriptToken, keyHash []byte) error {
	_, err := s.pool.Exec(ctx, `
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
	rows, err := s.pool.Query(ctx, `SELECT `+rowColumns+`, key_sha256 FROM identity.script_tokens WHERE id = $1`, string(id))
	if err != nil {
		return identity.ScriptToken{}, nil, fmt.Errorf("loading script token %s: %w", id, err)
	}
	var hash []byte
	tok, err := pgx.CollectExactlyOneRow(rows, func(row pgx.CollectableRow) (identity.ScriptToken, error) {
		return scanToken(row, &hash)
	})
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
	tag, err := s.pool.Exec(ctx, `
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
	rows, err := s.pool.Query(ctx, `SELECT `+rowColumns+` FROM identity.script_tokens ORDER BY created_at, id`)
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

// scanToken reads a row of rowColumns, then into extra any further
// columns.
func scanToken(row pgx.CollectableRow, extra ...any) (identity.ScriptToken, error) {
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
