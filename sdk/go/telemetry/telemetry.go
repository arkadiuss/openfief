package telemetry

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"
	"google.golang.org/grpc"
)

// Telemetry is a process's tracing, as Start set it up. Shut it down before
// the process exits, so the spans still buffered are sent.
type Telemetry struct {
	// provider exports spans; nil when export is off.
	provider *sdktrace.TracerProvider
}

// StartFromEnv is Start with the Config ParseConfig reads from the process
// environment.
func StartFromEnv(ctx context.Context, defaultService string) (*Telemetry, error) {
	cfg, err := ParseConfig(os.Getenv, defaultService)
	if err != nil {
		return nil, fmt.Errorf("tracing configuration: %w", err)
	}
	return Start(ctx, &cfg)
}

// Start sets up the process's tracing as cfg says and installs it as
// OpenTelemetry's global tracer provider and propagator, which the
// instrumentation in this package and the libraries OpenFief uses read.
// The propagator is always W3C trace context alone, never baggage, so a
// trace continues across a process that exports nothing. With export off
// the provider is a no-op: no span is recorded and nothing is sent. With
// it on, spans are sampled as cfg says, batched, passed through Redact and
// sent to cfg's endpoint, straight to it and never through a proxy; the
// first export opens the connection.
func Start(ctx context.Context, cfg *Config) (*Telemetry, error) {
	otel.SetTextMapPropagator(propagation.TraceContext{})
	if !cfg.Enabled() {
		otel.SetTracerProvider(noop.NewTracerProvider())
		return &Telemetry{}, nil
	}
	sampler, ok := samplers(cfg.SamplerArg)[cfg.Sampler]
	if !ok {
		return nil, fmt.Errorf("%w: %q", ErrBadSampler, cfg.Sampler)
	}
	exp, err := newExporter(ctx, cfg)
	if err != nil {
		return nil, err
	}
	res, err := resource.New(ctx,
		resource.WithAttributes(attribute.String("service.name", cfg.ServiceName)),
		resource.WithFromEnv(), resource.WithTelemetrySDK())
	if err != nil {
		return nil, errors.Join(fmt.Errorf("tracing resource: %w", err), exp.Shutdown(ctx))
	}
	tp := sdktrace.NewTracerProvider(sdktrace.WithBatcher(Redact(exp)),
		sdktrace.WithSampler(sampler), sdktrace.WithResource(res))
	otel.SetTracerProvider(tp)
	return &Telemetry{provider: tp}, nil
}

// Shutdown sends the spans still buffered and stops exporting, waiting at
// most until ctx is done. With export off it does nothing.
func (t *Telemetry) Shutdown(ctx context.Context) error {
	if t == nil || t.provider == nil {
		return nil
	}
	if err := t.provider.Shutdown(ctx); err != nil {
		return fmt.Errorf("shut down tracing: %w", err)
	}
	return nil
}

// newExporter returns the OTLP exporter cfg names, sending to its endpoint
// with its headers and no proxy: a plugin's HTTPS_PROXY is the egress
// proxy, which never reaches a collector inside the cluster.
func newExporter(ctx context.Context, cfg *Config) (*otlptrace.Exporter, error) {
	switch cfg.Protocol {
	case ProtocolGRPC:
		exp, err := otlptracegrpc.New(ctx, otlptracegrpc.WithEndpointURL(cfg.Endpoint),
			otlptracegrpc.WithHeaders(cfg.headers), otlptracegrpc.WithDialOption(grpc.WithNoProxy()))
		if err != nil {
			return nil, fmt.Errorf("OTLP gRPC exporter: %w", err)
		}
		return exp, nil
	case ProtocolHTTP:
		exp, err := otlptracehttp.New(ctx, otlptracehttp.WithEndpointURL(cfg.Endpoint),
			otlptracehttp.WithHeaders(cfg.headers), otlptracehttp.WithProxy(http.ProxyURL(nil)))
		if err != nil {
			return nil, fmt.Errorf("OTLP HTTP exporter: %w", err)
		}
		return exp, nil
	}
	return nil, ErrBadProtocol
}

// Run runs fn, a domain operation, with a context in a span called name,
// with attrs, from the tracer of the instrumentation scope, such as a
// package's import path. When fn fails, the span is marked as an error
// without the error's message, which may quote data; fn hands its results,
// error included, to its caller itself. With export off the span costs next
// to nothing.
func Run(ctx context.Context, scope, name string, fn func(ctx context.Context) error, attrs ...attribute.KeyValue) {
	ctx, span := otel.Tracer(scope).Start(ctx, name, trace.WithAttributes(attrs...))
	defer span.End()
	if err := fn(ctx); err != nil {
		span.SetStatus(codes.Error, "")
	}
}
