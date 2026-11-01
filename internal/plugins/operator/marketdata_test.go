package operator_test

import (
	"testing"

	"k8s.io/apimachinery/pkg/types"

	"github.com/arkadiuss/openfief/internal/plugins/operator/v1alpha1"
)

// TestReconcileMarketData checks that a connector whose manifest declares
// it a market-data connector waits for the PluginGrant to approve that,
// with marketData in status.pendingApproval, and once approved runs with
// marketData in status.effective, which is what the core reads before it
// tells the connector what the owner holds (D-123).
func TestReconcileMarketData(t *testing.T) {
	t.Parallel()
	tc := startCluster(t)
	r := tc.reconciler(signedImages{imageA})
	ctx := t.Context()
	p := samplePlugin("prices", imageA)
	p.Spec.Manifest.MarketData = true
	create(ctx, t, tc.client, p)
	g := sampleGrant("prices")
	create(ctx, t, tc.client, g)
	tc.reconcile(ctx, t, r, "prices")
	got := tc.state(ctx, t, "prices")
	if got.Pending == nil || !got.Pending.MarketData || got.Conditions["ApprovalRequired"] != "True/PermissionsPending" {
		t.Fatalf("unapproved market data: pending %+v, conditions %v; want marketData pending", got.Pending, got.Conditions)
	}
	if marketData(t, tc, "prices") {
		t.Error("status.effective.marketData is true before the grant approves it")
	}
	var stored v1alpha1.PluginGrant
	if err := tc.client.Get(ctx, types.NamespacedName{Namespace: "openfief", Name: "prices"}, &stored); err != nil {
		t.Fatalf("get grant: %v", err)
	}
	stored.Spec.MarketData = true
	if err := tc.client.Update(ctx, &stored); err != nil {
		t.Fatalf("approve market data: %v", err)
	}
	tc.reconcile(ctx, t, r, "prices")
	if got := tc.state(ctx, t, "prices"); got.Pending != nil || got.DeployedImage != imageA {
		t.Errorf("approved market data: pending %+v, deployed %q; want nothing pending and the image running", got.Pending, got.DeployedImage)
	}
	if !marketData(t, tc, "prices") {
		t.Error("status.effective.marketData is false once the grant approves it")
	}
}

// marketData reads the named Plugin's status.effective.marketData.
func marketData(t *testing.T, tc *testCluster, name string) bool {
	t.Helper()
	var p v1alpha1.Plugin
	if err := tc.client.Get(t.Context(), types.NamespacedName{Namespace: "openfief", Name: name}, &p); err != nil {
		t.Fatalf("get plugin: %v", err)
	}
	return p.Status.Effective != nil && p.Status.Effective.MarketData
}
