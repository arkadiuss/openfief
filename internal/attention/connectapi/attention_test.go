package connectapi_test

import (
	"context"
	"errors"
	"log/slog"
	"net/http/httptest"
	"testing"
	"time"

	"connectrpc.com/connect"

	"github.com/arkadiuss/openfief/internal/attention"
	"github.com/arkadiuss/openfief/internal/attention/app"
	"github.com/arkadiuss/openfief/internal/attention/connectapi"
	attentionv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/attention/v1"
	"github.com/arkadiuss/openfief/sdk/go/gen/openfief/attention/v1/attentionv1connect"
)

// Synthetic tokens and origin; nothing real is implied.
const (
	viewCaller   = "view-plugin-token"
	ownerSession = "ofsess_owner"
	readSession  = "ofsess_read"
	local        = "http://localhost:8080"
)

var (
	errDown    = errors.New("store down")
	errNotView = errors.New("not a view plugin")
)

// owners accepts ownerSession; readSession is not the owner's.
type owners struct{}

func (owners) CheckOwner(_ context.Context, _, token string) error {
	switch token {
	case ownerSession:
		return nil
	case readSession:
		return app.ErrNotOwner
	}
	return app.ErrNoSession
}

// store holds two open items: 1, a missing setting, and 2, import
// problems; it fails every call when err is set.
type store struct {
	items []attention.Item
	err   error
}

func newStore() *store {
	open := []attention.Change{{Seq: 1, State: attention.StateOpen, Reason: attention.ReasonRaised}}
	return &store{items: []attention.Item{
		{ID: 1, Kind: attention.KindSettingMissing, Subject: "marginal-rate", Changes: open},
		{ID: 2, Kind: attention.KindImportProblems, Subject: "b-1", Changes: open},
	}}
}

func (s *store) InTx(_ context.Context, fn func(app.Store) error) error { return fn(s) }
func (s *store) Lock(context.Context) error                             { return s.err }

func (s *store) Items(context.Context, attention.Kind) ([]attention.Item, error) {
	return s.items, s.err
}

func (s *store) Item(_ context.Context, id int64) (attention.Item, bool, error) {
	for _, it := range s.items {
		if it.ID == id {
			return it, true, s.err
		}
	}
	return attention.Item{}, false, s.err
}

func (s *store) Append(_ context.Context, step *attention.Step) (int64, error) {
	s.items[step.ItemID-1].Changes = append(s.items[step.ItemID-1].Changes, step.Change)
	return step.ItemID, s.err
}

// viewOnly admits calls bearing viewCaller, like cmd/core's view-plugin
// check.
func viewOnly() connect.UnaryInterceptorFunc {
	return func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			if req.Header().Get("Authorization") != "Bearer "+viewCaller {
				return nil, connect.NewError(connect.CodePermissionDenied, errNotView)
			}
			return next(ctx, req)
		}
	}
}

// bearer sets the Authorization header of every call.
func bearer(token string) connect.UnaryInterceptorFunc {
	return func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			req.Header().Set("Authorization", "Bearer "+token)
			return next(ctx, req)
		}
	}
}

// outcome is a call's code and error text, both empty when it succeeded.
type outcome struct {
	code connect.Code
	text string
}

// act acknowledges item id through client, or resolves it, for session,
// and returns the call's outcome.
func act(t *testing.T, client attentionv1connect.AttentionServiceClient, resolve bool, session string, id int64) outcome {
	t.Helper()
	var err error
	if resolve {
		_, err = client.ResolveAttentionItem(t.Context(), connect.NewRequest(&attentionv1.ResolveAttentionItemRequest{
			Origin: local, SessionToken: session, ItemId: id,
		}))
	} else {
		_, err = client.AcknowledgeAttentionItem(t.Context(), connect.NewRequest(&attentionv1.AcknowledgeAttentionItemRequest{
			Origin: local, SessionToken: session, ItemId: id,
		}))
	}
	if err == nil {
		return outcome{}
	}
	return outcome{code: connect.CodeOf(err), text: err.Error()}
}

func TestAttentionService(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		caller  string
		session string
		resolve bool
		id      int64
		fail    error
		want    connect.Code
	}{
		{name: "the owner acknowledges", caller: viewCaller, session: ownerSession, id: 1},
		{name: "the owner resolves import problems", caller: viewCaller, session: ownerSession, resolve: true, id: 2},
		{name: "a missing setting resolves only when set", caller: viewCaller, session: ownerSession, resolve: true, id: 1, want: connect.CodeFailedPrecondition},
		{name: "unknown item", caller: viewCaller, session: ownerSession, id: 9, want: connect.CodeNotFound},
		{name: "a read session", caller: viewCaller, session: readSession, id: 1, want: connect.CodePermissionDenied},
		{name: "no session", caller: viewCaller, id: 1, want: connect.CodeUnauthenticated},
		{name: "not a view plugin", caller: "script-token", session: ownerSession, id: 1, want: connect.CodePermissionDenied},
		{name: "the store fails", caller: viewCaller, session: ownerSession, id: 1, fail: errDown, want: connect.CodeInternal},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			s := newStore()
			s.err = tt.fail
			items := app.New(s, owners{}, func() time.Time { return time.Date(2026, 3, 2, 0, 0, 0, 0, time.UTC) })
			path, handler := connectapi.NewAttention(items, viewOnly(), slog.New(slog.DiscardHandler))
			srv := httptest.NewServer(handler)
			t.Cleanup(srv.Close)
			if path != attentionv1connect.AttentionServiceName+"/" && path != "/"+attentionv1connect.AttentionServiceName+"/" {
				t.Errorf("path = %q", path)
			}
			client := attentionv1connect.NewAttentionServiceClient(srv.Client(), srv.URL, connect.WithInterceptors(bearer(tt.caller)))
			got := act(t, client, tt.resolve, tt.session, tt.id)
			if got.code != tt.want {
				t.Fatalf("error %q (code %s), want code %s", got.text, got.code, tt.want)
			}
			if tt.want == connect.CodeInternal && got.text != "internal: internal error" {
				t.Errorf("internal error leaks %q", got.text)
			}
		})
	}
}
