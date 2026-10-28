package app

import (
	"context"
	"fmt"

	"github.com/arkadiuss/openfief/internal/portfolio"
)

// SaveKind checks k against the stored sleeve kinds and stores it: a new
// key adds a kind, an existing one renames it. Changing a stored kind's
// default role fails with an error wrapping portfolio.ErrKindRoleChange.
func (l *Ledger) SaveKind(ctx context.Context, k portfolio.Kind) error {
	kinds, err := l.store.Kinds(ctx)
	if err != nil {
		return fmt.Errorf("loading sleeve kinds: %w", err)
	}
	if err := portfolio.CheckKind(kinds, k); err != nil {
		return fmt.Errorf("saving sleeve kind: %w", err)
	}
	if err := l.store.SaveKind(ctx, k); err != nil {
		return fmt.Errorf("saving sleeve kind %s: %w", k.Key, err)
	}
	return nil
}

// Kinds returns every sleeve kind, in key order.
func (l *Ledger) Kinds(ctx context.Context) ([]portfolio.Kind, error) {
	kinds, err := l.store.Kinds(ctx)
	if err != nil {
		return nil, fmt.Errorf("loading sleeve kinds: %w", err)
	}
	return kinds, nil
}

// EnsureDefaults stores every built-in sleeve kind and default sleeve that
// is not stored yet, in one database transaction. Run it at install or on
// every start: it never changes a stored kind or sleeve, so the owner's
// renames and moves are kept, and running it again stores nothing.
func (l *Ledger) EnsureDefaults(ctx context.Context) error {
	err := l.store.InTx(ctx, func(tx Store) error {
		kinds, err := tx.Kinds(ctx)
		if err != nil {
			return fmt.Errorf("loading sleeve kinds: %w", err)
		}
		sleeves, err := tx.Sleeves(ctx)
		if err != nil {
			return fmt.Errorf("loading sleeves: %w", err)
		}
		missingKinds, missingSleeves := portfolio.MissingDefaults(kinds, sleeves)
		for _, k := range missingKinds {
			if err := tx.SaveKind(ctx, k); err != nil {
				return fmt.Errorf("saving sleeve kind %s: %w", k.Key, err)
			}
		}
		for _, s := range missingSleeves {
			if err := tx.SaveSleeve(ctx, s); err != nil {
				return fmt.Errorf("saving sleeve %s: %w", s.ID, err)
			}
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("ensuring default sleeves: %w", err)
	}
	return nil
}
