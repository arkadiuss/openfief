package connectapi

import (
	"context"
	"fmt"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/arkadiuss/openfief/internal/identity"
	"github.com/arkadiuss/openfief/internal/identity/app"
	identityv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/identity/v1"
)

// CreateGrant lets a script token or a view plugin view an account, a sleeve
// or the household (D-209), for an owner session.
func (s *authService) CreateGrant(ctx context.Context, req *connect.Request[identityv1.CreateGrantRequest]) (*connect.Response[identityv1.CreateGrantResponse], error) {
	m := req.Msg
	holder, target, err := grantSides(m.GetHolderKind(), m.GetHolderId(), m.GetTargetKind(), m.GetTargetId())
	if err != nil {
		return nil, s.fail(ctx, err)
	}
	g, err := s.auth.CreateGrant(ctx, m.GetOrigin(), m.GetSessionToken(), holder, target)
	if err != nil {
		return nil, s.fail(ctx, err)
	}
	return connect.NewResponse(&identityv1.CreateGrantResponse{Grant: pbGrant(&g)}), nil
}

// RevokeGrant ends a grant for an owner session.
func (s *authService) RevokeGrant(ctx context.Context, req *connect.Request[identityv1.RevokeGrantRequest]) (*connect.Response[identityv1.RevokeGrantResponse], error) {
	m := req.Msg
	holder, target, err := grantSides(m.GetHolderKind(), m.GetHolderId(), m.GetTargetKind(), m.GetTargetId())
	if err != nil {
		return nil, s.fail(ctx, err)
	}
	if err := s.auth.RevokeGrant(ctx, m.GetOrigin(), m.GetSessionToken(), holder, target); err != nil {
		return nil, s.fail(ctx, err)
	}
	return connect.NewResponse(&identityv1.RevokeGrantResponse{}), nil
}

// ListGrants returns the grants for an owner session.
func (s *authService) ListGrants(ctx context.Context, req *connect.Request[identityv1.ListGrantsRequest]) (*connect.Response[identityv1.ListGrantsResponse], error) {
	m := req.Msg
	match, err := grantMatch(m.GetHolderKind(), m.GetHolderId())
	if err != nil {
		return nil, s.fail(ctx, err)
	}
	grants, err := s.auth.ListGrants(ctx, m.GetOrigin(), m.GetSessionToken(), match, m.GetInForceOnly())
	if err != nil {
		return nil, s.fail(ctx, err)
	}
	out := &identityv1.ListGrantsResponse{Grants: make([]*identityv1.Grant, 0, len(grants))}
	for i := range grants {
		out.Grants = append(out.Grants, pbGrant(&grants[i]))
	}
	return connect.NewResponse(out), nil
}

// The grant enums, from the domain to the contract, as functions: the package
// holds no global state. A value the domain has and a table lacks fails the
// contract test of this package.

// pbHolderKinds is the contract value of each domain value.
func pbHolderKinds() map[identity.GrantHolderKind]identityv1.GrantHolderKind {
	return map[identity.GrantHolderKind]identityv1.GrantHolderKind{
		identity.GrantHolderPerson:      identityv1.GrantHolderKind_GRANT_HOLDER_KIND_PERSON,
		identity.GrantHolderScriptToken: identityv1.GrantHolderKind_GRANT_HOLDER_KIND_SCRIPT_TOKEN,
		identity.GrantHolderViewPlugin:  identityv1.GrantHolderKind_GRANT_HOLDER_KIND_VIEW_PLUGIN,
	}
}

// pbTargetKinds is the contract value of each domain value.
func pbTargetKinds() map[identity.GrantTargetKind]identityv1.GrantTargetKind {
	return map[identity.GrantTargetKind]identityv1.GrantTargetKind{
		identity.GrantTargetAccount:   identityv1.GrantTargetKind_GRANT_TARGET_KIND_ACCOUNT,
		identity.GrantTargetSleeve:    identityv1.GrantTargetKind_GRANT_TARGET_KIND_SLEEVE,
		identity.GrantTargetHousehold: identityv1.GrantTargetKind_GRANT_TARGET_KIND_HOUSEHOLD,
	}
}

// grantSides converts the holder and the target of a request. An unspecified
// or unknown kind wraps app.ErrInvalidRequest; a household's empty ID is
// "home".
func grantSides(hk identityv1.GrantHolderKind, hid string, tk identityv1.GrantTargetKind, tid string) (identity.GrantHolder, identity.GrantTarget, error) {
	holder, ok := invert(pbHolderKinds())[hk]
	if !ok {
		return identity.GrantHolder{}, identity.GrantTarget{}, fmt.Errorf("%w: unknown holder kind", app.ErrInvalidRequest)
	}
	target, ok := invert(pbTargetKinds())[tk]
	if !ok {
		return identity.GrantHolder{}, identity.GrantTarget{}, fmt.Errorf("%w: unknown target kind", app.ErrInvalidRequest)
	}
	if target == identity.GrantTargetHousehold && tid == "" {
		tid = identity.HouseholdID
	}
	return identity.GrantHolder{Kind: holder, ID: hid}, identity.GrantTarget{Kind: target, ID: tid}, nil
}

// grantMatch converts the holder a list asks for: none selects every holder.
func grantMatch(hk identityv1.GrantHolderKind, hid string) (identity.GrantMatch, error) {
	if hk == identityv1.GrantHolderKind_GRANT_HOLDER_KIND_UNSPECIFIED {
		if hid != "" {
			return identity.GrantMatch{}, fmt.Errorf("%w: a holder ID needs its kind", app.ErrInvalidRequest)
		}
		return identity.GrantMatch{}, nil
	}
	holder, ok := invert(pbHolderKinds())[hk]
	if !ok {
		return identity.GrantMatch{}, fmt.Errorf("%w: unknown holder kind", app.ErrInvalidRequest)
	}
	return identity.GrantMatch{Holder: identity.GrantHolder{Kind: holder, ID: hid}}, nil
}

// pbGrant converts a grant: kinds, IDs and times, nothing else.
func pbGrant(g *identity.Grant) *identityv1.Grant {
	out := &identityv1.Grant{
		HolderKind: pbHolderKinds()[g.Holder.Kind], HolderId: g.Holder.ID,
		TargetKind: pbTargetKinds()[g.Target.Kind], TargetId: g.Target.ID,
		CreatedAt: timestamppb.New(g.CreatedAt),
	}
	if !g.RevokedAt.IsZero() {
		out.RevokedAt = timestamppb.New(g.RevokedAt)
	}
	return out
}
