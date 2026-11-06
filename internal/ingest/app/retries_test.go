package app_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"github.com/arkadiuss/openfief/internal/ingest/app"
	"github.com/arkadiuss/openfief/internal/ledger"
)

// recordingTrigger records the instances it was asked to sync, and fails
// with the next error of errs.
type recordingTrigger struct {
	calls []string
	errs  []error
}

func (r *recordingTrigger) Retry(_ context.Context, instance string) error {
	r.calls = append(r.calls, instance)
	if len(r.errs) > 0 {
		err := r.errs[0]
		r.errs = r.errs[1:]
		if err != nil {
			return fmt.Errorf("retrying %s: %w", instance, err)
		}
	}
	return nil
}

// TestRetriesRefuseWithoutTheOwner checks who may retry a sync (D-163,
// D-196): only the owner, only when the authorizer lets the owner change
// every account bound to the instance, and only while syncs run.
func TestRetriesRefuseWithoutTheOwner(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		session string
		denied  []ledger.AccountID
		trigger bool
		want    error
		calls   []string
	}{
		{name: "the owner", session: "owner", trigger: true, calls: []string{"xtb-main"}},
		{name: "not the owner", session: "read", trigger: true, want: app.ErrNotOwner},
		{name: "a bound account the owner may not change", session: "owner", denied: []ledger.AccountID{"acc-2"}, trigger: true, want: app.ErrNotOwner},
		{name: "no syncs run", session: "owner", want: app.ErrSyncsNotRunning},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tr := &recordingTrigger{}
			r := &app.Retries{
				Owners: fakeOwners{err: app.ErrNotOwner, denied: tt.denied}, Bindings: &fakeBindings{accounts: []ledger.AccountID{"acc-1", "acc-2"}},
				Now: t0,
			}
			if tt.trigger {
				r.Trigger = tr
			}
			err := r.RetrySync(t.Context(), "https://ui.example.com", tt.session, "xtb-main")
			if !errors.Is(err, tt.want) || (tt.want == nil && err != nil) {
				t.Errorf("RetrySync = %v, want %v", err, tt.want)
			}
			if diff := cmp.Diff(tt.calls, tr.calls); diff != "" {
				t.Errorf("syncs started (-want +got):\n%s", diff)
			}
		})
	}
}

// TestRetriesRateLimit checks that one instance may be retried once a
// minute, and that a retry refused because a sync runs doesn't count.
func TestRetriesRateLimit(t *testing.T) {
	t.Parallel()
	c := &clock{now: t0()}
	tr := &recordingTrigger{errs: []error{app.ErrSyncRunning}}
	r := &app.Retries{Owners: fakeOwners{}, Bindings: &fakeBindings{}, Trigger: tr, Now: c.Now}
	retry := func() error { return r.RetrySync(t.Context(), "https://ui.example.com", "owner", "xtb-main") }
	steps := []struct {
		advance time.Duration
		want    error
	}{
		{want: app.ErrSyncRunning},
		{},
		{advance: 30 * time.Second, want: app.ErrRetryTooSoon},
		{advance: 30 * time.Second},
	}
	for i, st := range steps {
		c.advance(st.advance)
		if err := retry(); !errors.Is(err, st.want) || (st.want == nil && err != nil) {
			t.Errorf("retry %d = %v, want %v", i, err, st.want)
		}
	}
	if err := r.RetrySync(t.Context(), "https://ui.example.com", "owner", "nbp"); err != nil {
		t.Errorf("another instance's retry = %v, want it started", err)
	}
	if diff := cmp.Diff([]string{"xtb-main", "xtb-main", "xtb-main", "nbp"}, tr.calls); diff != "" {
		t.Errorf("syncs asked for (-want +got):\n%s", diff)
	}
}

// TestSchedulerRetry checks the scheduler's retry (D-196): an instance not
// listed is refused, a listed one syncs in the background and is recorded
// like a scheduled sync, and a retry while it syncs is refused.
func TestSchedulerRetry(t *testing.T) {
	t.Parallel()
	caller := &scriptedCaller{block: true, started: make(chan struct{}, 1), release: make(chan struct{})}
	store := newMemStore()
	s, _ := scheduler(t, caller, store, "xtb-main")
	if err := s.Retry(t.Context(), "nbp"); !errors.Is(err, app.ErrNoInstance) {
		t.Errorf("Retry of an unlisted instance = %v, want ErrNoInstance", err)
	}
	if err := s.Retry(t.Context(), "xtb-main"); err != nil {
		t.Fatalf("Retry: %v", err)
	}
	<-caller.started
	if err := s.Retry(t.Context(), "xtb-main"); !errors.Is(err, app.ErrSyncRunning) {
		t.Errorf("Retry while syncing = %v, want ErrSyncRunning", err)
	}
	close(caller.release)
	s.Wait()
	if diff := cmp.Diff([]string{"b-1:completed"}, store.recorded()); diff != "" {
		t.Errorf("outcomes (-want +got):\n%s", diff)
	}
}

// TestSchedulerRetryRespectsAnotherClaim checks that a retry is refused
// while another core's claim holds the instance (D-49), and goes ahead once
// it expires.
func TestSchedulerRetryRespectsAnotherClaim(t *testing.T) {
	t.Parallel()
	store := newMemStore()
	s, c := scheduler(t, &scriptedCaller{}, store, "xtb-main")
	if _, err := store.ClaimSync(t.Context(), "xtb-main", c.Now(), c.Now().Add(time.Hour)); err != nil {
		t.Fatalf("ClaimSync: %v", err)
	}
	if err := s.Retry(t.Context(), "xtb-main"); !errors.Is(err, app.ErrSyncRunning) {
		t.Errorf("Retry under another claim = %v, want ErrSyncRunning", err)
	}
	c.advance(2 * time.Hour)
	if err := s.Retry(t.Context(), "xtb-main"); err != nil {
		t.Errorf("Retry after the claim expired = %v", err)
	}
	s.Wait()
	if diff := cmp.Diff([]string{"b-1:completed"}, store.recorded()); diff != "" {
		t.Errorf("outcomes (-want +got):\n%s", diff)
	}
}
