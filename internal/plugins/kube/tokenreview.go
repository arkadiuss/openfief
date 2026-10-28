// Package kube is the egress proxy's Kubernetes adapter: it verifies plugin
// tokens with the TokenReview API (D-63) and follows the egress allow-list
// ConfigMap the operator writes (D-32).
package kube

import (
	"context"
	"fmt"
	"slices"

	authnv1 "k8s.io/api/authentication/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/arkadiuss/openfief/internal/plugins/app"
)

// TokenReviewer is app.TokenReviewer over the TokenReview API, for one
// audience. The proxy's ServiceAccount needs permission to create
// TokenReviews.
type TokenReviewer struct {
	client   kubernetes.Interface
	audience string
}

// NewTokenReviewer returns a TokenReviewer that uses client and accepts only
// tokens issued for audience; the egress proxy passes plugins.EgressAudience.
func NewTokenReviewer(client kubernetes.Interface, audience string) *TokenReviewer {
	return &TokenReviewer{client: client, audience: audience}
}

// ReviewToken asks the API server whether token is valid for the reviewer's
// audience and returns its username. A token the API server does not
// authenticate, or authenticates without that audience, is refused with
// app.ErrUnauthenticated. Errors never contain the token.
func (r *TokenReviewer) ReviewToken(ctx context.Context, token string) (string, error) {
	review := &authnv1.TokenReview{Spec: authnv1.TokenReviewSpec{
		Token:     token,
		Audiences: []string{r.audience},
	}}
	got, err := r.client.AuthenticationV1().TokenReviews().Create(ctx, review, metav1.CreateOptions{})
	if err != nil {
		return "", fmt.Errorf("create TokenReview: %w", err)
	}
	switch {
	case !got.Status.Authenticated:
		return "", fmt.Errorf("%w: %s", app.ErrUnauthenticated, got.Status.Error)
	case !slices.Contains(got.Status.Audiences, r.audience):
		return "", fmt.Errorf("%w: audience is not %q", app.ErrUnauthenticated, r.audience)
	}
	return got.Status.User.Username, nil
}
