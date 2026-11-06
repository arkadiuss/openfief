package app_test

import (
	"context"
	"errors"
	"slices"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/arkadiuss/openfief/internal/attention"
	"github.com/arkadiuss/openfief/internal/identity"
	identityapp "github.com/arkadiuss/openfief/internal/identity/app"
	"github.com/arkadiuss/openfief/internal/ingest"
	"github.com/arkadiuss/openfief/internal/ledger"
	"github.com/arkadiuss/openfief/internal/views"
	"github.com/arkadiuss/openfief/internal/views/app"
)

// holderViewer stands for a viewer (D-65), a holder of read access the
// core doesn't have yet: no holder it knows today.
const holderViewer identity.Holder = 99

// viewerOf returns t's context carrying a viewer.
func viewerOf(t *testing.T) context.Context {
	t.Helper()
	return identityapp.WithCaller(t.Context(), identity.Caller{Holder: holderViewer, ID: "viewer"})
}

// grantAuthorizer lets any caller view exactly the accounts in viewable,
// as SpiceDB will answer for a viewer granted them (D-163, D-164).
type grantAuthorizer struct {
	viewable []string
}

func (g grantAuthorizer) MayAct(_ context.Context, _ identity.Caller, a identity.Action, account string) (bool, error) {
	return a == identity.ActionView && slices.Contains(g.viewable, account), nil
}

// syncBindings binds xtb-main to acc-1 and files to acc-2; nbp-fx is bound
// to no account.
type syncBindings struct{}

func (syncBindings) Bindings(context.Context) ([]ingest.Binding, error) {
	return []ingest.Binding{{Instance: "files", Account: "acc-2"}, {Instance: "xtb-main", Account: "acc-1"}}, nil
}

// viewerHistory is history with nbp-fx's failing syncs and a batch of it
// that booked into no account added.
func viewerHistory(t *testing.T) (*fakeHistory, fakeChanges) {
	t.Helper()
	h, changes := history(t)
	h.attempts = append([]ingest.SyncAttempt{{ID: 5, Instance: "nbp-fx", Batch: "n-1", Ended: at(45), Class: ingest.SyncErrorTimeout, Failures: 1}}, h.attempts...)
	h.batches = append(h.batches, ingest.BatchRecord{ID: "n-0", Instance: "nbp-fx", Source: "sync", Finished: at(5)})
	h.statuses = append(h.statuses, ingest.SyncStatus{Instance: "nbp-fx", Failures: 1, LastAttempt: at(45)})
	return h, changes
}

// TestSyncsForViewers checks D-199: a caller sees the syncs, sync attempts
// and import activity of the connector instances bound to an account it may
// view; one bound to no account only a holder of read access other than a
// viewer sees (D-165).
func TestSyncsForViewers(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		ctx       func(t *testing.T) context.Context
		authz     identityapp.Authorizer
		instances []string
		batches   []ingest.BatchID
		failed    int
	}{
		{
			name: "a script token sees every instance", ctx: reader, authz: identityapp.Policy{},
			instances: []string{"xtb-main", "nbp-fx"}, batches: []ingest.BatchID{"u-2", "u-1", "b-1", "n-0"}, failed: 4,
		},
		{
			name: "a caller who may not view acc-1 loses xtb-main", ctx: reader, authz: &denyingAuthorizer{denied: "acc-1"},
			instances: []string{"nbp-fx"}, batches: []ingest.BatchID{"u-1", "n-0"}, failed: 1,
		},
		{
			name: "a viewer of acc-1 sees xtb-main, never nbp-fx", ctx: viewerOf, authz: grantAuthorizer{viewable: []string{"acc-1"}},
			instances: []string{"xtb-main"}, batches: []ingest.BatchID{"u-2", "u-1", "b-1"}, failed: 3,
		},
		{
			name: "a viewer of acc-2 sees no sync", ctx: viewerOf, authz: grantAuthorizer{viewable: []string{"acc-2"}},
			batches: []ingest.BatchID{"u-1"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			f := newFixture(t)
			h, changes := viewerHistory(t)
			v, err := app.New(f.ledger, f.valuations, f.statuses, f.rates, cur(t, "PLN"), cur(t, "PLN"),
				app.WithSyncHistory(h), app.WithRecordedChanges(changes), app.WithConnections(syncBindings{}), app.WithAuthorizer(tt.authz))
			if err != nil {
				t.Fatalf("app.New: %v", err)
			}
			ctx := tt.ctx(t)
			checkInstances(ctx, t, v, tt.instances)
			seen := activitySeen(ctx, t, v)
			if diff := cmp.Diff(tt.batches, seen.batches); diff != "" || seen.failed != tt.failed {
				t.Errorf("batches (-want +got):\n%s\nfailed attempts %d, want %d", diff, seen.failed, tt.failed)
			}
		})
	}
}

// checkInstances fails t unless the caller in ctx sees the sync status and
// attempts of exactly the instances in want, and of no other.
func checkInstances(ctx context.Context, t *testing.T, v *app.Views, want []string) {
	t.Helper()
	statuses, err := v.SyncStatuses(ctx)
	if err != nil {
		t.Fatalf("SyncStatuses: %v", err)
	}
	got := make([]string, 0, len(statuses))
	for i := range statuses {
		got = append(got, statuses[i].Instance)
	}
	if diff := cmp.Diff(slices.Sorted(slices.Values(want)), slices.Sorted(slices.Values(got))); diff != "" {
		t.Errorf("instances (-want +got):\n%s", diff)
	}
	for _, inst := range []string{"xtb-main", "nbp-fx"} {
		_, attemptsErr := v.SyncAttempts(ctx, inst, views.ActivityBound{}, 3)
		seen := slices.Contains(want, inst)
		if seen != (attemptsErr == nil) || (!seen && !errors.Is(attemptsErr, app.ErrUnknownInstance)) {
			t.Errorf("SyncAttempts(%s) = %v, want seen %v", inst, attemptsErr, seen)
		}
	}
}

// seenActivity is what a caller sees of the recent activity: the batches,
// newest first, and how many failed sync attempts.
type seenActivity struct {
	batches []ingest.BatchID
	failed  int
}

// activitySeen returns what the caller in ctx sees of v's recent activity.
func activitySeen(ctx context.Context, t *testing.T, v *app.Views) seenActivity {
	t.Helper()
	page, err := v.Activity(ctx, nil, 20)
	if err != nil {
		t.Fatalf("Activity: %v", err)
	}
	var out seenActivity
	for i := range page.Activities {
		switch a := &page.Activities[i]; {
		case a.Batch != nil:
			out.batches = append(out.batches, a.Batch.ID)
		case a.Kind == views.ActivitySyncFailed:
			out.failed++
		}
	}
	return out
}

// TestUnboundItemsForViewers checks that an attention item tied to no
// account, such as a failing sync's, is never a viewer's (D-165).
func TestUnboundItemsForViewers(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	v, err := app.New(f.ledger, f.valuations, f.statuses, f.rates, cur(t, "PLN"), cur(t, "PLN"),
		app.WithAttention(inboxItems()), app.WithAuthorizer(grantAuthorizer{viewable: []string{"acc-1", "acc-2"}}))
	if err != nil {
		t.Fatalf("app.New: %v", err)
	}
	list, err := v.Attention(viewerOf(t), &app.AttentionQuery{})
	if err != nil {
		t.Fatalf("Attention: %v", err)
	}
	for i := range list.Items {
		if len(list.Items[i].Accounts) == 0 {
			t.Errorf("a viewer sees item %d, tied to no account", list.Items[i].ID)
		}
	}
	if len(list.Items) == 0 {
		t.Error("a viewer of both accounts sees none of their items")
	}
}

// suggestionViews returns views over a withdrawal of 250 PLN from acc-1
// and its deposit into acc-2 a day later, suggested as a transfer, with
// item 8 about the pair, asking an authorizer that denies denied.
func suggestionViews(t *testing.T, denied ledger.AccountID) *app.Views {
	t.Helper()
	f := newFixture(t)
	day := date(t, "2026-03-02")
	w, wErr := ledger.NewWithdrawal(ledger.TransactionHeader{Key: ledger.TransactionKey{Source: "test", ExternalID: "w-1"}, Account: "acc-1", Date: day}, money(t, "250", "PLN"))
	d, dErr := ledger.NewDeposit(ledger.TransactionHeader{Key: ledger.TransactionKey{Source: "test", ExternalID: "d-1"}, Account: "acc-2", Date: day.AddDays(1)}, money(t, "250", "PLN"))
	if err := errors.Join(wErr, dErr); err != nil {
		t.Fatalf("building the legs: %v", err)
	}
	f.ledger.txs = map[ledger.AccountID][]ledger.Transaction{"acc-1": {w}, "acc-2": {d}}
	pair := ledger.TransferPair{Withdrawal: w.Key, Deposit: d.Key}
	items := fakeItems{item(8, attention.KindTransferSuggested, pair.String(), attention.StateOpen, "suggested")}
	v, err := app.New(f.ledger, f.valuations, f.statuses, f.rates, cur(t, "PLN"), cur(t, "PLN"),
		app.WithAttention(items), app.WithAuthorizer(&denyingAuthorizer{denied: denied}))
	if err != nil {
		t.Fatalf("app.New: %v", err)
	}
	return v
}

// TestSuggestedTransferItemTiedToBothAccounts checks that a transfer
// suggested between acc-1 and acc-2 is tied to both accounts, like a broken
// one, and hidden from a caller who may not view either (D-105, D-130).
func TestSuggestedTransferItemTiedToBothAccounts(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		q      app.AttentionQuery
		denied ledger.AccountID
		ids    []int64
	}{
		{name: "every item", ids: []int64{8}},
		{name: "the deposit's account", q: app.AttentionQuery{Account: "acc-2"}, ids: []int64{8}},
		{name: "another account", q: app.AttentionQuery{Account: "acc-3"}, ids: []int64{}},
		{name: "the withdrawal's account hidden", denied: "acc-1", ids: []int64{}},
		{name: "the deposit's account hidden", denied: "acc-2", ids: []int64{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			v := suggestionViews(t, tt.denied)
			list, err := v.Attention(reader(t), &tt.q)
			if err != nil {
				t.Fatalf("Attention: %v", err)
			}
			if diff := cmp.Diff(tt.ids, itemIDs(&list)); diff != "" {
				t.Errorf("items (-want +got):\n%s", diff)
			}
			checkBothAccounts(t, &list)
			for i := range list.Items {
				if list.Items[i].Group != attention.GroupDecisions {
					t.Errorf("group = %v, want decisions", list.Items[i].Group)
				}
			}
		})
	}
}
