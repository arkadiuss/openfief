package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"slices"
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
	// watermark, resets its failures, sets its next sync and releases the
	// claim, all at once.
	CompleteSync(ctx context.Context, instance string, b *ingest.Batch, watermark string, next, at time.Time) error
	// FailSync records a failed attempt of the open batch id at at, keeps
	// the batch open, sets the instance's failures and next sync and
	// releases the claim.
	FailSync(ctx context.Context, instance string, id ingest.BatchID, failures int, next, at time.Time) error
	// SyncStates returns the state of every instance that has one.
	SyncStates(ctx context.Context) ([]SyncState, error)
}

// ScopeSource says what market-data connectors are asked to price (D-93).
// PricingScope implements it.
type ScopeSource interface {
	MarketDataScope(ctx context.Context) (ingest.MarketDataScope, error)
}

// Scheduler owns connector instances' syncs (D-49): it syncs every running
// instance when its SyncPolicy says it is due, one sync per instance at a
// time, retries a failed batch with the same batch ID after a backoff, and
// stores the watermark of each completed sync. An instance that is no
// longer listed is no longer synced; its state and data stay (D-58).
//
// The mode follows the ledger (D-23, D-95): an instance whose bound
// accounts are all seeded syncs in transactions mode on its own, one with
// none seeded in opening mode, and one with both, such as a broker account
// mapped after the others were seeded, runs an opening batch and then a
// transactions batch (ingest.PlanSync). An instance bound to no account is
// a market-data connector: it syncs in transactions mode and is asked to
// price what Scope says (D-93).
type Scheduler struct {
	Connectors *Connectors
	Instances  InstanceLister
	Store      SyncStore
	// Scope says what to ask market-data connectors to price; nil asks for
	// nothing.
	Scope  ScopeSource
	Policy ingest.SyncPolicy
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
// mode in the modes its accounts call for, and returns the last batch. It
// fails with ErrSyncRunning while the instance syncs. A failure counts like
// a scheduled one.
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

// attempt runs one sync attempt of instance: it plans the attempt's
// batches, claims the instance, and runs each batch in turn, retrying an
// open batch first. When an opening batch is followed by a transactions
// batch, both ask from the watermark the instance had before them, so the
// seeded accounts miss none of their transactions, and the instance is
// claimed again for the second.
func (s *Scheduler) attempt(ctx context.Context, instance string, mode ingest.Mode) (SyncResult, error) {
	plan, err := s.plan(ctx, instance, mode)
	if err != nil {
		return SyncResult{}, fmt.Errorf("planning a sync of %s: %w", instance, err)
	}
	now := s.Now()
	claim, err := s.Store.ClaimSync(ctx, instance, now, now.Add(s.Timeout+leaseMargin))
	if err != nil {
		return SyncResult{}, fmt.Errorf("claiming %s: %w", instance, err)
	}
	batches := batchesFor(&claim, mode, plan.modes, s.NewBatchID)
	watermark := claim.Watermark
	var res SyncResult
	for i, b := range batches {
		if i > 0 {
			now = s.Now()
			if claim, err = s.Store.ClaimSync(ctx, instance, now, now.Add(s.Timeout+leaseMargin)); err != nil {
				return res, fmt.Errorf("claiming %s: %w", instance, err)
			}
		}
		call := SyncCall{ID: b.ID, Mode: b.Mode, Watermark: watermark, Scope: plan.scope}
		keep := i < len(batches)-1
		if res, err = s.run(ctx, instance, &claim, &call, keep); err != nil {
			return res, err
		}
	}
	return res, nil
}

// syncPlan is what an attempt asks for: the modes of its batches, and what
// a market-data connector is to price.
type syncPlan struct {
	modes []ingest.Mode
	scope ingest.MarketDataScope
}

// plan returns the plan of a sync of instance: mode when the caller asks
// for one, otherwise what its bound accounts' seeding calls for
// (ingest.PlanSync); and the market-data scope for an instance bound to no
// account.
func (s *Scheduler) plan(ctx context.Context, instance string, mode ingest.Mode) (syncPlan, error) {
	bound, err := s.Connectors.Bindings.BoundAccounts(ctx, instance)
	if err != nil {
		return syncPlan{}, fmt.Errorf("connector bindings: %w", err)
	}
	var p syncPlan
	if len(bound) == 0 && s.Scope != nil {
		if p.scope, err = s.Scope.MarketDataScope(ctx); err != nil {
			return syncPlan{}, fmt.Errorf("market-data scope: %w", err)
		}
	}
	if mode != 0 {
		p.modes = []ingest.Mode{mode}
		return p, nil
	}
	goLive, err := s.Connectors.GoLiveDates(ctx, bound)
	if err != nil {
		return syncPlan{}, err
	}
	p.modes = ingest.PlanSync(bound, goLive)
	return p, nil
}

// run runs one batch of an attempt on the claimed instance: it starts the
// batch unless it is the claim's open one, calls the connector and imports
// the delivery within the timeout, and records the outcome with the next
// sync's due time. With keep, the instance keeps its watermark, because
// another batch follows from it.
func (s *Scheduler) run(ctx context.Context, instance string, claim *SyncClaim, call *SyncCall, keep bool) (SyncResult, error) {
	b := OpenBatch{ID: call.ID, Mode: call.Mode}
	if claim.Open == nil || *claim.Open != b {
		if err := s.Store.StartBatch(ctx, instance, b, s.Now()); err != nil {
			return SyncResult{}, fmt.Errorf("starting a batch of %s: %w", instance, err)
		}
	}
	actx, cancel := context.WithTimeout(ctx, s.Timeout)
	res, syncErr := s.Connectors.Sync(actx, instance, call)
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
	watermark := res.Watermark
	if keep {
		watermark = call.Watermark
	}
	next := s.Policy.NextSync(ended, 0, s.Jitter(s.Policy.Jitter))
	if err := s.Store.CompleteSync(rctx, instance, &res.Batch, watermark, next, ended); err != nil {
		return SyncResult{}, fmt.Errorf("recording sync %s of %s: %w", b.ID, instance, err)
	}
	s.logger().Info("connector synced", "instance", instance, "batch", string(b.ID), "mode", b.Mode.String(),
		"transactions", res.Batch.Transactions, "new", res.Batch.Imported.New, "problems", len(res.Batch.Problems))
	return res, nil
}

// batchesFor returns the batches an attempt runs, in order. The open batch
// a failed attempt left is retried first, with its ID and mode, unless mode
// asks for another mode; an open opening batch is followed by a
// transactions batch when the plan has one. Otherwise each planned mode
// gets a new batch.
func batchesFor(claim *SyncClaim, mode ingest.Mode, planned []ingest.Mode, newID func() ingest.BatchID) []OpenBatch {
	if claim.Open != nil && (mode == 0 || mode == claim.Open.Mode) {
		out := []OpenBatch{*claim.Open}
		if claim.Open.Mode == ingest.ModeOpening && slices.Contains(planned, ingest.ModeTransactions) {
			out = append(out, OpenBatch{ID: newID(), Mode: ingest.ModeTransactions})
		}
		return out
	}
	out := make([]OpenBatch, 0, len(planned))
	for _, m := range planned {
		out = append(out, OpenBatch{ID: newID(), Mode: m})
	}
	return out
}

// logger returns Log, or a logger that discards everything.
func (s *Scheduler) logger() *slog.Logger {
	if s.Log == nil {
		return slog.New(slog.DiscardHandler)
	}
	return s.Log
}
