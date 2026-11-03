package postgres_test

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/arkadiuss/openfief/internal/attention"
	"github.com/arkadiuss/openfief/internal/attention/app"
	"github.com/arkadiuss/openfief/internal/attention/postgres"
)

// dsnVar names the server the tests create their databases on. task test:pg
// sets it to a disposable server; plain go test leaves it unset, and the
// tests skip, so task test stays offline.
const dsnVar = "OPENFIEF_PG_DSN"

// errStop makes a transaction roll back.
var errStop = errors.New("stop")

// at returns the time of the changes the tests store.
func at() time.Time { return time.Date(2026, 3, 2, 10, 0, 0, 0, time.UTC) }

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
	name := "attention_" + hex.EncodeToString(sum[:8]) + "_" + strconv.Itoa(os.Getpid())
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

// count returns how many rows table holds.
func count(t *testing.T, pool *pgxpool.Pool, table string) int {
	t.Helper()
	var n int
	if err := pool.QueryRow(t.Context(), "SELECT count(*) FROM attention."+table).Scan(&n); err != nil {
		t.Fatalf("counting %s: %v", table, err)
	}
	return n
}

// raise is the step that raises an item of kind and subject.
func raise(kind attention.Kind, subject string) attention.Step {
	return attention.Step{Kind: kind, Subject: subject, Change: attention.Change{
		Seq: 1, State: attention.StateOpen, Reason: attention.ReasonRaised, Detail: "found", At: at(),
	}}
}

func TestAppendKeepsHistoryAndWritesItsEvent(t *testing.T) {
	t.Parallel()
	pool := newPool(t)
	s := postgres.New(pool)
	ctx := t.Context()
	step := raise(attention.KindBrokerAccountUnmapped, "ibkr/U0000001")
	id, err := s.Append(ctx, &step)
	if err != nil {
		t.Fatalf("Append: %v", err)
	}
	ack := attention.Step{Kind: step.Kind, Subject: step.Subject, ItemID: id, Change: attention.Change{
		Seq: 2, State: attention.StateAcknowledged, Reason: attention.ReasonAcknowledged, Detail: "found", At: at().Add(time.Minute),
	}}
	if _, ackErr := s.Append(ctx, &ack); ackErr != nil {
		t.Fatalf("Append: %v", ackErr)
	}
	want := attention.Item{ID: id, Kind: step.Kind, Subject: step.Subject, Changes: []attention.Change{step.Change, ack.Change}}
	got, ok, err := s.Item(ctx, id)
	if err != nil || !ok {
		t.Fatalf("Item = %v, %v", ok, err)
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("Item (-want +got):\n%s", diff)
	}
	if _, found, itemErr := s.Item(ctx, id+1); found || itemErr != nil {
		t.Errorf("Item of an unknown ID = %v, %v", found, itemErr)
	}
	checkOutbox(t, s, id, &step, &ack)
}

// checkOutbox checks that s's outbox holds the events of step and ack to
// item id, in order, and that marking one published leaves the other.
func checkOutbox(t *testing.T, s *postgres.Store, id int64, step, ack *attention.Step) {
	t.Helper()
	ctx := t.Context()
	events, err := s.Unpublished(ctx, 10)
	if err != nil {
		t.Fatalf("Unpublished: %v", err)
	}
	// The events name the subject as read access sees it: only the
	// connector instance of an unmapped broker account (D-121, D-165).
	subject := attention.ReadSubject(step.Kind, step.Subject)
	wantEvents := []app.Event{
		{ID: 1, ItemID: id, Kind: step.Kind, Subject: subject, Change: step.Change},
		{ID: 2, ItemID: id, Kind: step.Kind, Subject: subject, Change: ack.Change},
	}
	if diff := cmp.Diff(wantEvents, events); diff != "" {
		t.Errorf("Unpublished (-want +got):\n%s", diff)
	}
	if err := s.MarkPublished(ctx, 1, at()); err != nil {
		t.Fatalf("MarkPublished: %v", err)
	}
	if err := s.MarkPublished(ctx, 1, at().Add(time.Hour)); err != nil {
		t.Fatalf("MarkPublished again: %v", err)
	}
	if events, err := s.Unpublished(ctx, 10); err != nil || len(events) != 1 || events[0].ID != 2 {
		t.Errorf("Unpublished after marking event 1 = %v, %v", events, err)
	}
}

func TestChangeAndEventCommitTogether(t *testing.T) {
	t.Parallel()
	pool := newPool(t)
	s := postgres.New(pool)
	ctx := t.Context()
	step := raise(attention.KindSettingMissing, "marginal-rate")
	err := s.InTx(ctx, func(tx app.Store) error {
		if _, err := tx.Append(ctx, &step); err != nil {
			return fmt.Errorf("appending: %w", err)
		}
		return errStop
	})
	if !errors.Is(err, errStop) {
		t.Fatalf("InTx = %v, want errStop", err)
	}
	if n := count(t, pool, "changes") + count(t, pool, "outbox"); n != 0 {
		t.Errorf("a rolled-back change left %d rows", n)
	}
	// An outbox that refuses the event must take the change with it.
	adminSQL := `CREATE FUNCTION attention.refuse_event() RETURNS trigger LANGUAGE plpgsql AS $$
		BEGIN RAISE EXCEPTION 'outbox down'; END; $$;
		CREATE TRIGGER outbox_down BEFORE INSERT ON attention.outbox FOR EACH ROW EXECUTE FUNCTION attention.refuse_event();`
	if _, err := pool.Exec(ctx, adminSQL); err != nil {
		t.Fatalf("breaking the outbox: %v", err)
	}
	if _, err := s.Append(ctx, &step); err == nil {
		t.Fatal("Append succeeded with the outbox refusing its event")
	}
	if n := count(t, pool, "items") + count(t, pool, "changes"); n != 0 {
		t.Errorf("a change whose event failed left %d rows", n)
	}
}

func TestStoreIsAppendOnly(t *testing.T) {
	t.Parallel()
	pool := newPool(t)
	s := postgres.New(pool)
	ctx := t.Context()
	step := raise(attention.KindImportProblems, "b-1")
	if _, err := s.Append(ctx, &step); err != nil {
		t.Fatalf("Append: %v", err)
	}
	if err := s.MarkPublished(ctx, 1, at()); err != nil {
		t.Fatalf("MarkPublished: %v", err)
	}
	for _, sql := range []string{
		"UPDATE attention.items SET subject = 'x'",
		"DELETE FROM attention.items",
		"UPDATE attention.changes SET state = 'resolved'",
		"DELETE FROM attention.changes",
		"UPDATE attention.outbox SET topic = 'x'",
		"DELETE FROM attention.outbox",
		"DELETE FROM attention.published",
		"TRUNCATE attention.changes CASCADE",
	} {
		_, err := pool.Exec(ctx, sql)
		var pgErr *pgconn.PgError
		if !errors.As(err, &pgErr) || pgErr.Code != "OFA01" {
			t.Errorf("%s: error %v, want the append-only refusal", sql, err)
		}
	}
	skip := step
	skip.ItemID, skip.Change.Seq = 1, 3
	if _, err := s.Append(ctx, &skip); err == nil {
		t.Error("appending change 3 after change 1 succeeded")
	}
}

func TestSweepOverPostgres(t *testing.T) {
	t.Parallel()
	pool := newPool(t)
	s := postgres.New(pool)
	ctx := t.Context()
	conds := []attention.Condition{{Kind: attention.KindTickerUnconfirmed, Subject: "xtb/AAA.US", Detail: "2 rows wait"}}
	check := app.Check{Kind: attention.KindTickerUnconfirmed, Find: func(context.Context) ([]attention.Condition, error) { return conds, nil }}
	items := app.New(s, nil, at, check)
	for range 2 {
		if err := items.Sweep(ctx); err != nil {
			t.Fatalf("Sweep: %v", err)
		}
	}
	conds = nil
	if err := items.Sweep(ctx); err != nil {
		t.Fatalf("Sweep: %v", err)
	}
	list, err := items.List(ctx)
	if err != nil || len(list) != 1 {
		t.Fatalf("List = %v, %v; want one item", list, err)
	}
	if got := []attention.State{list[0].Changes[0].State, list[0].State()}; !cmp.Equal(got, []attention.State{attention.StateOpen, attention.StateResolved}) {
		t.Errorf("history states = %v, want raised then resolved", got)
	}
	if n := count(t, pool, "outbox"); n != 2 {
		t.Errorf("outbox holds %d events, want one per change", n)
	}
}
