package main

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"connectrpc.com/connect"

	ingestv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/ingest/v1"
	"github.com/arkadiuss/openfief/sdk/go/gen/openfief/ingest/v1/ingestv1connect"
	portfoliov1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/portfolio/v1"
	"github.com/arkadiuss/openfief/sdk/go/gen/openfief/portfolio/v1/portfoliov1connect"
)

// listBrokerAccounts lists the unmapped broker accounts through the
// owner's AccountService, the only place they are listed (D-121), after
// checking that the read API serves none of them.
func listBrokerAccounts(t *testing.T, srv *httptest.Server, session string) *ingestv1.ListBrokerAccountsResponse {
	t.Helper()
	r := get(t, srv, "/v1/broker-accounts", session)
	if r.status != http.StatusNotImplemented || bytes.Contains(r.body, []byte("U000000")) {
		t.Errorf("GET /v1/broker-accounts = %d %s, want 501 naming no broker account (D-121)", r.status, r.body)
	}
	accounts := ingestv1connect.NewAccountServiceClient(srv.Client(), srv.URL, connect.WithInterceptors(viewBearer(uiCaller)))
	resp, err := accounts.ListBrokerAccounts(t.Context(), connect.NewRequest(&ingestv1.ListBrokerAccountsRequest{
		Origin: localUI, SessionToken: session, Unmapped: true,
	}))
	if err != nil {
		t.Fatalf("ListBrokerAccounts: %v", err)
	}
	// A caller that is no view plugin, such as a connector, gets none.
	other := ingestv1connect.NewAccountServiceClient(srv.Client(), srv.URL, connect.WithInterceptors(viewBearer("synthetic-not-a-view")))
	if _, err := other.ListBrokerAccounts(t.Context(), connect.NewRequest(&ingestv1.ListBrokerAccountsRequest{
		Origin: localUI, SessionToken: session,
	})); err == nil {
		t.Error("a caller that is no view plugin listed the broker accounts")
	}
	return resp.Msg
}

// checkRemovals runs D-120 through the real core once ibkr-main holds a
// deposit from U0000001: that account and its sleeve stay, and the broker
// account can't be remapped; an empty account takes U0000002, gives it up
// and is deleted; an unused sleeve and then its kind are deleted.
func checkRemovals(t *testing.T, srv *httptest.Server, session string) {
	t.Helper()
	ctx := t.Context()
	accounts := ingestv1connect.NewAccountServiceClient(srv.Client(), srv.URL, connect.WithInterceptors(viewBearer(uiCaller)))
	sleeves := portfoliov1connect.NewSleeveServiceClient(srv.Client(), srv.URL, connect.WithInterceptors(viewBearer(uiCaller)))
	refused := func(what string, err error) {
		t.Helper()
		if connect.CodeOf(err) != connect.CodeFailedPrecondition {
			t.Errorf("%s = %v, want FailedPrecondition", what, err)
		}
	}
	ok := func(what string, err error) {
		t.Helper()
		if err != nil {
			t.Errorf("%s: %v", what, err)
		}
	}
	remap := func(broker, account string) error {
		_, err := accounts.RemapBrokerAccount(ctx, connect.NewRequest(&ingestv1.RemapBrokerAccountRequest{
			Origin: localUI, SessionToken: session, ConnectorInstance: "ibkr-main", BrokerAccount: broker, AccountId: account,
		}))
		if err != nil {
			return fmt.Errorf("owner service: %w", err)
		}
		return nil
	}
	deleteAccount := func(id string) error {
		_, err := accounts.DeleteAccount(ctx, connect.NewRequest(&ingestv1.DeleteAccountRequest{Origin: localUI, SessionToken: session, AccountId: id}))
		if err != nil {
			return fmt.Errorf("owner service: %w", err)
		}
		return nil
	}
	refused("deleting ibkr-main, which holds a deposit", deleteAccount("ibkr-main"))
	refused("remapping U0000001, booked into ibkr-main", remap("U0000001", ""))
	_, err := accounts.CreateAccount(ctx, connect.NewRequest(&ingestv1.CreateAccountRequest{
		Origin: localUI, SessionToken: session, Name: "Spare", Sleeve: "gold-bars", Institution: "ibkr",
	}))
	ok("CreateAccount", err)
	ok("mapping U0000002 to the empty account", remap("U0000002", "spare"))
	ok("remapping U0000002 away from it", remap("U0000002", ""))
	ok("deleting the empty account", deleteAccount("spare"))
	ok("deleting it again", deleteAccount("spare"))
	_, err = sleeves.DeleteSleeve(ctx, connect.NewRequest(&portfoliov1.DeleteSleeveRequest{Origin: localUI, SessionToken: session, Id: "satellite"}))
	refused("deleting satellite, which holds ibkr-main", err)
	_, err = sleeves.DeleteSleeveKind(ctx, connect.NewRequest(&portfoliov1.DeleteSleeveKindRequest{Origin: localUI, SessionToken: session, Key: "gold"}))
	refused("deleting the kind gold-bars has", err)
	_, err = sleeves.DeleteSleeve(ctx, connect.NewRequest(&portfoliov1.DeleteSleeveRequest{Origin: localUI, SessionToken: session, Id: "gold-bars"}))
	ok("deleting the empty sleeve", err)
	_, err = sleeves.DeleteSleeveKind(ctx, connect.NewRequest(&portfoliov1.DeleteSleeveKindRequest{Origin: localUI, SessionToken: session, Key: "gold"}))
	ok("deleting the kind no sleeve has", err)
	if n := len(listBrokerAccounts(t, srv, session).GetBrokerAccounts()); n != 1 {
		t.Errorf("unmapped broker accounts after the removals = %d, want U0000002 alone", n)
	}
}
