package main

import (
	"context"
	"testing"

	"connectrpc.com/connect"

	ledgerv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/ledger/v1"
	"github.com/arkadiuss/openfief/sdk/go/gen/openfief/ledger/v1/ledgerv1connect"
)

// TestInstrumentService checks that the core's plugin port serves the
// owner's renames of instruments (D-197, D-79): the owner's view stores a
// name, while another caller, an unknown
// session and a malformed ISIN are refused.
func TestInstrumentService(t *testing.T) {
	t.Parallel()
	oc := ownerSession(t)
	bearer := func(token string) connect.Option {
		return connect.WithInterceptors(connect.UnaryInterceptorFunc(func(next connect.UnaryFunc) connect.UnaryFunc {
			return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
				req.Header().Set("Authorization", "Bearer "+token)
				return next(ctx, req)
			}
		}))
	}
	tests := []struct {
		name, caller, session, isin string
		want                        connect.Code
	}{
		{"the owner's view", uiCaller, oc.session, "PLTEST000010", 0},
		{"not a view plugin", "wrong", oc.session, "PLTEST000010", connect.CodePermissionDenied},
		{"an unknown session", uiCaller, "ofsess_nope_nope", "PLTEST000010", connect.CodeUnauthenticated},
		{"a malformed ISIN", uiCaller, oc.session, "PL1", connect.CodeInvalidArgument},
	}
	for _, tt := range tests {
		c := ledgerv1connect.NewInstrumentServiceClient(oc.srv.Client(), oc.srv.URL, bearer(tt.caller))
		_, err := c.RenameInstrument(t.Context(), connect.NewRequest(&ledgerv1.RenameInstrumentRequest{
			Origin: localUI, SessionToken: tt.session, Isin: tt.isin, Name: "Synthetic fund",
		}))
		if got := connect.CodeOf(err); (tt.want == 0 && err != nil) || (tt.want != 0 && got != tt.want) {
			t.Errorf("%s: RenameInstrument = %v, want %s", tt.name, err, tt.want)
		}
	}
}
