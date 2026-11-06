package app_test

import (
	"context"
	"errors"
	"slices"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/arkadiuss/openfief/internal/attention"
	"github.com/arkadiuss/openfief/internal/ingest"
	"github.com/arkadiuss/openfief/internal/jurisdiction/pl"
	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
	ledgerapp "github.com/arkadiuss/openfief/internal/ledger/app"
	"github.com/arkadiuss/openfief/internal/valuation"
	"github.com/arkadiuss/openfief/internal/views"
	"github.com/arkadiuss/openfief/internal/views/app"
)

// at is minute m of a synthetic morning.
func at(m int) time.Time {
	return time.Date(2026, time.November, 4, 6, 0, 0, 0, time.UTC).Add(time.Duration(m) * time.Minute)
}

// fakeItems lists fixed attention items.
type fakeItems []attention.Item

func (f fakeItems) List(context.Context) ([]attention.Item, error) { return f, nil }

// item is an attention item of kind about subject in state, with detail.
func item(id int64, kind attention.Kind, subject string, state attention.State, detail string) attention.Item {
	return attention.Item{ID: id, Kind: kind, Subject: subject, Changes: []attention.Change{{Seq: 1, State: state, Detail: detail, At: at(0)}}}
}

// inboxItems are one item of each scope: acc-1's balance assertion, acc-2's
// market-data gap, an instrument's price check, an instance's failing sync
// and import problems, and a missing setting.
func inboxItems() fakeItems {
	return fakeItems{
		item(1, attention.KindAssertionFlagged, "acc-1/2026-03-02/"+isinIE+"/ETF1", attention.StateOpen, "derived 2, reported 3"),
		item(2, "market-data-gap", "acc-2/fx_rates/EUR", attention.StateAcknowledged, "no rate for 5 days"),
		item(3, attention.KindPriceCheckFlagged, isinIE+"/ETF1", attention.StateOpen, "differs"),
		item(4, attention.KindSyncFailing, "xtb-main", attention.StateResolved, "failing since 2026-11-03T06:00Z"),
		item(5, attention.KindImportProblems, "b-1", attention.StateOpen, "xtb-main: 2 rows left out"),
		item(6, attention.KindSettingMissing, "marginal-rate", attention.StateOpen, "needed"),
	}
}

func TestAttentionViewsAndScopes(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		q      app.AttentionQuery
		denied ledger.AccountID
		ids    []int64
		counts [3]int
	}{
		{name: "every item", ids: []int64{1, 2, 3, 4, 5, 6}, counts: [3]int{4, 1, 1}},
		{name: "open", q: app.AttentionQuery{View: app.AttentionOpen}, ids: []int64{1, 3, 5, 6}, counts: [3]int{4, 1, 1}},
		{name: "acknowledged", q: app.AttentionQuery{View: app.AttentionAcknowledged}, ids: []int64{2}, counts: [3]int{4, 1, 1}},
		{name: "done", q: app.AttentionQuery{View: app.AttentionDone}, ids: []int64{4}, counts: [3]int{4, 1, 1}},
		{name: "an account's items", q: app.AttentionQuery{Account: "acc-1"}, ids: []int64{1}, counts: [3]int{1, 0, 0}},
		{name: "an instrument's items", q: app.AttentionQuery{Instrument: isinIE}, ids: []int64{1, 3}, counts: [3]int{2, 0, 0}},
		{name: "an instance's open items", q: app.AttentionQuery{Instance: "xtb-main", View: app.AttentionOpen}, ids: []int64{5}, counts: [3]int{1, 0, 1}},
		{name: "an account the caller may not view", denied: "acc-1", ids: []int64{2, 3, 4, 5, 6}, counts: [3]int{3, 1, 1}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			f := newFixture(t)
			v, err := app.New(f.ledger, f.valuations, f.statuses, f.rates, cur(t, "PLN"), cur(t, "PLN"),
				app.WithAttention(inboxItems()), app.WithAuthorizer(&denyingAuthorizer{denied: tt.denied}))
			if err != nil {
				t.Fatalf("app.New: %v", err)
			}
			list, err := v.Attention(reader(t), &tt.q)
			if err != nil {
				t.Fatalf("Attention: %v", err)
			}
			if diff := cmp.Diff(tt.ids, itemIDs(&list)); diff != "" {
				t.Errorf("items (-want +got):\n%s", diff)
			}
			if got := [3]int{list.Open, list.Acknowledged, list.Done}; got != tt.counts {
				t.Errorf("open, acknowledged, done = %v, want %v", got, tt.counts)
			}
		})
	}
}

// itemIDs returns the IDs of l's items, in order.
func itemIDs(l *app.AttentionList) []int64 {
	ids := make([]int64, 0, len(l.Items))
	for i := range l.Items {
		ids = append(ids, l.Items[i].ID)
	}
	return ids
}

func TestAttentionItemsCarryScopeAndGroup(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	v, err := app.New(f.ledger, f.valuations, f.statuses, f.rates, cur(t, "PLN"), cur(t, "PLN"), app.WithAttention(inboxItems()))
	if err != nil {
		t.Fatalf("app.New: %v", err)
	}
	list, err := v.Attention(reader(t), &app.AttentionQuery{})
	if err != nil {
		t.Fatalf("Attention: %v", err)
	}
	type scoped struct {
		Scope attention.Scope
		Group attention.Group
	}
	got := make([]scoped, 0, len(list.Items))
	for i := range list.Items {
		got = append(got, scoped{list.Items[i].Scope, list.Items[i].Group})
	}
	want := []scoped{
		{attention.Scope{Account: "acc-1", Instrument: isinIE}, attention.GroupNumbersToCheck},
		{attention.Scope{Account: "acc-2"}, attention.GroupDataComingIn},
		{attention.Scope{Instrument: isinIE}, attention.GroupNumbersToCheck},
		{attention.Scope{Instance: "xtb-main"}, attention.GroupDataComingIn},
		{attention.Scope{Instance: "xtb-main"}, attention.GroupDataComingIn},
		{attention.Scope{}, attention.GroupDecisions},
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("scopes and groups (-want +got):\n%s", diff)
	}
	none, err := newFixture(t).views(t).Attention(reader(t), &app.AttentionQuery{})
	if err != nil || len(none.Items) != 0 {
		t.Errorf("without the port = %+v, %v", none, err)
	}
}

// TestBrokenTransferItemTiedToBothAccounts checks that a transfer a
// correction broke is tied to both legs' accounts: each account's filter
// finds it, its scope names the withdrawal's, and a caller who may not view
// either account doesn't see it.
func TestBrokenTransferItemTiedToBothAccounts(t *testing.T) {
	t.Parallel()
	key := func(id string) ledger.TransactionKey { return ledger.TransactionKey{Source: "test", ExternalID: id} }
	pair := ledger.TransferPair{Withdrawal: key("w-1"), Deposit: key("d-1")}
	items := fakeItems{item(7, attention.KindTransferBroken, pair.String(), attention.StateOpen, "its legs no longer match")}
	tests := []struct {
		name   string
		q      app.AttentionQuery
		denied ledger.AccountID
		ids    []int64
	}{
		{name: "every item", ids: []int64{7}},
		{name: "the withdrawal's account", q: app.AttentionQuery{Account: "acc-1"}, ids: []int64{7}},
		{name: "the deposit's account", q: app.AttentionQuery{Account: "acc-2"}, ids: []int64{7}},
		{name: "another account", q: app.AttentionQuery{Account: "acc-3"}, ids: []int64{}},
		{name: "the withdrawal's account hidden", denied: "acc-1", ids: []int64{}},
		{name: "the deposit's account hidden", denied: "acc-2", ids: []int64{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			f := newFixture(t)
			f.ledger.txs = map[ledger.AccountID][]ledger.Transaction{
				"acc-1": {{Key: key("w-1"), Account: "acc-1"}}, "acc-2": {{Key: key("d-1"), Account: "acc-2"}},
			}
			f.ledger.links = []ledger.TransferLink{{Pair: pair, Decision: ledger.TransferConfirmed, Version: 1}}
			v, err := app.New(f.ledger, f.valuations, f.statuses, f.rates, cur(t, "PLN"), cur(t, "PLN"),
				app.WithAttention(items), app.WithAuthorizer(&denyingAuthorizer{denied: tt.denied}))
			if err != nil {
				t.Fatalf("app.New: %v", err)
			}
			list, err := v.Attention(reader(t), &tt.q)
			if err != nil {
				t.Fatalf("Attention: %v", err)
			}
			if diff := cmp.Diff(tt.ids, itemIDs(&list)); diff != "" {
				t.Errorf("items (-want +got):\n%s", diff)
			}
			checkBothAccounts(t, &list)
		})
	}
}

// checkBothAccounts checks that every item of l is tied to acc-1 and acc-2,
// its scope naming acc-1.
func checkBothAccounts(t *testing.T, l *app.AttentionList) {
	t.Helper()
	for i := range l.Items {
		got := &l.Items[i]
		if diff := cmp.Diff([]string{"acc-1", "acc-2"}, got.Accounts); diff != "" || got.Scope.Account != "acc-1" {
			t.Errorf("accounts (-want +got):\n%s; scope account %q, want acc-1", diff, got.Scope.Account)
		}
	}
}

// fakeHistory is the ingest store's batches, attempts and statuses.
type fakeHistory struct {
	batches  []ingest.BatchRecord
	attempts []ingest.SyncAttempt
	statuses []ingest.SyncStatus
	fail     error
}

func (f *fakeHistory) CompletedBatches(_ context.Context, q *ingest.BatchQuery) ([]ingest.BatchRecord, error) {
	var out []ingest.BatchRecord
	for i := range f.batches {
		if b := &f.batches[i]; (q.Before.IsZero() || b.Finished.Before(q.Before)) && len(out) < q.Limit && (len(q.IDs) == 0 || slices.Contains(q.IDs, b.ID)) {
			out = append(out, *b)
		}
	}
	return out, f.fail
}

func (f *fakeHistory) SyncAttempts(_ context.Context, q *ingest.AttemptQuery) ([]ingest.SyncAttempt, error) {
	var out []ingest.SyncAttempt
	for i := range f.attempts {
		a := &f.attempts[i]
		if attemptSelected(q, a) && len(out) < q.Limit {
			out = append(out, *a)
		}
	}
	return out, f.fail
}

// attemptSelected reports whether q selects a.
func attemptSelected(q *ingest.AttemptQuery, a *ingest.SyncAttempt) bool {
	after := q.Before.IsZero() || a.Ended.Before(q.Before) || (a.Ended.Equal(q.Before) && a.ID < q.BeforeID)
	return (q.Instance == "" || a.Instance == q.Instance) && (!q.FailedOnly || a.Failed()) && after
}

func (f *fakeHistory) SyncStatuses(context.Context) ([]ingest.SyncStatus, error) {
	return f.statuses, f.fail
}

// fakeChanges is the ledger's coupons and corrections, newest first.
type fakeChanges []ledgerapp.RecordedVersion

func (f fakeChanges) RecordedChanges(_ context.Context, before time.Time, _ int64, limit int) ([]ledgerapp.RecordedVersion, error) {
	var out []ledgerapp.RecordedVersion
	for i := range f {
		if c := &f[i]; (before.IsZero() || c.Recorded.Before(before)) && len(out) < limit {
			out = append(out, *c)
		}
	}
	return out, nil
}

// history is xtb-main's sync history: a sync at minute 10, failures at
// 20, 30 and 40, an upload into acc-1 and acc-2 at 25 and one into acc-1
// at 27; the ledger corrected an acc-2
// transaction at 35 and booked an acc-1 coupon at 15.
func history(t *testing.T) (*fakeHistory, fakeChanges) {
	t.Helper()
	fail := func(id int64, m, n int) ingest.SyncAttempt {
		return ingest.SyncAttempt{ID: id, Instance: "xtb-main", Batch: "b-2", Ended: at(m), Class: ingest.SyncErrorTimeout, Failures: n}
	}
	h := &fakeHistory{
		batches: []ingest.BatchRecord{
			{ID: "u-2", Instance: "files", Source: "upload", Finished: at(27), Accounts: []ledger.AccountID{"acc-1"}},
			{ID: "u-1", Instance: "files", Source: "upload", Finished: at(25), Accounts: []ledger.AccountID{"acc-1", "acc-2"}},
			{ID: "b-1", Instance: "xtb-main", Source: "sync", Finished: at(10)},
		},
		attempts: []ingest.SyncAttempt{fail(4, 40, 3), fail(3, 30, 2), fail(2, 20, 1), {ID: 1, Instance: "xtb-main", Batch: "b-1", Ended: at(10)}},
		statuses: []ingest.SyncStatus{{Instance: "xtb-main", Failures: 3, FailingSince: at(19), LastAttempt: at(40), LastClass: ingest.SyncErrorTimeout}},
	}
	key := ledger.TransactionKey{Source: "test", ExternalID: "x"}
	changes := fakeChanges{
		{VersionedTransaction: ledger.VersionedTransaction{Transaction: ledger.Transaction{Key: key, Account: "acc-2"}, Version: 2}, ID: 9, Recorded: at(35)},
		{VersionedTransaction: ledger.VersionedTransaction{Transaction: ledger.Transaction{Key: key, Account: "acc-1", Kind: ledger.KindInterest}, Version: 1}, ID: 8, Recorded: at(15)},
	}
	return h, changes
}

func TestActivity(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		denied ledger.AccountID
		want   []views.ActivityKind
	}{
		{name: "everything, newest first", want: []views.ActivityKind{
			views.ActivitySyncFailed, views.ActivityCorrection, views.ActivitySyncFailed, views.ActivityImport, views.ActivityImport,
			views.ActivitySyncFailed, views.ActivityCoupon, views.ActivitySync,
		}},
		{name: "a correction in an account the caller may not view", denied: "acc-2", want: []views.ActivityKind{
			views.ActivitySyncFailed, views.ActivitySyncFailed, views.ActivityImport, views.ActivityImport,
			views.ActivitySyncFailed, views.ActivityCoupon, views.ActivitySync,
		}},
		{name: "an upload booked only into an account the caller may not view", denied: "acc-1", want: []views.ActivityKind{
			views.ActivitySyncFailed, views.ActivityCorrection, views.ActivitySyncFailed, views.ActivityImport,
			views.ActivitySyncFailed, views.ActivitySync,
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			f := newFixture(t)
			h, changes := history(t)
			v, err := app.New(f.ledger, f.valuations, f.statuses, f.rates, cur(t, "PLN"), cur(t, "PLN"),
				app.WithSyncHistory(h), app.WithRecordedChanges(changes), app.WithAuthorizer(&denyingAuthorizer{denied: tt.denied}))
			if err != nil {
				t.Fatalf("app.New: %v", err)
			}
			if diff := cmp.Diff(tt.want, activityKinds(t, v)); diff != "" {
				t.Errorf("kinds (-want +got):\n%s", diff)
			}
		})
	}
}

// TestActivityBatchAccounts checks that an upload names the accounts it
// booked into that the caller may view, and none it may not (D-140, D-163).
func TestActivityBatchAccounts(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	h, changes := history(t)
	v, err := app.New(f.ledger, f.valuations, f.statuses, f.rates, cur(t, "PLN"), cur(t, "PLN"),
		app.WithSyncHistory(h), app.WithRecordedChanges(changes), app.WithAuthorizer(&denyingAuthorizer{denied: "acc-1"}))
	if err != nil {
		t.Fatalf("app.New: %v", err)
	}
	page, err := v.Activity(reader(t), nil, 10)
	if err != nil {
		t.Fatalf("Activity: %v", err)
	}
	got := map[ingest.BatchID][]ledger.AccountID{}
	for i := range page.Activities {
		if b := page.Activities[i].Batch; b != nil {
			got[b.ID] = b.Accounts
		}
	}
	want := map[ingest.BatchID][]ledger.AccountID{"u-1": {"acc-2"}, "b-1": nil}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("batches' accounts (-want +got):\n%s", diff)
	}
}

// activityKinds reads v's whole activity two entries a page and returns
// their kinds, in order.
func activityKinds(t *testing.T, v *app.Views) []views.ActivityKind {
	t.Helper()
	var got []views.ActivityKind
	var after *views.ActivityCursor
	for range 10 {
		page, err := v.Activity(reader(t), after, 2)
		if err != nil {
			t.Fatalf("Activity: %v", err)
		}
		for i := range page.Activities {
			got = append(got, page.Activities[i].Kind)
		}
		if page.Next == nil {
			break
		}
		after = page.Next
	}
	return got
}

func TestActivityWithoutPortsAndFailures(t *testing.T) {
	t.Parallel()
	page, err := newFixture(t).views(t).Activity(reader(t), nil, 10)
	if err != nil || len(page.Activities) != 0 || page.Next != nil {
		t.Errorf("without the ports = %+v, %v", page, err)
	}
	if _, sizeErr := newFixture(t).views(t).Activity(reader(t), nil, 0); !errors.Is(sizeErr, views.ErrInvalidPageSize) {
		t.Errorf("a page of 0 = %v", sizeErr)
	}
	f := newFixture(t)
	v, err := app.New(f.ledger, f.valuations, f.statuses, f.rates, cur(t, "PLN"), cur(t, "PLN"), app.WithSyncHistory(&fakeHistory{fail: errStore}))
	if err != nil {
		t.Fatalf("app.New: %v", err)
	}
	if _, err := v.Activity(reader(t), nil, 10); !errors.Is(err, errStore) {
		t.Errorf("a failing store = %v", err)
	}
	if _, err := v.SyncStatuses(reader(t)); !errors.Is(err, errStore) {
		t.Errorf("SyncStatuses with a failing store = %v", err)
	}
}

func TestSyncStatusesKeepFailing(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	h, _ := history(t)
	v, err := app.New(f.ledger, f.valuations, f.statuses, f.rates, cur(t, "PLN"), cur(t, "PLN"), app.WithSyncHistory(h))
	if err != nil {
		t.Fatalf("app.New: %v", err)
	}
	statuses, err := v.SyncStatuses(reader(t))
	if err != nil || len(statuses) != 1 || !statuses[0].KeepsFailing() {
		t.Fatalf("SyncStatuses = %+v, %v; want xtb-main keeping failing", statuses, err)
	}
}

func TestSyncAttemptsPages(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	h, _ := history(t)
	v, err := app.New(f.ledger, f.valuations, f.statuses, f.rates, cur(t, "PLN"), cur(t, "PLN"), app.WithSyncHistory(h))
	if err != nil {
		t.Fatalf("app.New: %v", err)
	}
	first, err := v.SyncAttempts(reader(t), "xtb-main", views.ActivityBound{}, 3)
	if err != nil {
		t.Fatalf("SyncAttempts: %v", err)
	}
	if len(first.Attempts) != 3 || first.Next == nil || first.Status.Failures != 3 {
		t.Fatalf("first page = %+v", first)
	}
	rest, err := v.SyncAttempts(reader(t), "xtb-main", *first.Next, 3)
	if err != nil || len(rest.Attempts) != 1 || rest.Attempts[0].ID != 1 || rest.Next != nil {
		t.Errorf("second page = %+v, %v", rest, err)
	}
}

func TestSyncAttemptsRefuses(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	h, _ := history(t)
	v, err := app.New(f.ledger, f.valuations, f.statuses, f.rates, cur(t, "PLN"), cur(t, "PLN"), app.WithSyncHistory(h))
	if err != nil {
		t.Fatalf("app.New: %v", err)
	}
	if _, err := v.SyncAttempts(reader(t), "nbp-fx", views.ActivityBound{}, 3); !errors.Is(err, app.ErrUnknownInstance) {
		t.Errorf("an instance never synced = %v, want ErrUnknownInstance", err)
	}
	if _, err := v.SyncAttempts(reader(t), "xtb-main", views.ActivityBound{}, views.MaxPageSize+1); !errors.Is(err, views.ErrInvalidPageSize) {
		t.Errorf("an oversized page = %v", err)
	}
}

// fakePrices is valuation's price store over fixed prices and sources.
type fakePrices struct {
	prices  []valuation.Price
	sources map[string]string
}

func (f *fakePrices) PriceHistory(_ context.Context, isin string, from, to kernel.Date) ([]valuation.Price, error) {
	var out []valuation.Price
	for _, p := range f.prices {
		if in, ok := views.ListingOf(p.Instrument); ok && in.ISIN.String() == isin && !p.Date.Before(from) && !p.Date.After(to) {
			out = append(out, p)
		}
	}
	return out, nil
}

func (f *fakePrices) LatestPrices(ctx context.Context, isin string) ([]valuation.Price, error) {
	all, err := f.PriceHistory(ctx, isin, kernel.Date{}, kernel.Date{}.AddDays(1_000_000))
	latest := map[[2]string]int{}
	var out []valuation.Price
	for _, p := range all {
		k := [2]string{p.Instrument, p.Source}
		if i, seen := latest[k]; seen {
			out[i] = p
			continue
		}
		latest[k] = len(out)
		out = append(out, p)
	}
	return out, err
}

func (f *fakePrices) SourceOf(_ context.Context, s valuation.Series) (valuation.SeriesSource, bool, error) {
	src, ok := f.sources[s.Key]
	return valuation.SeriesSource{Series: s, Source: src}, ok, nil
}

// instrumentViews returns the fixture's views with acc-1 holding ETF1 and
// ETF2 of isinIE, bought in two transactions, acc-2 buying ETF1 too, and
// prices of ETF1 from two sources.
func instrumentViews(t *testing.T, denied ledger.AccountID) (*app.Views, *fakePrices) {
	t.Helper()
	f := newFixture(t)
	etf1, etf2 := instrument(t, isinIE, "ETF1"), instrument(t, isinIE, "ETF2")
	buy := func(acc ledger.AccountID, id, day string, in ledger.Instrument) ledger.Transaction {
		h := ledger.TransactionHeader{Key: ledger.TransactionKey{Source: "test", ExternalID: id}, Account: acc, Date: date(t, day)}
		tx, err := ledger.NewBuy(h, in, kernel.NewQuantity(kernel.NewDecimal(1, 0)), money(t, "10", "USD"), money(t, "0", "USD"))
		if err != nil {
			t.Fatalf("NewBuy: %v", err)
		}
		return tx
	}
	f.ledger.txs = map[ledger.AccountID][]ledger.Transaction{
		"acc-1": {buy("acc-1", "b1", "2026-02-01", etf1), buy("acc-1", "b2", "2026-02-02", etf2)},
		"acc-2": {buy("acc-2", "b3", "2026-02-03", etf1)},
	}
	f.ledger.lots = map[ledger.AccountID][]ledger.Lot{"acc-1": {
		{Instrument: etf1, Opened: date(t, "2026-02-01"), Quantity: kernel.NewQuantity(kernel.NewDecimal(3, 0)), Cost: money(t, "30", "USD")},
	}}
	price := func(in ledger.Instrument, day, source string) valuation.Price {
		return valuation.Price{Instrument: in.String(), Date: date(t, day), Close: money(t, "10", "USD"), Source: source}
	}
	prices := &fakePrices{
		prices: []valuation.Price{
			price(etf1, "2026-03-01", "eodhd"), price(etf1, "2026-03-01", "ibkr"), price(etf1, "2026-03-02", "eodhd"),
		},
		sources: map[string]string{etf1.String(): "eodhd"},
	}
	v, err := app.New(f.ledger, f.valuations, f.statuses, f.rates, cur(t, "PLN"), cur(t, "PLN"),
		app.WithInstrumentPrices(prices), app.WithJurisdiction(pl.NewPack(), &fakeTerms{}),
		app.WithAuthorizer(&denyingAuthorizer{denied: denied}))
	if err != nil {
		t.Fatalf("app.New: %v", err)
	}
	return v, prices
}

func TestInstrument(t *testing.T) {
	t.Parallel()
	v, prices := instrumentViews(t, "")
	in, err := v.Instrument(reader(t), isinIE)
	if err != nil {
		t.Fatalf("Instrument: %v", err)
	}
	latest := prices.prices[2]
	want := []app.Listing{
		{Instrument: instrument(t, isinIE, "ETF1"), Source: "eodhd", Latest: &latest},
		{Instrument: instrument(t, isinIE, "ETF2")},
	}
	if diff := cmp.Diff(want, in.Listings); diff != "" {
		t.Errorf("listings (-want +got):\n%s", diff)
	}
	if in.AssetClass != views.AssetListed || len(in.Classes) == 0 || in.Classes[0].Wrapper != "oki" {
		t.Errorf("asset class %v, classes %+v", in.AssetClass, in.Classes)
	}
	if _, unknownErr := v.Instrument(reader(t), isinPL); !errors.Is(unknownErr, app.ErrUnknownInstrument) {
		t.Errorf("an instrument nobody holds or prices = %v, want ErrUnknownInstrument", unknownErr)
	}
}

func TestInstrumentPrices(t *testing.T) {
	t.Parallel()
	v, _ := instrumentViews(t, "")
	pricesOnly, err := v.InstrumentPrices(reader(t), isinIE, date(t, "2026-03-01"), date(t, "2026-03-01"))
	if err != nil {
		t.Fatalf("InstrumentPrices: %v", err)
	}
	marks := make([]bool, 0, len(pricesOnly))
	for i := range pricesOnly {
		marks = append(marks, pricesOnly[i].Primary)
	}
	if diff := cmp.Diff([]bool{true, false}, marks); diff != "" {
		t.Errorf("primary marks (-want +got):\n%s", diff)
	}
	if _, err := v.InstrumentPrices(reader(t), isinIE, date(t, "2026-03-02"), date(t, "2026-03-01")); !errors.Is(err, views.ErrInvalidRange) {
		t.Errorf("a reversed range = %v", err)
	}
}

func TestInstrumentAcrossViewableAccounts(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		denied   ledger.AccountID
		txs      []string
		holdings []ledger.AccountID
	}{
		{name: "every account", txs: []string{"b1", "b2", "b3"}, holdings: []ledger.AccountID{"acc-1", "acc-1"}},
		{name: "acc-1 hidden", denied: "acc-1", txs: []string{"b3"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			v, _ := instrumentViews(t, tt.denied)
			page, err := v.InstrumentTransactions(reader(t), isinIE, nil, 10)
			if err != nil {
				t.Fatalf("InstrumentTransactions: %v", err)
			}
			if diff := cmp.Diff(tt.txs, externalIDs(page.Transactions)); diff != "" {
				t.Errorf("transactions (-want +got):\n%s", diff)
			}
			holdings, err := v.InstrumentHoldings(reader(t), isinIE, date(t, "2026-03-02"))
			if err != nil {
				t.Fatalf("InstrumentHoldings: %v", err)
			}
			if diff := cmp.Diff(tt.holdings, holdingAccounts(holdings), cmpopts.EquateEmpty()); diff != "" {
				t.Errorf("holdings' accounts (-want +got):\n%s", diff)
			}
		})
	}
}

// externalIDs returns the external IDs of txs, in order.
func externalIDs(txs []ledger.Transaction) []string {
	ids := make([]string, 0, len(txs))
	for i := range txs {
		ids = append(ids, txs[i].Key.ExternalID)
	}
	return ids
}

// holdingAccounts returns the account of each of holdings, in order.
func holdingAccounts(holdings []views.InstrumentHolding) []ledger.AccountID {
	accounts := make([]ledger.AccountID, 0, len(holdings))
	for i := range holdings {
		accounts = append(accounts, holdings[i].Account)
	}
	return accounts
}

func TestPriceChecksOfAnInstrument(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	check := func(inst string) valuation.PriceCheck {
		return valuation.PriceCheck{CrossCheck: valuation.CrossCheckPrice{Instrument: inst}, Status: valuation.PriceMatched}
	}
	v, err := app.New(f.ledger, f.valuations, f.statuses, f.rates, cur(t, "PLN"), cur(t, "PLN"),
		app.WithPriceChecks(fakePriceChecks{checks: []valuation.PriceCheck{check(isinIE + "/ETF1"), check(isinPL + "/EDO0334"), check(isinIE + "/ETF2")}}))
	if err != nil {
		t.Fatalf("app.New: %v", err)
	}
	list, err := v.PriceChecks(reader(t), 0, isinIE)
	if err != nil {
		t.Fatalf("PriceChecks: %v", err)
	}
	if len(list.Checks) != 2 || list.Checks[1].CrossCheck.Instrument != isinIE+"/ETF2" {
		t.Errorf("checks of %s = %+v", isinIE, list.Checks)
	}
}
