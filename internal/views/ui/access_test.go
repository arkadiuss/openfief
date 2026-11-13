package ui

import (
	"errors"
	"net/http"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/google/go-cmp/cmp"
	"google.golang.org/protobuf/testing/protocmp"
	"google.golang.org/protobuf/types/known/durationpb"

	identityv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/identity/v1"
)

// errSpiceDBDown is the fake core's reason for a grant the authorization
// model did not take.
var errSpiceDBDown = errors.New("stored, but not passed on to the authorization model")

// errNoSuchSleeve is the fake core's refusal of a grant to a sleeve that isn't there.
var errNoSuchSleeve = errors.New("no such sleeve")

// TestAccessPage checks the owner's Access page (D-210): each script token
// with its status, expiry, MCP flag and grants in words, the view plugins with
// theirs, the grants forms for each, the form of a new token, and, while grants
// have no effect, the note that says so.
func TestAccessPage(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	rec := h.serve(t, withSession(request(http.MethodGet, "/access", ""), ownerSession))
	wantStatus(t, rec, http.StatusOK)
	wantBody(t, rec,
		"Script tokens", "backup-script", "fedcba9876543210", "active", "2026-02-01", "allowed", "Sleeve Core",
		`action="/access/tokens/fedcba9876543210/revoke"`, `action="/access/tokens/fedcba9876543210/grants"`,
		"old-script", "revoked",
		"View plugins", "ghostfolio", "Account Synthetic brokerage (synthetic-broker)", "sheets", "Nothing",
		`action="/access/plugins/ghostfolio/grants"`, `action="/access/plugins/sheets/grants"`,
		`action="/access/tokens"`, `name="grant" value="household:home"`, "Whole household",
		`name="grant" value="sleeve:core"`, `name="grant" value="account:`+accountID+`"`, `name="mcp"`, `name="days"`,
		"Grants take effect when SpiceDB authorization is on", `href="/access" aria-current="page"`,
	)
	// A revoked token has no way to change or revoke.
	wantNoBody(t, rec, `action="/access/tokens/0011223344556677/revoke"`, `action="/access/tokens/0011223344556677/grants"`)

	h.auth.locked(func() { h.auth.enforced = true })
	rec = h.serve(t, withSession(request(http.MethodGet, "/access", ""), ownerSession))
	wantStatus(t, rec, http.StatusOK)
	wantNoBody(t, rec, "Grants take effect when SpiceDB authorization is on")
}

// TestAccessPageViewPluginsUnavailable checks that a core that can't read its
// cluster leaves the tokens in place and says so where the plugins would be.
func TestAccessPageViewPluginsUnavailable(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	h.auth.locked(func() { h.auth.pluginsDown = true })
	rec := h.serve(t, withSession(request(http.MethodGet, "/access", ""), ownerSession))
	wantStatus(t, rec, http.StatusOK)
	wantBody(t, rec, "backup-script", "The installed view plugins can't be read now")
	wantNoBody(t, rec, `action="/access/plugins/`)
}

// TestAccessIsOwnerOnly checks that the Access page is in the owner's navigation
// and in no read session's.
func TestAccessIsOwnerOnly(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	rec := h.serve(t, withSession(request(http.MethodGet, "/settings", ""), ownerSession))
	wantBody(t, rec, `<a href="/access">Access</a>`)
	rec = h.serve(t, withSession(request(http.MethodGet, "/sleeves", ""), readSession))
	wantStatus(t, rec, http.StatusOK)
	wantNoBody(t, rec, `href="/access"`)
	rec = h.serve(t, withSession(request(http.MethodGet, "/access", ""), readSession))
	wantStatus(t, rec, http.StatusForbidden)
	wantNoBody(t, rec, "backup-script", "ghostfolio")
	h.auth.locked(func() {
		if len(h.auth.updates) != 0 {
			t.Errorf("updates = %v, want none", h.auth.updates)
		}
	})
}

// TestCreateScriptToken checks that a new script token's secret is shown once,
// in the response to its creation, with the lifetime, the MCP flag and the
// grants the form asked for; that a lifetime or a grant out of range never
// reaches the core; and that the secret is not on the page afterwards.
func TestCreateScriptToken(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	rec := h.serve(t, withSession(form("/access/tokens", "name=nightly&days=7&mcp=on&grant=sleeve%3Acore&grant=account%3A"+accountID+"&grant=sleeve%3Acore"), ownerSession))
	wantStatus(t, rec, http.StatusOK)
	wantBody(t, rec, "Script token nightly created", "ofs_0123456789abcdef_synthetic-secret", "shown only this once")
	var req *identityv1.CreateScriptTokenRequest
	h.auth.locked(func() { req = h.auth.tokenReq })
	want := &identityv1.CreateScriptTokenRequest{
		Origin: testOrigin, SessionToken: ownerSession, Name: "nightly", Ttl: durationDays(7), McpAllowed: true,
		Grants: []*identityv1.GrantTarget{
			{Kind: identityv1.GrantTargetKind_GRANT_TARGET_KIND_SLEEVE, Id: "core"},
			{Kind: identityv1.GrantTargetKind_GRANT_TARGET_KIND_ACCOUNT, Id: accountID},
		},
	}
	if diff := cmp.Diff(want, req, protocmp.Transform()); diff != "" {
		t.Errorf("CreateScriptToken request (-want +got):\n%s", diff)
	}
	rec = h.serve(t, withSession(request(http.MethodGet, "/access", ""), ownerSession))
	wantNoBody(t, rec, "synthetic-secret")

	for name, body := range map[string]string{
		"days over a year":  "name=nightly&days=400",
		"days not a number": "name=nightly&days=soon",
	} {
		rec = h.serve(t, withSession(form("/access/tokens", body), ownerSession))
		if rec.Code != http.StatusBadRequest {
			t.Errorf("%s: status = %d, want 400", name, rec.Code)
		}
		wantBody(t, rec, "from 1 to 365", `value="nightly"`)
	}
	for name, grant := range map[string]string{
		"an unknown kind":      "nobody%3Aroot",
		"no kind":              "core",
		"an empty ID":          "sleeve%3A",
		"a household not home": "household%3Aother",
		"an overlong ID":       "sleeve%3A" + longID(),
	} {
		rec = h.serve(t, withSession(form("/access/tokens", "name=x&grant="+grant), ownerSession))
		if rec.Code != http.StatusBadRequest {
			t.Errorf("%s: status = %d, want 400", name, rec.Code)
		}
	}
	h.auth.locked(func() {
		if h.auth.tokenReq != req {
			t.Error("a refused form reached the core")
		}
	})
}

// longID is an ID longer than any a grant target has.
func longID() string {
	b := make([]byte, maxChoiceChars)
	for i := range b {
		b[i] = 'a'
	}
	return string(b)
}

// TestCreateScriptTokenRefusals checks the core's answers to a new token: a
// grant to a sleeve that isn't there shows the core's reason and keeps the form;
// an authorization model that did not take the grants says nothing was kept and
// shows no secret.
func TestCreateScriptTokenRefusals(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	h.auth.locked(func() { h.auth.createErr = codeErr(connect.CodeNotFound, errNoSuchSleeve) })
	rec := h.serve(t, withSession(form("/access/tokens", "name=nightly&days=7&grant=sleeve%3Agone"), ownerSession))
	wantStatus(t, rec, http.StatusNotFound)
	wantBody(t, rec, "no such sleeve", `value="nightly"`, `value="7"`)

	h.auth.locked(func() { h.auth.createErr = codeErr(connect.CodeUnavailable, errSpiceDBDown) })
	rec = h.serve(t, withSession(form("/access/tokens", "name=nightly&days=7&grant=household%3Ahome"), ownerSession))
	wantStatus(t, rec, http.StatusServiceUnavailable)
	wantBody(t, rec, "The token was not created", "no secret exists", `value="nightly"`)
	wantNoBody(t, rec, "synthetic-secret", "Script token nightly created")
}

// TestRevokeScriptToken checks revoking a script token from the Access page,
// and the core's refusal shown there.
func TestRevokeScriptToken(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	rec := h.serve(t, withSession(form("/access/tokens/fedcba9876543210/revoke", ""), ownerSession))
	wantStatus(t, rec, http.StatusSeeOther)
	if got := rec.Header().Get("Location"); got != "/access#tokens-title" {
		t.Errorf("Location = %q, want /access#tokens-title", got)
	}
	if diff := cmp.Diff([]string{"fedcba9876543210"}, seen(h.auth, func() []string { return h.auth.revoked })); diff != "" {
		t.Errorf("revoked (-want +got):\n%s", diff)
	}
}

// TestUpdateGrants checks that the grants form of a token or a view plugin
// reaches the core as one UpdateGrants with exactly what the owner changed: the
// boxes newly checked are added, those unchecked are ended, those left alone are
// not sent, and a form that changes nothing sends nothing.
func TestUpdateGrants(t *testing.T) {
	t.Parallel()
	const (
		tokenKind  = "GRANT_HOLDER_KIND_SCRIPT_TOKEN/fedcba9876543210"
		pluginKind = "GRANT_HOLDER_KIND_VIEW_PLUGIN/ghostfolio"
	)
	tests := []struct {
		name     string
		path     string
		body     string
		want     []string
		location string
	}{
		{
			name: "a token drops its sleeve for the household", path: "/access/tokens/fedcba9876543210/grants", body: "grant=household%3Ahome",
			want:     []string{tokenKind + " +GRANT_TARGET_KIND_HOUSEHOLD/home -GRANT_TARGET_KIND_SLEEVE/core"},
			location: "/access#tokens-title",
		},
		{
			name: "a token keeps its sleeve and gains an account", path: "/access/tokens/fedcba9876543210/grants", body: "grant=sleeve%3Acore&grant=account%3A" + accountID,
			want:     []string{tokenKind + " +GRANT_TARGET_KIND_ACCOUNT/" + accountID},
			location: "/access#tokens-title",
		},
		{
			name: "a token with nothing checked loses its grants", path: "/access/tokens/fedcba9876543210/grants", body: "",
			want:     []string{tokenKind + " -GRANT_TARGET_KIND_SLEEVE/core"},
			location: "/access#tokens-title",
		},
		{
			name: "a token form that changes nothing", path: "/access/tokens/fedcba9876543210/grants", body: "grant=sleeve%3Acore",
			location: "/access#tokens-title",
		},
		{
			name: "a plugin gains the household", path: "/access/plugins/ghostfolio/grants", body: "grant=household%3Ahome&grant=account%3A" + accountID,
			want:     []string{pluginKind + " +GRANT_TARGET_KIND_HOUSEHOLD/home"},
			location: "/access#view-plugins-title",
		},
		{
			name: "a plugin with no grant gains a sleeve", path: "/access/plugins/sheets/grants", body: "grant=sleeve%3Abonds",
			want:     []string{"GRANT_HOLDER_KIND_VIEW_PLUGIN/sheets +GRANT_TARGET_KIND_SLEEVE/bonds"},
			location: "/access#view-plugins-title",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			h := newHarness(t)
			rec := h.serve(t, withSession(form(tt.path, tt.body), ownerSession))
			wantStatus(t, rec, http.StatusSeeOther)
			if got := rec.Header().Get("Location"); got != tt.location {
				t.Errorf("Location = %q, want %q", got, tt.location)
			}
			if diff := cmp.Diff(tt.want, seen(h.auth, func() []string { return h.auth.updates })); diff != "" {
				t.Errorf("UpdateGrants calls (-want +got):\n%s", diff)
			}
		})
	}
}

// TestUpdateGrantsRefusals checks the grants forms' failures: a holder the core
// doesn't list (a revoked token, an uninstalled plugin) is 404 before any change
// reaches it, a malformed choice is 400, and the core's refusal or an
// authorization model that did not take the change shows on the page.
func TestUpdateGrantsRefusals(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	for _, path := range []string{"/access/tokens/0011223344556677/grants", "/access/tokens/ffffffffffffffff/grants", "/access/plugins/gone/grants"} {
		rec := h.serve(t, withSession(form(path, "grant=household%3Ahome"), ownerSession))
		wantStatus(t, rec, http.StatusNotFound)
	}
	rec := h.serve(t, withSession(form("/access/plugins/ghostfolio/grants", "grant=bank%3Aall"), ownerSession))
	wantStatus(t, rec, http.StatusBadRequest)
	if got := seen(h.auth, func() []string { return h.auth.updates }); len(got) != 0 {
		t.Errorf("UpdateGrants calls = %v, want none", got)
	}

	h.auth.locked(func() { h.auth.updateErr = codeErr(connect.CodeNotFound, errNoSuchSleeve) })
	rec = h.serve(t, withSession(form("/access/plugins/ghostfolio/grants", "grant=sleeve%3Agone"), ownerSession))
	wantStatus(t, rec, http.StatusNotFound)
	wantBody(t, rec, "no such sleeve", "View plugins")

	h.auth.locked(func() { h.auth.updateErr = codeErr(connect.CodeUnavailable, errSpiceDBDown) })
	rec = h.serve(t, withSession(form("/access/plugins/ghostfolio/grants", "grant=household%3Ahome"), ownerSession))
	wantStatus(t, rec, http.StatusServiceUnavailable)
	wantBody(t, rec, "The change is stored, but SpiceDB did not take it yet", "Save it again")
}

// TestUngrantedItemLinksToAccess checks the Inbox item of a view plugin that
// can't read anything (D-196) links to the view plugins of the Access page,
// where the owner picks narrower grants than the whole household.
func TestUngrantedItemLinksToAccess(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	ungrantedItem(h, 60, "ghostfolio")
	rec := h.serve(t, withSession(request(http.MethodGet, "/inbox", ""), ownerSession))
	wantStatus(t, rec, http.StatusOK)
	wantBody(t, rec, `href="/access#view-plugins-title"`, "Open Access")
}

// durationDays is the proto duration of n days.
func durationDays(n int) *durationpb.Duration {
	return durationpb.New(time.Duration(n) * hoursPerDay * time.Hour)
}
