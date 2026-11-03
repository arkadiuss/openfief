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

// TestReconcileGrantWithoutInstitutions checks that a PluginGrant that
// names no institutions at all approves none (D-185): a connector whose
// manifest declares one waits for approval with it in
// status.pendingApproval and doesn't start; and when the field is removed
// from the grant of a running version, the institution waits for approval
// again and status.effective, what the core accepts deliveries for, names
// none.
func TestReconcileGrantWithoutInstitutions(t *testing.T) {
	t.Parallel()
	tc := startCluster(t)
	r := tc.reconciler(signedImages{imageA})
	ctx := t.Context()
	for _, name := range []string{"unnamed", "removed"} {
		p := samplePlugin(name, imageA)
		p.Spec.Manifest.Institutions = []string{"ibkr"}
		create(ctx, t, tc.client, p)
		g := sampleGrant(name)
		if name == "removed" {
			g.Spec.Institutions = &[]string{"ibkr"}
		}
		create(ctx, t, tc.client, g)
		tc.reconcile(ctx, t, r, name)
	}
	if got := tc.state(ctx, t, "unnamed"); !pendingIbkr(got.Pending) || got.DeployedImage != "" {
		t.Errorf("grant without institutions: pending %+v, deployed %q; want ibkr pending and nothing running", got.Pending, got.DeployedImage)
	}
	if got := tc.state(ctx, t, "removed"); got.Pending != nil || got.DeployedImage != imageA {
		t.Fatalf("approved institution: pending %+v, deployed %q; want the image running", got.Pending, got.DeployedImage)
	}
	setGrantInstitutions(t, tc, "removed", nil)
	tc.reconcile(ctx, t, r, "removed")
	if got := tc.state(ctx, t, "removed"); !pendingIbkr(got.Pending) {
		t.Errorf("grant whose institutions were removed: pending %+v; want ibkr pending", got.Pending)
	}
	if got := effectiveInstitutions(t, tc, "removed"); len(got) != 0 {
		t.Errorf("status.effective.institutions = %v, want none", got)
	}
}

// pendingIbkr reports whether pending holds exactly the institution ibkr.
func pendingIbkr(pending *v1alpha1.Permissions) bool {
	return pending != nil && pending.Institutions != nil && slices.Equal(*pending.Institutions, []string{"ibkr"})
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
