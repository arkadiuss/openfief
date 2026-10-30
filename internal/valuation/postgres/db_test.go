package postgres_test

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"os"
	"strconv"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
	"github.com/arkadiuss/openfief/internal/valuation"
	"github.com/arkadiuss/openfief/internal/valuation/postgres"
)

// dsnVar names the server the tests create their databases on. task test:pg
// sets it to a disposable server; plain go test leaves it unset, and the
// tests skip, so task test stays offline.
const dsnVar = "OPENFIEF_PG_DSN"

// Synthetic accounts and ISINs with valid check digits; nothing real is
// implied.
const (
	acc1 ledger.AccountID = "acc-1"
	acc2 ledger.AccountID = "acc-2"

	etfISIN  = "IETEST000026"
	bondISIN = "PL0000000014"
	// otherISIN is one more synthetic ISIN with a valid check digit.
	otherISIN = "PLTEST000010"
)

// newPool returns a pool on a fresh database with every valuation and River
// migration applied, dropped when the test ends.
func newPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	dsn := os.Getenv(dsnVar)
	if dsn == "" {
		t.Skip(dsnVar + " is not set: task test:pg runs this test on a disposable PostgreSQL")
	}
	ctx := t.Context()
	sum := sha256.Sum256([]byte(t.Name()))
	name := "valuation_" + hex.EncodeToString(sum[:8]) + "_" + strconv.Itoa(os.Getpid())
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
	if err := postgres.Migrate(ctx, pool); err != nil {
		t.Fatalf("Migrate: %v", err)
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

// newStore returns a store on a fresh database.
func newStore(t *testing.T) (*postgres.Store, *pgxpool.Pool) {
	t.Helper()
	pool := newPool(t)
	store, err := postgres.NewStore(pool)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	return store, pool
}

func dec(t *testing.T, s string) kernel.Decimal {
	t.Helper()
	d, err := kernel.ParseDecimal(s)
	if err != nil {
		t.Fatalf("ParseDecimal(%q): %v", s, err)
	}
	return d
}

func money(t *testing.T, amount, code string) kernel.Money {
	t.Helper()
	c, err := kernel.ParseCurrency(code)
	if err != nil {
		t.Fatalf("ParseCurrency(%q): %v", code, err)
	}
	return kernel.NewMoney(dec(t, amount), c)
}

func date(t *testing.T, s string) kernel.Date {
	t.Helper()
	d, err := kernel.ParseDate(s)
	if err != nil {
		t.Fatalf("ParseDate(%q): %v", s, err)
	}
	return d
}

func instrument(t *testing.T, isin, symbol string) ledger.Instrument {
	t.Helper()
	i, err := ledger.ParseISIN(isin)
	if err != nil {
		t.Fatalf("ParseISIN(%q): %v", isin, err)
	}
	return ledger.Instrument{ISIN: i, Symbol: symbol}
}

func reval(t *testing.T, account ledger.AccountID, from string) valuation.Revaluation {
	t.Helper()
	return valuation.Revaluation{Account: account, From: date(t, from)}
}

// textRows returns the single text column of sql's rows.
func textRows(t *testing.T, pool *pgxpool.Pool, sql string) []string {
	t.Helper()
	rows, err := pool.Query(t.Context(), sql)
	if err != nil {
		t.Fatalf("%s: %v", sql, err)
	}
	out, err := pgx.CollectRows(rows, pgx.RowTo[string])
	if err != nil {
		t.Fatalf("%s: %v", sql, err)
	}
	return out
}

// dirtyRows returns valuation.dirty as "account from_date claimed_from",
// with "-" for NULL.
func dirtyRows(t *testing.T, pool *pgxpool.Pool) []string {
	t.Helper()
	return textRows(t, pool, `
		SELECT account_id || ' ' || coalesce(from_date::text, '-') || ' ' || coalesce(claimed_from::text, '-')
		FROM valuation.dirty ORDER BY account_id`)
}

// liveJobs returns the RevalueAccount jobs waiting or running as
// "account state".
func liveJobs(t *testing.T, pool *pgxpool.Pool) []string {
	t.Helper()
	return textRows(t, pool, `
		SELECT (args->>'account') || ' ' || state FROM river.river_job
		WHERE kind = 'valuation.revalue_account' AND state NOT IN ('completed', 'cancelled', 'discarded')
		ORDER BY args->>'account'`)
}

// checkPending compares valuation.dirty and the live jobs with the wanted
// rows (dirtyRows, liveJobs).
func checkPending(t *testing.T, pool *pgxpool.Pool, wantDirty, wantJobs []string) {
	t.Helper()
	if diff := cmp.Diff(wantDirty, dirtyRows(t, pool), cmpopts.EquateEmpty()); diff != "" {
		t.Errorf("valuation.dirty (-want +got):\n%s", diff)
	}
	if diff := cmp.Diff(wantJobs, liveJobs(t, pool), cmpopts.EquateEmpty()); diff != "" {
		t.Errorf("jobs (-want +got):\n%s", diff)
	}
}

// countRows returns the count a counting query returns.
func countRows(t *testing.T, pool *pgxpool.Pool, sql string) int {
	t.Helper()
	var n int
	if err := pool.QueryRow(t.Context(), sql).Scan(&n); err != nil {
		t.Fatalf("%s: %v", sql, err)
	}
	return n
}
