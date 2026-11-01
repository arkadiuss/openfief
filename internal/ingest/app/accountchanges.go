package app

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"time"

	"github.com/arkadiuss/openfief/internal/ingest"
	"github.com/arkadiuss/openfief/internal/ledger"
)

// ErrBrokerAccountHasTransactions reports a broker account whose account
// holds transactions from its connector instance, so it can't be remapped
// (D-120).
var ErrBrokerAccountHasTransactions = errors.New("the account holds transactions from the broker account")

// AccountDeletions deletes an account that holds no transactions and tells
// where an account's transactions come from (D-120); ledger/app.Ledger
// implements it.
type AccountDeletions interface {
	// DeleteAccount deletes account id unless it holds transactions, which
	// fails with an error wrapping ledger.ErrAccountHasTransactions.
	DeleteAccount(ctx context.Context, id ledger.AccountID) error
	// HasTransactionsFrom reports whether account holds a transaction whose
	// key source starts with prefix.
	HasTransactionsFrom(ctx context.Context, account ledger.AccountID, prefix string) (bool, error)
}

// AccountConnections is what the ingest schema holds of accounts' broker
// accounts and bindings beyond what BrokerAccountStore does;
// ingest/postgres.Store implements it.
type AccountConnections interface {
	// AllBrokerAccounts returns every instance's broker accounts, in
	// instance then ID order.
	AllBrokerAccounts(ctx context.Context) ([]ingest.BrokerAccount, error)
	// ForgetAccount removes a deleted account's bindings and leaves the
	// broker accounts mapped to it unmapped.
	ForgetAccount(ctx context.Context, account ledger.AccountID) error
	// RemapBrokerAccount maps instance's broker account id to account, or
	// leaves it unmapped when account is empty, whatever it was mapped to.
	RemapBrokerAccount(ctx context.Context, instance string, id ingest.BrokerAccountID, account ledger.AccountID, at time.Time) error
}

// TermsForgetter removes a deleted account's tax wrapper, fee rule and
// stored valuations; valuation/postgres.Store implements it.
type TermsForgetter interface {
	ForgetAccount(ctx context.Context, account ledger.AccountID) error
}

// BrokerAccounts returns the broker accounts connector instances reported
// (D-72), in instance then ID order, only the unmapped ones with unmapped,
// for the owner only: broker account numbers are listed nowhere else
// (D-121).
func (a *Accounts) BrokerAccounts(ctx context.Context, origin, sessionToken string, unmapped bool) ([]ingest.BrokerAccount, error) {
	if err := a.Owners.CheckOwner(ctx, origin, sessionToken); err != nil {
		return nil, fmt.Errorf("listing broker accounts: %w", err)
	}
	all, err := a.Connections.AllBrokerAccounts(ctx)
	if err != nil {
		return nil, fmt.Errorf("listing broker accounts: %w", err)
	}
	if unmapped {
		all = slices.DeleteFunc(all, func(b ingest.BrokerAccount) bool { return b.Mapped() })
	}
	return all, nil
}

// Delete deletes the account id when it holds no transactions (D-120): the
// ledger deletes it first, refusing one that holds any with an error
// wrapping ledger.ErrAccountHasTransactions, then its bindings and
// broker account mappings go, then its terms and stored valuations. Nothing
// the ledger holds is deleted (D-18). Deleting an account not stored, as a
// retry after a failure between the steps does, finishes the clean-up and
// succeeds.
func (a *Accounts) Delete(ctx context.Context, origin, sessionToken string, id ledger.AccountID) error {
	if err := a.Owners.CheckOwner(ctx, origin, sessionToken); err != nil {
		return fmt.Errorf("deleting an account: %w", err)
	}
	if id == "" {
		return fmt.Errorf("deleting an account: %w", ErrNoAccount)
	}
	if err := a.Deletions.DeleteAccount(ctx, id); err != nil {
		return fmt.Errorf("deleting account %s: %w", id, err)
	}
	if err := a.Connections.ForgetAccount(ctx, id); err != nil {
		return fmt.Errorf("deleting account %s: %w", id, err)
	}
	if err := a.Forget.ForgetAccount(ctx, id); err != nil {
		return fmt.Errorf("deleting account %s: %w", id, err)
	}
	return nil
}

// Remap maps instance's broker account broker to the stored account, or
// leaves it unmapped when account is empty, while the account it is mapped
// to holds no transactions from instance (D-120): their keys are scoped to
// it (ingest.ConnectorKeySource), so none starts with its prefix. One that
// does fails with ErrBrokerAccountHasTransactions and changes nothing. The
// instance is bound to the new account.
func (a *Accounts) Remap(ctx context.Context, origin, sessionToken, instance string, broker ingest.BrokerAccountID, account ledger.AccountID) error {
	if err := a.Owners.CheckOwner(ctx, origin, sessionToken); err != nil {
		return fmt.Errorf("remapping a broker account: %w", err)
	}
	if err := ingest.ValidateInstanceName(instance); err != nil {
		return fmt.Errorf("remapping a broker account: connector instance: %w", err)
	}
	if err := a.checkRemappable(ctx, instance, broker, account); err != nil {
		return err
	}
	if account != "" {
		if _, err := a.stored(ctx, account); err != nil {
			return err
		}
	}
	if err := a.Connections.RemapBrokerAccount(ctx, instance, broker, account, a.Now()); err != nil {
		return fmt.Errorf("remapping a broker account of %s: %w", instance, err)
	}
	if account == "" {
		return nil
	}
	if err := a.Bindings.Bind(ctx, instance, account, a.Now()); err != nil {
		return fmt.Errorf("binding %s: %w", instance, err)
	}
	return nil
}

// checkRemappable checks that instance reported broker and that the
// account it is mapped to, unless that is account, holds no transaction
// from instance.
func (a *Accounts) checkRemappable(ctx context.Context, instance string, broker ingest.BrokerAccountID, account ledger.AccountID) error {
	known, err := a.Brokers.BrokerAccounts(ctx, instance)
	if err != nil {
		return fmt.Errorf("loading the broker accounts of %s: %w", instance, err)
	}
	i := slices.IndexFunc(known, func(b ingest.BrokerAccount) bool { return b.ID == broker })
	if i < 0 {
		return fmt.Errorf("connector instance %s: %w", instance, ErrUnknownBrokerAccount)
	}
	old := known[i].Account
	if old == "" || old == account {
		return nil
	}
	has, err := a.Deletions.HasTransactionsFrom(ctx, old, ingest.ConnectorKeySource(instance, ""))
	if err != nil {
		return fmt.Errorf("remapping a broker account of %s: %w", instance, err)
	}
	if has {
		return fmt.Errorf("account %s: %w", old, ErrBrokerAccountHasTransactions)
	}
	return nil
}
