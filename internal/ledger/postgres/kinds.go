package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/arkadiuss/openfief/internal/portfolio"
)

// SaveKind creates k, or renames the stored kind with its key. It never
// changes a stored kind's default role, and a trigger refuses any update
// that would.
func (s *Store) SaveKind(ctx context.Context, k portfolio.Kind) error {
	_, err := s.db.Exec(ctx, `
		INSERT INTO ledger.sleeve_kinds (key, name, default_role) VALUES ($1, $2, $3)
		ON CONFLICT (key) DO UPDATE SET name = EXCLUDED.name`,
		string(k.Key), k.Name, k.DefaultRole.String())
	if err != nil {
		return fmt.Errorf("saving sleeve kind %s: %w", k.Key, err)
	}
	return nil
}

// Kinds returns every sleeve kind, in key order.
func (s *Store) Kinds(ctx context.Context) ([]portfolio.Kind, error) {
	rows, err := s.db.Query(ctx, `SELECT key, name, default_role FROM ledger.sleeve_kinds ORDER BY key COLLATE "C"`)
	if err != nil {
		return nil, fmt.Errorf("loading sleeve kinds: %w", err)
	}
	out, err := pgx.CollectRows(rows, scanKind)
	if err != nil {
		return nil, fmt.Errorf("loading sleeve kinds: %w", err)
	}
	return out, nil
}

// scanKind reads one sleeve kind row.
func scanKind(row pgx.CollectableRow) (portfolio.Kind, error) {
	var key, name, role string
	if err := row.Scan(&key, &name, &role); err != nil {
		return portfolio.Kind{}, fmt.Errorf("reading a sleeve kind: %w", err)
	}
	r, err := portfolio.ParseRole(role)
	if err != nil {
		return portfolio.Kind{}, fmt.Errorf("%w: sleeve kind %s: %w", ErrCorruptLog, key, err)
	}
	return portfolio.Kind{Key: portfolio.KindKey(key), Name: name, DefaultRole: r}, nil
}
