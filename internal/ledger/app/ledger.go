package app

import (
	"context"
	"errors"
	"fmt"

	"github.com/arkadiuss/openfief/internal/ledger"
	"github.com/arkadiuss/openfief/internal/portfolio"
)

// ErrUnknownSleeve reports an account whose sleeve is not stored.
var ErrUnknownSleeve = errors.New("unknown sleeve")

// ErrUnknownAccount reports a transaction, balance assertion or query for an
// account that is not stored.
var ErrUnknownAccount = errors.New("unknown account")

// Ledger runs the ledger context's use cases over a Store.
type Ledger struct {
	store    Store
	onImport []ImportHook
}

// New returns the use cases over store. Every successful import calls
// onImport, in order, inside its database transaction.
func New(store Store, onImport ...ImportHook) *Ledger {
	return &Ledger{store: store, onImport: onImport}
}

// SaveSleeve checks s against the stored sleeve kinds and stores it,
// replacing a sleeve with its ID. A sleeve with no role gets its kind's
// default role; portfolio.CheckSleeve says when a sleeve's role may differ
// from it.
func (l *Ledger) SaveSleeve(ctx context.Context, s portfolio.Sleeve) error {
	kinds, err := l.store.Kinds(ctx)
	if err != nil {
		return fmt.Errorf("loading sleeve kinds: %w", err)
	}
	s, err = portfolio.CheckSleeve(kinds, s)
	if err != nil {
		return fmt.Errorf("saving sleeve: %w", err)
	}
	if err := l.store.SaveSleeve(ctx, s); err != nil {
		return fmt.Errorf("saving sleeve %s: %w", s.ID, err)
	}
	return nil
}

// SaveAccount validates a, checks that its sleeve is stored, and stores it,
// replacing an account with its ID.
func (l *Ledger) SaveAccount(ctx context.Context, a ledger.Account) error {
	if err := a.Validate(); err != nil {
		return fmt.Errorf("saving account: %w", err)
	}
	sleeves, err := l.store.Sleeves(ctx)
	if err != nil {
		return fmt.Errorf("loading sleeves: %w", err)
	}
	if !hasSleeve(sleeves, a.Sleeve) {
		return fmt.Errorf("%w %s for account %s", ErrUnknownSleeve, a.Sleeve, a.ID)
	}
	if err := l.store.SaveAccount(ctx, a); err != nil {
		return fmt.Errorf("saving account %s: %w", a.ID, err)
	}
	return nil
}

// hasSleeve reports whether sleeves include id.
func hasSleeve(sleeves []portfolio.Sleeve, id portfolio.SleeveID) bool {
	for i := range sleeves {
		if sleeves[i].ID == id {
			return true
		}
	}
	return false
}

// lookupAccount returns the account id stored in store, or an error wrapping
// ErrUnknownAccount.
func lookupAccount(ctx context.Context, store Registry, id ledger.AccountID) (ledger.Account, error) {
	a, err := store.Account(ctx, id)
	if errors.Is(err, ErrNotFound) {
		return ledger.Account{}, fmt.Errorf("%w: %s", ErrUnknownAccount, id)
	}
	if err != nil {
		return ledger.Account{}, fmt.Errorf("loading account %s: %w", id, err)
	}
	return a, nil
}
