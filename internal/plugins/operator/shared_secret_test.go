package operator_test

import (
	"testing"

	apierrors "k8s.io/apimachinery/pkg/api/errors"

	"github.com/arkadiuss/openfief/internal/plugins/operator"
	"github.com/arkadiuss/openfief/internal/plugins/operator/v1alpha1"
)

// testSharedSecret checks D-89 in the operator: a Secret's name joins the
// instance and the role with "--", which neither may hold, so instance
// clash with role a-token mounts clash--a-token, never clash-a's
// clash-a--token, and both run. The API server refuses a Plugin whose name
// or Secret role holds "--", so no instance-role pair can spell another's
// Secret. The operator's own refusal of a Secret another instance mounts
// (D-182) stays as a second layer; TestMayMountSecrets covers it.
func testSharedSecret(t *testing.T, tc *testCluster, r *operator.PluginReconciler) {
	t.Helper()
	ctx := t.Context()
	holder, gh := samplePlugin("clash-a", imageA), sampleGrant("clash-a")
	create(ctx, t, tc.client, holder)
	create(ctx, t, tc.client, gh)
	tc.reconcile(ctx, t, r, "clash-a")
	if got := tc.state(ctx, t, "clash-a"); got.DeployedImage != imageA {
		t.Fatalf("the first instance does not run: %+v", got)
	}
	clash, gc := samplePlugin("clash", imageB), sampleGrant("clash")
	clash.Spec.Manifest.Secrets = []v1alpha1.SecretMount{{Role: "a-token", MountPath: "/etc/openfief/token"}}
	gc.Spec.Secrets = []string{"clash--a-token"}
	create(ctx, t, tc.client, clash)
	create(ctx, t, tc.client, gc)
	tc.reconcile(ctx, t, r, "clash")
	if got := tc.state(ctx, t, "clash"); got.DeployedImage != imageB {
		t.Errorf("an instance whose name and role once spelled another's Secret does not run: %+v", got)
	}
	tc.reconcile(ctx, t, r, "clash-a")
	if got := tc.state(ctx, t, "clash-a"); got.DeployedImage != imageA {
		t.Errorf("the first instance stopped: %+v", got)
	}
	if err := tc.client.Create(ctx, samplePlugin("clash--b", imageA)); !apierrors.IsInvalid(err) {
		t.Errorf("create a Plugin named clash--b: error = %v, want Invalid", err)
	}
	if err := tc.client.Create(ctx, sampleGrant("clash--b")); !apierrors.IsInvalid(err) {
		t.Errorf("create a PluginGrant named clash--b: error = %v, want Invalid", err)
	}
	role := samplePlugin("clash-b", imageA)
	role.Spec.Manifest.Secrets = []v1alpha1.SecretMount{{Role: "b--token", MountPath: "/etc/openfief/token"}}
	if err := tc.client.Create(ctx, role); !apierrors.IsInvalid(err) {
		t.Errorf("create a Plugin with Secret role b--token: error = %v, want Invalid", err)
	}
}
