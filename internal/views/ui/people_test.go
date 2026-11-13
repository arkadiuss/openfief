package ui

import (
	"errors"
	"net/http"
	"regexp"
	"strings"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/google/go-cmp/cmp"
	"google.golang.org/protobuf/testing/protocmp"
	"google.golang.org/protobuf/types/known/durationpb"

	identityv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/identity/v1"
)

// errNameTaken is the fake core's refusal of a viewer's name.
var errNameTaken = errors.New("a viewer needs a name of 1 to 64 printable characters")

// errAlreadyRevoked is the fake core's refusal of revoking a viewer twice.
var errAlreadyRevoked = errors.New("no such viewer, or already revoked")

// TestAccessPeopleSection checks the People section of the owner's Access page
// (D-210): the owner first, each viewer with their status, what they may view in
// words, their sessions and passkeys, "no access yet" beside a viewer with no
// grant, a grants form and Revoke on an active viewer, and a revoked viewer
// still listed, with neither.
func TestAccessPeopleSection(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	rec := h.serve(t, withSession(request(http.MethodGet, "/access", ""), ownerSession))
	wantStatus(t, rec, http.StatusOK)
	wantBody(t, rec,
		"People", "Owner", "Every account",
		"Alice", "Sleeve Core", "Account Synthetic brokerage (synthetic-broker)",
		"Bob", "no access yet", "Carol", "revoked",
		`action="/access/people/`+alicePerson+`/grants"`, `action="/access/people/`+alicePerson+`/revoke"`,
		`action="/access/people/`+bobPerson+`/grants"`, `action="/access/people"`,
		`name="grant" value="sleeve:core" checked`, `id="person-`+alicePerson+`-sleeve-1"`,
		`name="minutes"`, `value="15" selected`, `name="grant" value="household:home"`, "Create invitation link",
	)
	// The "no access yet" mark is Bob's alone, and the revoked viewer and the
	// owner have no form to change or revoke them.
	if n := strings.Count(rec.Body.String(), "no access yet"); n != 1 {
		t.Errorf("the page marks %d people \"no access yet\", want one (Bob)", n)
	}
	wantNoBody(t, rec, `action="/access/people/`+carolPerson+`/`, `action="/access/people/owner/`)
}

// TestInviteViewer checks that the invitation form reaches the core as one
// InviteViewer with the name, the grants and the lifetime asked for, shows the
// one-time link with its expiry once and never again, and that a lifetime or a
// grant out of range never reaches the core.
func TestInviteViewer(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	rec := h.serve(t, withSession(form("/access/people", "name=Dana&minutes=60&grant=sleeve%3Acore&grant=sleeve%3Acore&grant=account%3A"+accountID), ownerSession))
	wantStatus(t, rec, http.StatusOK)
	wantBody(t, rec, "Invitation for Dana created", testOrigin+"/enroll#code=synthetic-invite-code", "2026-03-02 12:30 UTC", "shown only this once")
	var req *identityv1.InviteViewerRequest
	h.auth.locked(func() { req = h.auth.inviteReq })
	want := &identityv1.InviteViewerRequest{
		Origin: testOrigin, SessionToken: ownerSession, Name: "Dana", Ttl: durationMinutes(60),
		Grants: []*identityv1.GrantTarget{
			{Kind: identityv1.GrantTargetKind_GRANT_TARGET_KIND_SLEEVE, Id: "core"},
			{Kind: identityv1.GrantTargetKind_GRANT_TARGET_KIND_ACCOUNT, Id: accountID},
		},
	}
	if diff := cmp.Diff(want, req, protocmp.Transform()); diff != "" {
		t.Errorf("InviteViewer request (-want +got):\n%s", diff)
	}
	rec = h.serve(t, withSession(request(http.MethodGet, "/access", ""), ownerSession))
	wantNoBody(t, rec, "synthetic-invite-code", "Invitation for Dana")

	// A viewer who starts with no grant is allowed, with the default lifetime.
	rec = h.serve(t, withSession(form("/access/people", "name=Erin"), ownerSession))
	wantStatus(t, rec, http.StatusOK)
	h.auth.locked(func() { req = h.auth.inviteReq })
	if req.GetName() != "Erin" || len(req.GetGrants()) != 0 || req.GetTtl().AsDuration() != 15*time.Minute {
		t.Errorf("InviteViewer = %v, want Erin with no grants for 15 minutes", req)
	}

	sent := req
	for name, body := range map[string]string{
		"a lifetime that isn't offered": "name=Dana&minutes=5000",
		"a lifetime not a number":       "name=Dana&minutes=soon",
		"an unknown grant kind":         "name=Dana&grant=nobody%3Aroot",
		"a household not home":          "name=Dana&grant=household%3Aother",
		"an overlong grant ID":          "name=Dana&grant=sleeve%3A" + longID(),
	} {
		rec = h.serve(t, withSession(form("/access/people", body), ownerSession))
		if rec.Code != http.StatusBadRequest {
			t.Errorf("%s: status = %d, want 400", name, rec.Code)
		}
	}
	h.auth.locked(func() {
		if h.auth.inviteReq != sent {
			t.Error("a refused form reached the core")
		}
	})
}

// TestInviteViewerRefusals checks the core's answers to an invitation: its
// reason shows with the form as the owner filled it, and a missing passkey check
// (D-183) asks the owner to confirm with their passkey and posts the same form
// again, which then goes ahead; no link exists until the core made it.
func TestInviteViewerRefusals(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	h.auth.locked(func() { h.auth.inviteErr = codeErr(connect.CodeInvalidArgument, errNameTaken) })
	rec := h.serve(t, withSession(form("/access/people", "name=&minutes=60&grant=sleeve%3Acore"), ownerSession))
	wantStatus(t, rec, http.StatusBadRequest)
	wantBody(t, rec, "a viewer needs a name", `name="grant" value="sleeve:core" checked`, `value="60" selected`)
	wantNoBody(t, rec, "synthetic-invite-code")
	h.auth.locked(func() { h.auth.inviteErr = nil })

	h.auth.locked(func() { h.auth.stale = true })
	rec = h.serve(t, withSession(form("/access/people", "name=Dana&minutes=60&grant=sleeve%3Acore&grant=household%3Ahome"), ownerSession))
	wantStatus(t, rec, http.StatusForbidden)
	wantBody(t, rec, "Confirm with your passkey", "Inviting a viewer", `data-webauthn="check"`, `data-retry="passkey-check-retry"`,
		`action="/access/people"`, `name="name" value="Dana"`, `name="minutes" value="60"`,
		`name="grant" value="sleeve:core"`, `name="grant" value="household:home"`)
	wantNoBody(t, rec, "synthetic-invite-code", "Invitation for Dana")
	h.auth.locked(func() {
		if h.auth.inviteReq != nil {
			t.Error("an invitation reached the core's model without a passkey check")
		}
	})

	h.auth.locked(func() { h.auth.checked = true })
	rec = h.serve(t, withSession(form("/access/people", "name=Dana&minutes=60&grant=sleeve%3Acore"), ownerSession))
	wantStatus(t, rec, http.StatusOK)
	wantBody(t, rec, "Invitation for Dana created", "synthetic-invite-code")
}

// TestRevokePerson checks that revoking a viewer takes a confirmation: the first
// post shows what revoking does and reaches the core with nothing, the confirmed
// one revokes them with no passkey check, and a viewer the page doesn't offer
// (the owner, a revoked viewer, a stranger) is not found; the core's refusal
// shows on the page.
func TestRevokePerson(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	h.auth.locked(func() { h.auth.stale = true })
	revoked := func() []string { return seen(h.auth, func() []string { return h.auth.revokedPeople }) }

	rec := h.serve(t, withSession(form("/access/people/"+alicePerson+"/revoke", ""), ownerSession))
	wantStatus(t, rec, http.StatusOK)
	wantBody(t, rec, "Revoke Alice?", "2 active session(s)", "1 passkey(s)", "Yes, revoke Alice", `name="confirm" value="yes"`, `action="/access/people/`+alicePerson+`/revoke"`)
	if got := revoked(); len(got) != 0 {
		t.Fatalf("revoked %v before the owner confirmed", got)
	}

	rec = h.serve(t, withSession(form("/access/people/"+alicePerson+"/revoke", "confirm=yes"), ownerSession))
	wantStatus(t, rec, http.StatusSeeOther)
	if got := rec.Header().Get("Location"); got != "/access#people-title" {
		t.Errorf("Location = %q, want /access#people-title", got)
	}
	if diff := cmp.Diff([]string{alicePerson}, revoked()); diff != "" {
		t.Errorf("revoked people (-want +got):\n%s", diff)
	}

	for _, id := range []string{"owner", carolPerson, "ffffffffffffffff"} {
		rec = h.serve(t, withSession(form("/access/people/"+id+"/revoke", ""), ownerSession))
		wantStatus(t, rec, http.StatusNotFound)
	}
	h.auth.locked(func() { h.auth.revokePersonErr = codeErr(connect.CodeNotFound, errAlreadyRevoked) })
	rec = h.serve(t, withSession(form("/access/people/"+bobPerson+"/revoke", "confirm=yes"), ownerSession))
	wantStatus(t, rec, http.StatusNotFound)
	wantBody(t, rec, "already revoked", "People")
	if diff := cmp.Diff([]string{alicePerson}, revoked()); diff != "" {
		t.Errorf("revoked people after the refusals (-want +got):\n%s", diff)
	}
}

// TestUpdateViewerGrants checks that a viewer's grants form reaches the core as
// one UpdateGrants of a person holder with exactly the boxes changed, needing no
// passkey check, and that a revoked viewer or a stranger is not found.
func TestUpdateViewerGrants(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	h.auth.locked(func() { h.auth.stale = true })
	const kind = "GRANT_HOLDER_KIND_PERSON/"
	tests := []struct {
		name, person, body string
		want               []string
	}{
		{
			"alice drops her account for the household", alicePerson, "grant=household%3Ahome&grant=sleeve%3Acore",
			[]string{kind + alicePerson + " +GRANT_TARGET_KIND_HOUSEHOLD/home -GRANT_TARGET_KIND_ACCOUNT/" + accountID},
		},
		{
			"bob, who sees nothing, gains a sleeve", bobPerson, "grant=sleeve%3Abonds",
			[]string{kind + bobPerson + " +GRANT_TARGET_KIND_SLEEVE/bonds"},
		},
		{
			"alice with nothing checked loses everything", alicePerson, "",
			[]string{kind + alicePerson + " -GRANT_TARGET_KIND_SLEEVE/core -GRANT_TARGET_KIND_ACCOUNT/" + accountID},
		},
		{"a form that changes nothing", alicePerson, "grant=sleeve%3Acore&grant=account%3A" + accountID, nil},
	}
	for _, tt := range tests {
		h.auth.locked(func() { h.auth.updates = nil })
		rec := h.serve(t, withSession(form("/access/people/"+tt.person+"/grants", tt.body), ownerSession))
		wantStatus(t, rec, http.StatusSeeOther)
		if got := rec.Header().Get("Location"); got != "/access#people-title" {
			t.Errorf("%s: Location = %q, want /access#people-title", tt.name, got)
		}
		if diff := cmp.Diff(tt.want, seen(h.auth, func() []string { return h.auth.updates })); diff != "" {
			t.Errorf("%s: UpdateGrants calls (-want +got):\n%s", tt.name, diff)
		}
	}
	h.auth.locked(func() { h.auth.updates = nil })
	for _, id := range []string{carolPerson, "owner", "ffffffffffffffff"} {
		rec := h.serve(t, withSession(form("/access/people/"+id+"/grants", "grant=household%3Ahome"), ownerSession))
		wantStatus(t, rec, http.StatusNotFound)
	}
	rec := h.serve(t, withSession(form("/access/people/"+alicePerson+"/grants", "grant=bank%3Aall"), ownerSession))
	wantStatus(t, rec, http.StatusBadRequest)
	if got := seen(h.auth, func() []string { return h.auth.updates }); len(got) != 0 {
		t.Errorf("UpdateGrants calls = %v, want none", got)
	}
}

// TestPeopleActionsAreTheOwners checks that a viewer's, a read session's and a
// signed-out browser's posts to the People section are refused, and none of them
// reaches the core: the section's routes take the owner's guard like the rest of
// the Access page.
func TestPeopleActionsAreTheOwners(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	posts := []string{
		"/access/people", "/access/people/" + alicePerson + "/revoke", "/access/people/" + alicePerson + "/grants",
	}
	for _, session := range []string{viewerSession, emptyViewerSession, readSession, demoSession} {
		for _, p := range posts {
			rec := h.serve(t, withSession(form(p, "name=Dana&confirm=yes&grant=household%3Ahome"), session))
			if rec.Code != http.StatusForbidden {
				t.Errorf("POST %s with session %s = %d, want 403", p, session, rec.Code)
			}
		}
		rec := h.serve(t, withSession(request(http.MethodGet, "/access", ""), session))
		wantStatus(t, rec, http.StatusForbidden)
		wantNoBody(t, rec, "Alice", "Bob", "Carol")
	}
	h.postsNeedALogin(t, posts)
	h.auth.locked(func() {
		if h.auth.inviteReq != nil || len(h.auth.revokedPeople) != 0 || len(h.auth.updates) != 0 {
			t.Errorf("a refused post reached the core: invite %v, revoked %v, updates %v", h.auth.inviteReq, h.auth.revokedPeople, h.auth.updates)
		}
	})
}

// TestViewerWithNothingSharedSeesAnEmptyState checks D-210's viewer with no
// grant: the overview, accounts and sleeves pages say plainly that nothing has
// been shared with them yet, with no error and no chart, while a viewer who has
// something is shown it.
func TestViewerWithNothingSharedSeesAnEmptyState(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	h.writeToken(t, pluginToken)
	canvas := regexp.MustCompile(`<canvas|class="alert alert-error"`)
	for _, path := range []string{"/", "/accounts", "/sleeves"} {
		rec := h.serve(t, withSession(request(http.MethodGet, path, ""), emptyViewerSession))
		wantStatus(t, rec, http.StatusOK)
		wantBody(t, rec, "Nothing has been shared with you yet", "Viewer access", "Bob")
		wantNoBody(t, rec, `href="/settings"`, `href="/access"`, "Synthetic brokerage", "No accounts yet")
		if canvas.MatchString(rec.Body.String()) {
			t.Errorf("GET %s showed a chart or an error to a viewer who was shared nothing", path)
		}
	}
	for _, path := range []string{"/", "/accounts"} {
		rec := h.serve(t, withSession(request(http.MethodGet, path, ""), viewerSession))
		wantStatus(t, rec, http.StatusOK)
		wantNoBody(t, rec, "Nothing has been shared with you yet")
	}
	// The owner without accounts is not told that anything was shared.
	h.read.locked(func() { h.read.noAccounts = true })
	rec := h.serve(t, withSession(request(http.MethodGet, "/accounts", ""), ownerSession))
	wantStatus(t, rec, http.StatusOK)
	wantNoBody(t, rec, "Nothing has been shared with you yet")
}

// postsNeedALogin checks that posting to each path without a session ends at the login page, not in a form's answer.
func (h *harness) postsNeedALogin(t *testing.T, posts []string) {
	t.Helper()
	for _, p := range posts {
		rec := h.serve(t, form(p, "name=Dana"))
		if loc := rec.Header().Get("Location"); rec.Code != http.StatusSeeOther || !strings.HasPrefix(loc, "/login") {
			t.Errorf("POST %s without a session = %d to %q, want the login page", p, rec.Code, loc)
		}
	}
}

// durationMinutes is the proto duration of n minutes.
func durationMinutes(n int) *durationpb.Duration {
	return durationpb.New(time.Duration(n) * time.Minute)
}
