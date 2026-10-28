package kube_test

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	authnv1 "k8s.io/api/authentication/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
	k8stesting "k8s.io/client-go/testing"

	"github.com/arkadiuss/openfief/internal/plugins"
	"github.com/arkadiuss/openfief/internal/plugins/app"
	"github.com/arkadiuss/openfief/internal/plugins/kube"
)

// Synthetic token; not a real credential.
const sampleToken = "synthetic-token-sample"

// sampleUsername is plugin sample's ServiceAccount username.
const sampleUsername = "system:serviceaccount:openfief:plugin-sample"

var (
	errAPIDown    = errors.New("API server unavailable")
	errNotAReview = errors.New("reactor saw something other than a TokenReview")
)

// reviewClient answers TokenReviews with status, or fails with apiErr, and
// records the audiences each review asked for.
func reviewClient(status *authnv1.TokenReviewStatus, apiErr error, asked *[]string) *fake.Clientset {
	client := fake.NewClientset()
	client.PrependReactor("create", "tokenreviews", func(action k8stesting.Action) (bool, runtime.Object, error) {
		create, ok := action.(k8stesting.CreateAction)
		if !ok {
			return true, nil, errNotAReview
		}
		review, ok := create.GetObject().(*authnv1.TokenReview)
		if !ok {
			return true, nil, errNotAReview
		}
		*asked = append(*asked, review.Spec.Audiences...)
		if apiErr != nil {
			return true, nil, apiErr
		}
		return true, &authnv1.TokenReview{Status: *status}, nil
	})
	return client
}

func TestTokenReviewerReviewToken(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		status  authnv1.TokenReviewStatus
		apiErr  error
		want    string
		wantErr error
	}{
		{
			name:   "authenticated for the core audience",
			status: authnv1.TokenReviewStatus{Authenticated: true, Audiences: []string{plugins.CoreAudience}, User: authnv1.UserInfo{Username: sampleUsername}},
			want:   sampleUsername,
		},
		{
			name:    "not authenticated",
			status:  authnv1.TokenReviewStatus{Error: "token expired"},
			wantErr: app.ErrUnauthenticated,
		},
		{
			name:    "authenticated for another audience",
			status:  authnv1.TokenReviewStatus{Authenticated: true, Audiences: []string{"https://kubernetes.default.svc"}, User: authnv1.UserInfo{Username: sampleUsername}},
			wantErr: app.ErrUnauthenticated,
		},
		{name: "API error", apiErr: errAPIDown, wantErr: errAPIDown},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var asked []string
			reviewer := kube.NewTokenReviewer(reviewClient(&tt.status, tt.apiErr, &asked))
			got, err := reviewer.ReviewToken(t.Context(), sampleToken)
			if diff := cmp.Diff(tt.wantErr, err, cmpopts.EquateErrors()); diff != "" {
				t.Fatalf("ReviewToken() error (-want +got):\n%s", diff)
			}
			if got != tt.want {
				t.Errorf("ReviewToken() = %q, want %q", got, tt.want)
			}
			if err != nil && strings.Contains(err.Error(), sampleToken) {
				t.Errorf("ReviewToken() error %q contains the token", err)
			}
			if diff := cmp.Diff([]string{plugins.CoreAudience}, asked); diff != "" {
				t.Errorf("TokenReview audiences (-want +got):\n%s", diff)
			}
		})
	}
}

// allowListMap builds the operator's ConfigMap with the given data.
func allowListMap(name string, data map[string]string) *corev1.ConfigMap {
	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: plugins.Namespace},
		Data:       data,
	}
}

// errorLog collects what WatchAllowLists reports.
type errorLog struct {
	mu   sync.Mutex
	errs []error
}

func (l *errorLog) report(err error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.errs = append(l.errs, err)
}

func (l *errorLog) count() int {
	l.mu.Lock()
	defer l.mu.Unlock()
	return len(l.errs)
}

// waitFor polls cond until it holds, failing the test after a deadline.
func waitFor(t *testing.T, what string, cond func() bool) {
	t.Helper()
	ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
	defer cancel()
	tick := time.NewTicker(10 * time.Millisecond)
	defer tick.Stop()
	for !cond() {
		select {
		case <-ctx.Done():
			t.Fatalf("timed out waiting for %s", what)
		case <-tick.C:
		}
	}
}

// allows reports whether plugin sample's current list allows host:443.
func allows(a *kube.AllowLists, host string) bool {
	l, ok := a.AllowList("sample")
	return ok && l.Allows(host, plugins.DefaultEgressPort)
}

func TestWatchAllowListsFollowsConfigMap(t *testing.T) {
	t.Parallel()
	client := fake.NewClientset(allowListMap(plugins.EgressAllowListConfigMap, map[string]string{"sample": "api.example.com:443\n"}))
	log := &errorLog{}
	lists, err := kube.WatchAllowLists(t.Context(), client, log.report)
	if err != nil {
		t.Fatal(err)
	}
	if !allows(lists, "api.example.com") {
		t.Fatal("initial ConfigMap: sample may not reach api.example.com")
	}
	if _, ok := lists.AllowList("other"); ok {
		t.Error("a plugin not in the ConfigMap has an allow-list")
	}

	cms := client.CoreV1().ConfigMaps(plugins.Namespace)
	update := func(data map[string]string) {
		t.Helper()
		if _, err := cms.Update(t.Context(), allowListMap(plugins.EgressAllowListConfigMap, data), metav1.UpdateOptions{}); err != nil {
			t.Fatal(err)
		}
	}

	update(map[string]string{"sample": "data.example.org:443\n"})
	waitFor(t, "the updated list", func() bool {
		return allows(lists, "data.example.org") && !allows(lists, "api.example.com")
	})

	update(map[string]string{"sample": "API.example.com\n"})
	waitFor(t, "an undecodable ConfigMap to drop every list", func() bool {
		_, ok := lists.AllowList("sample")
		return !ok && log.count() == 1
	})

	update(map[string]string{"sample": "api.example.com:443\n"})
	waitFor(t, "the list to come back", func() bool { return allows(lists, "api.example.com") })

	if _, err := cms.Create(t.Context(), allowListMap("unrelated", map[string]string{"evil": "evil.example.net:443\n"}), metav1.CreateOptions{}); err != nil {
		t.Fatal(err)
	}
	if err := cms.Delete(t.Context(), plugins.EgressAllowListConfigMap, metav1.DeleteOptions{}); err != nil {
		t.Fatal(err)
	}
	waitFor(t, "a deleted ConfigMap to drop every list", func() bool {
		_, ok := lists.AllowList("sample")
		return !ok
	})
	if _, ok := lists.AllowList("evil"); ok {
		t.Error("a list from another ConfigMap was used")
	}
}

func TestWatchAllowListsWithoutConfigMap(t *testing.T) {
	t.Parallel()
	lists, err := kube.WatchAllowLists(t.Context(), fake.NewClientset(), func(error) {})
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := lists.AllowList("sample"); ok {
		t.Error("with no ConfigMap, a plugin has an allow-list")
	}
}

func TestWatchAllowListsNotSynced(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	_, err := kube.WatchAllowLists(ctx, fake.NewClientset(), func(error) {})
	if !errors.Is(err, kube.ErrNotSynced) {
		t.Errorf("WatchAllowLists() error = %v, want %v", err, kube.ErrNotSynced)
	}
}
