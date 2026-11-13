package app_test

import (
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"github.com/arkadiuss/openfief/internal/identity"
	"github.com/arkadiuss/openfief/internal/identity/app"
)

var errAuditDown = errors.New("audit log unavailable")

// trail returns the audit log as one line per entry, oldest first: actor
// kind, action, outcome and the kinds of the records it names.
func (f *authFixture) trail() []string {
	entries := f.store.entries()
	out := make([]string, 0, len(entries))
	for _, e := range entries {
		words := make([]string, 0, 3+len(e.Refs))
		words = append(words, string(e.Actor.Kind), string(e.Action), string(e.Outcome))
		for _, r := range e.Refs {
			words = append(words, string(r.Kind))
		}
		out = append(out, strings.Join(words, " "))
	}
	return out
}

// must fails the test now when err is not nil.
func must(t *testing.T, what string, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("%s: %v", what, err)
	}
}

func TestAuditTrail(t *testing.T) {
	t.Parallel()
	f := newAuthFixture(t, demoConfig{})
	ctx := t.Context()
	dev, first := f.enroll(t)
	f.clock.set(authStart().Add(time.Minute))
	second, err := f.login(t, dev, local)
	must(t, "login", err)
	tok, secret, err := f.auth.CreateScriptToken(ctx, local, second.Token, app.NewToken{Name: "nightly"})
	must(t, "CreateScriptToken", err)
	must(t, "RevokeScriptToken", f.auth.RevokeScriptToken(ctx, local, second.Token, string(tok.ID)))
	_, err = f.auth.CreateEnrollmentLink(ctx, local, second.Token, devo, 0)
	must(t, "CreateEnrollmentLink", err)
	must(t, "Logout", f.auth.Logout(ctx, local, first.Token))
	_, err = f.auth.SignOutEverywhere(ctx, local, second.Token)
	must(t, "SignOutEverywhere", err)
	want := []string{
		"system enrollment_link_issued ok enrollment_code",
		"enrollment_code enrollment_link_used ok enrollment_code passkey",
		"enrollment_code passkey_added ok passkey",
		"owner sign_in ok session",
		"owner sign_in ok session",
		"owner script_token_created ok script_token",
		"owner script_token_revoked ok script_token",
		"owner enrollment_link_issued ok enrollment_code",
		"owner sign_out ok session",
		"owner sign_out_everywhere ok",
	}
	if diff := cmp.Diff(want, f.trail()); diff != "" {
		t.Errorf("audit trail (-want +got):\n%s", diff)
	}
	// Every entry is a valid record, and none holds the token's secret.
	for i, e := range f.store.entries() {
		if validErr := e.Validate(); validErr != nil {
			t.Errorf("entry %d is not valid: %v", i, validErr)
		}
		if strings.Contains(fmt.Sprintf("%+v", e), secret) {
			t.Errorf("entry %d holds the script token secret", i)
		}
	}
}

// TestRegistrationBySessionNamesTheSession checks that a passkey added in an
// owner session is recorded under that session, which the ceremony carries
// from its begin to its finish.
func TestRegistrationBySessionNamesTheSession(t *testing.T) {
	t.Parallel()
	f := newAuthFixture(t, demoConfig{})
	_, first := f.enroll(t)
	if _, r, err := f.register(t, local, app.RegistrationGrant{SessionToken: first.Token}); err != nil || r.Session != nil {
		t.Fatalf("registering in a session = %+v, %v; want a passkey and no new session", r, err)
	}
	entries := f.store.entries()
	added := entries[len(entries)-1]
	want := identity.Actor{Kind: identity.ActorOwner, ID: string(first.Session.ID)}
	if added.Action != identity.ActionPasskeyAdded || added.Actor != want {
		t.Errorf("last entry = %+v, want passkey_added by %+v", added, want)
	}
}

func TestRefusedSignInIsRecorded(t *testing.T) {
	t.Parallel()
	f := newAuthFixture(t, demoConfig{})
	f.enroll(t)
	before := len(f.store.entries())
	ctx := t.Context()
	refusals := map[string]func() error{
		"an unknown ceremony": func() error {
			_, err := f.auth.Login(ctx, f.auth.Passkey(), local, app.Proof{CeremonyID: "nope", Credential: []byte("{}")})
			return wrap(err)
		},
		"an origin off the list": func() error {
			_, err := f.auth.Login(ctx, f.auth.Demo(), "https://evil.example.org", app.Proof{})
			return wrap(err)
		},
		"demo sessions off": func() error {
			_, err := f.auth.Login(ctx, f.auth.Demo(), local, app.Proof{})
			return wrap(err)
		},
		"a bad assertion": func() error {
			st, err := f.auth.BeginLogin(ctx, local)
			if err != nil {
				return wrap(err)
			}
			_, err = f.auth.Login(ctx, f.auth.Passkey(), local, app.Proof{CeremonyID: st.CeremonyID, Credential: []byte(`{"id":"x"}`)})
			return wrap(err)
		},
	}
	for name, refuse := range refusals {
		if err := refuse(); err == nil {
			t.Errorf("%s: the sign-in was not refused", name)
		}
	}
	entries := f.store.entries()[before:]
	if len(entries) != len(refusals) {
		t.Fatalf("%d entries for %d refusals", len(entries), len(refusals))
	}
	for _, e := range entries {
		wantEntry := identity.AuditEntry{Seq: e.Seq, At: e.At, Actor: identity.Actor{Kind: identity.ActorAnonymous}, Action: identity.ActionSignIn, Outcome: identity.OutcomeRefused}
		if diff := cmp.Diff(wantEntry, e); diff != "" {
			t.Errorf("refused sign-in entry (-want +got):\n%s", diff)
		}
	}
}

// TestRefusedSignInStaysRefusedWhenTheLogIsDown checks the choice for the one
// event with no change of state to share a transaction with: the sign-in is
// refused whether or not its entry is written, and the failed write is
// reported next to the refusal.
func TestRefusedSignInStaysRefusedWhenTheLogIsDown(t *testing.T) {
	t.Parallel()
	f := newAuthFixture(t, demoConfig{})
	f.store.failAudit(errAuditDown)
	_, err := f.auth.Login(t.Context(), f.auth.Passkey(), local, app.Proof{CeremonyID: "nope", Credential: []byte("{}")})
	if !errors.Is(err, app.ErrCeremony) {
		t.Errorf("Login = %v, want it to wrap %v: still a refusal", err, app.ErrCeremony)
	}
	if !errors.Is(err, errAuditDown) {
		t.Errorf("Login = %v, want it to carry the failed audit write", err)
	}
}

// TestAuditFailsClosed checks that each change of state fails, and leaves no
// trace, when its audit entry can't be written.
func TestAuditFailsClosed(t *testing.T) {
	t.Parallel()
	f := newAuthFixture(t, demoConfig{})
	ctx := t.Context()
	dev, first := f.enroll(t)
	_, second, err := f.register(t, local, app.RegistrationGrant{SessionToken: first.Token})
	if err != nil {
		t.Fatalf("registering a second passkey: %v", err)
	}
	proof := f.begunLogin(t, dev, local)
	sessions, passkeys, entries := len(f.store.sessions), len(f.store.passkeys), len(f.store.entries())
	f.store.failAudit(errAuditDown)
	actions := map[string]func() error{
		"sign-in": func() error {
			_, err := f.auth.Login(ctx, f.auth.Passkey(), local, proof)
			return wrap(err)
		},
		"sign-out": func() error { return f.auth.Logout(ctx, local, first.Token) },
		"sign-out everywhere": func() error {
			_, err := f.auth.SignOutEverywhere(ctx, local, first.Token)
			return wrap(err)
		},
		"passkey removal": func() error {
			_, err := f.auth.RemovePasskey(ctx, local, first.Token, second.Passkey.ID())
			return wrap(err)
		},
		"script token creation": func() error {
			_, _, err := f.auth.CreateScriptToken(ctx, local, first.Token, app.NewToken{Name: "nightly"})
			return wrap(err)
		},
		"enrollment link": func() error {
			_, err := f.auth.CreateEnrollmentLink(ctx, local, first.Token, devo, 0)
			return wrap(err)
		},
	}
	for name, act := range actions {
		if err := act(); !errors.Is(err, errAuditDown) {
			t.Errorf("%s = %v, want %v", name, err, errAuditDown)
		}
	}
	if got := len(f.store.sessions); got != sessions {
		t.Errorf("%d sessions after the failed changes, want %d", got, sessions)
	}
	if got := len(f.store.passkeys); got != passkeys {
		t.Errorf("%d passkeys after the failed changes, want %d", got, passkeys)
	}
	if got := len(f.store.entries()); got != entries {
		t.Errorf("%d entries after the failed changes, want %d", got, entries)
	}
	if _, _, err := f.auth.WhoAmI(ctx, local, first.Token); err != nil {
		t.Errorf("the session after the failed sign-outs: %v, want it still valid", err)
	}
}

// TestFailedRegistrationLeavesNoCodeUsed checks that an enrollment code is
// not used up by a registration whose audit entry fails: the owner can try
// again with the same link.
func TestFailedRegistrationLeavesNoCodeUsed(t *testing.T) {
	t.Parallel()
	f := newAuthFixture(t, demoConfig{})
	code := f.link(t, local, 0)
	f.store.failAudit(errAuditDown)
	if _, _, err := f.register(t, local, app.RegistrationGrant{EnrollmentCode: code}); !errors.Is(err, errAuditDown) {
		t.Fatalf("registering = %v, want %v", err, errAuditDown)
	}
	if len(f.store.passkeys) != 0 {
		t.Errorf("%d passkeys stored by a failed registration, want none", len(f.store.passkeys))
	}
	f.store.failAudit(nil)
	if _, r, err := f.register(t, local, app.RegistrationGrant{EnrollmentCode: code}); err != nil || r.Session == nil {
		t.Errorf("registering again with the same link = %+v, %v; want it to work", r, err)
	}
}

func TestAuditEntriesOwnerOnly(t *testing.T) {
	t.Parallel()
	f := newAuthFixture(t, demoConfig{on: true, marked: true})
	ctx := t.Context()
	_, owner := f.enroll(t)
	demo, err := f.auth.Login(ctx, f.auth.Demo(), local, app.Proof{})
	must(t, "demo login", err)
	all := &identity.AuditFilter{}
	_, err = f.auth.AuditEntries(ctx, local, demo.Token, all)
	if !errors.Is(err, app.ErrForbidden) {
		t.Errorf("AuditEntries in a demo (read) session = %v, want %v", err, app.ErrForbidden)
	}
	junk := map[string]string{"no session": "", "a junk token": "ofs_junk", "a script token secret": "ofs_0123456789abcdef_" + strings.Repeat("A", 43)}
	for name, token := range junk {
		_, err = f.auth.AuditEntries(ctx, local, token, all)
		if !errors.Is(err, app.ErrInvalidSession) {
			t.Errorf("AuditEntries with %s = %v, want %v", name, err, app.ErrInvalidSession)
		}
	}
	_, err = f.auth.AuditEntries(ctx, "https://evil.example.org", owner.Token, all)
	if !errors.Is(err, app.ErrForbidden) {
		t.Errorf("AuditEntries on another origin = %v, want %v", err, app.ErrForbidden)
	}
	_, err = f.auth.AuditEntries(ctx, local, owner.Token, &identity.AuditFilter{Limit: -3})
	if !errors.Is(err, app.ErrInvalidRequest) {
		t.Errorf("AuditEntries with a bad filter = %v, want %v", err, app.ErrInvalidRequest)
	}
}

// TestAuditEntriesFilteredAndNotLogged checks an owner session reads the log
// through a filter, and that reading is no entry of its own.
func TestAuditEntriesFilteredAndNotLogged(t *testing.T) {
	t.Parallel()
	f := newAuthFixture(t, demoConfig{on: true, marked: true})
	ctx := t.Context()
	_, owner := f.enroll(t)
	_, err := f.auth.Login(ctx, f.auth.Demo(), local, app.Proof{})
	must(t, "demo login", err)
	page, err := f.auth.AuditEntries(ctx, local, owner.Token, &identity.AuditFilter{Action: identity.ActionSignIn})
	must(t, "AuditEntries", err)
	if len(page.Entries) != 2 {
		t.Errorf("%d sign-in entries, want 2 (the owner's and the demo's)", len(page.Entries))
	}
	before := len(f.store.entries())
	_, err = f.auth.AuditEntries(ctx, local, owner.Token, &identity.AuditFilter{})
	must(t, "AuditEntries", err)
	if got := len(f.store.entries()); got != before {
		t.Errorf("reading the log wrote %d entries", got-before)
	}
}
