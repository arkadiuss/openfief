package kube_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/wait"
	dynamicfake "k8s.io/client-go/dynamic/fake"
	"k8s.io/client-go/kubernetes/fake"

	"github.com/arkadiuss/openfief/internal/ingest/kube"
	"github.com/arkadiuss/openfief/internal/plugins/operator"
)

// plugin returns a Plugin resource named name with the effective
// capabilities caps and the Ready condition ready.
func plugin(name string, ready bool, caps ...string) *unstructured.Unstructured {
	status := "False"
	if ready {
		status = "True"
	}
	effective := make([]any, len(caps))
	for i, c := range caps {
		effective[i] = c
	}
	return &unstructured.Unstructured{Object: map[string]any{
		"apiVersion": "openfief.io/v1alpha1",
		"kind":       "Plugin",
		"metadata":   map[string]any{"name": name, "namespace": "openfief"},
		"status": map[string]any{
			"conditions": []any{map[string]any{"type": "Ready", "status": status}},
			"effective":  map[string]any{"capabilities": effective},
		},
	}}
}

// service returns the operator's Service of plugin name at ip, port 8080.
func service(name, ip string) *corev1.Service {
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: "plugin-" + name, Namespace: "openfief",
			Labels: map[string]string{kube.LabelName: kube.PluginAppName, kube.LabelPlugin: name},
		},
		Spec: corev1.ServiceSpec{ClusterIP: ip, Ports: []corev1.ServicePort{{Name: kube.PortName, Port: 8080}}},
	}
}

func TestDiscovery(t *testing.T) {
	t.Parallel()
	deleting := plugin("old", true, "connector")
	deleting.SetDeletionTimestamp(&metav1.Time{Time: time.Date(2026, time.March, 2, 6, 0, 0, 0, time.UTC)})
	dyn := dynamicfake.NewSimpleDynamicClientWithCustomListKinds(runtime.NewScheme(),
		map[schema.GroupVersionResource]string{kube.PluginResource(): "PluginList"},
		plugin("ibkr-main", true, "connector"),
		plugin("ui", true, "view"),
		plugin("xtb", false, "connector"),
		plugin("nbp", true, "connector", "view"),
		deleting,
	)
	client := fake.NewClientset(service("ibkr-main", "192.0.2.10"), service("xtb", "192.0.2.11"), service("old", "192.0.2.12"))
	d := kube.NewDiscovery(dyn, client, "openfief", 0)
	if _, err := d.ConnectorInstances(t.Context()); !errors.Is(err, kube.ErrNotStarted) {
		t.Errorf("before Start = %v, want ErrNotStarted", err)
	}
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	if err := d.Start(ctx); err != nil {
		t.Fatalf("Start: %v", err)
	}
	got, err := d.ConnectorInstances(ctx)
	if err != nil {
		t.Fatalf("ConnectorInstances: %v", err)
	}
	if diff := cmp.Diff([]string{"ibkr-main"}, got); diff != "" {
		t.Errorf("instances (-want +got):\n%s", diff)
	}
	if url, urlErr := d.BaseURL("ibkr-main"); urlErr != nil || url != "http://192.0.2.10:8080" {
		t.Errorf("BaseURL = %q, %v", url, urlErr)
	}
	if _, urlErr := d.BaseURL("nbp"); !errors.Is(urlErr, kube.ErrNoAddress) {
		t.Errorf("BaseURL of an instance without a Service = %v, want ErrNoAddress", urlErr)
	}
	checkWatch(t, d, client)
}

// checkWatch creates nbp's Service, as the operator would, and checks that
// the watch picks it up.
func checkWatch(t *testing.T, d *kube.Discovery, client *fake.Clientset) {
	t.Helper()
	ctx := t.Context()
	if _, err := client.CoreV1().Services("openfief").Create(ctx, service("nbp", "192.0.2.13"), metav1.CreateOptions{}); err != nil {
		t.Fatalf("creating a Service: %v", err)
	}
	var got []string
	err := wait.PollUntilContextTimeout(ctx, time.Millisecond, 10*time.Second, true, func(ctx context.Context) (bool, error) {
		var listErr error
		got, listErr = d.ConnectorInstances(ctx)
		if listErr != nil {
			return false, fmt.Errorf("listing: %w", listErr)
		}
		return len(got) == 2, nil
	})
	if err != nil {
		t.Fatalf("waiting for nbp: %v (instances %v)", err, got)
	}
	if diff := cmp.Diff([]string{"ibkr-main", "nbp"}, got); diff != "" {
		t.Errorf("instances after the Service appeared (-want +got):\n%s", diff)
	}
}

func TestServiceAddress(t *testing.T) {
	t.Parallel()
	headless := service("a", corev1.ClusterIPNone)
	otherPort := service("a", "192.0.2.10")
	otherPort.Spec.Ports[0].Name = "metrics"
	v6 := service("a", "2001:db8::10")
	tests := []struct {
		name   string
		svc    *corev1.Service
		want   string
		wantOK bool
	}{
		{name: "IPv4", svc: service("a", "192.0.2.10"), want: "192.0.2.10:8080", wantOK: true},
		{name: "IPv6", svc: v6, want: "[2001:db8::10]:8080", wantOK: true},
		{name: "no ClusterIP", svc: service("a", "")},
		{name: "headless", svc: headless},
		{name: "no connect port", svc: otherPort},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, ok := kube.ServiceAddress(tt.svc)
			if got != tt.want || ok != tt.wantOK {
				t.Errorf("ServiceAddress = %q, %v; want %q, %v", got, ok, tt.want, tt.wantOK)
			}
		})
	}
}

// TestOperatorContract checks that the names Discovery relies on are the
// operator's.
func TestOperatorContract(t *testing.T) {
	t.Parallel()
	if kube.LabelPlugin != operator.LabelPlugin || kube.PortName != operator.PortName {
		t.Errorf("label %q and port %q, operator's %q and %q", kube.LabelPlugin, kube.PortName, operator.LabelPlugin, operator.PortName)
	}
}
