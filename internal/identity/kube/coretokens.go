// Package kube is the identity context's Kubernetes adapter. CoreTokens
// requests the tokens the core presents to plugin instances through the
// TokenRequest API (docs/DESIGN.md D-59).
package kube

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	authnv1 "k8s.io/api/authentication/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/arkadiuss/openfief/sdk/go/openfief"
)

// ErrNoInstance reports a token request for an empty instance name.
var ErrNoInstance = errors.New("no plugin instance")

const (
	// TokenTTL is how long a core token is valid: the shortest the API
	// server issues.
	TokenTTL = 10 * time.Minute
	// refreshBefore is how long before its expiry a cached token is
	// replaced, so a call never carries a token about to expire.
	refreshBefore = 2 * time.Minute
)

// CoreTokens issues the core's tokens for plugin instances (D-59): a
// ServiceAccount token of the core's own ServiceAccount, whose audience is
// openfief.InstanceAudience of the instance, valid for TokenTTL. Each
// instance's token is cached until refreshBefore its expiry. The core's
// ServiceAccount needs create on its own serviceaccounts/token.
type CoreTokens struct {
	client         kubernetes.Interface
	namespace      string
	serviceAccount string
	now            func() time.Time

	mu    sync.Mutex
	cache map[string]cachedToken
}

// cachedToken is a token and when it expires.
type cachedToken struct {
	token   string
	expires time.Time
}

// NewCoreTokens returns the tokens of ServiceAccount serviceAccount in
// namespace, requested with client, with now as the clock.
func NewCoreTokens(client kubernetes.Interface, namespace, serviceAccount string, now func() time.Time) *CoreTokens {
	return &CoreTokens{client: client, namespace: namespace, serviceAccount: serviceAccount, now: now, cache: map[string]cachedToken{}}
}

// TokenFor returns a token for instance, from the cache while it is valid
// for more than refreshBefore, else a new one. Errors never contain a
// token.
func (c *CoreTokens) TokenFor(ctx context.Context, instance string) (string, error) {
	if instance == "" {
		return "", ErrNoInstance
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if t, ok := c.cache[instance]; ok && c.now().Add(refreshBefore).Before(t.expires) {
		return t.token, nil
	}
	ttl := int64(TokenTTL / time.Second)
	req := &authnv1.TokenRequest{Spec: authnv1.TokenRequestSpec{
		Audiences:         []string{openfief.InstanceAudience(instance)},
		ExpirationSeconds: &ttl,
	}}
	got, err := c.client.CoreV1().ServiceAccounts(c.namespace).CreateToken(ctx, c.serviceAccount, req, metav1.CreateOptions{})
	if err != nil {
		return "", fmt.Errorf("requesting a token for %s: %w", instance, err)
	}
	t := cachedToken{token: got.Status.Token, expires: got.Status.ExpirationTimestamp.Time}
	if t.token == "" {
		return "", fmt.Errorf("requesting a token for %s: %w", instance, errNoToken)
	}
	c.cache[instance] = t
	return t.token, nil
}

// errNoToken reports a TokenRequest answered without a token.
var errNoToken = errors.New("the API server returned no token")
