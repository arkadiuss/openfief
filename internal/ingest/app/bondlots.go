package app

import (
	"context"
	"errors"
	"fmt"
	"time"

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

// ErrLotSyncing reports a bond lot deletion refused because the account's
// connector instance is syncing: the sync may already have read the lot,
// and would book its coupons after the deletion (D-154).
var ErrLotSyncing = errors.New("the account's connector is syncing it now; delete the lot again when the sync ends")

// deleteTimeout bounds a bond lot deletion's call and import, which hold
// the instance's claim; the claim lasts leaseMargin longer.
const deleteTimeout = 5 * time.Minute

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
	// Claims claims the account's instance against its syncs while Delete
	// runs, so no sync that read the lot books its coupons after the
	// deletion (D-154); nil claims nothing.
	Claims SyncClaims
	// Now returns the time a claim starts from; Claims needs it.
	Now func() time.Time
	// Today returns the date a lot may be bought on at the latest.
	Today func() kernel.Date
	// NewBatchID returns a fresh import batch ID.
	NewBatchID func() ingest.BatchID
	// Audit keeps each entry and deletion of a lot with its audit entry;
	// nil enters none.
	Audit ChangeAuditor
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
	// The connector is asked before the change's transaction opens: no network
	// call is made inside it (D-211). Only its writes are the change.
	prepared, err := b.Connectors.PrepareBondLots(ctx, instance, b.NewBatchID(), account, &ingest.BondLotEntries{Lots: []ingest.BondLot{*lot}})
	if err != nil {
		return ingest.Batch{}, fmt.Errorf("adding a bond lot: %w", err)
	}
	var batch ingest.Batch
	err = audited(ctx, b.Audit, origin, sessionToken, auditBondLotAdded, []string{accountRef(account), instanceRef(instance)}, func(ctx context.Context) error {
		var importErr error
		if batch, importErr = b.Connectors.Commit(ctx, prepared); importErr != nil {
			return fmt.Errorf("import bond lots into %s: %w", instance, importErr)
		}
		return b.Batches.RecordBatch(ctx, instance, &batch)
	})
	if err != nil {
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
// (D-18); deleting it again changes nothing. Delete holds the instance's
// sync claim from reading what was booked to importing the deletions, so
// a sync either ran before it, and what it booked for the lot is deleted
// too, or runs after it, when the lot is no longer open; while a sync
// holds the claim Delete wraps ErrLotSyncing. A key without
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
	batch, err := b.deleteClaimed(ctx, &deletion{
		origin: origin, sessionToken: sessionToken, instance: instance, account: account, key: key,
	})
	if err != nil {
		return ingest.Batch{}, fmt.Errorf("deleting a bond lot: %w", err)
	}
	return batch, nil
}

// deletion is a bond lot deletion the owner asked for: the lot key of account,
// whose connector instance is instance.
type deletion struct {
	origin, sessionToken string
	instance             string
	account              ledger.AccountID
	key                  ingest.BondLotKey
}

// deleteClaimed claims the instance against its syncs, deletes the lot
// (deleteLot), and releases the claim; a failed release fails the deletion
// too. The claim is not part of the change: it is held from before the
// connector is asked to after the deletion's transaction commits.
func (b *BondLots) deleteClaimed(ctx context.Context, d *deletion) (ingest.Batch, error) {
	instance := d.instance
	release, err := b.claim(ctx, instance)
	if err != nil {
		return ingest.Batch{}, err
	}
	batch, err := b.deleteLot(ctx, d)
	if rerr := release(); rerr != nil {
		return ingest.Batch{}, errors.Join(err, rerr)
	}
	if err != nil {
		return ingest.Batch{}, err
	}
	return batch, nil
}

// claim claims instance against its syncs for deleteTimeout and leaseMargin
// from now, and returns the function that releases that claim, and no
// other; while a sync holds it, it fails wrapping ErrLotSyncing. Without
// Claims it claims nothing.
func (b *BondLots) claim(ctx context.Context, instance string) (func() error, error) {
	if b.Claims == nil {
		return func() error { return nil }, nil
	}
	now := b.Now()
	until := now.Add(deleteTimeout + leaseMargin)
	if _, err := b.Claims.ClaimSync(ctx, instance, now, until); err != nil {
		if errors.Is(err, ErrSyncRunning) {
			return nil, fmt.Errorf("%w: %w", ErrLotSyncing, err)
		}
		return nil, fmt.Errorf("claiming %s: %w", instance, err)
	}
	return func() error {
		if err := b.Claims.ReleaseSync(context.WithoutCancel(ctx), instance, until); err != nil {
			return fmt.Errorf("releasing %s: %w", instance, err)
		}
		return nil
	}, nil
}

// deleteLot tells the instance what it booked in the account with the key as a
// deleted lot, and imports and records its delivery, within deleteTimeout.
// The connector is asked first, outside any transaction, and only the import
// and the record of the batch are the change, audited (D-211).
func (b *BondLots) deleteLot(ctx context.Context, d *deletion) (ingest.Batch, error) {
	ctx, cancel := context.WithTimeout(ctx, deleteTimeout)
	defer cancel()
	instance := d.instance
	booked, err := b.booked(ctx, instance, d.account)
	if err != nil {
		return ingest.Batch{}, err
	}
	entries := &ingest.BondLotEntries{Deleted: []ingest.BondLotKey{d.key}, Booked: booked}
	prepared, err := b.Connectors.PrepareBondLots(ctx, instance, b.NewBatchID(), d.account, entries)
	if err != nil {
		return ingest.Batch{}, err
	}
	var batch ingest.Batch
	err = audited(ctx, b.Audit, d.origin, d.sessionToken, auditBondLotDeleted, []string{accountRef(d.account), instanceRef(instance)}, func(ctx context.Context) error {
		var importErr error
		if batch, importErr = b.Connectors.Commit(ctx, prepared); importErr != nil {
			return fmt.Errorf("import bond lots into %s: %w", instance, importErr)
		}
		if recordErr := b.Batches.RecordBatch(ctx, instance, &batch); recordErr != nil {
			return fmt.Errorf("recording the batch: %w", recordErr)
		}
		return nil
	})
	if err != nil {
		return ingest.Batch{}, err
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
