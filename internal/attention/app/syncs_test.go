package app_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/arkadiuss/openfief/internal/attention"
	"github.com/arkadiuss/openfief/internal/attention/app"
	"github.com/arkadiuss/openfief/internal/ingest"
)

// errSyncs stands for the ingest store failing.
var errSyncs = errors.New("sync statuses unavailable")

// fakeSyncs answers fixed statuses, or err.
type fakeSyncs struct {
	statuses []ingest.SyncStatus
	err      error
}

func (f fakeSyncs) SyncStatuses(context.Context) ([]ingest.SyncStatus, error) {
	return f.statuses, f.err
}

func TestFailingSyncs(t *testing.T) {
	t.Parallel()
	since := time.Date(2026, time.November, 3, 6, 0, 30, 0, time.UTC)
	failing := ingest.SyncStatus{
		Instance: "xtb-main", Failures: 3, FailingSince: since, LastAttempt: since.Add(time.Hour),
		LastClass: ingest.SyncErrorTimeout,
	}
	tests := []struct {
		name     string
		statuses []ingest.SyncStatus
		err      error
		want     []attention.Condition
		wantErr  error
	}{
		{name: "none"},
		{
			name: "one keeps failing, one failed briefly, one synced",
			statuses: []ingest.SyncStatus{
				{Instance: "eodhd", LastAttempt: since, LastSuccess: since},
				{Instance: "nbp-fx", Failures: 1, FailingSince: since, LastAttempt: since},
				failing,
			},
			want: []attention.Condition{{Kind: attention.KindSyncFailing, Subject: "xtb-main", Detail: "failing since 2026-11-03T06:00Z"}},
		},
		{name: "the store fails", err: errSyncs, wantErr: errSyncs},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := app.FailingSyncs(fakeSyncs{statuses: tt.statuses, err: tt.err}).Find(t.Context())
			if diff := cmp.Diff(tt.wantErr, err, cmpopts.EquateErrors()); diff != "" {
				t.Fatalf("error (-want +got):\n%s", diff)
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("conditions (-want +got):\n%s", diff)
			}
		})
	}
}

// TestFailingSyncItem sweeps the check through a run of failures: it
// raises one item, keeps it acknowledged while the failures go on, clears
// it after a success and raises it again for a new run.
func TestFailingSyncItem(t *testing.T) {
	t.Parallel()
	since := time.Date(2026, time.November, 3, 6, 0, 0, 0, time.UTC)
	run := func(start time.Time, failures int) ingest.SyncStatus {
		return ingest.SyncStatus{
			Instance: "xtb-main", Failures: failures, FailingSince: start, LastAttempt: start.Add(time.Duration(failures) * 10 * time.Minute),
			LastClass: ingest.SyncErrorUnavailable,
		}
	}
	steps := []ingest.SyncStatus{
		run(since, 2), // not yet
		run(since, 3), // keeps failing: raised
		run(since, 5), // still the same run
		{Instance: "xtb-main", LastSuccess: since.Add(2 * time.Hour), LastAttempt: since.Add(2 * time.Hour)}, // cleared
		run(since.Add(3*time.Hour), 4), // a new run: reopened
	}
	item := attention.Item{ID: 1, Kind: attention.KindSyncFailing, Subject: "xtb-main"}
	states := make([]attention.State, 0, len(steps))
	for i, st := range steps {
		at := since.Add(time.Duration(i) * time.Hour)
		sweepSyncs(t, &item, &st, at)
		if i == 1 {
			ch, _, ackErr := attention.Acknowledge(&item, at)
			if ackErr != nil {
				t.Fatalf("Acknowledge: %v", ackErr)
			}
			item.Changes = append(item.Changes, ch)
		}
		states = append(states, item.State())
	}
	want := []attention.State{0, attention.StateAcknowledged, attention.StateAcknowledged, attention.StateResolved, attention.StateOpen}
	if diff := cmp.Diff(want, states); diff != "" {
		t.Errorf("states after each step (-want +got):\n%s", diff)
	}
}

// sweepSyncs runs the check over st and appends to item what reconciling
// it at the time at changes.
func sweepSyncs(t *testing.T, item *attention.Item, st *ingest.SyncStatus, at time.Time) {
	t.Helper()
	conds, err := app.FailingSyncs(fakeSyncs{statuses: []ingest.SyncStatus{*st}}).Find(t.Context())
	if err != nil {
		t.Fatalf("Find: %v", err)
	}
	var stored []attention.Item
	if len(item.Changes) > 0 {
		stored = []attention.Item{*item}
	}
	steps, err := attention.Reconcile(attention.KindSyncFailing, stored, conds, at)
	if err != nil {
		t.Fatalf("Reconcile: %v", err)
	}
	for _, s := range steps {
		item.Changes = append(item.Changes, s.Change)
	}
}
