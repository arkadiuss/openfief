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

// DefaultScriptTokenTTL is how long a script token created without a
// lifetime lasts.
const DefaultScriptTokenTTL = 30 * 24 * time.Hour

// CreateScriptToken issues a read-only script token (D-27) for an owner
// session (author decision, auth split: script tokens come from a logged-in
// session), lasting ttl, or DefaultScriptTokenTTL when ttl is zero. It
// returns the token and its secret, shown only here.
func (a *Auth) CreateScriptToken(ctx context.Context, origin, token, name string, ttl time.Duration) (identity.ScriptToken, string, error) {
	if _, _, err := a.sessionOn(ctx, origin, token, identity.SessionScopeOwner); err != nil {
		return identity.ScriptToken{}, "", err
	}
	if ttl == 0 {
		ttl = DefaultScriptTokenTTL
	}
	tok, secret, err := a.tokens.Issue(ctx, name, ttl)
	if errors.Is(err, identity.ErrInvalidName) || errors.Is(err, identity.ErrInvalidTTL) {
		return identity.ScriptToken{}, "", fmt.Errorf("%w: %w", ErrInvalidRequest, err)
	}
	if err != nil {
		return identity.ScriptToken{}, "", err
	}
	return tok, secret, nil
}

// ScriptTokens returns every script token for an owner session.
func (a *Auth) ScriptTokens(ctx context.Context, origin, token string) ([]identity.ScriptToken, error) {
	if _, _, err := a.sessionOn(ctx, origin, token, identity.SessionScopeOwner); err != nil {
		return nil, err
	}
	return a.tokens.List(ctx)
}

// RevokeScriptToken revokes the script token id for an owner session.
func (a *Auth) RevokeScriptToken(ctx context.Context, origin, token, id string) error {
	if _, _, err := a.sessionOn(ctx, origin, token, identity.SessionScopeOwner); err != nil {
		return err
	}
	tid, err := identity.ParseScriptTokenID(id)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidRequest, err)
	}
	err = a.tokens.Revoke(ctx, tid)
	if errors.Is(err, ErrUnknownToken) {
		return fmt.Errorf("%w: %w", ErrNotFound, err)
	}
	return err
}

// EnrollmentLink is a one-time link for creating a passkey on one origin.
type EnrollmentLink struct {
	URL       string
	ExpiresAt time.Time
}

// CreateEnrollmentLink issues, for an owner session on the allowed origin
// text, a one-time link for creating a passkey on target, another allowed
// origin or, when empty, origin itself (D-60). It lasts ttl, or
// identity.DefaultEnrollmentTTL when ttl is zero.
func (a *Auth) CreateEnrollmentLink(ctx context.Context, origin, token, target string, ttl time.Duration) (EnrollmentLink, error) {
	o, _, err := a.sessionOn(ctx, origin, token, identity.SessionScopeOwner)
	if err != nil {
		return EnrollmentLink{}, err
	}
	t := o
	if target != "" {
		if t, err = a.origin(target); err != nil {
			return EnrollmentLink{}, err
		}
	}
	return IssueEnrollmentLink(ctx, a.store, a.random, a.now(), t, ttl)
}

// IssueEnrollmentLink stores a one-time enrollment code for creating a
// passkey on origin and returns its link (D-60), with secrets from random
// (crypto/rand.Reader when nil), issued at now and lasting ttl, or
// identity.DefaultEnrollmentTTL when ttl is zero. It checks no session:
// besides CreateEnrollmentLink, core enroll-link calls it, the break-glass
// way in that prints the first link. The link's origin need not be on the
// serving core's allow-list yet; BeginRegistration checks both.
func IssueEnrollmentLink(ctx context.Context, store AuthStore, random io.Reader, now time.Time, origin identity.Origin, ttl time.Duration) (EnrollmentLink, error) {
	if random == nil {
		random = rand.Reader
	}
	now = now.UTC().Truncate(time.Second)
	if ttl == 0 {
		ttl = identity.DefaultEnrollmentTTL
	}
	idBytes := make([]byte, identity.IDBytes)
	key := make([]byte, identity.KeyBytes)
	if _, err := io.ReadFull(random, idBytes); err != nil {
		return EnrollmentLink{}, fmt.Errorf("issuing an enrollment link: random ID: %w", err)
	}
	if _, err := io.ReadFull(random, key); err != nil {
		return EnrollmentLink{}, fmt.Errorf("issuing an enrollment link: random key: %w", err)
	}
	c, err := identity.NewEnrollmentCode(identity.EnrollmentCodeID(hex.EncodeToString(idBytes)), origin, now, ttl)
	if err != nil {
		return EnrollmentLink{}, fmt.Errorf("%w: %w", ErrInvalidRequest, err)
	}
	if err := store.SaveEnrollmentCode(ctx, &c, identity.HashKey(key)); err != nil {
		return EnrollmentLink{}, fmt.Errorf("issuing an enrollment link: %w", err)
	}
	code := identity.FormatEnrollmentCode(c.ID, key)
	return EnrollmentLink{URL: identity.EnrollmentURL(origin, code), ExpiresAt: c.ExpiresAt}, nil
}
