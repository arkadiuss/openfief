package app_test

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"sync"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/arkadiuss/openfief/internal/ingest"
	"github.com/arkadiuss/openfief/internal/ingest/app"
	"github.com/arkadiuss/openfief/internal/ledger"
)

var errList = errors.New("listing failed")

// memInstance is one instance's state in memStore.
type memInstance struct {
	claim      app.SyncClaim
	next       time.Time
	leaseUntil time.Time
	hasState   bool
}

// memStore is app.SyncStore in memory. It records every batch's outcome.
type memStore struct {
	mu        sync.Mutex
	instances map[string]*memInstance
	// outcomes lists "<batch>:<outcome>" in order.
	outcomes []string
	// claimErr, when set, fails every claim.
	claimErr error
}

func newMemStore() *memStore { return &memStore{instances: map[string]*memInstance{}} }

func (m *memStore) ClaimSync(_ context.Context, instance string, now, until time.Time) (app.SyncClaim, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.claimErr != nil {
		return app.SyncClaim{}, m.claimErr
	}
	in := m.get(instance)
	if now.Before(in.leaseUntil) {
		return app.SyncClaim{}, app.ErrSyncRunning
	}
	in.leaseUntil, in.hasState = until, true
	c := in.claim
	if c.Open != nil {
		open := *c.Open
		c.Open = &open
	}
	return c, nil
}

func (m *memStore) StartBatch(_ context.Context, instance string, b app.OpenBatch, _ time.Time) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	in := m.get(instance)
	if in.claim.Open != nil {
		m.outcomes = append(m.outcomes, string(in.claim.Open.ID)+":abandoned")
	}
	in.claim.Open = &b
	return nil
}

func (m *memStore) CompleteSync(_ context.Context, instance string, b *ingest.Batch, watermark string, next, _ time.Time) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	in := m.get(instance)
	in.claim.Seeded = in.claim.Seeded || b.Mode == ingest.ModeOpening
	in.claim.Watermark, in.claim.Failures, in.claim.Open = watermark, 0, nil
	in.next, in.leaseUntil = next, time.Time{}
	m.outcomes = append(m.outcomes, string(b.ID)+":completed")
	return nil
}

func (m *memStore) FailSync(_ context.Context, instance string, id ingest.BatchID, failures int, next, _ time.Time) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	in := m.get(instance)
	in.claim.Failures, in.next, in.leaseUntil = failures, next, time.Time{}
	m.outcomes = append(m.outcomes, fmt.Sprintf("%s:failed %d", id, failures))
	return nil
}

func (m *memStore) SyncStates(context.Context) ([]app.SyncState, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var out []app.SyncState
	for name, in := range m.instances {
		if in.hasState {
			out = append(out, app.SyncState{Instance: name, NextSync: in.next})
		}
	}
	return out, nil
}

func (m *memStore) get(instance string) *memInstance {
	in, ok := m.instances[instance]
	if !ok {
		in = &memInstance{}
		m.instances[instance] = in
	}
	return in
}

// xtbState returns xtb-main's state and next sync.
func (m *memStore) xtbState() (app.SyncClaim, time.Time) {
	m.mu.Lock()
	defer m.mu.Unlock()
	in := m.get("xtb-main")
	return in.claim, in.next
}

func (m *memStore) recorded() []string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return append([]string(nil), m.outcomes...)
}

// listed lists fixed instances, or fails with err.
type listed struct {
	names []string
	err   error
}

func (l listed) ConnectorInstances(context.Context) ([]string, error) { return l.names, l.err }

// syncCall is one Sync call a scriptedCaller saw.
type syncCall struct {
	Instance  string
	ID        ingest.BatchID
	Mode      ingest.Mode
	Watermark string
}

// scriptedCaller answers Sync with a delivery whose watermark names the
// batch, or with the next error of errs; with block set, it waits for
// release or the call's context first.
type scriptedCaller struct {
	mu      sync.Mutex
	calls   []syncCall
	errs    []error
	block   bool
	started chan struct{}
	release chan struct{}
	base    app.Delivery
}

func (f *scriptedCaller) Sync(ctx context.Context, instance string, id ingest.BatchID, mode ingest.Mode, watermark string) (app.Delivery, error) {
	f.mu.Lock()
	f.calls = append(f.calls, syncCall{instance, id, mode, watermark})
	var err error
	if len(f.errs) > 0 {
		err, f.errs = f.errs[0], f.errs[1:]
	}
	f.mu.Unlock()
	if f.block {
		f.started <- struct{}{}
		select {
		case <-f.release:
		case <-ctx.Done():
			return app.Delivery{}, fmt.Errorf("waiting: %w", ctx.Err())
		}
	}
	if err != nil {
		return app.Delivery{}, err
	}
	d := f.base
	d.Watermark = "after-" + string(id)
	return d, nil
}

func (*scriptedCaller) ImportFile(context.Context, string, ingest.BatchID, ingest.Mode, app.NamedFile) (app.Delivery, error) {
	return app.Delivery{}, errCaller
}

func (*scriptedCaller) ImportBondLots(context.Context, string, ingest.BatchID, []ingest.BondLot) (app.Delivery, error) {
	return app.Delivery{}, errCaller
}

func (f *scriptedCaller) seen() []syncCall {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]syncCall(nil), f.calls...)
}

// clock is a settable clock.
type clock struct {
	mu  sync.Mutex
	now time.Time
}

func (c *clock) Now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.now
}

func (c *clock) advance(d time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.now = c.now.Add(d)
}

// t0 is when the scheduler tests begin.
func t0() time.Time { return time.Date(2026, time.March, 2, 6, 0, 0, 0, time.UTC) }

// scheduler returns a scheduler over the fakes with a six-hour interval,
// ten minutes of jitter on every success, retries after 5 minutes doubling
// up to an hour, and batch IDs b-1, b-2 and so on.
func scheduler(t *testing.T, caller *scriptedCaller, store *memStore, names ...string) (*app.Scheduler, *clock) {
	t.Helper()
	c := &clock{now: t0()}
	caller.base = app.Delivery{Institution: "xtb"}
	var ids int
	var idMu sync.Mutex
	return &app.Scheduler{
		Connectors: connectors(t, caller, &lockedLedger{}, &fakeStores{}),
		Instances:  listed{names: names},
		Store:      store,
		Policy:     ingest.SyncPolicy{Interval: 6 * time.Hour, Jitter: 30 * time.Minute, RetryFirst: 5 * time.Minute, RetryMax: time.Hour},
		Timeout:    time.Minute,
		Now:        c.Now,
		NewBatchID: func() ingest.BatchID {
			idMu.Lock()
			defer idMu.Unlock()
			ids++
			return ingest.BatchID(fmt.Sprintf("b-%d", ids))
		},
		Jitter: func(time.Duration) time.Duration { return 10 * time.Minute },
	}, c
}

// tick runs one tick and waits for the syncs it started.
func tick(t *testing.T, s *app.Scheduler) {
	t.Helper()
	if err := s.Tick(t.Context()); err != nil {
		t.Fatalf("Tick: %v", err)
	}
	s.Wait()
}

func TestSchedulerSyncsWhenDue(t *testing.T) {
	t.Parallel()
	caller, store := &scriptedCaller{}, newMemStore()
	s, c := scheduler(t, caller, store, "xtb-main")

	tick(t, s) // never synced: due at once, opening until seeded
	c.advance(6 * time.Hour)
	tick(t, s) // 6h is before 6h10m: not due
	c.advance(10 * time.Minute)
	tick(t, s) // due: transactions from the stored watermark

	want := []syncCall{
		{Instance: "xtb-main", ID: "b-1", Mode: ingest.ModeOpening},
		{Instance: "xtb-main", ID: "b-2", Mode: ingest.ModeTransactions, Watermark: "after-b-1"},
	}
	if diff := cmp.Diff(want, caller.seen()); diff != "" {
		t.Errorf("Sync calls (-want +got):\n%s", diff)
	}
	claim, next := store.xtbState()
	if !claim.Seeded || claim.Watermark != "after-b-2" || claim.Open != nil {
		t.Errorf("state = %+v, want seeded, watermark after-b-2, no open batch", claim)
	}
	if want := t0().Add(12*time.Hour + 20*time.Minute); !next.Equal(want) {
		t.Errorf("next sync = %s, want %s", next, want)
	}
}

func TestSchedulerRetriesWithTheSameBatch(t *testing.T) {
	t.Parallel()
	caller, store := &scriptedCaller{errs: []error{errCaller, errLedger}}, newMemStore()
	s, c := scheduler(t, caller, store, "xtb-main")

	tick(t, s) // fails: retry after 5m
	c.advance(4 * time.Minute)
	tick(t, s) // not due yet
	c.advance(time.Minute)
	tick(t, s) // fails again: retry after 10m
	_, next := store.xtbState()
	if want := t0().Add(15 * time.Minute); !next.Equal(want) {
		t.Errorf("next sync after two failures = %s, want %s", next, want)
	}
	c.advance(10 * time.Minute)
	tick(t, s) // succeeds

	first := syncCall{Instance: "xtb-main", ID: "b-1", Mode: ingest.ModeOpening}
	if diff := cmp.Diff([]syncCall{first, first, first}, caller.seen()); diff != "" {
		t.Errorf("Sync calls (-want +got):\n%s", diff)
	}
	if diff := cmp.Diff([]string{"b-1:failed 1", "b-1:failed 2", "b-1:completed"}, store.recorded()); diff != "" {
		t.Errorf("outcomes (-want +got):\n%s", diff)
	}
	if claim, _ := store.xtbState(); claim.Failures != 0 || claim.Watermark != "after-b-1" {
		t.Errorf("state after the success = %+v", claim)
	}
}

func TestSchedulerSyncsOnlyListedInstances(t *testing.T) {
	t.Parallel()
	caller, store := &scriptedCaller{}, newMemStore()
	s, c := scheduler(t, caller, store, "nbp", "xtb-main")
	tick(t, s)
	// xtb-main is uninstalled: it is no longer listed, and its state stays.
	s.Instances = listed{names: []string{"nbp"}}
	c.advance(7 * time.Hour)
	tick(t, s)
	got := map[string]int{}
	for _, call := range caller.seen() {
		got[call.Instance]++
	}
	if diff := cmp.Diff(map[string]int{"nbp": 2, "xtb-main": 1}, got); diff != "" {
		t.Errorf("syncs per instance (-want +got):\n%s", diff)
	}
	if claim, _ := store.xtbState(); claim.Watermark == "" {
		t.Error("the uninstalled instance's watermark was dropped")
	}
}

func TestSchedulerSingleFlight(t *testing.T) {
	t.Parallel()
	caller := &scriptedCaller{block: true, started: make(chan struct{}), release: make(chan struct{})}
	store := newMemStore()
	s, _ := scheduler(t, caller, store, "xtb-main")
	if err := s.Tick(t.Context()); err != nil {
		t.Fatalf("Tick: %v", err)
	}
	<-caller.started
	if err := s.Tick(t.Context()); err != nil { // the running sync is not started again
		t.Fatalf("Tick: %v", err)
	}
	if _, err := s.SyncNow(t.Context(), "xtb-main", 0); !errors.Is(err, app.ErrSyncRunning) {
		t.Errorf("SyncNow while syncing = %v, want ErrSyncRunning", err)
	}
	close(caller.release)
	s.Wait()
	if n := len(caller.seen()); n != 1 {
		t.Errorf("connector called %d times, want 1", n)
	}
}

func TestSchedulerSyncNow(t *testing.T) {
	t.Parallel()
	caller, store := &scriptedCaller{errs: []error{errCaller}}, newMemStore()
	s, _ := scheduler(t, caller, store, "xtb-main")
	if _, err := s.SyncNow(t.Context(), "xtb-main", 0); !errors.Is(err, errCaller) {
		t.Fatalf("first SyncNow = %v, want the connector's error", err)
	}
	// The owner asks for a transactions batch: the open opening batch is abandoned.
	got, err := s.SyncNow(t.Context(), "xtb-main", ingest.ModeTransactions)
	if err != nil {
		t.Fatalf("SyncNow: %v", err)
	}
	if got.Batch.ID != "b-2" || got.Batch.Mode != ingest.ModeTransactions || got.Watermark != "after-b-2" {
		t.Errorf("SyncNow = %+v", got)
	}
	if diff := cmp.Diff([]string{"b-1:failed 1", "b-1:abandoned", "b-2:completed"}, store.recorded()); diff != "" {
		t.Errorf("outcomes (-want +got):\n%s", diff)
	}
	if claim, _ := store.xtbState(); claim.Seeded {
		t.Error("a transactions batch marked the instance seeded")
	}
}

func TestSchedulerTimeout(t *testing.T) {
	t.Parallel()
	caller := &scriptedCaller{block: true, started: make(chan struct{}, 1), release: make(chan struct{})}
	store := newMemStore()
	s, _ := scheduler(t, caller, store, "xtb-main")
	s.Timeout = time.Nanosecond
	_, err := s.SyncNow(t.Context(), "xtb-main", 0)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("SyncNow = %v, want a deadline error", err)
	}
	if diff := cmp.Diff([]string{"b-1:failed 1"}, store.recorded()); diff != "" {
		t.Errorf("outcomes (-want +got):\n%s", diff)
	}
}

func TestSchedulerFailures(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		list     error
		claimErr error
		wantErr  error
	}{
		{name: "listing fails", list: errList, wantErr: errList},
		{name: "another replica holds the claim", claimErr: app.ErrSyncRunning},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			caller, store := &scriptedCaller{}, newMemStore()
			store.claimErr = tt.claimErr
			s, _ := scheduler(t, caller, store, "xtb-main")
			s.Instances = listed{names: []string{"xtb-main"}, err: tt.list}
			err := s.Tick(t.Context())
			s.Wait()
			if diff := cmp.Diff(tt.wantErr, err, cmpopts.EquateErrors()); diff != "" {
				t.Errorf("Tick error (-want +got):\n%s", diff)
			}
			if n := len(caller.seen()); n != 0 {
				t.Errorf("connector called %d times, want 0", n)
			}
		})
	}
}

func TestSchedulerRunStopsWithItsContext(t *testing.T) {
	t.Parallel()
	caller, store := &scriptedCaller{}, newMemStore()
	s, _ := scheduler(t, caller, store, "xtb-main")
	ctx, cancel := context.WithCancel(t.Context())
	ticks := make(chan time.Time)
	done := make(chan struct{})
	go func() {
		s.Run(ctx, ticks)
		close(done)
	}()
	ticks <- t0() // Run ticked once at start and now waits for a tick
	cancel()
	<-done
	if diff := cmp.Diff([]string{"b-1:completed"}, store.recorded()); diff != "" {
		t.Errorf("outcomes (-want +got):\n%s", diff)
	}
}

func TestBind(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		instance string
		account  ledger.AccountID
		wantErr  error
		want     []ledger.AccountID
	}{
		{name: "stored account", instance: "xtb-main", account: "acc-1", want: []ledger.AccountID{"acc-1"}},
		{name: "unknown account", instance: "xtb-main", account: "acc-9", wantErr: errUnknownAccount},
		{name: "invalid instance name", instance: "XTB", account: "acc-1", wantErr: ingest.ErrInvalidInstanceName},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			store := &memBindings{}
			err := app.Bind(t.Context(), registry{"acc-1"}, store, tt.instance, tt.account, t0())
			if diff := cmp.Diff(tt.wantErr, err, cmpopts.EquateErrors()); diff != "" {
				t.Errorf("Bind error (-want +got):\n%s", diff)
			}
			got, err := store.BoundAccounts(t.Context(), tt.instance)
			if err != nil {
				t.Fatalf("BoundAccounts: %v", err)
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("bound (-want +got):\n%s", diff)
			}
		})
	}
}

var errUnknownAccount = errors.New("unknown account")

// registry stores the listed accounts.
type registry []ledger.AccountID

func (r registry) Account(_ context.Context, id ledger.AccountID) (ledger.Account, error) {
	if slices.Contains(r, id) {
		return ledger.Account{ID: id}, nil
	}
	return ledger.Account{}, errUnknownAccount
}

// memBindings is app.BindingStore in memory, for one instance.
type memBindings struct {
	accounts []ledger.AccountID
}

func (m *memBindings) Bind(_ context.Context, _ string, account ledger.AccountID, _ time.Time) error {
	m.accounts = append(m.accounts, account)
	return nil
}

func (m *memBindings) BoundAccounts(context.Context, string) ([]ledger.AccountID, error) {
	return m.accounts, nil
}

// lockedLedger is a fakeLedger that concurrent syncs may share.
type lockedLedger struct {
	mu sync.Mutex
	l  fakeLedger
}

func (f *lockedLedger) Import(ctx context.Context, txs []ledger.Transaction, assertions []ledger.BalanceAssertion) (ingest.ImportSummary, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.l.Import(ctx, txs, assertions)
}

func (f *lockedLedger) Seed(ctx context.Context, txs []ledger.Transaction, assertions []ledger.BalanceAssertion) (ingest.ImportSummary, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.l.Seed(ctx, txs, assertions)
}
