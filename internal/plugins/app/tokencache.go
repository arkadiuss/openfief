package app

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/arkadiuss/openfief/internal/plugins"
)

const (
	// DefaultTokenCacheTTL is how long the egress proxy trusts a reviewed
	// token before asking the TokenReviewer again (D-63). It is well below a
	// projected token's lifetime (plugins.TokenExpirationSeconds).
	DefaultTokenCacheTTL = 5 * time.Minute
	// RefusalCacheTTL is how long a refused token stays refused without
	// asking the TokenReviewer again (D-184), so a caller repeating a bad
	// token can't turn each request into a TokenReview. It is short, so a
	// view plugin refused while the operator was still starting it gets in
	// soon after.
	RefusalCacheTTL = 10 * time.Second
	// MaxTokenBytes bounds a token the cache sends to the TokenReviewer: a
	// projected ServiceAccount token is about a kilobyte.
	MaxTokenBytes = 8 << 10
	// tokenCacheSize bounds the cache. Each plugin instance holds one or two
	// live tokens at a time, so this is far above any expected install.
	tokenCacheSize = 1024
	// jwtParts is the number of dot-separated parts of a compact JWS.
	jwtParts = 3
)

// ErrTokenCacheTTL reports a cache lifetime that is not positive or not
// below a projected token's lifetime.
var ErrTokenCacheTTL = errors.New("token cache TTL must be positive and below the projected token lifetime")

// CachingTokenReviewer is a TokenReviewer that remembers what it reviewed,
// so the egress proxy and the core don't ask the API server on every
// request (D-63, D-82), and a caller can't make them ask it at will
// (D-184).
//
// A token that isn't shaped like a JSON Web Token (JWTShaped) is refused
// without asking: a projected ServiceAccount token always is one. An
// accepted token is trusted until the earlier of the cache TTL and the
// token's own expiry; a refused one stays refused for RefusalCacheTTL. A
// failure to ask, such as an unreachable API server, is never cached. The
// cache keys on a SHA-256 digest of the token, so it holds no token, and
// keeps accepted and refused tokens apart, each bounded, so refusals never
// crowd out accepted tokens.
//
// Trade-off: a token the API server would now refuse, because its pod or
// ServiceAccount was deleted, stays accepted for at most the TTL.
type CachingTokenReviewer struct {
	next TokenReviewer
	ttl  time.Duration
	now  func() time.Time

	mu       sync.Mutex
	accepted map[[sha256.Size]byte]cachedReview
	refused  map[[sha256.Size]byte]cachedReview
}

// cachedReview is one reviewed token's username, empty for a refused one,
// and when to stop trusting the answer.
type cachedReview struct {
	username string
	expires  time.Time
}

// NewCachingTokenReviewer caches next's accepted tokens for ttl, reading the
// time from now. ttl must be positive and below plugins.TokenExpirationSeconds.
func NewCachingTokenReviewer(next TokenReviewer, ttl time.Duration, now func() time.Time) (*CachingTokenReviewer, error) {
	if ttl <= 0 || ttl >= time.Duration(plugins.TokenExpirationSeconds)*time.Second {
		return nil, fmt.Errorf("%w: %v", ErrTokenCacheTTL, ttl)
	}
	return &CachingTokenReviewer{
		next: next, ttl: ttl, now: now,
		accepted: make(map[[sha256.Size]byte]cachedReview),
		refused:  make(map[[sha256.Size]byte]cachedReview),
	}, nil
}

// ReviewToken implements TokenReviewer: a token not shaped like a JWT is
// refused at once; otherwise a cached answer while it is fresh, or else
// next's, cached when it accepts or refuses the token. Every refusal wraps
// ErrUnauthenticated.
func (c *CachingTokenReviewer) ReviewToken(ctx context.Context, token string) (string, error) {
	if !JWTShaped(token) {
		return "", fmt.Errorf("%w: not a JSON Web Token", ErrUnauthenticated)
	}
	key := sha256.Sum256([]byte(token))
	now := c.now()
	c.mu.Lock()
	username, accepted := fresh(c.accepted, key, now)
	_, refused := fresh(c.refused, key, now)
	c.mu.Unlock()
	switch {
	case accepted:
		return username, nil
	case refused:
		return "", fmt.Errorf("%w: refused within the last %v", ErrUnauthenticated, RefusalCacheTTL)
	}
	username, err := c.next.ReviewToken(ctx, token)
	if errors.Is(err, ErrUnauthenticated) {
		c.store(c.refused, key, cachedReview{expires: now.Add(RefusalCacheTTL)}, now)
	}
	if err != nil {
		return "", fmt.Errorf("review token: %w", err)
	}
	expires := now.Add(c.ttl)
	if exp, ok := tokenExpiry(token); ok && exp.Before(expires) {
		expires = exp
	}
	c.store(c.accepted, key, cachedReview{username: username, expires: expires}, now)
	return username, nil
}

// fresh returns the username entries holds for key while it is fresh at
// now. The caller holds the lock.
func fresh(entries map[[sha256.Size]byte]cachedReview, key [sha256.Size]byte, now time.Time) (string, bool) {
	e, ok := entries[key]
	if !ok || !now.Before(e.expires) {
		return "", false
	}
	return e.username, true
}

// store caches e in entries unless it is already stale, dropping stale
// entries when entries is full and skipping e if it is still full.
func (c *CachingTokenReviewer) store(entries map[[sha256.Size]byte]cachedReview, key [sha256.Size]byte, e cachedReview, now time.Time) {
	if !now.Before(e.expires) {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if len(entries) >= tokenCacheSize {
		for k, old := range entries {
			if !now.Before(old.expires) {
				delete(entries, k)
			}
		}
	}
	if len(entries) < tokenCacheSize {
		entries[key] = e
	}
}

// JWTShaped reports whether token looks like a compact JSON Web Token, as
// every projected ServiceAccount token does: at most MaxTokenBytes, three
// dot-separated base64url parts, of which the header and the claims decode
// to JSON objects. It verifies nothing; it only keeps session tokens,
// script tokens and junk away from the TokenReview API (D-184).
func JWTShaped(token string) bool {
	if token == "" || len(token) > MaxTokenBytes {
		return false
	}
	parts := strings.Split(token, ".")
	if len(parts) != jwtParts || parts[2] == "" {
		return false
	}
	for _, part := range parts[:2] {
		raw, err := base64.RawURLEncoding.DecodeString(part)
		if err != nil {
			return false
		}
		var obj map[string]json.RawMessage
		if json.Unmarshal(raw, &obj) != nil || obj == nil {
			return false
		}
	}
	_, err := base64.RawURLEncoding.DecodeString(parts[2])
	return err == nil
}

// tokenExpiry reads the "exp" claim of a compact JWT without verifying it.
// Only call it on a token the TokenReviewer has accepted, which verified the
// signature; it serves only to shorten how long the token is cached.
func tokenExpiry(token string) (time.Time, bool) {
	parts := strings.Split(token, ".")
	if len(parts) != jwtParts {
		return time.Time{}, false
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return time.Time{}, false
	}
	var claims struct {
		Exp *int64 `json:"exp"`
	}
	if err := json.Unmarshal(payload, &claims); err != nil || claims.Exp == nil {
		return time.Time{}, false
	}
	return time.Unix(*claims.Exp, 0), true
}
