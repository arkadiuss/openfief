package app_test

import (
	"errors"
	"slices"
	"strings"
	"testing"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"

	"github.com/arkadiuss/openfief/internal/ingest"
	"github.com/arkadiuss/openfief/internal/ingest/app"
	"github.com/arkadiuss/openfief/sdk/go/telemetry"
)

// TestConnectorsRunInSpans checks that a sync, a statement upload and a
// bond-lot entry each run in their named span, which names the instance
// and the mode and nothing else, and that a failure marks its span as an
// error without the error's message. It is the only test in this package
// that sets OpenTelemetry's tracer provider; the spans of other tests are
// told apart by the instance.
func TestConnectorsRunInSpans(t *testing.T) {
	t.Parallel()
	rec := tracetest.NewSpanRecorder()
	otel.SetTracerProvider(sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(rec)))
	const instance = "traced-main"
	d := delivery(t)
	d.Prices, d.FXRates, d.IndexReadings, d.IndexLevels = nil, nil, nil, nil
	c := connectors(t, &fakeCaller{delivery: d}, &fakeLedger{}, &fakeStores{})
	if _, err := c.Sync(t.Context(), instance, &app.SyncCall{ID: "b-1", Mode: ingest.ModeTransactions}); err != nil {
		t.Fatalf("Sync: %v", err)
	}
	if _, err := c.ImportFile(t.Context(), instance, "b-2", ingest.ModeOpening, app.NamedFile{Name: "s.csv", Contents: strings.NewReader("x")}); err != nil {
		t.Fatalf("ImportFile: %v", err)
	}
	failing := connectors(t, &fakeCaller{err: errCaller}, &fakeLedger{}, &fakeStores{})
	if _, err := failing.ImportBondLots(t.Context(), instance, "b-3", "acc-1", &ingest.BondLotEntries{}); !errors.Is(err, errCaller) {
		t.Fatalf("ImportBondLots = %v, want %v", err, errCaller)
	}
	checkSpans(t, rec, instance, map[string]spanWant{
		"ingest.sync":             {[]attribute.KeyValue{telemetry.KeyInstance.String(instance), telemetry.KeyMode.String("transactions")}, codes.Unset},
		"ingest.import_file":      {[]attribute.KeyValue{telemetry.KeyInstance.String(instance), telemetry.KeyMode.String("opening")}, codes.Unset},
		"ingest.import_bond_lots": {[]attribute.KeyValue{telemetry.KeyInstance.String(instance)}, codes.Error},
	})
}

// spanWant is what a span must carry: exactly attrs, and code.
type spanWant struct {
	attrs []attribute.KeyValue
	code  codes.Code
}

// checkSpans checks that rec ended each span of want naming instance, with
// what want says and no status description.
func checkSpans(t *testing.T, rec *tracetest.SpanRecorder, instance string, want map[string]spanWant) {
	t.Helper()
	seen := map[string]bool{}
	for _, s := range rec.Ended() {
		w, ok := want[s.Name()]
		if !ok || !slices.Contains(s.Attributes(), telemetry.KeyInstance.String(instance)) {
			continue
		}
		seen[s.Name()] = true
		if !sameAttrs(w.attrs, s.Attributes()) {
			t.Errorf("span %s attributes = %v, want %v", s.Name(), s.Attributes(), w.attrs)
		}
		if s.Status().Code != w.code || s.Status().Description != "" {
			t.Errorf("span %s status = %+v, want code %v and no description", s.Name(), s.Status(), w.code)
		}
	}
	for name := range want {
		if !seen[name] {
			t.Errorf("no span %s", name)
		}
	}
}

// sameAttrs reports whether got holds exactly want, in any order.
func sameAttrs(want, got []attribute.KeyValue) bool {
	if len(want) != len(got) {
		return false
	}
	for _, kv := range want {
		if !slices.Contains(got, kv) {
			return false
		}
	}
	return true
}
