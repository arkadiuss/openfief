package telemetry_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"slices"
	"sync"
	"testing"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
	collectormetrics "go.opentelemetry.io/proto/otlp/collector/metrics/v1"
	collectortrace "go.opentelemetry.io/proto/otlp/collector/trace/v1"
	commonpb "go.opentelemetry.io/proto/otlp/common/v1"
	tracepb "go.opentelemetry.io/proto/otlp/trace/v1"
	"google.golang.org/protobuf/proto"

	"github.com/arkadiuss/openfief/sdk/go/telemetry"
)

// errSynthetic is a failure whose message quotes data, as real errors may.
var errSynthetic = errors.New("price 1234.56 PLN for PL0000000000")

// collector is an OTLP/HTTP receiver that keeps the spans and the metrics it
// was sent, and the authorization header of the last metrics post.
type collector struct {
	mu      sync.Mutex
	spans   []*tracepb.Span
	metrics []*collectormetrics.ExportMetricsServiceRequest
	auth    string
	posts   int
}

// ServeHTTP decodes an export request: spans at /v1/traces, metrics at
// /v1/metrics, or at /otlp/v1/metrics, where a plugin sends them to the core.
func (c *collector) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "bad export", http.StatusBadRequest)
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	switch r.URL.Path {
	case "/v1/traces":
		var req collectortrace.ExportTraceServiceRequest
		if proto.Unmarshal(body, &req) != nil {
			http.Error(w, "bad export", http.StatusBadRequest)
			return
		}
		c.posts++
		for _, rs := range req.GetResourceSpans() {
			for _, ss := range rs.GetScopeSpans() {
				c.spans = append(c.spans, ss.GetSpans()...)
			}
		}
	case "/v1/metrics", "/otlp/v1/metrics":
		var req collectormetrics.ExportMetricsServiceRequest
		if proto.Unmarshal(body, &req) != nil {
			http.Error(w, "bad export", http.StatusBadRequest)
			return
		}
		c.posts++
		c.auth = r.Header.Get("Authorization")
		c.metrics = append(c.metrics, &req)
	default:
		http.Error(w, "bad export", http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/x-protobuf")
}

// metricsRequests returns the metrics requests received so far.
func (c *collector) metricsRequests() []*collectormetrics.ExportMetricsServiceRequest {
	c.mu.Lock()
	defer c.mu.Unlock()
	return slices.Clone(c.metrics)
}

// lastAuth returns the authorization header of the last metrics request.
func (c *collector) lastAuth() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.auth
}

// received returns the spans received so far by name.
func (c *collector) received() map[string]*tracepb.Span {
	c.mu.Lock()
	defer c.mu.Unlock()
	out := map[string]*tracepb.Span{}
	for _, s := range c.spans {
		out[s.GetName()] = s
	}
	return out
}

// startWith starts tracing with vars as the environment.
func startWith(t *testing.T, vars map[string]string) *telemetry.Telemetry {
	t.Helper()
	cfg, err := telemetry.ParseConfig(func(name string) string { return vars[name] }, "openfief-test")
	if err != nil {
		t.Fatalf("ParseConfig: %v", err)
	}
	tel, err := telemetry.Start(t.Context(), &cfg)
	if err != nil {
		t.Fatalf("Start: %v", err)
	}
	return tel
}

// TestStart is the only test that sets the global tracer provider and
// propagator, so the steps run in order.
func TestStart(t *testing.T) {
	t.Parallel()

	checkOff(t)

	recv := &collector{}
	srv := httptest.NewServer(recv)
	defer srv.Close()
	tel := startWith(t, map[string]string{telemetry.EnvEndpoint: srv.URL})
	counter := countLeakyEvents(t)
	ctx, root := otel.Tracer("test").Start(t.Context(), "root")
	inner := traceSeen(ctx, t, telemetry.ServerHandler, http.MethodPost, "/openfief.connector.v1.ConnectorService/Sync")
	if inner != root.SpanContext().TraceID() {
		t.Errorf("ServerHandler continued trace %s, want the caller's %s", inner, root.SpanContext().TraceID())
	}
	public := traceSeen(ctx, t, telemetry.PublicHandler, http.MethodPost, "/v1/accounts/acc-1")
	if public == root.SpanContext().TraceID() || !public.IsValid() {
		t.Errorf("PublicHandler continued the caller's trace %s, want a new one", public)
	}
	if probe := traceSeen(ctx, t, telemetry.ServerHandler, http.MethodGet, "/healthz"); probe.IsValid() {
		t.Errorf("a health probe ran in trace %s, want none", probe)
	}
	ran := false
	telemetry.Run(ctx, "test", "ingest.sync", func(context.Context) error { ran = true; return errSynthetic },
		telemetry.KeyInstance.String("nbp-fx"), attribute.String("openfief.amount", "1234.56"))
	if !ran {
		t.Error("Run did not run fn")
	}
	root.End()
	if recv.received()["root"] != nil {
		t.Error("a span was sent before the batch was due or shutdown flushed it")
	}
	if err := tel.Shutdown(t.Context()); err != nil {
		t.Fatalf("Shutdown: %v", err)
	}
	checkExported(t, recv.received())
	checkMetricsExported(t, recv, "openfief-test", counter)
	checkPluginMetrics(t)
	grpc := startWith(t, map[string]string{telemetry.EnvEndpoint: "http://127.0.0.1:4317", telemetry.EnvProtocol: "grpc"})
	if err := grpc.Shutdown(t.Context()); err != nil {
		t.Errorf("Shutdown of a gRPC exporter that sent nothing: %v", err)
	}
}

// metricCounter is what countLeakyEvents counted.
type metricCounter struct{ allowed, other int64 }

// countLeakyEvents records, through the global meter provider Start installed,
// an allowed counter whose data point carries a client address, a session and
// a question beside the words of the vocabulary, and a counter nobody
// allowed.
func countLeakyEvents(t *testing.T) metricCounter {
	t.Helper()
	m := otel.Meter("openfief-ui")
	events, err := m.Int64Counter(telemetry.MetricUIClientEvents)
	if err != nil {
		t.Fatalf("counter: %v", err)
	}
	other, err := m.Int64Counter("http.server.request.count")
	if err != nil {
		t.Fatalf("counter: %v", err)
	}
	events.Add(t.Context(), 2, metric.WithAttributes(attribute.String("kind", "turnstile-success"), attribute.String("browser", "chrome"),
		attribute.String("client.address", "203.0.113.9"), attribute.String("session", "ofsess_abc"), attribute.String("question", "my balance")))
	other.Add(t.Context(), 5)
	return metricCounter{allowed: 2, other: 5}
}

// bearer adds an Authorization header to every request, as a plugin's client
// for the core does.
type bearer struct{ next http.RoundTripper }

func (b bearer) RoundTrip(r *http.Request) (*http.Response, error) {
	r = r.Clone(r.Context())
	r.Header.Set("Authorization", "Bearer synthetic-plugin-token")
	return b.next.RoundTrip(r) //nolint:wrapcheck // a test transport
}

// checkPluginMetrics checks a plugin's way: StartMetricsTo sends the metrics,
// redacted, to the core's address with the client it is given, which carries
// the plugin's token, and to nothing else.
func checkPluginMetrics(t *testing.T) {
	t.Helper()
	core := &collector{}
	srv := httptest.NewServer(core)
	defer srv.Close()
	tel, err := telemetry.StartMetricsTo(t.Context(), srv.URL+"/otlp/v1/metrics", &http.Client{Transport: bearer{http.DefaultTransport}}, "ui")
	if err != nil {
		t.Fatalf("StartMetricsTo: %v", err)
	}
	counted := countLeakyEvents(t)
	if shutdownErr := tel.Shutdown(t.Context()); shutdownErr != nil {
		t.Fatalf("Shutdown: %v", shutdownErr)
	}
	checkMetricsExported(t, core, "ui", counted)
	if auth := core.lastAuth(); auth != "Bearer synthetic-plugin-token" {
		t.Errorf("the core got authorization %q, want the plugin's token", auth)
	}
}

// checkMetricsExported checks the metrics a collector received: one request,
// whose resource is the service name alone, with the allowed counter alone,
// whose point has the vocabulary's attributes alone and the value counted, and
// not the counter nobody allowed.
func checkMetricsExported(t *testing.T, recv *collector, service string, counted metricCounter) {
	t.Helper()
	reqs := recv.metricsRequests()
	if len(reqs) != 1 || len(reqs[0].GetResourceMetrics()) != 1 {
		t.Fatalf("%d metrics requests, want one with a resource", len(reqs))
	}
	rm := reqs[0].GetResourceMetrics()[0]
	if attrs := rm.GetResource().GetAttributes(); len(attrs) != 1 || attrs[0].GetKey() != "service.name" || attrs[0].GetValue().GetStringValue() != service {
		t.Errorf("resource attributes = %v, want the service name %q alone", attrs, service)
	}
	m := rm.GetScopeMetrics()[0].GetMetrics()
	if len(m) != 1 || m[0].GetName() != telemetry.MetricUIClientEvents {
		t.Fatalf("metrics = %v, want the allowed counter alone", m)
	}
	point := m[0].GetSum().GetDataPoints()[0]
	if keys := attributeKeys(point.GetAttributes()); !slices.Equal(keys, []string{"browser", "kind"}) || point.GetAsInt() != counted.allowed {
		t.Errorf("point = %v, want browser and kind alone, counting %d", point, counted.allowed)
	}
}

// attributeKeys returns the sorted keys of attrs.
func attributeKeys(attrs []*commonpb.KeyValue) []string {
	keys := make([]string, 0, len(attrs))
	for _, kv := range attrs {
		keys = append(keys, kv.GetKey())
	}
	slices.Sort(keys)
	return keys
}

// checkOff starts tracing without an endpoint and checks that no span
// records and that shutting down does nothing.
func checkOff(t *testing.T) {
	t.Helper()
	off := startWith(t, nil)
	_, span := otel.Tracer("test").Start(t.Context(), "off")
	if span.IsRecording() {
		t.Error("with no endpoint, a span records")
	}
	span.End()
	if err := off.Shutdown(t.Context()); err != nil {
		t.Errorf("Shutdown with export off: %v", err)
	}
}

// traceSeen sends a method request for path, through Transport within
// ctx's span, to a handler wrapped by wrap and returns the trace the
// handler ran in.
func traceSeen(ctx context.Context, t *testing.T, wrap func(http.Handler) http.Handler, method, path string) trace.TraceID {
	t.Helper()
	var seen trace.TraceID
	srv := httptest.NewServer(wrap(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		seen = trace.SpanContextFromContext(r.Context()).TraceID()
	})))
	defer srv.Close()
	req, err := http.NewRequestWithContext(ctx, method, srv.URL+path, http.NoBody)
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	resp, err := (&http.Client{Transport: telemetry.Transport(nil)}).Do(req)
	if err != nil {
		t.Fatalf("call: %v", err)
	}
	if err := resp.Body.Close(); err != nil {
		t.Fatalf("close: %v", err)
	}
	return seen
}

// checkExported checks what the collector received: every span, named by
// operation, with only the allowed attributes and no error message.
func checkExported(t *testing.T, got map[string]*tracepb.Span) {
	t.Helper()
	for _, name := range []string{"root", "ingest.sync", "openfief.connector.v1.ConnectorService/Sync", "POST"} {
		if got[name] == nil {
			t.Errorf("no span %q was exported; got %v", name, keys(got))
		}
	}
	if run, root := got["ingest.sync"], got["root"]; run != nil && root != nil {
		checkDomainSpan(t, run, root)
	}
}

// checkDomainSpan checks the span Begin started: an error without a
// message, the instance alone, a child of root.
func checkDomainSpan(t *testing.T, run, root *tracepb.Span) {
	t.Helper()
	if run.GetStatus().GetCode() != tracepb.Status_STATUS_CODE_ERROR || run.GetStatus().GetMessage() != "" {
		t.Errorf("ingest.sync status = %v, want an error without a message", run.GetStatus())
	}
	attrs := run.GetAttributes()
	if len(attrs) != 1 || attrs[0].GetKey() != string(telemetry.KeyInstance) || attrs[0].GetValue().GetStringValue() != "nbp-fx" {
		t.Errorf("ingest.sync attributes = %v, want the instance alone", attrs)
	}
	if !bytes.Equal(run.GetParentSpanId(), root.GetSpanId()) {
		t.Error("ingest.sync is not a child of the span it ran in")
	}
}

// keys returns the names of got.
func keys(got map[string]*tracepb.Span) []string {
	out := make([]string, 0, len(got))
	for k := range got {
		out = append(out, k)
	}
	return out
}
