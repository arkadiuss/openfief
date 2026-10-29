package kube_test

import (
	"errors"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	dynamicfake "k8s.io/client-go/dynamic/fake"
	"k8s.io/client-go/kubernetes/fake"

	"github.com/arkadiuss/openfief/internal/plugins/operator/v1alpha1"
	"github.com/arkadiuss/openfief/internal/views"
	"github.com/arkadiuss/openfief/internal/views/app"
	"github.com/arkadiuss/openfief/internal/views/kube"
)

// viewPlugin is a Plugin as the operator writes it for a view instance:
// running when ready, with effective capabilities caps and browser paths,
// converted to unstructured data so a renamed field in v1alpha1 fails.
func viewPlugin(t *testing.T, name string, ready bool, caps []string, paths ...string) *unstructured.Unstructured {
	t.Helper()
	status := metav1.ConditionFalse
	if ready {
		status = metav1.ConditionTrue
	}
	p := v1alpha1.Plugin{
		TypeMeta:   metav1.TypeMeta{APIVersion: "openfief.io/v1alpha1", Kind: "Plugin"},
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: "openfief"},
		Spec:       v1alpha1.PluginSpec{Image: testImage, Manifest: v1alpha1.Manifest{Capabilities: caps}},
		Status: v1alpha1.PluginStatus{
			Conditions: []metav1.Condition{{Type: v1alpha1.ConditionReady, Status: status, Reason: "Synthetic"}},
			Effective:  &v1alpha1.Permissions{Capabilities: caps, BrowserPaths: paths},
		},
	}
	obj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&p)
	if err != nil {
		t.Fatalf("to unstructured: %v", err)
	}
	return &unstructured.Unstructured{Object: obj}
}

func TestBrowserRouteOf(t *testing.T) {
	t.Parallel()
	view := []string{"view"}
	tests := []struct {
		name  string
		build func(t *testing.T) *unstructured.Unstructured
		want  views.BrowserRoute
		ok    bool
	}{
		{
			name: "a ready view with a browser path", ok: true, want: views.BrowserRoute{Instance: "ui", Prefix: "/"},
			build: func(t *testing.T) *unstructured.Unstructured { t.Helper(); return viewPlugin(t, "ui", true, view, "/") },
		},
		{
			name: "not ready",
			build: func(t *testing.T) *unstructured.Unstructured {
				t.Helper()
				return viewPlugin(t, "ui", false, view, "/")
			},
		},
		{
			name:  "no browser path granted",
			build: func(t *testing.T) *unstructured.Unstructured { t.Helper(); return viewPlugin(t, "ui", true, view) },
		},
		{
			name: "not a view",
			build: func(t *testing.T) *unstructured.Unstructured {
				t.Helper()
				return viewPlugin(t, "ibkr", true, []string{"connector"}, "/")
			},
		},
		{
			name: "being deleted",
			build: func(t *testing.T) *unstructured.Unstructured {
				t.Helper()
				u := viewPlugin(t, "ui", true, view, "/")
				now := metav1.NewTime(time.Date(2026, 10, 28, 12, 0, 0, 0, time.UTC))
				u.SetDeletionTimestamp(&now)
				return u
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, ok := kube.BrowserRouteOf(tt.build(t))
			if ok != tt.ok || got != tt.want {
				t.Errorf("BrowserRouteOf = %+v, %v; want %+v, %v", got, ok, tt.want, tt.ok)
			}
		})
	}
}

func TestBrowser(t *testing.T) {
	t.Parallel()
	dyn := dynamicfake.NewSimpleDynamicClientWithCustomListKinds(runtime.NewScheme(),
		map[schema.GroupVersionResource]string{kube.PluginResource(): "PluginList"},
		viewPlugin(t, "ui", true, []string{"view"}, "/"),
		viewPlugin(t, "gallery", true, []string{"view"}, "/gallery"),
		viewPlugin(t, "ibkr", true, []string{"connector"}))
	service := func(name, ip string) *corev1.Service {
		return &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name: "plugin-" + name, Namespace: "openfief",
				Labels: map[string]string{"app.kubernetes.io/name": "openfief-plugin", "openfief.io/plugin": name},
			},
			Spec: corev1.ServiceSpec{ClusterIP: ip, Ports: []corev1.ServicePort{{Name: "connect", Port: 8080}}},
		}
	}
	client := fake.NewClientset(service("ui", "198.51.100.10"), service("gallery", corev1.ClusterIPNone))
	b := kube.NewBrowser(dyn, client, "openfief", 0)
	if _, err := b.BrowserRoutes(t.Context()); !errors.Is(err, app.ErrViewsUnknown) {
		t.Errorf("BrowserRoutes before Start = %v, want ErrViewsUnknown", err)
	}
	if err := b.Start(t.Context()); err != nil {
		t.Fatalf("Start: %v", err)
	}
	routes, err := b.BrowserRoutes(t.Context())
	if err != nil {
		t.Fatalf("BrowserRoutes: %v", err)
	}
	want := []views.BrowserRoute{{Instance: "gallery", Prefix: "/gallery"}, {Instance: "ui", Prefix: "/"}}
	if diff := cmp.Diff(want, routes); diff != "" {
		t.Errorf("routes (-want +got):\n%s", diff)
	}
	if got, err := b.BaseURL("ui"); err != nil || got != "http://198.51.100.10:8080" {
		t.Errorf("BaseURL(ui) = %q, %v", got, err)
	}
	if _, err := b.BaseURL("gallery"); !errors.Is(err, kube.ErrNoAddress) {
		t.Errorf("BaseURL of a headless Service = %v, want ErrNoAddress", err)
	}
}
