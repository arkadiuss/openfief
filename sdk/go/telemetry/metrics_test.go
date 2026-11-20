package telemetry

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/google/go-cmp/cmp"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/instrumentation"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	"go.opentelemetry.io/otel/sdk/resource"
	collectormetrics "go.opentelemetry.io/proto/otlp/collector/metrics/v1"
	commonpb "go.opentelemetry.io/proto/otlp/common/v1"
	metricspb "go.opentelemetry.io/proto/otlp/metrics/v1"
	resourcepb "go.opentelemetry.io/proto/otlp/resource/v1"
	"google.golang.org/protobuf/proto"
)

// errCollectorDown is a collector that fails.
var errCollectorDown = errors.New("collector down")

func TestParseConfigMetrics(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		vars map[string]string
		want string
	}{
		{"off by default", nil, ""},
		{"a base endpoint gets the metrics path", map[string]string{EnvEndpoint: "http://collector.observability:4318"}, "http://collector.observability:4318/v1/metrics"},
		{"a base endpoint with a path and a slash", map[string]string{EnvEndpoint: "https://otlp.example.org/otlp/"}, "https://otlp.example.org/otlp/v1/metrics"},
		{
			"a metrics endpoint is used as it is and wins",
			map[string]string{EnvEndpoint: "http://ignored:4318", EnvMetricsEndpoint: "https://otlp.example.org/custom/metrics"},
			"https://otlp.example.org/custom/metrics",
		},
		{"grpc has no metrics path of its own", map[string]string{EnvEndpoint: "http://collector:4317", EnvProtocol: "grpc"}, ""},
		{
			"grpc with a metrics endpoint sends metrics to it over http",
			map[string]string{EnvEndpoint: "http://collector:4317", EnvProtocol: "grpc", EnvMetricsEndpoint: "http://collector:4318/v1/metrics"},
			"http://collector:4318/v1/metrics",
		},
		{"the none exporter turns metrics off", map[string]string{EnvEndpoint: "http://collector:4318", EnvMetricsExporter: "none"}, ""},
		{"OTEL_SDK_DISABLED turns metrics off", map[string]string{EnvEndpoint: "http://collector:4318", EnvSDKDisabled: "true"}, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := ParseConfig(env(tt.vars), "openfief-core")
			if err != nil {
				t.Fatalf("ParseConfig: %v", err)
			}
			if got.MetricsEndpoint != tt.want || got.MetricsEnabled() != (tt.want != "") {
				t.Errorf("MetricsEndpoint = %q, enabled %v; want %q", got.MetricsEndpoint, got.MetricsEnabled(), tt.want)
			}
			if off := MetricsOff(env(tt.vars)); off != (tt.vars[EnvMetricsExporter] == "none" || tt.vars[EnvSDKDisabled] == "true") {
				t.Errorf("MetricsOff = %v for %v", off, tt.vars)
			}
		})
	}
}

func TestParseConfigMetricsRefuses(t *testing.T) {
	t.Parallel()
	if _, err := ParseConfig(env(map[string]string{EnvMetricsExporter: "console"}), "x"); !errors.Is(err, ErrBadExporter) {
		t.Errorf("a console metrics exporter: error %v, want ErrBadExporter", err)
	}
	if _, err := ParseConfig(env(map[string]string{EnvMetricsEndpoint: "collector:4318"}), "x"); !errors.Is(err, ErrBadEndpoint) {
		t.Errorf("a metrics endpoint without a scheme: error %v, want ErrBadEndpoint", err)
	}
	if !MetricsOff(env(map[string]string{EnvMetricsExporter: "console"})) {
		t.Error("a metrics exporter it can't read does not count as off")
	}
}

// fakeExporter keeps what it is asked to export. Its other methods are the
// embedded exporter's, which these tests don't call.
type fakeExporter struct {
	sdkmetric.Exporter

	mu   sync.Mutex
	got  []*metricdata.ResourceMetrics
	fail error
}

func (f *fakeExporter) Export(_ context.Context, rm *metricdata.ResourceMetrics) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.got = append(f.got, rm)
	return f.fail
}

// leaky is a data point's attributes as a page's report could leave them:
// the vocabulary's, and what must never leave.
func leaky() attribute.Set {
	return attribute.NewSet(
		attribute.String("kind", "turnstile-error"), attribute.String("browser", "chrome"), attribute.String("code", "110200"),
		attribute.String("client.address", "203.0.113.9"), attribute.String("session", "ofsess_abc"),
		attribute.String("question", "what is my balance"), attribute.String("path", strings.Repeat("x", 200)),
	)
}

// leakyMetrics is a process's metrics as they come out of the SDK: a resource
// with the node's name and a pod's address, a scope with a version, the
// allowed counter and histogram and gauge, a metric nobody allowed, a kind of
// data nobody handles, and a scope of nothing allowed.
func leakyMetrics() *metricdata.ResourceMetrics {
	scope := instrumentation.Scope{Name: "openfief-ui", Version: "1.2.3", SchemaURL: "https://example.org/schema"}
	return &metricdata.ResourceMetrics{
		Resource: resource.NewSchemaless(attribute.String("service.name", "ui"), attribute.String("host.name", "node-1"), attribute.String("k8s.pod.ip", "10.42.0.9")),
		ScopeMetrics: []metricdata.ScopeMetrics{
			{Scope: scope, Metrics: []metricdata.Metrics{
				{
					Name: MetricUIClientEvents, Description: "events", Unit: "1",
					Data: metricdata.Sum[int64]{
						IsMonotonic: true, Temporality: metricdata.CumulativeTemporality,
						DataPoints: []metricdata.DataPoint[int64]{{
							Attributes: leaky(), Value: 3,
							Exemplars: []metricdata.Exemplar[int64]{{FilteredAttributes: []attribute.KeyValue{attribute.String("client.address", "203.0.113.9")}, Value: 1}},
						}},
					},
				},
				{
					Name: MetricUITurnstileTime, Unit: "s",
					Data: metricdata.Histogram[float64]{
						Temporality: metricdata.CumulativeTemporality,
						DataPoints:  []metricdata.HistogramDataPoint[float64]{{Attributes: leaky(), Count: 1, Sum: 2}},
					},
				},
				{
					Name: MetricAuthzAvailable,
					Data: metricdata.Gauge[int64]{DataPoints: []metricdata.DataPoint[int64]{{Attributes: attribute.NewSet(attribute.String("to", "available")), Value: 1}}},
				},
				{Name: "process.runtime.go.mem.heap_alloc", Data: metricdata.Gauge[int64]{DataPoints: []metricdata.DataPoint[int64]{{Value: 1}}}},
				{Name: MetricMCPUnauthenticated, Data: metricdata.Summary{}},
			}},
			{Scope: scope, Metrics: []metricdata.Metrics{{Name: "http.server.duration", Data: metricdata.Sum[int64]{}}}},
		},
	}
}

func TestRedactMetrics(t *testing.T) {
	t.Parallel()
	fake := &fakeExporter{}
	if err := RedactMetrics(fake).Export(t.Context(), leakyMetrics()); err != nil {
		t.Fatalf("Export: %v", err)
	}
	if len(fake.got) != 1 {
		t.Fatalf("%d exports, want 1", len(fake.got))
	}
	out := fake.got[0]
	wantResource := attribute.NewSet(attribute.String("service.name", "ui"))
	if gotResource := out.Resource.Set(); !gotResource.Equals(&wantResource) {
		t.Errorf("resource attributes = %v, want the service name alone", out.Resource.Attributes())
	}
	if len(out.ScopeMetrics) != 1 || out.ScopeMetrics[0].Scope.Name != "openfief-ui" || out.ScopeMetrics[0].Scope.Version != "" {
		t.Fatalf("scopes = %+v, want the one scope with its name alone", out.ScopeMetrics)
	}
	metrics := out.ScopeMetrics[0].Metrics
	names := make([]string, 0, len(metrics))
	for _, m := range metrics {
		names = append(names, m.Name)
	}
	if diff := cmp.Diff([]string{MetricUIClientEvents, MetricUITurnstileTime, MetricAuthzAvailable}, names); diff != "" {
		t.Errorf("metrics kept (-want +got):\n%s", diff)
	}
	checkRedactedPoints(t, metrics)
}

// checkRedactedPoints checks the counter's and the histogram's data points:
// the vocabulary's attributes alone, the value, and no exemplar.
func checkRedactedPoints(t *testing.T, metrics []metricdata.Metrics) {
	t.Helper()
	want := attribute.NewSet(attribute.String("kind", "turnstile-error"), attribute.String("browser", "chrome"), attribute.String("code", "110200"))
	sum, ok := metrics[0].Data.(metricdata.Sum[int64])
	if !ok || len(sum.DataPoints) != 1 {
		t.Fatalf("the counter = %#v, want one sum point", metrics[0].Data)
	}
	if got := sum.DataPoints[0]; !got.Attributes.Equals(&want) || got.Value != 3 || len(got.Exemplars) != 0 {
		t.Errorf("counter point = %+v, want the vocabulary's attributes alone, the value and no exemplar", got)
	}
	hist, ok := metrics[1].Data.(metricdata.Histogram[float64])
	if !ok || len(hist.DataPoints) != 1 || !hist.DataPoints[0].Attributes.Equals(&want) {
		t.Errorf("histogram = %#v, want the vocabulary's attributes alone", metrics[1].Data)
	}
}

func TestRedactMetricsExportsNothingWhenNothingIsLeft(t *testing.T) {
	t.Parallel()
	fake := &fakeExporter{}
	in := &metricdata.ResourceMetrics{ScopeMetrics: []metricdata.ScopeMetrics{{
		Metrics: []metricdata.Metrics{{Name: "http.server.duration", Data: metricdata.Sum[int64]{}}},
	}}}
	if err := RedactMetrics(fake).Export(t.Context(), in); err != nil {
		t.Fatalf("Export: %v", err)
	}
	if len(fake.got) != 0 {
		t.Errorf("%d exports of metrics that all must stay, want none", len(fake.got))
	}
}

func TestRedactMetricsReturnsTheExportersError(t *testing.T) {
	t.Parallel()
	failing := &fakeExporter{fail: errCollectorDown}
	in := &metricdata.ResourceMetrics{ScopeMetrics: []metricdata.ScopeMetrics{{
		Metrics: []metricdata.Metrics{{Name: MetricMCPUnauthenticated, Data: metricdata.Sum[int64]{}}},
	}}}
	if err := RedactMetrics(failing).Export(t.Context(), in); !errors.Is(err, errCollectorDown) {
		t.Errorf("Export error = %v, want the exporter's", err)
	}
}

// receivedRequest is a plugin's report as it arrives at the core: the plugin
// names itself and sends what it likes.
func receivedRequest() *collectormetrics.ExportMetricsServiceRequest {
	str := func(k, v string) *commonpb.KeyValue {
		return &commonpb.KeyValue{Key: k, Value: &commonpb.AnyValue{Value: &commonpb.AnyValue_StringValue{StringValue: v}}}
	}
	point := func() *metricspb.NumberDataPoint {
		return &metricspb.NumberDataPoint{
			Attributes: []*commonpb.KeyValue{
				str("kind", "turnstile-success"), str("browser", "firefox"), str("client.address", "203.0.113.9"),
				str("session", "ofsess_abc"), str("code", strings.Repeat("9", 100)),
			},
			Value:     &metricspb.NumberDataPoint_AsInt{AsInt: 4},
			Exemplars: []*metricspb.Exemplar{{TraceId: []byte("0123456789abcdef"), FilteredAttributes: []*commonpb.KeyValue{str("question", "q")}}},
		}
	}
	sum := func() *metricspb.Metric_Sum {
		return &metricspb.Metric_Sum{Sum: &metricspb.Sum{DataPoints: []*metricspb.NumberDataPoint{point()}}}
	}
	return &collectormetrics.ExportMetricsServiceRequest{ResourceMetrics: []*metricspb.ResourceMetrics{
		{
			Resource: &resourcepb.Resource{Attributes: []*commonpb.KeyValue{str("service.name", "authz-admin"), str("host.name", "node-1")}},
			ScopeMetrics: []*metricspb.ScopeMetrics{{
				Scope: &commonpb.InstrumentationScope{Name: "openfief-ui", Version: "9", Attributes: []*commonpb.KeyValue{str("a", "b")}},
				Metrics: []*metricspb.Metric{
					{Name: MetricUIClientEvents, Metadata: []*commonpb.KeyValue{str("m", "d")}, Data: sum()},
					{Name: MetricAuthzAvailable, Data: &metricspb.Metric_Gauge{Gauge: &metricspb.Gauge{DataPoints: []*metricspb.NumberDataPoint{point()}}}},
					{Name: "somebody.elses.metric", Data: sum()},
					{Name: MetricMCPUnauthenticated, Data: &metricspb.Metric_Summary{Summary: &metricspb.Summary{}}},
				},
			}},
		},
		{ScopeMetrics: []*metricspb.ScopeMetrics{{Metrics: []*metricspb.Metric{{Name: "only.unknown"}}}}},
	}}
}

func TestFilterOTLPMetrics(t *testing.T) {
	t.Parallel()
	req := receivedRequest()
	if got := FilterOTLPMetrics(req, "ui"); got != 2 {
		t.Errorf("FilterOTLPMetrics kept %d metrics, want 2", got)
	}
	if len(req.GetResourceMetrics()) != 1 {
		t.Fatalf("%d resources left, want the one with metrics", len(req.GetResourceMetrics()))
	}
	rm := req.GetResourceMetrics()[0]
	if attrs := rm.GetResource().GetAttributes(); len(attrs) != 1 || attrs[0].GetKey() != "service.name" || attrs[0].GetValue().GetStringValue() != "ui" {
		t.Errorf("resource attributes = %v, want the verified service name alone, not the plugin's word", attrs)
	}
	checkFilteredScope(t, rm.GetScopeMetrics()[0])
}

// checkFilteredScope checks the scope of receivedRequest after the filter: its
// name alone, and the two allowed metrics.
func checkFilteredScope(t *testing.T, scope *metricspb.ScopeMetrics) {
	t.Helper()
	if scope.GetScope().GetName() != "openfief-ui" || scope.GetScope().GetVersion() != "" || len(scope.GetScope().GetAttributes()) != 0 {
		t.Errorf("scope = %v, want its name alone", scope.GetScope())
	}
	if len(scope.GetMetrics()) != 2 {
		t.Fatalf("%d metrics left, want 2", len(scope.GetMetrics()))
	}
	for _, m := range scope.GetMetrics() {
		checkFilteredMetric(t, m)
	}
}

func TestFilterOTLPMetricsOfNothing(t *testing.T) {
	t.Parallel()
	if got := FilterOTLPMetrics(&collectormetrics.ExportMetricsServiceRequest{}, "ui"); got != 0 {
		t.Errorf("an empty request kept %d metrics", got)
	}
}

// checkFilteredMetric checks a metric of receivedRequest after the filter: no
// metadata, the vocabulary's attributes alone, no exemplar, the value.
func checkFilteredMetric(t *testing.T, m *metricspb.Metric) {
	t.Helper()
	if m.GetMetadata() != nil {
		t.Errorf("metric %s kept its metadata", m.GetName())
	}
	p := m.GetGauge().GetDataPoints()
	if sum := m.GetSum(); sum != nil {
		p = sum.GetDataPoints()
	}
	keys := make([]string, 0, len(p[0].GetAttributes()))
	for _, kv := range p[0].GetAttributes() {
		keys = append(keys, kv.GetKey())
	}
	if diff := cmp.Diff([]string{"kind", "browser"}, keys); diff != "" {
		t.Errorf("%s attributes (-want +got):\n%s", m.GetName(), diff)
	}
	if len(p[0].GetExemplars()) != 0 || p[0].GetAsInt() != 4 {
		t.Errorf("%s point = %v, want no exemplar and the value kept", m.GetName(), p[0])
	}
}

// collectorAnswering is an OTLP/HTTP collector that keeps the last report and
// the headers it came with, and answers with code.
type collectorAnswering struct {
	mu   sync.Mutex
	got  collectormetrics.ExportMetricsServiceRequest
	auth string
	path string
	code int
}

func (c *collectorAnswering) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	c.mu.Lock()
	defer c.mu.Unlock()
	c.auth, c.path = r.Header.Get("Authorization"), r.URL.Path
	if err != nil || proto.Unmarshal(body, &c.got) != nil {
		http.Error(w, "bad", http.StatusBadRequest)
		return
	}
	w.WriteHeader(c.code)
}

func TestMetricsForwarder(t *testing.T) {
	t.Parallel()
	recv := &collectorAnswering{code: http.StatusOK}
	srv := httptest.NewServer(recv)
	defer srv.Close()
	f, err := MetricsForwarderFromEnv(env(map[string]string{EnvEndpoint: srv.URL, EnvHeaders: "authorization=Basic%20c3ludGhldGlj"}))
	if err != nil || f == nil {
		t.Fatalf("MetricsForwarderFromEnv = %v, %v; want a forwarder", f, err)
	}
	req := receivedRequest()
	FilterOTLPMetrics(req, "ui")
	if forwardErr := f.Forward(t.Context(), req); forwardErr != nil {
		t.Fatalf("Forward: %v", forwardErr)
	}
	recv.mu.Lock()
	if recv.path != "/v1/metrics" || recv.auth != "Basic c3ludGhldGlj" || len(recv.got.GetResourceMetrics()) != 1 {
		t.Errorf("collector got path %q, authorization %q and %d resources; want /v1/metrics, the configured header and the one filtered resource",
			recv.path, recv.auth, len(recv.got.GetResourceMetrics()))
	}
	recv.code = http.StatusServiceUnavailable
	recv.mu.Unlock()
	err = f.Forward(t.Context(), req)
	if !errors.Is(err, ErrForward) {
		t.Fatalf("Forward to a collector that answers 503: error %v, want ErrForward", err)
	}
	if strings.Contains(err.Error(), srv.URL) {
		t.Errorf("error %q quotes the endpoint, which may hold a credential", err)
	}
}

func TestNoMetricsForwarderDropsTheReport(t *testing.T) {
	t.Parallel()
	var off *MetricsForwarder
	if err := off.Forward(t.Context(), receivedRequest()); err != nil {
		t.Errorf("a nil forwarder: %v, want it to drop the metrics", err)
	}
	none, err := MetricsForwarderFromEnv(env(nil))
	if err != nil || none != nil {
		t.Errorf("without an endpoint the forwarder is %v, %v; want nil", none, err)
	}
	if _, err := MetricsForwarderFromEnv(env(map[string]string{EnvEndpoint: "collector"})); !errors.Is(err, ErrBadEndpoint) {
		t.Errorf("a bad endpoint: error %v, want ErrBadEndpoint", err)
	}
}
