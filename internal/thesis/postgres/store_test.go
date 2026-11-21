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

	"github.com/arkadiuss/openfief/internal/thesis"
	"github.com/arkadiuss/openfief/internal/thesis/postgres"
)

// dsnVar names the server the tests create their databases on. task test:pg
// sets it to a disposable server; plain go test leaves it unset, and the
// tests skip, so task test stays offline.
const dsnVar = "OPENFIEF_PG_DSN"

// newPool returns a pool on a fresh database with the thesis migrations
// applied twice, dropped when the test ends.
func newPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	dsn := os.Getenv(dsnVar)
	if dsn == "" {
		t.Skip(dsnVar + " is not set: task test:pg runs this test on a disposable PostgreSQL")
	}
	ctx := t.Context()
	sum := sha256.Sum256([]byte(t.Name()))
	name := "thesis_" + hex.EncodeToString(sum[:8]) + "_" + strconv.Itoa(os.Getpid())
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

const id = "00000000000000a1"

// errHook is the error of an audit entry that could not be written.
var errHook = errors.New("the audit entry could not be written")

// day is the time the tests stamp versions with.
func day() time.Time { return time.Date(2026, 11, 21, 9, 0, 0, 0, time.UTC) }

func draft(statement string) *thesis.Draft {
	return &thesis.Draft{
		Statement: statement, Instruments: []string{"EXMP", "OTHR"}, Horizon: thesis.HorizonShortTerm,
		Invalidation: []string{"Growth stalls."}, Sources: []string{"The holdings card."},
	}
}

func TestSaveEditAndGetKeepEveryVersion(t *testing.T) {
	t.Parallel()
	store := postgres.New(newPool(t))
	ctx := t.Context()
	if err := store.Save(ctx, id, 0, day(), draft("first"), nil); err != nil {
		t.Fatalf("Save v1: %v", err)
	}
	if err := store.Save(ctx, id, 1, day().Add(time.Hour), draft("second"), nil); err != nil {
		t.Fatalf("Save v2: %v", err)
	}
	got, err := store.Get(ctx, id)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	want := thesis.Thesis{ID: id, Versions: []thesis.Version{
		{Number: 1, SavedAt: day(), Draft: *draft("first")},
		{Number: 2, SavedAt: day().Add(time.Hour), Draft: *draft("second")},
	}}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("Get (-want +got):\n%s", diff)
	}
}

func TestStaleAndMissingSavesStoreNothing(t *testing.T) {
	t.Parallel()
	store := postgres.New(newPool(t))
	ctx := t.Context()
	if err := store.Save(ctx, id, 0, day(), draft("first"), nil); err != nil {
		t.Fatalf("Save v1: %v", err)
	}
	if err := store.Save(ctx, id, 1, day(), draft("second"), nil); err != nil {
		t.Fatalf("Save v2: %v", err)
	}
	tests := []struct {
		name string
		id   string
		base int
		want error
	}{
		{"an edit on a version that is not the latest", id, 1, thesis.ErrStale},
		{"a create of an id that is taken", id, 0, thesis.ErrStale},
		{"an edit of a thesis that does not exist", "00000000000000b2", 1, thesis.ErrNotFound},
	}
	for _, tt := range tests {
		if err := store.Save(ctx, tt.id, tt.base, day(), draft("third"), nil); !errors.Is(err, tt.want) {
			t.Errorf("Save, %s: error = %v, want %v", tt.name, err, tt.want)
		}
	}
	got, err := store.Get(ctx, id)
	if err != nil || len(got.Versions) != 2 {
		t.Errorf("Get = %d versions, %v, want the two that were saved", len(got.Versions), err)
	}
}

func TestAFailingHookRollsTheVersionBack(t *testing.T) {
	t.Parallel()
	store := postgres.New(newPool(t))
	ctx := t.Context()
	boom := errHook
	if err := store.Save(ctx, id, 0, day(), draft("first"), func(context.Context) error { return boom }); !errors.Is(err, boom) {
		t.Fatalf("Save error = %v, want the hook's", err)
	}
	if _, err := store.Get(ctx, id); !errors.Is(err, thesis.ErrNotFound) {
		t.Errorf("Get after a failed hook: error = %v, want ErrNotFound", err)
	}
}

func TestListIsTheLatestVersionsNewestChangeFirst(t *testing.T) {
	t.Parallel()
	store := postgres.New(newPool(t))
	ctx := t.Context()
	other := "00000000000000b2"
	for _, s := range []struct {
		id   string
		base int
		at   time.Time
		text string
	}{
		{id, 0, day(), "a v1"}, {other, 0, day().Add(time.Hour), "b v1"}, {id, 1, day().Add(2 * time.Hour), "a v2"},
	} {
		if err := store.Save(ctx, s.id, s.base, s.at, draft(s.text), nil); err != nil {
			t.Fatalf("Save %s: %v", s.text, err)
		}
	}
	got, err := store.List(ctx)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	want := []thesis.Summary{
		{ID: id, Version: 2, CreatedAt: day(), UpdatedAt: day().Add(2 * time.Hour), Draft: *draft("a v2")},
		{ID: other, Version: 1, CreatedAt: day().Add(time.Hour), UpdatedAt: day().Add(time.Hour), Draft: *draft("b v1")},
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("List (-want +got):\n%s", diff)
	}
}

func TestASavedVersionIsNeverChangedOrRemoved(t *testing.T) {
	t.Parallel()
	pool := newPool(t)
	store := postgres.New(pool)
	ctx := t.Context()
	if err := store.Save(ctx, id, 0, day(), draft("first"), nil); err != nil {
		t.Fatalf("Save: %v", err)
	}
	for _, sql := range []string{
		"UPDATE thesis.thesis SET statement = 'changed'",
		"UPDATE thesis.thesis SET statement = 'changed' WHERE false",
		"DELETE FROM thesis.thesis",
		"TRUNCATE thesis.thesis",
	} {
		if _, err := pool.Exec(ctx, sql); err == nil {
			t.Errorf("%s: the database accepted it", sql)
		}
	}
	got, err := store.Get(ctx, id)
	if err != nil || got.Latest().Draft.Statement != "first" {
		t.Errorf("Get = %+v, %v, want the saved text", got, err)
	}
}

func TestTheTableRefusesWhatTheDomainRefuses(t *testing.T) {
	t.Parallel()
	pool := newPool(t)
	for name, sql := range map[string]string{
		"a horizon that is neither": "INSERT INTO thesis.thesis VALUES ('00000000000000c3', 1, now(), 's', 'medium_term', '{a}', '{b}', '{c}')",
		"no instrument":             "INSERT INTO thesis.thesis VALUES ('00000000000000c3', 1, now(), 's', 'long_term', '{}', '{b}', '{c}')",
		"no invalidation condition": "INSERT INTO thesis.thesis VALUES ('00000000000000c3', 1, now(), 's', 'long_term', '{a}', '{}', '{c}')",
		"an id of the wrong shape":  "INSERT INTO thesis.thesis VALUES ('nope', 1, now(), 's', 'long_term', '{a}', '{b}', '{c}')",
		"an empty statement":        "INSERT INTO thesis.thesis VALUES ('00000000000000c3', 1, now(), '', 'long_term', '{a}', '{b}', '{c}')",
		"a version below the first": "INSERT INTO thesis.thesis VALUES ('00000000000000c3', 0, now(), 's', 'long_term', '{a}', '{b}', '{c}')",
	} {
		if _, err := pool.Exec(t.Context(), sql); err == nil {
			t.Errorf("%s: the database accepted it", name)
		}
	}
}
