package operator_test

import (
	"slices"
	"testing"

	"k8s.io/apimachinery/pkg/types"

	"github.com/arkadiuss/openfief/internal/plugins/operator/v1alpha1"
)

// TestReconcileInstitutions checks D-185: a connector's institutions wait
// for the PluginGrant to approve them, with institutions in
// status.pendingApproval and nothing running, and once approved run with
// them in status.effective, which is what the core accepts deliveries for.
func TestReconcileInstitutions(t *testing.T) {
	t.Parallel()
	tc := startCluster(t)
	r := tc.reconciler(signedImages{imageA})
	ctx := t.Context()
	p := samplePlugin("broker", imageA)
	p.Spec.Manifest.Institutions = []string{"ibkr"}
	create(ctx, t, tc.client, p)
	g := sampleGrant("broker")
	none := []string{}
	g.Spec.Institutions = &none
	create(ctx, t, tc.client, g)
	tc.reconcile(ctx, t, r, "broker")
	got := tc.state(ctx, t, "broker")
	if got.Pending == nil || got.Pending.Institutions == nil || !slices.Equal(*got.Pending.Institutions, []string{"ibkr"}) ||
		got.DeployedImage != "" {
		t.Fatalf("unapproved institution: pending %+v, deployed %q; want ibkr pending and nothing running", got.Pending, got.DeployedImage)
	}
	setGrantInstitutions(t, tc, "broker", &[]string{"ibkr"})
	tc.reconcile(ctx, t, r, "broker")
	if got := tc.state(ctx, t, "broker"); got.Pending != nil || got.DeployedImage != imageA {
		t.Errorf("approved institution: pending %+v, deployed %q; want nothing pending and the image running", got.Pending, got.DeployedImage)
	}
	if got := effectiveInstitutions(t, tc, "broker"); !slices.Equal(got, []string{"ibkr"}) {
		t.Errorf("status.effective.institutions = %v, want [ibkr]", got)
	}
}

// TestReconcileLegacyGrantInstitutions checks that a PluginGrant made before
// D-185, which names no institutions at all, approves those of the version
// that runs under it, so it keeps running, while a version that declares
// another institution waits for the owner.
func TestReconcileLegacyGrantInstitutions(t *testing.T) {
	t.Parallel()
	tc := startCluster(t)
	r := tc.reconciler(signedImages{imageA, imageB})
	ctx := t.Context()
	p := samplePlugin("legacy", imageA)
	p.Spec.Manifest.Institutions = []string{"ibkr"}
	create(ctx, t, tc.client, p)
	g := sampleGrant("legacy")
	g.Spec.Institutions = &[]string{"ibkr"}
	create(ctx, t, tc.client, g)
	tc.reconcile(ctx, t, r, "legacy")
	// The grant as one made before D-185: no institutions field.
	setGrantInstitutions(t, tc, "legacy", nil)
	tc.reconcile(ctx, t, r, "legacy")
	if got := tc.state(ctx, t, "legacy"); got.Pending != nil || got.DeployedImage != imageA {
		t.Fatalf("legacy grant: pending %+v, deployed %q; want the running version kept", got.Pending, got.DeployedImage)
	}
	if got := effectiveInstitutions(t, tc, "legacy"); !slices.Equal(got, []string{"ibkr"}) {
		t.Errorf("legacy grant: status.effective.institutions = %v, want [ibkr]", got)
	}
	var stored v1alpha1.Plugin
	if err := tc.client.Get(ctx, types.NamespacedName{Namespace: "openfief", Name: "legacy"}, &stored); err != nil {
		t.Fatalf("get plugin: %v", err)
	}
	stored.Spec.Image = imageB
	stored.Spec.Manifest.Institutions = []string{"ibkr", "xtb"}
	if err := tc.client.Update(ctx, &stored); err != nil {
		t.Fatalf("upgrade the plugin: %v", err)
	}
	tc.reconcile(ctx, t, r, "legacy")
	got := tc.state(ctx, t, "legacy")
	if got.Pending == nil || got.Pending.Institutions == nil || !slices.Equal(*got.Pending.Institutions, []string{"xtb"}) ||
		got.DeployedImage != imageA {
		t.Errorf("legacy grant, new institution: pending %+v, deployed %q; want xtb pending and the old version running", got.Pending, got.DeployedImage)
	}
}

// setGrantInstitutions sets the named PluginGrant's institutions.
func setGrantInstitutions(t *testing.T, tc *testCluster, name string, institutions *[]string) {
	t.Helper()
	var stored v1alpha1.PluginGrant
	if err := tc.client.Get(t.Context(), types.NamespacedName{Namespace: "openfief", Name: name}, &stored); err != nil {
		t.Fatalf("get grant: %v", err)
	}
	stored.Spec.Institutions = institutions
	if err := tc.client.Update(t.Context(), &stored); err != nil {
		t.Fatalf("update grant: %v", err)
	}
}

// effectiveInstitutions reads the named Plugin's status.effective.institutions.
func effectiveInstitutions(t *testing.T, tc *testCluster, name string) []string {
	t.Helper()
	var p v1alpha1.Plugin
	if err := tc.client.Get(t.Context(), types.NamespacedName{Namespace: "openfief", Name: name}, &p); err != nil {
		t.Fatalf("get plugin: %v", err)
	}
	if p.Status.Effective == nil || p.Status.Effective.Institutions == nil {
		return nil
	}
	return *p.Status.Effective.Institutions
}
