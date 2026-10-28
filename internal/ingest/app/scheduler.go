package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/arkadiuss/openfief/internal/ingest"
)

// ErrSyncRunning reports a sync of an instance that is already syncing, in
// this process or, through the store's claim, in another.
var ErrSyncRunning = errors.New("a sync of the instance is already running")

// leaseMargin is how much longer than one attempt's timeout a claim on an
// instance lasts, so a claim outlives the attempt that holds it.
const leaseMargin = time.Minute

// InstanceLister lists the connector instances that run now (D-48): those
// to sync. An uninstalled instance is no longer listed (D-58).
type InstanceLister interface {
	ConnectorInstances(ctx context.Context) ([]string, error)
}

// OpenBatch is the import batch of an instance's sync that has not
// completed yet: a failed attempt leaves it open, and the next attempt
// retries it with the same ID and mode.
type OpenBatch struct {
	ID   ingest.BatchID
	Mode ingest.Mode
}

// SyncClaim is one instance's sync state, claimed for one attempt.
type SyncClaim struct {
	// Watermark is what the last successful sync returned, empty before it.
	Watermark string
	// Seeded reports whether an opening batch of the instance completed.
	Seeded bool
	// Failures counts the attempts that failed since the last success.
	Failures int
	// Open is the batch a failed attempt left open, nil when none.
	Open *OpenBatch
}

// SyncState is when an instance's next sync is due. An instance that has
// no state was never synced, and is due at once.
type SyncState struct {
	Instance string
	NextSync time.Time
}

// SyncStore keeps each connector instance's watermark and sync state, and
// the history of its import batches.
type SyncStore interface {
	// ClaimSync claims instance for one attempt from now until until, and
	// returns its state; it fails with ErrSyncRunning while another
	// unexpired claim holds it.
	ClaimSync(ctx context.Context, instance string, now, until time.Time) (SyncClaim, error)
	// StartBatch records b as the claimed instance's open batch, started at
	// at; a batch it replaces is recorded as abandoned.
	StartBatch(ctx context.Context, instance string, b OpenBatch, at time.Time) error
	// CompleteSync records the open batch b as completed at at, stores
	// watermark, marks the instance seeded after an opening batch, resets
	// its failures, sets its next sync and releases the claim, all at once.
	CompleteSync(ctx context.Context, instance string, b *ingest.Batch, watermark string, next, at time.Time) error
	// FailSync records a failed attempt of the open batch id at at, keeps
	// the batch open, sets the instance's failures and next sync and
	// releases the claim.
	FailSync(ctx context.Context, instance string, id ingest.BatchID, failures int, next, at time.Time) error
	// SyncStates returns the state of every instance that has one.
	SyncStates(ctx context.Context) ([]SyncState, error)
}

// Scheduler owns connector instances' syncs (D-49): it syncs every running
// instance when its SyncPolicy says it is due, one sync per instance at a
// time, retries a failed batch with the same batch ID after a backoff, and
// stores the watermark of each completed sync. An instance that is no
// longer listed is no longer synced; its state and data stay (D-58).
type Scheduler struct {
	Connectors *Connectors
	Instances  InstanceLister
	Store      SyncStore
	Policy     ingest.SyncPolicy
	// Timeout bounds one attempt: the call and the import.
	Timeout time.Duration
	Now     func() time.Time
	// NewBatchID returns a new, unique import batch ID.
	NewBatchID func() ingest.BatchID
	// Jitter returns a random duration in [0, limit).
	Jitter func(limit time.Duration) time.Duration
	// Log gets one line per attempt, with IDs and counts only; nil logs
	// nothing.
	Log *slog.Logger

	mu      sync.Mutex
	running map[string]bool
	wg      sync.WaitGroup
}

// Run ticks at once and then on every tick until ctx ends, then waits for
// the syncs it started. A failed tick is logged and the next one tried.
func (s *Scheduler) Run(ctx context.Context, ticks <-chan time.Time) {
	for {
		if err := s.Tick(ctx); err != nil {
			s.logger().Warn("connector sync tick failed", "error", err)
		}
		select {
		case <-ctx.Done():
			s.Wait()
			return
		case <-ticks:
		}
	}
}

// Tick starts, in the background, a sync of every listed instance that is
// due and not syncing in this process. Wait waits for them.
func (s *Scheduler) Tick(ctx context.Context) error {
	names, err := s.Instances.ConnectorInstances(ctx)
	if err != nil {
		return fmt.Errorf("listing connector instances: %w", err)
	}
	states, err := s.Store.SyncStates(ctx)
	if err != nil {
		return fmt.Errorf("loading sync states: %w", err)
	}
	next := make(map[string]time.Time, len(states))
	for _, st := range states {
		next[st.Instance] = st.NextSync
	}
	now := s.Now()
	for _, name := range names {
		if now.Before(next[name]) || !s.begin(name) {
			continue
		}
		s.wg.Go(func() {
			defer s.end(name)
			if _, err := s.attempt(ctx, name, 0); err != nil && !errors.Is(err, ErrSyncRunning) {
				s.logger().Warn("connector sync failed", "instance", name, "error", err)
			}
		})
	}
	return nil
}

// Wait waits for the syncs Tick started.
func (s *Scheduler) Wait() { s.wg.Wait() }

// SyncNow syncs instance at once and waits for it, in mode, or with a zero
// mode in the mode its state calls for. It fails with ErrSyncRunning while
// the instance syncs. A failure counts like a scheduled one.
func (s *Scheduler) SyncNow(ctx context.Context, instance string, mode ingest.Mode) (SyncResult, error) {
	if !s.begin(instance) {
		return SyncResult{}, fmt.Errorf("%w: %s", ErrSyncRunning, instance)
	}
	defer s.end(instance)
	return s.attempt(ctx, instance, mode)
}

// begin marks instance as syncing in this process, unless it already is.
func (s *Scheduler) begin(instance string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.running == nil {
		s.running = map[string]bool{}
	}
	if s.running[instance] {
		return false
	}
	s.running[instance] = true
	return true
}

// end marks instance as no longer syncing in this process.
func (s *Scheduler) end(instance string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.running, instance)
}

// attempt runs one sync attempt of instance: it claims the instance,
// retries its open batch or starts one, calls the connector and imports
// the delivery within the timeout, and records the outcome with the next
// sync's due time.
func (s *Scheduler) attempt(ctx context.Context, instance string, mode ingest.Mode) (SyncResult, error) {
	now := s.Now()
	claim, err := s.Store.ClaimSync(ctx, instance, now, now.Add(s.Timeout+leaseMargin))
	if err != nil {
		return SyncResult{}, fmt.Errorf("claiming %s: %w", instance, err)
	}
	b := batchFor(&claim, mode, s.NewBatchID)
	if claim.Open == nil || *claim.Open != b {
		if err := s.Store.StartBatch(ctx, instance, b, now); err != nil {
			return SyncResult{}, fmt.Errorf("starting a batch of %s: %w", instance, err)
		}
	}
	actx, cancel := context.WithTimeout(ctx, s.Timeout)
	res, syncErr := s.Connectors.Sync(actx, instance, b.ID, b.Mode, claim.Watermark)
	cancel()
	ended := s.Now()
	// Record the outcome even when ctx ended during the attempt.
	rctx := context.WithoutCancel(ctx)
	if syncErr != nil {
		failures := claim.Failures + 1
		if err := s.Store.FailSync(rctx, instance, b.ID, failures, s.Policy.NextSync(ended, failures, 0), ended); err != nil {
			return SyncResult{}, errors.Join(syncErr, fmt.Errorf("recording the failure: %w", err))
		}
		s.logger().Info("connector sync attempt failed", "instance", instance, "batch", string(b.ID), "failures", failures)
		return SyncResult{}, syncErr
	}
	next := s.Policy.NextSync(ended, 0, s.Jitter(s.Policy.Jitter))
	if err := s.Store.CompleteSync(rctx, instance, &res.Batch, res.Watermark, next, ended); err != nil {
		return SyncResult{}, fmt.Errorf("recording sync %s of %s: %w", b.ID, instance, err)
	}
	s.logger().Info("connector synced", "instance", instance, "batch", string(b.ID), "mode", b.Mode.String(),
		"transactions", res.Batch.Transactions, "new", res.Batch.Imported.New, "problems", len(res.Batch.Problems))
	return res, nil
}

// batchFor returns the batch an attempt runs: the open one, unless mode
// asks for another mode; otherwise a new batch in mode, or, with a zero
// mode, opening until the instance is seeded and transactions after (D-23).
func batchFor(claim *SyncClaim, mode ingest.Mode, newID func() ingest.BatchID) OpenBatch {
	if claim.Open != nil && (mode == 0 || mode == claim.Open.Mode) {
		return *claim.Open
	}
	if mode == 0 {
		mode = ingest.ModeOpening
		if claim.Seeded {
			mode = ingest.ModeTransactions
		}
	}
	return OpenBatch{ID: newID(), Mode: mode}
}

// logger returns Log, or a logger that discards everything.
func (s *Scheduler) logger() *slog.Logger {
	if s.Log == nil {
		return slog.New(slog.DiscardHandler)
	}
	return s.Log
}
