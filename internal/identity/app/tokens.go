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

// Get returns the token id, or an error wrapping ErrUnknownToken.
func (t *Tokens) Get(ctx context.Context, id identity.ScriptTokenID) (identity.ScriptToken, error) {
	tok, _, err := t.store.ScriptToken(ctx, id)
	if err != nil {
		return identity.ScriptToken{}, fmt.Errorf("loading script token %s: %w", id, err)
	}
	return tok, nil
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

// auditedAtomic is what a token's creation or revocation needs of a store
// besides Store: one transaction with its audit entry.
type auditedAtomic interface {
	Atomic
	Auditor
}

// IssueToken creates a read-only token for actor, as Tokens.Issue does, and
// records in the audit log, in the same transaction, that actor created it
// (D-211): if the entry can't be written, no token is created. The entry
// names the token by its ID; the secret is returned only here and never
// reaches the log. store must be over the same database as t's store, so one
// transaction covers both. Besides Auth.CreateScriptToken, core token issue
// calls it, as the core itself.
func IssueToken(ctx context.Context, store auditedAtomic, t *Tokens, actor identity.Actor, name string, ttl time.Duration) (identity.ScriptToken, string, error) {
	var (
		tok    identity.ScriptToken
		secret string
	)
	err := within(ctx, store, func(ctx context.Context) error {
		var issueErr error
		if tok, secret, issueErr = t.Issue(ctx, name, ttl); issueErr != nil {
			return issueErr
		}
		return appendTokenEntry(ctx, store, t, actor, identity.ActionScriptTokenCreated, tok.ID)
	})
	if err != nil {
		return identity.ScriptToken{}, "", err
	}
	return tok, secret, nil
}

// RevokeToken revokes the token id as Tokens.Revoke does, and records in the
// audit log, in the same transaction, that actor did (D-211).
func RevokeToken(ctx context.Context, store auditedAtomic, t *Tokens, actor identity.Actor, id identity.ScriptTokenID) error {
	return within(ctx, store, func(ctx context.Context) error {
		if err := t.Revoke(ctx, id); err != nil {
			return err
		}
		return appendTokenEntry(ctx, store, t, actor, identity.ActionScriptTokenRevoked, id)
	})
}

// appendTokenEntry appends the entry that actor did action to the token id,
// naming it by its ID alone.
func appendTokenEntry(ctx context.Context, store Auditor, t *Tokens, actor identity.Actor, action identity.AuditAction, id identity.ScriptTokenID) error {
	e, err := identity.NewAuditEntry(t.now(), actor, action, identity.OutcomeOK, identity.Ref{Kind: identity.RefScriptToken, ID: string(id)})
	if err != nil {
		return fmt.Errorf("building an audit entry: %w", err)
	}
	if err := store.AppendAudit(ctx, &e); err != nil {
		return fmt.Errorf("appending an audit entry: %w", err)
	}
	return nil
}
