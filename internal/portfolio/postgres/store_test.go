package postgres_test

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"os"
	"strconv"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/portfolio"
	"github.com/arkadiuss/openfief/internal/portfolio/postgres"
)

// dsnVar names the server the tests create their databases on. task test:pg
// sets it to a disposable server; plain go test leaves it unset, and the
// tests skip, so task test stays offline.
const dsnVar = "OPENFIEF_PG_DSN"

// newPool returns a pool on a fresh database with the portfolio migrations
// applied twice, dropped when the test ends.
func newPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	dsn := os.Getenv(dsnVar)
	if dsn == "" {
		t.Skip(dsnVar + " is not set: task test:pg runs this test on a disposable PostgreSQL")
	}
	ctx := t.Context()
	sum := sha256.Sum256([]byte(t.Name()))
	name := "portfolio_" + hex.EncodeToString(sum[:8]) + "_" + strconv.Itoa(os.Getpid())
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

// target builds a target; currency may be empty.
func target(t *testing.T, kind portfolio.TargetKind, name, currency, weight string) portfolio.Target {
	t.Helper()
	w, err := kernel.ParseDecimal(weight)
	if err != nil {
		t.Fatal(err)
	}
	tg := portfolio.Target{Kind: kind, Name: name, Weight: w}
	if currency != "" {
		if tg.Currency, err = kernel.ParseCurrency(currency); err != nil {
			t.Fatal(err)
		}
	}
	return tg
}

func TestModelRoundTrip(t *testing.T) {
	t.Parallel()
	store := postgres.New(newPool(t))
	ctx := t.Context()
	got, err := store.Model(ctx)
	if err != nil {
		t.Fatalf("Model: %v", err)
	}
	if diff := cmp.Diff(portfolio.Model{}, got); diff != "" {
		t.Errorf("a model never saved is off and empty (-want +got):\n%s", diff)
	}
	first := portfolio.Model{Enabled: true, Targets: []portfolio.Target{
		target(t, portfolio.TargetInstrument, "IETEST000026/ETF1", "", "60.5"),
		target(t, portfolio.TargetIndex, "FAKE20", "PLN", "39.5"),
	}}
	second := portfolio.Model{Targets: []portfolio.Target{target(t, portfolio.TargetIndex, "FAKE20", "PLN", "100")}}
	for _, want := range []portfolio.Model{first, second, {}} {
		if err := store.SaveModel(ctx, want); err != nil {
			t.Fatalf("SaveModel: %v", err)
		}
		got, err := store.Model(ctx)
		if err != nil {
			t.Fatalf("Model: %v", err)
		}
		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("round trip (-want +got):\n%s", diff)
		}
	}
}

func TestModelConstraints(t *testing.T) {
	t.Parallel()
	store := postgres.New(newPool(t))
	ctx := t.Context()
	if err := store.SaveModel(ctx, portfolio.Model{Enabled: true, Targets: []portfolio.Target{
		target(t, portfolio.TargetInstrument, "IETEST000026/ETF1", "", "100"),
	}}); err != nil {
		t.Fatal(err)
	}
	bad := []portfolio.Model{
		{Targets: []portfolio.Target{target(t, portfolio.TargetIndex, "FAKE20", "", "100")}},
		{Targets: []portfolio.Target{target(t, portfolio.TargetInstrument, "IETEST000026/ETF1", "USD", "100")}},
		{Targets: []portfolio.Target{target(t, portfolio.TargetInstrument, "IETEST000026/ETF1", "", "0.001")}},
		{Targets: []portfolio.Target{target(t, portfolio.TargetInstrument, "IETEST000026/ETF1", "", "50"), target(t, portfolio.TargetInstrument, "IETEST000026/ETF1", "", "50")}},
	}
	for i, m := range bad {
		if err := store.SaveModel(ctx, m); err == nil {
			t.Errorf("bad model %d was stored", i)
		}
	}
	got, err := store.Model(ctx)
	if err != nil || !got.Enabled || len(got.Targets) != 1 {
		t.Errorf("a refused save changed the model: %+v, %v", got, err)
	}
}
