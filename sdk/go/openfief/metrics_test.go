package openfief_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	collectormetrics "go.opentelemetry.io/proto/otlp/collector/metrics/v1"
	"google.golang.org/protobuf/proto"

	"github.com/arkadiuss/openfief/sdk/go/openfief"
	"github.com/arkadiuss/openfief/sdk/go/telemetry"
)

// staticToken is a TokenSource with one synthetic token.
type staticToken string

func (s staticToken) Token() (string, error) { return string(s), nil }

// fakeCore is the core's plugin port as far as a plugin's metrics report goes:
// it keeps the reports and the path and authorization they came with.
type fakeCore struct {
	mu      sync.Mutex
	reports []*collectormetrics.ExportMetricsServiceRequest
	path    string
	auth    string
}

func (c *fakeCore) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	var req collectormetrics.ExportMetricsServiceRequest
	c.mu.Lock()
	defer c.mu.Unlock()
	if err != nil || proto.Unmarshal(body, &req) != nil {
		http.Error(w, "bad", http.StatusBadRequest)
		return
	}
	c.path, c.auth = r.URL.Path, r.Header.Get("Authorization")
	c.reports = append(c.reports, &req)
}

// received is what a fakeCore was sent.
type received struct {
	reports    []*collectormetrics.ExportMetricsServiceRequest
	path, auth string
}

func (c *fakeCore) got() received {
	c.mu.Lock()
	defer c.mu.Unlock()
	return received{reports: c.reports, path: c.path, auth: c.auth}
}

// countOnce records one client event, with a client address beside the words
// of the vocabulary, through OpenTelemetry's global meter, which StartMetrics
// sets.
func countOnce(t *testing.T) {
	t.Helper()
	events, err := otel.Meter("openfief-ui").Int64Counter(telemetry.MetricUIClientEvents)
	if err != nil {
		t.Fatalf("counter: %v", err)
	}
	events.Add(t.Context(), 1, metric.WithAttributes(attribute.String("kind", "turnstile-success"), attribute.String("client.address", "203.0.113.9")))
}

// TestStartMetrics sets the global meter provider, so the steps run in order
// and no other test in this package sets it.
func TestStartMetrics(t *testing.T) {
	core := &fakeCore{}
	srv := httptest.NewServer(core)
	defer srv.Close()
	t.Setenv(openfief.EnvCoreAddress, strings.TrimPrefix(srv.URL, "http://"))
	t.Setenv(openfief.EnvInstance, "ui")
	t.Setenv(telemetry.EnvMetricsExporter, "")
	t.Setenv(telemetry.EnvSDKDisabled, "")

	stop, err := openfief.StartMetrics(t.Context(), staticToken("synthetic-plugin-token"))
	if err != nil {
		t.Fatalf("StartMetrics: %v", err)
	}
	countOnce(t)
	if shutdownErr := stop(t.Context()); shutdownErr != nil {
		t.Fatalf("shutdown: %v", shutdownErr)
	}
	got := core.got()
	if got.path != telemetry.MetricsPath || got.auth != "Bearer synthetic-plugin-token" {
		t.Errorf("the core got a report at %q with authorization %q, want %s with the plugin's token", got.path, got.auth, telemetry.MetricsPath)
	}
	if len(got.reports) != 1 || len(got.reports[0].GetResourceMetrics()) != 1 {
		t.Fatalf("%d reports, want one with a resource", len(got.reports))
	}
	rm := got.reports[0].GetResourceMetrics()[0]
	if got := rm.GetResource().GetAttributes()[0].GetValue().GetStringValue(); got != "ui" {
		t.Errorf("service name = %q, want the instance", got)
	}
	point := rm.GetScopeMetrics()[0].GetMetrics()[0].GetSum().GetDataPoints()[0]
	if len(point.GetAttributes()) != 1 || point.GetAttributes()[0].GetKey() != "kind" {
		t.Errorf("point attributes = %v, want the vocabulary's kind alone: the plugin redacts before it sends", point.GetAttributes())
	}
}

func TestStartMetricsIsOffWhenAskedAndRefusesWithoutACore(t *testing.T) {
	core := &fakeCore{}
	srv := httptest.NewServer(core)
	defer srv.Close()
	t.Setenv(openfief.EnvCoreAddress, strings.TrimPrefix(srv.URL, "http://"))
	t.Setenv(telemetry.EnvMetricsExporter, "none")
	stop, err := openfief.StartMetrics(t.Context(), staticToken("synthetic-plugin-token"))
	if err != nil {
		t.Fatalf("StartMetrics with the exporter off: %v", err)
	}
	if err := stop(t.Context()); err != nil {
		t.Errorf("shutdown of metrics that are off: %v", err)
	}
	if got := core.got(); len(got.reports) != 0 {
		t.Errorf("%d reports sent with the exporter off, want none", len(got.reports))
	}
	t.Setenv(telemetry.EnvMetricsExporter, "")
	t.Setenv(openfief.EnvCoreAddress, "")
	if _, err := openfief.StartMetrics(t.Context(), staticToken("x")); err == nil {
		t.Error("StartMetrics without a core address succeeded")
	}
}
