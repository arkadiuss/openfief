package operator_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	authenticationv1 "k8s.io/api/authentication/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"

	"github.com/arkadiuss/openfief/internal/plugins"
	"github.com/arkadiuss/openfief/internal/plugins/operator"
	"github.com/arkadiuss/openfief/sdk/go/openfief"
)

// errSourceDown is a key source that can't be read.
var errSourceDown = errors.New("synthetic: API server unavailable")

type failingSource struct{}

func (failingSource) SigningKeys(context.Context) (plugins.SigningKeys, error) {
	return plugins.SigningKeys{}, errSourceDown
}

// discoveryAs returns IssuerDiscovery reading the API server as the given
// user and groups, impersonated with the envtest admin's credentials.
func (tc *testCluster) discoveryAs(t *testing.T, user string, groups ...string) operator.IssuerDiscovery {
	t.Helper()
	cfg := rest.CopyConfig(tc.cfg.Config)
	cfg.Impersonate = rest.ImpersonationConfig{UserName: user, Groups: groups}
	dc, err := discovery.NewDiscoveryClientForConfig(cfg)
	if err != nil {
		t.Fatalf("discovery client: %v", err)
	}
	return operator.IssuerDiscovery{Client: dc.RESTClient()}
}

// TestSigningKeys runs against envtest's API server, which has Kubernetes'
// default RBAC: it checks that the operator's ServiceAccount may read the
// issuer discovery endpoints without a ClusterRole of OpenFief's, publishes
// the keys, and verifies a TokenRequest token of the core with the SDK's
// JWKSVerifier from the published ConfigMap alone, as a plugin does.
func TestSigningKeys(t *testing.T) {
	t.Parallel()
	tc := startCluster(t)
	ctx := t.Context()
	asOperator := tc.discoveryAs(t, "system:serviceaccount:openfief:openfief-operator",
		"system:serviceaccounts", "system:serviceaccounts:openfief", "system:authenticated")
	t.Run("a user outside system:serviceaccounts is refused", func(t *testing.T) {
		t.Parallel()
		_, err := tc.discoveryAs(t, "synthetic-user", "system:authenticated").SigningKeys(ctx)
		if !apierrors.IsForbidden(err) {
			t.Errorf("SigningKeys as a plain user = %v, want Forbidden", err)
		}
	})
	p := &operator.SigningKeysPublisher{Client: tc.client, Source: asOperator, Refresh: 3 * time.Minute}
	res, err := p.Reconcile(ctx, ctrl.Request{})
	if err != nil {
		t.Fatalf("Reconcile as the operator's ServiceAccount: %v", err)
	}
	if res.RequeueAfter != 3*time.Minute {
		t.Errorf("RequeueAfter = %v, want the refresh interval", res.RequeueAfter)
	}
	cm := tc.signingKeys(ctx, t)
	if want := tc.cfg.Config.Host; strings.TrimSuffix(cm.Data["issuer"], "/") != strings.TrimSuffix(want, "/") {
		t.Errorf("issuer = %q, want the API server's (%s)", cm.Data["issuer"], want)
	}
	t.Run("the SDK verifies the core's token with the published keys", func(t *testing.T) {
		t.Parallel()
		tc.verifyCoreTokens(ctx, t, cm)
	})
	t.Run("a deleted ConfigMap is published again", func(t *testing.T) {
		t.Parallel()
		tc.republish(ctx, t, p)
	})
}

// republish deletes the ConfigMap, checks that a failing source leaves it
// absent and reports the error, and that the next good reconcile restores it.
func (tc *testCluster) republish(ctx context.Context, t *testing.T, p *operator.SigningKeysPublisher) {
	t.Helper()
	if err := tc.client.Delete(ctx, &corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{
		Name: "openfief-sa-jwks", Namespace: "openfief",
	}}); err != nil {
		t.Fatalf("delete: %v", err)
	}
	failing := &operator.SigningKeysPublisher{Client: tc.client, Source: failingSource{}}
	if _, err := failing.Reconcile(ctx, ctrl.Request{}); !errors.Is(err, errSourceDown) {
		t.Errorf("Reconcile with a failing source = %v, want %v", err, errSourceDown)
	}
	if _, err := p.Reconcile(ctx, ctrl.Request{}); err != nil {
		t.Fatalf("Reconcile: %v", err)
	}
	if cm := tc.signingKeys(ctx, t); cm.Labels["app.kubernetes.io/managed-by"] != "openfief-operator" {
		t.Errorf("labels = %v, want managed by the operator", cm.Labels)
	}
}

func (tc *testCluster) signingKeys(ctx context.Context, t *testing.T) *corev1.ConfigMap {
	t.Helper()
	var cm corev1.ConfigMap
	if err := tc.client.Get(ctx, types.NamespacedName{Namespace: "openfief", Name: "openfief-sa-jwks"}, &cm); err != nil {
		t.Fatalf("get signing keys ConfigMap: %v", err)
	}
	return &cm
}

// verifyCoreTokens requests tokens of the core's ServiceAccount through the
// TokenRequest API, as the core does, and checks them with a JWKSVerifier
// reading the ConfigMap's files.
func (tc *testCluster) verifyCoreTokens(ctx context.Context, t *testing.T, cm *corev1.ConfigMap) {
	t.Helper()
	dir := t.TempDir()
	for name, value := range cm.Data {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(value), 0o600); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
	}
	cs, err := kubernetes.NewForConfig(tc.cfg.Config)
	if err != nil {
		t.Fatalf("clientset: %v", err)
	}
	for _, sa := range []string{"openfief-core", "plugin-other"} {
		create(ctx, t, tc.client, &corev1.ServiceAccount{ObjectMeta: metav1.ObjectMeta{Name: sa, Namespace: "openfief"}})
	}
	token := func(sa, audience string) string {
		expiry := int64(600)
		tr, err := cs.CoreV1().ServiceAccounts("openfief").CreateToken(ctx, sa, &authenticationv1.TokenRequest{
			Spec: authenticationv1.TokenRequestSpec{Audiences: []string{audience}, ExpirationSeconds: &expiry},
		}, metav1.CreateOptions{})
		if err != nil {
			t.Fatalf("TokenRequest for %s: %v", sa, err)
		}
		return tr.Status.Token
	}
	v := &openfief.JWKSVerifier{
		Dir: dir, Audience: openfief.InstanceAudience("nbp"), Subject: "system:serviceaccount:openfief:openfief-core",
	}
	if err := v.VerifyCore(ctx, token("openfief-core", "openfief-plugin:nbp")); err != nil {
		t.Errorf("the core's token for nbp: %v", err)
	}
	refused := map[string]string{
		"the core's token for another instance": token("openfief-core", "openfief-plugin:other"),
		"another ServiceAccount's token":        token("plugin-other", "openfief-plugin:nbp"),
		"a plugin's own core token":             token("plugin-other", plugins.CoreAudience),
	}
	for name, tok := range refused {
		if err := v.VerifyCore(ctx, tok); !errors.Is(err, openfief.ErrNotCore) {
			t.Errorf("%s: VerifyCore = %v, want ErrNotCore", name, err)
		}
	}
}

// TestSigningKeysPublisherRuns runs the publisher in a manager: it publishes
// the keys once it starts, without any event from the cluster, and puts a
// deleted ConfigMap back.
func TestSigningKeysPublisherRuns(t *testing.T) {
	t.Parallel()
	tc := startCluster(t)
	mgr, err := ctrl.NewManager(tc.cfg.Config, ctrl.Options{
		Scheme:  tc.scheme,
		Cache:   cache.Options{DefaultNamespaces: map[string]cache.Config{"openfief": {}}},
		Metrics: metricsserver.Options{BindAddress: "0"},
	})
	if err != nil {
		t.Fatalf("NewManager: %v", err)
	}
	dc, err := discovery.NewDiscoveryClientForConfig(tc.cfg.Config)
	if err != nil {
		t.Fatalf("discovery client: %v", err)
	}
	p := &operator.SigningKeysPublisher{Client: mgr.GetClient(), Source: operator.IssuerDiscovery{Client: dc.RESTClient()}}
	if err := p.SetupWithManager(mgr); err != nil {
		t.Fatalf("SetupWithManager: %v", err)
	}
	ctx, cancel := context.WithCancel(t.Context())
	done := make(chan error, 1)
	go func() { done <- mgr.Start(ctx) }()
	t.Cleanup(func() {
		cancel()
		if err := <-done; err != nil {
			t.Errorf("manager: %v", err)
		}
	})
	tc.waitForSigningKeys(ctx, t)
	if err := tc.client.Delete(ctx, &corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{
		Name: "openfief-sa-jwks", Namespace: "openfief",
	}}); err != nil {
		t.Fatalf("delete: %v", err)
	}
	tc.waitForSigningKeys(ctx, t)
}

// waitForSigningKeys waits up to 30 seconds for the ConfigMap to hold keys.
func (tc *testCluster) waitForSigningKeys(ctx context.Context, t *testing.T) {
	t.Helper()
	deadline := time.Now().Add(30 * time.Second)
	for time.Now().Before(deadline) {
		var cm corev1.ConfigMap
		err := tc.client.Get(ctx, types.NamespacedName{Namespace: "openfief", Name: "openfief-sa-jwks"}, &cm)
		if err == nil && cm.DeletionTimestamp == nil && cm.Data["jwks.json"] != "" {
			return
		}
		time.Sleep(100 * time.Millisecond)
	}
	t.Fatal("the signing keys ConfigMap was not published within 30 seconds")
}
