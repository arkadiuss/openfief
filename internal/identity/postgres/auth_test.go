package postgres_test

import (
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/arkadiuss/openfief/internal/identity"
	"github.com/arkadiuss/openfief/internal/identity/app"
	"github.com/arkadiuss/openfief/internal/identity/postgres"
	"github.com/arkadiuss/openfief/internal/identity/webauthn"
	"github.com/arkadiuss/openfief/internal/identity/webauthn/virtual"
)

// localUI is the synthetic origin of the store tests.
const localUI = "http://localhost:8080"

// localOrigin parses localUI.
func localOrigin(t *testing.T) identity.Origin {
	t.Helper()
	o, err := identity.ParseOrigin(localUI)
	if err != nil {
		t.Fatalf("ParseOrigin: %v", err)
	}
	return o
}

func TestOwnerStoredOnce(t *testing.T) {
	t.Parallel()
	store := postgres.New(newPool(t))
	ctx := t.Context()
	if _, err := store.Owner(ctx); !errors.Is(err, app.ErrNoOwner) {
		t.Fatalf("Owner before any = %v, want %v", err, app.ErrNoOwner)
	}
	first, err := identity.NewOwner(bytes.Repeat([]byte{1}, identity.OwnerHandleBytes), issuedAt())
	if err != nil {
		t.Fatalf("NewOwner: %v", err)
	}
	second := first
	second.Handle = bytes.Repeat([]byte{2}, identity.OwnerHandleBytes)
	for _, candidate := range []identity.Owner{first, second} {
		got, ensureErr := store.EnsureOwner(ctx, candidate)
		if ensureErr != nil || !bytes.Equal(got.Handle, first.Handle) {
			t.Fatalf("EnsureOwner = %x, %v; want the first owner kept", got.Handle, ensureErr)
		}
	}
}

func TestPasskeyStore(t *testing.T) {
	t.Parallel()
	store := postgres.New(newPool(t))
	ctx := t.Context()
	p := identity.Passkey{
		CredentialID: []byte{9, 9, 9}, Name: "laptop", RPID: "localhost", PublicKey: []byte{1, 2, 3},
		AttestationType: "none", AttestationFormat: "none", AAGUID: make([]byte, 16), SignCount: 4,
		Transports: []string{"internal", "hybrid"}, UserVerified: true, BackupEligible: true, CreatedAt: issuedAt(),
	}
	if err := store.SavePasskey(ctx, &p); err != nil {
		t.Fatalf("SavePasskey: %v", err)
	}
	if err := store.SavePasskey(ctx, &p); err == nil {
		t.Error("SavePasskey stored a duplicate credential")
	}
	p.SignCount, p.BackupState, p.LastUsedAt = 5, true, issuedAt().Add(time.Hour)
	if err := store.UpdatePasskeyUse(ctx, &p); err != nil {
		t.Fatalf("UpdatePasskeyUse: %v", err)
	}
	keys, err := store.Passkeys(ctx)
	if err != nil {
		t.Fatalf("Passkeys: %v", err)
	}
	if diff := cmp.Diff([]identity.Passkey{p}, keys); diff != "" {
		t.Errorf("Passkeys (-want +got):\n%s", diff)
	}
	if err := store.DeletePasskey(ctx, p.CredentialID, issuedAt()); !errors.Is(err, app.ErrLastPasskey) {
		t.Errorf("DeletePasskey of the last one = %v, want %v", err, app.ErrLastPasskey)
	}
	if err := store.DeletePasskey(ctx, []byte{7}, issuedAt()); !errors.Is(err, app.ErrUnknownPasskey) {
		t.Errorf("DeletePasskey of an unknown one = %v, want %v", err, app.ErrUnknownPasskey)
	}
}

func TestCeremonyStore(t *testing.T) {
	t.Parallel()
	store := postgres.New(newPool(t))
	ctx := t.Context()
	o := localOrigin(t)
	old := identity.NewCeremony(strings.Repeat("a", 32), identity.CeremonyLogin, o, []byte(`{"c":1}`), issuedAt())
	if err := store.SaveCeremony(ctx, &old); err != nil {
		t.Fatalf("SaveCeremony: %v", err)
	}
	c := identity.NewCeremony(strings.Repeat("b", 32), identity.CeremonyRegistration, o, []byte(`{"c":2}`), issuedAt().Add(identity.CeremonyTTL))
	if err := store.SaveCeremony(ctx, &c); err != nil {
		t.Fatalf("SaveCeremony: %v", err)
	}
	got, err := store.TakeCeremony(ctx, c.ID)
	if err != nil {
		t.Fatalf("TakeCeremony: %v", err)
	}
	if diff := cmp.Diff(c, got); diff != "" {
		t.Errorf("TakeCeremony (-want +got):\n%s", diff)
	}
	for _, id := range []string{c.ID, old.ID} {
		if _, err := store.TakeCeremony(ctx, id); !errors.Is(err, app.ErrUnknownCeremony) {
			t.Errorf("TakeCeremony(%s) again or expired = %v, want %v", id[:4], err, app.ErrUnknownCeremony)
		}
	}
}

// sessionKey is the synthetic key of the stored session.
func sessionKey() []byte { return bytes.Repeat([]byte{0x42}, identity.KeyBytes) }

func TestSessionStore(t *testing.T) {
	t.Parallel()
	store := postgres.New(newPool(t))
	ctx := t.Context()
	s, err := identity.NewSession("0123456789abcdef", identity.AuthenticatorDemo, identity.SessionScopeRead, "localhost", issuedAt(), identity.DefaultSessionPolicy())
	if err != nil {
		t.Fatalf("NewSession: %v", err)
	}
	if saveErr := store.SaveSession(ctx, &s, identity.HashKey(sessionKey())); saveErr != nil {
		t.Fatalf("SaveSession: %v", saveErr)
	}
	for _, at := range []time.Duration{time.Hour, time.Minute} { // a touch never moves back
		if touchErr := store.TouchSession(ctx, s.ID, issuedAt().Add(at)); touchErr != nil {
			t.Fatalf("TouchSession: %v", touchErr)
		}
	}
	for range 2 {
		if revokeErr := store.RevokeSession(ctx, s.ID, issuedAt().Add(2*time.Hour)); revokeErr != nil {
			t.Fatalf("RevokeSession: %v", revokeErr)
		}
	}
	got, hash, err := store.Session(ctx, s.ID)
	if err != nil {
		t.Fatalf("Session: %v", err)
	}
	want := s
	want.LastSeenAt, want.RevokedAt = issuedAt().Add(time.Hour), issuedAt().Add(2*time.Hour)
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("Session (-want +got):\n%s", diff)
	}
	if !bytes.Equal(hash, identity.HashKey(sessionKey())) {
		t.Errorf("stored hash = %x", hash)
	}
}

// passkeyFor returns a synthetic passkey with credential ID id.
func passkeyFor(id byte) identity.Passkey {
	return identity.Passkey{
		CredentialID: []byte{id, id}, Name: "key", RPID: "localhost", PublicKey: []byte{1, 2, 3},
		AttestationType: "none", AttestationFormat: "none", AAGUID: make([]byte, 16), Transports: []string{"internal"},
		UserVerified: true, CreatedAt: issuedAt(),
	}
}

// savedSession stores a session with id from authenticator a, signed in by
// passkey when it is not nil.
func savedSession(t *testing.T, store *postgres.Store, id identity.SessionID, a identity.Authenticator, passkey []byte) identity.Session {
	t.Helper()
	scope, err := a.Scope()
	if err != nil {
		t.Fatalf("Scope: %v", err)
	}
	s, err := identity.NewSession(id, a, scope, "localhost", issuedAt(), identity.DefaultSessionPolicy())
	if err != nil {
		t.Fatalf("NewSession: %v", err)
	}
	if passkey != nil {
		if s, err = s.SignedInWithPasskey(passkey); err != nil {
			t.Fatalf("SignedInWithPasskey: %v", err)
		}
	}
	if err := store.SaveSession(t.Context(), &s, identity.HashKey(sessionKey())); err != nil {
		t.Fatalf("SaveSession: %v", err)
	}
	return s
}

// wantRevokedEverywhere signs out everywhere in store, where sessions
// earlier and later are left of three, earlier revoked an hour after
// issuedAt, and checks that every one is revoked and earlier keeps its time.
func wantRevokedEverywhere(t *testing.T, store *postgres.Store, earlier, later identity.SessionID) {
	t.Helper()
	n, err := store.RevokeSessions(t.Context(), issuedAt().Add(2*time.Hour))
	if err != nil || n != 2 {
		t.Errorf("RevokeSessions = %d, %v; want the 2 sessions left", n, err)
	}
	revokedAt := map[identity.SessionID]time.Time{
		earlier: issuedAt().Add(time.Hour), later: issuedAt().Add(2 * time.Hour), "c123456789abcdef": issuedAt().Add(2 * time.Hour),
	}
	for id, want := range revokedAt {
		got, _, sessionErr := store.Session(t.Context(), id)
		if sessionErr != nil || !got.RevokedAt.Equal(want) {
			t.Errorf("session %s after signing out everywhere: revoked %v, %v; want %v", id, got.RevokedAt, sessionErr, want)
		}
	}
}

// TestSessionPasskeyChecks checks what D-183 stores: the passkey that signed
// a session in and its last passkey check, which never moves back; removing
// a passkey revokes only the sessions it signed in; signing out everywhere
// revokes every session left.
func TestSessionPasskeyChecks(t *testing.T) {
	t.Parallel()
	store := postgres.New(newPool(t))
	ctx := t.Context()
	first, second := passkeyFor(1), passkeyFor(2)
	for _, p := range []*identity.Passkey{&first, &second} {
		if err := store.SavePasskey(ctx, p); err != nil {
			t.Fatalf("SavePasskey: %v", err)
		}
	}
	a := savedSession(t, store, "a123456789abcdef", identity.AuthenticatorPasskey, first.CredentialID)
	b := savedSession(t, store, "b123456789abcdef", identity.AuthenticatorPasskey, second.CredentialID)
	savedSession(t, store, "c123456789abcdef", identity.AuthenticatorDemo, nil)
	for _, at := range []time.Duration{10 * time.Minute, time.Minute} { // a check never moves back
		if err := store.PasskeyCheckSession(ctx, b.ID, issuedAt().Add(at)); err != nil {
			t.Fatalf("PasskeyCheckSession: %v", err)
		}
	}
	if err := store.PasskeyCheckSession(ctx, "fedcba9876543210", issuedAt()); !errors.Is(err, app.ErrUnknownSession) {
		t.Errorf("PasskeyCheckSession(unknown) = %v, want %v", err, app.ErrUnknownSession)
	}
	if err := store.DeletePasskey(ctx, first.CredentialID, issuedAt().Add(time.Hour)); err != nil {
		t.Fatalf("DeletePasskey: %v", err)
	}
	wantA := a
	wantA.RevokedAt = issuedAt().Add(time.Hour)
	wantStored(t, store, "the removed passkey's session", &wantA)
	wantB := b
	wantB.PasskeyCheckedAt = issuedAt().Add(10 * time.Minute)
	wantStored(t, store, "the other passkey's session", &wantB)
	wantRevokedEverywhere(t, store, a.ID, b.ID)
}

// wantStored fails unless store holds want as it is.
func wantStored(t *testing.T, store *postgres.Store, what string, want *identity.Session) {
	t.Helper()
	got, _, err := store.Session(t.Context(), want.ID)
	if err != nil {
		t.Fatalf("Session: %v", err)
	}
	if diff := cmp.Diff(*want, got); diff != "" {
		t.Errorf("%s (-want +got):\n%s", what, diff)
	}
}

// TestSessionStoreRefusesDemoPasskey checks that the table refuses a passkey
// or a passkey check on a demo session.
func TestSessionStoreRefusesDemoPasskey(t *testing.T) {
	t.Parallel()
	store := postgres.New(newPool(t))
	demo, err := identity.NewSession("d123456789abcdef", identity.AuthenticatorDemo, identity.SessionScopeRead, "localhost", issuedAt(), identity.DefaultSessionPolicy())
	if err != nil {
		t.Fatalf("NewSession: %v", err)
	}
	demo.PasskeyID, demo.PasskeyCheckedAt = []byte{1, 1}, issuedAt()
	if err := store.SaveSession(t.Context(), &demo, identity.HashKey(sessionKey())); err == nil {
		t.Error("the table stored a demo session with a passkey")
	}
}

func TestSessionStoreRefuses(t *testing.T) {
	t.Parallel()
	store := postgres.New(newPool(t))
	ctx := t.Context()
	if _, _, err := store.Session(ctx, "fedcba9876543210"); !errors.Is(err, app.ErrUnknownSession) {
		t.Errorf("Session(unknown) = %v, want %v", err, app.ErrUnknownSession)
	}
	if err := store.RevokeSession(ctx, "fedcba9876543210", issuedAt()); !errors.Is(err, app.ErrUnknownSession) {
		t.Errorf("RevokeSession(unknown) = %v, want %v", err, app.ErrUnknownSession)
	}
	s, err := identity.NewSession("1123456789abcdef", identity.AuthenticatorPasskey, identity.SessionScopeOwner, "localhost", issuedAt(), identity.DefaultSessionPolicy())
	if err != nil {
		t.Fatalf("NewSession: %v", err)
	}
	s.Authenticator = identity.AuthenticatorDemo
	if err := store.SaveSession(ctx, &s, identity.HashKey(sessionKey())); err == nil {
		t.Error("the table stored an owner-scoped demo session")
	}
}

func TestEnrollmentCodeUsedOnce(t *testing.T) {
	t.Parallel()
	store := postgres.New(newPool(t))
	ctx := t.Context()
	key := bytes.Repeat([]byte{3}, identity.KeyBytes)
	c, err := identity.NewEnrollmentCode("0123456789abcdef", localOrigin(t), issuedAt(), 15*time.Minute)
	if err != nil {
		t.Fatalf("NewEnrollmentCode: %v", err)
	}
	if saveErr := store.SaveEnrollmentCode(ctx, &c, identity.HashKey(key)); saveErr != nil {
		t.Fatalf("SaveEnrollmentCode: %v", saveErr)
	}
	got, hash, err := store.EnrollmentCode(ctx, c.ID)
	if err != nil || !bytes.Equal(hash, identity.HashKey(key)) {
		t.Fatalf("EnrollmentCode = %x, %v", hash, err)
	}
	if diff := cmp.Diff(c, got); diff != "" {
		t.Errorf("EnrollmentCode (-want +got):\n%s", diff)
	}
	if ok := useConcurrently(t, store, c.ID, 8); ok != 1 {
		t.Errorf("%d concurrent uses succeeded, want 1", ok)
	}
	if _, _, err := store.EnrollmentCode(ctx, "fedcba9876543210"); !errors.Is(err, app.ErrUnknownEnrollmentCode) {
		t.Errorf("EnrollmentCode(unknown) = %v, want %v", err, app.ErrUnknownEnrollmentCode)
	}
}

// useConcurrently uses the code id n times at once and returns how many
// uses succeeded; every failure must be identity.ErrEnrollmentCodeUsed.
func useConcurrently(t *testing.T, store *postgres.Store, id identity.EnrollmentCodeID, n int) int {
	t.Helper()
	var wg sync.WaitGroup
	errs := make([]error, n)
	for i := range errs {
		wg.Go(func() { errs[i] = store.UseEnrollmentCode(t.Context(), id, issuedAt().Add(time.Minute)) })
	}
	wg.Wait()
	ok := 0
	for _, err := range errs {
		if err == nil {
			ok++
		} else if !errors.Is(err, identity.ErrEnrollmentCodeUsed) {
			t.Errorf("UseEnrollmentCode = %v", err)
		}
	}
	return ok
}

func TestAuthOverPostgres(t *testing.T) {
	t.Parallel()
	pool := newPool(t)
	ctx := t.Context()
	store := postgres.New(pool)
	now := issuedAt()
	clock := func() time.Time { return now }
	origins, err := identity.NewOrigins([]string{localUI})
	if err != nil {
		t.Fatalf("NewOrigins: %v", err)
	}
	auth, err := app.NewAuth(store, webauthn.New("OpenFief"), app.NewTokens(store, clock, nil), nil,
		app.AuthConfig{Origins: origins, Sessions: identity.DefaultSessionPolicy()}, clock, nil)
	if err != nil {
		t.Fatalf("NewAuth: %v", err)
	}
	link, err := app.IssueEnrollmentLink(ctx, store, nil, now, origins, localOrigin(t), 0)
	if err != nil {
		t.Fatalf("IssueEnrollmentLink: %v", err)
	}
	_, code, _ := strings.Cut(link.URL, "#code=")
	enrolled := enrollOver(t, auth, code)
	now = now.Add(time.Hour)
	session := loginOver(t, auth, enrolled.dev)
	_, secret, err := auth.CreateScriptToken(ctx, localUI, session, "nightly", 0)
	if err != nil {
		t.Fatalf("CreateScriptToken: %v", err)
	}
	for _, bearer := range []string{session, secret} {
		if _, err := auth.Authenticate(ctx, bearer); err != nil {
			t.Errorf("Authenticate: %v", err)
		}
	}
	checkNoAuthSecretStored(t, pool, []string{code, session, enrolled.session})
}

// enrolled is a passkey's authenticator and the session that came with it.
type enrolled struct {
	dev     *virtual.Authenticator
	session string
}

// enrollOver creates a passkey with code through auth.
func enrollOver(t *testing.T, auth *app.Auth, code string) enrolled {
	t.Helper()
	dev, err := virtual.New(nil)
	if err != nil {
		t.Fatalf("virtual.New: %v", err)
	}
	st, err := auth.BeginRegistration(t.Context(), localUI, app.RegistrationGrant{EnrollmentCode: code})
	if err != nil {
		t.Fatalf("BeginRegistration: %v", err)
	}
	cred, err := dev.Create(st.Options, localUI)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	r, err := auth.FinishRegistration(t.Context(), localUI, st.CeremonyID, cred, "laptop")
	if err != nil || r.Session == nil {
		t.Fatalf("FinishRegistration = %+v, %v", r, err)
	}
	return enrolled{dev: dev, session: r.Session.Token}
}

// loginOver logs in with dev through auth and returns the session token.
func loginOver(t *testing.T, auth *app.Auth, dev *virtual.Authenticator) string {
	t.Helper()
	st, err := auth.BeginLogin(t.Context(), localUI)
	if err != nil {
		t.Fatalf("BeginLogin: %v", err)
	}
	assertion, err := dev.Get(st.Options, localUI)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	ns, err := auth.Login(t.Context(), auth.Passkey(), localUI, app.Proof{CeremonyID: st.CeremonyID, Credential: assertion})
	if err != nil {
		t.Fatalf("Login: %v", err)
	}
	return ns.Token
}

// checkNoAuthSecretStored fails if the sessions or enrollment codes table
// holds the key of one of the secrets, as text or as hex.
func checkNoAuthSecretStored(t *testing.T, pool *pgxpool.Pool, secrets []string) {
	t.Helper()
	var dump string
	err := pool.QueryRow(t.Context(), `SELECT
		(SELECT string_agg(s::text, ' ') FROM identity.sessions s) || ' ' ||
		(SELECT string_agg(c::text, ' ') FROM identity.enrollment_codes c)`).Scan(&dump)
	if err != nil {
		t.Fatalf("dumping: %v", err)
	}
	for _, secret := range secrets {
		// The prefix (ofsess, ofenroll) ends at the first "_"; the base64url
		// key itself may contain "_".
		_, keyText, _ := strings.Cut(secret, "_")
		key, err := base64.RawURLEncoding.DecodeString(keyText)
		if err != nil {
			t.Fatalf("decoding a key: %v", err)
		}
		if strings.Contains(dump, keyText) || strings.Contains(dump, hex.EncodeToString(key)) {
			t.Error("an auth table holds a secret's key")
		}
	}
}
