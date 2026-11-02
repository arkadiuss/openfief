package app

import (
	"context"
	"errors"

	"github.com/arkadiuss/openfief/internal/identity"
)

// ErrInvalidToken reports a bearer token that may not read the API: unknown,
// expired, revoked or without read access. TokenVerifier implementations
// return it wrapped.
var ErrInvalidToken = errors.New("invalid token")

// TokenVerifier checks the bearer token of a read API request: a session, a
// script token or a plugin's projected ServiceAccount token (D-27).
type TokenVerifier interface {
	// VerifyToken returns the caller token stands for when it may read the
	// API, an error wrapping ErrInvalidToken when it may not, and any other
	// error when it cannot be checked now.
	VerifyToken(ctx context.Context, token string) (identity.Caller, error)
}
