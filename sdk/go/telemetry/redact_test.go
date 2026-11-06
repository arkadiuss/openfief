package telemetry

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

// containsAny reports whether s contains any of subs.
func containsAny(s string, subs ...string) bool {
	for _, sub := range subs {
		if strings.Contains(s, sub) {
			return true
		}
	}
	return false
}

func TestRedact(t *testing.T) {
	t.Parallel()
	stub := tracetest.SpanStub{
		Name: "openfief.connector.v1.ConnectorService/Sync",
		Attributes: []attribute.KeyValue{
			attribute.String("http.request.method", "POST"),
			attribute.Int("http.response.status_code", 200),
			KeyInstance.String("ibkr-flex"),
			attribute.String("url.full", "https://api.example.org/eod/PL0000000000?api_token=synthetic-token"),
			attribute.String("db.query.text", "INSERT INTO ledger.postings VALUES ($1, $2)"),
			attribute.StringSlice("db.query.parameters", []string{"1234.56", "PLN"}),
			attribute.String("client.address", "192.0.2.10"),
			attribute.String("openfief.amount", "1234.56 PLN"),
		},
		Events: []sdktrace.Event{{Name: "exception", Attributes: []attribute.KeyValue{
			attribute.String("exception.type", "*errors.errorString"),
			attribute.String("exception.message", "account U0000000 holds 12 units"),
		}}},
		Links:  []sdktrace.Link{{Attributes: []attribute.KeyValue{attribute.String("session", "synthetic-session")}}},
		Status: sdktrace.Status{Code: codes.Error, Description: "price 1234.56 for PL0000000000"},
	}
	mem := tracetest.NewInMemoryExporter()
	if err := Redact(mem).ExportSpans(t.Context(), []sdktrace.ReadOnlySpan{stub.Snapshot()}); err != nil {
		t.Fatalf("ExportSpans: %v", err)
	}
	spans := mem.GetSpans()
	if len(spans) != 1 {
		t.Fatalf("exported %d spans, want 1", len(spans))
	}
	checkRedacted(t, &spans[0], stub.Name)
	if err := Redact(mem).Shutdown(t.Context()); err != nil {
		t.Errorf("Shutdown: %v", err)
	}
}

// checkRedacted checks a span Redact exported: named name, with only its
// HTTP and instance attributes, its exception's type, its link without
// attributes and its error code without a description.
func checkRedacted(t *testing.T, got *tracetest.SpanStub, name string) {
	t.Helper()
	wantAttrs := []attribute.KeyValue{
		attribute.String("http.request.method", "POST"),
		attribute.Int("http.response.status_code", 200),
		KeyInstance.String("ibkr-flex"),
	}
	if diff := cmp.Diff(wantAttrs, got.Attributes, cmp.Comparer(func(a, b attribute.KeyValue) bool { return a == b })); diff != "" {
		t.Errorf("attributes (-want +got):\n%s", diff)
	}
	if got.Name != name {
		t.Errorf("name = %q, want %q", got.Name, name)
	}
	if len(got.Events) != 1 || len(got.Events[0].Attributes) != 1 || got.Events[0].Attributes[0].Key != "exception.type" {
		t.Errorf("events = %+v, want the exception's type alone", got.Events)
	}
	if len(got.Links) != 1 || len(got.Links[0].Attributes) != 0 {
		t.Errorf("links = %+v, want the link without attributes", got.Links)
	}
	if got.Status != (sdktrace.Status{Code: codes.Error}) {
		t.Errorf("status = %+v, want the error code without a description", got.Status)
	}
}
