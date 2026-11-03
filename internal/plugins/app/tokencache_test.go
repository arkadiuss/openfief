package app_test

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/arkadiuss/openfief/internal/plugins"
	"github.com/arkadiuss/openfief/internal/plugins/app"
)

// countingReviewer accepts every token except jwtNamed("refused") as plugin
// "sample", fails to review jwtNamed("unreachable") as an API server that
// doesn't answer would, and counts the reviews it was asked for.
type countingReviewer struct{ calls int }

// errUnreachable is a synthetic failure to ask the API server.
var errUnreachable = errors.New("synthetic: the API server didn't answer")

func (c *countingReviewer) ReviewToken(_ context.Context, token string) (string, error) {
	c.calls++
	switch token {
	case jwtNamed("refused"):
		return "", fmt.Errorf("%w: refused by the fake", app.ErrUnauthenticated)
	case jwtNamed("unreachable"):
		return "", errUnreachable
	}
	return "system:serviceaccount:openfief:plugin-sample", nil
}

// jwtNamed is a synthetic, unsigned compact JWT whose subject is name, as a
// projected ServiceAccount token is shaped.
func jwtNamed(name string) string {
	enc := base64.RawURLEncoding.EncodeToString
	return enc([]byte(`{"alg":"RS256"}`)) + "." + enc([]byte(`{"sub":"`+name+`"}`)) + ".c2lnbmF0dXJl"
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
	sample, refused := jwtNamed("sample"), jwtNamed("refused")
	type step = cacheStep
	tests := []struct {
		name  string
		steps []step
	}{
		{
			name:  "accepted token is cached until the TTL",
			steps: []step{{token: sample, calls: 1}, {token: sample, at: ttl - time.Second, calls: 1}, {token: sample, at: ttl, calls: 2}},
		},
		{
			name:  "distinct tokens are cached apart",
			steps: []step{{token: sample, calls: 1}, {token: jwtNamed("core"), calls: 2}, {token: sample, calls: 2}},
		},
		{
			name: "refused token stays refused briefly without asking again (D-184)",
			steps: []step{
				{token: refused, calls: 1, wantErr: app.ErrUnauthenticated},
				{token: refused, at: app.RefusalCacheTTL - time.Second, calls: 1, wantErr: app.ErrUnauthenticated},
				{token: refused, at: app.RefusalCacheTTL, calls: 2, wantErr: app.ErrUnauthenticated},
			},
		},
		{
			name: "a review that failed is never cached",
			steps: []step{
				{token: jwtNamed("unreachable"), calls: 1, wantErr: errUnreachable},
				{token: jwtNamed("unreachable"), calls: 2, wantErr: errUnreachable},
			},
		},
		{
			name: "a token not shaped like a JWT is refused without asking (D-184)",
			steps: []step{
				{token: "ofsess_synthetic-session", calls: 0, wantErr: app.ErrUnauthenticated},
				{token: "a.b.c", calls: 0, wantErr: app.ErrUnauthenticated},
				{token: sample, calls: 1},
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
		review(jwtNamed("synthetic-" + strconv.Itoa(i)))
	}
	review(jwtNamed("extra"))
	review(jwtNamed("extra"))
	if want := size + 2; next.calls != want {
		t.Errorf("full cache: %d reviews, want %d", next.calls, want)
	}
	now = t0.Add(time.Minute)
	review(jwtNamed("extra"))
	review(jwtNamed("extra"))
	if want := size + 3; next.calls != want {
		t.Errorf("after stale entries expire: %d reviews, want %d", next.calls, want)
	}
}

// TestCachingTokenReviewerRefusalsBounded fills the refusals (1024) and
// checks that accepted tokens are still cached: refusals never crowd them
// out (D-184).
func TestCachingTokenReviewerRefusalsBounded(t *testing.T) {
	t.Parallel()
	const size = 1024
	now := start()
	next := &refusingReviewer{}
	cache, err := app.NewCachingTokenReviewer(next, time.Minute, func() time.Time { return now })
	if err != nil {
		t.Fatal(err)
	}
	for i := range size + 1 {
		if _, err := cache.ReviewToken(t.Context(), jwtNamed("refused-"+strconv.Itoa(i))); !errors.Is(err, app.ErrUnauthenticated) {
			t.Fatalf("refusal %d: error = %v", i, err)
		}
	}
	for range 2 {
		if _, err := cache.ReviewToken(t.Context(), jwtNamed("sample")); err != nil {
			t.Fatal(err)
		}
	}
	if want := size + 2; next.calls != want {
		t.Errorf("%d reviews, want %d: the accepted token is reviewed once", next.calls, want)
	}
}

// refusingReviewer refuses every token whose subject starts with
// "refused", accepts the rest as plugin "sample", and counts reviews.
type refusingReviewer struct{ calls int }

func (r *refusingReviewer) ReviewToken(_ context.Context, token string) (string, error) {
	r.calls++
	if token != jwtNamed("sample") {
		return "", fmt.Errorf("%w: refused by the fake", app.ErrUnauthenticated)
	}
	return "system:serviceaccount:openfief:plugin-sample", nil
}

func TestJWTShaped(t *testing.T) {
	t.Parallel()
	enc := base64.RawURLEncoding.EncodeToString
	tests := []struct {
		name string
		in   string
		want bool
	}{
		{name: "a compact JWT", in: jwtNamed("sample"), want: true},
		{name: "a session token", in: "ofsess_synthetic"},
		{name: "a script token", in: "oftok_synthetic"},
		{name: "empty"},
		{name: "two parts", in: enc([]byte(`{"alg":"RS256"}`)) + "." + enc([]byte(`{}`))},
		{name: "no signature", in: enc([]byte(`{"alg":"RS256"}`)) + "." + enc([]byte(`{}`)) + "."},
		{name: "not base64url", in: "a+b.c/d.e=f"},
		{name: "claims not JSON", in: enc([]byte(`{"alg":"RS256"}`)) + "." + enc([]byte("not json")) + ".c2ln"},
		{name: "claims a JSON array", in: enc([]byte(`{"alg":"RS256"}`)) + "." + enc([]byte(`[1]`)) + ".c2ln"},
		{name: "too long", in: jwtNamed(strings.Repeat("x", app.MaxTokenBytes))},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := app.JWTShaped(tt.in); got != tt.want {
				t.Errorf("JWTShaped = %v, want %v", got, tt.want)
			}
		})
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
