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
	"github.com/arkadiuss/openfief/internal/valuation"
	valuationapp "github.com/arkadiuss/openfief/internal/valuation/app"
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
	if again := count(t, pool, "SELECT count(*) FROM ledger.transaction_versions"); again != versions {
		t.Errorf("loading another day without --reset-demo: %d transaction versions, want %d unchanged", again, versions)
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
	checkOneFlaggedPriceCheck(t, vstore, today)
}

// checkOneFlaggedPriceCheck checks that the price checks, as the core's
// read API serves them, compare IBKR's cross-check prices with the stored
// ones and flag exactly one: the latest statement's first instrument
// (D-110, docs/dev/demo.md).
func checkOneFlaggedPriceCheck(t *testing.T, vstore *valuationpg.Store, today kernel.Date) {
	t.Helper()
	checks, err := valuationapp.NewCrossChecks(vstore, nil).PriceChecks(t.Context())
	if err != nil {
		t.Fatalf("PriceChecks: %v", err)
	}
	var flagged []valuation.PriceCheck
	for i := range checks {
		if checks[i].CrossCheck.Source != demo.IBKRInstitution {
			t.Errorf("price check of %s from %q, want only IBKR's", checks[i].CrossCheck.Instrument, checks[i].CrossCheck.Source)
		}
		if checks[i].Status == valuation.PriceFlagged {
			flagged = append(flagged, checks[i])
		}
	}
	if len(flagged) != 1 || !flagged[0].CrossCheck.Date.Equal(today) {
		t.Errorf("%d price checks, %d flagged (%v); want exactly one flagged, on %s", len(checks), len(flagged), flagged, today)
	}
}

// checkValued checks that account has one valuation per holding on today,
// each with every input present and all three values (D-77), and is
// valued through today, idle. A foreign lot opened before the FX rates
// start leaves its net and liquidation values missing, so this catches
// rates that don't reach back to every lot's tax rate date.
func checkValued(t *testing.T, vstore *valuationpg.Store, account ledger.AccountID, today kernel.Date, holdings int) {
	t.Helper()
	rows, err := vstore.Valuations(t.Context(), account, today, today)
	if err != nil || len(rows) != holdings {
		t.Errorf("%s: %d valuations on %s, err %v; want one per holding (%d)", account, len(rows), today, err, holdings)
	}
	for i := range rows {
		checkThreeValues(t, &rows[i])
	}
	st, err := vstore.Status(t.Context(), account)
	if err != nil || st.Pending || st.Running || !st.ValuedThrough.Equal(today) {
		t.Errorf("%s: status %+v, err %v; want valued through %s, idle", account, st, err, today)
	}
}

// TestLoadTopsUpADemoLoadedEarlier checks that a load of another day
// without --reset-demo, refused as before, first records again the inputs
// of the dataset the database holds: a demo an older loader filled, here
// one without cross-check prices and with FX rates only from go-live, gets
// them, and its revaluations are enqueued for the core.
func TestLoadTopsUpADemoLoadedEarlier(t *testing.T) {
	t.Parallel()
	pool := emptyPool(t)
	cfg := testConfig(t, "2026-10-28", false)
	if err := load(t.Context(), pool, cfg, quiet()); err != nil {
		t.Fatalf("load into an empty database: %v", err)
	}
	goLive := demo.GoLive(cfg.Today).String()
	forgetNewerInputs(t, pool, goLive)
	versions := count(t, pool, "SELECT count(*) FROM ledger.transaction_versions")

	err := load(t.Context(), pool, testConfig(t, "2026-10-31", false), quiet())
	if !errors.Is(err, errAlreadyLoaded) {
		t.Fatalf("loading another day without --reset-demo: err = %v, want errAlreadyLoaded", err)
	}
	if again := count(t, pool, "SELECT count(*) FROM ledger.transaction_versions"); again != versions {
		t.Errorf("%d transaction versions, want %d unchanged", again, versions)
	}
	if n := count(t, pool, "SELECT count(*) FROM valuation.fx_rates WHERE date < '"+goLive+"'"); n == 0 {
		t.Errorf("no FX rate before go-live after the top-up, want those of the lots' tax dates")
	}
	if n := count(t, pool, "SELECT count(*) FROM demo.marker WHERE loaded_on = '"+cfg.Today.String()+"'"); n != 1 {
		t.Errorf("the marker no longer names %s", cfg.Today)
	}
	vstore, err := valuationpg.NewStore(pool)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	checkOneFlaggedPriceCheck(t, vstore, cfg.Today)
	if st, err := vstore.Status(t.Context(), demo.IBKRAccount); err != nil || !st.Pending {
		t.Errorf("IBKR status %+v, err %v; want a revaluation pending for the core", st, err)
	}
}

// checkThreeValues checks that row has every input and a positive gross
// value, and its net and liquidation values.
func checkThreeValues(t *testing.T, row *valuation.HoldingValuation) {
	t.Helper()
	if !row.Valued() || row.PLN.Sign() <= 0 {
		t.Errorf("%s: %s on %s: missing %s (%s), PLN %s", row.Account, row.Instrument, row.Date, row.Missing, row.Detail, row.PLN)
	}
	if row.NetMissing != 0 || row.LiquidationMissing != 0 {
		t.Errorf("%s: %s on %s: net missing %s, liquidation missing %s; want both values", row.Account, row.Instrument, row.Date, row.NetMissing, row.LiquidationMissing)
	}
}

// forgetNewerInputs deletes what an older loader didn't record: every
// cross-check price and every FX rate before goLive.
func forgetNewerInputs(t *testing.T, pool *pgxpool.Pool, goLive string) {
	t.Helper()
	for _, sql := range []string{
		"DELETE FROM valuation.cross_check_prices",
		"DELETE FROM valuation.fx_rates WHERE date < '" + goLive + "'",
	} {
		if _, err := pool.Exec(t.Context(), sql); err != nil {
			t.Fatalf("%s: %v", sql, err)
		}
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
