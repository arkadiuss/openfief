package main

import (
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"

	viewsv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/views/v1"
)

// stickAuditJobs inserts n audit-entry jobs that have failed failures attempts
// and wait for their next one a day from now, as a River queue holds them while
// the log is down, and returns their IDs.
func stickAuditJobs(t *testing.T, pool *pgxpool.Pool, n, failures int) []int64 {
	t.Helper()
	ids := make([]int64, 0, n)
	for range n {
		var id int64
		err := pool.QueryRow(t.Context(), `
			INSERT INTO river.river_job (state, kind, args, queue, max_attempts, attempt, scheduled_at, errors)
			VALUES ('retryable', $1, '{}', 'audit', 32767, $2::int, now() + interval '1 day',
			        ARRAY(SELECT '{"at":"2026-03-02T10:00:00Z","attempt":1,"error":"the log is down","trace":""}'::jsonb FROM generate_series(1, $2::int)))
			RETURNING id`, auditEntryArgs{}.Kind(), failures).Scan(&id)
		if err != nil {
			t.Fatalf("holding an audit job in the queue: %v", err)
		}
		ids = append(ids, id)
	}
	return ids
}

// finish marks job id completed, as River does when its worker succeeds.
func finish(t *testing.T, pool *pgxpool.Pool, id int64) {
	t.Helper()
	if _, err := pool.Exec(t.Context(), `UPDATE river.river_job SET state = 'completed', finalized_at = now() WHERE id = $1`, id); err != nil {
		t.Fatalf("finishing job %d: %v", id, err)
	}
}

// behindWatch runs the core's checks and reads the attention items back through
// the read API, as a view plugin would.
type behindWatch struct {
	t    *testing.T
	c    *core
	list func() *viewsv1.ListAttentionItemsResponse
}

// behindItems are the audit-log-behind items the read API lists, and how many
// items it counts open.
type behindItems struct {
	behind []*viewsv1.AttentionItem
	open   int32
}

// items runs a tick of the checks and returns what the read API lists then.
func (w *behindWatch) items() behindItems {
	w.t.Helper()
	w.c.attention.tick(w.t.Context(), quiet())
	got := w.list()
	out := behindItems{open: got.GetOpen()}
	for _, it := range got.GetItems() {
		if it.GetKind() == "audit-log-behind" {
			out.behind = append(out.behind, it)
		}
	}
	return out
}

// wantNone fails the test when the read API lists an audit-log-behind item or
// counts an item open, and says why none was expected.
func (w *behindWatch) wantNone(why string) {
	w.t.Helper()
	if got := w.items(); len(got.behind) != 0 || got.open != 0 {
		w.t.Fatalf("items = %v, open %d %s; want none", got.behind, got.open, why)
	}
}

// newBehindWatch starts a core over an empty database and returns what watches
// its attention items, with its pool.
func newBehindWatch(t *testing.T) (*behindWatch, *pgxpool.Pool) {
	t.Helper()
	dsn := emptyDatabase(t)
	c, srv, pool := startCore(t, dsn)
	secret, _, _ := strings.Cut(tokenCmd(t, dsn, "issue", "--name", "behind-test", "--ttl", "1h"), "\n")
	w := &behindWatch{t: t, c: c}
	w.list = func() *viewsv1.ListAttentionItemsResponse {
		var got viewsv1.ListAttentionItemsResponse
		getJSON(t, srv, "/v1/attention-items", secret, &got)
		return &got
	}
	return w, pool
}

// TestAuditLogBehindItemIsRaisedAtFiveFailures checks over a real database
// (D-211, D-196) that the core raises the attention item audit-log-behind, in the
// group of data coming in, once an audit entry's job has failed five attempts,
// naming how many changes wait, and does not for jobs that failed fewer.
func TestAuditLogBehindItemIsRaisedAtFiveFailures(t *testing.T) {
	t.Parallel()
	w, pool := newBehindWatch(t)
	w.wantNone("with nothing waiting")
	stickAuditJobs(t, pool, 3, 4)
	w.wantNone("with jobs that failed four times")
	stickAuditJobs(t, pool, 2, 5)
	got := w.items()
	if len(got.behind) != 1 || got.open != 1 {
		t.Fatalf("items = %v, open %d with two jobs that failed five times; want the one item", got.behind, got.open)
	}
	it := got.behind[0]
	wantDetail := "2 changes are waiting to enter the audit log; OpenFief keeps retrying. Check the database's health."
	if it.GetSubject() != "audit-log" || it.GetDetail() != wantDetail || it.GetState() != viewsv1.AttentionItemState_ATTENTION_ITEM_STATE_OPEN ||
		it.GetGroup() != viewsv1.AttentionItemGroup_ATTENTION_ITEM_GROUP_DATA_COMING_IN {
		t.Errorf("item = %v, want the open audit-log item of data coming in saying %q", it, wantDetail)
	}
}

// TestAuditLogBehindItemClearsWhenTheStuckEntriesGetThrough checks that the
// item's count follows the jobs that are still stuck, and that the item clears
// when none is, with the log caught up.
func TestAuditLogBehindItemClearsWhenTheStuckEntriesGetThrough(t *testing.T) {
	t.Parallel()
	w, pool := newBehindWatch(t)
	stuck := stickAuditJobs(t, pool, 2, 5)
	if got := w.items(); len(got.behind) != 1 {
		t.Fatalf("items = %v with two jobs stuck, want the item", got.behind)
	}
	finish(t, pool, stuck[0])
	got := w.items()
	if len(got.behind) != 1 || !strings.HasPrefix(got.behind[0].GetDetail(), "1 change is waiting to enter the audit log") {
		t.Errorf("items = %v after one job got through, want the item saying one change waits", got.behind)
	}
	finish(t, pool, stuck[1])
	got = w.items()
	if len(got.behind) != 1 || got.open != 0 || got.behind[0].GetState() != viewsv1.AttentionItemState_ATTENTION_ITEM_STATE_RESOLVED {
		t.Errorf("items = %v, open %d once the stuck jobs got through; want the item cleared", got.behind, got.open)
	}
}
