package operator_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/arkadiuss/openfief/internal/plugins/operator"
	"github.com/arkadiuss/openfief/internal/plugins/operator/v1alpha1"
)

// viewPlugin is a view plugin that declares browser path path, with the
// grant that approves it.
func viewPlugin(name, image, path string) (*v1alpha1.Plugin, *v1alpha1.PluginGrant) {
	return &v1alpha1.Plugin{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: "openfief"},
		Spec: v1alpha1.PluginSpec{Image: image, Manifest: v1alpha1.Manifest{
			Capabilities: []string{"view"}, BrowserPath: path,
		}},
	}, &v1alpha1.PluginGrant{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: "openfief"},
		Spec:       v1alpha1.Permissions{Capabilities: []string{"view"}, BrowserPaths: []string{path}},
	}
}

// testOneBrowserView checks D-114 in the operator: with view-a holding its
// browser path, view-b, applied as GitOps would with its own grant, runs
// nothing and says why in its Ready condition; once view-a is gone,
// view-b runs.
func testOneBrowserView(t *testing.T, tc *testCluster, r *operator.PluginReconciler) {
	t.Helper()
	ctx := t.Context()
	a, ga := viewPlugin("view-a", imageA, "/")
	b, gb := viewPlugin("view-b", imageB, "/gallery")
	create(ctx, t, tc.client, a)
	create(ctx, t, tc.client, ga)
	tc.reconcile(ctx, t, r, "view-a")
	if got := tc.state(ctx, t, "view-a"); got.DeployedImage != imageA {
		t.Fatalf("the first view does not run: %+v", got)
	}
	create(ctx, t, tc.client, b)
	create(ctx, t, tc.client, gb)
	tc.reconcile(ctx, t, r, "view-b")
	want := state{DenyAll: true, Conditions: conditions("False/BrowserPathHeld", "False/Approved", "True/Verified")}
	if diff := cmp.Diff(want, tc.state(ctx, t, "view-b"), cmpopts.EquateEmpty()); diff != "" {
		t.Errorf("a second view (-want +got):\n%s", diff)
	}
	tc.reconcile(ctx, t, r, "view-a")
	if got := tc.state(ctx, t, "view-a"); got.DeployedImage != imageA {
		t.Errorf("the first view stopped when a second came: %+v", got)
	}
	if err := tc.client.Delete(ctx, a); err != nil {
		t.Fatalf("delete view-a: %v", err)
	}
	tc.reconcile(ctx, t, r, "view-b")
	if got := tc.state(ctx, t, "view-b"); got.DeployedImage != imageB || got.Conditions["Ready"] != "False/Progressing" {
		t.Errorf("with the first view gone, the second does not run: %+v", got)
	}
}
