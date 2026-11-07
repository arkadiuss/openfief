package kube_test

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	k8sfake "k8s.io/client-go/kubernetes/fake"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/arkadiuss/openfief/internal/plugins"
	"github.com/arkadiuss/openfief/internal/plugins/kube"
	"github.com/arkadiuss/openfief/internal/plugins/operator/v1alpha1"
)

// auditPod is a pod in ns labelled app.kubernetes.io/name=name.
func auditPod(ns, name, app string) *corev1.Pod {
	return &corev1.Pod{ObjectMeta: metav1.ObjectMeta{
		Namespace: ns, Name: name, Labels: map[string]string{"app.kubernetes.io/name": app},
	}}
}

// TestEgressProxyLog checks that the log is read from every egress proxy
// pod in the openfief namespace, each ending in a newline, and from no
// other pod. client-go's fake answers every log request with "fake logs".
func TestEgressProxyLog(t *testing.T) {
	t.Parallel()
	clients := k8sfake.NewClientset(
		auditPod(plugins.Namespace, "egress-proxy-a", "egress-proxy"),
		auditPod(plugins.Namespace, "egress-proxy-b", "egress-proxy"),
		auditPod(plugins.Namespace, "openfief-core-a", "openfief-core"),
		auditPod("elsewhere", "egress-proxy-c", "egress-proxy"),
	)
	got, err := kube.EgressProxyLog(t.Context(), clients)
	if err != nil {
		t.Fatalf("EgressProxyLog: %v", err)
	}
	if want := "fake logs\nfake logs\n"; string(got) != want {
		t.Errorf("EgressProxyLog() = %q, want %q", got, want)
	}
}

// grantedEgressOf runs GrantedEgress on a controller-runtime fake holding
// objs.
func grantedEgressOf(t *testing.T, objs ...client.Object) (map[string][]plugins.EgressHost, error) {
	t.Helper()
	scheme := runtime.NewScheme()
	if err := v1alpha1.AddToScheme(scheme); err != nil {
		t.Fatalf("scheme: %v", err)
	}
	granted, err := kube.GrantedEgress(t.Context(), fake.NewClientBuilder().WithScheme(scheme).WithObjects(objs...).Build())
	if err != nil {
		return nil, fmt.Errorf("granted egress: %w", err)
	}
	return granted, nil
}

// auditGrant is a PluginGrant in ns approving egress.
func auditGrant(ns, name string, egress ...string) *v1alpha1.PluginGrant {
	return &v1alpha1.PluginGrant{
		ObjectMeta: metav1.ObjectMeta{Namespace: ns, Name: name},
		Spec:       v1alpha1.Permissions{Egress: egress},
	}
}

func TestGrantedEgress(t *testing.T) {
	t.Parallel()
	got, err := grantedEgressOf(t,
		auditGrant(plugins.Namespace, "nbp", "api.nbp.pl"),
		auditGrant(plugins.Namespace, "fixture", "fixture.example.org:8443", "example.com"),
		auditGrant(plugins.Namespace, "ui"),
		auditGrant("elsewhere", "other", "example.net"),
	)
	if err != nil {
		t.Fatalf("GrantedEgress: %v", err)
	}
	want := map[string][]plugins.EgressHost{
		"nbp":     {{Host: "api.nbp.pl", Port: 443}},
		"fixture": {{Host: "fixture.example.org", Port: 8443}, {Host: "example.com", Port: 443}},
		"ui":      {},
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("GrantedEgress() mismatch (-want +got):\n%s", diff)
	}
}

func TestGrantedEgressRefusesABadHost(t *testing.T) {
	t.Parallel()
	if _, err := grantedEgressOf(t, auditGrant(plugins.Namespace, "bad", "*.example.com")); err == nil {
		t.Error("GrantedEgress accepted a wildcard host, want an error")
	}
}
