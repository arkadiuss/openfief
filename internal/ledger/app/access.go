package app

import (
	"context"
	"fmt"

	"github.com/arkadiuss/openfief/internal/ledger"
)

// mayChange returns nil when the owner-scoped session token valid on origin
// may change every one of accounts (D-163); otherwise an error wrapping
// ErrNoSession or ErrNotOwner, or another error when owners can't tell.
func mayChange(ctx context.Context, owners Owners, origin, sessionToken string, accounts []ledger.AccountID) error {
	ok, err := owners.Changeable(ctx, origin, sessionToken, accounts)
	if err != nil {
		return fmt.Errorf("authorizing a change: %w", err)
	}
	allowed := make(map[ledger.AccountID]bool, len(ok))
	for _, a := range ok {
		allowed[a] = true
	}
	for _, a := range accounts {
		if !allowed[a] {
			return fmt.Errorf("%w: may not change account %s", ErrNotOwner, a)
		}
	}
	return nil
}
