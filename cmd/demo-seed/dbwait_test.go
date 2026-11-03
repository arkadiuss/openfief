package main

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

var (
	errRefused  = errors.New("connection refused")
	errCanceled = errors.New("canceled")
)

// fakeDB refuses the first refusals pings, then accepts, and records the
// waits between them; sleepErr fails every wait.
type fakeDB struct {
	refusals int
	pings    int
	slept    []time.Duration
	sleepErr error
}

func (f *fakeDB) ping(context.Context) error {
	f.pings++
	if f.pings <= f.refusals {
		return errRefused
	}
	return nil
}

func (f *fakeDB) sleep(_ context.Context, d time.Duration) error {
	f.slept = append(f.slept, d)
	return f.sleepErr
}

func TestWaitForDatabase(t *testing.T) {
	t.Parallel()
	ms := time.Millisecond
	tests := []struct {
		name      string
		db        fakeDB
		limit     time.Duration
		wantSlept []time.Duration
		wantErr   error
	}{
		{name: "ready at once", limit: time.Minute},
		{
			name: "ready after three refusals, backing off", db: fakeDB{refusals: 3}, limit: time.Minute,
			wantSlept: []time.Duration{500 * ms, time.Second, 2 * time.Second},
		},
		{
			name: "the backoff stops growing at its cap", db: fakeDB{refusals: 6}, limit: time.Minute,
			wantSlept: []time.Duration{500 * ms, time.Second, 2 * time.Second, 4 * time.Second, 5 * time.Second, 5 * time.Second},
		},
		{
			name: "still refused at the limit", db: fakeDB{refusals: 100}, limit: 10 * time.Second,
			wantSlept: []time.Duration{500 * ms, time.Second, 2 * time.Second, 4 * time.Second, 2500 * ms},
			wantErr:   errDatabaseNotReady,
		},
		{
			name: "the wait is cut short", db: fakeDB{refusals: 1, sleepErr: errCanceled}, limit: time.Minute,
			wantSlept: []time.Duration{500 * ms}, wantErr: errCanceled,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			db := tt.db
			err := waitForDatabase(t.Context(), db.ping, db.sleep, tt.limit)
			if diff := cmp.Diff(tt.wantErr, err, cmpopts.EquateErrors()); diff != "" {
				t.Errorf("waitForDatabase error (-want +got):\n%s", diff)
			}
			if diff := cmp.Diff(tt.wantSlept, db.slept); diff != "" {
				t.Errorf("waits (-want +got):\n%s", diff)
			}
		})
	}
}

func TestWaitForDatabaseKeepsTheLastError(t *testing.T) {
	t.Parallel()
	db := fakeDB{refusals: 100}
	err := waitForDatabase(t.Context(), db.ping, db.sleep, time.Second)
	if !errors.Is(err, errRefused) || !errors.Is(err, errDatabaseNotReady) {
		t.Errorf("waitForDatabase = %v, want errDatabaseNotReady wrapping the last refusal", err)
	}
}

func TestSleepCtxStopsWithTheContext(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	if err := sleepCtx(ctx, time.Hour); !errors.Is(err, context.Canceled) {
		t.Errorf("sleepCtx on a canceled context = %v, want context.Canceled", err)
	}
}
