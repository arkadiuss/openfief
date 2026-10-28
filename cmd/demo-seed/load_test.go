package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"log/slog"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/arkadiuss/openfief/internal/demo"
	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
	ledgerapp "github.com/arkadiuss/openfief/internal/ledger/app"
	ledgerpg "github.com/arkadiuss/openfief/internal/ledger/postgres"
	"github.com/arkadiuss/openfief/internal/portfolio"
	valuationpg "github.com/arkadiuss/openfief/internal/valuation/postgres"
)

// dsnVar names the server the tests create their databases on; task
// test:pg sets it to a disposable server, and without it the tests skip.
const dsnVar = "OPENFIEF_PG_DSN"

// emptyPool returns a pool on a fresh, empty database, dropped when the
// test ends. Unlike the contexts' own tests it applies no migration: the
// loader must find the database as an operator would give it.
func emptyPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	dsn := os.Getenv(dsnVar)
	if dsn == "" {
		t.Skip(dsnVar + " is not set: task test:pg runs this test on a disposable PostgreSQL")
	}
	ctx := t.Context()
	sum := sha256.Sum256([]byte(t.Name()))
	name := "demo_" + hex.EncodeToString(sum[:8]) + "_" + strconv.Itoa(os.Getpid())
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

// testConfig loads the dataset of today offline, from the committed rates.
func testConfig(t *testing.T, today string, reset bool) *config {
	t.Helper()
	d, err := kernel.ParseDate(today)
	if err != nil {
		t.Fatalf("ParseDate: %v", err)
	}
	return &config{Seed: defaultSeed, Today: d, Reset: reset, Offline: true, Wait: 2 * time.Minute}
}

// quiet discards the loader's logs.
func quiet() *slog.Logger { return slog.New(slog.DiscardHandler) }

// count returns the result of a count query.
func count(t *testing.T, pool *pgxpool.Pool, sql string) int64 {
	t.Helper()
	var n int64
	if err := pool.QueryRow(t.Context(), sql).Scan(&n); err != nil {
		t.Fatalf("%s: %v", sql, err)
	}
	return n
}

func TestLoadEndToEnd(t *testing.T) {
	t.Parallel()
	pool := emptyPool(t)
	cfg := testConfig(t, "2026-10-28", false)
	if err := load(t.Context(), pool, cfg, quiet()); err != nil {
		t.Fatalf("load into an empty database: %v", err)
	}
	checkLoaded(t, pool, cfg.Today)
	versions := count(t, pool, "SELECT count(*) FROM ledger.transaction_versions")

	if err := load(t.Context(), pool, cfg, quiet()); err != nil {
		t.Fatalf("loading the same day again: %v", err)
	}
	if again := count(t, pool, "SELECT count(*) FROM ledger.transaction_versions"); again != versions {
		t.Errorf("loading the same day again: %d transaction versions, want %d unchanged", again, versions)
	}

	next := testConfig(t, "2026-10-29", false)
	if err := load(t.Context(), pool, next, quiet()); !errors.Is(err, errAlreadyLoaded) {
		t.Errorf("loading another day without --reset-demo: err = %v, want errAlreadyLoaded", err)
	}
	next.Reset = true
	if err := load(t.Context(), pool, next, quiet()); err != nil {
		t.Fatalf("loading another day with --reset-demo: %v", err)
	}
	checkLoaded(t, pool, next.Today)
}

// checkLoaded checks that every demo account has holdings and is valued
// through today with every input present, and that no balance assertion
// was flagged.
func checkLoaded(t *testing.T, pool *pgxpool.Pool, today kernel.Date) {
	t.Helper()
	ctx := t.Context()
	store := ledgerpg.New(pool)
	l := ledgerapp.New(store)
	vstore, err := valuationpg.NewStore(pool)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	for _, a := range demo.Accounts() {
		holdings, err := l.HoldingsOn(ctx, a.ID, today)
		if err != nil || len(holdings) == 0 {
			t.Errorf("%s: %d holdings on %s, err %v; want some", a.ID, len(holdings), today, err)
		}
		checkValued(t, vstore, a.ID, today, len(holdings))
	}
	if flagged := count(t, pool, "SELECT count(*) FROM ledger.assertion_checks WHERE status <> 'matched'"); flagged != 0 {
		t.Errorf("%d flagged balance assertions, want none", flagged)
	}
}

// checkValued checks that account has one valuation per holding on today,
// each with every input present, and is valued through today, idle.
func checkValued(t *testing.T, vstore *valuationpg.Store, account ledger.AccountID, today kernel.Date, holdings int) {
	t.Helper()
	rows, err := vstore.Valuations(t.Context(), account, today, today)
	if err != nil || len(rows) != holdings {
		t.Errorf("%s: %d valuations on %s, err %v; want one per holding (%d)", account, len(rows), today, err, holdings)
	}
	for i := range rows {
		if !rows[i].Valued() || rows[i].PLN.Sign() <= 0 {
			t.Errorf("%s: %s on %s: missing %s (%s), PLN %s", account, rows[i].Instrument, today, rows[i].Missing, rows[i].Detail, rows[i].PLN)
		}
	}
	st, err := vstore.Status(t.Context(), account)
	if err != nil || st.Pending || st.Running || !st.ValuedThrough.Equal(today) {
		t.Errorf("%s: status %+v, err %v; want valued through %s, idle", account, st, err, today)
	}
}

func TestLoadRefusesADatabaseThatIsNotADemo(t *testing.T) {
	t.Parallel()
	for _, reset := range []bool{false, true} {
		t.Run("reset="+strconv.FormatBool(reset), func(t *testing.T) {
			t.Parallel()
			checkRefused(t, reset)
		})
	}
}

// checkRefused checks that the loader, with or without reset, refuses a
// database with someone's sleeve in it and leaves it as it was.
func checkRefused(t *testing.T, reset bool) {
	t.Helper()
	pool := emptyPool(t)
	ctx := t.Context()
	if err := ledgerpg.Migrate(ctx, pool); err != nil {
		t.Fatalf("Migrate: %v", err)
	}
	theirs := portfolio.Sleeve{ID: "core", Name: "Someone's sleeve", Kind: portfolio.KindCore, Role: portfolio.RoleDefensive}
	if err := ledgerapp.New(ledgerpg.New(pool)).SaveSleeve(ctx, theirs); err != nil {
		t.Fatalf("SaveSleeve: %v", err)
	}
	err := load(ctx, pool, testConfig(t, "2026-10-28", reset), quiet())
	if !errors.Is(err, demo.ErrNotDemoDatabase) {
		t.Fatalf("load: err = %v, want ErrNotDemoDatabase", err)
	}
	if n := count(t, pool, "SELECT count(*) FROM ledger.sleeves"); n != 1 {
		t.Errorf("%d sleeves after the refusal, want the 1 that was there", n)
	}
	if n := count(t, pool, "SELECT count(*) FROM information_schema.schemata WHERE schema_name IN ('demo', 'valuation', 'river')"); n != 0 {
		t.Errorf("the refusal created %d schemas, want none", n)
	}
}
