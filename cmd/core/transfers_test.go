package main

import (
	"context"
	"net/http/httptest"
	"testing"

	"connectrpc.com/connect"

	ledgerv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/ledger/v1"
	"github.com/arkadiuss/openfief/sdk/go/gen/openfief/ledger/v1/ledgerv1connect"
)

// transferAPI is a generated TransferService client.
type transferAPI struct {
	ledgerv1connect.TransferServiceClient
}

// transferClient returns a TransferService client of srv that presents
// token as the view plugin's.
func transferClient(srv *httptest.Server, token string) transferAPI {
	bearer := connect.UnaryInterceptorFunc(func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			req.Header().Set("Authorization", "Bearer "+token)
			return next(ctx, req)
		}
	})
	return transferAPI{ledgerv1connect.NewTransferServiceClient(srv.Client(), srv.URL, connect.WithInterceptors(bearer))}
}

// TestTransferService checks that the core's plugin port serves the
// owner's transfer decisions (D-105, D-79): the owner's view reaches the
// ledger, which answers that legs it doesn't hold are not found, while
// another caller and an unknown session are refused before it.
func TestTransferService(t *testing.T) {
	t.Parallel()
	oc := ownerSession(t)
	tests := []struct {
		name, caller, session string
		want                  connect.Code
	}{
		{"the owner's view, legs not stored", uiCaller, oc.session, connect.CodeNotFound},
		{"not a view plugin", "wrong", oc.session, connect.CodePermissionDenied},
		{"an unknown session", uiCaller, "ofsess_nope_nope", connect.CodeUnauthenticated},
	}
	for _, tt := range tests {
		_, err := transferClient(oc.srv, tt.caller).ConfirmTransfer(t.Context(), connect.NewRequest(&ledgerv1.ConfirmTransferRequest{
			Origin: localUI, SessionToken: tt.session,
			Withdrawal: &ledgerv1.TransactionKey{Source: "test", ExternalId: "w1"},
			Deposit:    &ledgerv1.TransactionKey{Source: "test", ExternalId: "d1"},
		}))
		if connect.CodeOf(err) != tt.want {
			t.Errorf("%s: ConfirmTransfer = %v, want %s", tt.name, err, tt.want)
		}
	}
}
