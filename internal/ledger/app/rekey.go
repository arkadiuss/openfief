package app

import (
	"context"
	"errors"
	"fmt"
	"slices"

	"github.com/arkadiuss/openfief/internal/ledger"
	"github.com/arkadiuss/openfief/sdk/go/telemetry"
)

// RekeyInstrument moves every account's holding of from over to the
// instrument to (D-214):
// each current transaction that posts to from gets a new version, keyed,
// dated and worded as it is, whose postings name to instead
// (ledger.Rekeyed), so it supersedes the old one and keeps it (D-18). All
// of it is one database transaction: each account is locked, in ID order,
// and each new version is decided against its key's current version when
// it is appended, so an import that lands meanwhile is re-keyed rather
// than overwritten. Then every touched account has its balance assertions
// checked and the import hooks run, such as enqueueing a revaluation from
// the earliest date re-keyed (D-36). With nothing posting to from, it
// changes nothing, so running it again is safe. It runs in a span
// ledger.import naming the mode rekey, never the instrument.
func (l *Ledger) RekeyInstrument(ctx context.Context, from, to ledger.Instrument) error {
	if err := from.Validate(); err != nil {
		return fmt.Errorf("re-keying: %w", err)
	}
	if err := to.Validate(); err != nil {
		return fmt.Errorf("re-keying: %w", err)
	}
	var err error
	telemetry.Run(ctx, tracerScope, "ledger.import", func(ctx context.Context) error {
		err = l.rekeyAll(ctx, from, to)
		return err
	}, telemetry.KeyImport.String("rekey"))
	return err
}

// rekeyAll is RekeyInstrument without its span: one database transaction.
func (l *Ledger) rekeyAll(ctx context.Context, from, to ledger.Instrument) error {
	if err := l.store.InTx(ctx, func(tx Store) error { return l.rekeyInTx(ctx, tx, from, to) }); err != nil {
		return fmt.Errorf("re-keying %s to %s: %w", from, to, err)
	}
	return nil
}

// rekeyInTx is RekeyInstrument's work inside its database transaction tx.
func (l *Ledger) rekeyInTx(ctx context.Context, tx Store, from, to ledger.Instrument) error {
	accounts, err := tx.Accounts(ctx)
	if err != nil {
		return fmt.Errorf("loading accounts: %w", err)
	}
	var (
		sum     ImportSummary
		touched []ledger.AccountID
	)
	for i := range accounts {
		moved, err := rekeyAccount(ctx, tx, accounts[i].ID, from, to, &sum)
		if err != nil {
			return err
		}
		if moved {
			touched = append(touched, accounts[i].ID)
		}
	}
	if len(touched) == 0 {
		return nil
	}
	return l.afterRekey(ctx, tx, touched, &sum)
}

// rekeyAccount locks account id in tx and re-keys each of its current
// transactions that posts to from (rekeyOne), counting them in sum, and
// reports whether it re-keyed any.
func rekeyAccount(ctx context.Context, tx Store, id ledger.AccountID, from, to ledger.Instrument, sum *ImportSummary) (bool, error) {
	log, err := lockedLog(ctx, tx, id)
	if err != nil {
		return false, err
	}
	current, err := currentIn(log, id)
	if err != nil {
		return false, err
	}
	keys := postingKeys(current, from)
	for _, k := range keys {
		if err := rekeyOne(ctx, tx, k, from, to, sum); err != nil {
			return false, err
		}
	}
	return len(keys) > 0, nil
}

// afterRekey checks the balance assertions of each touched account and
// runs the import hooks with sum, in tx.
func (l *Ledger) afterRekey(ctx context.Context, tx Store, touched []ledger.AccountID, sum *ImportSummary) error {
	for _, id := range touched {
		checks, err := checkAccount(ctx, tx, id, nil)
		if err != nil {
			return err
		}
		sum.Checks = append(sum.Checks, checks...)
	}
	for _, hook := range l.onImport {
		if err := hook(ctx, tx, sum); err != nil {
			return fmt.Errorf("import hook: %w", err)
		}
	}
	return nil
}

// postingKeys returns the keys of txs that post to a holding of in, in key
// order.
func postingKeys(txs []ledger.Transaction, in ledger.Instrument) []ledger.TransactionKey {
	var out []ledger.TransactionKey
	for i := range txs {
		if postsToHolding(&txs[i], in) {
			out = append(out, txs[i].Key)
		}
	}
	slices.SortFunc(out, compareKeys)
	return out
}

// postsToHolding reports whether t posts to a holding of in.
func postsToHolding(t *ledger.Transaction, in ledger.Instrument) bool {
	return slices.ContainsFunc(t.Postings, func(p ledger.Posting) bool {
		return p.Book.Kind == ledger.BookHolding && p.Book.Instrument == in
	})
}

// rekeyOne appends the re-keyed version of key's current version, when that
// still posts to from, counting it in sum. On a version conflict it reloads
// and decides again, up to MaxAppendAttempts times.
func rekeyOne(ctx context.Context, store TransactionLog, key ledger.TransactionKey, from, to ledger.Instrument, sum *ImportSummary) error {
	for range MaxAppendAttempts {
		versions, err := store.Versions(ctx, key)
		if err != nil {
			return fmt.Errorf("loading versions of %s: %w", key, err)
		}
		current := slices.MaxFunc(versions, func(a, b ledger.VersionedTransaction) int { return a.Version - b.Version })
		next, moved := ledger.Rekeyed(&current.Transaction, from, to)
		if !moved {
			return nil
		}
		err = store.Append(ctx, &ledger.VersionedTransaction{Transaction: next, Version: current.Version + 1})
		if err == nil {
			sum.count(next.Account, ledger.ImportNewVersion)
			sum.affect(next.Account, next.Date)
			return nil
		}
		if !errors.Is(err, ErrVersionConflict) {
			return fmt.Errorf("re-keying %s: %w", key, err)
		}
	}
	return &VersionConflictError{Key: key, Attempts: MaxAppendAttempts}
}
