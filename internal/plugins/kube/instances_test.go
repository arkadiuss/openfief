package kube_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/arkadiuss/openfief/internal/plugins"
	"github.com/arkadiuss/openfief/internal/plugins/app"
	"github.com/arkadiuss/openfief/internal/plugins/kube"
	"github.com/arkadiuss/openfief/internal/plugins/operator/v1alpha1"
)

// sampleImage is a synthetic digest-pinned image.
const sampleImage = "registry.example.com/example/sample@sha256:a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1"

// sampleYAML is what installing sample renders for GitOps.
const sampleYAML = `---
apiVersion: openfief.io/v1alpha1
kind: Plugin
metadata:
  name: sample
  namespace: openfief
spec:
  config:
    interval: 24h
  image: registry.example.com/example/sample@sha256:a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1
  manifest:
    capabilities:
    - connector
    configSchema:
      type: object
    egress:
    - api.example.com:443
    secrets:
    - mountPath: /etc/openfief/sample
      name: sample-credentials
---
apiVersion: openfief.io/v1alpha1
kind: PluginGrant
metadata:
  name: sample
  namespace: openfief
spec:
  capabilities:
  - connector
  egress:
  - api.example.com:443
  secrets:
  - sample-credentials
`

// samplePlan is an install of sample that requests one egress host and one
// Secret, with a config.
func samplePlan(t *testing.T, name string) (plugins.Plugin, plugins.Grant) {
	t.Helper()
	img, err := plugins.ParseImageRef(sampleImage)
	if err != nil {
		t.Fatalf("image: %v", err)
	}
	signed := plugins.SignedImage{Image: img, Manifest: plugins.Manifest{
		Capabilities: []plugins.Capability{plugins.CapabilityConnector},
		Egress:       []plugins.EgressHost{{Host: "api.example.com", Port: 443}},
		Secrets:      []plugins.SecretMount{{Name: "sample-credentials", MountPath: "/etc/openfief/sample"}},
		ConfigSchema: []byte(`{"type":"object"}`),
	}}
	p, g, err := plugins.PlanInstall(name, &signed, []byte(`{"interval":"24h"}`))
	if err != nil {
		t.Fatalf("plan: %v", err)
	}
	return p, g
}

// fakeCluster is an Instances store on a fake API server holding objs, with
// the operator's types.
func fakeCluster(t *testing.T, objs ...client.Object) *kube.Instances {
	t.Helper()
	scheme := runtime.NewScheme()
	if err := errors.Join(clientgoscheme.AddToScheme(scheme), v1alpha1.AddToScheme(scheme)); err != nil {
		t.Fatalf("scheme: %v", err)
	}
	return &kube.Instances{Client: fake.NewClientBuilder().WithScheme(scheme).WithObjects(objs...).Build()}
}

func secret(name string) *corev1.Secret {
	return &corev1.Secret{ObjectMeta: metav1.ObjectMeta{Namespace: plugins.Namespace, Name: name}}
}

func TestResourcesYAML(t *testing.T) {
	t.Parallel()
	p, g := samplePlan(t, "sample")
	got, err := kube.ResourcesYAML(&p, &g)
	if err != nil {
		t.Fatalf("render: %v", err)
	}
	if diff := cmp.Diff(sampleYAML, string(got)); diff != "" {
		t.Errorf("YAML (-want +got):\n%s", diff)
	}
}

func TestInstancesCreateAndList(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	running := &v1alpha1.Plugin{
		ObjectMeta: metav1.ObjectMeta{Namespace: plugins.Namespace, Name: "running"},
		Spec:       v1alpha1.PluginSpec{Image: sampleImage, Manifest: v1alpha1.Manifest{Capabilities: []string{"view"}}},
		Status: v1alpha1.PluginStatus{
			Conditions: []metav1.Condition{
				{Type: v1alpha1.ConditionReady, Status: metav1.ConditionFalse, Reason: "NoGrant"},
				{Type: v1alpha1.ConditionSignatureVerified, Status: metav1.ConditionTrue, Reason: "SignedByTrustedSigner"},
			},
			PendingApproval: &v1alpha1.Permissions{Capabilities: []string{"view"}},
		},
	}
	orphan := &v1alpha1.PluginGrant{
		ObjectMeta: metav1.ObjectMeta{Namespace: plugins.Namespace, Name: "orphan"},
		Spec:       v1alpha1.Permissions{Secrets: []string{"orphan-key"}},
	}
	store := fakeCluster(t, running, orphan)
	p, g := samplePlan(t, "sample")
	if err := store.Create(ctx, &p, &g); err != nil {
		t.Fatalf("create: %v", err)
	}
	for _, name := range []string{"sample", "running", "orphan"} {
		other, otherGrant := samplePlan(t, name)
		if err := store.Create(ctx, &other, &otherGrant); !errors.Is(err, app.ErrInstanceExists) {
			t.Errorf("create %s again: error = %v, want ErrInstanceExists", name, err)
		}
	}
	got, err := store.List(ctx)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	want := []app.Instance{
		{Name: "orphan", HasGrant: true, Secrets: []string{"orphan-key"}},
		{Name: "running", Image: sampleImage, HasPlugin: true, Signature: "SignedByTrustedSigner", Ready: "False (NoGrant)", Pending: true},
		{Name: "sample", Image: sampleImage, HasPlugin: true, HasGrant: true, Secrets: []string{"sample-credentials"}},
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("instances (-want +got):\n%s", diff)
	}
}

func TestInstancesDelete(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	p, g := samplePlan(t, "sample")
	plugin, grant := kube.Resources(&p, &g)
	store := fakeCluster(t, plugin, grant, secret("sample-credentials"), secret("shared"), secret("unrelated"))
	c := store.Client
	// "missing" was deleted by hand already: that is not an error.
	if err := store.Delete(ctx, "sample", []string{"sample-credentials", "missing"}); err != nil {
		t.Fatalf("delete: %v", err)
	}
	gone := []client.Object{&v1alpha1.Plugin{}, &v1alpha1.PluginGrant{}, &corev1.Secret{}}
	for i, name := range []string{"sample", "sample", "sample-credentials"} {
		if err := c.Get(ctx, client.ObjectKey{Namespace: plugins.Namespace, Name: name}, gone[i]); !apierrors.IsNotFound(err) {
			t.Errorf("%T %s: get error = %v, want not found", gone[i], name, err)
		}
	}
	for _, name := range []string{"shared", "unrelated"} {
		if err := c.Get(ctx, client.ObjectKey{Namespace: plugins.Namespace, Name: name}, &corev1.Secret{}); err != nil {
			t.Errorf("Secret %s was not to be deleted: %v", name, err)
		}
	}
	if err := store.Delete(ctx, "sample", nil); err != nil {
		t.Errorf("deleting an instance that is gone: %v", err)
	}
}

func TestOwnerSigners(t *testing.T) {
	t.Parallel()
	cm := func(data map[string]string) *corev1.ConfigMap {
		return &corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Namespace: plugins.Namespace, Name: kube.TrustedSignersConfigMap}, Data: data}
	}
	tests := []struct {
		name    string
		objs    []client.Object
		want    []string
		wantErr error
	}{
		{name: "no ConfigMap"},
		{name: "no key", objs: []client.Object{cm(map[string]string{"other": "{}"})}},
		{
			name: "one signer",
			objs: []client.Object{cm(map[string]string{kube.TrustedSignersKey: `{"signers":[{"name":"my-ci","issuer":"https://issuer.example.com","subject":"https://example.com/ci"}]}`})},
			want: []string{"my-ci"},
		},
		{
			name:    "invalid document",
			objs:    []client.Object{cm(map[string]string{kube.TrustedSignersKey: `{"signers":[{"name":"x","extra":1}]}`})},
			wantErr: plugins.ErrInvalidCatalog,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			signers, err := kube.OwnerSigners(context.Background(), fakeCluster(t, tt.objs...).Client)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("error = %v, want %v", err, tt.wantErr)
			}
			var got []string
			for _, s := range signers {
				got = append(got, s.Name)
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("signers (-want +got):\n%s", diff)
			}
		})
	}
}
