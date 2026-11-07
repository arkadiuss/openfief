package operator_test

import (
	"context"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/arkadiuss/openfief/internal/plugins"
	"github.com/arkadiuss/openfief/internal/plugins/operator"
	"github.com/arkadiuss/openfief/internal/plugins/operator/v1alpha1"
)

func TestStampCABundle(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		before   map[string]string
		checksum string
		want     map[string]string
	}{
		{
			name: "no Secret stamps nothing", before: map[string]string{operator.AnnotationConfigHash: "c"},
			want: map[string]string{operator.AnnotationConfigHash: "c"},
		},
		{
			name: "a Secret's checksum joins the config's", before: map[string]string{operator.AnnotationConfigHash: "c"}, checksum: "abc",
			want: map[string]string{operator.AnnotationConfigHash: "c", operator.AnnotationCABundleHash: "abc"},
		},
		{name: "a template without annotations", checksum: "abc", want: map[string]string{operator.AnnotationCABundleHash: "abc"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			d := &appsv1.Deployment{}
			d.Spec.Template.Annotations = tt.before
			operator.StampCABundle(d, tt.checksum)
			if diff := cmp.Diff(tt.want, d.Spec.Template.Annotations); diff != "" {
				t.Errorf("pod template annotations (-want +got):\n%s", diff)
			}
		})
	}
}

// caBundleState is what testCABundle checks after a reconcile: the
// CABundleMissing condition as "Status/Reason", "" when absent, the
// checksum on the pod template, whether the pod mounts the CA bundle
// Secret, the Ready condition as "Status/Reason": Progressing, the
// Deployment applied and waiting for no bundle, and whether the reconcile
// asked to run again.
type caBundleState struct {
	Condition string
	Checksum  string
	Mounted   bool
	Ready     string
	Requeue   bool
}

// testCABundle checks that an instance expecting an extra CA bundle
// (spec.extraCa) runs with no checksum stamped and says so in the
// CABundleMissing condition; that a stamped checksum turns it False and is
// copied to the pod template, so a new stamp rolls the pod; that the
// operator reads no Secret for it, so a Secret appearing changes nothing
// and nothing is read again on a timer; and that an instance expecting
// none neither mounts the Secret nor reports it.
func testCABundle(t *testing.T, tc *testCluster, r *operator.PluginReconciler) {
	t.Helper()
	ctx := t.Context()
	const name = "cabundle"
	p := samplePlugin(name, imageA)
	p.Spec.ExtraCA = true
	p.Spec.Manifest.Secrets = append(p.Spec.Manifest.Secrets,
		v1alpha1.SecretMount{Role: plugins.CABundleRole, MountPath: "/etc/openfief/ca-bundle", Optional: true})
	g := sampleGrant(name)
	g.Spec.Secrets = append(g.Spec.Secrets, name+"--ca-bundle")
	create(ctx, t, tc.client, p)
	create(ctx, t, tc.client, g)
	check := func(step string, want caBundleState) {
		t.Helper()
		req := ctrl.Request{NamespacedName: types.NamespacedName{Namespace: "openfief", Name: name}}
		res, err := r.Reconcile(ctx, req)
		if err != nil {
			t.Fatalf("%s: Reconcile: %v", step, err)
		}
		got := tc.caBundleState(ctx, t, name)
		got.Requeue = res.RequeueAfter > 0
		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("%s (-want +got):\n%s", step, diff)
		}
	}
	check("nothing stamped", caBundleState{Condition: "True/ChecksumNotStamped", Mounted: true, Ready: "False/Progressing"})

	// A Secret the owner created without the CLI changes nothing: the
	// operator never reads it.
	s := &corev1.Secret{ObjectMeta: metav1.ObjectMeta{
		Name: name + "--ca-bundle", Namespace: "openfief", Labels: map[string]string{plugins.OwnerInstanceLabel: name},
	}, Data: map[string][]byte{plugins.CABundleKey: []byte("synthetic")}}
	create(ctx, t, tc.client, s)
	check("a Secret, nothing stamped", caBundleState{Condition: "True/ChecksumNotStamped", Mounted: true, Ready: "False/Progressing"})

	first, second := plugins.CABundleChecksum([]byte("first")), plugins.CABundleChecksum([]byte("second"))
	stamp(ctx, t, tc, func(spec *v1alpha1.PluginSpec) { spec.CABundleSHA256 = first })
	check("stamped", caBundleState{Condition: "False/ChecksumStamped", Checksum: first, Mounted: true, Ready: "False/Progressing"})
	stamp(ctx, t, tc, func(spec *v1alpha1.PluginSpec) { spec.CABundleSHA256 = second })
	check("rotated", caBundleState{Condition: "False/ChecksumStamped", Checksum: second, Mounted: true, Ready: "False/Progressing"})
	stamp(ctx, t, tc, func(spec *v1alpha1.PluginSpec) { spec.CABundleSHA256 = "" })
	check("unset", caBundleState{Condition: "True/ChecksumNotStamped", Mounted: true, Ready: "False/Progressing"})

	stamp(ctx, t, tc, func(spec *v1alpha1.PluginSpec) { spec.ExtraCA, spec.CABundleSHA256 = false, first })
	check("not expected", caBundleState{Ready: "False/Progressing"})
}

// stamp changes the spec of testCABundle's Plugin with change.
func stamp(ctx context.Context, t *testing.T, tc *testCluster, change func(*v1alpha1.PluginSpec)) {
	t.Helper()
	var cur v1alpha1.Plugin
	if err := tc.client.Get(ctx, types.NamespacedName{Namespace: "openfief", Name: "cabundle"}, &cur); err != nil {
		t.Fatalf("get plugin: %v", err)
	}
	change(&cur.Spec)
	if err := tc.client.Update(ctx, &cur); err != nil {
		t.Fatalf("update plugin: %v", err)
	}
}

// caBundleState reads what testCABundle checks of Plugin name, and fails t
// when the CABundleMissing condition's message shows the Secret's data.
func (tc *testCluster) caBundleState(ctx context.Context, t *testing.T, name string) caBundleState {
	t.Helper()
	var p v1alpha1.Plugin
	if err := tc.client.Get(ctx, types.NamespacedName{Namespace: "openfief", Name: name}, &p); err != nil {
		t.Fatalf("get plugin: %v", err)
	}
	var s caBundleState
	if c := meta.FindStatusCondition(p.Status.Conditions, v1alpha1.ConditionCABundleMissing); c != nil {
		s.Condition = string(c.Status) + "/" + c.Reason
		if strings.Contains(c.Message, "synthetic") {
			t.Errorf("the condition shows the Secret's data: %s", c.Message)
		}
	}
	if c := meta.FindStatusCondition(p.Status.Conditions, v1alpha1.ConditionReady); c != nil {
		s.Ready = string(c.Status) + "/" + c.Reason
	}
	var d appsv1.Deployment
	if err := tc.client.Get(ctx, types.NamespacedName{Namespace: "openfief", Name: "plugin-" + name}, &d); err != nil {
		t.Fatalf("get deployment: %v", err)
	}
	s.Checksum = d.Spec.Template.Annotations[operator.AnnotationCABundleHash]
	for i := range d.Spec.Template.Spec.Volumes {
		v := &d.Spec.Template.Spec.Volumes[i]
		s.Mounted = s.Mounted || (v.Secret != nil && v.Secret.SecretName == name+"--ca-bundle")
	}
	return s
}
