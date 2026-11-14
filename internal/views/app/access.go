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
	if lister, ok := v.authz.(identityapp.AccountLister); ok {
		// One call for every account, as SpiceDB answers a list (D-164).
		set, err := lister.ActableAccounts(ctx, caller, identity.ActionView)
		if err != nil {
			return nil, fmt.Errorf("authorizing the views of the accounts: %w", err)
		}
		out := make([]ledger.Account, 0, len(accounts))
		for i := range accounts {
			if _, ok := set[string(accounts[i].ID)]; ok {
				out = append(out, accounts[i])
			}
		}
		return out, nil
	}
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

// unboundVisible reports whether the caller in ctx may see what is tied to
// a connector instance and to no account (D-165, D-199): the owner and the
// other holders of read access (D-137), script tokens, view plugins and
// demo sessions, which view every account until SpiceDB answers (D-163),
// and nobody else, so a viewer (D-65) never does.
func unboundVisible(ctx context.Context) bool {
	switch identityapp.CallerFrom(ctx).Holder {
	case identity.HolderOwner, identity.HolderScriptToken, identity.HolderViewPlugin, identity.HolderDemoSession:
		return true
	case identity.HolderViewer:
		return false
	}
	return false
}

// instanceAccounts returns the accounts each connector instance is bound to
// (the connector bindings, D-72), by instance; none without Connections.
func (v *Views) instanceAccounts(ctx context.Context) (map[string][]ledger.AccountID, error) {
	out := map[string][]ledger.AccountID{}
	if v.connections == nil {
		return out, nil
	}
	bindings, err := v.connections.Bindings(ctx)
	if err != nil {
		return nil, fmt.Errorf("loading connector bindings: %w", err)
	}
	for _, b := range bindings {
		out[b.Instance] = append(out[b.Instance], b.Account)
	}
	return out, nil
}

// mayViewAny reports whether the caller in ctx may see a record tied to
// accounts, such as a connector instance's syncs tied to the accounts it is
// bound to (D-199): when it may view any one of them; a record tied to none
// as unboundVisible says.
func (v *Views) mayViewAny(ctx context.Context, accounts []ledger.AccountID) (bool, error) {
	if len(accounts) == 0 {
		return unboundVisible(ctx), nil
	}
	caller := identityapp.CallerFrom(ctx)
	for _, a := range accounts {
		ok, err := v.authz.MayAct(ctx, caller, identity.ActionView, string(a))
		if err != nil {
			return false, fmt.Errorf("authorizing a view of %s: %w", a, err)
		}
		if ok {
			return true, nil
		}
	}
	return false, nil
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
