package app

import (
	"context"
	"fmt"
	"time"

	"github.com/arkadiuss/openfief/internal/ingest"
	"github.com/arkadiuss/openfief/internal/ledger"
)

// BindingStore keeps connector bindings: binding an instance to an account
// it is already bound to changes nothing.
type BindingStore interface {
	ConnectorBindings
	Bind(ctx context.Context, instance string, account ledger.AccountID, at time.Time) error
}

// AccountRegistry finds a stored account, or fails.
type AccountRegistry interface {
	Account(ctx context.Context, id ledger.AccountID) (ledger.Account, error)
}

// Bind gives the connector instance a connector binding to account, so its
// deliveries may write to it, after checking the instance's name and that
// the account is stored.
func Bind(ctx context.Context, accounts AccountRegistry, store BindingStore, instance string, account ledger.AccountID, at time.Time) error {
	if err := ingest.ValidateInstanceName(instance); err != nil {
		return fmt.Errorf("binding: %w", err)
	}
	if _, err := accounts.Account(ctx, account); err != nil {
		return fmt.Errorf("binding %s to %s: %w", instance, account, err)
	}
	if err := store.Bind(ctx, instance, account, at); err != nil {
		return fmt.Errorf("binding %s to %s: %w", instance, account, err)
	}
	return nil
}
