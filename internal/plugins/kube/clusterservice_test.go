package kube_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/arkadiuss/openfief/internal/plugins"
	"github.com/arkadiuss/openfief/internal/plugins/app"
	"github.com/arkadiuss/openfief/internal/plugins/kube"
	"github.com/arkadiuss/openfief/internal/plugins/operator/v1alpha1"
)

// gfService is a synthetic in-cluster Ghostfolio.
func gfService() plugins.ClusterService {
	return plugins.ClusterService{Namespace: "ghostfolio", Service: "ghostfolio", Port: 3333}
}

// gfPlan is an install of a view plugin that asks for an in-cluster service
// and the host of its config field url, with gfService picked.
func gfPlan(t *testing.T) (plugins.Plugin, plugins.Grant) {
	t.Helper()
	img, err := plugins.ParseImageRef(sampleImage)
	if err != nil {
		t.Fatalf("image: %v", err)
	}
	signed := plugins.SignedImage{Image: img, Manifest: plugins.Manifest{
		Capabilities:     []plugins.Capability{plugins.CapabilityView},
		EgressFromConfig: []string{"url"},
		ClusterService:   &plugins.ClusterServiceRequest{Description: "the owner's Ghostfolio"},
	}}
	cs := gfService()
	p, g, err := plugins.PlanInstall("gf", &signed, []byte(`{"url":"https://gf.example.com"}`), &cs)
	if err != nil {
		t.Fatalf("plan: %v", err)
	}
	return p, g
}

func gfSvc() *corev1.Service {
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{Namespace: "ghostfolio", Name: "ghostfolio"},
		Spec: corev1.ServiceSpec{Selector: map[string]string{"app": "ghostfolio"}, Ports: []corev1.ServicePort{
			{Name: "http", Port: 3333, TargetPort: intstr.FromInt32(3000)},
		}},
	}
}

func TestResourcesWithClusterService(t *testing.T) {
	t.Parallel()
	p, g := gfPlan(t)
	plugin, grant := kube.Resources(&p, &g)
	ref := v1alpha1.ClusterServiceRef{Namespace: "ghostfolio", Service: "ghostfolio", Port: 3333}
	if diff := cmp.Diff(&ref, plugin.Spec.ClusterService); diff != "" {
		t.Errorf("spec.clusterService (-want +got):\n%s", diff)
	}
	if diff := cmp.Diff([]string{"url"}, plugin.Spec.Manifest.EgressFromConfig); diff != "" {
		t.Errorf("spec.manifest.egressFromConfig (-want +got):\n%s", diff)
	}
	want := v1alpha1.Permissions{
		Capabilities: []string{"view"}, Egress: []string{"gf.example.com:443"},
		ClusterServices: []v1alpha1.ClusterServiceRef{ref},
	}
	if diff := cmp.Diff(want, grant.Spec); diff != "" {
		t.Errorf("grant (-want +got):\n%s", diff)
	}
	doc, err := kube.ResourcesYAML(&p, &g)
	if err != nil {
		t.Fatalf("render: %v", err)
	}
	for _, s := range []string{"kind: Role\n", "kind: RoleBinding\n", "namespace: ghostfolio\n", "- ghostfolio\n", "kind: Plugin\n"} {
		if !strings.Contains(string(doc), s) {
			t.Errorf("YAML lacks %q:\n%s", s, doc)
		}
	}
}

func TestServiceAccess(t *testing.T) {
	t.Parallel()
	role, binding := kube.ServiceAccess("gf", gfService())
	wantRules := []rbacv1.PolicyRule{{APIGroups: []string{""}, Resources: []string{"services"}, ResourceNames: []string{"ghostfolio"}, Verbs: []string{"get"}}}
	if role.Namespace != "ghostfolio" || role.Name != "openfief-plugin-gf" || !cmp.Equal(wantRules, role.Rules) {
		t.Errorf("role %s/%s %+v, want get on that one Service only", role.Namespace, role.Name, role.Rules)
	}
	wantSubjects := []rbacv1.Subject{{Kind: "ServiceAccount", Name: "openfief-operator", Namespace: "openfief"}}
	if !cmp.Equal(wantSubjects, binding.Subjects) || binding.RoleRef.Name != role.Name {
		t.Errorf("binding %+v to %+v, want the operator bound to the role", binding.Subjects, binding.RoleRef)
	}
}

func TestInstancesClusterService(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	policy := &networkingv1.NetworkPolicy{ObjectMeta: metav1.ObjectMeta{Namespace: "guarded", Name: "default-deny"}}
	guardedSvc := gfSvc()
	guardedSvc.Namespace = "guarded"
	store := fakeCluster(t, gfSvc(), guardedSvc, policy)
	info, err := store.ClusterService(ctx, gfService())
	if err != nil {
		t.Fatalf("ClusterService: %v", err)
	}
	want := app.ClusterServiceInfo{TargetPort: "3000", Selector: map[string]string{"app": "ghostfolio"}}
	if diff := cmp.Diff(want, info); diff != "" {
		t.Errorf("info (-want +got):\n%s", diff)
	}
	guarded := gfService()
	guarded.Namespace = "guarded"
	if got, guardedErr := store.ClusterService(ctx, guarded); guardedErr != nil || !got.NamespaceHasNetworkPolicies {
		t.Errorf("guarded: %+v, %v; want its NetworkPolicies seen", got, guardedErr)
	}
	other := gfService()
	other.Port = 80
	if _, portErr := store.ClusterService(ctx, other); !errors.Is(portErr, app.ErrClusterService) {
		t.Errorf("no such port: error = %v, want ErrClusterService", portErr)
	}
	missing := gfService()
	missing.Service = "absent"
	if _, missingErr := store.ClusterService(ctx, missing); !apierrors.IsNotFound(missingErr) {
		t.Errorf("no such service: error = %v, want not found", missingErr)
	}
}

func TestIngressRule(t *testing.T) {
	t.Parallel()
	cs := gfService()
	cs.Namespace = "guarded"
	info := app.ClusterServiceInfo{TargetPort: "3000", Selector: map[string]string{"app": "ghostfolio"}}
	rule := kube.IngressRule("gf", cs, &info)
	in := rule.Spec.Ingress[0]
	if rule.Namespace != "guarded" || in.Ports[0].Port.IntValue() != 3000 ||
		in.From[0].PodSelector.MatchLabels["openfief.io/plugin"] != "gf" ||
		in.From[0].NamespaceSelector.MatchLabels["kubernetes.io/metadata.name"] != "openfief" ||
		!cmp.Equal(info.Selector, rule.Spec.PodSelector.MatchLabels) {
		t.Errorf("ingress rule %+v, want the instance's pods from openfief to ghostfolio's on port 3000", rule.Spec)
	}
	doc, err := kube.IngressRuleYAML("gf", cs, &info)
	if err != nil || !strings.Contains(string(doc), "kind: NetworkPolicy") {
		t.Errorf("IngressRuleYAML: %s, %v", doc, err)
	}
}

func TestInstancesCreateDeleteServiceAccess(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	store := fakeCluster(t, gfSvc())
	p, g := gfPlan(t)
	if err := store.Create(ctx, &p, &g); err != nil {
		t.Fatalf("create: %v", err)
	}
	checkServiceAccess(t, store, true)
	all, err := store.List(ctx)
	if err != nil || len(all) != 1 || all[0].ClusterService == nil || *all[0].ClusterService != gfService() {
		t.Errorf("list = %+v, %v; want gf naming its in-cluster service", all, err)
	}
	if err := store.Delete(ctx, "gf", nil); err != nil {
		t.Fatalf("delete: %v", err)
	}
	checkServiceAccess(t, store, false)
}

// checkServiceAccess fails t unless the Role and RoleBinding that let the
// operator read gf's in-cluster service exist exactly when want.
func checkServiceAccess(t *testing.T, store *kube.Instances, want bool) {
	t.Helper()
	key := client.ObjectKey{Namespace: "ghostfolio", Name: "openfief-plugin-gf"}
	for _, obj := range []client.Object{&rbacv1.Role{}, &rbacv1.RoleBinding{}} {
		err := store.Client.Get(context.Background(), key, obj)
		if (err == nil) != want || (err != nil && !apierrors.IsNotFound(err)) {
			t.Errorf("%T: error = %v, want it to exist: %v", obj, err, want)
		}
	}
}
