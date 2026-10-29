package postgres_test

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/arkadiuss/openfief/internal/ingest"
	"github.com/arkadiuss/openfief/internal/ingest/app"
	"github.com/arkadiuss/openfief/internal/ingest/postgres"
	"github.com/arkadiuss/openfief/internal/ledger"
)

// dsnVar names the server the tests create their databases on. task test:pg
// sets it to a disposable server; plain go test leaves it unset, and the
// tests skip.
const dsnVar = "OPENFIEF_PG_DSN"

// newPool returns a pool on a fresh database with the ingest migrations
// applied, dropped when the test ends.
func newPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	dsn := os.Getenv(dsnVar)
	if dsn == "" {
		t.Skip(dsnVar + " is not set: task test:pg runs this test on a disposable PostgreSQL")
	}
	ctx := t.Context()
	sum := sha256.Sum256([]byte(t.Name()))
	name := "ingest_" + hex.EncodeToString(sum[:8]) + "_" + strconv.Itoa(os.Getpid())
	adminExec(ctx, t, dsn, "DROP DATABASE IF EXISTS "+name+" WITH (FORCE)")
	adminExec(ctx, t, dsn, "CREATE DATABASE "+name)
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		t.Fatalf("parsing %s: %v", dsnVar, err)
	}
	cfg.ConnConfig.Database = name
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		t.Fatalf("connecting to %s: %v", name, err)
	}
	t.Cleanup(func() {
		pool.Close()
		adminExec(context.WithoutCancel(ctx), t, dsn, "DROP DATABASE IF EXISTS "+name+" WITH (FORCE)")
	})
	for range 2 { // a second run applies nothing
		if err := postgres.Migrate(ctx, pool); err != nil {
			t.Fatalf("Migrate: %v", err)
		}
	}
	return pool
}

// adminExec runs sql on the server's own database.
func adminExec(ctx context.Context, t *testing.T, dsn, sql string) {
	t.Helper()
	conn, err := pgx.Connect(ctx, dsn)
	if err != nil {
		t.Fatalf("connecting to %s: %v", dsnVar, err)
	}
	defer func() {
		if err := conn.Close(ctx); err != nil {
			t.Errorf("closing: %v", err)
		}
	}()
	if _, err := conn.Exec(ctx, sql); err != nil {
		t.Fatalf("%s: %v", sql, err)
	}
}

// t0 is when the store tests begin.
func t0() time.Time { return time.Date(2026, time.March, 2, 6, 0, 0, 0, time.UTC) }

// claim claims instance from the given time for ten minutes, failing the test on error.
func claim(t *testing.T, s *postgres.Store, instance string, from time.Time) app.SyncClaim {
	t.Helper()
	c, err := s.ClaimSync(t.Context(), instance, from, from.Add(10*time.Minute))
	if err != nil {
		t.Fatalf("ClaimSync at %s: %v", from, err)
	}
	return c
}

// batchState returns a stored batch's state and attempts, such as
// "completed after 2".
func batchState(t *testing.T, pool *pgxpool.Pool, id string) string {
	t.Helper()
	var (
		state    string
		attempts int
	)
	if err := pool.QueryRow(t.Context(), "SELECT state, attempts FROM ingest.import_batches WHERE id = $1", id).Scan(&state, &attempts); err != nil {
		t.Fatalf("loading batch %s: %v", id, err)
	}
	return state + " after " + strconv.Itoa(attempts)
}

func TestSyncStoreLifecycle(t *testing.T) {
	t.Parallel()
	pool := newPool(t)
	s := postgres.New(pool)
	ctx := t.Context()

	if diff := cmp.Diff(app.SyncClaim{}, claim(t, s, "xtb-main", t0())); diff != "" {
		t.Errorf("first claim (-want +got):\n%s", diff)
	}
	if _, err := s.ClaimSync(ctx, "xtb-main", t0().Add(time.Minute), t0().Add(time.Hour)); !errors.Is(err, app.ErrSyncRunning) {
		t.Errorf("claim while claimed = %v, want ErrSyncRunning", err)
	}
	states, err := s.SyncStates(ctx)
	if err != nil {
		t.Fatalf("SyncStates: %v", err)
	}
	if diff := cmp.Diff([]app.SyncState{{Instance: "xtb-main"}}, states); diff != "" {
		t.Errorf("states before an attempt ended (-want +got):\n%s", diff)
	}

	open := app.OpenBatch{ID: "b-1", Mode: ingest.ModeOpening}
	if err := s.StartBatch(ctx, "xtb-main", open, t0()); err != nil {
		t.Fatalf("StartBatch: %v", err)
	}
	retry := t0().Add(5 * time.Minute)
	if err := s.FailSync(ctx, "xtb-main", "b-1", 1, retry, t0().Add(time.Minute)); err != nil {
		t.Fatalf("FailSync: %v", err)
	}
	want := app.SyncClaim{Failures: 1, Open: &open}
	if diff := cmp.Diff(want, claim(t, s, "xtb-main", t0().Add(2*time.Minute))); diff != "" {
		t.Errorf("claim after a failure (-want +got):\n%s", diff)
	}

	checkCompletion(t, pool, s)
}

func TestSyncStoreClaimExpires(t *testing.T) {
	t.Parallel()
	s := postgres.New(newPool(t))
	claim(t, s, "nbp", t0())
	if _, err := s.ClaimSync(t.Context(), "nbp", t0().Add(9*time.Minute), t0().Add(time.Hour)); !errors.Is(err, app.ErrSyncRunning) {
		t.Errorf("claim before the lease expired = %v, want ErrSyncRunning", err)
	}
	claim(t, s, "nbp", t0().Add(10*time.Minute)) // a crashed attempt's claim lapses
	claim(t, s, "other", t0().Add(11*time.Minute))
}

func TestSyncStoreStartBatchAbandonsTheOpenOne(t *testing.T) {
	t.Parallel()
	pool := newPool(t)
	s := postgres.New(pool)
	ctx := t.Context()
	claim(t, s, "xtb-main", t0())
	if err := s.StartBatch(ctx, "xtb-main", app.OpenBatch{ID: "b-1", Mode: ingest.ModeOpening}, t0()); err != nil {
		t.Fatalf("StartBatch b-1: %v", err)
	}
	if err := s.StartBatch(ctx, "xtb-main", app.OpenBatch{ID: "b-2", Mode: ingest.ModeTransactions}, t0()); err != nil {
		t.Fatalf("StartBatch b-2: %v", err)
	}
	if got := batchState(t, pool, "b-1"); got != "abandoned after 0" {
		t.Errorf("b-1 is %s, want abandoned after 0", got)
	}
	if err := s.FailSync(ctx, "xtb-main", "b-2", 1, t0(), t0()); err != nil {
		t.Fatalf("FailSync: %v", err)
	}
	want := app.SyncClaim{Failures: 1, Open: &app.OpenBatch{ID: "b-2", Mode: ingest.ModeTransactions}}
	if diff := cmp.Diff(want, claim(t, s, "xtb-main", t0().Add(time.Minute))); diff != "" {
		t.Errorf("claim (-want +got):\n%s", diff)
	}
	if err := s.StartBatch(ctx, "xtb-main", app.OpenBatch{ID: "b-2", Mode: ingest.ModeOpening}, t0()); err == nil {
		t.Error("StartBatch reused a batch ID")
	}
}

func TestBindingStore(t *testing.T) {
	t.Parallel()
	s := postgres.New(newPool(t))
	ctx := t.Context()
	for _, a := range []ledger.AccountID{"acc-2", "acc-1", "acc-2"} {
		if err := s.Bind(ctx, "xtb-main", a, t0()); err != nil {
			t.Fatalf("Bind %s: %v", a, err)
		}
	}
	got, err := s.BoundAccounts(ctx, "xtb-main")
	if err != nil {
		t.Fatalf("BoundAccounts: %v", err)
	}
	if diff := cmp.Diff([]ledger.AccountID{"acc-1", "acc-2"}, got); diff != "" {
		t.Errorf("bound (-want +got):\n%s", diff)
	}
	other, err := s.BoundAccounts(ctx, "nbp")
	if err != nil {
		t.Fatalf("BoundAccounts: %v", err)
	}
	if len(other) != 0 {
		t.Errorf("an unbound instance has %v", other)
	}
	if err := s.Bind(ctx, "Not-A-Name", "acc-1", t0()); err == nil {
		t.Error("the table took an invalid instance name")
	}
}

func TestBoundInstances(t *testing.T) {
	t.Parallel()
	s := postgres.New(newPool(t))
	ctx := t.Context()
	for _, instance := range []string{"xtb-main", "bonds-main"} {
		if err := s.Bind(ctx, instance, "acc-2", t0()); err != nil {
			t.Fatalf("Bind %s: %v", instance, err)
		}
	}
	instances, err := s.BoundInstances(ctx, "acc-2")
	if err != nil {
		t.Fatalf("BoundInstances: %v", err)
	}
	if diff := cmp.Diff([]string{"bonds-main", "xtb-main"}, instances); diff != "" {
		t.Errorf("instances bound to acc-2 (-want +got):\n%s", diff)
	}
	if none, noneErr := s.BoundInstances(ctx, "acc-9"); noneErr != nil || len(none) != 0 {
		t.Errorf("BoundInstances of an unbound account = %v, %v; want none", none, noneErr)
	}
}

// checkCompletion completes xtb-main's open batch b-1 and checks the state
// and the stored counts, and that a completed batch can't complete or fail.
func checkCompletion(t *testing.T, pool *pgxpool.Pool, s *postgres.Store) {
	t.Helper()
	ctx := t.Context()
	b := &ingest.Batch{
		ID: "b-1", Institution: "xtb", Source: app.SourceSync, Mode: ingest.ModeOpening, Transactions: 2, Assertions: 1,
		Imported: ingest.ImportSummary{New: 2}, Problems: []ingest.Problem{ingest.Skipped("row 3", "summary row")},
	}
	next := t0().Add(6 * time.Hour)
	if err := s.CompleteSync(ctx, "xtb-main", b, "w-1", next, t0().Add(3*time.Minute)); err != nil {
		t.Fatalf("CompleteSync: %v", err)
	}
	if diff := cmp.Diff(app.SyncClaim{Watermark: "w-1", Seeded: true}, claim(t, s, "xtb-main", t0().Add(4*time.Minute))); diff != "" {
		t.Errorf("claim after the success (-want +got):\n%s", diff)
	}
	if got := batchState(t, pool, "b-1"); got != "completed after 2" {
		t.Errorf("batch b-1 is %s, want completed after 2", got)
	}
	var problems, newTxs int
	if err := pool.QueryRow(ctx, "SELECT problems, new FROM ingest.import_batches WHERE id = 'b-1'").Scan(&problems, &newTxs); err != nil {
		t.Fatalf("loading counts: %v", err)
	}
	if problems != 1 || newTxs != 2 {
		t.Errorf("stored problems = %d, new = %d; want 1 and 2", problems, newTxs)
	}
	if err := s.CompleteSync(ctx, "xtb-main", b, "w-2", next, t0().Add(5*time.Minute)); !errors.Is(err, postgres.ErrBatchNotOpen) {
		t.Errorf("completing a completed batch = %v, want ErrBatchNotOpen", err)
	}
	if err := s.FailSync(ctx, "xtb-main", "b-1", 1, next, t0().Add(5*time.Minute)); !errors.Is(err, postgres.ErrBatchNotOpen) {
		t.Errorf("failing a completed batch = %v, want ErrBatchNotOpen", err)
	}
}
