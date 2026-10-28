package app_test

import (
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/arkadiuss/openfief/internal/identity"
	"github.com/arkadiuss/openfief/internal/identity/app"
	"github.com/arkadiuss/openfief/internal/identity/webauthn"
	"github.com/arkadiuss/openfief/internal/identity/webauthn/virtual"
)

// The synthetic origins the tests allow.
const (
	local = "http://localhost:8080"
	devo  = "https://dev.example.com"
)

// authStart is when the tests' clock starts.
func authStart() time.Time { return time.Date(2026, 10, 28, 10, 0, 0, 0, time.UTC) }

// authFixture is the auth use cases over in-memory stores, the real
// WebAuthn ceremonies and a settable clock.
type authFixture struct {
	auth  *app.Auth
	store *memAuthStore
	clock *clock
}

// demoConfig is whether the demo authenticator is on, and whether the
// database is marked as a demo.
type demoConfig struct {
	on, marked bool
}

func newAuthFixture(t *testing.T, demo demoConfig) *authFixture {
	t.Helper()
	origins, err := identity.NewOrigins([]string{local, devo})
	if err != nil {
		t.Fatalf("NewOrigins: %v", err)
	}
	clk := &clock{t: authStart()}
	store := newMemAuthStore()
	cfg := app.AuthConfig{Origins: origins, Sessions: identity.DefaultSessionPolicy(), DemoSessions: demo.on}
	auth, err := app.NewAuth(store, webauthn.New("OpenFief"), app.NewTokens(newMemStore(), clk.now, nil), demoMarker(demo.marked), cfg, clk.now, nil)
	if err != nil {
		t.Fatalf("NewAuth: %v", err)
	}
	return &authFixture{auth: auth, store: store, clock: clk}
}

// origin parses text.
func origin(t *testing.T, text string) identity.Origin {
	t.Helper()
	o, err := identity.ParseOrigin(text)
	if err != nil {
		t.Fatalf("ParseOrigin: %v", err)
	}
	return o
}

// link issues an enrollment link for target and returns its code.
func (f *authFixture) link(t *testing.T, target string, ttl time.Duration) string {
	t.Helper()
	l, err := app.IssueEnrollmentLink(t.Context(), f.store, nil, f.clock.now(), origin(t, target), ttl)
	if err != nil {
		t.Fatalf("IssueEnrollmentLink: %v", err)
	}
	prefix := target + identity.EnrollPath + "#code="
	code, ok := strings.CutPrefix(l.URL, prefix)
	if !ok {
		t.Fatalf("link %q does not start with %q", l.URL, prefix)
	}
	return code
}

// register creates a passkey on originText as grant allows, with a new
// virtual authenticator.
func (f *authFixture) register(t *testing.T, originText string, grant app.RegistrationGrant) (*virtual.Authenticator, app.Registered, error) {
	t.Helper()
	dev, err := virtual.New(nil)
	if err != nil {
		t.Fatalf("virtual.New: %v", err)
	}
	st, err := f.auth.BeginRegistration(t.Context(), originText, grant)
	if err != nil {
		return nil, app.Registered{}, fmt.Errorf("BeginRegistration: %w", err)
	}
	cred, err := dev.Create(st.Options, originText)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	r, err := f.auth.FinishRegistration(t.Context(), originText, st.CeremonyID, cred, "")
	if err != nil {
		return nil, app.Registered{}, fmt.Errorf("FinishRegistration: %w", err)
	}
	return dev, r, nil
}

// enroll creates the first passkey on localhost through an enrollment link
// and returns the authenticator and the owner session that comes with it.
func (f *authFixture) enroll(t *testing.T) (*virtual.Authenticator, app.NewSession) {
	t.Helper()
	dev, r, err := f.register(t, local, app.RegistrationGrant{EnrollmentCode: f.link(t, local, 0)})
	if err != nil {
		t.Fatalf("registering with an enrollment code: %v", err)
	}
	if r.Session == nil {
		t.Fatal("a registration with an enrollment code issued no session")
	}
	return dev, *r.Session
}

// begunLogin begins a login on originText and returns its ceremony ID and
// dev's assertion.
func (f *authFixture) begunLogin(t *testing.T, dev *virtual.Authenticator, originText string) app.Proof {
	t.Helper()
	st, err := f.auth.BeginLogin(t.Context(), originText)
	if err != nil {
		t.Fatalf("BeginLogin: %v", err)
	}
	assertion, err := dev.Get(st.Options, originText)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	return app.Proof{CeremonyID: st.CeremonyID, Credential: assertion}
}

// login logs in with dev on originText.
func (f *authFixture) login(t *testing.T, dev *virtual.Authenticator, originText string) (app.NewSession, error) {
	t.Helper()
	ns, err := f.auth.Login(t.Context(), f.auth.Passkey(), originText, f.begunLogin(t, dev, originText))
	if err != nil {
		return app.NewSession{}, fmt.Errorf("Login: %w", err)
	}
	return ns, nil
}

func TestEnrollThenLogin(t *testing.T) {
	t.Parallel()
	f := newAuthFixture(t, demoConfig{})
	dev, first := f.enroll(t)
	if first.Session.Scope != identity.SessionScopeOwner || first.Session.RPID != "localhost" ||
		!strings.HasPrefix(first.Token, identity.SessionPrefix) {
		t.Errorf("enrollment session = %+v", first.Session)
	}
	if keys := f.store.passkeys; len(keys) != 1 || keys[0].Name != "localhost" || keys[0].RPID != "localhost" {
		t.Fatalf("stored passkeys = %+v", keys)
	}
	checkLoginAfterAnHour(t, f, dev)
}

// checkLoginAfterAnHour logs in with dev an hour after the start and checks
// the session and the passkey's recorded use.
func checkLoginAfterAnHour(t *testing.T, f *authFixture, dev *virtual.Authenticator) {
	t.Helper()
	f.clock.set(authStart().Add(time.Hour))
	ns, err := f.login(t, dev, local)
	if err != nil {
		t.Fatal(err)
	}
	owner, s, err := f.auth.WhoAmI(t.Context(), local, ns.Token)
	if err != nil {
		t.Fatalf("WhoAmI: %v", err)
	}
	if s.Scope != identity.SessionScopeOwner || s.Authenticator != identity.AuthenticatorPasskey || len(owner.Handle) != identity.OwnerHandleBytes {
		t.Errorf("WhoAmI = %+v, %+v", owner, s)
	}
	if got := f.store.passkeys[0]; got.SignCount != 1 || !got.LastUsedAt.Equal(authStart().Add(time.Hour)) {
		t.Errorf("after login the passkey has sign count %d, last used %s", got.SignCount, got.LastUsedAt)
	}
}

func TestLoginMethods(t *testing.T) {
	t.Parallel()
	f := newAuthFixture(t, demoConfig{})
	f.enroll(t)
	tests := []struct {
		origin string
		want   app.LoginMethods
	}{
		{origin: local, want: app.LoginMethods{PasskeyRegistered: true}},
		{origin: devo, want: app.LoginMethods{}},
	}
	for _, tt := range tests {
		got, err := f.auth.LoginMethods(t.Context(), tt.origin)
		if err != nil || got != tt.want {
			t.Errorf("LoginMethods(%s) = %+v, %v; want %+v", tt.origin, got, err, tt.want)
		}
	}
	if _, err := f.auth.LoginMethods(t.Context(), "https://evil.example.org"); !errors.Is(err, app.ErrForbidden) {
		t.Errorf("LoginMethods off the list = %v, want %v", err, app.ErrForbidden)
	}
}

func TestEnrollmentCodeRefused(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		// use returns the code to begin with on local, after doing what
		// the case needs.
		use func(t *testing.T, f *authFixture) string
	}{
		{name: "no code and no session", use: func(*testing.T, *authFixture) string { return "" }},
		{name: "malformed", use: func(*testing.T, *authFixture) string { return "ofenroll_xyz" }},
		{name: "unknown", use: func(*testing.T, *authFixture) string {
			return identity.FormatEnrollmentCode("0123456789abcdef", make([]byte, identity.KeyBytes))
		}},
		{name: "wrong key", use: func(t *testing.T, f *authFixture) string {
			t.Helper()
			id, _, err := identity.ParseEnrollmentCode(f.link(t, local, 0))
			if err != nil {
				t.Fatalf("ParseEnrollmentCode: %v", err)
			}
			return identity.FormatEnrollmentCode(id, make([]byte, identity.KeyBytes))
		}},
		{name: "expired", use: func(t *testing.T, f *authFixture) string {
			t.Helper()
			code := f.link(t, local, time.Minute)
			f.clock.set(authStart().Add(time.Minute))
			return code
		}},
		{name: "issued for another origin", use: func(t *testing.T, f *authFixture) string {
			t.Helper()
			return f.link(t, devo, 0)
		}},
		{name: "already used", use: func(t *testing.T, f *authFixture) string {
			t.Helper()
			code := f.link(t, local, 0)
			if _, _, err := f.register(t, local, app.RegistrationGrant{EnrollmentCode: code}); err != nil {
				t.Fatalf("first use: %v", err)
			}
			return code
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			f := newAuthFixture(t, demoConfig{})
			code := tt.use(t, f)
			if _, _, err := f.register(t, local, app.RegistrationGrant{EnrollmentCode: code}); !errors.Is(err, app.ErrForbidden) {
				t.Errorf("register = %v, want %v", err, app.ErrForbidden)
			}
		})
	}
}

func TestEnrollmentCodeUsedOnceAcrossCeremonies(t *testing.T) {
	t.Parallel()
	f := newAuthFixture(t, demoConfig{})
	grant := app.RegistrationGrant{EnrollmentCode: f.link(t, local, 0)}
	type begun struct {
		dev *virtual.Authenticator
		st  app.Started
	}
	ceremonies := make([]begun, 2)
	for i := range ceremonies {
		dev, err := virtual.New(nil)
		if err != nil {
			t.Fatalf("virtual.New: %v", err)
		}
		st, err := f.auth.BeginRegistration(t.Context(), local, grant)
		if err != nil {
			t.Fatalf("BeginRegistration %d: %v", i, err)
		}
		ceremonies[i] = begun{dev: dev, st: st}
	}
	for i, wantErr := range []error{nil, app.ErrForbidden} {
		cred, err := ceremonies[i].dev.Create(ceremonies[i].st.Options, local)
		if err != nil {
			t.Fatalf("Create: %v", err)
		}
		if _, err := f.auth.FinishRegistration(t.Context(), local, ceremonies[i].st.CeremonyID, cred, "laptop"); !errors.Is(err, wantErr) {
			t.Errorf("FinishRegistration %d = %v, want %v", i, err, wantErr)
		}
	}
	if len(f.store.passkeys) != 1 {
		t.Errorf("stored %d passkeys from one code, want 1", len(f.store.passkeys))
	}
}

func TestPasskeysPerHost(t *testing.T) {
	t.Parallel()
	f := newAuthFixture(t, demoConfig{})
	dev, ns := f.enroll(t)
	if _, err := f.login(t, dev, devo); !errors.Is(err, app.ErrAuthenticationFailed) {
		t.Errorf("login on another host with a localhost passkey = %v, want %v", err, app.ErrAuthenticationFailed)
	}
	if _, err := f.auth.Passkeys(t.Context(), devo, ns.Token); !errors.Is(err, app.ErrInvalidSession) {
		t.Errorf("a localhost session used on another host = %v, want %v", err, app.ErrInvalidSession)
	}
	// A localhost owner session issues a link for the new host (D-60).
	link, err := f.auth.CreateEnrollmentLink(t.Context(), local, ns.Token, devo, 0)
	if err != nil {
		t.Fatalf("CreateEnrollmentLink: %v", err)
	}
	code, ok := strings.CutPrefix(link.URL, devo+"/enroll#code=")
	if !ok || !link.ExpiresAt.Equal(authStart().Add(identity.DefaultEnrollmentTTL)) {
		t.Fatalf("link = %+v", link)
	}
	devDevice, r, err := f.register(t, devo, app.RegistrationGrant{EnrollmentCode: code})
	if err != nil || r.Passkey.RPID != "dev.example.com" {
		t.Fatalf("registering on the new host = %+v, %v", r, err)
	}
	if _, err := f.login(t, devDevice, devo); err != nil {
		t.Errorf("login on the new host: %v", err)
	}
	if _, err := f.login(t, devDevice, local); !errors.Is(err, app.ErrAuthenticationFailed) {
		t.Errorf("login on localhost with the new host's passkey = %v, want %v", err, app.ErrAuthenticationFailed)
	}
}

func TestSignCountRegressed(t *testing.T) {
	t.Parallel()
	f := newAuthFixture(t, demoConfig{})
	dev, _ := f.enroll(t)
	if _, err := f.login(t, dev, local); err != nil {
		t.Fatalf("first login: %v", err)
	}
	dev.SignCount = 0 // a clone that has signed less often
	_, err := f.login(t, dev, local)
	if !errors.Is(err, identity.ErrSignCountRegressed) || !errors.Is(err, app.ErrAuthenticationFailed) {
		t.Errorf("login with a regressed counter = %v, want %v", err, identity.ErrSignCountRegressed)
	}
}

func TestSyncedPasskeyKeepsAZeroCounter(t *testing.T) {
	t.Parallel()
	f := newAuthFixture(t, demoConfig{})
	dev, err := virtual.New(nil)
	if err != nil {
		t.Fatalf("virtual.New: %v", err)
	}
	dev.Synced = true
	st, err := f.auth.BeginRegistration(t.Context(), local, app.RegistrationGrant{EnrollmentCode: f.link(t, local, 0)})
	if err != nil {
		t.Fatalf("BeginRegistration: %v", err)
	}
	cred, err := dev.Create(st.Options, local)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if _, err := f.auth.FinishRegistration(t.Context(), local, st.CeremonyID, cred, "phone"); err != nil {
		t.Fatalf("FinishRegistration: %v", err)
	}
	for range 2 {
		if _, err := f.login(t, dev, local); err != nil {
			t.Errorf("login with a synced passkey: %v", err)
		}
	}
}

func TestCeremonyRefused(t *testing.T) {
	t.Parallel()
	f := newAuthFixture(t, demoConfig{})
	dev, _ := f.enroll(t)
	proof := f.begunLogin(t, dev, local)
	if _, err := f.auth.Login(t.Context(), f.auth.Passkey(), local, proof); err != nil {
		t.Fatalf("first finish: %v", err)
	}
	if _, err := f.auth.Login(t.Context(), f.auth.Passkey(), local, proof); !errors.Is(err, app.ErrCeremony) {
		t.Errorf("second finish = %v, want %v", err, app.ErrCeremony)
	}
	late := f.begunLogin(t, dev, local)
	f.clock.set(authStart().Add(identity.CeremonyTTL))
	if _, err := f.auth.Login(t.Context(), f.auth.Passkey(), local, late); !errors.Is(err, app.ErrCeremony) {
		t.Errorf("finish after expiry = %v, want %v", err, app.ErrCeremony)
	}
	if _, err := f.auth.BeginLogin(t.Context(), "https://evil.example.org"); !errors.Is(err, app.ErrForbidden) {
		t.Errorf("BeginLogin on an origin off the list = %v, want %v", err, app.ErrForbidden)
	}
}

func TestSessionLifetime(t *testing.T) {
	t.Parallel()
	f := newAuthFixture(t, demoConfig{})
	_, ns := f.enroll(t)
	// Used every 11 hours, it outlives the 12-hour idle timeout...
	for at := 11 * time.Hour; at < identity.DefaultSessionMax; at += 11 * time.Hour {
		if err := f.whoAt(t, ns.Token, at); err != nil {
			t.Fatalf("WhoAmI at %s: %v", at, err)
		}
	}
	// ...but not the 7-day absolute expiry.
	if err := f.whoAt(t, ns.Token, identity.DefaultSessionMax); !errors.Is(err, identity.ErrSessionExpired) {
		t.Errorf("WhoAmI at the absolute expiry = %v, want %v", err, identity.ErrSessionExpired)
	}
	g := newAuthFixture(t, demoConfig{})
	_, idle := g.enroll(t)
	if err := g.whoAt(t, idle.Token, identity.DefaultSessionIdle); !errors.Is(err, identity.ErrSessionIdle) || !errors.Is(err, app.ErrInvalidSession) {
		t.Errorf("WhoAmI after 12 idle hours = %v, want %v", err, identity.ErrSessionIdle)
	}
}

// whoAt calls WhoAmI with token at the start plus at.
func (f *authFixture) whoAt(t *testing.T, token string, at time.Duration) error {
	t.Helper()
	f.clock.set(authStart().Add(at))
	if _, _, err := f.auth.WhoAmI(t.Context(), local, token); err != nil {
		return fmt.Errorf("WhoAmI: %w", err)
	}
	return nil
}

func TestLogout(t *testing.T) {
	t.Parallel()
	f := newAuthFixture(t, demoConfig{})
	_, ns := f.enroll(t)
	if err := f.auth.Logout(t.Context(), local, ns.Token); err != nil {
		t.Fatalf("Logout: %v", err)
	}
	if _, _, err := f.auth.WhoAmI(t.Context(), local, ns.Token); !errors.Is(err, identity.ErrSessionRevoked) {
		t.Errorf("WhoAmI after Logout = %v, want %v", err, identity.ErrSessionRevoked)
	}
	if _, err := f.auth.Authenticate(t.Context(), ns.Token); !errors.Is(err, app.ErrInvalidToken) {
		t.Errorf("read API with a revoked session = %v, want %v", err, app.ErrInvalidToken)
	}
}

// ownerAction is one owner-only use case, called with a session token.
type ownerAction struct {
	name string
	do   func(t *testing.T, f *authFixture, token string) error
}

// ownerActions are the owner-only use cases.
func ownerActions() []ownerAction {
	return []ownerAction{
		{name: "CreateScriptToken", do: func(t *testing.T, f *authFixture, token string) error {
			t.Helper()
			_, _, err := f.auth.CreateScriptToken(t.Context(), local, token, "script", 0)
			return wrap(err)
		}},
		{name: "ScriptTokens", do: func(t *testing.T, f *authFixture, token string) error {
			t.Helper()
			_, err := f.auth.ScriptTokens(t.Context(), local, token)
			return wrap(err)
		}},
		{name: "Passkeys", do: func(t *testing.T, f *authFixture, token string) error {
			t.Helper()
			_, err := f.auth.Passkeys(t.Context(), local, token)
			return wrap(err)
		}},
		{name: "CreateEnrollmentLink", do: func(t *testing.T, f *authFixture, token string) error {
			t.Helper()
			_, err := f.auth.CreateEnrollmentLink(t.Context(), local, token, "", 0)
			return wrap(err)
		}},
		{name: "BeginRegistration", do: func(t *testing.T, f *authFixture, token string) error {
			t.Helper()
			_, err := f.auth.BeginRegistration(t.Context(), local, app.RegistrationGrant{SessionToken: token})
			return wrap(err)
		}},
	}
}

// wrap wraps a non-nil use case error.
func wrap(err error) error {
	if err != nil {
		return fmt.Errorf("use case: %w", err)
	}
	return nil
}

func TestOwnerActions(t *testing.T) {
	t.Parallel()
	f := newAuthFixture(t, demoConfig{on: true, marked: true})
	_, owner := f.enroll(t)
	read, err := f.auth.Login(t.Context(), f.auth.Demo(), local, app.Proof{})
	if err != nil {
		t.Fatalf("demo login: %v", err)
	}
	unknown := "ofsess_0123456789abcdef_" + strings.Repeat("A", 43)
	for _, a := range ownerActions() {
		// With no session at all, BeginRegistration has no grant: forbidden.
		if err := a.do(t, f, ""); !errors.Is(err, app.ErrInvalidSession) && !errors.Is(err, app.ErrForbidden) {
			t.Errorf("%s without a session = %v, want %v", a.name, err, app.ErrInvalidSession)
		}
		if err := a.do(t, f, unknown); !errors.Is(err, app.ErrInvalidSession) {
			t.Errorf("%s with an unknown session = %v, want %v", a.name, err, app.ErrInvalidSession)
		}
		if err := a.do(t, f, read.Token); !errors.Is(err, app.ErrForbidden) || !errors.Is(err, identity.ErrSessionScope) {
			t.Errorf("%s with a read session = %v, want %v", a.name, err, app.ErrForbidden)
		}
		if err := a.do(t, f, owner.Token); err != nil {
			t.Errorf("%s with an owner session: %v", a.name, err)
		}
	}
	if _, _, err := f.auth.WhoAmI(t.Context(), local, read.Token); err != nil {
		t.Errorf("WhoAmI with a read session: %v", err)
	}
}

func TestScriptTokensFromSession(t *testing.T) {
	t.Parallel()
	f := newAuthFixture(t, demoConfig{})
	_, ns := f.enroll(t)
	ctx := t.Context()
	tok, secret, err := f.auth.CreateScriptToken(ctx, local, ns.Token, "nightly", 0)
	if err != nil {
		t.Fatalf("CreateScriptToken: %v", err)
	}
	if !tok.ExpiresAt.Equal(tok.CreatedAt.Add(app.DefaultScriptTokenTTL)) {
		t.Errorf("default lifetime: expires %s", tok.ExpiresAt)
	}
	for _, token := range []string{secret, ns.Token} {
		if _, err := f.auth.Authenticate(ctx, token); err != nil {
			t.Errorf("Authenticate(%.8s...) = %v", token, err)
		}
	}
	if _, _, err := f.auth.CreateScriptToken(ctx, local, ns.Token, "bad name", 0); !errors.Is(err, app.ErrInvalidRequest) {
		t.Errorf("bad name = %v, want %v", err, app.ErrInvalidRequest)
	}
	if err := f.auth.RevokeScriptToken(ctx, local, ns.Token, string(tok.ID)); err != nil {
		t.Fatalf("RevokeScriptToken: %v", err)
	}
	if _, err := f.auth.Authenticate(ctx, secret); !errors.Is(err, app.ErrInvalidToken) {
		t.Errorf("Authenticate after revoke = %v, want %v", err, app.ErrInvalidToken)
	}
	if err := f.auth.RevokeScriptToken(ctx, local, ns.Token, "0123456789abcdef"); !errors.Is(err, app.ErrNotFound) {
		t.Errorf("revoking an unknown token = %v, want %v", err, app.ErrNotFound)
	}
}

func TestAuthenticateRefusesJunk(t *testing.T) {
	t.Parallel()
	f := newAuthFixture(t, demoConfig{})
	for _, junk := range []string{"", "ofsess_nope", "ofs_nope", "Bearer x"} {
		if _, err := f.auth.Authenticate(t.Context(), junk); !errors.Is(err, app.ErrInvalidToken) {
			t.Errorf("Authenticate(%q) = %v, want %v", junk, err, app.ErrInvalidToken)
		}
	}
}

func TestRemovePasskey(t *testing.T) {
	t.Parallel()
	f := newAuthFixture(t, demoConfig{})
	_, ns := f.enroll(t)
	first := f.store.passkeys[0].ID()
	if err := f.auth.RemovePasskey(t.Context(), local, ns.Token, first); !errors.Is(err, app.ErrLastPasskey) {
		t.Errorf("removing the last passkey = %v, want %v", err, app.ErrLastPasskey)
	}
	if _, _, err := f.register(t, local, app.RegistrationGrant{SessionToken: ns.Token}); err != nil {
		t.Fatalf("adding a passkey with a session: %v", err)
	}
	if err := f.auth.RemovePasskey(t.Context(), local, ns.Token, first); err != nil {
		t.Errorf("removing one of two passkeys: %v", err)
	}
	if err := f.auth.RemovePasskey(t.Context(), local, ns.Token, first); !errors.Is(err, app.ErrNotFound) {
		t.Errorf("removing it again = %v, want %v", err, app.ErrNotFound)
	}
}

func TestDemoSessions(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		demo demoConfig
		want bool
	}{
		{name: "flag on, database marked", demo: demoConfig{on: true, marked: true}, want: true},
		{name: "flag on, database not marked", demo: demoConfig{on: true}},
		{name: "flag off, database marked", demo: demoConfig{marked: true}},
		{name: "flag off, database not marked", demo: demoConfig{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			f := newAuthFixture(t, tt.demo)
			methods, err := f.auth.LoginMethods(t.Context(), local)
			if err != nil || methods.DemoSessions != tt.want {
				t.Errorf("LoginMethods = %+v, %v; want demo %v", methods, err, tt.want)
			}
			ns, err := f.auth.Login(t.Context(), f.auth.Demo(), local, app.Proof{})
			if !tt.want {
				if !errors.Is(err, app.ErrForbidden) {
					t.Errorf("demo login = %v, want %v", err, app.ErrForbidden)
				}
				return
			}
			checkDemoSession(t, f, &ns, err)
		})
	}
}

// checkDemoSession checks that a demo login succeeded with a read session
// the read API accepts.
func checkDemoSession(t *testing.T, f *authFixture, ns *app.NewSession, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("demo login: %v", err)
	}
	if ns.Session.Scope != identity.SessionScopeRead || ns.Session.Authenticator != identity.AuthenticatorDemo {
		t.Errorf("demo session = %+v, want read scope from the demo authenticator", ns.Session)
	}
	if _, err := f.auth.Authenticate(t.Context(), ns.Token); err != nil {
		t.Errorf("read API with a demo session: %v", err)
	}
}
