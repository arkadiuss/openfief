package app

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/arkadiuss/openfief/internal/identity"
	identityapp "github.com/arkadiuss/openfief/internal/identity/app"
	"github.com/arkadiuss/openfief/internal/ingest"
	"github.com/arkadiuss/openfief/internal/ledger"
	ledgerapp "github.com/arkadiuss/openfief/internal/ledger/app"
	"github.com/arkadiuss/openfief/internal/views"
)

// ErrUnknownInstance reports a connector instance that has never been
// synced.
var ErrUnknownInstance = errors.New("unknown connector instance")

// SyncHistory reads the connectors' import batches and sync attempts (D-49,
// D-196); *ingest/postgres.Store implements it.
type SyncHistory interface {
	// CompletedBatches returns the completed batches q selects, newest
	// first.
	CompletedBatches(ctx context.Context, q *ingest.BatchQuery) ([]ingest.BatchRecord, error)
	// SyncAttempts returns the sync attempts q selects, newest first.
	SyncAttempts(ctx context.Context, q *ingest.AttemptQuery) ([]ingest.SyncAttempt, error)
	// SyncStatuses returns every instance's sync status, in instance
	// order.
	SyncStatuses(ctx context.Context) ([]ingest.SyncStatus, error)
}

// RecordedChanges reads the ledger's coupons and corrections with when each
// was recorded (D-193); *ledger/postgres.Store implements it.
type RecordedChanges interface {
	// RecordedChanges returns up to limit coupons and corrections, newest
	// first, after the one recorded at before with row ID beforeID unless
	// before is zero.
	RecordedChanges(ctx context.Context, before time.Time, beforeID int64, limit int) ([]ledgerapp.RecordedVersion, error)
}

// WithSyncHistory reads the batches and sync attempts from h; without it
// the activity has no imports or syncs and no instance has a status.
func WithSyncHistory(h SyncHistory) Option {
	return func(v *Views) { v.syncs = h }
}

// WithRecordedChanges reads the coupons and corrections from c; without it
// the activity has none.
func WithRecordedChanges(c RecordedChanges) Option {
	return func(v *Views) { v.changes = c }
}

// Activity returns the page of at most size recent activities after the
// cursor after, or from the newest when after is nil (D-193): completed
// import batches, failed sync attempts, and coupons and corrections in the
// accounts the caller may view (D-163). An import or sync is listed when
// it booked into an account the caller may view, naming only those; one
// that booked into none, and a failed sync attempt, are tied to their
// connector instance, and listed when the caller may view an account the
// instance is bound to, or, for an instance bound to none, as
// unboundVisible says (D-199, D-165).
func (v *Views) Activity(ctx context.Context, after *views.ActivityCursor, size int) (views.ActivityPage, error) {
	if size < 1 || size > views.MaxPageSize {
		return views.ActivityPage{}, fmt.Errorf("%w: %d", views.ErrInvalidPageSize, size)
	}
	var from views.ActivityCursor
	if after != nil {
		from = *after
	}
	fetch := size + 1
	var fetched [3][]views.Activity
	var err error
	bound, err := v.instanceAccounts(ctx)
	if err != nil {
		return views.ActivityPage{}, err
	}
	if fetched[views.SourceBatches], err = v.batchActivity(ctx, from[views.SourceBatches], fetch, bound); err != nil {
		return views.ActivityPage{}, err
	}
	if fetched[views.SourceFailures], err = v.failureActivity(ctx, from[views.SourceFailures], fetch, bound); err != nil {
		return views.ActivityPage{}, err
	}
	if fetched[views.SourceChanges], err = v.changeActivity(ctx, from[views.SourceChanges], fetch); err != nil {
		return views.ActivityPage{}, err
	}
	page, err := views.MergeActivity(&from, &fetched, fetch, size)
	if err != nil {
		return views.ActivityPage{}, fmt.Errorf("merging the activity: %w", err)
	}
	return page, nil
}

// batchActivity returns up to fetch completed batches after b, each naming
// only the accounts it booked into that the caller may view, with only
// their counts (D-163, D-199: views.ScopeBatch), and
// hidden when it booked into some and the caller may view none of them;
// one that booked into none is hidden unless the caller may see its
// instance, bound to the accounts in bound (D-199).
func (v *Views) batchActivity(ctx context.Context, b views.ActivityBound, fetch int, bound map[string][]ledger.AccountID) ([]views.Activity, error) {
	if v.syncs == nil {
		return nil, nil
	}
	batches, err := v.syncs.CompletedBatches(ctx, &ingest.BatchQuery{Before: b.At, BeforeID: ingest.BatchID(b.Key), Limit: fetch})
	if err != nil {
		return nil, fmt.Errorf("loading import batches: %w", err)
	}
	out := make([]views.Activity, len(batches))
	for i := range batches {
		booked := batches[i].Accounts
		viewable, viewErr := v.viewableIDs(ctx, booked)
		if viewErr != nil {
			return nil, viewErr
		}
		batches[i] = views.ScopeBatch(&batches[i], viewable)
		out[i] = views.BatchActivity(&batches[i])
		visible := len(batches[i].Accounts) > 0
		if len(booked) == 0 {
			if visible, err = v.mayViewAny(ctx, bound[batches[i].Instance]); err != nil {
				return nil, err
			}
		}
		out[i].Hidden = !visible
	}
	return out, nil
}

// viewableIDs returns those of accounts the caller in ctx may view, in
// their order (D-163); none for none.
func (v *Views) viewableIDs(ctx context.Context, accounts []ledger.AccountID) ([]ledger.AccountID, error) {
	caller := identityapp.CallerFrom(ctx)
	var out []ledger.AccountID
	for _, a := range accounts {
		ok, err := v.authz.MayAct(ctx, caller, identity.ActionView, string(a))
		if err != nil {
			return nil, fmt.Errorf("authorizing a view of %s: %w", a, err)
		}
		if ok {
			out = append(out, a)
		}
	}
	return out, nil
}

// failureActivity returns up to fetch failed sync attempts after b, each
// hidden unless the caller may see its instance, bound to the accounts in
// bound (D-199).
func (v *Views) failureActivity(ctx context.Context, b views.ActivityBound, fetch int, bound map[string][]ledger.AccountID) ([]views.Activity, error) {
	if v.syncs == nil {
		return nil, nil
	}
	failed, err := v.syncs.SyncAttempts(ctx, &ingest.AttemptQuery{FailedOnly: true, Before: b.At, BeforeID: b.ID, Limit: fetch})
	if err != nil {
		return nil, fmt.Errorf("loading failed sync attempts: %w", err)
	}
	out := make([]views.Activity, len(failed))
	for i := range failed {
		out[i] = views.FailureActivity(&failed[i])
		visible, viewErr := v.mayViewAny(ctx, bound[failed[i].Instance])
		if viewErr != nil {
			return nil, viewErr
		}
		out[i].Hidden = !visible
	}
	return out, nil
}

// changeActivity returns up to fetch coupons and corrections after b, each
// hidden when it is in an account the caller may not view (D-163).
func (v *Views) changeActivity(ctx context.Context, b views.ActivityBound, fetch int) ([]views.Activity, error) {
	if v.changes == nil {
		return nil, nil
	}
	changes, err := v.changes.RecordedChanges(ctx, b.At, b.ID, fetch)
	if err != nil {
		return nil, fmt.Errorf("loading coupons and corrections: %w", err)
	}
	caller := identityapp.CallerFrom(ctx)
	out := make([]views.Activity, len(changes))
	for i := range changes {
		c := &changes[i]
		out[i] = views.ChangeActivity(&c.VersionedTransaction, c.ID, c.Recorded)
		ok, authzErr := v.authz.MayAct(ctx, caller, identity.ActionView, string(c.Transaction.Account))
		if authzErr != nil {
			return nil, fmt.Errorf("authorizing a view of %s: %w", c.Transaction.Account, authzErr)
		}
		out[i].Hidden = !ok
	}
	return out, nil
}

// SyncStatuses returns where each connector instance's syncs stand, in
// instance order (D-196), of the instances the caller may see: one bound
// to an account the caller may view, and one bound to none as
// unboundVisible says (D-199, D-165); none without the sync history.
func (v *Views) SyncStatuses(ctx context.Context) ([]ingest.SyncStatus, error) {
	if v.syncs == nil {
		return nil, nil
	}
	all, err := v.syncs.SyncStatuses(ctx)
	if err != nil {
		return nil, fmt.Errorf("loading sync statuses: %w", err)
	}
	bound, err := v.instanceAccounts(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]ingest.SyncStatus, 0, len(all))
	for i := range all {
		visible, viewErr := v.mayViewAny(ctx, bound[all[i].Instance])
		if viewErr != nil {
			return nil, viewErr
		}
		if visible {
			out = append(out, all[i])
		}
	}
	return out, nil
}

// SyncAttemptPage is one page of an instance's sync attempts, newest
// first, with the instance's status; Next is nil on the last page.
type SyncAttemptPage struct {
	Status   ingest.SyncStatus
	Attempts []ingest.SyncAttempt
	Next     *views.ActivityBound
	// Booked are, by batch, the counts of each account the batch of an
	// attempt that went through booked into, of the accounts the caller
	// may view (D-199).
	Booked map[ingest.BatchID][]ingest.AccountCounts
}

// SyncAttempts returns the page of at most size of instance's sync
// attempts after the bound after, or from the newest when after is the
// zero bound, with its status (D-196). An instance never synced, or one the
// caller may not see (SyncStatuses, D-199), is ErrUnknownInstance.
func (v *Views) SyncAttempts(ctx context.Context, instance string, after views.ActivityBound, size int) (SyncAttemptPage, error) {
	if size < 1 || size > views.MaxPageSize {
		return SyncAttemptPage{}, fmt.Errorf("%w: %d", views.ErrInvalidPageSize, size)
	}
	statuses, err := v.SyncStatuses(ctx)
	if err != nil {
		return SyncAttemptPage{}, err
	}
	var page SyncAttemptPage
	found := false
	for i := range statuses {
		if statuses[i].Instance == instance {
			page.Status, found = statuses[i], true
		}
	}
	if !found {
		return SyncAttemptPage{}, fmt.Errorf("%w: %s", ErrUnknownInstance, instance)
	}
	attempts, err := v.syncs.SyncAttempts(ctx, &ingest.AttemptQuery{Instance: instance, Before: after.At, BeforeID: after.ID, Limit: size + 1})
	if err != nil {
		return SyncAttemptPage{}, fmt.Errorf("loading sync attempts of %s: %w", instance, err)
	}
	if len(attempts) > size {
		attempts = attempts[:size]
		last := &attempts[size-1]
		page.Next = &views.ActivityBound{At: last.Ended, ID: last.ID}
	}
	page.Attempts = attempts
	if page.Booked, err = v.bookedBy(ctx, attempts); err != nil {
		return SyncAttemptPage{}, err
	}
	return page, nil
}

// bookedBy returns, by batch, the counts of the accounts the caller may
// view that the batch of each of attempts that went through booked into
// (D-199).
func (v *Views) bookedBy(ctx context.Context, attempts []ingest.SyncAttempt) (map[ingest.BatchID][]ingest.AccountCounts, error) {
	var ids []ingest.BatchID
	for i := range attempts {
		if !attempts[i].Failed() {
			ids = append(ids, attempts[i].Batch)
		}
	}
	out := map[ingest.BatchID][]ingest.AccountCounts{}
	if len(ids) == 0 {
		return out, nil
	}
	batches, err := v.syncs.CompletedBatches(ctx, &ingest.BatchQuery{IDs: ids, Limit: len(ids)})
	if err != nil {
		return nil, fmt.Errorf("loading the attempts' batches: %w", err)
	}
	for i := range batches {
		viewable, viewErr := v.viewableIDs(ctx, batches[i].Accounts)
		if viewErr != nil {
			return nil, viewErr
		}
		out[batches[i].ID] = views.ScopeBatch(&batches[i], viewable).PerAccount
	}
	return out, nil
}
