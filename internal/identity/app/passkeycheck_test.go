package app_test

import (
	"errors"
	"testing"
	"time"

	"github.com/arkadiuss/openfief/internal/identity"
	"github.com/arkadiuss/openfief/internal/identity/app"
	"github.com/arkadiuss/openfief/internal/identity/webauthn/virtual"
)

// protectedActions are the actions D-183 protects with a recent passkey
// check: enrolling a passkey with a session, issuing an enrollment link and
// creating a script token. Removing a passkey is checked in
// TestPasskeyCheckGuardsRemoval, since it needs a second passkey.
func protectedActions() []ownerAction {
	var out []ownerAction
	for _, a := range ownerActions() {
		switch a.name {
		case "CreateScriptToken", "CreateEnrollmentLink", "BeginRegistration":
			out = append(out, a)
		}
	}
	return out
}

// passkeyCheck runs a passkey check with dev in the session of token.
func (f *authFixture) passkeyCheck(t *testing.T, dev *virtual.Authenticator, token string) (identity.Session, error) {
	t.Helper()
	st, err := f.auth.BeginPasskeyCheck(t.Context(), local, token)
	if err != nil {
		return identity.Session{}, wrap(err)
	}
	assertion, err := dev.Get(st.Options, local)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	s, err := f.auth.FinishPasskeyCheck(t.Context(), local, token, st.CeremonyID, assertion)
	return s, wrap(err)
}

// wantProtected checks every protected action in the session of token:
// allowed when wantErr is nil, otherwise failing with it.
func (f *authFixture) wantProtected(t *testing.T, when, token string, wantErr error) {
	t.Helper()
	for _, a := range protectedActions() {
		err := a.do(t, f, token)
		if wantErr == nil && err != nil {
			t.Errorf("%s %s: %v", a.name, when, err)
		}
		if wantErr != nil && !errors.Is(err, wantErr) {
			t.Errorf("%s %s = %v, want %v", a.name, when, err, wantErr)
		}
	}
}

// TestPasskeyCheckWindow checks D-183: a session that signed in with a
// passkey may take the protected actions for five minutes; after that each
// is refused until a passkey check in the same session, which opens five
// more minutes. Other owner actions need no check.
func TestPasskeyCheckWindow(t *testing.T) {
	t.Parallel()
	f := newAuthFixture(t, demoConfig{})
	dev, ns := f.enroll(t)
	if !ns.Session.PasskeyCheckedAt.Equal(authStart()) || len(ns.Session.PasskeyID) == 0 {
		t.Errorf("a session from an enrollment = %+v, want its passkey and a check at sign-in", ns.Session)
	}
	f.clock.set(authStart().Add(identity.PasskeyCheckWindow - time.Second))
	f.wantProtected(t, "within five minutes of signing in", ns.Token, nil)
	later := authStart().Add(identity.PasskeyCheckWindow)
	f.clock.set(later)
	f.wantProtected(t, "five minutes after signing in", ns.Token, identity.ErrPasskeyCheckRequired)
	f.wantProtected(t, "five minutes after signing in", ns.Token, app.ErrPasskeyCheckRequired)
	for _, a := range ownerActions() {
		if a.name != "ScriptTokens" && a.name != "Passkeys" {
			continue
		}
		if err := a.do(t, f, ns.Token); err != nil {
			t.Errorf("%s needs no passkey check: %v", a.name, err)
		}
	}
	checked, err := f.passkeyCheck(t, dev, ns.Token)
	if err != nil {
		t.Fatalf("passkey check: %v", err)
	}
	if !checked.PasskeyCheckedAt.Equal(later) {
		t.Errorf("checked at %s, want %s", checked.PasskeyCheckedAt, later)
	}
	f.wantProtected(t, "after a passkey check", ns.Token, nil)
}

// TestLoginCountsAsPasskeyCheck checks that a new login is a check of its
// own session only.
func TestLoginCountsAsPasskeyCheck(t *testing.T) {
	t.Parallel()
	f := newAuthFixture(t, demoConfig{})
	dev, ns := f.enroll(t)
	f.clock.set(authStart().Add(time.Hour))
	fresh, err := f.login(t, dev, local)
	if err != nil {
		t.Fatalf("login: %v", err)
	}
	f.wantProtected(t, "right after a login", fresh.Token, nil)
	f.wantProtected(t, "in the older session", ns.Token, app.ErrPasskeyCheckRequired)
}

// TestPasskeyCheckGuardsRemoval checks that removing a passkey needs a
// recent passkey check too (D-183).
func TestPasskeyCheckGuardsRemoval(t *testing.T) {
	t.Parallel()
	f := newAuthFixture(t, demoConfig{})
	dev, ns := f.enroll(t)
	if _, _, err := f.register(t, local, app.RegistrationGrant{SessionToken: ns.Token}); err != nil {
		t.Fatalf("adding a passkey: %v", err)
	}
	second := f.store.passkeys[1].ID()
	f.clock.set(authStart().Add(identity.PasskeyCheckWindow))
	if _, err := f.auth.RemovePasskey(t.Context(), local, ns.Token, second); !errors.Is(err, app.ErrPasskeyCheckRequired) {
		t.Errorf("RemovePasskey without a recent check = %v, want %v", err, app.ErrPasskeyCheckRequired)
	}
	if len(f.store.passkeys) != 2 {
		t.Fatalf("a refused removal removed a passkey")
	}
	if _, err := f.passkeyCheck(t, dev, ns.Token); err != nil {
		t.Fatalf("passkey check: %v", err)
	}
	if _, err := f.auth.RemovePasskey(t.Context(), local, ns.Token, second); err != nil {
		t.Errorf("RemovePasskey after a check: %v", err)
	}
}

// TestPasskeyCheckNeedsOwnerSession checks that a passkey check needs an
// owner session: a demo session or none can't begin one.
func TestPasskeyCheckNeedsOwnerSession(t *testing.T) {
	t.Parallel()
	f := newAuthFixture(t, demoConfig{on: true, marked: true})
	read, loginErr := f.auth.Login(t.Context(), f.auth.Demo(), local, app.Proof{})
	if loginErr != nil {
		t.Fatalf("demo login: %v", loginErr)
	}
	if _, err := f.auth.BeginPasskeyCheck(t.Context(), local, read.Token); !errors.Is(err, identity.ErrSessionScope) {
		t.Errorf("BeginPasskeyCheck with a demo session = %v, want %v", err, identity.ErrSessionScope)
	}
	if _, err := f.auth.BeginPasskeyCheck(t.Context(), local, ""); !errors.Is(err, app.ErrInvalidSession) {
		t.Errorf("BeginPasskeyCheck without a session = %v, want %v", err, app.ErrInvalidSession)
	}
}

// TestPasskeyCheckCeremonyRefused checks that a passkey check is a ceremony
// of its own kind and verifies the assertion; a failed check records
// nothing.
func TestPasskeyCheckCeremonyRefused(t *testing.T) {
	t.Parallel()
	f := newAuthFixture(t, demoConfig{})
	dev, ns := f.enroll(t)
	f.clock.set(authStart().Add(identity.PasskeyCheckWindow))
	login := f.begunLogin(t, dev, local)
	if _, err := f.auth.FinishPasskeyCheck(t.Context(), local, ns.Token, login.CeremonyID, login.Credential); !errors.Is(err, app.ErrCeremony) {
		t.Errorf("a check finished with a login's ceremony = %v, want %v", err, app.ErrCeremony)
	}
	st, beginErr := f.auth.BeginPasskeyCheck(t.Context(), local, ns.Token)
	if beginErr != nil {
		t.Fatalf("BeginPasskeyCheck: %v", beginErr)
	}
	assertion, getErr := dev.Get(st.Options, local)
	if getErr != nil {
		t.Fatalf("Get: %v", getErr)
	}
	if _, err := f.auth.Login(t.Context(), f.auth.Passkey(), local, app.Proof{CeremonyID: st.CeremonyID, Credential: assertion}); !errors.Is(err, app.ErrCeremony) {
		t.Errorf("a login finished with a check's ceremony = %v, want %v", err, app.ErrCeremony)
	}
	elsewhere := *dev
	elsewhere.RPID = "dev.example.com"
	if _, err := f.passkeyCheck(t, &elsewhere, ns.Token); !errors.Is(err, app.ErrAuthenticationFailed) {
		t.Errorf("a check with an assertion for another host = %v, want %v", err, app.ErrAuthenticationFailed)
	}
	if _, err := f.auth.CreateEnrollmentLink(t.Context(), local, ns.Token, "", 0); !errors.Is(err, app.ErrPasskeyCheckRequired) {
		t.Errorf("CreateEnrollmentLink after a failed check = %v, want %v", err, app.ErrPasskeyCheckRequired)
	}
}

// TestSignOutEverywhere checks that the owner can revoke every session at
// once, on every host and of either scope, their own included (D-183), and
// that a read session can't.
func TestSignOutEverywhere(t *testing.T) {
	t.Parallel()
	f := newAuthFixture(t, demoConfig{on: true, marked: true})
	dev, ns := f.enroll(t)
	other, loginErr := f.login(t, dev, local)
	if loginErr != nil {
		t.Fatalf("login: %v", loginErr)
	}
	read, demoErr := f.auth.Login(t.Context(), f.auth.Demo(), devo, app.Proof{})
	if demoErr != nil {
		t.Fatalf("demo login: %v", demoErr)
	}
	if _, err := f.auth.SignOutEverywhere(t.Context(), devo, read.Token); !errors.Is(err, identity.ErrSessionScope) {
		t.Errorf("SignOutEverywhere with a read session = %v, want %v", err, identity.ErrSessionScope)
	}
	// Long after signing in: no passkey check is needed to sign out.
	f.clock.set(authStart().Add(time.Hour))
	if n, err := f.auth.SignOutEverywhere(t.Context(), local, other.Token); err != nil || n != 3 {
		t.Errorf("SignOutEverywhere = %d, %v; want 3 sessions revoked", n, err)
	}
	for _, s := range []struct{ origin, token string }{{local, ns.Token}, {local, other.Token}, {devo, read.Token}} {
		if _, _, err := f.auth.WhoAmI(t.Context(), s.origin, s.token); !errors.Is(err, identity.ErrSessionRevoked) {
			t.Errorf("WhoAmI on %s after signing out everywhere = %v, want %v", s.origin, err, identity.ErrSessionRevoked)
		}
	}
	if _, err := f.login(t, dev, local); err != nil {
		t.Errorf("logging in again: %v", err)
	}
}
