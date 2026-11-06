package app

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/arkadiuss/openfief/internal/identity"
	identityapp "github.com/arkadiuss/openfief/internal/identity/app"
	"github.com/arkadiuss/openfief/internal/ingest"
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
// accounts the caller may view (D-163). Imports and syncs are tied to their
// connector instance only, as an item tied to no account is (D-165).
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
	if fetched[views.SourceBatches], err = v.batchActivity(ctx, from[views.SourceBatches], fetch); err != nil {
		return views.ActivityPage{}, err
	}
	if fetched[views.SourceFailures], err = v.failureActivity(ctx, from[views.SourceFailures], fetch); err != nil {
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
// only the accounts it booked into that the caller may view (D-163), and
// hidden when it booked into some and the caller may view none of them.
func (v *Views) batchActivity(ctx context.Context, b views.ActivityBound, fetch int) ([]views.Activity, error) {
	if v.syncs == nil {
		return nil, nil
	}
	batches, err := v.syncs.CompletedBatches(ctx, &ingest.BatchQuery{Before: b.At, BeforeID: ingest.BatchID(b.Key), Limit: fetch})
	if err != nil {
		return nil, fmt.Errorf("loading import batches: %w", err)
	}
	caller := identityapp.CallerFrom(ctx)
	out := make([]views.Activity, len(batches))
	for i := range batches {
		booked := batches[i].Accounts
		batches[i].Accounts = nil
		for _, a := range booked {
			ok, authzErr := v.authz.MayAct(ctx, caller, identity.ActionView, string(a))
			if authzErr != nil {
				return nil, fmt.Errorf("authorizing a view of %s: %w", a, authzErr)
			}
			if ok {
				batches[i].Accounts = append(batches[i].Accounts, a)
			}
		}
		out[i] = views.BatchActivity(&batches[i])
		out[i].Hidden = len(booked) > 0 && len(batches[i].Accounts) == 0
	}
	return out, nil
}

// failureActivity returns up to fetch failed sync attempts after b.
func (v *Views) failureActivity(ctx context.Context, b views.ActivityBound, fetch int) ([]views.Activity, error) {
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
// instance order (D-196); none without the sync history.
func (v *Views) SyncStatuses(ctx context.Context) ([]ingest.SyncStatus, error) {
	if v.syncs == nil {
		return nil, nil
	}
	out, err := v.syncs.SyncStatuses(ctx)
	if err != nil {
		return nil, fmt.Errorf("loading sync statuses: %w", err)
	}
	return out, nil
}

// SyncAttemptPage is one page of an instance's sync attempts, newest
// first, with the instance's status; Next is nil on the last page.
type SyncAttemptPage struct {
	Status   ingest.SyncStatus
	Attempts []ingest.SyncAttempt
	Next     *views.ActivityBound
}

// SyncAttempts returns the page of at most size of instance's sync
// attempts after the bound after, or from the newest when after is the
// zero bound, with its status (D-196). An instance never synced is
// ErrUnknownInstance.
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
	return page, nil
}
