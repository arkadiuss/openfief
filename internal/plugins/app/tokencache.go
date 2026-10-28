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
	// tokenCacheSize bounds the cache. Each plugin instance holds one or two
	// live tokens at a time, so this is far above any expected install.
	tokenCacheSize = 1024
	// jwtParts is the number of dot-separated parts of a compact JWS.
	jwtParts = 3
)

// ErrTokenCacheTTL reports a cache lifetime that is not positive or not
// below a projected token's lifetime.
var ErrTokenCacheTTL = errors.New("token cache TTL must be positive and below the projected token lifetime")

// CachingTokenReviewer is a TokenReviewer that remembers accepted tokens, so
// the egress proxy does not ask the API server on every CONNECT (D-63).
//
// An accepted token is trusted until the earlier of the cache TTL and the
// token's own expiry. Refused tokens are never cached. The cache keys on a
// SHA-256 digest of the token, so it holds no token.
//
// Trade-off: a token the API server would now refuse, because its pod or
// ServiceAccount was deleted, stays accepted for at most the TTL.
type CachingTokenReviewer struct {
	next TokenReviewer
	ttl  time.Duration
	now  func() time.Time

	mu      sync.Mutex
	entries map[[sha256.Size]byte]cachedReview
}

// cachedReview is one accepted token's username and when to stop trusting it.
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
		entries: make(map[[sha256.Size]byte]cachedReview),
	}, nil
}

// ReviewToken implements TokenReviewer: a cached answer while it is fresh,
// otherwise next's, cached when it accepts the token.
func (c *CachingTokenReviewer) ReviewToken(ctx context.Context, token string) (string, error) {
	key := sha256.Sum256([]byte(token))
	now := c.now()
	if username, ok := c.lookup(key, now); ok {
		return username, nil
	}
	username, err := c.next.ReviewToken(ctx, token)
	if err != nil {
		return "", fmt.Errorf("review token: %w", err)
	}
	expires := now.Add(c.ttl)
	if exp, ok := tokenExpiry(token); ok && exp.Before(expires) {
		expires = exp
	}
	c.store(key, cachedReview{username: username, expires: expires}, now)
	return username, nil
}

// lookup returns the cached username for key while it is fresh at now.
func (c *CachingTokenReviewer) lookup(key [sha256.Size]byte, now time.Time) (string, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	e, ok := c.entries[key]
	if !ok || !now.Before(e.expires) {
		return "", false
	}
	return e.username, true
}

// store caches e unless it is already stale, dropping stale entries when the
// cache is full and skipping e if it is still full.
func (c *CachingTokenReviewer) store(key [sha256.Size]byte, e cachedReview, now time.Time) {
	if !now.Before(e.expires) {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if len(c.entries) >= tokenCacheSize {
		for k, old := range c.entries {
			if !now.Before(old.expires) {
				delete(c.entries, k)
			}
		}
	}
	if len(c.entries) < tokenCacheSize {
		c.entries[key] = e
	}
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
