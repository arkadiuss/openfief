package operator

import (
	"context"
	"errors"
	"testing"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.opentelemetry.io/otel/trace"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/arkadiuss/openfief/sdk/go/telemetry"
)

// errReconcile is a reconcile's failure.
var errReconcile = errors.New("synthetic reconcile failure")

// TestTracedReconciles checks that traced runs each reconcile in a span
// operator.reconcile naming the controller alone, passes the reconcile its
// span and its result back, marks a failure as an error without its message
// and wraps the failure with the controller's name. It is the only test in
// this package that sets OpenTelemetry's tracer provider; its spans are
// told apart from the reconciles of other tests by their parent.
func TestTracedReconciles(t *testing.T) {
	t.Parallel()
	rec := tracetest.NewSpanRecorder()
	otel.SetTracerProvider(sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(rec)))
	ctx, root := otel.Tracer("test").Start(t.Context(), "root")
	defer root.End()
	var inSpan bool
	r := traced("plugin", reconcile.Func(func(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
		inSpan = trace.SpanFromContext(ctx).SpanContext().IsValid()
		if req.Name == "broken" {
			return reconcile.Result{}, errReconcile
		}
		return reconcile.Result{RequeueAfter: time.Minute}, nil
	}))
	ok := reconcile.Request{}
	ok.Name = "nbp-fx"
	if res, err := r.Reconcile(ctx, ok); err != nil || res.RequeueAfter != time.Minute || !inSpan {
		t.Errorf("Reconcile = %+v, %v, in a span %v; want the reconcile's result in a span", res, err, inSpan)
	}
	broken := reconcile.Request{}
	broken.Name = "broken"
	_, err := r.Reconcile(ctx, broken)
	if !errors.Is(err, errReconcile) || err.Error() != "plugin controller: "+errReconcile.Error() {
		t.Errorf("failed Reconcile = %v, want %v wrapped with the controller", err, errReconcile)
	}
	got := reconcileStatuses(t, rec, root.SpanContext().SpanID())
	if len(got) != 2 || got[0] != codes.Unset || got[1] != codes.Error {
		t.Errorf("reconcile spans' status = %v, want unset then error", got)
	}
}

// reconcileStatuses checks the spans rec ended under parent and returns
// their status codes in order.
func reconcileStatuses(t *testing.T, rec *tracetest.SpanRecorder, parent trace.SpanID) []codes.Code {
	t.Helper()
	var got []codes.Code
	for _, s := range rec.Ended() {
		if s.Parent().SpanID() != parent {
			continue // another test's reconcile
		}
		attrs := s.Attributes()
		if s.Name() != "operator.reconcile" || len(attrs) != 1 || attrs[0] != telemetry.KeyController.String("plugin") || s.Status().Description != "" {
			t.Errorf("span %s with %v and %+v, want operator.reconcile naming the controller alone", s.Name(), attrs, s.Status())
		}
		got = append(got, s.Status().Code)
	}
	return got
}
