package kube_test

import (
	"context"
	"errors"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/wait"
	dynamicfake "k8s.io/client-go/dynamic/fake"

	"github.com/arkadiuss/openfief/internal/plugins/operator"
	"github.com/arkadiuss/openfief/internal/plugins/operator/v1alpha1"
	"github.com/arkadiuss/openfief/internal/views"
	"github.com/arkadiuss/openfief/internal/views/app"
	"github.com/arkadiuss/openfief/internal/views/kube"
)

// testImage is a synthetic image reference; not of any real image.
const testImage = "registry.example.com/openfief/sample@sha256:a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1"

// operatorPlugin is a Plugin as the operator writes it: a running
// connector its CI signed, converted to the unstructured data the adapter
// reads, so a renamed field in v1alpha1 fails the test.
func operatorPlugin(t *testing.T, name string) *unstructured.Unstructured {
	t.Helper()
	p := v1alpha1.Plugin{
		TypeMeta:   metav1.TypeMeta{APIVersion: "openfief.io/v1alpha1", Kind: "Plugin"},
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: "openfief"},
		Spec: v1alpha1.PluginSpec{
			Image:    testImage,
			Manifest: v1alpha1.Manifest{Capabilities: []string{"connector", "view"}},
		},
		Status: v1alpha1.PluginStatus{
			Conditions: []metav1.Condition{
				{Type: v1alpha1.ConditionReady, Status: metav1.ConditionTrue, Reason: operator.ReasonRunning},
				{Type: v1alpha1.ConditionSignatureVerified, Status: metav1.ConditionTrue, Reason: operator.ReasonTrustedSigner},
				{Type: v1alpha1.ConditionApprovalRequired, Status: metav1.ConditionFalse, Reason: operator.ReasonApproved},
			},
			Effective: &v1alpha1.Permissions{Capabilities: []string{"connector"}},
			Signature: &v1alpha1.SignatureStatus{Trust: "Unverified", SignerKind: "OpenFiefCI", Signer: "openfief-ci"},
		},
	}
	obj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&p)
	if err != nil {
		t.Fatalf("to unstructured: %v", err)
	}
	return &unstructured.Unstructured{Object: obj}
}

// signedRecord is what the adapter must read from operatorPlugin(name).
func signedRecord(name string) views.PluginRecord {
	return views.PluginRecord{
		Instance: name, Image: testImage,
		Requested: []string{"connector", "view"}, Effective: []string{"connector"},
		Ready:     views.Condition{Status: "True", Reason: "Running"},
		Signature: views.Condition{Status: "True", Reason: "SignedByTrustedSigner"},
		Trust:     "Unverified", SignerKind: "OpenFiefCI", Signer: "openfief-ci",
	}
}

func TestRecordOf(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		obj  func(t *testing.T) *unstructured.Unstructured
		want views.PluginRecord
	}{
		{"as the operator writes it", func(t *testing.T) *unstructured.Unstructured {
			t.Helper()
			return operatorPlugin(t, "ibkr-main")
		}, signedRecord("ibkr-main")},
		{"not reconciled yet", func(t *testing.T) *unstructured.Unstructured {
			t.Helper()
			u := operatorPlugin(t, "fresh")
			delete(u.Object, "status")
			return u
		}, views.PluginRecord{Instance: "fresh", Image: testImage, Requested: []string{"connector", "view"}}},
		{"malformed fields read as empty", func(t *testing.T) *unstructured.Unstructured {
			t.Helper()
			return &unstructured.Unstructured{Object: map[string]any{
				"metadata": map[string]any{"name": "odd"},
				"spec":     map[string]any{"image": 7},
				"status":   map[string]any{"conditions": []any{"Ready", map[string]any{"type": "Ready", "status": true}}},
			}}
		}, views.PluginRecord{Instance: "odd"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if diff := cmp.Diff(tt.want, kube.RecordOf(tt.obj(t))); diff != "" {
				t.Errorf("RecordOf (-want +got):\n%s", diff)
			}
		})
	}
}

func TestPlugins(t *testing.T) {
	t.Parallel()
	deleting := operatorPlugin(t, "old")
	deleting.SetDeletionTimestamp(&metav1.Time{Time: time.Date(2026, time.March, 2, 6, 0, 0, 0, time.UTC)})
	other := operatorPlugin(t, "elsewhere")
	other.SetNamespace("default")
	dyn := dynamicfake.NewSimpleDynamicClientWithCustomListKinds(runtime.NewScheme(),
		map[schema.GroupVersionResource]string{kube.PluginResource(): "PluginList"},
		operatorPlugin(t, "ibkr-main"), deleting, other,
	)
	p := kube.NewPlugins(dyn, "openfief", 0)
	if _, err := p.InstalledPlugins(t.Context()); !errors.Is(err, app.ErrPluginsUnknown) {
		t.Errorf("before Start = %v, want ErrPluginsUnknown", err)
	}
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	if err := p.Start(ctx); err != nil {
		t.Fatalf("Start: %v", err)
	}
	old := signedRecord("old")
	old.Deleting = true
	if diff := cmp.Diff([]views.PluginRecord{signedRecord("ibkr-main"), old}, sorted(ctx, t, p)); diff != "" {
		t.Errorf("InstalledPlugins (-want +got):\n%s", diff)
	}

	if _, err := dyn.Resource(kube.PluginResource()).Namespace("openfief").Create(ctx, operatorPlugin(t, "ui"), metav1.CreateOptions{}); err != nil {
		t.Fatalf("create: %v", err)
	}
	err := wait.PollUntilContextTimeout(ctx, 10*time.Millisecond, 10*time.Second, true, func(ctx context.Context) (bool, error) {
		return len(sorted(ctx, t, p)) == 3, nil
	})
	if err != nil {
		t.Errorf("a Plugin added after Start was not listed: %v", err)
	}
}

func TestStartGivesUp(t *testing.T) {
	t.Parallel()
	dyn := dynamicfake.NewSimpleDynamicClientWithCustomListKinds(runtime.NewScheme(),
		map[schema.GroupVersionResource]string{kube.PluginResource(): "PluginList"})
	p := kube.NewPlugins(dyn, "openfief", 0)
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	if err := p.Start(ctx); err == nil {
		t.Error("Start with an ended context succeeded, want an error")
	}
}

// sorted returns p's records in name order.
func sorted(ctx context.Context, t *testing.T, p *kube.Plugins) []views.PluginRecord {
	t.Helper()
	got, err := p.InstalledPlugins(ctx)
	if err != nil {
		t.Fatalf("InstalledPlugins: %v", err)
	}
	slices.SortFunc(got, func(a, b views.PluginRecord) int { return strings.Compare(a.Instance, b.Instance) })
	return got
}
