package app_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"github.com/arkadiuss/openfief/internal/ingest"
	"github.com/arkadiuss/openfief/internal/ingest/app"
	"github.com/arkadiuss/openfief/internal/ledger"
)

// errDatabase stands for a store failure that names no known error.
var errDatabase = errors.New("connection reset")

func TestSyncErrorClassOf(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		err  error
		want ingest.SyncErrorClass
	}{
		{name: "no error", err: nil, want: 0},
		{name: "a delivery's transaction past its timeout", err: fmt.Errorf("deliver: %w: %w", app.ErrDeliveryTimeout, errDatabase), want: ingest.SyncErrorDeliveryTimeout},
		{name: "past the timeout", err: fmt.Errorf("syncing: %w", context.DeadlineExceeded), want: ingest.SyncErrorTimeout},
		{name: "a timeout reported as unreachable stays a timeout", err: fmt.Errorf("%w: %w", app.ErrConnectorUnavailable, context.DeadlineExceeded), want: ingest.SyncErrorTimeout},
		{name: "the core stopping", err: fmt.Errorf("waiting: %w", context.Canceled), want: ingest.SyncErrorInterrupted},
		{name: "an instance that can't be reached", err: fmt.Errorf("%w: no endpoint", app.ErrConnectorUnavailable), want: ingest.SyncErrorUnavailable},
		{name: "the connector's own error", err: fmt.Errorf("connector xtb-main: %w: bad credentials", app.ErrConnectorFailed), want: ingest.SyncErrorConnector},
		{name: "an answer that doesn't translate", err: fmt.Errorf("%w: amount", app.ErrInvalidDelivery), want: ingest.SyncErrorInvalidDelivery},
		{name: "an institution the manifest doesn't declare", err: fmt.Errorf("delivery: %w", ingest.ErrUndeclaredInstitution), want: ingest.SyncErrorRefused},
		{name: "an account not bound", err: ingest.ErrAccountNotBound, want: ingest.SyncErrorRefused},
		{name: "a transaction the ledger refuses", err: fmt.Errorf("importing: %w", ledger.ErrInvalidTransaction), want: ingest.SyncErrorBooking},
		{name: "unbalanced postings", err: fmt.Errorf("importing: %w", &ledger.UnbalancedError{Key: ledger.TransactionKey{Source: "s", ExternalID: "1"}}), want: ingest.SyncErrorBooking},
		{name: "seeding outside an opening batch", err: app.ErrSeedingOutsideOpening, want: ingest.SyncErrorBooking},
		{name: "anything else", err: errDatabase, want: ingest.SyncErrorOther},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := app.SyncErrorClassOf(tt.err); got != tt.want {
				t.Errorf("SyncErrorClassOf = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestSchedulerRecordsAttempts checks that every attempt is recorded with
// its batch, start, end, error class, failures in a row and next sync, and
// a success with no class and no failures.
func TestSchedulerRecordsAttempts(t *testing.T) {
	t.Parallel()
	caller, store := &scriptedCaller{errs: []error{fmt.Errorf("%w: down", app.ErrConnectorUnavailable)}}, newMemStore()
	s, c := scheduler(t, caller, store, "xtb-main")

	tick(t, s) // fails: retry after 5m
	c.advance(5 * time.Minute)
	tick(t, s) // succeeds

	retry := t0().Add(5 * time.Minute)
	want := []ingest.SyncAttempt{
		{
			Instance: "xtb-main", Batch: "b-1", Mode: ingest.ModeOpening, Started: t0(), Ended: t0(),
			Class: ingest.SyncErrorUnavailable, Failures: 1, NextSync: retry,
		},
		{
			Instance: "xtb-main", Batch: "b-1", Mode: ingest.ModeOpening, Started: retry, Ended: retry,
			NextSync: retry.Add(6*time.Hour + 10*time.Minute),
		},
	}
	store.mu.Lock()
	got := append([]ingest.SyncAttempt(nil), store.attempts...)
	store.mu.Unlock()
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("attempts (-want +got):\n%s", diff)
	}
}
