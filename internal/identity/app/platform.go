package app

import (
	"context"
	"fmt"
	"io"

	"github.com/arkadiuss/openfief/internal/identity"
)

// SecretCreator creates a Secret only when none of its name exists (D-216).
// It never reads, changes or deletes one. *identity/kube.SecretCreator
// implements it.
type SecretCreator interface {
	// CreateIfMissing creates s and reports true, or reports false and
	// leaves the Secret that exists untouched.
	CreateIfMissing(ctx context.Context, s identity.PlatformSecret) (created bool, err error)
}

// PlatformOutcome is what happened to one of the platform's Secrets.
type PlatformOutcome struct {
	// Name is the Secret's name, never its value.
	Name string
	// Created is true when the Secret was made, false when one existed and
	// was kept as it is.
	Created bool
}

// InitPlatformCredentials makes the platform's credentials (D-216): each
// value is identity.PlatformCredentialBytes bytes read from random, and each
// Secret of identity.PlatformSecrets is created only if it is missing, so
// an owner who made one first (D-29) keeps theirs, and running it again
// rotates nothing. It returns the outcomes so far and stops at the first
// failure. Neither an outcome nor an error holds a value.
func InitPlatformCredentials(ctx context.Context, creator SecretCreator, random io.Reader) ([]PlatformOutcome, error) {
	var values [2]string
	for i := range values {
		raw := make([]byte, identity.PlatformCredentialBytes)
		if _, err := io.ReadFull(random, raw); err != nil {
			return nil, fmt.Errorf("reading random bytes: %w", err)
		}
		v, err := identity.PlatformCredential(raw)
		if err != nil {
			return nil, fmt.Errorf("credential %d: %w", i+1, err)
		}
		values[i] = v
	}
	secrets := identity.PlatformSecrets(values[0], values[1])
	outcomes := make([]PlatformOutcome, 0, len(secrets))
	for _, s := range secrets {
		created, err := creator.CreateIfMissing(ctx, s)
		if err != nil {
			return outcomes, fmt.Errorf("secret %s: %w", s.Name, err)
		}
		outcomes = append(outcomes, PlatformOutcome{Name: s.Name, Created: created})
	}
	return outcomes, nil
}
