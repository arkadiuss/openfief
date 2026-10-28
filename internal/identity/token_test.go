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

// testID is a synthetic token ID.
const testID identity.ScriptTokenID = "0123456789abcdef"

// createdAt is when the test tokens are created.
func createdAt() time.Time { return time.Date(2026, 10, 28, 12, 0, 0, 0, time.UTC) }

func TestNewScriptToken(t *testing.T) {
	t.Parallel()
	created := createdAt()
	tests := []struct {
		name    string
		id      identity.ScriptTokenID
		tokName string
		scope   identity.Scope
		ttl     time.Duration
		want    identity.ScriptToken
		wantErr error
	}{
		{
			name: "valid", id: testID, tokName: "demo-script_1.0", scope: identity.ScopeRead, ttl: 720 * time.Hour,
			want: identity.ScriptToken{
				ID: testID, Name: "demo-script_1.0", Scope: identity.ScopeRead,
				CreatedAt: created, ExpiresAt: created.Add(720 * time.Hour),
			},
		},
		{
			name: "longest lifetime", id: testID, tokName: "x", scope: identity.ScopeRead, ttl: identity.MaxTTL,
			want: identity.ScriptToken{ID: testID, Name: "x", Scope: identity.ScopeRead, CreatedAt: created, ExpiresAt: created.Add(identity.MaxTTL)},
		},
		{name: "empty name", id: testID, tokName: "", scope: identity.ScopeRead, ttl: time.Hour, wantErr: identity.ErrInvalidName},
		{name: "long name", id: testID, tokName: strings.Repeat("a", 65), scope: identity.ScopeRead, ttl: time.Hour, wantErr: identity.ErrInvalidName},
		{name: "space in name", id: testID, tokName: "my token", scope: identity.ScopeRead, ttl: time.Hour, wantErr: identity.ErrInvalidName},
		{name: "control character in name", id: testID, tokName: "a\nb", scope: identity.ScopeRead, ttl: time.Hour, wantErr: identity.ErrInvalidName},
		{name: "zero lifetime", id: testID, tokName: "x", scope: identity.ScopeRead, ttl: 0, wantErr: identity.ErrInvalidTTL},
		{name: "negative lifetime", id: testID, tokName: "x", scope: identity.ScopeRead, ttl: -time.Hour, wantErr: identity.ErrInvalidTTL},
		{name: "lifetime over a year", id: testID, tokName: "x", scope: identity.ScopeRead, ttl: identity.MaxTTL + time.Second, wantErr: identity.ErrInvalidTTL},
		{name: "zero scope", id: testID, tokName: "x", scope: 0, ttl: time.Hour, wantErr: identity.ErrInvalidScope},
		{name: "unknown scope", id: testID, tokName: "x", scope: 7, ttl: time.Hour, wantErr: identity.ErrInvalidScope},
		{name: "bad ID", id: "XYZ", tokName: "x", scope: identity.ScopeRead, ttl: time.Hour, wantErr: identity.ErrInvalidID},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := identity.NewScriptToken(tt.id, tt.tokName, tt.scope, created, tt.ttl)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("NewScriptToken error = %v, want %v", err, tt.wantErr)
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("NewScriptToken (-want +got):\n%s", diff)
			}
		})
	}
}

func TestScriptTokenCheck(t *testing.T) {
	t.Parallel()
	created := createdAt()
	expires := created.Add(time.Hour)
	tests := []struct {
		name       string
		revokedAt  time.Time
		now        time.Time
		want       identity.Scope
		wantErr    error
		wantStatus string
	}{
		{name: "active", now: created, want: identity.ScopeRead, wantStatus: "active"},
		{name: "a nanosecond before expiry", now: expires.Add(-time.Nanosecond), want: identity.ScopeRead, wantStatus: "active"},
		{name: "at expiry", now: expires, want: identity.ScopeRead, wantErr: identity.ErrTokenExpired, wantStatus: "expired"},
		{name: "after expiry", now: expires.Add(time.Hour), want: identity.ScopeRead, wantErr: identity.ErrTokenExpired, wantStatus: "expired"},
		{name: "revoked", revokedAt: created.Add(time.Minute), now: created.Add(2 * time.Minute), want: identity.ScopeRead, wantErr: identity.ErrTokenRevoked, wantStatus: "revoked"},
		{name: "revoked wins over expired", revokedAt: created.Add(time.Minute), now: expires.Add(time.Hour), want: identity.ScopeRead, wantErr: identity.ErrTokenRevoked, wantStatus: "revoked"},
		{name: "scope not granted", now: created, want: 2, wantErr: identity.ErrScopeNotGranted, wantStatus: "active"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tok := identity.ScriptToken{ID: testID, Name: "x", Scope: identity.ScopeRead, CreatedAt: created, ExpiresAt: expires, RevokedAt: tt.revokedAt}
			if err := tok.Check(tt.now, tt.want); !errors.Is(err, tt.wantErr) {
				t.Errorf("Check = %v, want %v", err, tt.wantErr)
			}
			if got := tok.Status(tt.now); got != tt.wantStatus {
				t.Errorf("Status = %q, want %q", got, tt.wantStatus)
			}
		})
	}
}

func TestScope(t *testing.T) {
	t.Parallel()
	tests := []struct {
		text    string
		want    identity.Scope
		wantErr error
	}{
		{text: "read", want: identity.ScopeRead},
		{text: "write", wantErr: identity.ErrInvalidScope},
		{text: "", wantErr: identity.ErrInvalidScope},
		{text: "Scope(0)", wantErr: identity.ErrInvalidScope},
	}
	for _, tt := range tests {
		t.Run(tt.text, func(t *testing.T) {
			t.Parallel()
			got, err := identity.ParseScope(tt.text)
			if !errors.Is(err, tt.wantErr) || got != tt.want {
				t.Fatalf("ParseScope(%q) = %v, %v; want %v, %v", tt.text, got, err, tt.want, tt.wantErr)
			}
			if err == nil && got.String() != tt.text {
				t.Errorf("String = %q, want %q", got.String(), tt.text)
			}
		})
	}
	if got := identity.Scope(0).String(); got != "Scope(0)" {
		t.Errorf("Scope(0).String() = %q", got)
	}
}

func TestParseScriptTokenID(t *testing.T) {
	t.Parallel()
	tests := []struct {
		text    string
		wantErr error
	}{
		{text: "0123456789abcdef"},
		{text: "0123456789ABCDEF", wantErr: identity.ErrInvalidID},
		{text: "0123456789abcde", wantErr: identity.ErrInvalidID},
		{text: "0123456789abcdeg", wantErr: identity.ErrInvalidID},
		{text: "", wantErr: identity.ErrInvalidID},
	}
	for _, tt := range tests {
		t.Run(tt.text, func(t *testing.T) {
			t.Parallel()
			got, err := identity.ParseScriptTokenID(tt.text)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("ParseScriptTokenID(%q) error = %v, want %v", tt.text, err, tt.wantErr)
			}
			if err == nil && string(got) != tt.text {
				t.Errorf("ParseScriptTokenID(%q) = %q", tt.text, got)
			}
		})
	}
}

func TestSecret(t *testing.T) {
	t.Parallel()
	key := bytes.Repeat([]byte{0xfb}, identity.KeyBytes)
	// 0xfb 0xfb 0xfb is the bits 111110 111111 101111 111011: "-_v7" in base64url.
	secret := identity.FormatSecret(testID, key)
	if want := "ofs_0123456789abcdef_" + strings.Repeat("-_v7", 10) + "-_s"; secret != want {
		t.Fatalf("FormatSecret = %q, want %q", secret, want)
	}
	tests := []struct {
		name    string
		secret  string
		wantErr error
	}{
		{name: "valid", secret: secret},
		{name: "no prefix", secret: strings.TrimPrefix(secret, "ofs_"), wantErr: identity.ErrMalformedSecret},
		{name: "other prefix", secret: "ghp_" + strings.TrimPrefix(secret, "ofs_"), wantErr: identity.ErrMalformedSecret},
		{name: "no key", secret: "ofs_0123456789abcdef", wantErr: identity.ErrMalformedSecret},
		{name: "bad ID", secret: "ofs_0123_" + strings.Repeat("A", 43), wantErr: identity.ErrMalformedSecret},
		{name: "short key", secret: "ofs_0123456789abcdef_AAAA", wantErr: identity.ErrMalformedSecret},
		{name: "padded key", secret: secret + "=", wantErr: identity.ErrMalformedSecret},
		{name: "not base64url", secret: "ofs_0123456789abcdef_" + strings.Repeat("+", 43), wantErr: identity.ErrMalformedSecret},
		{name: "empty", secret: "", wantErr: identity.ErrMalformedSecret},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			checkParsed(t, tt.secret, key, tt.wantErr)
		})
	}
}

// checkParsed checks that ParseSecret(secret) returns testID and key, or
// fails with wantErr and an error that does not contain the secret.
func checkParsed(t *testing.T, secret string, key []byte, wantErr error) {
	t.Helper()
	id, got, err := identity.ParseSecret(secret)
	if !errors.Is(err, wantErr) {
		t.Fatalf("ParseSecret error = %v, want %v", err, wantErr)
	}
	if err != nil {
		if secret != "" && strings.Contains(err.Error(), secret) {
			t.Errorf("ParseSecret error %q contains the secret", err)
		}
		return
	}
	if id != testID || !bytes.Equal(got, key) {
		t.Errorf("ParseSecret = %q, %x; want %q, %x", id, got, testID, key)
	}
}

func TestKeyMatches(t *testing.T) {
	t.Parallel()
	key := bytes.Repeat([]byte{1}, identity.KeyBytes)
	other := bytes.Repeat([]byte{2}, identity.KeyBytes)
	stored := identity.HashKey(key)
	tests := []struct {
		name   string
		key    []byte
		stored []byte
		want   bool
	}{
		{name: "same key", key: key, stored: stored, want: true},
		{name: "other key", key: other, stored: stored},
		{name: "truncated hash", key: key, stored: stored[:16]},
		{name: "no hash", key: key, stored: nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := identity.KeyMatches(tt.key, tt.stored); got != tt.want {
				t.Errorf("KeyMatches = %v, want %v", got, tt.want)
			}
		})
	}
	if len(stored) != 32 || bytes.Equal(stored, key) {
		t.Errorf("HashKey = %x: want a 32-byte SHA-256 hash, not the key", stored)
	}
}
