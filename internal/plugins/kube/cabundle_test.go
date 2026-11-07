package kube_test

import (
	"context"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	rbacv1 "k8s.io/api/rbac/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/arkadiuss/openfief/internal/plugins"
	"github.com/arkadiuss/openfief/internal/plugins/kube"
)

// caPlan is an install of instance rates, a connector whose manifest asks
// for the optional CA bundle Secret, expecting a bundle when extraCA.
func caPlan(t *testing.T, extraCA bool) (plugins.Plugin, plugins.Grant) {
	t.Helper()
	img, err := plugins.ParseImageRef(sampleImage)
	if err != nil {
		t.Fatalf("image: %v", err)
	}
	signed := plugins.SignedImage{Image: img, Manifest: plugins.Manifest{
		Capabilities: []plugins.Capability{plugins.CapabilityConnector},
		Secrets:      []plugins.SecretMount{{Role: plugins.CABundleRole, MountPath: "/etc/openfief/ca-bundle", Optional: true}},
	}}
	p, g, err := plugins.PlanInstall("rates", &signed, nil, nil)
	if err != nil {
		t.Fatalf("plan: %v", err)
	}
	if err := p.SetExtraCA(extraCA); err != nil {
		t.Fatalf("SetExtraCA: %v", err)
	}
	return p, g
}

func TestCABundleAccess(t *testing.T) {
	t.Parallel()
	role, binding := kube.CABundleAccess("rates")
	wantRules := []rbacv1.PolicyRule{{APIGroups: []string{""}, Resources: []string{"secrets"}, ResourceNames: []string{"rates--ca-bundle"}, Verbs: []string{"get"}}}
	if role.Namespace != "openfief" || role.Name != "openfief-plugin-rates-ca-bundle" || !cmp.Equal(wantRules, role.Rules) {
		t.Errorf("role %s/%s %+v, want get on that one Secret only", role.Namespace, role.Name, role.Rules)
	}
	wantSubjects := []rbacv1.Subject{{Kind: "ServiceAccount", Name: "openfief-operator", Namespace: "openfief"}}
	if !cmp.Equal(wantSubjects, binding.Subjects) || binding.RoleRef.Name != role.Name {
		t.Errorf("binding %+v to %+v, want the operator bound to the role", binding.Subjects, binding.RoleRef)
	}
}

// TestInstancesCABundleAccess checks that installing an instance that
// expects a CA bundle lets the operator get that Secret, that an upgrade
// that no longer expects one takes that back and one that does gives it
// again, that uninstalling removes it, that an instance expecting none
// never gets it, and that --print's YAML holds it exactly when expected.
func TestInstancesCABundleAccess(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	store := fakeCluster(t)
	p, g := caPlan(t, true)
	if err := store.Create(ctx, &p, &g); err != nil {
		t.Fatalf("create: %v", err)
	}
	checkCABundleAccess(t, store, "after install", true)
	off, _ := caPlan(t, false)
	if err := store.Update(ctx, &off, &g); err != nil {
		t.Fatalf("update: %v", err)
	}
	checkCABundleAccess(t, store, "after an upgrade expecting none", false)
	if err := store.Update(ctx, &p, &g); err != nil {
		t.Fatalf("update: %v", err)
	}
	checkCABundleAccess(t, store, "after an upgrade expecting one again", true)
	if err := store.Delete(ctx, "rates", nil); err != nil {
		t.Fatalf("delete: %v", err)
	}
	checkCABundleAccess(t, store, "after uninstall", false)
	if err := store.Create(ctx, &off, &g); err != nil {
		t.Fatalf("create: %v", err)
	}
	checkCABundleAccess(t, store, "after an install expecting none", false)

	for _, tt := range []struct {
		plugin *plugins.Plugin
		want   bool
	}{{&p, true}, {&off, false}} {
		out, err := kube.ResourcesYAML(tt.plugin, &g)
		if err != nil {
			t.Fatalf("ResourcesYAML: %v", err)
		}
		if got := strings.Contains(string(out), "name: openfief-plugin-rates-ca-bundle"); got != tt.want {
			t.Errorf("YAML with extraCA %v holds the CA bundle Role: %v, want %v\n%s", tt.plugin.ExtraCA, got, tt.want, out)
		}
	}
}

// checkCABundleAccess fails t unless the Role and RoleBinding that let the
// operator get rates's CA bundle Secret exist exactly when want.
func checkCABundleAccess(t *testing.T, store *kube.Instances, step string, want bool) {
	t.Helper()
	key := client.ObjectKey{Namespace: "openfief", Name: "openfief-plugin-rates-ca-bundle"}
	for _, obj := range []client.Object{&rbacv1.Role{}, &rbacv1.RoleBinding{}} {
		err := store.Client.Get(context.Background(), key, obj)
		if (err == nil) != want || (err != nil && !apierrors.IsNotFound(err)) {
			t.Errorf("%s: %T: error = %v, want it to exist: %v", step, obj, err, want)
		}
	}
}
