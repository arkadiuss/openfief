package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"maps"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	v1 "github.com/authzed/authzed-go/proto/authzed/api/v1"
	"go.opentelemetry.io/otel/attribute"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	"google.golang.org/grpc"

	"github.com/arkadiuss/openfief/internal/identity"
	identityapp "github.com/arkadiuss/openfief/internal/identity/app"
	"github.com/arkadiuss/openfief/internal/identity/spicedb"
	"github.com/arkadiuss/openfief/internal/kernel"
)

// recorded is a slog.Handler that keeps each record's message and signals on
// updates, so a test waits on a log line instead of sleeping.
type recorded struct {
	mu      sync.Mutex
	lines   []string
	updates chan struct{}
}

func newRecorded() *recorded { return &recorded{updates: make(chan struct{}, 16)} }

func (r *recorded) Enabled(context.Context, slog.Level) bool { return true }
func (r *recorded) WithAttrs([]slog.Attr) slog.Handler       { return r }
func (r *recorded) WithGroup(string) slog.Handler            { return r }

func (r *recorded) Handle(_ context.Context, rec slog.Record) error { //nolint:gocritic // slog.Handler's signature
	r.mu.Lock()
	r.lines = append(r.lines, rec.Level.String()+" "+rec.Message)
	r.mu.Unlock()
	select {
	case r.updates <- struct{}{}:
	default:
	}
	return nil
}

func (r *recorded) all() []string {
	r.mu.Lock()
	defer r.mu.Unlock()
	return append([]string(nil), r.lines...)
}

// scriptedAuthorizer answers each check with err, which a test sets.
type scriptedAuthorizer struct {
	mu  sync.Mutex
	err error
}

func (s *scriptedAuthorizer) MayAct(context.Context, identity.Caller, identity.Action, string) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.err == nil, s.err
}

func (s *scriptedAuthorizer) set(err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.err = err
}

// listingAuthorizer is scriptedAuthorizer that also lists accounts.
type listingAuthorizer struct{ scriptedAuthorizer }

func (s *listingAuthorizer) ActableAccounts(context.Context, identity.Caller, identity.Action) (map[string]struct{}, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return nil, s.err
}

var errRefusal = errors.New("not an outage: a refusal")

var errConnectionRefused = errors.New("connection refused")

var errDown = fmt.Errorf("%w: SpiceDB does not answer", kernel.ErrAuthzUnavailable)

// testMonitor returns a monitor over a recorded log and a meter of an
// in-memory reader, with no probe.
func testMonitor(t *testing.T, ping func(context.Context) error, every time.Duration) (*authzMonitor, *recorded, *sdkmetric.ManualReader) {
	t.Helper()
	reader := sdkmetric.NewManualReader()
	meter := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader)).Meter("test")
	log := newRecorded()
	m, err := newAuthzMonitor(t.Context(), slog.New(log), meter, ping, every)
	if err != nil {
		t.Fatalf("newAuthzMonitor: %v", err)
	}
	return m, log, reader
}

// TestAuthzMonitorLogsEachChangeOnce checks that the transition from
// available to unavailable and back is logged once each, however many checks
// fail or succeed in between, that only an error that wraps
// kernel.ErrAuthzUnavailable is an outage (a refusal's error is not), and
// that the first answer logs nothing.
func TestAuthzMonitorLogsEachChangeOnce(t *testing.T) {
	t.Parallel()
	m, log, _ := testMonitor(t, nil, time.Hour)
	inner := &scriptedAuthorizer{}
	az := m.watch(authorization{authorizer: inner}).authorizer
	owner := identity.Caller{Holder: identity.HolderOwner, ID: "s"}
	check := func(want error) {
		t.Helper()
		if _, err := az.MayAct(t.Context(), owner, identity.ActionView, "bonds"); !errors.Is(err, want) {
			t.Fatalf("MayAct = %v, want %v", err, want)
		}
	}
	check(nil)
	check(nil)
	inner.set(errRefusal)
	check(errRefusal)
	if got := log.all(); len(got) != 0 {
		t.Fatalf("logged %v while the model answered, want nothing", got)
	}
	inner.set(errDown)
	for range 3 {
		check(errDown)
	}
	inner.set(nil)
	check(nil)
	check(nil)
	inner.set(errDown)
	check(errDown)
	want := []string{
		"WARN authorization is unavailable: SpiceDB does not answer; every check is refused and the requests that need one get 503 until it does",
		"INFO authorization is available again: SpiceDB answers",
		"WARN authorization is unavailable: SpiceDB does not answer; every check is refused and the requests that need one get 503 until it does",
	}
	if got := log.all(); strings.Join(got, "\n") != strings.Join(want, "\n") {
		t.Errorf("log = %q, want %q", got, want)
	}
}

// TestAuthzMonitorProbeSeesTheWayBack checks that after an outage the monitor
// asks SpiceDB until it answers and then logs that it is back, with no request
// to see it, and stops asking.
func TestAuthzMonitorProbeSeesTheWayBack(t *testing.T) {
	t.Parallel()
	var asked atomic.Int32
	answers := make(chan error)
	ping := func(context.Context) error {
		asked.Add(1)
		return <-answers
	}
	m, log, _ := testMonitor(t, ping, time.Millisecond)
	m.observe(errDown)
	answers <- errDown // still down: the probe asks again
	answers <- nil
	for {
		if lines := log.all(); len(lines) == 2 {
			if lines[1] != "INFO authorization is available again: SpiceDB answers" {
				t.Fatalf("log = %q", lines)
			}
			break
		}
		<-log.updates
	}
	if n := asked.Load(); n != 2 {
		t.Errorf("SpiceDB was asked %d times, want 2", n)
	}
	// The probe has ended with the outage: another outage starts a new one.
	m.observe(errDown)
	answers <- nil
	for len(log.all()) < 4 {
		<-log.updates
	}
}

// metrics is what the reader holds of the monitor's two instruments.
type metrics struct {
	// gauge holds the gauge's values, none while it is absent.
	gauge []int64
	// changes is the counter by the state it went to.
	changes map[string]int64
}

// collect reads the monitor's instruments from reader.
func collect(t *testing.T, reader *sdkmetric.ManualReader) metrics {
	t.Helper()
	var rm metricdata.ResourceMetrics
	if err := reader.Collect(t.Context(), &rm); err != nil {
		t.Fatalf("Collect: %v", err)
	}
	got := metrics{changes: map[string]int64{}}
	for _, sm := range rm.ScopeMetrics {
		for _, mt := range sm.Metrics {
			switch d := mt.Data.(type) {
			case metricdata.Gauge[int64]:
				for _, p := range d.DataPoints {
					got.gauge = append(got.gauge, p.Value)
				}
			case metricdata.Sum[int64]:
				for _, p := range d.DataPoints {
					v, _ := p.Attributes.Value(attribute.Key("to"))
					got.changes[v.AsString()] = p.Value
				}
			}
		}
	}
	return got
}

// TestAuthzMonitorMetric checks the metric: the gauge says 1 while the model
// answers and 0 while it can't, and is absent before the first check, and the
// counter counts each change by the state it went to.
func TestAuthzMonitorMetric(t *testing.T) {
	t.Parallel()
	m, _, reader := testMonitor(t, nil, time.Hour)
	if got := collect(t, reader); len(got.gauge) != 0 {
		t.Errorf("gauge before any check = %v, want absent", got.gauge)
	}
	m.observe(nil)
	if got := collect(t, reader); !slices.Equal(got.gauge, []int64{1}) {
		t.Errorf("gauge while the model answers = %v, want 1", got.gauge)
	}
	m.observe(errDown)
	m.observe(errDown)
	got := collect(t, reader)
	if want := (map[string]int64{"unavailable": 1}); !slices.Equal(got.gauge, []int64{0}) || !maps.Equal(got.changes, want) {
		t.Errorf("gauge, changes during the outage = %v, %v, want 0 and %v", got.gauge, got.changes, want)
	}
	m.observe(nil)
	got = collect(t, reader)
	if want := (map[string]int64{"unavailable": 1, "available": 1}); !slices.Equal(got.gauge, []int64{1}) || !maps.Equal(got.changes, want) {
		t.Errorf("gauge, changes after it came back = %v, %v, want 1 and %v", got.gauge, got.changes, want)
	}
}

// TestAuthzMonitorKeepsTheListing checks that the wrapper lists accounts
// exactly when the authorizer it wraps does, so the views keep their one-call
// list, and that a failing list is an outage too.
func TestAuthzMonitorKeepsTheListing(t *testing.T) {
	t.Parallel()
	m, log, _ := testMonitor(t, nil, time.Hour)
	if _, ok := m.watch(authorization{authorizer: &scriptedAuthorizer{}}).authorizer.(identityapp.AccountLister); ok {
		t.Error("the wrapper of an authorizer that doesn't list does")
	}
	inner := &listingAuthorizer{}
	inner.set(errDown)
	lister, ok := m.watch(authorization{authorizer: inner}).authorizer.(identityapp.AccountLister)
	if !ok {
		t.Fatal("the wrapper of an authorizer that lists doesn't")
	}
	if _, err := lister.ActableAccounts(t.Context(), identity.Caller{}, identity.ActionView); !errors.Is(err, kernel.ErrAuthzUnavailable) {
		t.Errorf("ActableAccounts = %v, want the outage passed on unchanged", err)
	}
	if got := log.all(); len(got) != 1 || !strings.HasPrefix(got[0], "WARN authorization is unavailable") {
		t.Errorf("log = %q, want the outage once", got)
	}
}

// failingPerms is a SpiceDB that fails every read while down is set, and
// otherwise answers as writeLog does.
type failingPerms struct {
	writeLog

	down atomic.Bool
}

func (f *failingPerms) ReadRelationships(ctx context.Context, in *v1.ReadRelationshipsRequest, opts ...grpc.CallOption) (grpc.ServerStreamingClient[v1.ReadRelationshipsResponse], error) {
	if f.down.Load() {
		return nil, errConnectionRefused
	}
	return f.writeLog.ReadRelationships(ctx, in, opts...)
}

// readyServer returns a server wired with az, over a database that answers.
func readyServer(az authorization) *server {
	s := &server{dbPing: func(context.Context) error { return nil }}
	api := http.NotFoundHandler()
	s.browser.Store(&api)
	s.ready(&core{access: access{authorization: az}})
	return s
}

// answer is what an endpoint said: its status and its text.
type answer struct {
	code int
	body string
}

func readyz(t *testing.T, s *server) answer {
	t.Helper()
	rec := httptest.NewRecorder()
	s.browserHandler().ServeHTTP(rec, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/readyz", http.NoBody))
	return answer{rec.Code, strings.TrimSpace(rec.Body.String())}
}

// TestReadyOnlyAfterSpiceDBAnsweredWithItsSchema checks, with the key in the
// environment, that /readyz is 503 until the start's reconcile has succeeded,
// which is SpiceDB answering and holding the schema; a reconcile that fails
// leaves the core not ready; and then it is ready.
func TestReadyOnlyAfterSpiceDBAnsweredWithItsSchema(t *testing.T) {
	t.Parallel()
	perms := &failingPerms{}
	perms.down.Store(true)
	az := authorization{authorizer: &scriptedAuthorizer{}, syncer: spicedb.NewSyncer(perms), started: new(atomic.Bool)}
	s := readyServer(az)
	load := func(context.Context) (spicedb.Desired, error) { return spicedb.Desired{}, nil }
	const waiting = "waiting for SpiceDB to answer and take its schema"
	if got := readyz(t, s); got != (answer{http.StatusServiceUnavailable, waiting}) {
		t.Fatalf("GET /readyz before the reconcile = %v, want 503 %q", got, waiting)
	}
	reconcileAccess(t.Context(), az.syncer, load, slog.New(slog.DiscardHandler), az.markStarted)
	if got := readyz(t, s); got != (answer{http.StatusServiceUnavailable, waiting}) {
		t.Fatalf("GET /readyz after a reconcile SpiceDB refused = %v, want 503 %q", got, waiting)
	}
	perms.down.Store(false)
	reconcileAccess(t.Context(), az.syncer, load, slog.New(slog.DiscardHandler), az.markStarted)
	if got := readyz(t, s); got != (answer{http.StatusOK, "ok"}) {
		t.Errorf("GET /readyz after the reconcile = %v, want 200 ok", got)
	}
}

// TestReadyStaysReadyDuringAnOutage checks that once the core is ready a
// SpiceDB outage does not flip /readyz: checks fail closed instead (503 from
// the APIs, tested with the views), and /readyz stays 200 through the outage
// and after it, so the Service keeps the core and its pods are not all taken
// out at once by one SpiceDB restart.
func TestReadyStaysReadyDuringAnOutage(t *testing.T) {
	t.Parallel()
	m, _, _ := testMonitor(t, nil, time.Hour)
	inner := &scriptedAuthorizer{}
	az := m.watch(authorization{authorizer: inner, syncer: spicedb.NewSyncer(&writeLog{}), started: new(atomic.Bool)})
	az.markStarted()
	s := readyServer(az)
	if got := readyz(t, s); got.code != http.StatusOK {
		t.Fatalf("GET /readyz after the start = %v, want 200", got)
	}
	inner.set(errDown)
	if _, err := az.authorizer.MayAct(t.Context(), identity.Caller{Holder: identity.HolderOwner, ID: "s"}, identity.ActionView, "bonds"); !errors.Is(err, kernel.ErrAuthzUnavailable) {
		t.Fatalf("MayAct during the outage = %v, want unavailable", err)
	}
	if m.state.Load() != authzUnavailable {
		t.Fatal("the monitor didn't see the outage")
	}
	if got := readyz(t, s); got != (answer{http.StatusOK, "ok"}) {
		t.Errorf("GET /readyz during the outage = %v, want 200 ok", got)
	}
	inner.set(nil)
	if _, err := az.authorizer.MayAct(t.Context(), identity.Caller{Holder: identity.HolderOwner, ID: "s"}, identity.ActionView, "bonds"); err != nil {
		t.Fatalf("MayAct after it came back = %v", err)
	}
	if got := readyz(t, s); got.code != http.StatusOK {
		t.Errorf("GET /readyz after the outage = %v, want 200", got)
	}
}
