package app

import (
	"context"
	"errors"
	"fmt"

	"github.com/arkadiuss/openfief/internal/ingest"
	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
)

// Errors an Owners implementation wraps, so the use cases can tell a caller
// who must log in again from one who may not do this.
var (
	// ErrNoSession reports a session token that is missing, unknown,
	// expired or revoked.
	ErrNoSession = errors.New("no valid session")
	// ErrNotOwner reports a valid session without owner scope, or an origin
	// the core does not serve the owner on.
	ErrNotOwner = errors.New("only the owner may do this")
)

// ErrNoConnector reports an account that has no connector instance to
// hand a bond lot or a statement file to: none is bound to it, or more than
// one is, so the core can't tell which is the right one.
var ErrNoConnector = errors.New("no single connector instance is bound to the account")

// Owners checks that a session is the owner's, with owner scope, on the
// origin the browser uses (author decision, auth split), and which accounts
// it may change, as the core's authorizer answers (D-163). cmd/core
// implements it over the identity context.
type Owners interface {
	// CheckOwner returns nil for an owner-scoped session token valid on
	// origin, an error wrapping ErrNoSession or ErrNotOwner when it is not,
	// and any other error when it can't tell.
	CheckOwner(ctx context.Context, origin, sessionToken string) error
	// Changeable returns those of accounts the owner-scoped session token
	// valid on origin may change, in their order; it fails as CheckOwner
	// does.
	Changeable(ctx context.Context, origin, sessionToken string, accounts []ledger.AccountID) ([]ledger.AccountID, error)
}

// AccountBindings finds the connector instances bound to an account.
// *ingest/postgres.Store implements it.
type AccountBindings interface {
	BoundInstances(ctx context.Context, account ledger.AccountID) ([]string, error)
}

// BatchRecorder records a completed import batch of a connector instance
// that is not a sync, such as an upload or a manual entry, in the import
// batch history; syncs are recorded by their SyncStore. Recording a batch
// ID again is an error. *ingest/postgres.Store implements it.
type BatchRecorder interface {
	RecordBatch(ctx context.Context, instance string, b *ingest.Batch) error
}

// AccountTransactions reads an account's current transactions;
// ledger/app.Ledger implements it.
type AccountTransactions interface {
	CurrentTransactions(ctx context.Context, account ledger.AccountID) ([]ledger.Transaction, error)
}

// BondLots takes treasury bond lots the owner enters by hand in the
// built-in UI (D-9) and hands each to the connector instance bound to its
// account, which maps it; the core imports what it delivers like any
// other batch (D-49). Nothing else writes a lot into the ledger.
type BondLots struct {
	Owners     Owners
	Bindings   AccountBindings
	Connectors *Connectors
	// Batches records each completed batch in the import batch history.
	Batches BatchRecorder
	// Booked reads the account's current transactions, from which Delete
	// tells the connector what it booked there (D-154); nil tells it
	// nothing, so only the lot's own transactions are deleted.
	Booked AccountTransactions
	// Today returns the date a lot may be bought on at the latest.
	Today func() kernel.Date
	// NewBatchID returns a fresh import batch ID.
	NewBatchID func() ingest.BatchID
}

// Add checks that the caller is the owner (an owner-scoped session on
// origin) and that lot is valid as of today, then hands it to the one
// connector instance bound to account as a new import batch, and returns
// the completed batch, after recording it in the import batch history.
// Errors wrap ErrNoSession or ErrNotOwner for a caller
// who may not, ingest.ErrInvalidBondLot for a lot that breaks a rule,
// ErrNoConnector for an account without exactly one bound instance,
// ErrEntryRefused or ErrNoEntries for a connector that won't take it,
// ErrConnectorUnavailable for one that can't be reached, and
// ingest.ErrAccountNotBound for a delivery into another account.
// Entering the same lot again imports nothing new.
func (b *BondLots) Add(ctx context.Context, origin, sessionToken string, account ledger.AccountID, lot *ingest.BondLot) (ingest.Batch, error) {
	if err := mayChange(ctx, b.Owners, origin, sessionToken, account); err != nil {
		return ingest.Batch{}, fmt.Errorf("adding a bond lot: %w", err)
	}
	if err := lot.Validate(b.Today()); err != nil {
		return ingest.Batch{}, fmt.Errorf("adding a bond lot: %w", err)
	}
	instance, err := boundInstance(ctx, b.Bindings, account)
	if err != nil {
		return ingest.Batch{}, fmt.Errorf("adding a bond lot: %w", err)
	}
	batch, err := b.Connectors.ImportBondLots(ctx, instance, b.NewBatchID(), account, &ingest.BondLotEntries{Lots: []ingest.BondLot{*lot}})
	if err != nil {
		return ingest.Batch{}, fmt.Errorf("adding a bond lot: %w", err)
	}
	if err := b.Batches.RecordBatch(ctx, instance, &batch); err != nil {
		return ingest.Batch{}, fmt.Errorf("adding a bond lot: %w", err)
	}
	return batch, nil
}

// Delete deletes the lot of account that key names, as a correction
// (D-100), for the owner whose session sessionToken is, on origin: it hands
// the key to the one connector instance bound to account, with the
// account's current transactions that instance delivered (ingest.BookedBy),
// and the connector delivers a deletion of each of the lot's transactions
// and of each one it booked for the lot since, such as its coupons, so
// their cash goes too (D-154); Delete imports them as a new transactions
// batch. Each deletion supersedes its transaction, which the ledger keeps
// (D-18); deleting it again changes nothing. A key without
// a series or date wraps ingest.ErrInvalidBondLot, and a lot never entered
// wraps ledger.ErrNothingToDelete; otherwise Delete fails as Add does.
func (b *BondLots) Delete(ctx context.Context, origin, sessionToken string, account ledger.AccountID, key ingest.BondLotKey) (ingest.Batch, error) {
	if err := mayChange(ctx, b.Owners, origin, sessionToken, account); err != nil {
		return ingest.Batch{}, fmt.Errorf("deleting a bond lot: %w", err)
	}
	if err := key.Validate(); err != nil {
		return ingest.Batch{}, fmt.Errorf("deleting a bond lot: %w", err)
	}
	instance, err := boundInstance(ctx, b.Bindings, account)
	if err != nil {
		return ingest.Batch{}, fmt.Errorf("deleting a bond lot: %w", err)
	}
	booked, err := b.booked(ctx, instance, account)
	if err != nil {
		return ingest.Batch{}, fmt.Errorf("deleting a bond lot: %w", err)
	}
	entries := &ingest.BondLotEntries{Deleted: []ingest.BondLotKey{key}, Booked: booked}
	batch, err := b.Connectors.ImportBondLots(ctx, instance, b.NewBatchID(), account, entries)
	if err != nil {
		return ingest.Batch{}, fmt.Errorf("deleting a bond lot: %w", err)
	}
	if err := b.Batches.RecordBatch(ctx, instance, &batch); err != nil {
		return ingest.Batch{}, fmt.Errorf("deleting a bond lot: %w", err)
	}
	return batch, nil
}

// booked returns the current transactions of account that instance
// delivered, keyed as it keyed them, or none without Booked.
func (b *BondLots) booked(ctx context.Context, instance string, account ledger.AccountID) ([]ledger.TransactionHeader, error) {
	if b.Booked == nil {
		return nil, nil
	}
	current, err := b.Booked.CurrentTransactions(ctx, account)
	if err != nil {
		return nil, fmt.Errorf("transactions of %s: %w", account, err)
	}
	return ingest.BookedBy(instance, current), nil
}

// boundInstance returns the one connector instance bound to account, or an
// error wrapping ErrNoConnector when none or several are.
func boundInstance(ctx context.Context, bindings AccountBindings, account ledger.AccountID) (string, error) {
	instances, err := bindings.BoundInstances(ctx, account)
	if err != nil {
		return "", fmt.Errorf("connector bindings of %s: %w", account, err)
	}
	if len(instances) != 1 {
		return "", fmt.Errorf("%s: %w (%d bound)", account, ErrNoConnector, len(instances))
	}
	return instances[0], nil
}
