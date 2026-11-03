package operator_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/arkadiuss/openfief/internal/plugins/operator"
	"github.com/arkadiuss/openfief/internal/plugins/operator/v1alpha1"
)

// testSharedSecret checks D-89 and D-182 in the operator: Secret names join
// the instance and the role with a hyphen, so instance clash with role
// a-token names clash-a's token Secret. With clash-a running, clash,
// applied as GitOps would with a grant for that Secret, runs nothing and
// says why in its Ready condition; once clash-a is gone, clash runs.
func testSharedSecret(t *testing.T, tc *testCluster, r *operator.PluginReconciler) {
	t.Helper()
	ctx := t.Context()
	holder, gh := samplePlugin("clash-a", imageA), sampleGrant("clash-a")
	create(ctx, t, tc.client, holder)
	create(ctx, t, tc.client, gh)
	tc.reconcile(ctx, t, r, "clash-a")
	if got := tc.state(ctx, t, "clash-a"); got.DeployedImage != imageA {
		t.Fatalf("the Secret's instance does not run: %+v", got)
	}
	clash, gc := samplePlugin("clash", imageB), sampleGrant("clash")
	clash.Spec.Manifest.Secrets = []v1alpha1.SecretMount{{Role: "a-token", MountPath: "/etc/openfief/token"}}
	gc.Spec.Secrets = []string{"clash-a-token"}
	create(ctx, t, tc.client, clash)
	create(ctx, t, tc.client, gc)
	tc.reconcile(ctx, t, r, "clash")
	want := state{DenyAll: true, Conditions: conditions("False/SecretHeld", "False/Approved", "True/Verified")}
	if diff := cmp.Diff(want, tc.state(ctx, t, "clash"), cmpopts.EquateEmpty()); diff != "" {
		t.Errorf("an instance naming another's Secret (-want +got):\n%s", diff)
	}
	tc.reconcile(ctx, t, r, "clash-a")
	if got := tc.state(ctx, t, "clash-a"); got.DeployedImage != imageA {
		t.Errorf("the Secret's instance stopped when another named its Secret: %+v", got)
	}
	if err := tc.client.Delete(ctx, holder); err != nil {
		t.Fatalf("delete clash-a: %v", err)
	}
	tc.reconcile(ctx, t, r, "clash")
	if got := tc.state(ctx, t, "clash"); got.DeployedImage != imageB {
		t.Errorf("with the Secret's instance gone, the other does not run: %+v", got)
	}
}
