package app

import (
	"context"
	"fmt"

	"github.com/arkadiuss/openfief/internal/portfolio"
)

// SleeveDeletions deletes sleeves and kinds nothing refers to (D-120).
// *Registry implements it.
type SleeveDeletions interface {
	// DeleteSleeve deletes sleeve id unless an account is in it, which
	// fails with an error wrapping portfolio.ErrSleeveHasAccounts.
	DeleteSleeve(ctx context.Context, id portfolio.SleeveID) error
	// DeleteKind deletes kind key unless a sleeve has it, which fails with
	// an error wrapping portfolio.ErrKindInUse.
	DeleteKind(ctx context.Context, key portfolio.KindKey) error
}

// DeleteSleeve deletes the stored sleeve id when it has no accounts
// (D-120). An unknown sleeve fails with ErrNotFound, one with accounts with
// an error wrapping portfolio.ErrSleeveHasAccounts.
func (s *Sleeves) DeleteSleeve(ctx context.Context, origin, sessionToken string, id portfolio.SleeveID) error {
	if err := s.Owners.CheckOwner(ctx, origin, sessionToken); err != nil {
		return fmt.Errorf("deleting a sleeve: %w", err)
	}
	sleeves, err := s.Store.Sleeves(ctx)
	if err != nil {
		return fmt.Errorf("deleting a sleeve: %w", err)
	}
	if _, ok := sleeveOf(sleeves, id); !ok {
		return fmt.Errorf("deleting sleeve %q: %w", id, ErrNotFound)
	}
	err = audited(ctx, s.Audit, origin, sessionToken, auditSleeveDeleted, []string{"sleeve:" + string(id)}, func(ctx context.Context) error {
		return s.Deletions.DeleteSleeve(ctx, id)
	})
	if err != nil {
		return fmt.Errorf("deleting sleeve %s: %w", id, err)
	}
	return nil
}

// DeleteKind deletes the stored sleeve kind key when no sleeve has it,
// built-in kinds too (D-120). An unknown kind fails with ErrNotFound, one
// in use with an error wrapping portfolio.ErrKindInUse.
func (s *Sleeves) DeleteKind(ctx context.Context, origin, sessionToken string, key portfolio.KindKey) error {
	if err := s.Owners.CheckOwner(ctx, origin, sessionToken); err != nil {
		return fmt.Errorf("deleting a sleeve kind: %w", err)
	}
	kinds, err := s.Store.Kinds(ctx)
	if err != nil {
		return fmt.Errorf("deleting a sleeve kind: %w", err)
	}
	if _, ok := kindOf(kinds, key); !ok {
		return fmt.Errorf("deleting sleeve kind %q: %w", key, ErrNotFound)
	}
	err = audited(ctx, s.Audit, origin, sessionToken, auditSleeveKindDeleted, []string{"sleeve_kind:" + string(key)}, func(ctx context.Context) error {
		return s.Deletions.DeleteKind(ctx, key)
	})
	if err != nil {
		return fmt.Errorf("deleting sleeve kind %s: %w", key, err)
	}
	return nil
}
