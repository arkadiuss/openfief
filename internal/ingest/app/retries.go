package app

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/arkadiuss/openfief/internal/ingest"
)

// Errors of the owner's sync retries (D-196).
var (
	// ErrRetryTooSoon reports a retry of an instance the owner retried less
	// than Retries.Every ago.
	ErrRetryTooSoon = errors.New("this connector instance was retried a moment ago; try again in a minute")
	// ErrNoInstance reports a retry of an instance that is not a running
	// connector.
	ErrNoInstance = errors.New("no running connector instance of that name")
	// ErrSyncsNotRunning reports a retry while the core runs no connector
	// syncs, such as outside a cluster or before discovery has started.
	ErrSyncsNotRunning = errors.New("the core runs no connector syncs now")
)

// DefaultRetryEvery is how often the owner may retry one instance's sync.
const DefaultRetryEvery = time.Minute

// SyncTrigger starts one sync of an instance at once, in the background;
// *Scheduler implements it. It fails with ErrNoInstance for an instance
// that isn't running and with ErrSyncRunning while one of its syncs runs.
type SyncTrigger interface {
	Retry(ctx context.Context, instance string) error
}

// Retries takes the owner's requests to sync a connector instance now, the
// Inbox's "Retry now" on a sync that keeps failing (D-196): only the owner,
// only when the authorizer lets the owner change every account bound to the
// instance (D-163), at most once per Every for each instance, and never
// while a sync of it runs (the scheduler's single-flight claim, D-49).
type Retries struct {
	Owners   Owners
	Bindings ConnectorBindings
	// Trigger starts the sync; nil refuses every retry with
	// ErrSyncsNotRunning.
	Trigger SyncTrigger
	// Every is how often one instance may be retried; zero is
	// DefaultRetryEvery.
	Every time.Duration
	Now   func() time.Time
	// Audit keeps each retry with its audit entry; nil enters none.
	Audit ChangeAuditor

	mu   sync.Mutex
	last map[string]time.Time
}

// RetrySync starts one sync of instance for the owner-scoped session token
// valid on origin, and returns once it has started. A refused retry, for
// want of an owner, an instance or a minute since the last, starts nothing;
// one refused because the instance syncs, or isn't running, doesn't count
// against the minute.
func (r *Retries) RetrySync(ctx context.Context, origin, sessionToken, instance string) error {
	if err := r.Owners.CheckOwner(ctx, origin, sessionToken); err != nil {
		return fmt.Errorf("retrying a sync: %w", err)
	}
	if err := ingest.ValidateInstanceName(instance); err != nil {
		return fmt.Errorf("retrying a sync: %w: %w", ErrNoInstance, err)
	}
	bound, err := r.Bindings.BoundAccounts(ctx, instance)
	if err != nil {
		return fmt.Errorf("retrying a sync: accounts bound to %s: %w", instance, err)
	}
	if len(bound) > 0 {
		if refused := mayChange(ctx, r.Owners, origin, sessionToken, bound...); refused != nil {
			return fmt.Errorf("retrying a sync of %s: %w", instance, refused)
		}
	}
	if r.Trigger == nil {
		return fmt.Errorf("retrying a sync of %s: %w", instance, ErrSyncsNotRunning)
	}
	previous, err := r.reserve(instance)
	if err != nil {
		return err
	}
	err = audited(ctx, r.Audit, origin, sessionToken, auditSyncRetried, []string{instanceRef(instance)}, func(ctx context.Context) error {
		return r.Trigger.Retry(ctx, instance)
	})
	if err != nil {
		r.release(instance, previous)
		return fmt.Errorf("retrying a sync of %s: %w", instance, err)
	}
	return nil
}

// reserve records a retry of instance now, unless the last was less than
// Every ago, and returns the time of the one before it.
func (r *Retries) reserve(instance string) (time.Time, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	now := r.Now()
	every := r.Every
	if every <= 0 {
		every = DefaultRetryEvery
	}
	previous, ok := r.last[instance]
	if ok && now.Before(previous.Add(every)) {
		return time.Time{}, fmt.Errorf("retrying a sync of %s: %w", instance, ErrRetryTooSoon)
	}
	if r.last == nil {
		r.last = map[string]time.Time{}
	}
	r.last[instance] = now
	return previous, nil
}

// release forgets the retry of instance reserve recorded, putting back the
// time of the one before it, when the retry didn't start.
func (r *Retries) release(instance string, previous time.Time) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if previous.IsZero() {
		delete(r.last, instance)
		return
	}
	r.last[instance] = previous
}
