package app

import (
	"context"
	"errors"
	"fmt"

	"github.com/arkadiuss/openfief/internal/ledger"
	"github.com/arkadiuss/openfief/internal/portfolio"
)

// RegistryStore stores the sleeve kinds, the sleeves and which account is
// in which sleeve, in the portfolio's own schema. Accounts are named by the
// ledger's account ID; the ledger knows nothing of sleeves.
// *portfolio/postgres.Store implements it.
type RegistryStore interface {
	// SaveKind creates k, or renames the stored kind with its key. It never
	// changes a stored kind's default role.
	SaveKind(ctx context.Context, k portfolio.Kind) error
	// Kinds returns every sleeve kind, in key order.
	Kinds(ctx context.Context) ([]portfolio.Kind, error)
	// SaveSleeve creates s or replaces the stored sleeve with its ID, with
	// its kind, role and role override. Its kind must be stored.
	SaveSleeve(ctx context.Context, s portfolio.Sleeve) error
	// Sleeves returns every sleeve, in ID order.
	Sleeves(ctx context.Context) ([]portfolio.Sleeve, error)
	// DeleteSleeve deletes sleeve id unless an account is in it, which
	// fails with an error wrapping portfolio.ErrSleeveHasAccounts.
	DeleteSleeve(ctx context.Context, id portfolio.SleeveID) error
	// DeleteKind deletes kind key unless a sleeve has it, which fails with
	// an error wrapping portfolio.ErrKindInUse.
	DeleteKind(ctx context.Context, key portfolio.KindKey) error
	// PlaceAccount puts account in sleeve, replacing the sleeve it was in.
	// The sleeve must be stored.
	PlaceAccount(ctx context.Context, account ledger.AccountID, sleeve portfolio.SleeveID) error
	// AccountSleeves returns the sleeve of every account in one.
	AccountSleeves(ctx context.Context) (map[ledger.AccountID]portfolio.SleeveID, error)
	// RemoveAccount takes account out of its sleeve; one in none stays so.
	RemoveAccount(ctx context.Context, account ledger.AccountID) error
	// InTx runs fn with a RegistryStore whose every call belongs to one
	// database transaction, committed when fn returns nil and rolled back
	// otherwise.
	InTx(ctx context.Context, fn func(tx RegistryStore) error) error
}

// AccountReader reads an account from the ledger, the context that owns
// accounts. *ledger/app.Ledger implements it.
type AccountReader interface {
	// Account returns the account id, or an error when it is not stored.
	Account(ctx context.Context, id ledger.AccountID) (ledger.Account, error)
}

// PlacementObserver is told, after the database committed it, which account
// is in which sleeve, for whoever keeps a copy of it: the authorization
// model's account and sleeve relationships (D-164, D-209). A failure is the
// observer's to repair, so the Registry returns its error wrapped in
// ErrPlacementNotObserved after the change is stored, and a retry, or the
// observer's own reconcile at start, brings it up to date. Both calls are
// idempotent. *identity/spicedb.Syncer, through cmd/core, implements it.
type PlacementObserver interface {
	// AccountPlaced reports that account is now in sleeve.
	AccountPlaced(ctx context.Context, account ledger.AccountID, sleeve portfolio.SleeveID) error
	// AccountRemoved reports that account is in no sleeve.
	AccountRemoved(ctx context.Context, account ledger.AccountID) error
}

// ErrPlacementNotObserved reports a placement or removal that is stored but
// that the PlacementObserver could not take: the change itself succeeded.
var ErrPlacementNotObserved = errors.New("stored, but not passed on to the authorization model")

// Registry keeps the sleeve kinds, the sleeves and which account is in
// which sleeve, each checked against the portfolio rules
// (portfolio.CheckKind, portfolio.CheckSleeve) before it is stored. It
// serves the owner's sleeve entries (Sleeves) and the contexts that place
// accounts in sleeves; it checks no session itself.
type Registry struct {
	store    RegistryStore
	accounts AccountReader
	observer PlacementObserver
}

// RegistryOption configures a Registry.
type RegistryOption func(*Registry)

// WithPlacementObserver has the registry tell o of every placement and
// removal of an account, after it is stored.
func WithPlacementObserver(o PlacementObserver) RegistryOption {
	return func(r *Registry) { r.observer = o }
}

// NewRegistry returns the registry over store, reading accounts through
// accounts.
func NewRegistry(store RegistryStore, accounts AccountReader, opts ...RegistryOption) *Registry {
	r := &Registry{store: store, accounts: accounts}
	for _, o := range opts {
		o(r)
	}
	return r
}

// SaveKind checks k against the stored sleeve kinds and stores it: a new
// key adds a kind, an existing one renames it. Changing a stored kind's
// default role fails with an error wrapping portfolio.ErrKindRoleChange.
func (r *Registry) SaveKind(ctx context.Context, k portfolio.Kind) error {
	kinds, err := r.store.Kinds(ctx)
	if err != nil {
		return fmt.Errorf("loading sleeve kinds: %w", err)
	}
	if err := portfolio.CheckKind(kinds, k); err != nil {
		return fmt.Errorf("saving sleeve kind: %w", err)
	}
	if err := r.store.SaveKind(ctx, k); err != nil {
		return fmt.Errorf("saving sleeve kind %s: %w", k.Key, err)
	}
	return nil
}

// Kinds returns every sleeve kind, in key order.
func (r *Registry) Kinds(ctx context.Context) ([]portfolio.Kind, error) {
	kinds, err := r.store.Kinds(ctx)
	if err != nil {
		return nil, fmt.Errorf("loading sleeve kinds: %w", err)
	}
	return kinds, nil
}

// SaveSleeve checks s against the stored sleeve kinds and stores it,
// replacing a sleeve with its ID. A sleeve with no role gets its kind's
// default role; portfolio.CheckSleeve says when a sleeve's role may differ
// from it.
func (r *Registry) SaveSleeve(ctx context.Context, s portfolio.Sleeve) error {
	kinds, err := r.store.Kinds(ctx)
	if err != nil {
		return fmt.Errorf("loading sleeve kinds: %w", err)
	}
	s, err = portfolio.CheckSleeve(kinds, s)
	if err != nil {
		return fmt.Errorf("saving sleeve: %w", err)
	}
	if err := r.store.SaveSleeve(ctx, s); err != nil {
		return fmt.Errorf("saving sleeve %s: %w", s.ID, err)
	}
	return nil
}

// Sleeves returns every sleeve, in ID order.
func (r *Registry) Sleeves(ctx context.Context) ([]portfolio.Sleeve, error) {
	s, err := r.store.Sleeves(ctx)
	if err != nil {
		return nil, fmt.Errorf("loading sleeves: %w", err)
	}
	return s, nil
}

// DeleteSleeve deletes sleeve id when no account is in it (D-120); one with
// accounts stays, with an error wrapping portfolio.ErrSleeveHasAccounts.
func (r *Registry) DeleteSleeve(ctx context.Context, id portfolio.SleeveID) error {
	if err := r.store.DeleteSleeve(ctx, id); err != nil {
		return fmt.Errorf("deleting sleeve %s: %w", id, err)
	}
	return nil
}

// DeleteKind deletes sleeve kind key when no sleeve has it, built-in kinds
// too (D-120); one in use stays, with an error wrapping portfolio.ErrKindInUse.
func (r *Registry) DeleteKind(ctx context.Context, key portfolio.KindKey) error {
	if err := r.store.DeleteKind(ctx, key); err != nil {
		return fmt.Errorf("deleting sleeve kind %s: %w", key, err)
	}
	return nil
}

// PlaceAccount puts the stored account in sleeve, replacing the sleeve it
// was in: an account is in one sleeve. An account the ledger does not
// hold fails with the ledger's error, and a sleeve that is not stored with
// an error wrapping ErrNotFound.
func (r *Registry) PlaceAccount(ctx context.Context, account ledger.AccountID, sleeve portfolio.SleeveID) error {
	if _, err := r.accounts.Account(ctx, account); err != nil {
		return fmt.Errorf("placing account %s: %w", account, err)
	}
	sleeves, err := r.store.Sleeves(ctx)
	if err != nil {
		return fmt.Errorf("placing account %s: %w", account, err)
	}
	if _, ok := sleeveOf(sleeves, sleeve); !ok {
		return fmt.Errorf("placing account %s: sleeve %q: %w", account, sleeve, ErrNotFound)
	}
	if err := r.store.PlaceAccount(ctx, account, sleeve); err != nil {
		return fmt.Errorf("placing account %s: %w", account, err)
	}
	if r.observer != nil {
		if err := r.observer.AccountPlaced(ctx, account, sleeve); err != nil {
			return fmt.Errorf("placing account %s: %w: %w", account, ErrPlacementNotObserved, err)
		}
	}
	return nil
}

// AccountSleeves returns the sleeve of every account in one; an account
// in none, such as one created before its sleeve was chosen, is left out.
func (r *Registry) AccountSleeves(ctx context.Context) (map[ledger.AccountID]portfolio.SleeveID, error) {
	m, err := r.store.AccountSleeves(ctx)
	if err != nil {
		return nil, fmt.Errorf("loading the accounts' sleeves: %w", err)
	}
	return m, nil
}

// RemoveAccount takes a deleted account out of its sleeve (D-120), so the
// sleeve can be deleted in turn. Removing one in no sleeve changes
// nothing.
func (r *Registry) RemoveAccount(ctx context.Context, account ledger.AccountID) error {
	if err := r.store.RemoveAccount(ctx, account); err != nil {
		return fmt.Errorf("removing account %s from its sleeve: %w", account, err)
	}
	if r.observer != nil {
		if err := r.observer.AccountRemoved(ctx, account); err != nil {
			return fmt.Errorf("removing account %s from its sleeve: %w: %w", account, ErrPlacementNotObserved, err)
		}
	}
	return nil
}

// EnsureDefaults stores every built-in sleeve kind that is not stored yet
// and, while no sleeve is stored at all, the default sleeves (D-71: an
// install starts with bonds, core and satellite), in one database
// transaction (portfolio.MissingDefaults). Run it at install or on every
// start: it never changes a stored kind or sleeve, so the owner's renames
// and moves are kept, and once any sleeve is stored, the owner's or the
// demo dataset's, it adds no sleeve.
func (r *Registry) EnsureDefaults(ctx context.Context) error {
	err := r.store.InTx(ctx, func(tx RegistryStore) error {
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
