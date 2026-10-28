package app

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/arkadiuss/openfief/internal/identity"
)

// Errors of the token use cases.
var (
	// ErrUnknownToken reports a token ID that is not stored. Stores return it
	// wrapped.
	ErrUnknownToken = errors.New("unknown script token")
	// ErrInvalidToken reports a secret that may not be used: malformed,
	// unknown, not matching its stored hash, expired, revoked or without the
	// scope. Verify returns it wrapped, with the reason but never the secret.
	ErrInvalidToken = errors.New("invalid script token")
)

// Store keeps script tokens with the SHA-256 hash of their key, never the
// key or the secret. *identity/postgres.Store implements it.
type Store interface {
	// SaveScriptToken stores a new token with its key's hash.
	SaveScriptToken(ctx context.Context, t *identity.ScriptToken, keyHash []byte) error
	// ScriptToken returns the token id and its key's hash, or an error
	// wrapping ErrUnknownToken.
	ScriptToken(ctx context.Context, id identity.ScriptTokenID) (identity.ScriptToken, []byte, error)
	// RevokeScriptToken marks the token id revoked at at, keeping an earlier
	// revocation's time, or fails with an error wrapping ErrUnknownToken.
	RevokeScriptToken(ctx context.Context, id identity.ScriptTokenID, at time.Time) error
	// ScriptTokens returns every token, in creation then ID order.
	ScriptTokens(ctx context.Context) ([]identity.ScriptToken, error)
}

// Tokens are the script token use cases.
type Tokens struct {
	store  Store
	now    func() time.Time
	random io.Reader
}

// NewTokens returns the use cases over store, reading the time from now and
// new IDs and keys from random; a nil random is crypto/rand.Reader.
func NewTokens(store Store, now func() time.Time, random io.Reader) *Tokens {
	if random == nil {
		random = rand.Reader
	}
	return &Tokens{store: store, now: now, random: random}
}

// Issue creates a read-only token named name that expires ttl from now, and
// returns it with its secret. The secret is returned only here: the store
// keeps its key's hash, so no one can show it again.
func (t *Tokens) Issue(ctx context.Context, name string, ttl time.Duration) (identity.ScriptToken, string, error) {
	idBytes := make([]byte, identity.IDBytes)
	key := make([]byte, identity.KeyBytes)
	if _, err := io.ReadFull(t.random, idBytes); err != nil {
		return identity.ScriptToken{}, "", fmt.Errorf("issuing a script token: random ID: %w", err)
	}
	if _, err := io.ReadFull(t.random, key); err != nil {
		return identity.ScriptToken{}, "", fmt.Errorf("issuing a script token: random key: %w", err)
	}
	id := identity.ScriptTokenID(hex.EncodeToString(idBytes))
	tok, err := identity.NewScriptToken(id, name, identity.ScopeRead, t.now().UTC().Truncate(time.Second), ttl)
	if err != nil {
		return identity.ScriptToken{}, "", fmt.Errorf("issuing a script token: %w", err)
	}
	if err := t.store.SaveScriptToken(ctx, &tok, identity.HashKey(key)); err != nil {
		return identity.ScriptToken{}, "", fmt.Errorf("issuing a script token: %w", err)
	}
	return tok, identity.FormatSecret(id, key), nil
}

// Verify returns nil when secret is a stored token's, matches its hash, and
// may be used for want now. Otherwise it returns an error wrapping
// ErrInvalidToken, or another error when the store fails. The error never
// contains the secret.
func (t *Tokens) Verify(ctx context.Context, secret string, want identity.Scope) error {
	id, key, err := identity.ParseSecret(secret)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidToken, err)
	}
	tok, hash, err := t.store.ScriptToken(ctx, id)
	if errors.Is(err, ErrUnknownToken) {
		return fmt.Errorf("%w: unknown", ErrInvalidToken)
	}
	if err != nil {
		return fmt.Errorf("verifying a script token: %w", err)
	}
	if !identity.KeyMatches(key, hash) {
		return fmt.Errorf("%w: key does not match", ErrInvalidToken)
	}
	if err := tok.Check(t.now(), want); err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidToken, err)
	}
	return nil
}

// Revoke revokes the token id from now on. Revoking a revoked token again
// keeps its first revocation; an unknown ID fails with an error wrapping
// ErrUnknownToken.
func (t *Tokens) Revoke(ctx context.Context, id identity.ScriptTokenID) error {
	if _, err := identity.ParseScriptTokenID(string(id)); err != nil {
		return fmt.Errorf("revoking a script token: %w", err)
	}
	if err := t.store.RevokeScriptToken(ctx, id, t.now().UTC()); err != nil {
		return fmt.Errorf("revoking script token %s: %w", id, err)
	}
	return nil
}

// List returns every token, expired and revoked ones included, in creation
// order.
func (t *Tokens) List(ctx context.Context) ([]identity.ScriptToken, error) {
	toks, err := t.store.ScriptTokens(ctx)
	if err != nil {
		return nil, fmt.Errorf("listing script tokens: %w", err)
	}
	return toks, nil
}
