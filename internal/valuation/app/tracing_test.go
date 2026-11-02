package app_test

import (
	"errors"
	"slices"
	"testing"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

// TestRevalueRunsInASpan checks that a revaluation runs in a span
// valuation.revalue with no attributes, so it names no account, date or
// value, and that a failed one marks its span as an error without the
// error's message. It is the only test in this package that sets
// OpenTelemetry's tracer provider; its spans are told apart from other
// tests' by their parent.
func TestRevalueRunsInASpan(t *testing.T) {
	t.Parallel()
	rec := tracetest.NewSpanRecorder()
	otel.SetTracerProvider(sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(rec)))
	ctx, root := otel.Tracer("test").Start(t.Context(), "root")
	s := marketStore(t, []string{"2026-03-02"}, []string{"2026-03-02"})
	if _, err := revaluer(t, etfLedger(t), s, "2026-03-06").Revalue(ctx, acc1, date(t, "2026-03-02")); err != nil {
		t.Fatalf("Revalue: %v", err)
	}
	broken := etfLedger(t)
	broken.err = errStoreDown
	if _, err := revaluer(t, broken, newFakeStore(), "2026-03-02").Revalue(ctx, acc1, date(t, "2026-03-02")); !errors.Is(err, errStoreDown) {
		t.Fatalf("Revalue over a failing log = %v, want %v", err, errStoreDown)
	}
	root.End()
	var got []codes.Code
	for _, sp := range rec.Ended() {
		if sp.Parent().SpanID() != root.SpanContext().SpanID() {
			continue
		}
		if sp.Name() != "valuation.revalue" || len(sp.Attributes()) != 0 || sp.Status().Description != "" {
			t.Errorf("span %s with %v and %+v, want valuation.revalue with nothing on it", sp.Name(), sp.Attributes(), sp.Status())
		}
		got = append(got, sp.Status().Code)
	}
	if want := []codes.Code{codes.Unset, codes.Error}; !slices.Equal(want, got) {
		t.Errorf("revaluation spans' status = %v, want %v", got, want)
	}
}
