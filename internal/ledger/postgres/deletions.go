package postgres

import (
	"context"
	"fmt"

	"github.com/arkadiuss/openfief/internal/ledger"
)

// DeleteAccount deletes account id in one statement unless a transaction
// version or a balance assertion check is in it (D-120), which fails with
// ledger.ErrAccountHasTransactions. Nothing of the transaction log is
// deleted: its triggers refuse any delete, and its foreign keys refuse to
// lose an account it refers to (D-18).
func (s *Store) DeleteAccount(ctx context.Context, id ledger.AccountID) error {
	var used bool
	err := s.db.QueryRow(ctx, `
		WITH used AS (
			SELECT EXISTS (SELECT 1 FROM ledger.transaction_versions WHERE account_id = $1)
				OR EXISTS (SELECT 1 FROM ledger.assertion_checks WHERE account_id = $1) AS used),
		gone AS (DELETE FROM ledger.accounts WHERE id = $1 AND NOT (SELECT used FROM used) RETURNING id)
		SELECT used FROM used`, string(id)).Scan(&used)
	if err != nil {
		return fmt.Errorf("deleting account %s: %w", id, err)
	}
	if used {
		return fmt.Errorf("account %s: %w", id, ledger.ErrAccountHasTransactions)
	}
	return nil
}

// HasTransactionsFrom reports whether a version of a transaction in account
// has a key source starting with prefix.
func (s *Store) HasTransactionsFrom(ctx context.Context, account ledger.AccountID, prefix string) (bool, error) {
	var has bool
	err := s.db.QueryRow(ctx, `
		SELECT EXISTS (SELECT 1 FROM ledger.transaction_versions
			WHERE account_id = $1 AND left(source, length($2)) = $2)`, string(account), prefix).Scan(&has)
	if err != nil {
		return false, fmt.Errorf("transactions of %s: %w", account, err)
	}
	return has, nil
}
