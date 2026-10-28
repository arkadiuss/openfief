package app

import (
	"context"
	"errors"
	"fmt"
	"slices"

	"github.com/arkadiuss/openfief/internal/ledger"
)

// Seed imports the seeding of accounts at go-live (D-23): their opening lots
// and opening cash, against the equity ledger account (D-33), with balance
// assertions. It works like Import, in one database transaction, but first
// locks each account it seeds, in ID order, and decides the seeding against
// that account's log (ledger.DecideSeed). An account that is already seeded
// refuses the whole seeding with an error wrapping a
// *ledger.AlreadySeededError, unless every incoming transaction of it is
// unchanged: then the seeding changes nothing for it, and the summary lists
// it in AlreadySeeded. The lock means two concurrent seedings of one
// account can't both pass the check. Seed checks only the accounts of txs,
// so it relies on keys whose source names one account, as connectors key
// them (ledger.TransactionKey).
func (l *Ledger) Seed(ctx context.Context, txs []ledger.Transaction, assertions []ledger.BalanceAssertion) (ImportSummary, error) {
	return l.importAll(ctx, txs, assertions, true)
}

// guardSeed locks and decides the seeding of every account of txs in tx, in
// ID order, and returns the accounts already seeded with the same content.
// txs must be valid and their accounts stored.
func guardSeed(ctx context.Context, tx Store, txs []ledger.Transaction) ([]ledger.AccountID, error) {
	accounts := make([]ledger.AccountID, 0, len(txs))
	for i := range txs {
		accounts = append(accounts, txs[i].Account)
	}
	slices.Sort(accounts)
	var unchanged []ledger.AccountID
	for _, id := range slices.Compact(accounts) {
		if err := tx.LockAccount(ctx, id); err != nil {
			if errors.Is(err, ErrNotFound) {
				return nil, fmt.Errorf("%w: %s", ErrUnknownAccount, id)
			}
			return nil, fmt.Errorf("locking account %s: %w", id, err)
		}
		log, err := tx.Log(ctx, id)
		if err != nil {
			return nil, fmt.Errorf("loading the log of %s: %w", id, err)
		}
		d, err := ledger.DecideSeed(log, id, txs)
		if err != nil {
			return nil, fmt.Errorf("seeding %s: %w", id, err)
		}
		if d == ledger.SeedUnchanged {
			unchanged = append(unchanged, id)
		}
	}
	return unchanged, nil
}
