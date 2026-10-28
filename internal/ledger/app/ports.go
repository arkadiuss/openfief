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
	// SaveSleeve creates s or replaces the stored sleeve with its ID.
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

// Store is everything the ledger's use cases persist.
type Store interface {
	TransactionLog
	Registry
	AssertionStore
	// InTx runs fn with a Store whose every call belongs to one database
	// transaction, committed when fn returns nil and rolled back otherwise.
	// On a Store that InTx passed in, it runs fn in that same transaction.
	// It is the seam for work that must commit with an import, such as
	// enqueueing a revaluation (D-36).
	InTx(ctx context.Context, fn func(tx Store) error) error
}
