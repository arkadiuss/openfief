package postgres_test

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"os"
	"strconv"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
	"github.com/arkadiuss/openfief/internal/ledger/app"
	"github.com/arkadiuss/openfief/internal/ledger/postgres"
	"github.com/arkadiuss/openfief/internal/portfolio"
)

// dsnVar names the server the tests create their databases on. task test:pg
// sets it to a disposable server; plain go test leaves it unset, and the
// tests skip, so task test stays offline.
const dsnVar = "OPENFIEF_PG_DSN"

// newPool returns a pool on a fresh database with every migration applied,
// dropped when the test ends.
func newPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	dsn := os.Getenv(dsnVar)
	if dsn == "" {
		t.Skip(dsnVar + " is not set: task test:pg runs this test on a disposable PostgreSQL")
	}
	ctx := t.Context()
	sum := sha256.Sum256([]byte(t.Name()))
	name := "ledger_" + hex.EncodeToString(sum[:8]) + "_" + strconv.Itoa(os.Getpid())
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

// sqlState returns err's PostgreSQL SQLSTATE, or "" if it has none.
func sqlState(err error) string {
	if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
		return pgErr.Code
	}
	return ""
}

// Synthetic IDs and ISINs with valid check digits; nothing real is implied.
const (
	isinA = "PLTEST000010"
	isinB = "IETEST000026"

	acc1 ledger.AccountID   = "acc-1"
	acc2 ledger.AccountID   = "acc-2"
	core portfolio.SleeveID = "core"
)

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

func qty(t *testing.T, s string) kernel.Quantity {
	t.Helper()
	return kernel.NewQuantity(dec(t, s))
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

// must unwraps a transaction constructor's result: must(t)(ledger.NewBuy(...)).
func must(t *testing.T) func(ledger.Transaction, error) ledger.Transaction {
	t.Helper()
	return func(tx ledger.Transaction, err error) ledger.Transaction {
		t.Helper()
		if err != nil {
			t.Fatalf("building transaction: %v", err)
		}
		return tx
	}
}

// header returns a header in account from source "test".
func header(t *testing.T, account ledger.AccountID, id, day string) ledger.TransactionHeader {
	t.Helper()
	return ledger.TransactionHeader{
		Key:     ledger.TransactionKey{Source: "test", ExternalID: id},
		Account: account,
		Date:    date(t, day),
	}
}

func versioned(t *ledger.Transaction, version int) ledger.VersionedTransaction {
	return ledger.VersionedTransaction{Transaction: *t, Version: version}
}

// seededStore returns a store on a fresh database with the sleeve core and
// the accounts acc-1 and acc-2.
func seededStore(t *testing.T) (*postgres.Store, *pgxpool.Pool) {
	t.Helper()
	pool := newPool(t)
	store := postgres.New(pool)
	if err := store.SaveSleeve(t.Context(), portfolio.Sleeve{ID: core, Name: "Core", Role: portfolio.RoleDefensive}); err != nil {
		t.Fatalf("SaveSleeve: %v", err)
	}
	for _, id := range []ledger.AccountID{acc1, acc2} {
		if err := store.SaveAccount(t.Context(), ledger.Account{ID: id, Institution: "test-broker", Name: string(id), Sleeve: core}); err != nil {
			t.Fatalf("SaveAccount(%s): %v", id, err)
		}
	}
	return store, pool
}

// appendAll appends vs in order.
func appendAll(t *testing.T, store app.TransactionLog, vs []ledger.VersionedTransaction) {
	t.Helper()
	for i := range vs {
		if err := store.Append(t.Context(), &vs[i]); err != nil {
			t.Fatalf("Append(%s v%d): %v", vs[i].Transaction.Key, vs[i].Version, err)
		}
	}
}
