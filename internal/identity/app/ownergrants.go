package app

import (
	"context"
	"errors"
	"fmt"

	"github.com/arkadiuss/openfief/internal/identity"
)

// errNoGrants reports a call that needs the grant use cases on an Auth built
// without them.
var errNoGrants = errors.New("the grant use cases are not configured")

// CreateGrant lets holder view target for an owner session (D-209): see
// Grants.Create. A read-scoped session, a demo session, is refused with an
// error wrapping ErrForbidden.
func (a *Auth) CreateGrant(ctx context.Context, origin, token string, holder identity.GrantHolder, target identity.GrantTarget) (identity.Grant, error) {
	_, s, err := a.sessionOn(ctx, origin, token, identity.SessionScopeOwner)
	if err != nil {
		return identity.Grant{}, err
	}
	if a.grants == nil {
		return identity.Grant{}, errNoGrants
	}
	return a.grants.Create(ctx, ownerActor(&s), holder, target)
}

// RevokeGrant ends the grant of target to holder for an owner session: see
// Grants.Revoke.
func (a *Auth) RevokeGrant(ctx context.Context, origin, token string, holder identity.GrantHolder, target identity.GrantTarget) error {
	_, s, err := a.sessionOn(ctx, origin, token, identity.SessionScopeOwner)
	if err != nil {
		return err
	}
	if a.grants == nil {
		return errNoGrants
	}
	return a.grants.Revoke(ctx, ownerActor(&s), holder, target)
}

// ListGrants returns the grants m selects for an owner session, oldest first,
// those in force only or the history too. Reading them is no audit entry.
func (a *Auth) ListGrants(ctx context.Context, origin, token string, m identity.GrantMatch, inForceOnly bool) ([]identity.Grant, error) {
	if _, _, err := a.sessionOn(ctx, origin, token, identity.SessionScopeOwner); err != nil {
		return nil, err
	}
	if a.grants == nil {
		return nil, errNoGrants
	}
	grants, err := a.grants.List(ctx, m, inForceOnly)
	if err != nil {
		return nil, fmt.Errorf("listing grants: %w", err)
	}
	return grants, nil
}
