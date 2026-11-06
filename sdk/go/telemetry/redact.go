package telemetry

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel/attribute"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

// OpenFief's own span attributes: the only ones of its own a span may carry
// out of the process. Each holds an internal name, never an amount, a
// quantity, an account number, an ISIN or a credential (GR-12).
const (
	// KeyInstance is a connector instance's name (D-48).
	KeyInstance = attribute.Key("openfief.connector.instance")
	// KeyMode is an import batch's mode: opening or transactions.
	KeyMode = attribute.Key("openfief.ingest.mode")
	// KeyImport is how the ledger imports: import, seed or opening_lots.
	KeyImport = attribute.Key("openfief.ledger.import")
	// KeyController is the operator controller that reconciles.
	KeyController = attribute.Key("openfief.operator.controller")
)

// exportable reports whether an attribute may leave the process: the keys
// that name what ran and how it ended, never what it carried. Everything
// else, URLs, paths, query text and parameters, addresses, user agents,
// headers and error messages included, is dropped before export.
func exportable(k attribute.Key) bool {
	switch k {
	case KeyInstance, KeyMode, KeyImport, KeyController,
		// HTTP and RPC: the method, the route template and the outcome.
		"http.request.method", "http.response.status_code", "http.route",
		"rpc.system", "rpc.service", "rpc.method",
		"network.protocol.name", "network.protocol.version", "error.type",
		// PostgreSQL: the system, the operation (SELECT, INSERT) and the
		// SQLSTATE of a failure; never the statement or its arguments.
		"db.system", "db.system.name", "db.operation.name", "pgx.sql_state",
		// River jobs: their kind, queue, attempt and outcome; never args.
		"kind", "kinds", "queue", "attempt", "priority", "status",
		// An error event's Go type, never its message.
		"exception.type":
		return true
	}
	return false
}

// Redact returns exp behind the privacy filter every exported span passes
// (GR-12): only exportable attributes are kept, on the span, its events and
// its links; a status keeps its code but loses its description, which is
// an error's message. Span names are left as they are: OpenFief names spans
// by their operation, never by what they carry.
func Redact(exp sdktrace.SpanExporter) RedactingExporter {
	return RedactingExporter{next: exp}
}

// RedactingExporter is the exporter Redact returns.
type RedactingExporter struct {
	next sdktrace.SpanExporter
}

// ExportSpans hands the next exporter redacted views of spans.
func (r RedactingExporter) ExportSpans(ctx context.Context, spans []sdktrace.ReadOnlySpan) error {
	out := make([]sdktrace.ReadOnlySpan, len(spans))
	for i, s := range spans {
		out[i] = redactedSpan{ReadOnlySpan: s}
	}
	if err := r.next.ExportSpans(ctx, out); err != nil {
		return fmt.Errorf("export spans: %w", err)
	}
	return nil
}

// Shutdown shuts the next exporter down.
func (r RedactingExporter) Shutdown(ctx context.Context) error {
	if err := r.next.Shutdown(ctx); err != nil {
		return fmt.Errorf("shut down exporter: %w", err)
	}
	return nil
}

// redactedSpan is a span as Redact exports it.
type redactedSpan struct {
	sdktrace.ReadOnlySpan
}

// Attributes returns the span's exportable attributes.
func (s redactedSpan) Attributes() []attribute.KeyValue {
	return keep(s.ReadOnlySpan.Attributes())
}

// Events returns the span's events with their exportable attributes.
func (s redactedSpan) Events() []sdktrace.Event {
	events := s.ReadOnlySpan.Events()
	out := make([]sdktrace.Event, len(events))
	for i := range events {
		out[i] = events[i]
		out[i].Attributes = keep(events[i].Attributes)
	}
	return out
}

// Links returns the span's links with their exportable attributes.
func (s redactedSpan) Links() []sdktrace.Link {
	links := s.ReadOnlySpan.Links()
	out := make([]sdktrace.Link, len(links))
	for i := range links {
		out[i] = links[i]
		out[i].Attributes = keep(links[i].Attributes)
	}
	return out
}

// Status returns the span's status code without its description.
func (s redactedSpan) Status() sdktrace.Status {
	return sdktrace.Status{Code: s.ReadOnlySpan.Status().Code}
}

// keep returns the exportable attributes of attrs.
func keep(attrs []attribute.KeyValue) []attribute.KeyValue {
	out := make([]attribute.KeyValue, 0, len(attrs))
	for _, kv := range attrs {
		if exportable(kv.Key) {
			out = append(out, kv)
		}
	}
	return out
}
