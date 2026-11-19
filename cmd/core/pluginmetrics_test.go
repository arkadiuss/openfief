package main

import (
	"bytes"
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	collectormetrics "go.opentelemetry.io/proto/otlp/collector/metrics/v1"
	commonpb "go.opentelemetry.io/proto/otlp/common/v1"
	metricspb "go.opentelemetry.io/proto/otlp/metrics/v1"
	resourcepb "go.opentelemetry.io/proto/otlp/resource/v1"
	"google.golang.org/protobuf/proto"

	identityconnect "github.com/arkadiuss/openfief/internal/identity/connectapi"
	"github.com/arkadiuss/openfief/sdk/go/telemetry"
)

// errCallerDown is a failure to check a caller, which is no refusal.
var errCallerDown = errors.New("the API server is down")

// metricsCallers admits the bearer token "ui" as the view plugin ui, refuses
// "connector" as no view plugin and cannot check "down".
type metricsCallers struct{}

func (metricsCallers) VerifyCaller(_ context.Context, token string) (string, error) {
	switch token {
	case "ui":
		return "ui", nil
	case "down":
		return "", errCallerDown
	}
	return "", identityconnect.ErrCallerRefused
}

// pluginReport is what a view plugin sends: a counter per name whose point
// carries a client address and a session beside the vocabulary's words, under
// a resource that names another service.
func pluginReport(names ...string) []byte {
	str := func(k, v string) *commonpb.KeyValue {
		return &commonpb.KeyValue{Key: k, Value: &commonpb.AnyValue{Value: &commonpb.AnyValue_StringValue{StringValue: v}}}
	}
	metrics := make([]*metricspb.Metric, 0, len(names))
	for _, name := range names {
		metrics = append(metrics, &metricspb.Metric{Name: name, Data: &metricspb.Metric_Sum{Sum: &metricspb.Sum{DataPoints: []*metricspb.NumberDataPoint{{
			Attributes: []*commonpb.KeyValue{str("kind", "turnstile-error"), str("browser", "chrome"), str("client.address", "203.0.113.9"), str("session", "ofsess_abc")},
			Value:      &metricspb.NumberDataPoint_AsInt{AsInt: 1},
		}}}}})
	}
	b, err := proto.Marshal(&collectormetrics.ExportMetricsServiceRequest{ResourceMetrics: []*metricspb.ResourceMetrics{{
		Resource:     &resourcepb.Resource{Attributes: []*commonpb.KeyValue{str("service.name", "somebody-else"), str("host.name", "node-1")}},
		ScopeMetrics: []*metricspb.ScopeMetrics{{Scope: &commonpb.InstrumentationScope{Name: "openfief-ui"}, Metrics: metrics}},
	}}})
	if err != nil {
		panic(err)
	}
	return b
}

// upstream is the core's own OTLP endpoint: a collector that keeps what it is
// sent and answers with code, 200 when it is 0.
type upstream struct {
	mu   sync.Mutex
	got  []*collectormetrics.ExportMetricsServiceRequest
	code int
}

func (u *upstream) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	var req collectormetrics.ExportMetricsServiceRequest
	u.mu.Lock()
	defer u.mu.Unlock()
	if err != nil || r.URL.Path != "/v1/metrics" || proto.Unmarshal(body, &req) != nil {
		http.Error(w, "bad", http.StatusBadRequest)
		return
	}
	u.got = append(u.got, &req)
	if u.code != 0 {
		w.WriteHeader(u.code)
	}
}

// take returns the reports the upstream got since the last take.
func (u *upstream) take() []*collectormetrics.ExportMetricsServiceRequest {
	u.mu.Lock()
	defer u.mu.Unlock()
	out := u.got
	u.got = nil
	return out
}

// reportServer is the plugin port with the metrics path, in front of an
// upstream collector.
type reportServer struct {
	t    *testing.T
	url  string
	up   *upstream
	logs *bytes.Buffer
	coll *httptest.Server
}

func newReportServer(t *testing.T) *reportServer {
	t.Helper()
	up := &upstream{}
	coll := httptest.NewServer(up)
	t.Cleanup(coll.Close)
	fwd, err := telemetry.MetricsForwarderFromEnv(func(k string) string {
		if k == telemetry.EnvEndpoint {
			return coll.URL
		}
		return ""
	})
	if err != nil {
		t.Fatalf("forwarder: %v", err)
	}
	logs := &bytes.Buffer{}
	srv := httptest.NewServer(withPluginMetrics(http.NotFoundHandler(), metricsCallers{}, fwd, slog.New(slog.NewJSONHandler(logs, nil))))
	t.Cleanup(srv.Close)
	return &reportServer{t: t, url: srv.URL, up: up, logs: logs, coll: coll}
}

// post sends a report with token and content type, and returns the status.
func (s *reportServer) post(token, contentType string, body []byte) int {
	s.t.Helper()
	req, err := http.NewRequestWithContext(s.t.Context(), http.MethodPost, s.url+telemetry.MetricsPath, bytes.NewReader(body))
	if err != nil {
		s.t.Fatalf("request: %v", err)
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	req.Header.Set("Content-Type", contentType)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		s.t.Fatalf("post: %v", err)
	}
	if _, err := io.Copy(io.Discard, resp.Body); err != nil {
		s.t.Fatalf("read: %v", err)
	}
	if err := resp.Body.Close(); err != nil {
		s.t.Fatalf("close: %v", err)
	}
	return resp.StatusCode
}

func TestPluginMetricsAreAnsweredAndForwarded(t *testing.T) {
	t.Parallel()
	s := newReportServer(t)
	tests := []struct {
		name        string
		token       string
		contentType string
		body        []byte
		want        int
		forwarded   int
	}{
		{"no token", "", protobufType, pluginReport(telemetry.MetricUIClientEvents), http.StatusUnauthorized, 0},
		{"a caller that is no view plugin", "connector", protobufType, pluginReport(telemetry.MetricUIClientEvents), http.StatusForbidden, 0},
		{"a caller that can't be checked", "down", protobufType, pluginReport(telemetry.MetricUIClientEvents), http.StatusInternalServerError, 0},
		{"a body that isn't protobuf", "ui", "application/json", pluginReport(telemetry.MetricUIClientEvents), http.StatusUnsupportedMediaType, 0},
		{"a body that is no report", "ui", protobufType, []byte("not protobuf \xff\xfe"), http.StatusBadRequest, 0},
		{"a body over the cap", "ui", protobufType, make([]byte, maxMetricsReportBytes+1), http.StatusRequestEntityTooLarge, 0},
		{"a report of nothing allowed", "ui", protobufType, pluginReport("somebody.elses.metric"), http.StatusOK, 0},
		{"a report", "ui", protobufType, pluginReport(telemetry.MetricUIClientEvents, "somebody.elses.metric"), http.StatusOK, 1},
		{"a report with a parameter on its type", "ui", protobufType + "; charset=binary", pluginReport(telemetry.MetricUIClientEvents), http.StatusOK, 1},
	}
	for _, tt := range tests {
		if code := s.post(tt.token, tt.contentType, tt.body); code != tt.want {
			t.Errorf("%s: status %d, want %d", tt.name, code, tt.want)
		}
		if got := len(s.up.take()); got != tt.forwarded {
			t.Errorf("%s: %d reports forwarded, want %d", tt.name, got, tt.forwarded)
		}
	}
}

func TestPluginMetricsAreForwardedFilteredAndNamedByTheVerifiedPlugin(t *testing.T) {
	t.Parallel()
	s := newReportServer(t)
	if code := s.post("ui", protobufType, pluginReport(telemetry.MetricUIClientEvents, "somebody.elses.metric")); code != http.StatusOK {
		t.Fatalf("a good report: status %d", code)
	}
	got := s.up.take()
	if len(got) != 1 || len(got[0].GetResourceMetrics()) != 1 {
		t.Fatalf("forwarded %d reports, want 1 with a resource", len(got))
	}
	rm := got[0].GetResourceMetrics()[0]
	if attrs := rm.GetResource().GetAttributes(); len(attrs) != 1 || attrs[0].GetValue().GetStringValue() != "ui" {
		t.Errorf("resource attributes = %v, want service.name ui, the verified plugin, and nothing else", attrs)
	}
	metrics := rm.GetScopeMetrics()[0].GetMetrics()
	if len(metrics) != 1 || metrics[0].GetName() != telemetry.MetricUIClientEvents {
		t.Fatalf("metrics = %v, want the allowed one alone", metrics)
	}
	attrs := metrics[0].GetSum().GetDataPoints()[0].GetAttributes()
	keys := make([]string, 0, len(attrs))
	for _, kv := range attrs {
		keys = append(keys, kv.GetKey())
	}
	if strings.Join(keys, ",") != "kind,browser" {
		t.Errorf("point attributes = %v, want kind and browser", keys)
	}
}

func TestAnUpstreamThatIsDownDoesNotFailThePlugin(t *testing.T) {
	t.Parallel()
	s := newReportServer(t)
	s.up.mu.Lock()
	s.up.code = http.StatusServiceUnavailable
	s.up.mu.Unlock()
	if code := s.post("ui", protobufType, pluginReport(telemetry.MetricUIClientEvents)); code != http.StatusOK {
		t.Errorf("an upstream that is down: status %d, want 200", code)
	}
	if out := s.logs.String(); !strings.Contains(out, "forwarding a plugin's metrics failed") || !strings.Contains(out, `"plugin":"ui"`) || strings.Contains(out, s.coll.URL) {
		t.Errorf("log %q, want the failure with the plugin and without the endpoint", out)
	}
}

func TestPluginMetricsWithoutAnOTLPEndpointAreDropped(t *testing.T) {
	t.Parallel()
	h := withPluginMetrics(http.NotFoundHandler(), metricsCallers{}, nil, slog.New(slog.DiscardHandler))
	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, telemetry.MetricsPath, bytes.NewReader(pluginReport(telemetry.MetricUIClientEvents)))
	req.Header.Set("Authorization", "Bearer ui")
	req.Header.Set("Content-Type", protobufType)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("status %d, want 200: the report is taken and dropped", rec.Code)
	}
	if rec.Header().Get("Cache-Control") != "no-store" {
		t.Errorf("Cache-Control = %q, want no-store", rec.Header().Get("Cache-Control"))
	}
}

func TestPluginMetricsPathLeavesEveryOtherCallToThePluginPort(t *testing.T) {
	t.Parallel()
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Seen", r.Method+" "+r.URL.Path)
	})
	h := withPluginMetrics(inner, metricsCallers{}, nil, slog.New(slog.DiscardHandler))
	for _, tt := range []struct{ method, path string }{
		{http.MethodGet, telemetry.MetricsPath},
		{http.MethodPost, "/openfief.views.v1.ReadService/ListAccounts"},
		{http.MethodPost, "/otlp/v1/traces"},
	} {
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequestWithContext(t.Context(), tt.method, tt.path, http.NoBody))
		if got := rec.Header().Get("X-Seen"); got != tt.method+" "+tt.path {
			t.Errorf("%s %s went to %q, want the plugin port's handler", tt.method, tt.path, got)
		}
	}
}

func TestPluginMetricsForwarderFromEnv(t *testing.T) {
	t.Parallel()
	env := func(vars map[string]string) func(string) string { return func(k string) string { return vars[k] } }
	if f, err := pluginMetricsForwarder(env(nil)); err != nil || f != nil {
		t.Errorf("without an endpoint: %v, %v; want no forwarder", f, err)
	}
	if f, err := pluginMetricsForwarder(env(map[string]string{telemetry.EnvEndpoint: "http://otel-collector.observability:4318"})); err != nil || f == nil {
		t.Errorf("with an endpoint: %v, %v; want a forwarder", f, err)
	}
	if f, err := pluginMetricsForwarder(env(map[string]string{telemetry.EnvEndpoint: "http://c:4318", telemetry.EnvMetricsExporter: "none"})); err != nil || f != nil {
		t.Errorf("with metrics off: %v, %v; want no forwarder", f, err)
	}
	if _, err := pluginMetricsForwarder(env(map[string]string{telemetry.EnvEndpoint: "collector"})); err == nil {
		t.Error("a bad endpoint was accepted")
	}
}
