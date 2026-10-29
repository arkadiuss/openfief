package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"
	"time"

	authnv1 "k8s.io/api/authentication/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	dynamicfake "k8s.io/client-go/dynamic/fake"
	"k8s.io/client-go/kubernetes/fake"
	k8stesting "k8s.io/client-go/testing"

	ingestkube "github.com/arkadiuss/openfief/internal/ingest/kube"
)

// frontDoorTimeout bounds the wait for the core to read its view plugins.
const frontDoorTimeout = 10 * time.Second

// browserCluster is a fake cluster with the view plugin ui installed,
// ready, with its browser path / granted, and served at uiURL; the
// TokenRequest API issues coreToken.
func browserCluster(t *testing.T, uiURL string) *cluster {
	t.Helper()
	u, err := url.Parse(uiURL)
	if err != nil {
		t.Fatal(err)
	}
	port, err := strconv.ParseInt(u.Port(), 10, 32)
	if err != nil {
		t.Fatal(err)
	}
	dyn := dynamicfake.NewSimpleDynamicClientWithCustomListKinds(runtime.NewScheme(),
		map[schema.GroupVersionResource]string{ingestkube.PluginResource(): "PluginList"},
		&unstructured.Unstructured{Object: map[string]any{
			"apiVersion": "openfief.io/v1alpha1", "kind": "Plugin",
			"metadata": map[string]any{"name": "ui", "namespace": "openfief"},
			"status": map[string]any{
				"conditions": []any{map[string]any{"type": "Ready", "status": "True"}},
				"effective":  map[string]any{"capabilities": []any{"view"}, "browserPaths": []any{"/"}},
			},
		}})
	client := fake.NewClientset(&corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: "plugin-ui", Namespace: "openfief",
			Labels: map[string]string{ingestkube.LabelName: ingestkube.PluginAppName, ingestkube.LabelPlugin: "ui"},
		},
		Spec: corev1.ServiceSpec{ClusterIP: u.Hostname(), Ports: []corev1.ServicePort{{Name: ingestkube.PortName, Port: int32(port)}}},
	})
	client.PrependReactor("create", "serviceaccounts", func(k8stesting.Action) (bool, runtime.Object, error) {
		return true, &authnv1.TokenRequest{Status: authnv1.TokenRequestStatus{
			Token: coreToken, ExpirationTimestamp: metav1.NewTime(time.Now().Add(10 * time.Minute)),
		}}, nil
	})
	return &cluster{client: client, dynamic: dyn, namespace: "openfief"}
}

// TestFrontDoor checks D-81 over the core's own server: a browser's path
// the core doesn't serve reaches the view plugin ui with the core's token
// and the browser's Host and scheme, while the read API, the AuthService
// and the health endpoints stay the core's.
func TestFrontDoor(t *testing.T) {
	t.Parallel()
	type seen struct{ path, auth, host, proto string }
	got := make(chan seen, 1)
	ui := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got <- seen{path: r.URL.Path, auth: r.Header.Get("Authorization"), host: r.Host, proto: r.Header.Get("X-Forwarded-Proto")}
		w.WriteHeader(http.StatusTeapot)
	}))
	t.Cleanup(ui.Close)
	_, srv, _ := startCoreIn(t, emptyDatabase(t), browserCluster(t, ui.URL), func(*serveConfig) {})
	status := waitRouted(t, srv, "/login")
	if status != http.StatusTeapot {
		t.Fatalf("GET /login through the core = %d, want the UI's 418", status)
	}
	want := seen{path: "/login", auth: "Bearer " + coreToken, host: "localhost:8080", proto: "http"}
	if s := <-got; s != want {
		t.Errorf("the UI saw %+v, want %+v", s, want)
	}
	for path, wantStatus := range map[string]int{
		"/v1/accounts": http.StatusUnauthorized, "/healthz": http.StatusOK,
		"/openfief.identity.v1.AuthService/GetLoginMethods": http.StatusMethodNotAllowed,
	} {
		if r := browserGet(t, srv, path); r != wantStatus {
			t.Errorf("GET %s = %d, want the core's %d", path, r, wantStatus)
		}
	}
	select {
	case s := <-got:
		t.Errorf("the UI got a request for a core path: %+v", s)
	default:
	}
}

// TestFrontDoorOutsideACluster checks that outside a cluster, with no view
// plugin to proxy to, the core answers a browser's path itself with 404.
func TestFrontDoorOutsideACluster(t *testing.T) {
	t.Parallel()
	_, srv, _ := startCoreWith(t, emptyDatabase(t), func(*serveConfig) {})
	if r := browserGet(t, srv, "/login"); r != http.StatusNotFound {
		t.Errorf("GET /login outside a cluster = %d, want 404", r)
	}
}

// waitRouted asks the core for path until it no longer answers 503 while
// it reads its view plugins, and returns the status it answered.
func waitRouted(t *testing.T, srv *httptest.Server, path string) int {
	t.Helper()
	ctx, cancel := context.WithTimeout(t.Context(), frontDoorTimeout)
	defer cancel()
	tick := time.NewTicker(20 * time.Millisecond)
	defer tick.Stop()
	for {
		if status := browserGet(t, srv, path); status != http.StatusServiceUnavailable {
			return status
		}
		select {
		case <-ctx.Done():
			t.Fatalf("the core did not route %s within %s", path, frontDoorTimeout)
		case <-tick.C:
		}
	}
}

// browserGet sends GET path to the core as a browser through kubectl
// port-forward to localhost:8080 would, and returns the status.
func browserGet(t *testing.T, srv *httptest.Server, path string) int {
	t.Helper()
	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, srv.URL+path, http.NoBody)
	if err != nil {
		t.Fatal(err)
	}
	req.Host = "localhost:8080"
	resp, err := srv.Client().Do(req)
	if err != nil {
		t.Fatal(err)
	}
	if err := resp.Body.Close(); err != nil {
		t.Fatal(err)
	}
	return resp.StatusCode
}
