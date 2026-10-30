package app

import (
	"context"
	"errors"
	"fmt"

	"github.com/arkadiuss/openfief/internal/ledger"
)

// Errors of the owner's entries.
var (
	// ErrNoSession reports a session token that is missing, unknown,
	// expired or revoked.
	ErrNoSession = errors.New("no valid session")
	// ErrNotOwner reports a valid session without owner scope, or an origin
	// the core does not serve the owner on.
	ErrNotOwner = errors.New("only the owner may do this")
	// ErrUnknownTransaction reports a transfer leg whose key has no stored
	// transaction.
	ErrUnknownTransaction = errors.New("unknown transaction")
)

// Owners checks that a caller is the owner (D-79).
type Owners interface {
	// CheckOwner returns nil for an owner-scoped session token valid on
	// origin, an error wrapping ErrNoSession or ErrNotOwner when it is not,
	// and any other error when it can't tell.
	CheckOwner(ctx context.Context, origin, sessionToken string) error
}

// TransferLinks returns the current link of each pair the owner decided on
// (D-105), in withdrawal then deposit key order.
func (l *Ledger) TransferLinks(ctx context.Context) ([]ledger.TransferLink, error) {
	return currentLinks(ctx, l.store)
}

// DecideTransfer records decision on pair (D-105) in one database
// transaction and returns the pair's current link. Both legs must be stored
// and their current versions one possible transfer (ledger.CheckTransfer);
// a leg that is not stored fails with ErrUnknownTransaction. A decision
// equal to the pair's current one appends nothing; any other appends the
// pair's next version, which supersedes the earlier ones (D-18).
// Confirming a leg of another confirmed transfer fails with
// ledger.ErrTransferTaken.
func (l *Ledger) DecideTransfer(ctx context.Context, pair ledger.TransferPair, decision ledger.TransferDecision) (ledger.TransferLink, error) {
	var out ledger.TransferLink
	err := l.store.InTx(ctx, func(tx Store) error {
		var err error
		out, err = decideIn(ctx, tx, pair, decision)
		return err
	})
	if err != nil {
		return ledger.TransferLink{}, fmt.Errorf("deciding transfer %s: %w", pair, err)
	}
	return out, nil
}

// Transfers are the owner's decisions on suggested transfers (D-105), for
// an owner-scoped session only (D-79).
type Transfers struct {
	Owners Owners
	Ledger TransferDecider
}

// TransferDecider records decisions on transfers; *Ledger implements it.
type TransferDecider interface {
	// DecideTransfer records decision on pair and returns the pair's
	// current link (Ledger.DecideTransfer).
	DecideTransfer(ctx context.Context, pair ledger.TransferPair, decision ledger.TransferDecision) (ledger.TransferLink, error)
}

// Decide records the owner's decision on pair: confirmed, the two legs are
// one transfer between the owner's own accounts and leave the
// time-weighted return's cash flows; rejected, they stay cash flows and the
// pair is not suggested again. Either supersedes an earlier decision on the
// pair. It returns the pair's current link.
func (t *Transfers) Decide(ctx context.Context, origin, sessionToken string, pair ledger.TransferPair, decision ledger.TransferDecision) (ledger.TransferLink, error) {
	if err := t.Owners.CheckOwner(ctx, origin, sessionToken); err != nil {
		return ledger.TransferLink{}, fmt.Errorf("deciding a transfer: %w", err)
	}
	link, err := t.Ledger.DecideTransfer(ctx, pair, decision)
	if err != nil {
		return ledger.TransferLink{}, fmt.Errorf("deciding a transfer: %w", err)
	}
	return link, nil
}

// decideIn decides pair on tx, a Store that InTx passed in, holding the
// transfer links' lock while it checks which legs are taken and appends.
func decideIn(ctx context.Context, tx Store, pair ledger.TransferPair, decision ledger.TransferDecision) (ledger.TransferLink, error) {
	if err := tx.LockTransferLinks(ctx); err != nil {
		return ledger.TransferLink{}, fmt.Errorf("locking transfer links: %w", err)
	}
	if err := checkLegs(ctx, tx, pair); err != nil {
		return ledger.TransferLink{}, err
	}
	current, err := currentLinks(ctx, tx)
	if err != nil {
		return ledger.TransferLink{}, err
	}
	link, changed, err := ledger.DecideTransferLink(current, pair, decision)
	if err != nil {
		return ledger.TransferLink{}, fmt.Errorf("deciding: %w", err)
	}
	if !changed {
		return link, nil
	}
	if err := tx.AppendTransferLink(ctx, &link); err != nil {
		return ledger.TransferLink{}, fmt.Errorf("storing: %w", err)
	}
	return link, nil
}

// checkLegs checks that the current versions of pair's legs are one
// possible transfer (ledger.CheckTransfer).
func checkLegs(ctx context.Context, store TransactionLog, pair ledger.TransferPair) error {
	w, err := currentByKey(ctx, store, pair.Withdrawal)
	if err != nil {
		return err
	}
	d, err := currentByKey(ctx, store, pair.Deposit)
	if err != nil {
		return err
	}
	if err := ledger.CheckTransfer(&w, &d); err != nil {
		return fmt.Errorf("checking the legs: %w", err)
	}
	return nil
}

// currentLinks returns the current version of each stored pair's link.
func currentLinks(ctx context.Context, store TransferLinkStore) ([]ledger.TransferLink, error) {
	all, err := store.TransferLinks(ctx)
	if err != nil {
		return nil, fmt.Errorf("loading transfer links: %w", err)
	}
	current, err := ledger.CurrentTransferLinks(all)
	if err != nil {
		return nil, fmt.Errorf("transfer links: %w", err)
	}
	return current, nil
}

// currentByKey returns the current version of the transaction key, or an
// error wrapping ErrUnknownTransaction when none is stored.
func currentByKey(ctx context.Context, store TransactionLog, key ledger.TransactionKey) (ledger.Transaction, error) {
	versions, err := store.Versions(ctx, key)
	if err != nil {
		return ledger.Transaction{}, fmt.Errorf("loading %s: %w", key, err)
	}
	current, err := ledger.CurrentTransactions(versions)
	if err != nil {
		return ledger.Transaction{}, fmt.Errorf("transaction %s: %w", key, err)
	}
	if len(current) != 1 {
		return ledger.Transaction{}, fmt.Errorf("transaction %s: %w", key, ErrUnknownTransaction)
	}
	return current[0], nil
}
