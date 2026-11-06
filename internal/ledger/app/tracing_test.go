package app_test

import (
	"errors"
	"slices"
	"testing"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"

	"github.com/arkadiuss/openfief/internal/ledger"
	"github.com/arkadiuss/openfief/internal/ledger/app"
	"github.com/arkadiuss/openfief/sdk/go/telemetry"
)

// TestImportsRunInSpans checks that an import, a seed and an addition of
// opening lots each run in a span ledger.import naming which it is and
// nothing else, and that a refused import marks its span as an error
// without the error's message. It is the only test in this package that
// sets OpenTelemetry's tracer provider; its spans are told apart from
// other tests' by their parent.
func TestImportsRunInSpans(t *testing.T) {
	t.Parallel()
	rec := tracetest.NewSpanRecorder()
	otel.SetTracerProvider(sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(rec)))
	ctx, root := otel.Tracer("test").Start(t.Context(), "root")
	f := newImportFixture(t)
	l, _ := seeded(t)
	if _, err := l.Import(ctx, []ledger.Transaction{f.deposit}, nil); err != nil {
		t.Fatalf("Import: %v", err)
	}
	if _, err := l.Import(ctx, []ledger.Transaction{f.elsewhere}, nil); !errors.Is(err, app.ErrUnknownAccount) {
		t.Fatalf("Import into an unknown account = %v, want %v", err, app.ErrUnknownAccount)
	}
	_, seedErr := l.Seed(ctx, nil, nil)
	_, lotsErr := l.AddOpeningLots(ctx, nil, nil)
	root.End()
	var got []string
	for _, s := range rec.Ended() {
		if s.Parent().SpanID() != root.SpanContext().SpanID() {
			continue
		}
		if s.Name() != "ledger.import" || len(s.Attributes()) != 1 || s.Attributes()[0].Key != telemetry.KeyImport || s.Status().Description != "" {
			t.Errorf("span %s with %v and %+v, want ledger.import with the mode alone", s.Name(), s.Attributes(), s.Status())
			continue
		}
		got = append(got, s.Attributes()[0].Value.AsString()+":"+statusOf(s.Status().Code))
	}
	want := []string{"import:ok", "import:error", "seed:" + outcome(seedErr), "opening_lots:" + outcome(lotsErr)}
	if !slices.Equal(want, got) {
		t.Errorf("spans = %v, want %v", got, want)
	}
}

// statusOf names a span's status code.
func statusOf(c codes.Code) string {
	if c == codes.Error {
		return "error"
	}
	return "ok"
}

// outcome names how a call ended, as statusOf names its span's.
func outcome(err error) string {
	if err != nil {
		return "error"
	}
	return "ok"
}
