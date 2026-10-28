package app_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"github.com/arkadiuss/openfief/internal/identity"
	"github.com/arkadiuss/openfief/internal/identity/app"
)

var errStore = errors.New("store unavailable")

// stored is one token in memStore.
type stored struct {
	token identity.ScriptToken
	hash  []byte
}

// memStore is an in-memory app.Store.
type memStore struct {
	mu     sync.Mutex
	tokens map[identity.ScriptTokenID]stored
	fail   error
}

func newMemStore() *memStore { return &memStore{tokens: map[identity.ScriptTokenID]stored{}} }

func (m *memStore) SaveScriptToken(_ context.Context, t *identity.ScriptToken, hash []byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.fail != nil {
		return m.fail
	}
	m.tokens[t.ID] = stored{token: *t, hash: slices.Clone(hash)}
	return nil
}

func (m *memStore) ScriptToken(_ context.Context, id identity.ScriptTokenID) (identity.ScriptToken, []byte, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.fail != nil {
		return identity.ScriptToken{}, nil, m.fail
	}
	s, ok := m.tokens[id]
	if !ok {
		return identity.ScriptToken{}, nil, fmt.Errorf("%w: %s", app.ErrUnknownToken, id)
	}
	return s.token, s.hash, nil
}

func (m *memStore) RevokeScriptToken(_ context.Context, id identity.ScriptTokenID, at time.Time) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	s, ok := m.tokens[id]
	if !ok {
		return fmt.Errorf("%w: %s", app.ErrUnknownToken, id)
	}
	if s.token.RevokedAt.IsZero() {
		s.token.RevokedAt = at
	}
	m.tokens[id] = s
	return nil
}

func (m *memStore) ScriptTokens(context.Context) ([]identity.ScriptToken, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]identity.ScriptToken, 0, len(m.tokens))
	for id := range m.tokens {
		out = append(out, m.tokens[id].token)
	}
	slices.SortFunc(out, func(a, b identity.ScriptToken) int { return strings.Compare(string(a.ID), string(b.ID)) })
	return out, nil
}

// clock is a settable clock.
type clock struct {
	mu sync.Mutex
	t  time.Time
}

func (c *clock) now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.t
}

func (c *clock) set(t time.Time) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.t = t
}

// issuedAt is when the tests issue tokens; Issue drops the nanoseconds.
func issuedAt() time.Time { return time.Date(2026, 10, 28, 9, 30, 15, 999, time.UTC) }

// fixture issues one 24-hour token from fixed random bytes.
type fixture struct {
	store  *memStore
	clock  *clock
	tokens *app.Tokens
	token  identity.ScriptToken
	secret string
}

func newFixture(t *testing.T) *fixture {
	t.Helper()
	f := &fixture{store: newMemStore(), clock: &clock{t: issuedAt()}}
	random := bytes.NewReader(append(bytes.Repeat([]byte{0xab}, identity.IDBytes), bytes.Repeat([]byte{7}, identity.KeyBytes)...))
	f.tokens = app.NewTokens(f.store, f.clock.now, random)
	tok, secret, err := f.tokens.Issue(t.Context(), "demo", 24*time.Hour)
	if err != nil {
		t.Fatalf("Issue: %v", err)
	}
	f.token, f.secret = tok, secret
	return f
}

func TestIssue(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	created := issuedAt().Truncate(time.Second)
	want := identity.ScriptToken{
		ID: "abababababababab", Name: "demo", Scope: identity.ScopeRead, CreatedAt: created, ExpiresAt: created.Add(24 * time.Hour),
	}
	if diff := cmp.Diff(want, f.token); diff != "" {
		t.Errorf("Issue token (-want +got):\n%s", diff)
	}
	key := bytes.Repeat([]byte{7}, identity.KeyBytes)
	if wantSecret := identity.FormatSecret(want.ID, key); f.secret != wantSecret {
		t.Errorf("Issue secret = %q, want %q", f.secret, wantSecret)
	}
	s := f.store.tokens[want.ID]
	if !bytes.Equal(s.hash, identity.HashKey(key)) {
		t.Errorf("stored hash = %x, want the key's SHA-256", s.hash)
	}
	if bytes.Contains(s.hash, key) || strings.Contains(fmt.Sprint(s), f.secret) {
		t.Error("the store holds the key or the secret")
	}
	listed, err := f.tokens.List(t.Context())
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if diff := cmp.Diff([]identity.ScriptToken{want}, listed); diff != "" {
		t.Errorf("List (-want +got):\n%s", diff)
	}
}

func TestIssueFails(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		tokName string
		ttl     time.Duration
		random  []byte
		fail    error
		wantErr error // nil: any error
	}{
		{name: "bad name", tokName: "my token", ttl: time.Hour, wantErr: identity.ErrInvalidName},
		{name: "bad lifetime", tokName: "x", ttl: 0, wantErr: identity.ErrInvalidTTL},
		{name: "no random bytes", tokName: "x", ttl: time.Hour, random: []byte{}},
		{name: "too few random bytes for the key", tokName: "x", ttl: time.Hour, random: make([]byte, identity.IDBytes+1)},
		{name: "store fails", tokName: "x", ttl: time.Hour, fail: errStore, wantErr: errStore},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			store := newMemStore()
			store.fail = tt.fail
			random := tt.random
			if random == nil {
				random = make([]byte, identity.IDBytes+identity.KeyBytes)
			}
			tokens := app.NewTokens(store, issuedAt, bytes.NewReader(random))
			_, secret, err := tokens.Issue(t.Context(), tt.tokName, tt.ttl)
			if err == nil || (tt.wantErr != nil && !errors.Is(err, tt.wantErr)) {
				t.Fatalf("Issue error = %v, want %v", err, tt.wantErr)
			}
			if secret != "" || len(store.tokens) != 0 {
				t.Errorf("a failed Issue returned a secret or stored a token")
			}
		})
	}
}

func TestVerify(t *testing.T) {
	t.Parallel()
	expires := issuedAt().Truncate(time.Second).Add(24 * time.Hour)
	tests := []struct {
		name    string
		secret  func(f *fixture) string
		prepare func(t *testing.T, f *fixture)
		want    identity.Scope
		wantErr []error
	}{
		{name: "issued token", secret: (*fixture).issued, want: identity.ScopeRead},
		{name: "just before expiry", secret: (*fixture).issued, want: identity.ScopeRead, prepare: at(expires.Add(-time.Second))},
		{name: "at expiry", secret: (*fixture).issued, want: identity.ScopeRead, prepare: at(expires), wantErr: []error{app.ErrInvalidToken, identity.ErrTokenExpired}},
		{name: "revoked", secret: (*fixture).issued, want: identity.ScopeRead, prepare: revoke, wantErr: []error{app.ErrInvalidToken, identity.ErrTokenRevoked}},
		{name: "scope not granted", secret: (*fixture).issued, want: 2, wantErr: []error{app.ErrInvalidToken, identity.ErrScopeNotGranted}},
		{name: "malformed", secret: func(*fixture) string { return "not-a-token" }, want: identity.ScopeRead, wantErr: []error{app.ErrInvalidToken, identity.ErrMalformedSecret}},
		{name: "unknown ID", secret: (*fixture).otherID, want: identity.ScopeRead, wantErr: []error{app.ErrInvalidToken}},
		{name: "wrong key", secret: (*fixture).otherKey, want: identity.ScopeRead, wantErr: []error{app.ErrInvalidToken}},
		{name: "store fails", secret: (*fixture).issued, want: identity.ScopeRead, prepare: failStore, wantErr: []error{errStore}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			f := newFixture(t)
			if tt.prepare != nil {
				tt.prepare(t, f)
			}
			secret := tt.secret(f)
			checkVerified(t, f.tokens.Verify(t.Context(), secret, tt.want), secret, tt.wantErr)
		})
	}
}

// checkVerified checks Verify's error: nil when wantErr is empty, else
// wrapping each of wantErr, never containing the secret, and never calling a
// store failure an invalid token.
func checkVerified(t *testing.T, err error, secret string, wantErr []error) {
	t.Helper()
	if len(wantErr) == 0 && err != nil {
		t.Fatalf("Verify = %v, want nil", err)
	}
	for _, want := range wantErr {
		if !errors.Is(err, want) {
			t.Errorf("Verify = %v, want an error wrapping %v", err, want)
		}
	}
	if err != nil && strings.Contains(err.Error(), secret) {
		t.Errorf("Verify error %q contains the secret", err)
	}
	if errors.Is(err, errStore) && errors.Is(err, app.ErrInvalidToken) {
		t.Error("a store failure was reported as an invalid token")
	}
}

// issued returns the fixture's secret.
func (f *fixture) issued() string { return f.secret }

// otherID returns a well-formed secret with an ID that is not stored.
func (f *fixture) otherID() string {
	return identity.FormatSecret("cdcdcdcdcdcdcdcd", bytes.Repeat([]byte{7}, identity.KeyBytes))
}

// otherKey returns the stored ID with another key.
func (f *fixture) otherKey() string {
	return identity.FormatSecret(f.token.ID, bytes.Repeat([]byte{8}, identity.KeyBytes))
}

// at sets the fixture's clock to t.
func at(now time.Time) func(t *testing.T, f *fixture) {
	return func(_ *testing.T, f *fixture) { f.clock.set(now) }
}

// revoke revokes the fixture's token.
func revoke(t *testing.T, f *fixture) {
	t.Helper()
	if err := f.tokens.Revoke(t.Context(), f.token.ID); err != nil {
		t.Fatalf("Revoke: %v", err)
	}
}

// failStore makes the fixture's store fail.
func failStore(_ *testing.T, f *fixture) { f.store.fail = errStore }

func TestRevoke(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	first := issuedAt().Add(time.Hour)
	f.clock.set(first)
	if err := f.tokens.Revoke(t.Context(), f.token.ID); err != nil {
		t.Fatalf("Revoke: %v", err)
	}
	f.clock.set(first.Add(time.Hour))
	if err := f.tokens.Revoke(t.Context(), f.token.ID); err != nil {
		t.Fatalf("Revoke again: %v", err)
	}
	if got := f.store.tokens[f.token.ID].token.RevokedAt; !got.Equal(first) {
		t.Errorf("RevokedAt = %s, want the first revocation %s", got, first)
	}
	tests := []struct {
		id      identity.ScriptTokenID
		wantErr error
	}{
		{id: "cdcdcdcdcdcdcdcd", wantErr: app.ErrUnknownToken},
		{id: "nope", wantErr: identity.ErrInvalidID},
	}
	for _, tt := range tests {
		if err := f.tokens.Revoke(t.Context(), tt.id); !errors.Is(err, tt.wantErr) {
			t.Errorf("Revoke(%q) = %v, want %v", tt.id, err, tt.wantErr)
		}
	}
}
