package postgres_test

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/arkadiuss/openfief/internal/ingest"
	"github.com/arkadiuss/openfief/internal/ingest/app"
	"github.com/arkadiuss/openfief/internal/ingest/postgres"
	"github.com/arkadiuss/openfief/internal/ledger"
)

// failed returns a failed attempt of xtb-main's batch id, the failures-th in
// a row, that started a minute before it ended, at the time at, and scheduled the next
// at next; it was unreachable.
func failed(id ingest.BatchID, failures int, next, at time.Time) *ingest.SyncAttempt {
	return &ingest.SyncAttempt{
		Instance: "xtb-main", Batch: id, Mode: ingest.ModeOpening, Started: at.Add(-time.Minute), Ended: at,
		Class: ingest.SyncErrorUnavailable, Failures: failures, NextSync: next,
	}
}

// succeeded returns the successful attempt of instance that completed b at
// at, a minute after it started, and scheduled the next at next.
func succeeded(instance string, b *ingest.Batch, next, at time.Time) *ingest.SyncAttempt {
	return &ingest.SyncAttempt{
		Instance: instance, Batch: b.ID, Mode: b.Mode, Started: at.Add(-time.Minute), Ended: at, NextSync: next,
	}
}

// syncHistory runs xtb-main through two failed attempts of b-1, a success,
// and three failures of b-2 that span 20 minutes, and nbp through one
// success; it returns the store and its pool.
func syncHistory(t *testing.T) (*postgres.Store, *pgxpool.Pool) {
	t.Helper()
	pool := newPool(t)
	s := postgres.New(pool)
	ctx := t.Context()
	at := func(m int) time.Time { return t0().Add(time.Duration(m) * time.Minute) }
	claim(t, s, "xtb-main", at(0))
	if err := s.StartBatch(ctx, "xtb-main", app.OpenBatch{ID: "b-1", Mode: ingest.ModeOpening}, at(0)); err != nil {
		t.Fatalf("StartBatch b-1: %v", err)
	}
	mustFail(t, s, failed("b-1", 1, at(6), at(1)))
	claim(t, s, "xtb-main", at(6))
	mustFail(t, s, failed("b-1", 2, at(17), at(7)))
	claim(t, s, "xtb-main", at(17))
	b1 := &ingest.Batch{ID: "b-1", Institution: "xtb", Source: app.SourceSync, Mode: ingest.ModeOpening, Transactions: 3, Imported: ingest.ImportSummary{New: 3}, Accounts: []ledger.AccountID{"acc-1", "acc-2"}}
	if err := s.CompleteSync(ctx, b1, "w-1", succeeded("xtb-main", b1, at(100), at(18)), time.Time{}); err != nil {
		t.Fatalf("CompleteSync b-1: %v", err)
	}
	for i, m := range []int{101, 107, 120} {
		claim(t, s, "xtb-main", at(m-1))
		if i == 0 {
			if err := s.StartBatch(ctx, "xtb-main", app.OpenBatch{ID: "b-2", Mode: ingest.ModeTransactions}, at(100)); err != nil {
				t.Fatalf("StartBatch b-2: %v", err)
			}
		}
		a := failed("b-2", i+1, at(m+5), at(m))
		a.Mode = ingest.ModeTransactions
		if i == 2 {
			a.Class = ingest.SyncErrorTimeout
		}
		mustFail(t, s, a)
	}
	claim(t, s, "nbp", at(30))
	if err := s.StartBatch(ctx, "nbp", app.OpenBatch{ID: "n-1", Mode: ingest.ModeTransactions}, at(30)); err != nil {
		t.Fatalf("StartBatch n-1: %v", err)
	}
	n1 := &ingest.Batch{ID: "n-1", Institution: "nbp", Source: app.SourceSync, Mode: ingest.ModeTransactions}
	if err := s.CompleteSync(ctx, n1, "w-n", succeeded("nbp", n1, at(400), at(31)), time.Time{}); err != nil {
		t.Fatalf("CompleteSync n-1: %v", err)
	}
	return s, pool
}

// mustFail records a, failing the test when it can't.
func mustFail(t *testing.T, s *postgres.Store, a *ingest.SyncAttempt) {
	t.Helper()
	if err := s.FailSync(t.Context(), a); err != nil {
		t.Fatalf("FailSync %s #%d: %v", a.Batch, a.Failures, err)
	}
}

// TestSyncAttempts checks that every attempt is kept with its batch, times,
// error class, failures and next sync, newest first, for one instance or
// all, failed ones only on request, and paged by when it ended.
func TestSyncAttempts(t *testing.T) {
	t.Parallel()
	s, _ := syncHistory(t)
	at := func(m int) time.Time { return t0().Add(time.Duration(m) * time.Minute) }
	all, err := s.SyncAttempts(t.Context(), &ingest.AttemptQuery{Instance: "xtb-main", Limit: 10})
	if err != nil {
		t.Fatalf("SyncAttempts: %v", err)
	}
	type row struct {
		Batch    ingest.BatchID
		Ended    time.Time
		Class    ingest.SyncErrorClass
		Failures int
		Next     time.Time
	}
	got := make([]row, 0, len(all))
	for _, a := range all {
		got = append(got, row{a.Batch, a.Ended, a.Class, a.Failures, a.NextSync})
	}
	want := []row{
		{"b-2", at(120), ingest.SyncErrorTimeout, 3, at(125)},
		{"b-2", at(107), ingest.SyncErrorUnavailable, 2, at(112)},
		{"b-2", at(101), ingest.SyncErrorUnavailable, 1, at(106)},
		{"b-1", at(18), 0, 0, at(100)},
		{"b-1", at(7), ingest.SyncErrorUnavailable, 2, at(17)},
		{"b-1", at(1), ingest.SyncErrorUnavailable, 1, at(6)},
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Fatalf("xtb-main's attempts (-want +got):\n%s", diff)
	}
	if first := all[len(all)-1]; first.Mode != ingest.ModeOpening || !first.Started.Equal(at(0)) || first.Instance != "xtb-main" {
		t.Errorf("first attempt = %+v", first)
	}
}

// TestSyncAttemptsPages checks that the failed attempts page after the last
// one of a previous page, and that every instance's attempts are listed
// without an instance.
func TestSyncAttemptsPages(t *testing.T) {
	t.Parallel()
	s, _ := syncHistory(t)
	all, err := s.SyncAttempts(t.Context(), &ingest.AttemptQuery{Instance: "xtb-main", Limit: 10})
	if err != nil || len(all) != 6 {
		t.Fatalf("SyncAttempts = %d attempts, %v; want 6", len(all), err)
	}
	page, err := s.SyncAttempts(t.Context(), &ingest.AttemptQuery{FailedOnly: true, Before: all[1].Ended, BeforeID: all[1].ID, Limit: 2})
	if err != nil {
		t.Fatalf("SyncAttempts page: %v", err)
	}
	if len(page) != 2 || page[0].ID != all[2].ID || page[1].ID != all[4].ID {
		t.Errorf("failed attempts after the second = %+v, want b-2's first and b-1's second", page)
	}
	every, err := s.SyncAttempts(t.Context(), &ingest.AttemptQuery{Limit: 100})
	if err != nil {
		t.Fatalf("SyncAttempts of every instance: %v", err)
	}
	if len(every) != 7 {
		t.Errorf("every instance's attempts = %d, want 7", len(every))
	}
}

// TestSyncStatuses checks each instance's status: its latest attempt and
// success, the failures since that success with when they began and the
// latest class, and its next sync; and that the attempts are append-only.
func TestSyncStatuses(t *testing.T) {
	t.Parallel()
	s, pool := syncHistory(t)
	at := func(m int) time.Time { return t0().Add(time.Duration(m) * time.Minute) }
	got, err := s.SyncStatuses(t.Context())
	if err != nil {
		t.Fatalf("SyncStatuses: %v", err)
	}
	want := []ingest.SyncStatus{
		{Instance: "nbp", LastAttempt: at(31), LastSuccess: at(31), NextSync: at(400)},
		{
			Instance: "xtb-main", LastAttempt: at(120), LastSuccess: at(18), Failures: 3, FailingSince: at(100),
			LastClass: ingest.SyncErrorTimeout, NextSync: at(125),
		},
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("statuses (-want +got):\n%s", diff)
	}
	if !got[1].KeepsFailing() || got[0].KeepsFailing() {
		t.Errorf("KeepsFailing = %v for nbp, %v for xtb-main; want false, true", got[0].KeepsFailing(), got[1].KeepsFailing())
	}
	if _, err := pool.Exec(t.Context(), "DELETE FROM ingest.sync_attempts"); err == nil {
		t.Error("deleting sync attempts succeeded, want the append-only trigger to refuse it")
	}
}

// TestCompletedBatches checks that completed batches of every source come
// newest first with their counts, open ones not at all, and page by when
// they finished.
func TestCompletedBatches(t *testing.T) {
	t.Parallel()
	s, _ := syncHistory(t)
	got, err := s.CompletedBatches(t.Context(), &ingest.BatchQuery{Limit: 10})
	if err != nil {
		t.Fatalf("CompletedBatches: %v", err)
	}
	at := func(m int) time.Time { return t0().Add(time.Duration(m) * time.Minute) }
	want := []ingest.BatchRecord{
		{ID: "n-1", Instance: "nbp", Source: "sync", Mode: ingest.ModeTransactions, Finished: at(31), Institution: "nbp"},
		{ID: "b-1", Instance: "xtb-main", Source: "sync", Mode: ingest.ModeOpening, Finished: at(18), Institution: "xtb", Transactions: 3, New: 3, Accounts: []ledger.AccountID{"acc-1", "acc-2"}},
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("completed batches (-want +got):\n%s", diff)
	}
	rest, err := s.CompletedBatches(t.Context(), &ingest.BatchQuery{Before: got[0].Finished, BeforeID: got[0].ID, Limit: 10})
	if err != nil {
		t.Fatalf("CompletedBatches after n-1: %v", err)
	}
	if len(rest) != 1 || rest[0].ID != "b-1" {
		t.Errorf("batches after n-1 = %+v, want b-1 only", rest)
	}
}

// TestRecordedBatchKeepsAccounts checks that a recorded upload keeps the
// accounts it booked into (D-140), and lists them with the completed
// batches.
func TestRecordedBatchKeepsAccounts(t *testing.T) {
	t.Parallel()
	s, _ := syncHistory(t)
	counts := []ingest.AccountCounts{{Account: "acc-2", Transactions: 4, New: 3, NewVersions: 1, FailedAssertions: 1}}
	upload := &ingest.Batch{
		ID: "u-1", Institution: "xtb", Source: app.SourceUpload, Mode: ingest.ModeTransactions, Accounts: []ledger.AccountID{"acc-2"},
		PerAccount: counts,
	}
	if err := s.RecordBatch(t.Context(), "xtb-main", upload); err != nil {
		t.Fatalf("RecordBatch u-1: %v", err)
	}
	newest, err := s.CompletedBatches(t.Context(), &ingest.BatchQuery{Limit: 1})
	if err != nil {
		t.Fatalf("CompletedBatches after the upload: %v", err)
	}
	if len(newest) != 1 || newest[0].ID != "u-1" || !cmp.Equal(newest[0].Accounts, []ledger.AccountID{"acc-2"}) {
		t.Errorf("newest batch = %+v, want the upload u-1 booked into acc-2", newest)
	}
	if len(newest) == 1 && !cmp.Equal(newest[0].PerAccount, counts) {
		t.Errorf("the upload's counts per account = %+v, want %+v (D-199)", newest[0].PerAccount, counts)
	}
}

// TestCompletedBatchesByID checks that the batch history selects completed
// batches by their IDs, for a sync's attempts (D-199).
func TestCompletedBatchesByID(t *testing.T) {
	t.Parallel()
	s, _ := syncHistory(t)
	byID, err := s.CompletedBatches(t.Context(), &ingest.BatchQuery{IDs: []ingest.BatchID{"b-1"}, Limit: 5})
	if err != nil {
		t.Fatalf("CompletedBatches of b-1: %v", err)
	}
	if len(byID) != 1 || byID[0].ID != "b-1" {
		t.Errorf("batches with ID b-1 = %+v, want b-1 alone", byID)
	}
}
