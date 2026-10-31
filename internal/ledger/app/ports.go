package app

import (
	"context"
	"errors"

	"github.com/arkadiuss/openfief/internal/ledger"
	"github.com/arkadiuss/openfief/internal/portfolio"
)

// ErrVersionConflict reports an append of a version that its key already
// has: another import appended it first. Stores return it wrapped.
var ErrVersionConflict = errors.New("transaction version already exists")

// ErrNotFound reports a sleeve or account that is not stored. Stores return
// it wrapped.
var ErrNotFound = errors.New("not found")

// TransactionLog stores the versioned transaction log. It only ever appends:
// no stored version is changed or removed (D-18).
type TransactionLog interface {
	// Append stores v as version v.Version of its key, with its postings. It
	// fails with an error wrapping ErrVersionConflict when the key already
	// has that version, so two imports can never both append it.
	Append(ctx context.Context, v *ledger.VersionedTransaction) error
	// Versions returns every stored version of key, in version order, or none
	// for a key never imported.
	Versions(ctx context.Context, key ledger.TransactionKey) ([]ledger.VersionedTransaction, error)
	// Log returns every stored version of every key that has a version in
	// account, in key then version order: all of a key's versions, so the
	// ledger can tell which is current even if a correction moved it.
	Log(ctx context.Context, account ledger.AccountID) ([]ledger.VersionedTransaction, error)
}

// Registry stores sleeves and accounts.
type Registry interface {
	// SaveSleeve creates s or replaces the stored sleeve with its ID, with
	// its kind, role and offensive override. Its kind must be stored.
	SaveSleeve(ctx context.Context, s portfolio.Sleeve) error
	// Sleeves returns every sleeve, in ID order.
	Sleeves(ctx context.Context) ([]portfolio.Sleeve, error)
	// SaveAccount creates a or replaces the stored account with its ID.
	SaveAccount(ctx context.Context, a ledger.Account) error
	// Account returns the account with id, or an error wrapping ErrNotFound.
	Account(ctx context.Context, id ledger.AccountID) (ledger.Account, error)
	// Accounts returns every account, in ID order.
	Accounts(ctx context.Context) ([]ledger.Account, error)
}

// KindStore stores sleeve kinds. Sleeves live in the ledger's store
// because the portfolio context has no store of its own in M0, and so do
// their kinds.
type KindStore interface {
	// SaveKind creates k, or renames the stored kind with its key. It never
	// changes a stored kind's default role.
	SaveKind(ctx context.Context, k portfolio.Kind) error
	// Kinds returns every sleeve kind, in key order.
	Kinds(ctx context.Context) ([]portfolio.Kind, error)
}

// AssertionStore stores balance assertions with the result of their latest
// check.
type AssertionStore interface {
	// SaveAssertionCheck creates c or replaces the stored check of the same
	// assertion: the same account, date and ledger account.
	SaveAssertionCheck(ctx context.Context, c *AssertionCheck) error
	// AssertionChecks returns account's checks, in date then ledger account
	// order.
	AssertionChecks(ctx context.Context, account ledger.AccountID) ([]AssertionCheck, error)
}

// TransferLinkStore stores the owner's transfer links (D-105). It only ever
// appends: no stored version is changed or removed (D-18).
type TransferLinkStore interface {
	// AppendTransferLink stores l as version l.Version of its pair. It
	// fails with an error wrapping ErrVersionConflict when the pair already
	// has that version.
	AppendTransferLink(ctx context.Context, l *ledger.TransferLink) error
	// TransferLinks returns every stored version of every pair's link, in
	// pair then version order.
	TransferLinks(ctx context.Context) ([]ledger.TransferLink, error)
	// LockTransferLinks keeps every other LockTransferLinks waiting until
	// the database transaction of a Store that InTx passed in ends, so that
	// one decision at a time checks which legs are taken. On a Store that
	// InTx did not pass in it fails.
	LockTransferLinks(ctx context.Context) error
}

// Store is everything the ledger's use cases persist.
type Store interface {
	TransactionLog
	Registry
	KindStore
	AssertionStore
	TransferLinkStore
	// InTx runs fn with a Store whose every call belongs to one database
	// transaction, committed when fn returns nil and rolled back otherwise.
	// On a Store that InTx passed in, it runs fn in that same transaction.
	// It is the seam for work that must commit with an import, such as
	// enqueueing a revaluation (D-36).
	InTx(ctx context.Context, fn func(tx Store) error) error
	// LockAccount locks the account id until the database transaction of a
	// Store that InTx passed in ends, so that one transaction at a time
	// decides whether to seed it (D-23). It fails with an error wrapping
	// ErrNotFound for an account that is not stored, and on a Store that
	// InTx did not pass in.
	LockAccount(ctx context.Context, id ledger.AccountID) error
}
