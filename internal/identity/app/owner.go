package app

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"slices"
	"time"

	"github.com/arkadiuss/openfief/internal/identity"
)

// DefaultScriptTokenTTL is how long a script token created without a
// lifetime lasts.
const DefaultScriptTokenTTL = 30 * 24 * time.Hour

// CreateScriptToken issues a read-only script token (D-27) for an owner
// session (author decision, auth split: script tokens come from a logged-in
// session) with a passkey check in the last five minutes (D-183), lasting
// ttl, or DefaultScriptTokenTTL when ttl is zero. It returns the token and
// its secret, shown only here.
func (a *Auth) CreateScriptToken(ctx context.Context, origin, token, name string, ttl time.Duration) (identity.ScriptToken, string, error) {
	_, s, err := a.recentOwner(ctx, origin, token)
	if err != nil {
		return identity.ScriptToken{}, "", err
	}
	if ttl == 0 {
		ttl = DefaultScriptTokenTTL
	}
	tok, secret, err := IssueToken(ctx, a.store, a.tokens, ownerActor(&s), name, ttl)
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
	_, s, err := a.sessionOn(ctx, origin, token, identity.SessionScopeOwner)
	if err != nil {
		return err
	}
	tid, err := identity.ParseScriptTokenID(id)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidRequest, err)
	}
	err = RevokeToken(ctx, a.store, a.tokens, ownerActor(&s), tid)
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
// text with a passkey check in the last five minutes (D-183), a one-time link for creating a passkey on target, another allowed
// origin or, when empty, origin itself (D-60). It lasts ttl, or
// identity.DefaultEnrollmentTTL when ttl is zero.
func (a *Auth) CreateEnrollmentLink(ctx context.Context, origin, token, target string, ttl time.Duration) (EnrollmentLink, error) {
	o, s, err := a.recentOwner(ctx, origin, token)
	if err != nil {
		return EnrollmentLink{}, err
	}
	t := o
	if target != "" {
		if t, err = a.origin(target); err != nil {
			return EnrollmentLink{}, err
		}
	}
	return IssueEnrollmentLink(ctx, a.store, a.random, a.now(), a.cfg.Origins, t, ttl, ownerActor(&s))
}

// IssueEnrollmentLink stores a one-time enrollment code for creating a
// passkey on origin and returns its link (D-60), recording in the audit log,
// in the same transaction, that actor issued it (D-211), with secrets from random
// (crypto/rand.Reader when nil), issued at now and lasting ttl, or
// identity.DefaultEnrollmentTTL when ttl is zero. It checks no session:
// besides CreateEnrollmentLink, core enroll-link calls it, the break-glass
// way in that prints the first link. allowed is the core's one setting of
// the origins it accepts, which serving reads too (D-126): an origin not on
// it is refused with an error wrapping ErrForbidden and
// identity.ErrOriginNotAllowed, and nothing is stored, since a passkey made
// there could never sign in.
func IssueEnrollmentLink(ctx context.Context, store AuthStore, random io.Reader, now time.Time, allowed identity.Origins, origin identity.Origin, ttl time.Duration, actor identity.Actor) (EnrollmentLink, error) {
	if !slices.Contains(allowed.List(), origin) {
		return EnrollmentLink{}, fmt.Errorf("%w: %w: %s; the core accepts %s", ErrForbidden, identity.ErrOriginNotAllowed, origin, allowed)
	}
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
	if err := storeLink(ctx, store, &c, identity.HashKey(key), actor, now); err != nil {
		return EnrollmentLink{}, fmt.Errorf("issuing an enrollment link: %w", err)
	}
	code := identity.FormatEnrollmentCode(c.ID, key)
	return EnrollmentLink{URL: identity.EnrollmentURL(origin, code), ExpiresAt: c.ExpiresAt}, nil
}

// storeLink stores the enrollment code c with its key's hash and the audit
// entry that actor issued it at now, in one transaction.
func storeLink(ctx context.Context, store AuthStore, c *identity.EnrollmentCode, keyHash []byte, actor identity.Actor, now time.Time) error {
	entry, err := identity.NewAuditEntry(now, actor, identity.ActionEnrollmentLinkIssued, identity.OutcomeOK,
		identity.Ref{Kind: identity.RefEnrollmentCode, ID: string(c.ID)})
	if err != nil {
		return fmt.Errorf("building an audit entry: %w", err)
	}
	return within(ctx, store, func(ctx context.Context) error {
		if saveErr := store.SaveEnrollmentCode(ctx, c, keyHash); saveErr != nil {
			return fmt.Errorf("saving the enrollment code: %w", saveErr)
		}
		if appendErr := store.AppendAudit(ctx, &entry); appendErr != nil {
			return fmt.Errorf("appending the audit entry: %w", appendErr)
		}
		return nil
	})
}
