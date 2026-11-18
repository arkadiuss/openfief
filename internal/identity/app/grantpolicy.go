package app

import (
	"context"
	"fmt"

	"github.com/arkadiuss/openfief/internal/identity"
)

// GrantReader reads the grants in force that say what a holder may view.
// *identity/postgres.Store implements it (GrantStore's Grants).
type GrantReader interface {
	Grants(ctx context.Context, m identity.GrantMatch, inForceOnly bool) ([]identity.Grant, error)
}

// TargetCover says which accounts a grant target covers now, as the
// household's accounts and sleeves hold them: the account itself, every
// account in a sleeve, every account of the household. cmd/core implements it
// over the ledger and the portfolio.
type TargetCover interface {
	// TargetAccounts returns the IDs of the accounts t covers.
	TargetAccounts(ctx context.Context, t identity.GrantTarget) ([]string, error)
}

// GrantPolicy is the Authorizer under --authz=policy (D-65, D-209): the owner
// and a demo session are answered as identity.MayAct answers, and a viewer, a
// script token and a view plugin may view exactly the accounts their grants
// cover and change none. D-163's interim rule that every other holder of read
// access views every account ended when the grants became the truth: existing
// tokens and plugins were migrated to explicit household grants. It fails
// closed: an error reading the grants or the accounts is a refusal.
type GrantPolicy struct {
	grants GrantReader
	cover  TargetCover
}

var (
	_ Authorizer    = (*GrantPolicy)(nil)
	_ AccountLister = (*GrantPolicy)(nil)
)

// NewGrantPolicy returns the policy that limits viewers to their grants,
// read from grants, over the accounts cover says each grant covers.
func NewGrantPolicy(grants GrantReader, cover TargetCover) *GrantPolicy {
	return &GrantPolicy{grants: grants, cover: cover}
}

// EnforcesTokenGrants reports that the policy limits script tokens and view
// plugins to their grants.
func (*GrantPolicy) EnforcesTokenGrants() bool { return true }

// MayAct implements Authorizer.
func (p *GrantPolicy) MayAct(ctx context.Context, c identity.Caller, a identity.Action, account string) (bool, error) {
	holder, granted := grantHolderOf(c)
	if !granted {
		return identity.MayAct(c, a), nil
	}
	if a != identity.ActionView {
		return false, nil
	}
	set, err := p.holderAccounts(ctx, holder)
	if err != nil {
		return false, err
	}
	_, ok := set[account]
	return ok, nil
}

// ActableAccounts implements AccountLister: for a holder of grants the accounts
// they cover, for any other caller every account or none, as MayAct says.
func (p *GrantPolicy) ActableAccounts(ctx context.Context, c identity.Caller, a identity.Action) (map[string]struct{}, error) {
	if holder, granted := grantHolderOf(c); granted {
		if a != identity.ActionView {
			return map[string]struct{}{}, nil
		}
		return p.holderAccounts(ctx, holder)
	}
	if !identity.MayAct(c, a) {
		return map[string]struct{}{}, nil
	}
	all, err := p.cover.TargetAccounts(ctx, identity.Household())
	if err != nil {
		return nil, fmt.Errorf("listing the accounts: %w", err)
	}
	return toSet(all), nil
}

// grantHolderOf returns the holder of the grants that limit c, and whether c
// is limited by grants: a viewer, a script token or a view plugin is.
func grantHolderOf(c identity.Caller) (identity.GrantHolder, bool) {
	kinds := map[identity.Holder]identity.GrantHolderKind{
		identity.HolderViewer:      identity.GrantHolderPerson,
		identity.HolderScriptToken: identity.GrantHolderScriptToken,
		identity.HolderViewPlugin:  identity.GrantHolderViewPlugin,
	}
	kind, ok := kinds[c.Holder]
	return identity.GrantHolder{Kind: kind, ID: c.ID}, ok
}

// holderAccounts returns the accounts the grants in force of holder cover.
func (p *GrantPolicy) holderAccounts(ctx context.Context, holder identity.GrantHolder) (map[string]struct{}, error) {
	if err := holder.Validate(); err != nil {
		return map[string]struct{}{}, nil //nolint:nilerr // a caller whose ID is no holder's holds nothing
	}
	grants, err := p.grants.Grants(ctx, identity.GrantMatch{Holder: holder}, true)
	if err != nil {
		return nil, fmt.Errorf("reading a holder's grants: %w", err)
	}
	out := map[string]struct{}{}
	for i := range grants {
		accounts, err := p.cover.TargetAccounts(ctx, grants[i].Target)
		if err != nil {
			return nil, fmt.Errorf("listing the accounts a grant covers: %w", err)
		}
		for _, a := range accounts {
			out[a] = struct{}{}
		}
	}
	return out, nil
}

// toSet returns ids as a set.
func toSet(ids []string) map[string]struct{} {
	out := make(map[string]struct{}, len(ids))
	for _, id := range ids {
		out[id] = struct{}{}
	}
	return out
}
