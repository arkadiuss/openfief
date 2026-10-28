package kube_test

import (
	"errors"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	authnv1 "k8s.io/api/authentication/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
	k8stesting "k8s.io/client-go/testing"

	"github.com/arkadiuss/openfief/internal/identity/kube"
)

var (
	errAPIDown   = errors.New("API server unavailable")
	errNotTokens = errors.New("reactor saw something other than a TokenRequest")
)

// tokenRequest is what the fake API server was asked for.
type tokenRequest struct {
	Namespace, ServiceAccount string
	Audiences                 []string
	ExpirationSeconds         int64
}

// tokenServer answers TokenRequests of the core's ServiceAccount with
// synthetic tokens "synthetic-<n>" that expire TTL after now, or fails
// with err, and records each request.
type tokenServer struct {
	mu    sync.Mutex
	asked []tokenRequest
	err   error
	empty bool
	now   time.Time
}

func (s *tokenServer) client() *fake.Clientset {
	client := fake.NewClientset()
	client.PrependReactor("create", "serviceaccounts", func(action k8stesting.Action) (bool, runtime.Object, error) {
		create, ok := action.(k8stesting.CreateAction)
		if !ok || action.GetSubresource() != "token" {
			return true, nil, errNotTokens
		}
		req, ok := create.GetObject().(*authnv1.TokenRequest)
		if !ok {
			return true, nil, errNotTokens
		}
		s.mu.Lock()
		defer s.mu.Unlock()
		name := ""
		if named, ok := action.(k8stesting.CreateActionImpl); ok {
			name = named.Name
		}
		s.asked = append(s.asked, tokenRequest{action.GetNamespace(), name, req.Spec.Audiences, *req.Spec.ExpirationSeconds})
		if s.err != nil {
			return true, nil, s.err
		}
		resp := &authnv1.TokenRequest{Status: authnv1.TokenRequestStatus{
			ExpirationTimestamp: metav1.NewTime(s.now.Add(kube.TokenTTL)),
		}}
		if !s.empty {
			resp.Status.Token = "synthetic-" + strconv.Itoa(len(s.asked))
		}
		return true, resp, nil
	})
	return client
}

// clock is a settable clock.
type clock struct{ now time.Time }

func (c *clock) Now() time.Time { return c.now }

func TestCoreTokensCaches(t *testing.T) {
	t.Parallel()
	c := &clock{now: time.Date(2026, time.March, 2, 6, 0, 0, 0, time.UTC)}
	srv := &tokenServer{now: c.now}
	tokens := kube.NewCoreTokens(srv.client(), "openfief", "openfief-core", c.Now)
	get := func(instance string) string {
		t.Helper()
		tok, err := tokens.TokenFor(t.Context(), instance)
		if err != nil {
			t.Fatalf("TokenFor(%s): %v", instance, err)
		}
		return tok
	}
	first := get("ibkr-main")
	c.now = c.now.Add(7*time.Minute + 59*time.Second) // still more than 2 minutes left
	if again := get("ibkr-main"); again != first {
		t.Errorf("token within its validity = %q, want the cached %q", again, first)
	}
	other := get("nbp")
	if other == first {
		t.Error("two instances share a token")
	}
	c.now = c.now.Add(time.Second) // 2 minutes left: replaced
	srv.now = c.now
	if renewed := get("ibkr-main"); renewed == first {
		t.Error("a token about to expire was not replaced")
	}
	want := []tokenRequest{
		{"openfief", "openfief-core", []string{"openfief-plugin:ibkr-main"}, 600},
		{"openfief", "openfief-core", []string{"openfief-plugin:nbp"}, 600},
		{"openfief", "openfief-core", []string{"openfief-plugin:ibkr-main"}, 600},
	}
	if diff := cmp.Diff(want, srv.asked); diff != "" {
		t.Errorf("token requests (-want +got):\n%s", diff)
	}
}

func TestCoreTokensFailures(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, time.March, 2, 6, 0, 0, 0, time.UTC)
	tests := []struct {
		name     string
		srv      *tokenServer
		instance string
		wantErr  error
		wantAsks int
	}{
		{name: "API server fails", srv: &tokenServer{err: errAPIDown, now: now}, instance: "ibkr-main", wantErr: errAPIDown, wantAsks: 2},
		{name: "no token in the answer", srv: &tokenServer{empty: true, now: now}, instance: "ibkr-main", wantAsks: 2},
		{name: "no instance", srv: &tokenServer{now: now}, wantErr: kube.ErrNoInstance},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tokens := kube.NewCoreTokens(tt.srv.client(), "openfief", "openfief-core", func() time.Time { return now })
			for range 2 { // a failure is not cached
				tok, err := tokens.TokenFor(t.Context(), tt.instance)
				if err == nil || tok != "" || (tt.wantErr != nil && !errors.Is(err, tt.wantErr)) {
					t.Errorf("TokenFor = %q, %v; want an error wrapping %v", tok, err, tt.wantErr)
				}
			}
			if len(tt.srv.asked) != tt.wantAsks {
				t.Errorf("asked %d times, want %d", len(tt.srv.asked), tt.wantAsks)
			}
		})
	}
}
