package connectapi_test

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/google/go-cmp/cmp"
	"google.golang.org/protobuf/testing/protocmp"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/arkadiuss/openfief/internal/identity"
	"github.com/arkadiuss/openfief/internal/identity/app"
	"github.com/arkadiuss/openfief/internal/identity/connectapi"
	"github.com/arkadiuss/openfief/internal/identity/webauthn"
	identityv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/identity/v1"
	"github.com/arkadiuss/openfief/sdk/go/gen/openfief/identity/v1/identityv1connect"
)

// inviteStore is a grantStore that also keeps enrollment codes that invite
// viewers, in memory.
type inviteStore struct {
	*grantStore

	codes []identity.EnrollmentCode
}

func (s *inviteStore) PendingInvites(_ context.Context, now time.Time) ([]identity.EnrollmentCode, error) {
	var out []identity.EnrollmentCode
	for i := range s.codes {
		if s.codes[i].Pending(now) {
			out = append(out, s.codes[i])
		}
	}
	return out, nil
}

func (s *inviteStore) CancelEnrollmentCode(_ context.Context, id identity.EnrollmentCodeID, at time.Time) error {
	for i := range s.codes {
		if s.codes[i].ID == id && s.codes[i].Pending(at) {
			s.codes[i].CancelledAt = at
			return nil
		}
	}
	return app.ErrUnknownEnrollmentCode
}

// invitesServed is the AuthService over an inviteStore with two pending
// invitations and one expired, an owner session whose passkey check is stale,
// and a client that presents the view plugin ui.
type invitesServed struct {
	client apiClient
	token  string
	store  *inviteStore
}

func newInvitesServed(t *testing.T) invitesServed {
	t.Helper()
	origins, err := identity.NewOrigins([]string{local})
	if err != nil {
		t.Fatalf("NewOrigins: %v", err)
	}
	now := func() time.Time { return time.Date(2026, 10, 28, 12, 0, 0, 0, time.UTC) }
	base := &sessionStore{sessions: map[identity.SessionID]identity.Session{}, hashes: map[identity.SessionID][]byte{}}
	store := &inviteStore{grantStore: &grantStore{sessionStore: base}}
	at := now().Add(-30 * time.Minute)
	for _, c := range []struct {
		id, name string
		ttl      time.Duration
		targets  []identity.GrantTarget
	}{
		{"00000000000000a1", "Alice", 24 * time.Hour, []identity.GrantTarget{{Kind: identity.GrantTargetSleeve, ID: "core"}}},
		{"00000000000000b2", "Bob", time.Hour, nil},
		{"00000000000000c3", "Carol", 15 * time.Minute, nil}, // expired by now
	} {
		code, codeErr := identity.NewEnrollmentCode(identity.EnrollmentCodeID(c.id), mustOrigin(t, local), at, c.ttl)
		if codeErr != nil {
			t.Fatalf("NewEnrollmentCode: %v", codeErr)
		}
		if codeErr = code.WithViewer(&identity.ViewerInvite{Name: c.name, Targets: c.targets}); codeErr != nil {
			t.Fatalf("WithViewer: %v", codeErr)
		}
		store.codes = append(store.codes, code)
	}
	grants := app.NewGrants(store, nil, directory{}, nil, now)
	auth, err := app.NewAuth(store, webauthn.New("OpenFief"), nil, marked{},
		app.AuthConfig{Origins: origins, Sessions: identity.DefaultSessionPolicy(), DemoSessions: true}, now, nil, app.WithGrants(grants))
	if err != nil {
		t.Fatalf("NewAuth: %v", err)
	}
	path, handler := connectapi.New(auth, callers{}, slog.New(slog.DiscardHandler))
	mux := http.NewServeMux()
	mux.Handle(path, handler)
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	client := apiClient{identityv1connect.NewAuthServiceClient(srv.Client(), srv.URL, connect.WithInterceptors(bearer(viewCaller)))}
	return invitesServed{client: client, token: staleOwnerSession(t, base), store: store}
}

// mustOrigin parses text as an origin.
func mustOrigin(t *testing.T, text string) identity.Origin {
	t.Helper()
	o, err := identity.ParseOrigin(text)
	if err != nil {
		t.Fatalf("ParseOrigin: %v", err)
	}
	return o
}

// TestInviteCallsAreOwnerOnly checks that a demo (read) session, no session, an
// unknown session and a caller that is no view plugin are refused ListInvites
// and CancelInvite, and that nothing is cancelled.
func TestInviteCallsAreOwnerOnly(t *testing.T) {
	t.Parallel()
	s := newInvitesServed(t)
	ctx := t.Context()
	demo := demoSession(t, s.client)
	for _, tc := range []struct {
		name  string
		token string
		want  connect.Code
	}{
		{"a demo (read) session", demo, connect.CodePermissionDenied},
		{"no session", "", connect.CodeUnauthenticated},
	} {
		_, err := s.client.ListInvites(ctx, connect.NewRequest(&identityv1.ListInvitesRequest{Origin: local, SessionToken: tc.token}))
		wantCode(t, "ListInvites with "+tc.name, err, tc.want)
		_, err = s.client.CancelInvite(ctx, connect.NewRequest(&identityv1.CancelInviteRequest{Origin: local, SessionToken: tc.token, InviteId: "00000000000000a1"}))
		wantCode(t, "CancelInvite with "+tc.name, err, tc.want)
	}
	for name, caller := range map[string]string{"no caller token": "", "a caller that is no view plugin": "connector-token"} {
		_, err := newClient(t, caller).CancelInvite(ctx, connect.NewRequest(&identityv1.CancelInviteRequest{Origin: local, SessionToken: s.token, InviteId: "00000000000000a1"}))
		wantCode(t, "CancelInvite from "+name, err, connect.CodePermissionDenied)
	}
	for i := range s.store.codes {
		if !s.store.codes[i].CancelledAt.IsZero() {
			t.Errorf("invitation %s was cancelled by a refused call", s.store.codes[i].ID)
		}
	}
}

// TestListAndCancelInvites checks ListInvites over Connect: the pending
// invitations, oldest first, each with its ID, name, grants and times and never
// the expired one; that CancelInvite, with the owner's session whose passkey
// check is stale (it needs none), removes one; and that cancelling what is gone,
// expired or unknown is NOT_FOUND.
func TestListAndCancelInvites(t *testing.T) {
	t.Parallel()
	s := newInvitesServed(t)
	ctx := t.Context()
	list := func() []*identityv1.Invite {
		resp, err := s.client.ListInvites(ctx, connect.NewRequest(&identityv1.ListInvitesRequest{Origin: local, SessionToken: s.token}))
		if err != nil {
			t.Fatalf("ListInvites: %v", err)
		}
		return resp.Msg.GetInvites()
	}
	created := func(c *identity.EnrollmentCode) *identityv1.Invite {
		return &identityv1.Invite{Id: string(c.ID), Name: c.Viewer.Name}
	}
	at := time.Date(2026, 10, 28, 11, 30, 0, 0, time.UTC)
	alice, bob := created(&s.store.codes[0]), created(&s.store.codes[1])
	alice.Grants = []*identityv1.GrantTarget{{Kind: identityv1.GrantTargetKind_GRANT_TARGET_KIND_SLEEVE, Id: "core"}}
	for _, p := range []struct {
		inv *identityv1.Invite
		ttl time.Duration
	}{{alice, 24 * time.Hour}, {bob, time.Hour}} {
		p.inv.CreatedAt, p.inv.ExpiresAt = timestamppb.New(at), timestamppb.New(at.Add(p.ttl))
	}
	if diff := cmp.Diff([]*identityv1.Invite{alice, bob}, list(), protocmp.Transform()); diff != "" {
		t.Errorf("ListInvites (-want +got):\n%s", diff)
	}

	_, err := s.client.CancelInvite(ctx, connect.NewRequest(&identityv1.CancelInviteRequest{Origin: local, SessionToken: s.token, InviteId: "00000000000000a1"}))
	if err != nil {
		t.Fatalf("CancelInvite: %v", err)
	}
	if diff := cmp.Diff([]*identityv1.Invite{bob}, list(), protocmp.Transform()); diff != "" {
		t.Errorf("ListInvites after cancelling Alice (-want +got):\n%s", diff)
	}
	for name, id := range map[string]string{"cancelled already": "00000000000000a1", "expired": "00000000000000c3", "unknown": "ffffffffffffffff"} {
		_, err = s.client.CancelInvite(ctx, connect.NewRequest(&identityv1.CancelInviteRequest{Origin: local, SessionToken: s.token, InviteId: id}))
		wantCode(t, "CancelInvite of one "+name, err, connect.CodeNotFound)
	}
	if got := s.store.audit[len(s.store.audit)-1]; got.Action != identity.ActionEnrollmentLinkCancelled || got.Outcome != identity.OutcomeOK {
		t.Errorf("last audit entry = %+v, want the cancellation", got)
	}
}
