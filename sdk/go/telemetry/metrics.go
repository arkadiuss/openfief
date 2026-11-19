package telemetry

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
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

// OpenFief's metrics: the only instruments whose data may leave a process
// (D-172), by name, as a span's attributes are allow-listed (Redact). Each
// counts or times what happened and never what it carried (GR-12).
const (
	// MetricUIClientEvents counts the events the UI's pages reported about
	// themselves, by kind, browser family and a Turnstile error's code (D-227).
	MetricUIClientEvents = "openfief.ui.client_events"
	// MetricUITurnstileTime is the time from a Turnstile check starting to its
	// token arriving, by browser family and path (D-227).
	MetricUITurnstileTime = "openfief.ui.turnstile.token_seconds"
	// MetricMCPUnauthenticated counts the MCP endpoint's requests refused
	// with 401.
	MetricMCPUnauthenticated = "openfief.mcp.unauthenticated"
	// MetricAuthzChanges counts how often the authorization model became
	// available or unavailable, by the state it changed to.
	MetricAuthzChanges = "openfief.authz.availability.changes"
	// MetricAuthzAvailable is 1 while the authorization model answers, 0
	// while it can't.
	MetricAuthzAvailable = "openfief.authz.available"
)

// The attribute keys a data point of those metrics may carry out of the
// process. Each is a word of a fixed vocabulary, never text a person typed.
const (
	// metricKeyKind is a client event's kind, a word of D-227's vocabulary.
	metricKeyKind = attribute.Key("kind")
	// metricKeyBrowser is a browser family, such as chrome.
	metricKeyBrowser = attribute.Key("browser")
	// metricKeyCode is a Turnstile error's six-digit code, or "other".
	metricKeyCode = attribute.Key("code")
	// metricKeyPath is the path a Turnstile check ran on, a fixed word.
	metricKeyPath = attribute.Key("path")
	// metricKeyTo is the state the authorization model changed to.
	metricKeyTo = attribute.Key("to")
)

// MetricsPath is where a plugin sends its metrics to the core, on the plugin
// port, as OTLP/HTTP with a protobuf body.
const MetricsPath = "/otlp/v1/metrics"

const (
	// serviceNameKey is the one resource attribute a metric keeps.
	serviceNameKey = "service.name"
	// maxLabelLen is the longest string value a data point's attribute keeps:
	// the words above are far shorter, so a longer one is no vocabulary.
	maxLabelLen = 64
	// forwardTimeout bounds the forwarding of one plugin's metrics.
	forwardTimeout = 10 * time.Second
	// metricsContentType is the media type of an OTLP/HTTP protobuf body.
	metricsContentType = "application/x-protobuf"
	// maxForwardResponse caps how much of the collector's answer is read.
	maxForwardResponse = 1 << 16
)

// exportableMetric reports whether the instrument called name may leave the
// process.
func exportableMetric(name string) bool {
	switch name {
	case MetricUIClientEvents, MetricUITurnstileTime, MetricMCPUnauthenticated, MetricAuthzChanges, MetricAuthzAvailable:
		return true
	}
	return false
}

// exportableMetricKey reports whether a data point's attribute may leave the
// process.
func exportableMetricKey(k attribute.Key) bool {
	switch k {
	case metricKeyKind, metricKeyBrowser, metricKeyCode, metricKeyPath, metricKeyTo:
		return true
	}
	return false
}

// RedactMetrics returns exp behind the privacy filter every exported metric
// passes (D-172, GR-12): only the metrics named above are kept, each data
// point with only the attributes above, the scope with its name alone, the
// resource with its service name alone, and no exemplar, which may hold a
// trace and attributes of its own. Everything else is dropped before export.
func RedactMetrics(exp sdkmetric.Exporter) RedactingMetricExporter {
	return RedactingMetricExporter{Exporter: exp}
}

// RedactingMetricExporter is the exporter RedactMetrics returns: the next
// exporter, with Export in front of it.
type RedactingMetricExporter struct {
	sdkmetric.Exporter
}

// Export hands the next exporter the redacted metrics, and nothing when none
// is left, so a process with no metric to send makes no call.
func (r RedactingMetricExporter) Export(ctx context.Context, rm *metricdata.ResourceMetrics) error {
	out := redactResourceMetrics(rm)
	if len(out.ScopeMetrics) == 0 {
		return nil
	}
	if err := r.Exporter.Export(ctx, out); err != nil {
		return fmt.Errorf("export metrics: %w", err)
	}
	return nil
}

// redactResourceMetrics returns the part of rm that may leave the process.
func redactResourceMetrics(rm *metricdata.ResourceMetrics) *metricdata.ResourceMetrics {
	out := &metricdata.ResourceMetrics{Resource: serviceResource(rm)}
	for _, sm := range rm.ScopeMetrics {
		kept := metricdata.ScopeMetrics{Scope: scopeNamed(sm.Scope.Name)}
		for i := range sm.Metrics {
			m := &sm.Metrics[i]
			next := metricdata.Metrics{Name: m.Name, Description: m.Description, Unit: m.Unit}
			if exportableMetric(m.Name) && redactInto(&next, m.Data) {
				kept.Metrics = append(kept.Metrics, next)
			}
		}
		if len(kept.Metrics) > 0 {
			out.ScopeMetrics = append(out.ScopeMetrics, kept)
		}
	}
	return out
}

// redactInto sets dst's data to data with only exportable attributes and no
// exemplars, and says whether it did: a kind of data it doesn't know is
// dropped.
func redactInto(dst *metricdata.Metrics, data metricdata.Aggregation) bool {
	switch d := data.(type) {
	case metricdata.Sum[int64]:
		d.DataPoints = redactPoints(d.DataPoints)
		dst.Data = d
	case metricdata.Sum[float64]:
		d.DataPoints = redactPoints(d.DataPoints)
		dst.Data = d
	case metricdata.Gauge[int64]:
		d.DataPoints = redactPoints(d.DataPoints)
		dst.Data = d
	case metricdata.Gauge[float64]:
		d.DataPoints = redactPoints(d.DataPoints)
		dst.Data = d
	case metricdata.Histogram[int64]:
		d.DataPoints = redactHistogramPoints(d.DataPoints)
		dst.Data = d
	case metricdata.Histogram[float64]:
		d.DataPoints = redactHistogramPoints(d.DataPoints)
		dst.Data = d
	default:
		return false
	}
	return true
}

// redactPoints returns pts with only exportable attributes and no exemplars.
func redactPoints[N int64 | float64](pts []metricdata.DataPoint[N]) []metricdata.DataPoint[N] {
	out := make([]metricdata.DataPoint[N], len(pts))
	for i := range pts {
		out[i] = pts[i]
		out[i].Attributes = keepLabels(pts[i].Attributes)
		out[i].Exemplars = nil
	}
	return out
}

// redactHistogramPoints is redactPoints for a histogram's.
func redactHistogramPoints[N int64 | float64](pts []metricdata.HistogramDataPoint[N]) []metricdata.HistogramDataPoint[N] {
	out := make([]metricdata.HistogramDataPoint[N], len(pts))
	for i := range pts {
		out[i] = pts[i]
		out[i].Attributes = keepLabels(pts[i].Attributes)
		out[i].Exemplars = nil
	}
	return out
}

// keepLabels returns the exportable attributes of set whose string values are
// short.
func keepLabels(set attribute.Set) attribute.Set {
	kept := make([]attribute.KeyValue, 0, set.Len())
	for iter := set.Iter(); iter.Next(); {
		kv := iter.Attribute()
		if exportableMetricKey(kv.Key) && (kv.Value.Type() != attribute.STRING || len(kv.Value.AsString()) <= maxLabelLen) {
			kept = append(kept, kv)
		}
	}
	return attribute.NewSet(kept...)
}

// FilterOTLPMetrics applies the same filter to the metrics a plugin sent the
// core in req, which it changes in place: only the metrics and attributes
// RedactMetrics keeps stay, and the resource becomes the service name alone,
// service, which the core takes from the plugin's verified token and never
// from the plugin's word. A metric of a kind it doesn't know, a summary, is
// dropped. It returns how many metrics are left.
func FilterOTLPMetrics(req *collectormetrics.ExportMetricsServiceRequest, service string) int {
	var total int
	resources := req.GetResourceMetrics()[:0]
	for _, rm := range req.GetResourceMetrics() {
		rm.Resource = &resourcepb.Resource{Attributes: []*commonpb.KeyValue{{
			Key: serviceNameKey, Value: &commonpb.AnyValue{Value: &commonpb.AnyValue_StringValue{StringValue: service}},
		}}}
		rm.SchemaUrl = ""
		n := filterScopes(rm)
		total += n
		if n > 0 {
			resources = append(resources, rm)
		}
	}
	req.ResourceMetrics = resources
	return total
}

// filterScopes filters the scopes of rm in place, keeps those with a metric
// left, and returns how many metrics are left in all.
func filterScopes(rm *metricspb.ResourceMetrics) int {
	var total int
	scopes := rm.GetScopeMetrics()[:0]
	for _, sm := range rm.GetScopeMetrics() {
		sm.Scope = &commonpb.InstrumentationScope{Name: truncated(sm.GetScope().GetName())}
		sm.SchemaUrl = ""
		metrics := sm.GetMetrics()[:0]
		for _, m := range sm.GetMetrics() {
			if exportableMetric(m.GetName()) && filterMetric(m) {
				metrics = append(metrics, m)
			}
		}
		sm.Metrics = metrics
		total += len(metrics)
		if len(metrics) > 0 {
			scopes = append(scopes, sm)
		}
	}
	rm.ScopeMetrics = scopes
	return total
}

// filterMetric strips m's data points of attributes that may not leave and of
// exemplars, and says whether m is of a kind that is kept.
func filterMetric(m *metricspb.Metric) bool {
	m.Metadata = nil
	m.Description = truncated(m.GetDescription())
	switch d := m.GetData().(type) {
	case *metricspb.Metric_Gauge:
		for _, p := range d.Gauge.GetDataPoints() {
			p.Attributes, p.Exemplars = keepAttributes(p.GetAttributes()), nil
		}
	case *metricspb.Metric_Sum:
		for _, p := range d.Sum.GetDataPoints() {
			p.Attributes, p.Exemplars = keepAttributes(p.GetAttributes()), nil
		}
	case *metricspb.Metric_Histogram:
		for _, p := range d.Histogram.GetDataPoints() {
			p.Attributes, p.Exemplars = keepAttributes(p.GetAttributes()), nil
		}
	case *metricspb.Metric_ExponentialHistogram:
		for _, p := range d.ExponentialHistogram.GetDataPoints() {
			p.Attributes, p.Exemplars = keepAttributes(p.GetAttributes()), nil
		}
	default:
		return false
	}
	return true
}

// keepAttributes returns the attributes of a received data point that may
// leave: an exportable key with a short string, a number or a boolean.
func keepAttributes(attrs []*commonpb.KeyValue) []*commonpb.KeyValue {
	out := make([]*commonpb.KeyValue, 0, len(attrs))
	for _, kv := range attrs {
		if !exportableMetricKey(attribute.Key(kv.GetKey())) {
			continue
		}
		switch v := kv.GetValue().GetValue().(type) {
		case *commonpb.AnyValue_StringValue:
			if len(v.StringValue) <= maxLabelLen {
				out = append(out, kv)
			}
		case *commonpb.AnyValue_IntValue, *commonpb.AnyValue_BoolValue, *commonpb.AnyValue_DoubleValue:
			out = append(out, kv)
		}
	}
	return out
}

// truncated returns s cut to the longest label.
func truncated(s string) string {
	if len(s) > maxLabelLen {
		return s[:maxLabelLen]
	}
	return s
}

// MetricsForwarder sends the metrics the core received from its plugins on to
// the core's own OTLP endpoint, after FilterOTLPMetrics, so the plugins never
// reach a collector themselves (D-172). A nil forwarder drops them: metrics
// are off.
type MetricsForwarder struct {
	url     string
	headers map[string]string
	client  *http.Client
}

// ErrForward reports a collector that didn't take forwarded metrics. It says
// nothing of the endpoint, which may hold a credential.
var ErrForward = errors.New("the collector did not take the metrics")

// MetricsForwarderFromEnv returns the forwarder of the OpenTelemetry
// variables getenv holds, such as os.Getenv, or nil when metrics are off.
func MetricsForwarderFromEnv(getenv func(string) string) (*MetricsForwarder, error) {
	cfg, err := ParseConfig(getenv, "")
	if err != nil {
		return nil, fmt.Errorf("metrics configuration: %w", err)
	}
	return NewMetricsForwarder(&cfg), nil
}

// NewMetricsForwarder returns the forwarder of cfg's metrics endpoint, or nil
// when cfg exports no metrics. It sends straight to the endpoint with cfg's
// headers, never through a proxy.
func NewMetricsForwarder(cfg *Config) *MetricsForwarder {
	if !cfg.MetricsEnabled() {
		return nil
	}
	transport, ok := http.DefaultTransport.(*http.Transport)
	if !ok {
		transport = &http.Transport{}
	}
	transport = transport.Clone()
	transport.Proxy = nil
	return &MetricsForwarder{url: cfg.MetricsEndpoint, headers: cfg.headers, client: &http.Client{Transport: transport, Timeout: forwardTimeout}}
}

// Forward sends req, already filtered, to the collector. On a nil forwarder
// it does nothing.
func (f *MetricsForwarder) Forward(ctx context.Context, req *collectormetrics.ExportMetricsServiceRequest) error {
	if f == nil {
		return nil
	}
	body, err := proto.Marshal(req)
	if err != nil {
		return fmt.Errorf("marshal metrics: %w", err)
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, f.url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("%w: building the request", ErrForward)
	}
	httpReq.Header.Set("Content-Type", metricsContentType)
	for k, v := range f.headers {
		httpReq.Header.Set(k, v)
	}
	resp, err := f.client.Do(httpReq)
	if err != nil {
		return fmt.Errorf("%w: no answer", ErrForward)
	}
	_, drainErr := io.Copy(io.Discard, io.LimitReader(resp.Body, maxForwardResponse))
	closeErr := resp.Body.Close()
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("%w: status %d", ErrForward, resp.StatusCode)
	}
	if err := errors.Join(drainErr, closeErr); err != nil {
		return fmt.Errorf("%w: answer not read", ErrForward)
	}
	return nil
}

// newMeterProvider returns the meter provider that sends metrics through exp,
// redacted, named service in its resource.
func newMeterProvider(exp sdkmetric.Exporter, service string) *sdkmetric.MeterProvider {
	return sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(sdkmetric.NewPeriodicReader(RedactMetrics(exp))),
		sdkmetric.WithResource(serviceOnly(service)),
	)
}

// newMetricsExporter returns the OTLP/HTTP metrics exporter for cfg's metrics
// endpoint: straight to it, never through a proxy.
func newMetricsExporter(ctx context.Context, cfg *Config) (*otlpmetrichttp.Exporter, error) {
	exp, err := otlpmetrichttp.New(ctx, otlpmetrichttp.WithEndpointURL(cfg.MetricsEndpoint),
		otlpmetrichttp.WithHeaders(cfg.headers), otlpmetrichttp.WithProxy(http.ProxyURL(nil)))
	if err != nil {
		return nil, fmt.Errorf("OTLP HTTP metrics exporter: %w", err)
	}
	return exp, nil
}

// serviceOnly is the resource of a process that exports metrics: its service
// name and nothing else.
func serviceOnly(name string) *resource.Resource {
	return resource.NewSchemaless(attribute.String(serviceNameKey, name))
}

// serviceResource is the resource of rm cut to its service name.
func serviceResource(rm *metricdata.ResourceMetrics) *resource.Resource {
	if rm.Resource != nil {
		if v, ok := rm.Resource.Set().Value(serviceNameKey); ok {
			return serviceOnly(v.AsString())
		}
	}
	return resource.Empty()
}

// scopeNamed is an instrumentation scope with a name and nothing else.
func scopeNamed(name string) instrumentation.Scope {
	return instrumentation.Scope{Name: truncated(name)}
}
