package kube_test

import (
	"errors"
	"testing"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	dynamicfake "k8s.io/client-go/dynamic/fake"
	"k8s.io/client-go/kubernetes/fake"

	"github.com/arkadiuss/openfief/internal/ingest/kube"
)

// TestModelProviderBaseURL checks that the core reaches an instance as a
// model provider only while it runs with that capability (D-205): a ready
// model-provider instance has an address; a connector and a view are no model
// provider, an instance that is not ready or not reconciled yet is not running,
// and one that does not exist is no such instance.
func TestModelProviderBaseURL(t *testing.T) {
	t.Parallel()
	dyn := dynamicfake.NewSimpleDynamicClientWithCustomListKinds(runtime.NewScheme(),
		map[schema.GroupVersionResource]string{kube.PluginResource(): "PluginList"},
		plugin("llm", true, "model-provider"),
		plugin("llm-down", false, "model-provider"),
		plugin("xtb", true, "connector"),
		plugin("ui", true, "view"),
		unreconciled("unreconciled"),
	)
	client := fake.NewClientset(service("llm", "192.0.2.20"), service("llm-down", "192.0.2.21"), service("xtb", "192.0.2.22"))
	d := kube.NewDiscovery(dyn, client, "openfief", 0)
	if _, err := d.ModelProviderBaseURL("llm"); !errors.Is(err, kube.ErrNotStarted) {
		t.Errorf("before Start = %v, want ErrNotStarted", err)
	}
	if err := d.Start(t.Context()); err != nil {
		t.Fatalf("Start: %v", err)
	}
	if url, err := d.ModelProviderBaseURL("llm"); err != nil || url != "http://192.0.2.20:8080" {
		t.Errorf("ModelProviderBaseURL(llm) = %q, %v", url, err)
	}
	for name, want := range map[string]error{
		"llm-down": kube.ErrNotRunning, "xtb": kube.ErrNotModelProvider, "ui": kube.ErrNotModelProvider,
		"unreconciled": kube.ErrNotRunning, "absent": kube.ErrNoSuchInstance,
	} {
		if _, err := d.ModelProviderBaseURL(name); !errors.Is(err, want) {
			t.Errorf("ModelProviderBaseURL(%s) = %v, want %v", name, err, want)
		}
	}
}

// unreconciled returns a Plugin resource the operator has not given a status
// yet: its capabilities are not known.
func unreconciled(name string) *unstructured.Unstructured {
	return &unstructured.Unstructured{Object: map[string]any{
		"apiVersion": "openfief.io/v1alpha1",
		"kind":       "Plugin",
		"metadata":   map[string]any{"name": name, "namespace": "openfief"},
	}}
}
