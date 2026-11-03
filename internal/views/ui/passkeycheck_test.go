package ui

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/google/go-cmp/cmp"
	"google.golang.org/protobuf/types/known/durationpb"

	identityv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/identity/v1"
)

// errPasskeyCheck is the fake core's reason for refusing an action D-183
// protects.
var errPasskeyCheck = errors.New("passkey check required")

// needsCheck refuses an action D-183 protects, as the core does, while the
// fake is stale and no passkey check finished.
func (f *fakeAuth) needsCheck() error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if !f.stale || f.checked {
		return nil
	}
	ce := connect.NewError(connect.CodeFailedPrecondition, errPasskeyCheck)
	detail, err := connect.NewErrorDetail(&identityv1.PasskeyCheckRequired{Window: durationpb.New(passkeyWindow)})
	if err != nil {
		return fmt.Errorf("fake core: %w", err)
	}
	ce.AddDetail(detail)
	return fmt.Errorf("fake core: %w", ce)
}

// passkeyWindow is the five minutes D-183 allows after a passkey check.
const passkeyWindow = 5 * time.Minute

// checkOptions are request options a browser's virtual authenticator can
// answer on localhost: a discoverable-credential assertion with user
// verification.
const checkOptions = `{"challenge":"c3ludGhldGljLWNoYWxsZW5nZQ","rpId":"localhost","userVerification":"required","timeout":60000}`

func (f *fakeAuth) BeginPasskeyCheck(_ context.Context, r *connect.Request[identityv1.BeginPasskeyCheckRequest]) (*connect.Response[identityv1.BeginPasskeyCheckResponse], error) {
	if err := f.record(r.Header(), r.Msg.GetOrigin()); err != nil {
		return nil, err
	}
	if err := owner(r.Msg.GetSessionToken()); err != nil {
		return nil, err
	}
	return connect.NewResponse(&identityv1.BeginPasskeyCheckResponse{CeremonyId: "c-check", OptionsJson: checkOptions}), nil
}

func (f *fakeAuth) FinishPasskeyCheck(_ context.Context, r *connect.Request[identityv1.FinishPasskeyCheckRequest]) (*connect.Response[identityv1.FinishPasskeyCheckResponse], error) {
	if err := f.record(r.Header(), r.Msg.GetOrigin()); err != nil {
		return nil, err
	}
	if err := owner(r.Msg.GetSessionToken()); err != nil {
		return nil, err
	}
	if r.Msg.GetCeremonyId() != "c-check" || r.Msg.GetCredentialJson() == "" {
		return nil, codeErr(connect.CodeFailedPrecondition, nil)
	}
	f.mu.Lock()
	f.checked = true
	f.checks = append(f.checks, r.Msg.GetCredentialJson())
	f.mu.Unlock()
	return connect.NewResponse(&identityv1.FinishPasskeyCheckResponse{Session: issued(identityv1.SessionScope_SESSION_SCOPE_OWNER)}), nil
}

func (f *fakeAuth) SignOutEverywhere(_ context.Context, r *connect.Request[identityv1.SignOutEverywhereRequest]) (*connect.Response[identityv1.SignOutEverywhereResponse], error) {
	if err := f.record(r.Header(), r.Msg.GetOrigin()); err != nil {
		return nil, err
	}
	if err := owner(r.Msg.GetSessionToken()); err != nil {
		return nil, err
	}
	f.mu.Lock()
	f.everywhere = append(f.everywhere, r.Msg.GetSessionToken())
	f.mu.Unlock()
	return connect.NewResponse(&identityv1.SignOutEverywhereResponse{RevokedSessions: 3}), nil
}

// TestPasskeyCheckPrompt checks D-183 in the settings: an action the core
// refuses for want of a recent passkey check shows "Confirm with your
// passkey" with the action and its fields to post again, nothing done; the
// check goes to the core; the same post then goes ahead.
func TestPasskeyCheckPrompt(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	h.auth.locked(func() { h.auth.stale = true })
	actions := []struct {
		name   string
		post   *http.Request
		fields []string
	}{
		{
			name: "script token", post: form("/settings/tokens", "name=nightly&days=7"),
			fields: []string{`action="/settings/tokens"`, `name="name" value="nightly"`, `name="days" value="7"`, "Creating a script token"},
		},
		{
			name: "enrollment link", post: form("/settings/enrollment-links", "target_origin=https%3A%2F%2Fother.example.com"),
			fields: []string{`action="/settings/enrollment-links"`, `name="target_origin" value="https://other.example.com"`},
		},
		{
			name: "remove passkey", post: form("/settings/passkeys/pk1/remove", ""),
			fields: []string{`action="/settings/passkeys/pk1/remove"`, "Removing a passkey"},
		},
	}
	for _, a := range actions {
		rec := h.serve(t, withSession(a.post, ownerSession))
		wantStatus(t, rec, http.StatusForbidden)
		wantBody(t, rec, append([]string{"Confirm with your passkey", `data-webauthn="check"`, `data-retry="passkey-check-retry"`}, a.fields...)...)
		wantNoBody(t, rec, "synthetic-secret", "synthetic-code")
	}
	if got := seen(h.auth, func() []string { return h.auth.removed }); len(got) != 0 {
		t.Errorf("removed %v without a passkey check", got)
	}

	rec := h.serve(t, withSession(request(http.MethodPost, "/passkey-check/begin", `{}`), ownerSession))
	wantStatus(t, rec, http.StatusOK)
	wantBody(t, rec, `"ceremonyId":"c-check"`, "userVerification")
	rec = h.serve(t, withSession(request(http.MethodPost, "/passkey-check/finish", `{"ceremonyId":"c-check","credential":"{\"id\":\"x\"}"}`), ownerSession))
	wantStatus(t, rec, http.StatusOK)
	wantBody(t, rec, `"checked":true`)
	if diff := cmp.Diff([]string{`{"id":"x"}`}, seen(h.auth, func() []string { return h.auth.checks })); diff != "" {
		t.Errorf("passkey checks (-want +got):\n%s", diff)
	}

	rec = h.serve(t, withSession(form("/settings/tokens", "name=nightly&days=7"), ownerSession))
	wantStatus(t, rec, http.StatusOK)
	wantBody(t, rec, "Script token nightly created", "ofs_0123456789abcdef_synthetic-secret")
}

// TestPasskeyCheckRefusals checks that a passkey check needs the owner's
// session, and that adding a passkey from a session tells the page's
// script to run a check first.
func TestPasskeyCheckRefusals(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	h.auth.locked(func() { h.auth.stale = true })
	rec := h.serve(t, withSession(request(http.MethodPost, "/passkey-check/begin", `{}`), readSession))
	wantStatus(t, rec, http.StatusForbidden)
	rec = h.serve(t, request(http.MethodPost, "/passkey-check/begin", `{}`))
	wantStatus(t, rec, http.StatusForbidden)
	rec = h.serve(t, withSession(request(http.MethodPost, "/passkey-check/finish", `{"ceremonyId":"c-login","credential":"{}"}`), ownerSession))
	wantStatus(t, rec, http.StatusBadRequest)

	rec = h.serve(t, withSession(request(http.MethodPost, "/enroll/begin", `{}`), ownerSession))
	wantStatus(t, rec, http.StatusForbidden)
	wantBody(t, rec, `"passkeyCheck":true`)
	h.auth.locked(func() { h.auth.checked = true })
	rec = h.serve(t, withSession(request(http.MethodPost, "/enroll/begin", `{}`), ownerSession))
	wantStatus(t, rec, http.StatusOK)
	wantNoBody(t, rec, "passkeyCheck")
}

// TestRemovePasskeySignsOut checks that removing the passkey that signed in
// this session drops the cookie and opens the login page (D-183).
func TestRemovePasskeySignsOut(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	h.auth.locked(func() { h.auth.ownPasskey = true })
	rec := h.serve(t, withSession(form("/settings/passkeys/pk1/remove", ""), ownerSession))
	wantStatus(t, rec, http.StatusSeeOther)
	if got := rec.Header().Get("Location"); got != "/login" {
		t.Errorf("Location = %q, want /login", got)
	}
	if c := responseCookie(rec); c == nil || c.MaxAge >= 0 {
		t.Errorf("cookie = %v, want it dropped", c)
	}
}

// TestSignOutEverywhere checks the settings' "Sign out everywhere": it
// posts to /logout, so the service worker empties its cache, the core
// revokes every session, and the cookie goes; a read session is refused
// and keeps its cookie.
func TestSignOutEverywhere(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	rec := h.serve(t, withSession(request(http.MethodGet, "/settings", ""), ownerSession))
	wantBody(t, rec, "Sign out everywhere", `name="everywhere" value="yes"`)
	rec = h.serve(t, withSession(form("/logout", "everywhere=yes"), ownerSession))
	wantStatus(t, rec, http.StatusSeeOther)
	if got := rec.Header().Get("Location"); got != "/login" {
		t.Errorf("Location = %q, want /login", got)
	}
	if c := responseCookie(rec); c == nil || c.MaxAge >= 0 {
		t.Errorf("cookie = %v, want it dropped", c)
	}
	if diff := cmp.Diff([]string{ownerSession}, seen(h.auth, func() []string { return h.auth.everywhere })); diff != "" {
		t.Errorf("signed out everywhere (-want +got):\n%s", diff)
	}
	if got := seen(h.auth, func() []string { return h.auth.loggedOut }); len(got) != 0 {
		t.Errorf("also logged out %v alone", got)
	}
	rec = h.serve(t, withSession(form("/logout", "everywhere=yes"), readSession))
	wantStatus(t, rec, http.StatusForbidden)
	if c := responseCookie(rec); c != nil {
		t.Errorf("a refused sign-out everywhere dropped the cookie: %v", c)
	}
}
