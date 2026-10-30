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

// pluginOnlyPaths are one procedure of each service only plugins call: the
// AuthService and the owner services (D-115).
func pluginOnlyPaths() []string {
	return []string{
		"/openfief.identity.v1.AuthService/GetLoginMethods",
		"/openfief.ingest.v1.AccountService/CreateAccount",
		"/openfief.ingest.v1.BondLotService/AddBondLot",
		"/openfief.ingest.v1.StatementService/UploadStatement",
		"/openfief.portfolio.v1.SleeveService/AddSleeveKind",
		"/openfief.portfolio.v1.ModelPortfolioService/GetModelPortfolio",
	}
}

// TestFrontDoor checks D-81 and D-115 over the core's own servers: on the
// browser port, a browser's path the core doesn't serve reaches the view
// plugin ui with the core's token and the browser's Host and scheme, the
// read API and the health endpoints stay the core's, and the services only
// plugins call are not there; the plugin port serves the read API and
// those services, and neither the front door nor the health endpoints.
func TestFrontDoor(t *testing.T) {
	t.Parallel()
	type seen struct{ path, auth, host, proto string }
	got := make(chan seen, 1)
	ui := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got <- seen{path: r.URL.Path, auth: r.Header.Get("Authorization"), host: r.Host, proto: r.Header.Get("X-Forwarded-Proto")}
		w.WriteHeader(http.StatusTeapot)
	}))
	t.Cleanup(ui.Close)
	_, ports, _ := startCorePorts(t, emptyDatabase(t), browserCluster(t, ui.URL), func(*serveConfig) {})
	status := waitRouted(t, ports.browser, "/login")
	if status != http.StatusTeapot {
		t.Fatalf("GET /login through the core = %d, want the UI's 418", status)
	}
	want := seen{path: "/login", auth: "Bearer " + coreToken, host: "localhost:8080", proto: "http"}
	if s := <-got; s != want {
		t.Errorf("the UI saw %+v, want %+v", s, want)
	}
	browserWant := map[string]int{"/v1/accounts": http.StatusUnauthorized, "/healthz": http.StatusOK}
	pluginWant := map[string]int{"/v1/accounts": http.StatusUnauthorized, "/healthz": http.StatusNotFound, "/login": http.StatusNotFound}
	for _, path := range pluginOnlyPaths() {
		browserWant[path] = http.StatusNotFound
		pluginWant[path] = mounted
	}
	checkStatuses(t, "browser", ports.browser, browserWant)
	checkStatuses(t, "plugin", ports.plugin, pluginWant)
	select {
	case s := <-got:
		t.Errorf("the UI got a request for a core path: %+v", s)
	default:
	}
}

// mounted stands for any status but 404 in checkStatuses: the service is
// there, and refuses a plain GET as its protocol does (405, 415).
const mounted = 0

// checkStatuses checks that GET of each path on the core's port srv
// answers the status want gives it.
func checkStatuses(t *testing.T, port string, srv *httptest.Server, want map[string]int) {
	t.Helper()
	for path, wantStatus := range want {
		r := browserGet(t, srv, path)
		switch {
		case wantStatus == mounted && r == http.StatusNotFound:
			t.Errorf("GET %s on the %s port = 404, want the service there", path, port)
		case wantStatus != mounted && r != wantStatus:
			t.Errorf("GET %s on the %s port = %d, want %d", path, port, r, wantStatus)
		}
	}
}

// TestFrontDoorOutsideACluster checks that outside a cluster, with no view
// plugin to proxy to, the core answers a browser's path itself with 404,
// and its browser port still serves no plugin-only service.
func TestFrontDoorOutsideACluster(t *testing.T) {
	t.Parallel()
	_, srv, _ := startCoreWith(t, emptyDatabase(t), func(*serveConfig) {})
	want := map[string]int{"/login": http.StatusNotFound}
	for _, path := range pluginOnlyPaths() {
		want[path] = http.StatusNotFound
	}
	checkStatuses(t, "browser", srv, want)
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
