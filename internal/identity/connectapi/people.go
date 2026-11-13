package connectapi

import (
	"context"
	"log/slog"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/arkadiuss/openfief/internal/identity"
	"github.com/arkadiuss/openfief/internal/kernel"
	identityv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/identity/v1"
)

// InviteViewer issues the enrollment link that invites a viewer (D-65).
func (s *authService) InviteViewer(ctx context.Context, req *connect.Request[identityv1.InviteViewerRequest]) (*connect.Response[identityv1.InviteViewerResponse], error) {
	m := req.Msg
	targets, err := grantTargets(m.GetGrants())
	if err != nil {
		return nil, s.fail(ctx, err)
	}
	link, err := s.auth.InviteViewer(ctx, m.GetOrigin(), m.GetSessionToken(), m.GetName(), targets, m.GetTargetOrigin(), m.GetTtl().AsDuration())
	if err != nil {
		return nil, s.fail(ctx, err)
	}
	return connect.NewResponse(&identityv1.InviteViewerResponse{Url: link.URL, ExpiresAt: timestamppb.New(link.ExpiresAt)}), nil
}

// ListPeople returns the owner and every viewer, with their access.
func (s *authService) ListPeople(ctx context.Context, req *connect.Request[identityv1.ListPeopleRequest]) (*connect.Response[identityv1.ListPeopleResponse], error) {
	people, err := s.auth.People(ctx, req.Msg.GetOrigin(), req.Msg.GetSessionToken())
	if err != nil {
		return nil, s.fail(ctx, err)
	}
	out := &identityv1.ListPeopleResponse{People: make([]*identityv1.PersonAccess, 0, len(people)), GrantsEnforced: s.auth.GrantsEnforced()}
	for i := range people {
		pa := &identityv1.PersonAccess{
			Person: pbPerson(&people[i].Person), Passkeys: int32(people[i].Passkeys), ActiveSessions: int32(people[i].ActiveSessions), //nolint:gosec // a count of a person's passkeys or sessions
		}
		for j := range people[i].Grants {
			pa.Grants = append(pa.Grants, pbGrant(&people[i].Grants[j]))
		}
		out.People = append(out.People, pa)
	}
	return connect.NewResponse(out), nil
}

// RevokePerson revokes a viewer.
func (s *authService) RevokePerson(ctx context.Context, req *connect.Request[identityv1.RevokePersonRequest]) (*connect.Response[identityv1.RevokePersonResponse], error) {
	r, err := s.auth.RevokePerson(ctx, req.Msg.GetOrigin(), req.Msg.GetSessionToken(), req.Msg.GetPersonId())
	if err != nil {
		return nil, s.fail(ctx, err)
	}
	return connect.NewResponse(&identityv1.RevokePersonResponse{
		SessionsEnded: int32(r.Sessions), PasskeysDisabled: int32(r.Passkeys), GrantsEnded: int32(r.Grants), //nolint:gosec // counts of what one viewer held
	}), nil
}

// pbPerson converts a person; it never carries the WebAuthn user handle.
func pbPerson(p *identity.Person) *identityv1.Person {
	kind := identityv1.PersonKind_PERSON_KIND_VIEWER
	if p.Kind == identity.PersonOwner {
		kind = identityv1.PersonKind_PERSON_KIND_OWNER
	}
	out := &identityv1.Person{Id: p.ID, Kind: kind, Name: p.Name, CreatedAt: timestamppb.New(p.CreatedAt)}
	if !p.RevokedAt.IsZero() {
		out.RevokedAt = timestamppb.New(p.RevokedAt)
	}
	return out
}

// logNotProjected logs, by class only, that a viewer's grants are stored but
// could not be passed on to the authorization model.
func (s *authService) logNotProjected(ctx context.Context, err error) {
	if err == nil {
		return
	}
	s.log.LogAttrs(ctx, slog.LevelError, "a viewer's grants are stored but not passed on to the authorization model",
		slog.String("error", kernel.LogSafe(err, context.Canceled, context.DeadlineExceeded)))
}
