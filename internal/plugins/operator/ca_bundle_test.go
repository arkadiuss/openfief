package operator_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

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

// forbiddenSecrets is a client.Reader that may read no Secret, as the
// operator without the owner's Role for a CA bundle.
type forbiddenSecrets struct{}

func (forbiddenSecrets) Get(_ context.Context, key client.ObjectKey, _ client.Object, _ ...client.GetOption) error {
	return forbidden(key.Name)
}

func (forbiddenSecrets) List(context.Context, client.ObjectList, ...client.ListOption) error {
	return forbidden("")
}

// forbidden is the API server's refusal to read Secret name.
func forbidden(name string) *apierrors.StatusError {
	return apierrors.NewForbidden(schema.GroupResource{Resource: "secrets"}, name, errNoRole)
}

// errNoRole is why forbiddenSecrets refuses.
var errNoRole = errors.New("no Role grants it")

// caBundleState is what testCABundle checks after a reconcile: the
// CABundleMissing condition as "Status/Reason", "" when absent, the
// checksum on the pod template, whether the pod mounts the CA bundle
// Secret, and the Ready condition as "Status/Reason": Progressing, the
// Deployment applied and waiting for no bundle.
type caBundleState struct {
	Condition string
	Checksum  string
	Mounted   bool
	Ready     string
}

// testCABundle checks that an instance expecting an extra CA bundle
// (spec.extraCa) runs while its Secret is missing and says so in the
// CABundleMissing condition, never with the Secret's data; that creating,
// changing and deleting the Secret changes the checksum on the pod
// template, which rolls the pod, and that it is read again every minute;
// that a Secret the operator may not read is reported as unknown; and that
// an instance expecting none neither mounts the Secret nor reports it.
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
		tc.reconcile(ctx, t, r, name)
		if diff := cmp.Diff(want, tc.caBundleState(ctx, t, name)); diff != "" {
			t.Errorf("%s (-want +got):\n%s", step, diff)
		}
	}
	check("missing", caBundleState{Condition: "True/SecretMissing", Mounted: true, Ready: "False/Progressing"})
	// The operator may only get the Secret, never watch it: it reads it
	// again every minute.
	req := ctrl.Request{NamespacedName: types.NamespacedName{Namespace: "openfief", Name: name}}
	if res, err := r.Reconcile(ctx, req); err != nil || res.RequeueAfter <= 0 {
		t.Errorf("Reconcile = %+v, %v; want a requeue while a bundle is expected", res, err)
	}

	data := map[string][]byte{"ca.crt": []byte("-----BEGIN CERTIFICATE-----\nAA==\n-----END CERTIFICATE-----\n")}
	s := &corev1.Secret{ObjectMeta: metav1.ObjectMeta{
		Name: name + "--ca-bundle", Namespace: "openfief", Labels: map[string]string{plugins.OwnerInstanceLabel: name},
	}, Data: data}
	create(ctx, t, tc.client, s)
	check("created", caBundleState{
		Condition: "False/SecretPresent", Checksum: plugins.SecretDataChecksum(data), Mounted: true, Ready: "False/Progressing",
	})

	changed := map[string][]byte{"ca.crt": []byte("-----BEGIN CERTIFICATE-----\nBB==\n-----END CERTIFICATE-----\n")}
	s.Data = changed
	if err := tc.client.Update(ctx, s); err != nil {
		t.Fatalf("update secret: %v", err)
	}
	check("changed", caBundleState{
		Condition: "False/SecretPresent", Checksum: plugins.SecretDataChecksum(changed), Mounted: true, Ready: "False/Progressing",
	})

	if err := tc.client.Delete(ctx, s); err != nil {
		t.Fatalf("delete secret: %v", err)
	}
	check("deleted", caBundleState{Condition: "True/SecretMissing", Mounted: true, Ready: "False/Progressing"})

	// Without the owner's Role for the Secret, the operator can't tell, and
	// says how to let it; the plugin still runs.
	unread := *r
	unread.SecretReader = forbiddenSecrets{}
	tc.reconcile(ctx, t, &unread, name)
	want := caBundleState{Condition: "Unknown/SecretUnreadable", Mounted: true, Ready: "False/Progressing"}
	if got := tc.caBundleState(ctx, t, name); got != want {
		t.Errorf("unreadable: %+v, want %+v", got, want)
	}

	var cur v1alpha1.Plugin
	if err := tc.client.Get(ctx, types.NamespacedName{Namespace: "openfief", Name: name}, &cur); err != nil {
		t.Fatalf("get plugin: %v", err)
	}
	cur.Spec.ExtraCA = false
	if err := tc.client.Update(ctx, &cur); err != nil {
		t.Fatalf("update plugin: %v", err)
	}
	check("not expected", caBundleState{Ready: "False/Progressing"})
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
		if strings.Contains(c.Message, "BEGIN CERTIFICATE") {
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
