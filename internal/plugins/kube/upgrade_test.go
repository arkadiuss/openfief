package kube_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
	rbacv1 "k8s.io/api/rbac/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/arkadiuss/openfief/internal/plugins"
	"github.com/arkadiuss/openfief/internal/plugins/app"
	"github.com/arkadiuss/openfief/internal/plugins/kube"
	"github.com/arkadiuss/openfief/internal/plugins/operator/v1alpha1"
)

// otherImage is a synthetic second version of sample.
const otherImage = "registry.example.com/example/sample@sha256:0f0f0f0f0f0f0f0f0f0f0f0f0f0f0f0f0f0f0f0f0f0f0f0f0f0f0f0f0f0f0f0f"

func TestInstancesGet(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	p, g := samplePlan(t, "sample")
	plugin, grant := kube.Resources(&p, &g)
	store := fakeCluster(t, plugin, grant)
	gotP, gotG, err := store.Get(ctx, "sample")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if diff := cmp.Diff(p, gotP); diff != "" {
		t.Errorf("plugin (-want +got):\n%s", diff)
	}
	if diff := cmp.Diff(g.Permissions, gotG.Permissions); diff != "" {
		t.Errorf("grant (-want +got):\n%s", diff)
	}
	gp, gg := gfPlan(t)
	gfPlugin, gfGrant := kube.Resources(&gp, &gg)
	gfStore := fakeCluster(t, gfPlugin, gfGrant)
	if got, _, gfErr := gfStore.Get(ctx, "gf"); gfErr != nil || got.ClusterService == nil || *got.ClusterService != gfService() {
		t.Errorf("Get gf = %+v, %v; want its in-cluster service", got.ClusterService, gfErr)
	}
	if _, _, missingErr := store.Get(ctx, "nope"); !errors.Is(missingErr, app.ErrNotInstalled) {
		t.Errorf("Get of a missing instance: %v, want ErrNotInstalled", missingErr)
	}
	onlyPlugin := fakeCluster(t, plugin)
	if _, _, noGrantErr := onlyPlugin.Get(ctx, "sample"); !errors.Is(noGrantErr, app.ErrNotInstalled) {
		t.Errorf("Get without a grant: %v, want ErrNotInstalled", noGrantErr)
	}
}

func TestInstancesUpdate(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	p, g := gfPlan(t)
	store := fakeCluster(t, gfSvc())
	if err := store.Create(ctx, &p, &g); err != nil {
		t.Fatalf("create: %v", err)
	}
	next := p
	next.Image = otherImage
	next.ClusterService = nil
	nextGrant := plugins.Grant{Plugin: "gf", Permissions: next.Requested()}
	if err := store.Update(ctx, &next, &nextGrant); err != nil {
		t.Fatalf("update: %v", err)
	}
	var plugin v1alpha1.Plugin
	var grant v1alpha1.PluginGrant
	key := client.ObjectKey{Namespace: plugins.Namespace, Name: "gf"}
	if err := store.Client.Get(ctx, key, &plugin); err != nil || plugin.Spec.Image != otherImage || plugin.Spec.ClusterService != nil {
		t.Errorf("plugin after update: %+v, %v; want the new image, no in-cluster service", plugin.Spec, err)
	}
	if err := store.Client.Get(ctx, key, &grant); err != nil || len(grant.Spec.ClusterServices) != 0 {
		t.Errorf("grant after update: %+v, %v; want no in-cluster service", grant.Spec, err)
	}
	// The access to the service it no longer names is taken back.
	access := client.ObjectKey{Namespace: "ghostfolio", Name: "openfief-plugin-gf"}
	if err := store.Client.Get(ctx, access, &rbacv1.Role{}); !apierrors.IsNotFound(err) {
		t.Errorf("Role after update: %v, want not found", err)
	}
	missing := next
	missing.Name = "nope"
	if err := store.Update(ctx, &missing, &plugins.Grant{Plugin: "nope"}); !errors.Is(err, app.ErrNotInstalled) {
		t.Errorf("Update of a missing instance: %v, want ErrNotInstalled", err)
	}
}
