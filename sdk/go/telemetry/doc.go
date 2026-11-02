// Package telemetry is OpenFief's distributed tracing with OpenTelemetry,
// shared by the core, the operator and plugins built on this SDK.
//
// Export is off by default: until an OTLP endpoint is configured, Start
// installs a no-op tracer provider, which records nothing and opens no
// connection. ParseConfig reads the standard OpenTelemetry variables,
// OTEL_EXPORTER_OTLP_ENDPOINT, OTEL_EXPORTER_OTLP_PROTOCOL (grpc or
// http/protobuf), OTEL_EXPORTER_OTLP_HEADERS, OTEL_TRACES_SAMPLER and its
// _ARG, OTEL_SERVICE_NAME and OTEL_RESOURCE_ATTRIBUTES, so any collector
// works. Shutdown sends the spans still buffered.
//
// Spans never carry money amounts, quantities, account numbers, ISINs,
// tokens, session IDs or SQL arguments (GR-12): every exported span passes
// Redact, which keeps only an allow-list of attributes and drops error
// messages. ServerHandler, PublicHandler and Transport trace HTTP and
// Connect calls between OpenFief's processes and propagate W3C trace
// context; nothing propagates it to the internet. Run runs a domain
// operation in a named span.
package telemetry
