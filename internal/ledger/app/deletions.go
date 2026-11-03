package app

import (
	"context"
	"fmt"

	"github.com/arkadiuss/openfief/internal/ledger"
)

// DeletionStore deletes registry entries nothing refers to (D-120). Each
// delete checks and deletes in one statement, so a write that refers to the
// entry can't slip in between, and deleting what isn't stored changes
// nothing.
type DeletionStore interface {
	// DeleteAccount deletes account id unless a transaction version or a
	// balance assertion is in it, which fails with an error wrapping
	// ledger.ErrAccountHasTransactions.
	DeleteAccount(ctx context.Context, id ledger.AccountID) error
	// HasTransactionsFrom reports whether any version of a transaction in
	// account has a key source starting with prefix; an empty prefix asks
	// for any transaction.
	HasTransactionsFrom(ctx context.Context, account ledger.AccountID, prefix string) (bool, error)
}

// DeleteAccount deletes account id when it holds no transactions (D-120);
// one that holds any stays, with an error wrapping
// ledger.ErrAccountHasTransactions. Deleting an account not stored changes
// nothing.
func (l *Ledger) DeleteAccount(ctx context.Context, id ledger.AccountID) error {
	if err := l.store.DeleteAccount(ctx, id); err != nil {
		return fmt.Errorf("deleting account %s: %w", id, err)
	}
	return nil
}

// HasTransactionsFrom reports whether account holds any transaction, of any
// version, whose key source starts with prefix, such as a connector
// instance's scoped sources; an empty prefix asks for any.
func (l *Ledger) HasTransactionsFrom(ctx context.Context, account ledger.AccountID, prefix string) (bool, error) {
	has, err := l.store.HasTransactionsFrom(ctx, account, prefix)
	if err != nil {
		return false, fmt.Errorf("transactions of %s: %w", account, err)
	}
	return has, nil
}
