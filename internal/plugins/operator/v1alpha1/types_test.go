package v1alpha1_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/arkadiuss/openfief/internal/plugins/operator/v1alpha1"
)

func TestAddToScheme(t *testing.T) {
	t.Parallel()
	s := runtime.NewScheme()
	if err := v1alpha1.AddToScheme(s); err != nil {
		t.Fatalf("AddToScheme: %v", err)
	}
	for _, kind := range []string{"Plugin", "PluginList", "PluginGrant", "PluginGrantList"} {
		gvk := schema.GroupVersionKind{Group: "openfief.io", Version: "v1alpha1", Kind: kind}
		if !s.Recognizes(gvk) {
			t.Errorf("scheme does not recognize %s", gvk)
		}
	}
}

// samplePlugin is a synthetic Plugin with every field set; the digest is
// not of any real image.
func samplePlugin() *v1alpha1.Plugin {
	return &v1alpha1.Plugin{
		ObjectMeta: metav1.ObjectMeta{Name: "sample", Namespace: "openfief"},
		Spec: v1alpha1.PluginSpec{
			Image: "registry.example.com/openfief/sample@sha256:a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1",
			Manifest: v1alpha1.Manifest{
				Capabilities: []string{"connector"},
				Egress:       []string{"api.example.com"},
				Secrets:      []v1alpha1.SecretMount{{Name: "sample-token", MountPath: "/etc/openfief/token"}},
				DataTiers:    []string{"T1"},
				ConfigSchema: &runtime.RawExtension{Raw: []byte(`{"type":"object"}`)},
			},
			Config: &runtime.RawExtension{Raw: []byte(`{"interval":"24h"}`)},
		},
		Status: v1alpha1.PluginStatus{
			Conditions:      []metav1.Condition{{Type: "Ready", Status: metav1.ConditionFalse, Reason: "NoGrant"}},
			PendingApproval: &v1alpha1.Permissions{Egress: []string{"api.example.com:443"}},
		},
	}
}

func TestDeepCopyIsIndependent(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		obj  runtime.Object
		edit func(runtime.Object)
	}{
		{
			name: "Plugin",
			obj:  samplePlugin(),
			edit: func(o runtime.Object) {
				if p, ok := o.(*v1alpha1.Plugin); ok {
					p.Spec.Manifest.Egress[0] = "evil.example.net"
					p.Spec.Config.Raw[2] = 'X'
					p.Status.PendingApproval.Egress[0] = "evil.example.net:443"
				}
			},
		},
		{
			name: "PluginGrant",
			obj: &v1alpha1.PluginGrant{
				ObjectMeta: metav1.ObjectMeta{Name: "sample", Namespace: "openfief"},
				Spec:       v1alpha1.Permissions{Capabilities: []string{"connector"}, Secrets: []string{"sample-token"}},
			},
			edit: func(o runtime.Object) {
				if g, ok := o.(*v1alpha1.PluginGrant); ok {
					g.Spec.Secrets[0] = "other-token"
				}
			},
		},
		{
			name: "PluginList",
			obj:  &v1alpha1.PluginList{Items: []v1alpha1.Plugin{*samplePlugin()}},
			edit: func(o runtime.Object) {
				if l, ok := o.(*v1alpha1.PluginList); ok {
					l.Items[0].Spec.Image = "changed"
				}
			},
		},
		{
			name: "PluginGrantList",
			obj:  &v1alpha1.PluginGrantList{Items: []v1alpha1.PluginGrant{{Spec: v1alpha1.Permissions{Egress: []string{"api.example.com"}}}}},
			edit: func(o runtime.Object) {
				if l, ok := o.(*v1alpha1.PluginGrantList); ok {
					l.Items[0].Spec.Egress[0] = "evil.example.net"
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			copied := tt.obj.DeepCopyObject()
			if diff := cmp.Diff(tt.obj, copied); diff != "" {
				t.Fatalf("DeepCopyObject differs (-original +copy):\n%s", diff)
			}
			tt.edit(copied)
			if cmp.Equal(tt.obj, copied) {
				t.Errorf("editing the copy changed the original")
			}
		})
	}
}
