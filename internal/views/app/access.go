package app

import (
	"context"
	"fmt"

	"github.com/arkadiuss/openfief/internal/identity"
	identityapp "github.com/arkadiuss/openfief/internal/identity/app"
	"github.com/arkadiuss/openfief/internal/ingest"
	"github.com/arkadiuss/openfief/internal/ledger"
)

// WithAuthorizer has the views ask a whether the caller may view each
// account (D-163); without it they ask identityapp.Policy, the only
// implementation until SpiceDB answers (D-164).
func WithAuthorizer(a identityapp.Authorizer) Option {
	return func(v *Views) { v.authz = a }
}

// mayView returns nil when the caller in ctx (identityapp.CallerFrom) may
// view account, and an error wrapping ErrUnknownAccount when it may not:
// a caller can't tell an account it may not view from one not stored. A
// caller-less ctx may view nothing.
func (v *Views) mayView(ctx context.Context, account ledger.AccountID) error {
	ok, err := v.authz.MayAct(ctx, identityapp.CallerFrom(ctx), identity.ActionView, string(account))
	if err != nil {
		return fmt.Errorf("authorizing a view of %s: %w", account, err)
	}
	if !ok {
		return fmt.Errorf("%w: %s", ErrUnknownAccount, account)
	}
	return nil
}

// viewable returns those of accounts the caller in ctx may view, in their
// order: what every list, total, allocation and performance is computed
// over (D-163).
func (v *Views) viewable(ctx context.Context, accounts []ledger.Account) ([]ledger.Account, error) {
	caller := identityapp.CallerFrom(ctx)
	out := make([]ledger.Account, 0, len(accounts))
	for i := range accounts {
		ok, err := v.authz.MayAct(ctx, caller, identity.ActionView, string(accounts[i].ID))
		if err != nil {
			return nil, fmt.Errorf("authorizing a view of %s: %w", accounts[i].ID, err)
		}
		if ok {
			out = append(out, accounts[i])
		}
	}
	return out, nil
}

// viewableBindings returns those of bindings whose account the caller in
// ctx may view, in their order.
func (v *Views) viewableBindings(ctx context.Context, bindings []ingest.Binding) ([]ingest.Binding, error) {
	caller := identityapp.CallerFrom(ctx)
	out := make([]ingest.Binding, 0, len(bindings))
	for i := range bindings {
		ok, err := v.authz.MayAct(ctx, caller, identity.ActionView, string(bindings[i].Account))
		if err != nil {
			return nil, fmt.Errorf("authorizing a view of %s: %w", bindings[i].Account, err)
		}
		if ok {
			out = append(out, bindings[i])
		}
	}
	return out, nil
}
