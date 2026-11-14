package connectapi_test

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/google/go-cmp/cmp"

	"github.com/arkadiuss/openfief/internal/identity"
	"github.com/arkadiuss/openfief/internal/identity/app"
	"github.com/arkadiuss/openfief/internal/identity/connectapi"
	"github.com/arkadiuss/openfief/internal/identity/webauthn"
	identityv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/identity/v1"
	"github.com/arkadiuss/openfief/sdk/go/gen/openfief/identity/v1/identityv1connect"
)

// grantStore is a sessionStore that also keeps grants, in memory.
type grantStore struct {
	*sessionStore

	grants []identity.Grant
}

func (g *grantStore) SaveGrant(_ context.Context, gr *identity.Grant) (bool, error) {
	for i := range g.grants {
		if g.grants[i].Active() && g.grants[i].Holder == gr.Holder && g.grants[i].Target == gr.Target {
			return false, nil
		}
	}
	g.grants = append(g.grants, *gr)
	return true, nil
}

func (g *grantStore) RevokeGrants(_ context.Context, m identity.GrantMatch, at time.Time) ([]identity.Grant, error) {
	var ended []identity.Grant
	for i := range g.grants {
		if g.grants[i].Active() && m.Matches(&g.grants[i]) {
			g.grants[i] = g.grants[i].Revoked(at)
			ended = append(ended, g.grants[i])
		}
	}
	return ended, nil
}

func (g *grantStore) Grants(_ context.Context, m identity.GrantMatch, inForceOnly bool) ([]identity.Grant, error) {
	var out []identity.Grant
	for i := range g.grants {
		if m.Matches(&g.grants[i]) && (!inForceOnly || g.grants[i].Active()) {
			out = append(out, g.grants[i])
		}
	}
	return out, nil
}

func (*grantStore) GrantCutoverDone(context.Context) (bool, error)            { return false, nil }
func (*grantStore) MarkGrantCutover(context.Context, time.Time) (bool, error) { return true, nil }

// directory knows the account main, the sleeve core and the plugin ui.
type directory struct{}

func (directory) TargetExists(_ context.Context, t identity.GrantTarget) (bool, error) {
	return t.Kind == identity.GrantTargetHousehold || t.ID == "main" || t.ID == "core", nil
}

func (directory) ViewPluginInstalled(_ context.Context, name string) (bool, error) {
	return name == "ui", nil
}

func (directory) ViewPlugins(context.Context) ([]string, error) { return []string{"ui"}, nil }

func (directory) AccessViewPlugins(context.Context) ([]string, error) { return []string{"ui"}, nil }

// grantsServed is the AuthService over a store with grants, an owner session
// planted in it and a client that presents the view plugin ui.
type grantsServed struct {
	client apiClient
	token  string
}

func newGrantsServed(t *testing.T) grantsServed {
	t.Helper()
	origins, err := identity.NewOrigins([]string{local})
	if err != nil {
		t.Fatalf("NewOrigins: %v", err)
	}
	now := func() time.Time { return time.Date(2026, 10, 28, 12, 0, 0, 0, time.UTC) }
	base := &sessionStore{sessions: map[identity.SessionID]identity.Session{}, hashes: map[identity.SessionID][]byte{}}
	store := &grantStore{sessionStore: base}
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
	return grantsServed{client: client, token: staleOwnerSession(t, base)}
}

// plugin is the grant holder ui.
func (g grantsServed) create(t *testing.T, target identityv1.GrantTargetKind, id string) (*identityv1.CreateGrantResponse, error) {
	t.Helper()
	resp, err := g.client.CreateGrant(t.Context(), connect.NewRequest(&identityv1.CreateGrantRequest{
		Origin: local, SessionToken: g.token, HolderKind: identityv1.GrantHolderKind_GRANT_HOLDER_KIND_VIEW_PLUGIN, HolderId: "ui",
		TargetKind: target, TargetId: id,
	}))
	if err != nil {
		return nil, err //nolint:wrapcheck // the test asserts on the Connect error itself
	}
	return resp.Msg, nil
}

// TestGrantCallsAreOwnerOnly checks that a demo session, no session, an
// unknown session and a caller that is no view plugin are refused every
// grant call, and that nothing is granted (D-209, D-137).
func TestGrantCallsAreOwnerOnly(t *testing.T) {
	t.Parallel()
	g := newGrantsServed(t)
	ctx := t.Context()
	demo := demoSession(t, g.client)
	unknown := identity.FormatSessionToken("0123456789abcdef", make([]byte, identity.KeyBytes))
	for _, tc := range []struct {
		name  string
		token string
		want  connect.Code
	}{
		{"a demo (read) session", demo, connect.CodePermissionDenied},
		{"no session", "", connect.CodeUnauthenticated},
		{"an unknown session", unknown, connect.CodeUnauthenticated},
	} {
		_, err := g.client.CreateGrant(ctx, connect.NewRequest(&identityv1.CreateGrantRequest{
			Origin: local, SessionToken: tc.token, HolderKind: identityv1.GrantHolderKind_GRANT_HOLDER_KIND_VIEW_PLUGIN, HolderId: "ui",
			TargetKind: identityv1.GrantTargetKind_GRANT_TARGET_KIND_HOUSEHOLD,
		}))
		wantCode(t, "CreateGrant with "+tc.name, err, tc.want)
		_, err = g.client.RevokeGrant(ctx, connect.NewRequest(&identityv1.RevokeGrantRequest{
			Origin: local, SessionToken: tc.token, HolderKind: identityv1.GrantHolderKind_GRANT_HOLDER_KIND_VIEW_PLUGIN, HolderId: "ui",
			TargetKind: identityv1.GrantTargetKind_GRANT_TARGET_KIND_HOUSEHOLD,
		}))
		wantCode(t, "RevokeGrant with "+tc.name, err, tc.want)
		_, err = g.client.ListGrants(ctx, connect.NewRequest(&identityv1.ListGrantsRequest{Origin: local, SessionToken: tc.token}))
		wantCode(t, "ListGrants with "+tc.name, err, tc.want)
	}
	for name, caller := range map[string]string{"no caller token": "", "a caller that is no view plugin": "connector-token"} {
		_, err := newClient(t, caller).CreateGrant(ctx, connect.NewRequest(&identityv1.CreateGrantRequest{Origin: local, SessionToken: g.token}))
		wantCode(t, "CreateGrant from "+name, err, connect.CodePermissionDenied)
		_, err = newClient(t, caller).ListGrants(ctx, connect.NewRequest(&identityv1.ListGrantsRequest{Origin: local, SessionToken: g.token}))
		wantCode(t, "ListGrants from "+name, err, connect.CodePermissionDenied)
	}
	got, err := g.client.ListGrants(ctx, connect.NewRequest(&identityv1.ListGrantsRequest{Origin: local, SessionToken: g.token}))
	if err != nil || len(got.Msg.GetGrants()) != 0 {
		t.Errorf("ListGrants after the refusals = %v, %v, want none", got, err)
	}
}

// revokeVia ends the grant of the plugin ui to a target as the owner.
func (g grantsServed) revokeVia(kind identityv1.GrantTargetKind, id string) error {
	_, err := g.client.RevokeGrant(context.Background(), connect.NewRequest(&identityv1.RevokeGrantRequest{
		Origin: local, SessionToken: g.token, HolderKind: identityv1.GrantHolderKind_GRANT_HOLDER_KIND_VIEW_PLUGIN, HolderId: "ui",
		TargetKind: kind, TargetId: id,
	}))
	return err //nolint:wrapcheck // the test asserts on the Connect error itself
}

// lines lists the grants as the owner, one line each: the kind and ID of the
// target and whether the grant is in force or ended.
func (g grantsServed) lines(t *testing.T, inForceOnly bool, kind identityv1.GrantHolderKind, id string) []string {
	t.Helper()
	resp, err := g.client.ListGrants(t.Context(), connect.NewRequest(&identityv1.ListGrantsRequest{
		Origin: local, SessionToken: g.token, InForceOnly: inForceOnly, HolderKind: kind, HolderId: id,
	}))
	if err != nil {
		t.Fatalf("ListGrants: %v", err)
	}
	out := make([]string, 0, len(resp.Msg.GetGrants()))
	for _, gr := range resp.Msg.GetGrants() {
		state := "in force"
		if gr.GetRevokedAt() != nil {
			state = "ended"
		}
		out = append(out, gr.GetTargetKind().String()+" "+gr.GetTargetId()+" "+state)
	}
	return out
}

// TestCreateGrantCall checks CreateGrant over Connect: each kind of target,
// the grant in force when it is made again, and the codes of what can't be
// granted.
func TestCreateGrantCall(t *testing.T) {
	t.Parallel()
	g := newGrantsServed(t)
	for _, tc := range []struct {
		kind identityv1.GrantTargetKind
		id   string
	}{
		{identityv1.GrantTargetKind_GRANT_TARGET_KIND_ACCOUNT, "main"},
		{identityv1.GrantTargetKind_GRANT_TARGET_KIND_SLEEVE, "core"},
		{identityv1.GrantTargetKind_GRANT_TARGET_KIND_HOUSEHOLD, ""}, // "home" by default
	} {
		resp, err := g.create(t, tc.kind, tc.id)
		if err != nil {
			t.Fatalf("CreateGrant(%v): %v", tc.kind, err)
		}
		gr := resp.GetGrant()
		if gr.GetHolderKind() != identityv1.GrantHolderKind_GRANT_HOLDER_KIND_VIEW_PLUGIN || gr.GetHolderId() != "ui" ||
			gr.GetTargetKind() != tc.kind || gr.GetCreatedAt() == nil || gr.GetRevokedAt() != nil {
			t.Errorf("CreateGrant(%v) = %v", tc.kind, gr)
		}
	}
	if again, err := g.create(t, identityv1.GrantTargetKind_GRANT_TARGET_KIND_HOUSEHOLD, "home"); err != nil || again.GetGrant().GetTargetId() != "home" {
		t.Errorf("CreateGrant of the household again = %v, %v, want the grant in force, on home", again, err)
	}
	g.checkTargetRefusals(t)
	g.checkHolderRefusals(t)
}

// checkTargetRefusals checks the codes CreateGrant answers a target that
// can't be granted with.
func (g grantsServed) checkTargetRefusals(t *testing.T) {
	t.Helper()
	for name, tc := range map[string]struct {
		kind identityv1.GrantTargetKind
		id   string
		want connect.Code
	}{
		"an unknown account":        {identityv1.GrantTargetKind_GRANT_TARGET_KIND_ACCOUNT, "nowhere", connect.CodeNotFound},
		"an unknown sleeve":         {identityv1.GrantTargetKind_GRANT_TARGET_KIND_SLEEVE, "nowhere", connect.CodeNotFound},
		"an account named a secret": {identityv1.GrantTargetKind_GRANT_TARGET_KIND_ACCOUNT, "ofs_x", connect.CodeInvalidArgument},
		"no target kind":            {identityv1.GrantTargetKind_GRANT_TARGET_KIND_UNSPECIFIED, "main", connect.CodeInvalidArgument},
		"an unknown target kind":    {99, "main", connect.CodeInvalidArgument},
		"a household that is away":  {identityv1.GrantTargetKind_GRANT_TARGET_KIND_HOUSEHOLD, "away", connect.CodeInvalidArgument},
	} {
		_, err := g.create(t, tc.kind, tc.id)
		wantCode(t, "CreateGrant of "+name, err, tc.want)
	}
}

// checkHolderRefusals checks the codes CreateGrant answers a holder that
// can't be granted anything with.
func (g grantsServed) checkHolderRefusals(t *testing.T) {
	t.Helper()
	for name, tc := range map[string]struct {
		kind identityv1.GrantHolderKind
		id   string
		want connect.Code
	}{
		"a plugin not installed":       {identityv1.GrantHolderKind_GRANT_HOLDER_KIND_VIEW_PLUGIN, "gone", connect.CodeNotFound},
		"a person that does not exist": {identityv1.GrantHolderKind_GRANT_HOLDER_KIND_PERSON, "0123456789abcdef", connect.CodeNotFound},
		"no holder kind":               {identityv1.GrantHolderKind_GRANT_HOLDER_KIND_UNSPECIFIED, "ui", connect.CodeInvalidArgument},
		"an unknown holder kind":       {99, "ui", connect.CodeInvalidArgument},
		"a plugin named by a sentence": {identityv1.GrantHolderKind_GRANT_HOLDER_KIND_VIEW_PLUGIN, "my plugin", connect.CodeInvalidArgument},
	} {
		_, err := g.client.CreateGrant(t.Context(), connect.NewRequest(&identityv1.CreateGrantRequest{
			Origin: local, SessionToken: g.token, HolderKind: tc.kind, HolderId: tc.id, TargetKind: identityv1.GrantTargetKind_GRANT_TARGET_KIND_HOUSEHOLD,
		}))
		wantCode(t, "CreateGrant to "+name, err, tc.want)
	}
}

// TestRevokeAndListGrantCalls checks RevokeGrant and ListGrants over Connect:
// a revoked grant stays as history, revoking again changes nothing, a grant
// never made is not found, and the list filters by holder and by whether the
// grant is in force.
func TestRevokeAndListGrantCalls(t *testing.T) {
	t.Parallel()
	g := newGrantsServed(t)
	ctx := t.Context()
	for _, tc := range []struct {
		kind identityv1.GrantTargetKind
		id   string
	}{
		{identityv1.GrantTargetKind_GRANT_TARGET_KIND_ACCOUNT, "main"},
		{identityv1.GrantTargetKind_GRANT_TARGET_KIND_SLEEVE, "core"},
		{identityv1.GrantTargetKind_GRANT_TARGET_KIND_HOUSEHOLD, "home"},
	} {
		if _, err := g.create(t, tc.kind, tc.id); err != nil {
			t.Fatalf("CreateGrant(%v): %v", tc.kind, err)
		}
	}
	sleeveKind := identityv1.GrantTargetKind_GRANT_TARGET_KIND_SLEEVE
	if err := g.revokeVia(sleeveKind, "core"); err != nil {
		t.Fatalf("RevokeGrant: %v", err)
	}
	if err := g.revokeVia(sleeveKind, "core"); err != nil {
		t.Errorf("RevokeGrant again: %v, want nothing to change", err)
	}
	wantCode(t, "RevokeGrant of a grant never made", g.revokeVia(identityv1.GrantTargetKind_GRANT_TARGET_KIND_ACCOUNT, "other"), connect.CodeNotFound)

	plugin := identityv1.GrantHolderKind_GRANT_HOLDER_KIND_VIEW_PLUGIN
	all := []string{"GRANT_TARGET_KIND_ACCOUNT main in force", "GRANT_TARGET_KIND_SLEEVE core ended", "GRANT_TARGET_KIND_HOUSEHOLD home in force"}
	if diff := cmp.Diff(all, g.lines(t, false, 0, "")); diff != "" {
		t.Errorf("ListGrants (-want +got):\n%s", diff)
	}
	if diff := cmp.Diff([]string{all[0], all[2]}, g.lines(t, true, plugin, "ui")); diff != "" {
		t.Errorf("ListGrants of the plugin in force (-want +got):\n%s", diff)
	}
	if got := g.lines(t, false, plugin, "other"); len(got) != 0 {
		t.Errorf("ListGrants of another plugin = %v, want none", got)
	}
	_, err := g.client.ListGrants(ctx, connect.NewRequest(&identityv1.ListGrantsRequest{Origin: local, SessionToken: g.token, HolderId: "ui"}))
	wantCode(t, "ListGrants with a holder ID and no kind", err, connect.CodeInvalidArgument)
	_, err = g.client.ListGrants(ctx, connect.NewRequest(&identityv1.ListGrantsRequest{Origin: local, SessionToken: g.token, HolderKind: 99}))
	wantCode(t, "ListGrants of an unknown holder kind", err, connect.CodeInvalidArgument)
}

// TestGrantAuditOverConnect checks that the audit log, read as the owner,
// shows a grant and a revocation naming the plugin, then the target, by kind
// and ID and nothing else.
func TestGrantAuditOverConnect(t *testing.T) {
	t.Parallel()
	g := newGrantsServed(t)
	if _, err := g.create(t, identityv1.GrantTargetKind_GRANT_TARGET_KIND_ACCOUNT, "main"); err != nil {
		t.Fatalf("CreateGrant: %v", err)
	}
	if _, err := g.create(t, identityv1.GrantTargetKind_GRANT_TARGET_KIND_HOUSEHOLD, ""); err != nil {
		t.Fatalf("CreateGrant: %v", err)
	}
	if err := g.revokeVia(identityv1.GrantTargetKind_GRANT_TARGET_KIND_ACCOUNT, "main"); err != nil {
		t.Fatalf("RevokeGrant: %v", err)
	}
	entries, err := g.client.ListAuditEntries(t.Context(), connect.NewRequest(&identityv1.ListAuditEntriesRequest{Origin: local, SessionToken: g.token}))
	if err != nil {
		t.Fatalf("ListAuditEntries: %v", err)
	}
	trail := make([]string, 0, len(entries.Msg.GetEntries()))
	for _, e := range slices.Backward(entries.Msg.GetEntries()) {
		words := make([]string, 0, 2+len(e.GetRefs()))
		words = append(words, e.GetActor().GetKind().String(), e.GetAction().String())
		for _, r := range e.GetRefs() {
			words = append(words, r.GetKind().String()+":"+r.GetId())
		}
		trail = append(trail, strings.Join(words, " "))
	}
	const (
		owner  = "AUDIT_ACTOR_KIND_OWNER AUDIT_ACTION_GRANT_"
		plugin = " AUDIT_REF_KIND_VIEW_PLUGIN:ui"
	)
	want := []string{
		owner + "CREATED" + plugin + " AUDIT_REF_KIND_ACCOUNT:main",
		owner + "CREATED" + plugin + " AUDIT_REF_KIND_HOUSEHOLD:home",
		owner + "REVOKED" + plugin + " AUDIT_REF_KIND_ACCOUNT:main",
	}
	if diff := cmp.Diff(want, trail); diff != "" {
		t.Errorf("audit entries (-want +got):\n%s", diff)
	}
}
