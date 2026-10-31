package connectapi_test

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http/httptest"
	"testing"

	"connectrpc.com/connect"
	gocmp "github.com/google/go-cmp/cmp"
	"google.golang.org/protobuf/testing/protocmp"

	"github.com/arkadiuss/openfief/internal/ledger"
	"github.com/arkadiuss/openfief/internal/ledger/app"
	"github.com/arkadiuss/openfief/internal/ledger/connectapi"
	ledgerv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/ledger/v1"
	"github.com/arkadiuss/openfief/sdk/go/gen/openfief/ledger/v1/ledgerv1connect"
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

// decider answers each decision with version 1 of it, or fails with err.
type decider struct{ err error }

func (d decider) DecideTransfer(_ context.Context, pair ledger.TransferPair, decision ledger.TransferDecision) (ledger.TransferLink, error) {
	if d.err != nil {
		return ledger.TransferLink{}, fmt.Errorf("deciding: %w", d.err)
	}
	return ledger.TransferLink{Pair: pair, Decision: decision, Version: 1}, nil
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

// client is the generated client, as a struct so helpers return no
// interface.
type client struct {
	ledgerv1connect.TransferServiceClient
}

// newClient serves the TransferService over d and returns a client calling
// it with the caller token.
func newClient(t *testing.T, d decider, caller string) client {
	t.Helper()
	_, handler := connectapi.NewTransfers(&app.Transfers{Owners: owners{}, Ledger: d}, viewOnly(), slog.New(slog.DiscardHandler))
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	return client{ledgerv1connect.NewTransferServiceClient(srv.Client(), srv.URL, connect.WithInterceptors(bearer(caller)))}
}

func TestTransferService(t *testing.T) {
	t.Parallel()
	w := &ledgerv1.TransactionKey{Source: "test", ExternalId: "w1"}
	d := &ledgerv1.TransactionKey{Source: "test", ExternalId: "d1"}
	ctx := t.Context()
	c := newClient(t, decider{}, viewCaller)
	confirmed, err := c.ConfirmTransfer(ctx, connect.NewRequest(&ledgerv1.ConfirmTransferRequest{
		Origin: local, SessionToken: ownerSession, Withdrawal: w, Deposit: d,
	}))
	want := &ledgerv1.TransferLink{Withdrawal: w, Deposit: d, Decision: ledgerv1.TransferDecision_TRANSFER_DECISION_CONFIRMED, Version: 1}
	if err != nil || !gocmp.Equal(want, confirmed.Msg.GetLink(), protocmp.Transform()) {
		t.Fatalf("ConfirmTransfer = %v, %v; want %v", confirmed, err, want)
	}
	rejected, err := c.RejectTransfer(ctx, connect.NewRequest(&ledgerv1.RejectTransferRequest{
		Origin: local, SessionToken: ownerSession, Withdrawal: w, Deposit: d,
	}))
	want.Decision = ledgerv1.TransferDecision_TRANSFER_DECISION_REJECTED
	if err != nil || !gocmp.Equal(want, rejected.Msg.GetLink(), protocmp.Transform()) {
		t.Fatalf("RejectTransfer = %v, %v; want %v", rejected, err, want)
	}
}

func TestTransferServiceErrors(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name            string
		caller, session string
		err             error
		want            connect.Code
	}{
		{"not a view plugin", "other", ownerSession, nil, connect.CodePermissionDenied},
		{"no session", viewCaller, "nope", nil, connect.CodeUnauthenticated},
		{"a read session", viewCaller, readSession, nil, connect.CodePermissionDenied},
		{"not a transfer", viewCaller, ownerSession, ledger.ErrInvalidTransfer, connect.CodeInvalidArgument},
		{"unknown leg", viewCaller, ownerSession, app.ErrUnknownTransaction, connect.CodeNotFound},
		{"taken leg", viewCaller, ownerSession, ledger.ErrTransferTaken, connect.CodeFailedPrecondition},
		{"store down", viewCaller, ownerSession, errDown, connect.CodeInternal},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			c := newClient(t, decider{err: tt.err}, tt.caller)
			_, err := c.ConfirmTransfer(t.Context(), connect.NewRequest(&ledgerv1.ConfirmTransferRequest{
				Origin: local, SessionToken: tt.session,
				Withdrawal: &ledgerv1.TransactionKey{Source: "test", ExternalId: "w1"},
				Deposit:    &ledgerv1.TransactionKey{Source: "test", ExternalId: "d1"},
			}))
			if got := connect.CodeOf(err); got != tt.want {
				t.Errorf("ConfirmTransfer error = %v (%v), want %v", got, err, tt.want)
			}
		})
	}
}
