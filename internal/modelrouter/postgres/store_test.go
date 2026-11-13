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

	"github.com/arkadiuss/openfief/internal/modelrouter"
	"github.com/arkadiuss/openfief/internal/modelrouter/postgres"
)

// dsnVar names the server the tests create their databases on. task test:pg
// sets it to a disposable server; plain go test leaves it unset, and the
// tests skip, so task test stays offline.
const dsnVar = "OPENFIEF_PG_DSN"

// newPool returns a pool on a fresh database with the modelrouter migrations
// applied twice, dropped when the test ends.
func newPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	dsn := os.Getenv(dsnVar)
	if dsn == "" {
		t.Skip(dsnVar + " is not set: task test:pg runs this test on a disposable PostgreSQL")
	}
	ctx := t.Context()
	sum := sha256.Sum256([]byte(t.Name()))
	name := "modelrouter_" + hex.EncodeToString(sum[:8]) + "_" + strconv.Itoa(os.Getpid())
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

func refs(t *testing.T, ss ...string) []modelrouter.Ref {
	t.Helper()
	out, err := modelrouter.ParseRefs(ss)
	if err != nil {
		t.Fatal(err)
	}
	return out
}

func TestSettingsAreOffUntilSaved(t *testing.T) {
	t.Parallel()
	store := postgres.New(newPool(t))
	got, err := store.Settings(t.Context())
	if err != nil {
		t.Fatalf("Settings: %v", err)
	}
	if diff := cmp.Diff(modelrouter.Settings{}, got, cmpopts.EquateEmpty()); diff != "" {
		t.Errorf("settings before any save (-want +got):\n%s", diff)
	}
}

func TestSettingsRoundTrip(t *testing.T) {
	t.Parallel()
	store := postgres.New(newPool(t))
	ctx := t.Context()
	steps := []error{
		store.SaveEnabled(ctx, true),
		store.SaveToggle(ctx, modelrouter.FeatureAsk, true),
		store.SaveToggle(ctx, modelrouter.FeatureMCP, true),
		store.SaveToggle(ctx, modelrouter.FeatureMCP, false),
		store.SaveRoute(ctx, modelrouter.RoleAsk, refs(t, "openrouter/openai/gpt-oss-120b", "ollama/qwen3:8b")),
		store.SaveRoute(ctx, modelrouter.RoleMaintainer, refs(t, "ollama/qwen3:8b")),
		store.SaveDataTierModels(ctx, modelrouter.DataTierT2, refs(t, "openrouter/openai/gpt-oss-120b", "ollama/qwen3:8b")),
		store.SaveDataTierModels(ctx, modelrouter.DataTierT0, refs(t, "ollama/qwen3:8b")),
	}
	checkSteps(t, steps)
	got, err := store.Settings(ctx)
	if err != nil {
		t.Fatalf("Settings: %v", err)
	}
	ask, err := modelrouter.NewRoute(modelrouter.RoleAsk, refs(t, "openrouter/openai/gpt-oss-120b", "ollama/qwen3:8b"))
	if err != nil {
		t.Fatal(err)
	}
	maintainer, err := modelrouter.NewRoute(modelrouter.RoleMaintainer, refs(t, "ollama/qwen3:8b"))
	if err != nil {
		t.Fatal(err)
	}
	policy, err := modelrouter.NewDataTierPolicy(map[modelrouter.DataTier][]modelrouter.Ref{
		modelrouter.DataTierT2: refs(t, "openrouter/openai/gpt-oss-120b", "ollama/qwen3:8b"),
		modelrouter.DataTierT0: refs(t, "ollama/qwen3:8b"),
	})
	if err != nil {
		t.Fatal(err)
	}
	want := modelrouter.Settings{
		Switch: modelrouter.Switch{Enabled: true, Toggles: map[modelrouter.Feature]bool{modelrouter.FeatureAsk: true}},
		Routes: map[modelrouter.Role]modelrouter.Route{modelrouter.RoleAsk: ask, modelrouter.RoleMaintainer: maintainer},
		Policy: policy,
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("settings (-want +got):\n%s", diff)
	}
}

func TestSettingsReplaceAndClear(t *testing.T) {
	t.Parallel()
	store := postgres.New(newPool(t))
	ctx := t.Context()
	steps := []error{
		store.SaveRoute(ctx, modelrouter.RoleAsk, refs(t, "a/one", "a/two", "a/three")),
		store.SaveRoute(ctx, modelrouter.RoleAsk, refs(t, "a/three", "a/one")),
	}
	checkSteps(t, steps)
	got, err := store.Settings(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(refs(t, "a/three", "a/one"), got.Routes[modelrouter.RoleAsk].Models); diff != "" {
		t.Errorf("replaced route (-want +got):\n%s", diff)
	}
	steps = []error{
		store.SaveRoute(ctx, modelrouter.RoleAsk, nil),
		store.SaveDataTierModels(ctx, modelrouter.DataTierT2, refs(t, "a/one")),
		store.SaveDataTierModels(ctx, modelrouter.DataTierT2, nil),
	}
	checkSteps(t, steps)
	cleared, err := store.Settings(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(cleared.Routes) != 0 || len(cleared.Policy) != 0 {
		t.Errorf("after clearing: routes %v, policy %v", cleared.Routes, cleared.Policy)
	}
}

// checkSteps fails the test for the first save that failed.
func checkSteps(t *testing.T, steps []error) {
	t.Helper()
	for i, err := range steps {
		if err != nil {
			t.Fatalf("step %d: %v", i, err)
		}
	}
}

func TestSettingsSwitchTurnsOffAgain(t *testing.T) {
	t.Parallel()
	store := postgres.New(newPool(t))
	ctx := t.Context()
	for _, on := range []bool{true, false} {
		if err := store.SaveEnabled(ctx, on); err != nil {
			t.Fatal(err)
		}
		got, err := store.Settings(ctx)
		if err != nil {
			t.Fatal(err)
		}
		if got.Switch.Enabled != on {
			t.Errorf("after SaveEnabled(%v): enabled = %v", on, got.Switch.Enabled)
		}
	}
}
