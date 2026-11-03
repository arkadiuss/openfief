package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/arkadiuss/openfief/internal/ledger"
	"github.com/arkadiuss/openfief/internal/portfolio"
	"github.com/arkadiuss/openfief/internal/portfolio/app"
)

var _ app.RegistryStore = (*Store)(nil)

// SaveKind creates k, or renames the stored kind with its key. It never
// changes a stored kind's default role, and a trigger refuses any update
// that would.
func (s *Store) SaveKind(ctx context.Context, k portfolio.Kind) error {
	_, err := s.db.Exec(ctx, `
		INSERT INTO portfolio.sleeve_kinds (key, name, default_role) VALUES ($1, $2, $3)
		ON CONFLICT (key) DO UPDATE SET name = EXCLUDED.name`,
		string(k.Key), k.Name, k.DefaultRole.String())
	if err != nil {
		return fmt.Errorf("saving sleeve kind %s: %w", k.Key, err)
	}
	return nil
}

// Kinds returns every sleeve kind, in key order.
func (s *Store) Kinds(ctx context.Context) ([]portfolio.Kind, error) {
	rows, err := s.db.Query(ctx, `SELECT key, name, default_role FROM portfolio.sleeve_kinds ORDER BY key COLLATE "C"`)
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
		return portfolio.Kind{}, fmt.Errorf("%w: sleeve kind %s: %w", ErrCorrupt, key, err)
	}
	return portfolio.Kind{Key: portfolio.KindKey(key), Name: name, DefaultRole: r}, nil
}

// SaveSleeve creates sl or replaces the sleeve with its ID. Its kind must
// be stored.
func (s *Store) SaveSleeve(ctx context.Context, sl portfolio.Sleeve) error {
	_, err := s.db.Exec(ctx, `
		INSERT INTO portfolio.sleeves (id, name, kind, role, role_override) VALUES ($1, $2, $3, $4, $5)
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
		SELECT id, name, kind, role, role_override FROM portfolio.sleeves ORDER BY id COLLATE "C"`)
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
		return portfolio.Sleeve{}, fmt.Errorf("%w: sleeve %s: %w", ErrCorrupt, id, err)
	}
	return portfolio.Sleeve{
		ID: portfolio.SleeveID(id), Name: name, Kind: portfolio.KindKey(kind), Role: r, RoleOverride: override,
	}, nil
}

// DeleteSleeve deletes sleeve id in one statement unless an account is in
// it, which fails with portfolio.ErrSleeveHasAccounts.
func (s *Store) DeleteSleeve(ctx context.Context, id portfolio.SleeveID) error {
	var used bool
	err := s.db.QueryRow(ctx, `
		WITH used AS (SELECT EXISTS (SELECT 1 FROM portfolio.account_sleeves WHERE sleeve_id = $1) AS used),
		gone AS (DELETE FROM portfolio.sleeves WHERE id = $1 AND NOT (SELECT used FROM used) RETURNING id)
		SELECT used FROM used`, string(id)).Scan(&used)
	if err != nil {
		return fmt.Errorf("deleting sleeve %s: %w", id, err)
	}
	if used {
		return fmt.Errorf("sleeve %s: %w", id, portfolio.ErrSleeveHasAccounts)
	}
	return nil
}

// DeleteKind deletes sleeve kind key in one statement unless a sleeve has
// it, which fails with portfolio.ErrKindInUse.
func (s *Store) DeleteKind(ctx context.Context, key portfolio.KindKey) error {
	var used bool
	err := s.db.QueryRow(ctx, `
		WITH used AS (SELECT EXISTS (SELECT 1 FROM portfolio.sleeves WHERE kind = $1) AS used),
		gone AS (DELETE FROM portfolio.sleeve_kinds WHERE key = $1 AND NOT (SELECT used FROM used) RETURNING key)
		SELECT used FROM used`, string(key)).Scan(&used)
	if err != nil {
		return fmt.Errorf("deleting sleeve kind %s: %w", key, err)
	}
	if used {
		return fmt.Errorf("sleeve kind %s: %w", key, portfolio.ErrKindInUse)
	}
	return nil
}

// PlaceAccount puts account in sleeve, replacing the sleeve it was in. The
// sleeve's foreign key refuses one that is not stored.
func (s *Store) PlaceAccount(ctx context.Context, account ledger.AccountID, sleeve portfolio.SleeveID) error {
	_, err := s.db.Exec(ctx, `
		INSERT INTO portfolio.account_sleeves (account_id, sleeve_id) VALUES ($1, $2)
		ON CONFLICT (account_id) DO UPDATE SET sleeve_id = EXCLUDED.sleeve_id`,
		string(account), string(sleeve))
	if err != nil {
		return fmt.Errorf("placing account %s in sleeve %s: %w", account, sleeve, err)
	}
	return nil
}

// AccountSleeves returns the sleeve of every account in one.
func (s *Store) AccountSleeves(ctx context.Context) (map[ledger.AccountID]portfolio.SleeveID, error) {
	rows, err := s.db.Query(ctx, `SELECT account_id, sleeve_id FROM portfolio.account_sleeves`)
	if err != nil {
		return nil, fmt.Errorf("loading the accounts' sleeves: %w", err)
	}
	out := map[ledger.AccountID]portfolio.SleeveID{}
	var account, sleeve string
	_, err = pgx.ForEachRow(rows, []any{&account, &sleeve}, func() error {
		out[ledger.AccountID(account)] = portfolio.SleeveID(sleeve)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("loading the accounts' sleeves: %w", err)
	}
	return out, nil
}

// RemoveAccount takes account out of its sleeve.
func (s *Store) RemoveAccount(ctx context.Context, account ledger.AccountID) error {
	if _, err := s.db.Exec(ctx, `DELETE FROM portfolio.account_sleeves WHERE account_id = $1`, string(account)); err != nil {
		return fmt.Errorf("removing account %s from its sleeve: %w", account, err)
	}
	return nil
}
