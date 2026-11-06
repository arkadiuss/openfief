package connectapi_test

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http/httptest"
	"testing"
	"time"

	"connectrpc.com/connect"

	"github.com/arkadiuss/openfief/internal/ingest/app"
	"github.com/arkadiuss/openfief/internal/ingest/connectapi"
	ingestv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/ingest/v1"
	"github.com/arkadiuss/openfief/sdk/go/gen/openfief/ingest/v1/ingestv1connect"
)

// errTriggerDown is a trigger that fails for no reason the owner can act on.
var errTriggerDown = errors.New("synthetic: trigger down")

// trigger starts every retry, or fails with err.
type trigger struct{ err error }

func (t trigger) Retry(_ context.Context, instance string) error {
	if t.err != nil {
		return fmt.Errorf("retrying %s: %w", instance, t.err)
	}
	return nil
}

// TestSyncServiceRetry checks RetrySync's answers over Connect (D-196):
// the owner's retry starts, and each refusal has the code its contract
// names.
func TestSyncServiceRetry(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		caller  string
		session string
		trigger app.SyncTrigger
		want    connect.Code
	}{
		{name: "the owner's retry", trigger: trigger{}},
		{name: "not a view plugin", caller: "other", trigger: trigger{}, want: connect.CodePermissionDenied},
		{name: "a read session", session: readSession, trigger: trigger{}, want: connect.CodePermissionDenied},
		{name: "no session", session: "unknown", trigger: trigger{}, want: connect.CodeUnauthenticated},
		{name: "no such instance", trigger: trigger{err: app.ErrNoInstance}, want: connect.CodeNotFound},
		{name: "a sync runs", trigger: trigger{err: app.ErrSyncRunning}, want: connect.CodeFailedPrecondition},
		{name: "no scheduler", want: connect.CodeUnavailable},
		{name: "anything else", trigger: trigger{err: errTriggerDown}, want: connect.CodeInternal},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			r := &app.Retries{
				Owners: owners{}, Bindings: bindings{}, Trigger: tt.trigger,
				Now: func() time.Time { return time.Date(2026, time.March, 2, 6, 0, 0, 0, time.UTC) },
			}
			_, handler := connectapi.NewSyncs(r, viewOnly(), slog.New(slog.DiscardHandler))
			srv := httptest.NewServer(handler)
			t.Cleanup(srv.Close)
			caller, session := viewCaller, ownerSession
			if tt.caller != "" {
				caller = tt.caller
			}
			if tt.session != "" {
				session = tt.session
			}
			c := ingestv1connect.NewSyncServiceClient(srv.Client(), srv.URL, connect.WithInterceptors(bearer(caller)))
			_, err := c.RetrySync(t.Context(), connect.NewRequest(&ingestv1.RetrySyncRequest{
				Origin: local, SessionToken: session, Instance: "bonds-main",
			}))
			if got := connect.CodeOf(err); err != nil && got != tt.want || err == nil && tt.want != 0 {
				t.Errorf("RetrySync = %v, want code %v", err, tt.want)
			}
		})
	}
}

// TestSyncServiceRetryTooSoon checks that a second retry of one instance
// within the minute is RESOURCE_EXHAUSTED.
func TestSyncServiceRetryTooSoon(t *testing.T) {
	t.Parallel()
	r := &app.Retries{
		Owners: owners{}, Bindings: bindings{}, Trigger: trigger{},
		Now: func() time.Time { return time.Date(2026, time.March, 2, 6, 0, 0, 0, time.UTC) },
	}
	_, handler := connectapi.NewSyncs(r, viewOnly(), slog.New(slog.DiscardHandler))
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	c := ingestv1connect.NewSyncServiceClient(srv.Client(), srv.URL, connect.WithInterceptors(bearer(viewCaller)))
	req := &ingestv1.RetrySyncRequest{Origin: local, SessionToken: ownerSession, Instance: "bonds-main"}
	if _, err := c.RetrySync(t.Context(), connect.NewRequest(req)); err != nil {
		t.Fatalf("first RetrySync: %v", err)
	}
	if _, err := c.RetrySync(t.Context(), connect.NewRequest(req)); connect.CodeOf(err) != connect.CodeResourceExhausted {
		t.Errorf("second RetrySync = %v, want RESOURCE_EXHAUSTED", err)
	}
}
