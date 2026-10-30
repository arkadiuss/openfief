package operator_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/arkadiuss/openfief/internal/plugins"
	"github.com/arkadiuss/openfief/internal/plugins/operator"
	"github.com/arkadiuss/openfief/internal/plugins/operator/v1alpha1"
)

// ghostfolioSelector selects a synthetic in-cluster Ghostfolio's pods.
func ghostfolioSelector() map[string]string { return map[string]string{"app": "ghostfolio"} }

// ghostfolioSvc is a synthetic in-cluster Ghostfolio's Service in
// namespace ghostfolio: port 3333 to pod port 3000.
func ghostfolioSvc(target intstr.IntOrString) *corev1.Service {
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{Name: "ghostfolio", Namespace: "ghostfolio"},
		Spec: corev1.ServiceSpec{
			ClusterIP: "192.0.2.30", Selector: ghostfolioSelector(),
			Ports: []corev1.ServicePort{{Name: "http", Port: 3333, Protocol: corev1.ProtocolTCP, TargetPort: target}},
		},
	}
}

func TestClusterWithClusterService(t *testing.T) {
	t.Parallel()
	headless := ghostfolioSvc(intstr.FromInt32(3000))
	headless.Spec.ClusterIP = "None"
	noSelector := ghostfolioSvc(intstr.FromInt32(3000))
	noSelector.Spec.Selector = nil
	tests := []struct {
		name    string
		svc     *corev1.Service
		port    uint16
		wantErr error
	}{
		{name: "usable", svc: ghostfolioSvc(intstr.FromInt32(3000)), port: 3333},
		{name: "headless", svc: headless, port: 3333, wantErr: operator.ErrServiceUnusable},
		{name: "no selector", svc: noSelector, port: 3333, wantErr: operator.ErrServiceUnusable},
		{name: "no such port", svc: ghostfolioSvc(intstr.FromInt32(3000)), port: 80, wantErr: operator.ErrServiceUnusable},
		{name: "named target port", svc: ghostfolioSvc(intstr.FromString("http")), port: 3333, wantErr: operator.ErrServiceUnusable},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			c := cluster()
			err := c.AddClusterService(tt.svc, tt.port)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("AddClusterService error = %v, want %v", err, tt.wantErr)
			}
			if err != nil {
				return
			}
			want := &plugins.Service{Address: "192.0.2.30", Port: 3333, TargetPort: 3000}
			if diff := cmp.Diff(want, c.Topology.ClusterService); diff != "" {
				t.Errorf("Topology.ClusterService (-want +got):\n%s", diff)
			}
			if c.ClusterServiceNamespace != "ghostfolio" || !cmp.Equal(c.ClusterServiceSelector, ghostfolioSelector()) {
				t.Errorf("selector %s %v, want ghostfolio's pods", c.ClusterServiceNamespace, c.ClusterServiceSelector)
			}
		})
	}
}

// inClusterPlugin is a view plugin that asks for an in-cluster service and
// names ghostfolio/<service>:3333.
func inClusterPlugin(name, service string) *v1alpha1.Plugin {
	return &v1alpha1.Plugin{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: "openfief"},
		Spec: v1alpha1.PluginSpec{
			Image: imageA,
			Manifest: v1alpha1.Manifest{
				Capabilities:   []string{"view"},
				ClusterService: &v1alpha1.ClusterServiceRequest{Description: "the owner's Ghostfolio"},
			},
			ClusterService: &v1alpha1.ClusterServiceRef{Namespace: "ghostfolio", Service: service, Port: 3333},
		},
	}
}

// inClusterGrant approves the view capability and, when service is not
// empty, ghostfolio/<service>:3333.
func inClusterGrant(name, service string) *v1alpha1.PluginGrant {
	g := &v1alpha1.PluginGrant{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: "openfief"},
		Spec:       v1alpha1.Permissions{Capabilities: []string{"view"}},
	}
	if service != "" {
		g.Spec.ClusterServices = []v1alpha1.ClusterServiceRef{{Namespace: "ghostfolio", Service: service, Port: 3333}}
	}
	return g
}

func TestReconcileClusterService(t *testing.T) {
	t.Parallel()
	tc := startCluster(t)
	ctx := t.Context()
	create(ctx, t, tc.client, &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "ghostfolio"}})
	svc := ghostfolioSvc(intstr.FromInt32(3000))
	svc.Spec.ClusterIP = "" // the API server assigns it
	create(ctx, t, tc.client, svc)
	r := tc.reconciler(signedImages{imageA})
	t.Run("granted: a NetworkPolicy rule to its pods and its ClusterIP", func(t *testing.T) {
		t.Parallel()
		testClusterServiceGranted(t, tc, r)
	})
	t.Run("not approved: waits", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()
		create(ctx, t, tc.client, inClusterPlugin("gf-pending", "ghostfolio"))
		create(ctx, t, tc.client, inClusterGrant("gf-pending", ""))
		tc.reconcile(ctx, t, r, "gf-pending")
		s := tc.state(ctx, t, "gf-pending")
		want := &v1alpha1.Permissions{ClusterServices: []v1alpha1.ClusterServiceRef{{Namespace: "ghostfolio", Service: "ghostfolio", Port: 3333}}}
		if !s.DenyAll || !cmp.Equal(want, s.Pending) {
			t.Errorf("state %+v, want deny-all with the service pending", s)
		}
	})
	t.Run("missing service: nothing runs, checked again", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()
		create(ctx, t, tc.client, inClusterPlugin("gf-missing", "absent"))
		create(ctx, t, tc.client, inClusterGrant("gf-missing", "absent"))
		res := reconcileResult(ctx, t, r, "gf-missing")
		s := tc.state(ctx, t, "gf-missing")
		if !s.DenyAll || s.Conditions["Ready"] != "False/ClusterUnavailable" || res.RequeueAfter == 0 {
			t.Errorf("state %+v, requeue %s; want deny-all, ClusterUnavailable and a retry", s, res.RequeueAfter)
		}
	})
}

// testClusterServiceGranted runs a plugin whose grant approves its
// in-cluster service: its NetworkPolicy reaches that service's pods on the
// pod port, it gets the ClusterIP, and nothing goes through the proxy.
func testClusterServiceGranted(t *testing.T, tc *testCluster, r *operator.PluginReconciler) {
	t.Helper()
	ctx := t.Context()
	create(ctx, t, tc.client, inClusterPlugin("gf", "ghostfolio"))
	create(ctx, t, tc.client, inClusterGrant("gf", "ghostfolio"))
	tc.reconcile(ctx, t, r, "gf")
	var np networkingv1.NetworkPolicy
	get(ctx, t, tc, "plugin-gf", &np)
	tcp, port := corev1.ProtocolTCP, intstr.FromInt32(3000)
	wantRule := networkingv1.NetworkPolicyEgressRule{
		To: []networkingv1.NetworkPolicyPeer{{
			NamespaceSelector: &metav1.LabelSelector{MatchLabels: map[string]string{"kubernetes.io/metadata.name": "ghostfolio"}},
			PodSelector:       &metav1.LabelSelector{MatchLabels: ghostfolioSelector()},
		}},
		Ports: []networkingv1.NetworkPolicyPort{{Protocol: &tcp, Port: &port}},
	}
	if n := len(np.Spec.Egress); n != 2 || !cmp.Equal(wantRule, np.Spec.Egress[1]) {
		t.Errorf("egress rules = %+v, want the core's and %+v", np.Spec.Egress, wantRule)
	}
	var created corev1.Service
	if err := tc.client.Get(ctx, types.NamespacedName{Namespace: "ghostfolio", Name: "ghostfolio"}, &created); err != nil {
		t.Fatalf("get service: %v", err)
	}
	env := podEnv(ctx, t, tc, "plugin-gf")
	if got := env["OPENFIEF_CLUSTER_SERVICE_ADDRESS"]; got != created.Spec.ClusterIP+":3333" {
		t.Errorf("OPENFIEF_CLUSTER_SERVICE_ADDRESS = %q, want %s:3333", got, created.Spec.ClusterIP)
	}
	if _, ok := env["HTTPS_PROXY"]; ok {
		t.Errorf("a plugin with only an in-cluster service got HTTPS_PROXY")
	}
	if s := tc.state(ctx, t, "gf"); s.AllowList != "" {
		t.Errorf("allow-list %q, want none: the in-cluster service is not reached through the proxy", s.AllowList)
	}
}

// podEnv reads the named Deployment's container environment.
func podEnv(ctx context.Context, t *testing.T, tc *testCluster, name string) map[string]string {
	t.Helper()
	var d appsv1.Deployment
	get(ctx, t, tc, name, &d)
	env := map[string]string{}
	for _, e := range d.Spec.Template.Spec.Containers[0].Env {
		env[e.Name] = e.Value
	}
	return env
}

// get reads one object of the openfief namespace.
func get(ctx context.Context, t *testing.T, tc *testCluster, name string, obj client.Object) {
	t.Helper()
	if err := tc.client.Get(ctx, types.NamespacedName{Namespace: "openfief", Name: name}, obj); err != nil {
		t.Fatalf("get %s: %v", name, err)
	}
}

// reconcileResult runs one reconcile of the named Plugin and returns its
// result.
func reconcileResult(ctx context.Context, t *testing.T, r *operator.PluginReconciler, name string) ctrl.Result {
	t.Helper()
	res, err := r.Reconcile(ctx, ctrl.Request{NamespacedName: types.NamespacedName{Namespace: "openfief", Name: name}})
	if err != nil {
		t.Fatalf("Reconcile(%s): %v", name, err)
	}
	return res
}
