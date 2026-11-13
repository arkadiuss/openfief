package connectapi_test

import (
	"testing"

	"connectrpc.com/connect"

	identityv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/identity/v1"
)

// TestAccessCallsAreOwnerOnly checks that a demo (read) session, no session, an
// unknown session and a caller that is no view plugin are refused UpdateGrants
// and ListViewPlugins (D-210), and that nothing is granted.
func TestAccessCallsAreOwnerOnly(t *testing.T) {
	t.Parallel()
	g := newGrantsServed(t)
	ctx := t.Context()
	demo := demoSession(t, g.client)
	for _, tc := range []struct {
		name  string
		token string
		want  connect.Code
	}{
		{"a demo (read) session", demo, connect.CodePermissionDenied},
		{"no session", "", connect.CodeUnauthenticated},
	} {
		_, err := g.client.UpdateGrants(ctx, connect.NewRequest(&identityv1.UpdateGrantsRequest{
			Origin: local, SessionToken: tc.token, HolderKind: identityv1.GrantHolderKind_GRANT_HOLDER_KIND_VIEW_PLUGIN, HolderId: "ui",
			Add: []*identityv1.GrantTarget{{Kind: identityv1.GrantTargetKind_GRANT_TARGET_KIND_HOUSEHOLD}},
		}))
		wantCode(t, "UpdateGrants with "+tc.name, err, tc.want)
		_, err = g.client.ListViewPlugins(ctx, connect.NewRequest(&identityv1.ListViewPluginsRequest{Origin: local, SessionToken: tc.token}))
		wantCode(t, "ListViewPlugins with "+tc.name, err, tc.want)
	}
	for name, caller := range map[string]string{"no caller token": "", "a caller that is no view plugin": "connector-token"} {
		_, err := newClient(t, caller).ListViewPlugins(ctx, connect.NewRequest(&identityv1.ListViewPluginsRequest{Origin: local, SessionToken: g.token}))
		wantCode(t, "ListViewPlugins from "+name, err, connect.CodePermissionDenied)
	}
	if got := g.lines(t, true, identityv1.GrantHolderKind_GRANT_HOLDER_KIND_UNSPECIFIED, ""); len(got) != 0 {
		t.Errorf("grants after the refusals = %v, want none", got)
	}
}

// TestUpdateGrantsCall checks UpdateGrants over Connect: it adds and ends
// grants in one call, answers with those in force, and is refused a target that
// isn't there (NOT_FOUND) or an unknown kind (INVALID_ARGUMENT) before changing
// anything. The household's empty ID is read as "home".
func TestUpdateGrantsCall(t *testing.T) {
	t.Parallel()
	g := newGrantsServed(t)
	ctx := t.Context()
	const holder = identityv1.GrantHolderKind_GRANT_HOLDER_KIND_VIEW_PLUGIN
	target := func(kind identityv1.GrantTargetKind, id string) *identityv1.GrantTarget {
		return &identityv1.GrantTarget{Kind: kind, Id: id}
	}
	update := func(add, remove []*identityv1.GrantTarget) (*identityv1.UpdateGrantsResponse, error) {
		resp, err := g.client.UpdateGrants(ctx, connect.NewRequest(&identityv1.UpdateGrantsRequest{
			Origin: local, SessionToken: g.token, HolderKind: holder, HolderId: "ui", Add: add, Remove: remove,
		}))
		if err != nil {
			return nil, err //nolint:wrapcheck // the test asserts on the Connect error itself
		}
		return resp.Msg, nil
	}
	resp, err := update([]*identityv1.GrantTarget{
		target(identityv1.GrantTargetKind_GRANT_TARGET_KIND_HOUSEHOLD, ""),
		target(identityv1.GrantTargetKind_GRANT_TARGET_KIND_SLEEVE, "core"),
	}, nil)
	if err != nil || len(resp.GetGrants()) != 2 {
		t.Fatalf("UpdateGrants = %v, %v, want the two grants in force", resp, err)
	}
	resp, err = update(nil, []*identityv1.GrantTarget{target(identityv1.GrantTargetKind_GRANT_TARGET_KIND_HOUSEHOLD, "home")})
	if err != nil || len(resp.GetGrants()) != 1 || resp.GetGrants()[0].GetTargetId() != "core" {
		t.Fatalf("UpdateGrants removing the household = %v, %v, want the sleeve alone", resp, err)
	}

	_, err = update([]*identityv1.GrantTarget{target(identityv1.GrantTargetKind_GRANT_TARGET_KIND_ACCOUNT, "main"), target(identityv1.GrantTargetKind_GRANT_TARGET_KIND_SLEEVE, "nowhere")}, nil)
	wantCode(t, "UpdateGrants to a sleeve that isn't there", err, connect.CodeNotFound)
	_, err = update([]*identityv1.GrantTarget{{Kind: identityv1.GrantTargetKind_GRANT_TARGET_KIND_UNSPECIFIED, Id: "main"}}, nil)
	wantCode(t, "UpdateGrants to an unspecified kind", err, connect.CodeInvalidArgument)
	_, err = g.client.UpdateGrants(ctx, connect.NewRequest(&identityv1.UpdateGrantsRequest{Origin: local, SessionToken: g.token, HolderId: "ui"}))
	wantCode(t, "UpdateGrants without a holder kind", err, connect.CodeInvalidArgument)
	if got := g.lines(t, true, holder, "ui"); len(got) != 1 {
		t.Errorf("grants after the refusals = %v, want only the sleeve", got)
	}
}

// TestListViewPluginsCall checks that ListViewPlugins shows the plugin with the
// grants it holds, and that grants are not enforced under the policy.
func TestListViewPluginsCall(t *testing.T) {
	t.Parallel()
	g := newGrantsServed(t)
	ctx := t.Context()
	_, err := g.create(t, identityv1.GrantTargetKind_GRANT_TARGET_KIND_SLEEVE, "core")
	if err != nil {
		t.Fatalf("CreateGrant: %v", err)
	}
	list, err := g.client.ListViewPlugins(ctx, connect.NewRequest(&identityv1.ListViewPluginsRequest{Origin: local, SessionToken: g.token}))
	if err != nil {
		t.Fatalf("ListViewPlugins: %v", err)
	}
	plugins := list.Msg.GetPlugins()
	if len(plugins) != 1 || plugins[0].GetInstance() != "ui" || len(plugins[0].GetGrants()) != 1 || list.Msg.GetGrantsEnforced() {
		t.Errorf("ListViewPlugins = %v, want ui with its sleeve, and grants not enforced under the policy", list.Msg)
	}
}
