package identity_test

import (
	"bytes"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"github.com/arkadiuss/openfief/internal/identity"
)

// mustOrigin parses a synthetic origin.
func mustOrigin(t *testing.T, text string) identity.Origin {
	t.Helper()
	o, err := identity.ParseOrigin(text)
	if err != nil {
		t.Fatalf("ParseOrigin(%q): %v", text, err)
	}
	return o
}

func TestParseOrigin(t *testing.T) {
	t.Parallel()
	tests := []struct {
		text     string
		wantText string
		wantRPID string
		wantErr  error
	}{
		{text: "https://dev.openfief.dev", wantText: "https://dev.openfief.dev", wantRPID: "dev.openfief.dev"},
		{text: "https://dev.example.com:443", wantText: "https://dev.example.com", wantRPID: "dev.example.com"},
		{text: "https://dev.example.com:8443", wantText: "https://dev.example.com:8443", wantRPID: "dev.example.com"},
		{text: "http://localhost:8080", wantText: "http://localhost:8080", wantRPID: "localhost"},
		{text: "http://localhost", wantText: "http://localhost", wantRPID: "localhost"},
		{text: "http://dev.example.com", wantErr: identity.ErrInvalidOrigin},
		{text: "http://127.0.0.1:8080", wantErr: identity.ErrInvalidOrigin},
		{text: "https://192.0.2.1", wantErr: identity.ErrInvalidOrigin},
		{text: "https://[2001:db8::1]", wantErr: identity.ErrInvalidOrigin},
		{text: "https://Dev.example.com", wantErr: identity.ErrInvalidOrigin},
		{text: "https://dev.example.com/", wantErr: identity.ErrInvalidOrigin},
		{text: "https://dev.example.com/enroll", wantErr: identity.ErrInvalidOrigin},
		{text: "https://user@dev.example.com", wantErr: identity.ErrInvalidOrigin},
		{text: "https://dev.example.com:0", wantErr: identity.ErrInvalidOrigin},
		{text: "https://dev.example.com:65536", wantErr: identity.ErrInvalidOrigin},
		{text: "https://dev.example.com:080", wantErr: identity.ErrInvalidOrigin},
		{text: "https://-dev.example.com", wantErr: identity.ErrInvalidOrigin},
		{text: "https://dev..example.com", wantErr: identity.ErrInvalidOrigin},
		{text: "ftp://dev.example.com", wantErr: identity.ErrInvalidOrigin},
		{text: "dev.example.com", wantErr: identity.ErrInvalidOrigin},
		{text: "null", wantErr: identity.ErrInvalidOrigin},
		{text: "", wantErr: identity.ErrInvalidOrigin},
	}
	for _, tt := range tests {
		t.Run(tt.text, func(t *testing.T) {
			t.Parallel()
			o, err := identity.ParseOrigin(tt.text)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("ParseOrigin error = %v, want %v", err, tt.wantErr)
			}
			if err == nil && (o.String() != tt.wantText || o.RPID() != tt.wantRPID) {
				t.Errorf("ParseOrigin = %s (RP ID %s), want %s (%s)", o, o.RPID(), tt.wantText, tt.wantRPID)
			}
		})
	}
}

func TestOrigins(t *testing.T) {
	t.Parallel()
	if _, err := identity.NewOrigins(nil); !errors.Is(err, identity.ErrInvalidOrigin) {
		t.Errorf("NewOrigins(nil) = %v, want %v", err, identity.ErrInvalidOrigin)
	}
	if _, err := identity.NewOrigins([]string{"http://dev.example.com"}); !errors.Is(err, identity.ErrInvalidOrigin) {
		t.Errorf("NewOrigins with a bad origin = %v, want %v", err, identity.ErrInvalidOrigin)
	}
	list, err := identity.NewOrigins([]string{"https://dev.example.com", "http://localhost:8080", "https://dev.example.com:443"})
	if err != nil {
		t.Fatalf("NewOrigins: %v", err)
	}
	if got := len(list.List()); got != 2 {
		t.Errorf("List has %d origins, want 2 (duplicates dropped)", got)
	}
	tests := []struct {
		text    string
		wantErr error
	}{
		{text: "https://dev.example.com"},
		{text: "http://localhost:8080"},
		{text: "http://localhost:8081", wantErr: identity.ErrOriginNotAllowed},
		{text: "https://other.example.com", wantErr: identity.ErrOriginNotAllowed},
		{text: "garbage", wantErr: identity.ErrInvalidOrigin},
	}
	for _, tt := range tests {
		if _, err := list.Allowed(tt.text); !errors.Is(err, tt.wantErr) {
			t.Errorf("Allowed(%q) = %v, want %v", tt.text, err, tt.wantErr)
		}
	}
}

// sessionAt is when the test sessions are issued.
func sessionAt() time.Time { return time.Date(2026, 10, 28, 8, 0, 0, 0, time.UTC) }

func TestNewSession(t *testing.T) {
	t.Parallel()
	policy := identity.SessionPolicy{Idle: time.Hour, Max: 24 * time.Hour}
	tests := []struct {
		name    string
		id      identity.SessionID
		auth    identity.Authenticator
		scope   identity.SessionScope
		rpID    string
		policy  identity.SessionPolicy
		wantErr error
	}{
		{name: "passkey, owner scope", id: "0123456789abcdef", auth: identity.AuthenticatorPasskey, scope: identity.SessionScopeOwner, rpID: "localhost", policy: policy},
		{name: "passkey, read scope", id: "0123456789abcdef", auth: identity.AuthenticatorPasskey, scope: identity.SessionScopeRead, rpID: "localhost", policy: policy},
		{name: "demo, read scope", id: "0123456789abcdef", auth: identity.AuthenticatorDemo, scope: identity.SessionScopeRead, rpID: "localhost", policy: policy},
		{name: "demo can't give owner scope", id: "0123456789abcdef", auth: identity.AuthenticatorDemo, scope: identity.SessionScopeOwner, rpID: "localhost", policy: policy, wantErr: identity.ErrSessionScope},
		{name: "no authenticator", id: "0123456789abcdef", scope: identity.SessionScopeRead, rpID: "localhost", policy: policy, wantErr: identity.ErrSessionScope},
		{name: "bad ID", id: "XYZ", auth: identity.AuthenticatorPasskey, scope: identity.SessionScopeOwner, rpID: "localhost", policy: policy, wantErr: identity.ErrMalformedSession},
		{name: "bad host", id: "0123456789abcdef", auth: identity.AuthenticatorPasskey, scope: identity.SessionScopeOwner, rpID: "192.0.2.1", policy: policy, wantErr: identity.ErrMalformedSession},
		{name: "idle longer than max", id: "0123456789abcdef", auth: identity.AuthenticatorPasskey, scope: identity.SessionScopeOwner, rpID: "localhost", policy: identity.SessionPolicy{Idle: 2 * time.Hour, Max: time.Hour}, wantErr: identity.ErrInvalidSessionPolicy},
		{name: "zero idle", id: "0123456789abcdef", auth: identity.AuthenticatorPasskey, scope: identity.SessionScopeOwner, rpID: "localhost", policy: identity.SessionPolicy{Max: time.Hour}, wantErr: identity.ErrInvalidSessionPolicy},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			s, err := identity.NewSession(tt.id, tt.auth, tt.scope, tt.rpID, sessionAt(), tt.policy)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("NewSession error = %v, want %v", err, tt.wantErr)
			}
			if err != nil {
				return
			}
			want := identity.Session{
				ID: tt.id, Scope: tt.scope, Authenticator: tt.auth, RPID: tt.rpID,
				CreatedAt: sessionAt(), LastSeenAt: sessionAt(), ExpiresAt: sessionAt().Add(24 * time.Hour), IdleTimeout: time.Hour,
			}
			if diff := cmp.Diff(want, s); diff != "" {
				t.Errorf("NewSession (-want +got):\n%s", diff)
			}
		})
	}
	if err := identity.DefaultSessionPolicy().Validate(); err != nil {
		t.Errorf("the default policy is invalid: %v", err)
	}
}

func TestSessionCheck(t *testing.T) {
	t.Parallel()
	base, err := identity.NewSession("0123456789abcdef", identity.AuthenticatorDemo, identity.SessionScopeRead, "localhost", sessionAt(),
		identity.SessionPolicy{Idle: time.Hour, Max: 3 * time.Hour})
	if err != nil {
		t.Fatalf("NewSession: %v", err)
	}
	tests := []struct {
		name     string
		lastSeen time.Duration
		revoked  bool
		at       time.Duration
		rpID     string
		want     identity.SessionScope
		wantErr  error
		wantIdle time.Duration
	}{
		{name: "fresh", at: time.Minute, rpID: "localhost", want: identity.SessionScopeRead, wantIdle: time.Hour},
		{name: "no host check", at: time.Minute, want: identity.SessionScopeRead, wantIdle: time.Hour},
		{name: "just before idle expiry", at: time.Hour - time.Nanosecond, want: identity.SessionScopeRead, wantIdle: time.Hour},
		{name: "at idle expiry", at: time.Hour, want: identity.SessionScopeRead, wantErr: identity.ErrSessionIdle, wantIdle: time.Hour},
		{name: "kept alive by use", lastSeen: 2 * time.Hour, at: 2*time.Hour + 30*time.Minute, want: identity.SessionScopeRead, wantIdle: 3 * time.Hour},
		{name: "absolute expiry wins", lastSeen: 2*time.Hour + 50*time.Minute, at: 3 * time.Hour, want: identity.SessionScopeRead, wantErr: identity.ErrSessionExpired, wantIdle: 3 * time.Hour},
		{name: "revoked", revoked: true, at: time.Minute, want: identity.SessionScopeRead, wantErr: identity.ErrSessionRevoked, wantIdle: time.Hour},
		{name: "another host", at: time.Minute, rpID: "dev.example.com", want: identity.SessionScopeRead, wantErr: identity.ErrSessionWrongHost, wantIdle: time.Hour},
		{name: "read scope can't act as owner", at: time.Minute, want: identity.SessionScopeOwner, wantErr: identity.ErrSessionScope, wantIdle: time.Hour},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			s := base
			s.LastSeenAt = sessionAt().Add(tt.lastSeen)
			if tt.revoked {
				s.RevokedAt = sessionAt()
			}
			if err := s.Check(sessionAt().Add(tt.at), tt.rpID, tt.want); !errors.Is(err, tt.wantErr) {
				t.Errorf("Check = %v, want %v", err, tt.wantErr)
			}
			if got := s.IdleExpiresAt(); !got.Equal(sessionAt().Add(tt.wantIdle)) {
				t.Errorf("IdleExpiresAt = %s, want %s", got, sessionAt().Add(tt.wantIdle))
			}
		})
	}
}

func TestSessionTouched(t *testing.T) {
	t.Parallel()
	s := identity.Session{LastSeenAt: sessionAt()}
	if got, store := s.Touched(sessionAt().Add(59 * time.Second)); store || !got.LastSeenAt.Equal(sessionAt()) {
		t.Errorf("Touched within a minute = %s, %v; want no write", got.LastSeenAt, store)
	}
	if got, store := s.Touched(sessionAt().Add(time.Minute)); !store || !got.LastSeenAt.Equal(sessionAt().Add(time.Minute)) {
		t.Errorf("Touched after a minute = %s, %v; want a write", got.LastSeenAt, store)
	}
}

func TestSessionScopes(t *testing.T) {
	t.Parallel()
	for _, sc := range []identity.SessionScope{identity.SessionScopeRead, identity.SessionScopeOwner} {
		got, err := identity.ParseSessionScope(sc.String())
		if err != nil || got != sc {
			t.Errorf("ParseSessionScope(%q) = %v, %v", sc, got, err)
		}
	}
	if _, err := identity.ParseSessionScope("admin"); !errors.Is(err, identity.ErrSessionScope) {
		t.Errorf("ParseSessionScope(admin) = %v", err)
	}
	if identity.SessionScope(0).String() != "SessionScope(0)" {
		t.Error("the zero scope prints wrongly")
	}
	if identity.SessionScopeRead.Includes(identity.SessionScopeOwner) || !identity.SessionScopeOwner.Includes(identity.SessionScopeRead) ||
		identity.SessionScopeOwner.Includes(0) {
		t.Error("Includes: owner includes read, read does not include owner, nothing includes the zero scope")
	}
}

func TestAuthenticators(t *testing.T) {
	t.Parallel()
	for _, a := range []identity.Authenticator{identity.AuthenticatorPasskey, identity.AuthenticatorDemo} {
		got, err := identity.ParseAuthenticator(a.String())
		if err != nil || got != a {
			t.Errorf("ParseAuthenticator(%q) = %v, %v", a, got, err)
		}
	}
	if _, err := identity.ParseAuthenticator("password"); err == nil {
		t.Error("ParseAuthenticator(password) succeeded")
	}
	if identity.Authenticator(0).String() != "Authenticator(0)" {
		t.Error("the zero authenticator prints wrongly")
	}
	if _, err := identity.Authenticator(0).Scope(); !errors.Is(err, identity.ErrSessionScope) {
		t.Errorf("the zero authenticator's scope = %v", err)
	}
}

func TestSessionToken(t *testing.T) {
	t.Parallel()
	key := bytes.Repeat([]byte{0xfb}, identity.KeyBytes)
	token := identity.FormatSessionToken("0123456789abcdef", key)
	if !strings.HasPrefix(token, "ofsess_0123456789abcdef_") {
		t.Fatalf("FormatSessionToken = %q", token)
	}
	id, got, err := identity.ParseSessionToken(token)
	if err != nil || id != "0123456789abcdef" || !bytes.Equal(got, key) {
		t.Errorf("ParseSessionToken = %q, %x, %v", id, got, err)
	}
	scriptSecret := identity.FormatSecret("0123456789abcdef", key)
	for _, bad := range []string{"", scriptSecret, "ofsess_0123_" + strings.Repeat("A", 43), "ofsess_0123456789abcdef_AAAA"} {
		if _, _, err := identity.ParseSessionToken(bad); !errors.Is(err, identity.ErrMalformedSession) || (bad != "" && strings.Contains(err.Error(), bad)) {
			t.Errorf("ParseSessionToken(%q) = %v, want %v without the token", bad, err, identity.ErrMalformedSession)
		}
	}
}

func TestPasskeyUsed(t *testing.T) {
	t.Parallel()
	used := sessionAt().Add(time.Hour)
	tests := []struct {
		name    string
		stored  uint32
		got     uint32
		wantErr error
	}{
		{name: "counter went up", stored: 5, got: 6},
		{name: "both zero (synced passkey)", stored: 0, got: 0},
		{name: "first count", stored: 0, got: 1},
		{name: "same counter", stored: 5, got: 5, wantErr: identity.ErrSignCountRegressed},
		{name: "counter went down", stored: 5, got: 4, wantErr: identity.ErrSignCountRegressed},
		{name: "counter reset to zero", stored: 5, got: 0, wantErr: identity.ErrSignCountRegressed},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			p := identity.Passkey{CredentialID: []byte{1}, SignCount: tt.stored}
			q, err := p.Used(tt.got, true, used)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("Used = %v, want %v", err, tt.wantErr)
			}
			if err == nil && (q.SignCount != tt.got || !q.BackupState || !q.LastUsedAt.Equal(used) || p.SignCount != tt.stored) {
				t.Errorf("Used = %+v (original %+v)", q, p)
			}
		})
	}
}

func TestPasskeyChecks(t *testing.T) {
	t.Parallel()
	good := identity.Passkey{CredentialID: []byte{1, 2}, PublicKey: []byte{3}, RPID: "localhost", Name: "laptop"}
	if err := good.CheckHost("localhost"); err != nil {
		t.Errorf("CheckHost(localhost) = %v", err)
	}
	if err := good.CheckHost("dev.example.com"); !errors.Is(err, identity.ErrPasskeyWrongHost) {
		t.Errorf("CheckHost(another host) = %v, want %v", err, identity.ErrPasskeyWrongHost)
	}
	id, err := identity.ParsePasskeyID(good.ID())
	if err != nil || !bytes.Equal(id, good.CredentialID) {
		t.Errorf("ParsePasskeyID(ID()) = %x, %v", id, err)
	}
	if _, err := identity.ParsePasskeyID("not base64!"); !errors.Is(err, identity.ErrInvalidPasskey) {
		t.Errorf("ParsePasskeyID(junk) = %v", err)
	}
	other := good
	other.CredentialID = []byte{1, 2}
	if !good.SameCredential(&other) {
		t.Error("SameCredential is false for equal IDs")
	}
}

func TestPasskeyValidate(t *testing.T) {
	t.Parallel()
	good := identity.Passkey{CredentialID: []byte{1, 2}, PublicKey: []byte{3}, RPID: "localhost", Name: "laptop"}
	if err := good.Validate(); err != nil {
		t.Errorf("Validate = %v", err)
	}
	bad := []func(p *identity.Passkey){
		func(p *identity.Passkey) { p.CredentialID = nil },
		func(p *identity.Passkey) { p.PublicKey = nil },
		func(p *identity.Passkey) { p.RPID = "192.0.2.1" },
		func(p *identity.Passkey) { p.Name = "" },
		func(p *identity.Passkey) { p.Name = strings.Repeat("x", 65) },
		func(p *identity.Passkey) { p.Name = "a\nb" },
	}
	for i, edit := range bad {
		p := good
		edit(&p)
		if err := p.Validate(); !errors.Is(err, identity.ErrInvalidPasskey) {
			t.Errorf("bad passkey %d: Validate = %v, want %v", i, err, identity.ErrInvalidPasskey)
		}
	}
	if _, err := identity.NewOwner(make([]byte, 31), sessionAt()); !errors.Is(err, identity.ErrInvalidOwner) {
		t.Errorf("NewOwner with a short handle = %v", err)
	}
}

func TestEnrollmentCode(t *testing.T) {
	t.Parallel()
	local := mustOrigin(t, "http://localhost:8080")
	devOrigin := mustOrigin(t, "https://dev.example.com")
	c, err := identity.NewEnrollmentCode("0123456789abcdef", local, sessionAt(), 15*time.Minute)
	if err != nil {
		t.Fatalf("NewEnrollmentCode: %v", err)
	}
	tests := []struct {
		name    string
		used    bool
		at      time.Duration
		origin  identity.Origin
		wantErr error
	}{
		{name: "fresh", at: time.Minute, origin: local},
		{name: "at expiry", at: 15 * time.Minute, origin: local, wantErr: identity.ErrEnrollmentCodeExpired},
		{name: "used", used: true, at: time.Minute, origin: local, wantErr: identity.ErrEnrollmentCodeUsed},
		{name: "another origin", at: time.Minute, origin: devOrigin, wantErr: identity.ErrEnrollmentCodeWrongOrigin},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			code := c
			if tt.used {
				code.UsedAt = sessionAt()
			}
			if err := code.Check(sessionAt().Add(tt.at), tt.origin); !errors.Is(err, tt.wantErr) {
				t.Errorf("Check = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestNewEnrollmentCodeRefused(t *testing.T) {
	t.Parallel()
	local := mustOrigin(t, "http://localhost:8080")
	for _, ttl := range []time.Duration{0, -time.Minute, identity.MaxEnrollmentTTL + time.Second} {
		if _, err := identity.NewEnrollmentCode("0123456789abcdef", local, sessionAt(), ttl); !errors.Is(err, identity.ErrInvalidEnrollmentTTL) {
			t.Errorf("NewEnrollmentCode ttl %s = %v", ttl, err)
		}
	}
	if _, err := identity.NewEnrollmentCode("x", local, sessionAt(), time.Minute); !errors.Is(err, identity.ErrMalformedEnrollmentCode) {
		t.Errorf("NewEnrollmentCode with a bad ID = %v", err)
	}
	if _, err := identity.NewEnrollmentCode("0123456789abcdef", identity.Origin{}, sessionAt(), time.Minute); !errors.Is(err, identity.ErrInvalidOrigin) {
		t.Errorf("NewEnrollmentCode without an origin = %v", err)
	}
}

func TestEnrollmentCodeText(t *testing.T) {
	t.Parallel()
	local := mustOrigin(t, "http://localhost:8080")
	key := bytes.Repeat([]byte{1}, identity.KeyBytes)
	text := identity.FormatEnrollmentCode("0123456789abcdef", key)
	url := identity.EnrollmentURL(local, text)
	if want := "http://localhost:8080/enroll#code=ofenroll_0123456789abcdef_"; !strings.HasPrefix(url, want) {
		t.Errorf("EnrollmentURL = %q, want prefix %q", url, want)
	}
	id, got, err := identity.ParseEnrollmentCode(text)
	if err != nil || id != "0123456789abcdef" || !bytes.Equal(got, key) {
		t.Errorf("ParseEnrollmentCode = %q, %x, %v", id, got, err)
	}
	if _, _, err := identity.ParseEnrollmentCode(identity.FormatSessionToken("0123456789abcdef", key)); !errors.Is(err, identity.ErrMalformedEnrollmentCode) {
		t.Errorf("a session token parsed as an enrollment code: %v", err)
	}
}

func TestCeremony(t *testing.T) {
	t.Parallel()
	local := mustOrigin(t, "http://localhost:8080")
	devOrigin := mustOrigin(t, "https://dev.example.com")
	c := identity.NewCeremony("00112233445566778899aabbccddeeff", identity.CeremonyLogin, local, []byte("{}"), sessionAt())
	if !c.ExpiresAt.Equal(sessionAt().Add(identity.CeremonyTTL)) {
		t.Errorf("ExpiresAt = %s", c.ExpiresAt)
	}
	tests := []struct {
		name    string
		at      time.Duration
		kind    identity.CeremonyKind
		origin  identity.Origin
		wantErr error
	}{
		{name: "in time", at: time.Minute, kind: identity.CeremonyLogin, origin: local},
		{name: "expired", at: identity.CeremonyTTL, kind: identity.CeremonyLogin, origin: local, wantErr: identity.ErrCeremonyExpired},
		{name: "other kind", at: time.Minute, kind: identity.CeremonyRegistration, origin: local, wantErr: identity.ErrCeremonyMismatch},
		{name: "other origin", at: time.Minute, kind: identity.CeremonyLogin, origin: devOrigin, wantErr: identity.ErrCeremonyMismatch},
	}
	for _, tt := range tests {
		if err := c.Check(sessionAt().Add(tt.at), tt.kind, tt.origin); !errors.Is(err, tt.wantErr) {
			t.Errorf("%s: Check = %v, want %v", tt.name, err, tt.wantErr)
		}
	}
	for _, k := range []identity.CeremonyKind{identity.CeremonyRegistration, identity.CeremonyLogin} {
		got, err := identity.ParseCeremonyKind(k.String())
		if err != nil || got != k {
			t.Errorf("ParseCeremonyKind(%q) = %v, %v", k, got, err)
		}
	}
	if _, err := identity.ParseCeremonyKind("dance"); !errors.Is(err, identity.ErrCeremonyMismatch) {
		t.Errorf("ParseCeremonyKind(dance) = %v", err)
	}
	if identity.CeremonyKind(0).String() != "CeremonyKind(0)" {
		t.Error("the zero kind prints wrongly")
	}
}
