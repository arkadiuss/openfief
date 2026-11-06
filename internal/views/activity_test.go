package views_test

import (
	"errors"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"github.com/arkadiuss/openfief/internal/ingest"
	"github.com/arkadiuss/openfief/internal/ledger"
	"github.com/arkadiuss/openfief/internal/views"
)

// activityAt is minute m of a synthetic day.
func activityAt(m int) time.Time {
	return time.Date(2026, time.November, 4, 6, 0, 0, 0, time.UTC).Add(time.Duration(m) * time.Minute)
}

// batchAt is a completed sync batch id finished at minute m.
func batchAt(id string, m int) views.Activity {
	return views.BatchActivity(&ingest.BatchRecord{ID: ingest.BatchID(id), Source: "sync", Finished: activityAt(m)})
}

// failureAt is failed attempt id ended at minute m.
func failureAt(id int64, m int) views.Activity {
	return views.FailureActivity(&ingest.SyncAttempt{ID: id, Ended: activityAt(m), Class: ingest.SyncErrorTimeout, Failures: 1})
}

// changeAt is ledger version id recorded at minute m, of version v; hidden
// when the caller may not view its account.
func changeAt(id int64, m, v int, hidden bool) views.Activity {
	a := views.ChangeActivity(&ledger.VersionedTransaction{Version: v}, id, activityAt(m))
	a.Hidden = hidden
	return a
}

// names lists the activities as "<kind>@<minute>".
func names(acts []views.Activity) []string {
	out := make([]string, 0, len(acts))
	for i := range acts {
		a := &acts[i]
		kind := map[views.ActivityKind]string{
			views.ActivityImport: "import", views.ActivitySync: "sync", views.ActivitySyncFailed: "failed",
			views.ActivityCoupon: "coupon", views.ActivityCorrection: "correction",
		}[a.Kind]
		out = append(out, kind+"@"+a.At.Format("1504"))
	}
	return out
}

func TestMergeActivity(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		fetched  [3][]views.Activity
		size     int
		want     []string
		wantNext bool
	}{
		{
			name: "newest first across sources, sources in order at the same time",
			fetched: [3][]views.Activity{
				{batchAt("b-2", 30), batchAt("b-1", 10)},
				{failureAt(7, 30), failureAt(6, 20)},
				{changeAt(3, 25, 1, false), changeAt(2, 5, 2, false)},
			},
			size: 10,
			want: []string{"sync@0630", "failed@0630", "coupon@0625", "failed@0620", "sync@0610", "correction@0605"},
		},
		{
			name: "a page short of everything has a next page",
			fetched: [3][]views.Activity{
				{batchAt("b-2", 30), batchAt("b-1", 10)},
				nil,
				{changeAt(3, 25, 1, false)},
			},
			size: 2, want: []string{"sync@0630", "coupon@0625"}, wantNext: true,
		},
		{
			name: "hidden changes are paged past",
			fetched: [3][]views.Activity{
				{batchAt("b-1", 10)},
				nil,
				{changeAt(3, 25, 1, true), changeAt(2, 20, 2, false)},
			},
			size: 10, want: []string{"correction@0620", "sync@0610"},
		},
		{
			name: "the merge stops where a full source runs out",
			fetched: [3][]views.Activity{
				{batchAt("b-1", 10)},
				nil,
				{changeAt(5, 40, 1, true), changeAt(4, 35, 1, true), changeAt(3, 30, 1, true)},
			},
			size: 2, want: []string{}, wantNext: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			page, err := views.MergeActivity(&views.ActivityCursor{}, &tt.fetched, tt.size+1, tt.size)
			if err != nil {
				t.Fatalf("MergeActivity: %v", err)
			}
			if diff := cmp.Diff(tt.want, names(page.Activities)); diff != "" {
				t.Errorf("activities (-want +got):\n%s", diff)
			}
			if (page.Next != nil) != tt.wantNext {
				t.Errorf("next = %v, want one: %v", page.Next, tt.wantNext)
			}
		})
	}
}

// TestMergeActivityPages pages through three sources two at a time and
// checks that every activity comes once, newest first, and that each page's
// cursor names the last activity it took from each source.
func TestMergeActivityPages(t *testing.T) {
	t.Parallel()
	all := [3][]views.Activity{
		{batchAt("b-3", 50), batchAt("b-2", 30), batchAt("b-1", 10)},
		{failureAt(9, 40), failureAt(8, 20)},
		{changeAt(4, 45, 2, false), changeAt(3, 30, 1, true), changeAt(2, 5, 1, false)},
	}
	var after views.ActivityCursor
	var got []string
	for range 10 {
		var fetched [3][]views.Activity
		for s := range all {
			fetched[s] = after3(all[s], after[s], 3)
		}
		page, err := views.MergeActivity(&after, &fetched, 3, 2)
		if err != nil {
			t.Fatalf("MergeActivity: %v", err)
		}
		got = append(got, names(page.Activities)...)
		if page.Next == nil {
			break
		}
		after = *page.Next
	}
	want := []string{"sync@0650", "correction@0645", "failed@0640", "sync@0630", "failed@0620", "sync@0610", "coupon@0605"}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("pages (-want +got):\n%s", diff)
	}
}

// after3 returns at most n of src, newest first, after bound b, as a store
// does.
func after3(src []views.Activity, b views.ActivityBound, n int) []views.Activity {
	var out []views.Activity
	for i := range src {
		a := src[i]
		if !b.IsZero() && !a.At.Before(b.At) {
			continue
		}
		if len(out) < n {
			out = append(out, a)
		}
	}
	return out
}

func TestActivityCursor(t *testing.T) {
	t.Parallel()
	c := views.ActivityCursor{
		views.SourceBatches:  {At: activityAt(10), Key: "b-1"},
		views.SourceChanges:  {At: activityAt(5), ID: 42},
		views.SourceFailures: {},
	}
	got, err := views.ParseActivityCursor(c.Encode())
	if err != nil {
		t.Fatalf("ParseActivityCursor: %v", err)
	}
	if diff := cmp.Diff(c, got); diff != "" {
		t.Errorf("round trip (-want +got):\n%s", diff)
	}
	b := views.ActivityBound{At: activityAt(3), ID: 7}
	gotBound, err := views.ParseActivityBound(b.Encode())
	if err != nil || gotBound != b {
		t.Errorf("bound round trip = %+v, %v", gotBound, err)
	}
	for _, bad := range []string{"%%%", "bm90IGpzb24", "W10", c.Encode()[:4]} {
		if _, err := views.ParseActivityCursor(bad); !errors.Is(err, views.ErrInvalidCursor) {
			t.Errorf("ParseActivityCursor(%q) = %v, want ErrInvalidCursor", bad, err)
		}
	}
	if _, err := views.ParseActivityBound(c.Encode()); !errors.Is(err, views.ErrInvalidCursor) {
		t.Errorf("ParseActivityBound of a cursor without a failures bound = %v", err)
	}
	if _, err := views.MergeActivity(&c, &[3][]views.Activity{}, 1, 0); !errors.Is(err, views.ErrInvalidPageSize) {
		t.Errorf("MergeActivity of size 0 = %v", err)
	}
}

func TestActivityKinds(t *testing.T) {
	t.Parallel()
	upload := views.BatchActivity(&ingest.BatchRecord{ID: "u-1", Source: "upload"})
	coupon := views.ChangeActivity(&ledger.VersionedTransaction{Version: 1}, 1, activityAt(0))
	correction := views.ChangeActivity(&ledger.VersionedTransaction{Version: 3}, 2, activityAt(0))
	got := []views.ActivityKind{upload.Kind, batchAt("b", 0).Kind, failureAt(1, 0).Kind, coupon.Kind, correction.Kind}
	want := []views.ActivityKind{views.ActivityImport, views.ActivitySync, views.ActivitySyncFailed, views.ActivityCoupon, views.ActivityCorrection}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("kinds (-want +got):\n%s", diff)
	}
}
