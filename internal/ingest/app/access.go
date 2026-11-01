package app

import (
	"context"
	"fmt"

	"github.com/arkadiuss/openfief/internal/ledger"
)

// mayChange returns nil when the owner-scoped session token valid on origin
// may change every one of accounts (D-163), asked before anything about
// them is read; otherwise an error wrapping ErrNoSession or ErrNotOwner, so
// a refused account looks the same whether it is stored or not, or another
// error when owners can't tell.
func mayChange(ctx context.Context, owners Owners, origin, sessionToken string, accounts ...ledger.AccountID) error {
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

// changeable returns those of accounts the owner-scoped session token valid
// on origin may change (D-163), in their order: an account the owner may
// not change takes no file.
func (u *Uploads) changeable(ctx context.Context, origin, sessionToken string, accounts []FileAccount) ([]FileAccount, error) {
	ids := make([]ledger.AccountID, 0, len(accounts))
	for i := range accounts {
		ids = append(ids, accounts[i].Account)
	}
	ok, err := u.Owners.Changeable(ctx, origin, sessionToken, ids)
	if err != nil {
		return nil, fmt.Errorf("listing accounts that take files: %w", err)
	}
	allowed := make(map[ledger.AccountID]bool, len(ok))
	for _, a := range ok {
		allowed[a] = true
	}
	out := make([]FileAccount, 0, len(accounts))
	for i := range accounts {
		if allowed[accounts[i].Account] {
			out = append(out, accounts[i])
		}
	}
	return out, nil
}
