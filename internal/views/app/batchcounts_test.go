package app_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"

	identityapp "github.com/arkadiuss/openfief/internal/identity/app"
	"github.com/arkadiuss/openfief/internal/ingest"
	"github.com/arkadiuss/openfief/internal/ledger"
	"github.com/arkadiuss/openfief/internal/views"
	"github.com/arkadiuss/openfief/internal/views/app"
)

// countedHistory is a sync of xtb-main and an upload of files, each booked
// into acc-1 and acc-2 with their counts per account (D-199).
func countedHistory() *fakeHistory {
	a1 := ingest.AccountCounts{Account: "acc-1", Transactions: 2, New: 2}
	a2 := ingest.AccountCounts{Account: "acc-2", Transactions: 3, New: 1, NewVersions: 1, FailedAssertions: 1}
	booked := []ledger.AccountID{"acc-1", "acc-2"}
	return &fakeHistory{
		batches: []ingest.BatchRecord{
			{
				ID: "u-1", Instance: "files", Source: "upload", Finished: at(25), Transactions: 5, New: 3, NewVersions: 1, FailedAssertions: 1,
				Problems: 2, Accounts: booked, PerAccount: []ingest.AccountCounts{a1, a2},
			},
			{
				ID: "b-1", Instance: "xtb-main", Source: "sync", Finished: at(10), Transactions: 5, New: 3, NewVersions: 1, FailedAssertions: 1,
				Accounts: booked, PerAccount: []ingest.AccountCounts{a1, a2},
			},
		},
		attempts: []ingest.SyncAttempt{
			{ID: 2, Instance: "xtb-main", Batch: "b-2", Ended: at(20), Class: ingest.SyncErrorTimeout, Failures: 1},
			{ID: 1, Instance: "xtb-main", Batch: "b-1", Ended: at(10)},
		},
		statuses: []ingest.SyncStatus{{Instance: "xtb-main", Failures: 1, LastAttempt: at(20)}},
	}
}

// uploadSeen returns the upload u-1 as the caller in ctx sees it in v's
// recent activity, or nil.
func uploadSeen(ctx context.Context, t *testing.T, v *app.Views) *ingest.BatchRecord {
	t.Helper()
	page, err := v.Activity(ctx, nil, 10)
	if err != nil {
		t.Fatalf("Activity: %v", err)
	}
	for i := range page.Activities {
		if b := page.Activities[i].Batch; b != nil && b.ID == "u-1" {
			return b
		}
	}
	return nil
}

// TestBatchCountsPerViewableAccount checks D-199's counts: a caller who may
// view every account a batch booked into sees its totals and each
// account's counts; a viewer of one of them sees that account alone, its
// counts as the batch's, and no left-out rows; a sync attempt that went
// through names the accounts its batch booked into that the caller may
// view, with their counts, and a failed one none.
func TestBatchCountsPerViewableAccount(t *testing.T) {
	t.Parallel()
	a1 := ingest.AccountCounts{Account: "acc-1", Transactions: 2, New: 2}
	a2 := ingest.AccountCounts{Account: "acc-2", Transactions: 3, New: 1, NewVersions: 1, FailedAssertions: 1}
	tests := []struct {
		name   string
		ctx    func(t *testing.T) context.Context
		authz  identityapp.Authorizer
		upload ingest.BatchRecord
		booked []ingest.AccountCounts
	}{
		{
			name: "every account viewable", ctx: reader, authz: identityapp.Policy{},
			upload: ingest.BatchRecord{
				ID: "u-1", Instance: "files", Source: "upload", Finished: at(25), Transactions: 5, New: 3, NewVersions: 1,
				FailedAssertions: 1, Problems: 2, Accounts: []ledger.AccountID{"acc-1", "acc-2"}, PerAccount: []ingest.AccountCounts{a1, a2},
			},
			booked: []ingest.AccountCounts{a1, a2},
		},
		{
			name: "a viewer of acc-1", ctx: viewerOf, authz: grantAuthorizer{viewable: []string{"acc-1"}},
			upload: ingest.BatchRecord{
				ID: "u-1", Instance: "files", Source: "upload", Finished: at(25), Transactions: 2, New: 2,
				Accounts: []ledger.AccountID{"acc-1"}, PerAccount: []ingest.AccountCounts{a1},
			},
			booked: []ingest.AccountCounts{a1},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			f := newFixture(t)
			v, err := app.New(f.ledger, f.valuations, f.statuses, f.rates, cur(t, "PLN"), cur(t, "PLN"),
				app.WithSyncHistory(countedHistory()), app.WithConnections(syncBindings{}), app.WithAuthorizer(tt.authz))
			if err != nil {
				t.Fatalf("app.New: %v", err)
			}
			ctx := tt.ctx(t)
			if diff := cmp.Diff(&tt.upload, uploadSeen(ctx, t, v)); diff != "" {
				t.Errorf("the upload in the activity (-want +got):\n%s", diff)
			}
			attempts, err := v.SyncAttempts(ctx, "xtb-main", views.ActivityBound{}, 5)
			if err != nil {
				t.Fatalf("SyncAttempts: %v", err)
			}
			want := map[ingest.BatchID][]ingest.AccountCounts{"b-1": tt.booked}
			if diff := cmp.Diff(want, attempts.Booked); diff != "" {
				t.Errorf("accounts the attempts booked into (-want +got):\n%s", diff)
			}
		})
	}
}
