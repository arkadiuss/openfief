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

// GrantPolicy is the Authorizer under --authz=policy once viewers exist
// (D-65): everyone is answered as identity.MayAct answers, except a viewer,
// who may view exactly the accounts their grants cover (D-209) and change
// none. It fails closed: an error reading the grants or the accounts is a
// refusal. Script tokens and view plugins keep viewing every account under
// the policy, as before; their grants take effect under SpiceDB.
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

// MayAct implements Authorizer.
func (p *GrantPolicy) MayAct(ctx context.Context, c identity.Caller, a identity.Action, account string) (bool, error) {
	if c.Holder != identity.HolderViewer {
		return identity.MayAct(c, a), nil
	}
	if a != identity.ActionView {
		return false, nil
	}
	set, err := p.viewerAccounts(ctx, c.ID)
	if err != nil {
		return false, err
	}
	_, ok := set[account]
	return ok, nil
}

// ActableAccounts implements AccountLister: for a viewer the accounts their
// grants cover, for any other caller every account or none, as MayAct says.
func (p *GrantPolicy) ActableAccounts(ctx context.Context, c identity.Caller, a identity.Action) (map[string]struct{}, error) {
	if c.Holder == identity.HolderViewer {
		if a != identity.ActionView {
			return map[string]struct{}{}, nil
		}
		return p.viewerAccounts(ctx, c.ID)
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

// viewerAccounts returns the accounts the grants in force of the viewer
// person cover.
func (p *GrantPolicy) viewerAccounts(ctx context.Context, person string) (map[string]struct{}, error) {
	holder := identity.GrantHolder{Kind: identity.GrantHolderPerson, ID: person}
	if err := holder.Validate(); err != nil {
		return map[string]struct{}{}, nil //nolint:nilerr // a caller whose ID is no person's holds nothing
	}
	grants, err := p.grants.Grants(ctx, identity.GrantMatch{Holder: holder}, true)
	if err != nil {
		return nil, fmt.Errorf("reading a viewer's grants: %w", err)
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
