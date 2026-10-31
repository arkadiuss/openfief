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
	return l.importAll(ctx, txs, assertions, importSeed)
}

// AddOpeningLots imports txs, among them opening lots that a manually
// entered account receives one at a time (D-98), with assertions. It works
// like Import, in one database transaction, but first locks each account of
// txs, in ID order, and checks its opening lots against its log
// (ledger.CheckOpeningLots): all on the account's go-live date when it is
// seeded, so its go-live never moves, and otherwise on one date with no
// transaction of the account on or before it. Each lot is keyed on its own,
// so adding one again unchanged changes nothing, and with another content
// is a new version that supersedes it (D-18).
func (l *Ledger) AddOpeningLots(ctx context.Context, txs []ledger.Transaction, assertions []ledger.BalanceAssertion) (ImportSummary, error) {
	return l.importAll(ctx, txs, assertions, importOpeningLots)
}

// guardSeed locks and decides the seeding of every account of txs in tx, in
// ID order, and returns the accounts already seeded with the same content.
// txs must be valid and their accounts stored.
func guardSeed(ctx context.Context, tx Store, txs []ledger.Transaction) ([]ledger.AccountID, error) {
	var unchanged []ledger.AccountID
	for _, id := range accountsOf(txs) {
		log, err := lockedLog(ctx, tx, id)
		if err != nil {
			return nil, err
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

// guardOpeningLots locks every account of txs in tx, in ID order, and checks
// the opening lots of each against its log (ledger.CheckOpeningLots). txs
// must be valid and their accounts stored.
func guardOpeningLots(ctx context.Context, tx Store, txs []ledger.Transaction) error {
	for _, id := range accountsOf(txs) {
		log, err := lockedLog(ctx, tx, id)
		if err != nil {
			return err
		}
		if err := ledger.CheckOpeningLots(log, id, txs); err != nil {
			return fmt.Errorf("adding opening lots to %s: %w", id, err)
		}
	}
	return nil
}

// accountsOf returns the accounts of txs, each once, in ID order.
func accountsOf(txs []ledger.Transaction) []ledger.AccountID {
	accounts := make([]ledger.AccountID, 0, len(txs))
	for i := range txs {
		accounts = append(accounts, txs[i].Account)
	}
	slices.Sort(accounts)
	return slices.Compact(accounts)
}

// lockedLog locks account id in tx and returns its log.
func lockedLog(ctx context.Context, tx Store, id ledger.AccountID) ([]ledger.VersionedTransaction, error) {
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
	return log, nil
}
