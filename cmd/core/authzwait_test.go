package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/arkadiuss/openfief/internal/identity/spicedb"
)

// testEndpoint is SpiceDB's address in the tests' messages.
const testEndpoint = "openfief-spicedb:50051"

var (
	// errNoDatabase stands for a database the diagnosis can't ask.
	errNoDatabase = errors.New("database unreachable")
	// errScanShape reports a scan into other columns than the diagnosis's two booleans.
	errScanShape = errors.New("the diagnosis scans two booleans")
)

// TestSpiceDBWaitReason checks what a core waiting for SpiceDB says, in its
// readiness and its log (D-164): the role or the database SpiceDB needs
// missing from the openfief-db cluster, by name, with the install step that
// adds them; a key SpiceDB refuses; SpiceDB not answering although its store
// is there; and SpiceDB not answering while the database can't be asked.
func TestSpiceDBWaitReason(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		code     codes.Code
		store    spicedbStore
		storeErr error
		want     string
	}{
		{
			name: "install step 5 skipped: neither role nor database", code: codes.Unavailable,
			want: "the openfief-db cluster has no role spicedb and no database spicedb yet: add the managed role spicedb and the Database spicedb " +
				"of the release's database file to yours and apply it (install step 5); CloudNativePG creates the role once the Secret openfief-db-spicedb exists",
		},
		{
			name: "the role not reconciled yet", code: codes.Unavailable, store: spicedbStore{database: true},
			want: "the openfief-db cluster has no role spicedb yet: add the managed role spicedb and the Database spicedb " +
				"of the release's database file to yours and apply it (install step 5); CloudNativePG creates the role once the Secret openfief-db-spicedb exists",
		},
		{
			name: "the database not reconciled yet", code: codes.DeadlineExceeded, store: spicedbStore{role: true},
			want: "the openfief-db cluster has no database spicedb yet: add the managed role spicedb and the Database spicedb " +
				"of the release's database file to yours and apply it (install step 5); CloudNativePG creates the role once the Secret openfief-db-spicedb exists",
		},
		{
			name: "a key SpiceDB refuses", code: codes.Unauthenticated, store: spicedbStore{role: true, database: true},
			want: "SpiceDB at openfief-spicedb:50051 refuses the core's preshared key: SpiceDB and the core read different keys of the Secret openfief-spicedb-key; restart SpiceDB, then the core",
		},
		{
			name: "SpiceDB down with its store there", code: codes.Unavailable, store: spicedbStore{role: true, database: true},
			want: "SpiceDB at openfief-spicedb:50051 has not answered; the openfief-db cluster has its role and database spicedb, " +
				"so check that the Deployment openfief-spicedb is ready and that the Secret openfief-db-spicedb exists",
		},
		{
			name: "SpiceDB down and the database not asked", code: codes.Unavailable, storeErr: errNoDatabase,
			want: "SpiceDB at openfief-spicedb:50051 has not answered, and the database could not be asked whether it has SpiceDB's role spicedb and database spicedb",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := spicedbWaitReason(testEndpoint, tt.code, tt.store, tt.storeErr); got != tt.want {
				t.Errorf("spicedbWaitReason =\n%q\nwant\n%q", got, tt.want)
			}
		})
	}
}

// storeRow is a pgx.Row of the diagnosis's one query: the role's and the
// database's existence, or an error.
type storeRow struct {
	store spicedbStore
	err   error
}

func (r storeRow) Scan(dest ...any) error {
	if r.err != nil {
		return r.err
	}
	if len(dest) != 2 {
		return fmt.Errorf("%w, not %d columns", errScanShape, len(dest))
	}
	role, okRole := dest[0].(*bool)
	database, okDatabase := dest[1].(*bool)
	if !okRole || !okDatabase {
		return errScanShape
	}
	*role, *database = r.store.role, r.store.database
	return nil
}

// fakeCatalog answers the diagnosis's query with row and keeps what it was
// asked.
type fakeCatalog struct {
	row  storeRow
	sql  string
	args []any
}

func (c *fakeCatalog) QueryRow(_ context.Context, sql string, args ...any) pgx.Row {
	c.sql, c.args = sql, args
	return c.row
}

// TestSpiceDBDiagnosis checks that the diagnosis asks the database's catalog
// for SpiceDB's role and database by name, and reads the failed call's gRPC
// code through the errors that wrap it.
func TestSpiceDBDiagnosis(t *testing.T) {
	t.Parallel()
	catalog := &fakeCatalog{}
	diagnose := spicedbDiagnosis(catalog, testEndpoint)
	failed := fmt.Errorf("reconciling SpiceDB: writing the schema: %w", status.Error(codes.Unavailable, "connection refused"))
	if got, want := diagnose(t.Context(), failed), spicedbWaitReason(testEndpoint, codes.Unavailable, spicedbStore{}, nil); got != want {
		t.Errorf("with neither role nor database = %q, want %q", got, want)
	}
	if !strings.Contains(catalog.sql, "pg_catalog.pg_roles") || !strings.Contains(catalog.sql, "pg_catalog.pg_database") ||
		fmt.Sprint(catalog.args) != "[spicedb spicedb]" {
		t.Errorf("asked %q with %v, want pg_roles and pg_database for spicedb", catalog.sql, catalog.args)
	}
	catalog.row = storeRow{store: spicedbStore{role: true, database: true}}
	refused := fmt.Errorf("reconciling SpiceDB: %w", status.Error(codes.Unauthenticated, "invalid preshared key"))
	if got := diagnose(t.Context(), refused); !strings.Contains(got, "refuses the core's preshared key") {
		t.Errorf("with a refused key = %q, want the key named", got)
	}
	catalog.row = storeRow{err: errNoDatabase}
	if got := diagnose(t.Context(), failed); !strings.Contains(got, "the database could not be asked") {
		t.Errorf("with the database unreachable = %q, want that said", got)
	}
}

// TestNotReadySaysWhatIsMissing checks the upgrade that skipped install step
// 5: the core has its key, SpiceDB doesn't answer and the database has no
// role or database spicedb. Every check stays refused, /readyz is 503 and
// says what is missing, by name, and so does the log, at every attempt; once
// SpiceDB answers the core is ready.
func TestNotReadySaysWhatIsMissing(t *testing.T) {
	t.Parallel()
	perms := &failingPerms{}
	perms.down.Store(true)
	catalog := &fakeCatalog{}
	az := authorization{
		authorizer: &scriptedAuthorizer{}, syncer: spicedb.NewSyncer(perms),
		start: &startState{diagnose: spicedbDiagnosis(catalog, testEndpoint)},
	}
	s := readyServer(az)
	load := func(context.Context) (spicedb.Desired, error) { return spicedb.Desired{}, nil }
	if got := readyz(t, s); got != (answer{http.StatusServiceUnavailable, schemaPending}) {
		t.Fatalf("GET /readyz before the first attempt = %v, want 503 %q", got, schemaPending)
	}
	var logged bytes.Buffer
	reconcileAccess(t.Context(), az.syncer, load, slog.New(slog.NewTextHandler(&logged, nil)), az.markStarted, az.noteFailure)
	missing := spicedbWaitReason(testEndpoint, codes.Unknown, spicedbStore{}, nil)
	if got, want := readyz(t, s), (answer{http.StatusServiceUnavailable, schemaPending + ": " + missing}); got != want {
		t.Fatalf("GET /readyz with no role and no database =\n%v\nwant\n%v", got, want)
	}
	for _, want := range []string{"level=WARN", "every authorization check is refused and the core is not ready until it is", "no role spicedb and no database spicedb", "install step 5"} {
		if !strings.Contains(logged.String(), want) {
			t.Errorf("the log lacks %q:\n%s", want, logged.String())
		}
	}
	perms.down.Store(false)
	reconcileAccess(t.Context(), az.syncer, load, slog.New(slog.DiscardHandler), az.markStarted, az.noteFailure)
	if got := readyz(t, s); got != (answer{http.StatusOK, "ok"}) {
		t.Errorf("GET /readyz once SpiceDB answered = %v, want 200 ok", got)
	}
}

// TestReadSpiceDBStoreOnPostgres checks, under task test:pg, that the
// diagnosis's query runs on a real PostgreSQL as the core's own role, which
// has no right of its own on the catalog, and answers what the server holds:
// whether the role spicedb exists (to_regrole) and the database spicedb.
func TestReadSpiceDBStoreOnPostgres(t *testing.T) {
	t.Parallel()
	pool, err := pgxpool.New(t.Context(), emptyDatabase(t))
	if err != nil {
		t.Fatalf("connecting: %v", err)
	}
	defer pool.Close()
	got, err := readSpiceDBStore(t.Context(), pool)
	if err != nil {
		t.Fatalf("readSpiceDBStore: %v", err)
	}
	var want spicedbStore
	if err := pool.QueryRow(t.Context(), `SELECT to_regrole('spicedb') IS NOT NULL,
		EXISTS (SELECT FROM pg_catalog.pg_database WHERE datname = 'spicedb')`).Scan(&want.role, &want.database); err != nil {
		t.Fatalf("asking the server: %v", err)
	}
	if got != want {
		t.Errorf("readSpiceDBStore = %+v, want %+v", got, want)
	}
}

// demoLedger is what a fake load reads: the accounts the loader has written
// so far, each in a sleeve; each reconcile is one receive on calls, and
// answers what the test sends on results.
type demoLedger struct {
	mu       sync.Mutex
	accounts []string
	calls    chan struct{}
	results  chan error
}

func newDemoLedger(accounts ...string) *demoLedger {
	return &demoLedger{accounts: accounts, calls: make(chan struct{}), results: make(chan error)}
}

func (d *demoLedger) set(accounts ...string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.accounts = accounts
}

func (d *demoLedger) load(context.Context) (spicedb.Desired, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	sleeves := make(map[string]string, len(d.accounts))
	for _, a := range d.accounts {
		sleeves[a] = "core"
	}
	return spicedb.Desired{Accounts: slices.Clone(d.accounts), Sleeves: sleeves, DemoSessions: true}, nil
}

func (d *demoLedger) reconcile(ctx context.Context) error {
	select {
	case d.calls <- struct{}{}:
	case <-ctx.Done():
		return fmt.Errorf("the fake reconcile: %w", ctx.Err())
	}
	select {
	case err := <-d.results:
		return err
	case <-ctx.Done():
		return fmt.Errorf("the fake reconcile: %w", ctx.Err())
	}
}

// answer waits for the next reconcile and has it return err.
func (d *demoLedger) answer(t *testing.T, err error) {
	t.Helper()
	waitCalls(t, d.calls, 1)
	d.results <- err
}

// quietFor fails t when a reconcile runs within a few polls.
func quietFor(t *testing.T, calls <-chan struct{}, what string) {
	t.Helper()
	select {
	case <-calls:
		t.Fatalf("reconciled %s", what)
	case <-time.After(20 * time.Millisecond):
	}
}

// TestDemoFollowsTheLoader checks that a demo core reconciles SpiceDB as soon
// as a poll reads accounts or sleeves other than those it last reconciled, as
// after the demo's loader reseeded the ledger next to it: nothing before the
// start's reconcile succeeded, once when it first reads the accounts, not
// again while they stay the same, nothing while the ledger is empty between
// the loader's drop and its accounts, again once they are back; a failure is
// logged as a warning and retried at the next poll; and it stops with ctx.
func TestDemoFollowsTheLoader(t *testing.T) {
	t.Parallel()
	az := authorization{start: new(startState)}
	d := newDemoLedger("demo-ibkr", "demo-xtb")
	var logged lockedBuffer
	ctx, stop := context.WithCancel(t.Context())
	done := make(chan struct{})
	go func() {
		az.followDemo(ctx, d.load, d.reconcile, time.Millisecond, time.Hour, slog.New(slog.NewTextHandler(&logged, nil)))
		close(done)
	}()
	quietFor(t, d.calls, "before the start's reconcile succeeded")
	az.markStarted()
	d.answer(t, nil)
	quietFor(t, d.calls, "again with the same accounts")
	d.set() // the loader dropped the ledger
	quietFor(t, d.calls, "an empty ledger")
	d.set("demo-ibkr", "demo-xtb", "demo-bonds") // and wrote its accounts again
	d.answer(t, errNoDatabase)                   // a failure is retried at the next poll
	d.answer(t, nil)
	quietFor(t, d.calls, "once the reseed's accounts were reconciled")
	stop()
	select {
	case <-done:
	case <-time.After(waitDeadline):
		t.Fatal("followDemo did not stop with its context")
	}
	if out := logged.String(); !strings.Contains(out, "level=WARN") || !strings.Contains(out, "reconcile of the demo's accounts failed") {
		t.Errorf("a failed reconcile was not logged as a warning:\n%s", out)
	}
}

// TestDemoSweepsWhateverItReads checks the backstop: every sweep reconciles
// although the accounts it reads have not changed.
func TestDemoSweepsWhateverItReads(t *testing.T) {
	t.Parallel()
	az := authorization{start: new(startState)}
	az.markStarted()
	d := newDemoLedger("demo-xtb")
	go az.followDemo(t.Context(), d.load, d.reconcile, time.Hour, time.Millisecond, slog.New(slog.DiscardHandler))
	for range 3 {
		d.answer(t, nil)
	}
}

// TestPlacementsIgnoreOrder checks that the same accounts and sleeves in
// another order read as the same, and a moved account as a change.
func TestPlacementsIgnoreOrder(t *testing.T) {
	t.Parallel()
	a := spicedb.Desired{Accounts: []string{"b", "a"}, Sleeves: map[string]string{"a": "core", "b": "bonds"}}
	b := spicedb.Desired{Accounts: []string{"a", "b"}, Sleeves: map[string]string{"a": "core", "b": "bonds"}}
	moved := spicedb.Desired{Accounts: []string{"a", "b"}, Sleeves: map[string]string{"a": "bonds", "b": "bonds"}}
	if placements(&a) != placements(&b) {
		t.Error("the same placements in another order read as a change")
	}
	if placements(&a) == placements(&moved) {
		t.Error("a moved account reads as no change")
	}
}

// lockedBuffer is a bytes.Buffer a logger may write while the test reads it.
type lockedBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (b *lockedBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	n, err := b.buf.Write(p)
	if err != nil {
		return n, fmt.Errorf("buffering a log line: %w", err)
	}
	return n, nil
}

func (b *lockedBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.String()
}

// waitCalls waits for n receives on calls, failing t at waitDeadline.
func waitCalls(t *testing.T, calls <-chan struct{}, n int) {
	t.Helper()
	for range n {
		select {
		case <-calls:
		case <-time.After(waitDeadline):
			t.Fatal("no reconcile came")
		}
	}
}
