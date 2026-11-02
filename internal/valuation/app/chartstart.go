package app

import (
	"context"
	"errors"
	"fmt"
	"slices"

	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
)

// ErrUnknownAccount reports a setting of an account that is not stored.
var ErrUnknownAccount = errors.New("unknown account")

// AccountOwners checks that a session is the owner's and asks which
// accounts it may change (D-163); cmd/core answers it with the identity
// context's sessions and the one authorizer.
type AccountOwners interface {
	Owners
	// Changeable returns those of accounts the owner-scoped session token
	// valid on origin may change, in their order; it fails as CheckOwner
	// does.
	Changeable(ctx context.Context, origin, sessionToken string, accounts []ledger.AccountID) ([]ledger.AccountID, error)
}

// ChartStartStore keeps the charts' start the owner chose (D-138) as a
// setting (D-167), the portfolio's and each account's.
// *postgres.Store implements it.
type ChartStartStore interface {
	// ChartStart returns the start stored for account's charts, or for the
	// portfolio's when account is empty, and the zero date when none is.
	ChartStart(ctx context.Context, account ledger.AccountID) (kernel.Date, error)
	// SaveChartStart stores start for account's charts, or for the
	// portfolio's when account is empty; a zero start clears the setting.
	SaveChartStart(ctx context.Context, account ledger.AccountID, start kernel.Date) error
}

// ChartStarts is the owner's setting of where the time charts start
// (D-138, D-167): one for the portfolio's charts and one for each
// account's. No stored value depends on it, so it revalues nothing.
type ChartStarts struct {
	owners   AccountOwners
	store    ChartStartStore
	accounts AccountLister
}

// NewChartStarts returns the charts' start setting over store, with the
// owner's sessions and the accounts they may change checked by owners and
// the stored accounts listed by accounts.
func NewChartStarts(owners AccountOwners, store ChartStartStore, accounts AccountLister) *ChartStarts {
	return &ChartStarts{owners: owners, store: store, accounts: accounts}
}

// Set stores start as where account's charts start, or the portfolio's
// when account is empty, for the owner; a zero start goes back to the
// default (D-138) and clears the setting (D-167). The portfolio's needs
// the owner's session; an account's also needs the authorizer's leave to
// change it (D-163), asked before anything about it is read, so a refused
// account fails with ErrNotOwner whether it is stored or not. An account
// that is not stored is an error wrapping ErrUnknownAccount.
func (c *ChartStarts) Set(ctx context.Context, origin, sessionToken string, account ledger.AccountID, start kernel.Date) error {
	if err := c.mayChange(ctx, origin, sessionToken, account); err != nil {
		return fmt.Errorf("setting the charts' start: %w", err)
	}
	if account != "" {
		all, err := c.accounts.Accounts(ctx)
		if err != nil {
			return fmt.Errorf("setting the charts' start: %w", err)
		}
		if !slices.ContainsFunc(all, func(a ledger.Account) bool { return a.ID == account }) {
			return fmt.Errorf("setting the charts' start: %w: %s", ErrUnknownAccount, account)
		}
	}
	if err := c.store.SaveChartStart(ctx, account, start); err != nil {
		return fmt.Errorf("setting the charts' start: %w", err)
	}
	return nil
}

// mayChange returns nil when the session on origin is the owner's and,
// for an account, may change it (D-163).
func (c *ChartStarts) mayChange(ctx context.Context, origin, sessionToken string, account ledger.AccountID) error {
	if account == "" {
		if err := c.owners.CheckOwner(ctx, origin, sessionToken); err != nil {
			return fmt.Errorf("authorizing a change: %w", err)
		}
		return nil
	}
	ok, err := c.owners.Changeable(ctx, origin, sessionToken, []ledger.AccountID{account})
	if err != nil {
		return fmt.Errorf("authorizing a change: %w", err)
	}
	if !slices.Contains(ok, account) {
		return fmt.Errorf("%w: may not change account %s", ErrNotOwner, account)
	}
	return nil
}
