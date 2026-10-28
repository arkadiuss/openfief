package app_test

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/arkadiuss/openfief/internal/plugins"
	"github.com/arkadiuss/openfief/internal/plugins/app"
)

// countingReviewer accepts every token except refusedToken as plugin
// "sample" and counts the reviews it was asked for.
type countingReviewer struct{ calls int }

const refusedToken = "synthetic-token-refused"

func (c *countingReviewer) ReviewToken(_ context.Context, token string) (string, error) {
	c.calls++
	if token == refusedToken {
		return "", fmt.Errorf("%w: refused by the fake", app.ErrUnauthenticated)
	}
	return "system:serviceaccount:openfief:plugin-sample", nil
}

// start is the fake clock's first reading; synthetic.
func start() time.Time { return time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC) }

// cacheStep reviews token at start()+at and expects calls reviews in total.
type cacheStep struct {
	token   string
	at      time.Duration
	calls   int
	wantErr error
}

// runCacheSteps runs steps against a fresh cache over a countingReviewer.
func runCacheSteps(t *testing.T, ttl time.Duration, steps []cacheStep) {
	t.Helper()
	now := start()
	next := &countingReviewer{}
	cache, err := app.NewCachingTokenReviewer(next, ttl, func() time.Time { return now })
	if err != nil {
		t.Fatal(err)
	}
	for i, s := range steps {
		now = start().Add(s.at)
		username, err := cache.ReviewToken(t.Context(), s.token)
		if !errors.Is(err, s.wantErr) {
			t.Fatalf("step %d: ReviewToken() error = %v, want %v", i, err, s.wantErr)
		}
		if err == nil && username != "system:serviceaccount:openfief:plugin-sample" {
			t.Errorf("step %d: ReviewToken() = %q", i, username)
		}
		if next.calls != s.calls {
			t.Errorf("step %d: %d reviews, want %d", i, next.calls, s.calls)
		}
	}
}

// jwtExpiring is a synthetic, unsigned compact JWT whose exp claim is exp.
func jwtExpiring(exp time.Time) string {
	enc := base64.RawURLEncoding.EncodeToString
	return enc([]byte(`{"alg":"RS256"}`)) + "." +
		enc([]byte(`{"exp":`+strconv.FormatInt(exp.Unix(), 10)+`}`)) + ".c2lnbmF0dXJl"
}

func TestCachingTokenReviewer(t *testing.T) {
	t.Parallel()
	ttl, t0 := app.DefaultTokenCacheTTL, start()
	type step = cacheStep
	tests := []struct {
		name  string
		steps []step
	}{
		{
			name:  "accepted token is cached until the TTL",
			steps: []step{{token: sampleToken, calls: 1}, {token: sampleToken, at: ttl - time.Second, calls: 1}, {token: sampleToken, at: ttl, calls: 2}},
		},
		{
			name:  "distinct tokens are cached apart",
			steps: []step{{token: sampleToken, calls: 1}, {token: coreToken, calls: 2}, {token: sampleToken, calls: 2}},
		},
		{
			name: "refused token is never cached",
			steps: []step{
				{token: refusedToken, calls: 1, wantErr: app.ErrUnauthenticated},
				{token: refusedToken, calls: 2, wantErr: app.ErrUnauthenticated},
			},
		},
		{
			name: "token expiring before the TTL is cached until its expiry",
			steps: []step{
				{token: jwtExpiring(t0.Add(time.Minute)), calls: 1},
				{token: jwtExpiring(t0.Add(time.Minute)), at: time.Minute - time.Second, calls: 1},
				{token: jwtExpiring(t0.Add(time.Minute)), at: time.Minute, calls: 2},
			},
		},
		{
			name: "token expiring after the TTL is cached for the TTL",
			steps: []step{
				{token: jwtExpiring(t0.Add(time.Hour)), calls: 1},
				{token: jwtExpiring(t0.Add(time.Hour)), at: ttl, calls: 2},
			},
		},
		{
			name: "token already past its expiry is not cached",
			steps: []step{
				{token: jwtExpiring(t0.Add(-time.Second)), calls: 1},
				{token: jwtExpiring(t0.Add(-time.Second)), calls: 2},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			runCacheSteps(t, ttl, tt.steps)
		})
	}
}

// TestCachingTokenReviewerBounded fills the cache (1024 entries) and checks
// that a full cache stops caching new tokens until stale entries can go.
func TestCachingTokenReviewerBounded(t *testing.T) {
	t.Parallel()
	const size = 1024
	t0 := start()
	now := t0
	next := &countingReviewer{}
	cache, err := app.NewCachingTokenReviewer(next, time.Minute, func() time.Time { return now })
	if err != nil {
		t.Fatal(err)
	}
	review := func(token string) {
		t.Helper()
		if _, err := cache.ReviewToken(t.Context(), token); err != nil {
			t.Fatal(err)
		}
	}
	for i := range size {
		review("synthetic-token-" + strconv.Itoa(i))
	}
	review("synthetic-token-extra")
	review("synthetic-token-extra")
	if want := size + 2; next.calls != want {
		t.Errorf("full cache: %d reviews, want %d", next.calls, want)
	}
	now = t0.Add(time.Minute)
	review("synthetic-token-extra")
	review("synthetic-token-extra")
	if want := size + 3; next.calls != want {
		t.Errorf("after stale entries expire: %d reviews, want %d", next.calls, want)
	}
}

func TestNewCachingTokenReviewerTTL(t *testing.T) {
	t.Parallel()
	lifetime := time.Duration(plugins.TokenExpirationSeconds) * time.Second
	tests := []struct {
		ttl     time.Duration
		wantErr error
	}{
		{ttl: time.Second},
		{ttl: lifetime - time.Second},
		{ttl: 0, wantErr: app.ErrTokenCacheTTL},
		{ttl: -time.Second, wantErr: app.ErrTokenCacheTTL},
		{ttl: lifetime, wantErr: app.ErrTokenCacheTTL},
	}
	for _, tt := range tests {
		t.Run(tt.ttl.String(), func(t *testing.T) {
			t.Parallel()
			_, err := app.NewCachingTokenReviewer(&countingReviewer{}, tt.ttl, time.Now)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("NewCachingTokenReviewer(%v) error = %v, want %v", tt.ttl, err, tt.wantErr)
			}
		})
	}
}
