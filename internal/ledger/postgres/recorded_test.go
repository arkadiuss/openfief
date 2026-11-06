package postgres_test

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"github.com/arkadiuss/openfief/internal/ledger"
	"github.com/arkadiuss/openfief/internal/ledger/app"
	"github.com/arkadiuss/openfief/internal/ledger/postgres"
)

// recordedFixture stores the log fixture and a coupon after it, and returns
// the store, the fixture and the coupon.
func recordedFixture(t *testing.T) (*postgres.Store, logFixture, ledger.Transaction) {
	t.Helper()
	store, _ := seededStore(t)
	f := newLogFixture(t)
	coupon := must(t)(ledger.NewInterest(header(t, acc1, "11-coupon", "2026-03-07"), instrument(t, isinB, "TSTB"),
		money(t, "12.50", "PLN"), money(t, "2.38", "PLN")))
	appendAll(t, store, append(f.all, versioned(&coupon, 1)))
	return store, f, coupon
}

// recordedChanges returns every recorded change of store, failing the test
// when it can't.
func recordedChanges(t *testing.T, store *postgres.Store) []app.RecordedVersion {
	t.Helper()
	got, err := store.RecordedChanges(t.Context(), time.Time{}, 0, 10)
	if err != nil {
		t.Fatalf("RecordedChanges: %v", err)
	}
	return got
}

// TestRecordedChanges checks that the activity's versions are the
// corrections and the coupons, newest first with their postings, and that
// the rest of the log is left out.
func TestRecordedChanges(t *testing.T) {
	t.Parallel()
	store, f, coupon := recordedFixture(t)
	got := recordedChanges(t, store)
	versions := make([]ledger.VersionedTransaction, 0, len(got))
	marks := make([][2]bool, 0, len(got))
	for i := range got {
		versions = append(versions, got[i].VersionedTransaction)
		marks = append(marks, [2]bool{got[i].IsCoupon(), got[i].IsCorrection()})
	}
	checkResult(t, "RecordedChanges", []ledger.VersionedTransaction{versioned(&coupon, 1), f.all[10], f.all[8]}, versions, nil)
	if diff := cmp.Diff([][2]bool{{true, false}, {false, true}, {false, true}}, marks); diff != "" {
		t.Errorf("coupon and correction marks (-want +got):\n%s", diff)
	}
}

// TestRecordedChangesPages checks that each change has its record time and
// row ID, and that a page continues after the last change of the one
// before.
func TestRecordedChangesPages(t *testing.T) {
	t.Parallel()
	store, _, _ := recordedFixture(t)
	got := recordedChanges(t, store)
	if len(got) != 3 || got[1].Recorded.IsZero() || got[1].ID == 0 {
		t.Fatalf("RecordedChanges = %+v, want three with record times and IDs", got)
	}
	rest, err := store.RecordedChanges(t.Context(), got[1].Recorded, got[1].ID, 10)
	if err != nil || len(rest) != 1 || rest[0].ID != got[2].ID {
		t.Errorf("after the second = %+v, %v; want the third only", rest, err)
	}
	one, err := store.RecordedChanges(t.Context(), time.Time{}, 0, 1)
	if err != nil || len(one) != 1 || one[0].ID != got[0].ID {
		t.Errorf("a page of one = %+v, %v", one, err)
	}
}
