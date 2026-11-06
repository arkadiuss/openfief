package app

import (
	"context"
	"errors"
	"fmt"

	"github.com/arkadiuss/openfief/internal/ledger"
)

// ErrUnknownAccount reports a transaction, balance assertion or query for an
// account that is not stored.
var ErrUnknownAccount = errors.New("unknown account")

// Ledger runs the ledger context's use cases over a Store.
type Ledger struct {
	store    Store
	onImport []ImportHook
	// attention raises and clears cash checkpoints' items (D-158); nil
	// raises none.
	attention AttentionRaiser
}

// New returns the use cases over store. Every successful import calls
// onImport, in order, inside its database transaction.
func New(store Store, onImport ...ImportHook) *Ledger {
	return &Ledger{store: store, onImport: onImport}
}

// SaveAccount validates a and stores it, replacing an account with its ID.
func (l *Ledger) SaveAccount(ctx context.Context, a ledger.Account) error {
	if err := a.Validate(); err != nil {
		return fmt.Errorf("saving account: %w", err)
	}
	if err := l.store.SaveAccount(ctx, a); err != nil {
		return fmt.Errorf("saving account %s: %w", a.ID, err)
	}
	return nil
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
