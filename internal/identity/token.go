package identity

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"fmt"
	"time"
)

// Errors about script tokens. Checks return them wrapped.
var (
	// ErrInvalidName reports a token name that is empty, too long or has
	// characters other than letters, digits, '.', '_' and '-'.
	ErrInvalidName = errors.New("invalid script token name")
	// ErrInvalidTTL reports a lifetime that is not positive or exceeds
	// MaxTTL.
	ErrInvalidTTL = errors.New("invalid script token lifetime")
	// ErrInvalidScope reports a scope outside the declared constants, or
	// text that names none.
	ErrInvalidScope = errors.New("invalid script token scope")
	// ErrInvalidID reports a token ID that is not IDBytes bytes in lowercase
	// hex.
	ErrInvalidID = errors.New("invalid script token ID")
	// ErrMalformedSecret reports text that is not a script token secret.
	ErrMalformedSecret = errors.New("malformed script token secret")
	// ErrTokenExpired reports a token used at or after its expiry.
	ErrTokenExpired = errors.New("script token expired")
	// ErrTokenRevoked reports a revoked token.
	ErrTokenRevoked = errors.New("script token revoked")
	// ErrScopeNotGranted reports a token used for a scope it does not have.
	ErrScopeNotGranted = errors.New("script token scope not granted")
)

const (
	// SecretPrefix starts every script token secret, so a leaked one is easy
	// to recognize and to scan for.
	SecretPrefix = "ofs_"
	// IDBytes is the number of random bytes in a token ID.
	IDBytes = 8
	// KeyBytes is the number of random bytes in a token's key, the part of
	// the secret that is hashed.
	KeyBytes = 32
	// MaxTTL is the longest lifetime a script token may have.
	MaxTTL = 365 * 24 * time.Hour
	// maxNameLen is the longest token name.
	maxNameLen = 64
	// secretSep separates the ID and the key in a secret.
	secretSep = "_"
)

// Scope is what a script token may do. In M0 the only scope is ScopeRead:
// the public API is read-only (D-6), and so is every script token (D-27).
type Scope int

// The scopes. The zero Scope is not a scope.
const (
	// ScopeRead may read the read API.
	ScopeRead Scope = iota + 1
)

// String returns the scope's name, such as "read".
func (s Scope) String() string {
	if s == ScopeRead {
		return "read"
	}
	return fmt.Sprintf("Scope(%d)", int(s))
}

// ParseScope parses a scope name as String writes it.
func ParseScope(s string) (Scope, error) {
	if s == ScopeRead.String() {
		return ScopeRead, nil
	}
	return 0, fmt.Errorf("%w: %q", ErrInvalidScope, s)
}

// ScriptTokenID identifies a script token. It is not secret: the owner
// lists and revokes tokens by it, and the log may name it.
type ScriptTokenID string

// ParseScriptTokenID checks that s is IDBytes bytes in lowercase hex.
func ParseScriptTokenID(s string) (ScriptTokenID, error) {
	if !isHexID(s) {
		return "", fmt.Errorf("%w: want %d lowercase hex digits", ErrInvalidID, hex.EncodedLen(IDBytes))
	}
	return ScriptTokenID(s), nil
}

// ScriptToken is a scoped, expiring personal API token for the owner's
// scripts (D-27). It holds no secret: the secret is shown once at issue,
// and only its hash is stored.
type ScriptToken struct {
	ID        ScriptTokenID
	Name      string
	Scope     Scope
	CreatedAt time.Time
	ExpiresAt time.Time
	// RevokedAt is when the token was revoked; zero while it is not.
	RevokedAt time.Time
}

// NewScriptToken returns a token with id, name and scope, created at
// createdAt and expiring ttl later. The name must be 1 to 64 letters,
// digits, '.', '_' or '-', so that listing it prints nothing unexpected;
// ttl must be positive and at most MaxTTL.
func NewScriptToken(id ScriptTokenID, name string, scope Scope, createdAt time.Time, ttl time.Duration) (ScriptToken, error) {
	if _, err := ParseScriptTokenID(string(id)); err != nil {
		return ScriptToken{}, err
	}
	if err := checkName(name); err != nil {
		return ScriptToken{}, err
	}
	if _, err := ParseScope(scope.String()); err != nil {
		return ScriptToken{}, err
	}
	if ttl <= 0 || ttl > MaxTTL {
		return ScriptToken{}, fmt.Errorf("%w: %s, want more than 0 and at most %s", ErrInvalidTTL, ttl, MaxTTL)
	}
	return ScriptToken{ID: id, Name: name, Scope: scope, CreatedAt: createdAt, ExpiresAt: createdAt.Add(ttl)}, nil
}

// checkName checks a token name (see NewScriptToken).
func checkName(name string) error {
	if name == "" || len(name) > maxNameLen {
		return fmt.Errorf("%w: want 1 to %d characters", ErrInvalidName, maxNameLen)
	}
	for _, r := range name {
		if !nameRune(r) {
			return fmt.Errorf("%w: only letters, digits, '.', '_' and '-'", ErrInvalidName)
		}
	}
	return nil
}

// nameRune reports whether r may appear in a token name: an ASCII letter or
// digit, '.', '_' or '-'.
func nameRune(r rune) bool {
	switch {
	case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9':
		return true
	}
	return r == '.' || r == '_' || r == '-'
}

// Check returns nil when t may be used for want at now: it is not revoked,
// now is before its expiry, and it has the scope. Otherwise it returns
// ErrTokenRevoked, ErrTokenExpired or ErrScopeNotGranted, in that order.
func (t *ScriptToken) Check(now time.Time, want Scope) error {
	switch {
	case !t.RevokedAt.IsZero():
		return ErrTokenRevoked
	case !now.Before(t.ExpiresAt):
		return ErrTokenExpired
	case t.Scope != want:
		return fmt.Errorf("%w: %s", ErrScopeNotGranted, want)
	}
	return nil
}

// Status returns "revoked", "expired" or "active": what Check would say
// about t at now, for listing.
func (t *ScriptToken) Status(now time.Time) string {
	switch {
	case !t.RevokedAt.IsZero():
		return "revoked"
	case !now.Before(t.ExpiresAt):
		return "expired"
	}
	return "active"
}

// FormatSecret returns the secret of the token id with key:
// SecretPrefix, the ID, "_" and the key in unpadded base64url.
func FormatSecret(id ScriptTokenID, key []byte) string {
	return formatSecret(SecretPrefix, string(id), key)
}

// ParseSecret splits a secret written by FormatSecret into the token's ID
// and key. Its errors never contain the secret.
func ParseSecret(secret string) (ScriptTokenID, []byte, error) {
	parts, err := parseSecret(SecretPrefix, secret, ErrMalformedSecret)
	if err != nil {
		return "", nil, err
	}
	return ScriptTokenID(parts.id), parts.key, nil
}

// HashKey returns the SHA-256 hash of a token's key: the only form of the
// secret that is stored.
func HashKey(key []byte) []byte {
	sum := sha256.Sum256(key)
	return sum[:]
}

// KeyMatches reports whether key hashes to stored, comparing the hashes in
// constant time.
func KeyMatches(key, stored []byte) bool {
	return subtle.ConstantTimeCompare(HashKey(key), stored) == 1
}
