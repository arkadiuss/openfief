package main

import (
	"context"
	"fmt"
	"net/http/httptest"
	"strings"
	"testing"

	"connectrpc.com/connect"

	connectorv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/connector/v1"
	"github.com/arkadiuss/openfief/sdk/go/gen/openfief/connector/v1/connectorv1connect"
	ingestv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/ingest/v1"
	"github.com/arkadiuss/openfief/sdk/go/gen/openfief/ingest/v1/ingestv1connect"
	portfoliov1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/portfolio/v1"
	"github.com/arkadiuss/openfief/sdk/go/gen/openfief/portfolio/v1/portfoliov1connect"
	viewsv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/views/v1"
)

// brokerConnector serves the connector contract as a broker connector that
// knows no OpenFief account (D-72): every Sync delivers one deposit named
// by the synthetic broker account U0000001 and reports U0000002 too.
type brokerConnector struct {
	connectorv1connect.UnimplementedConnectorServiceHandler
}

func (brokerConnector) Sync(_ context.Context, req *connect.Request[connectorv1.SyncRequest]) (*connect.Response[connectorv1.SyncResponse], error) {
	if req.Header().Get("Authorization") != "Bearer "+coreToken {
		return nil, fmt.Errorf("sync: %w", connect.NewError(connect.CodeUnauthenticated, errNotCore))
	}
	cash := &connectorv1.Book{Kind: connectorv1.BookKind_BOOK_KIND_CASH, Currency: "USD"}
	equity := &connectorv1.Book{Kind: connectorv1.BookKind_BOOK_KIND_EQUITY, Currency: "USD"}
	return connect.NewResponse(&connectorv1.SyncResponse{Delivery: &connectorv1.Delivery{
		BatchId: req.Msg.GetBatchId(), Institution: "ibkr",
		Transactions: []*connectorv1.Transaction{{
			Key:           &connectorv1.TransactionKey{Source: "cash", ExternalId: "D1"},
			BrokerAccount: "U0000001", Kind: connectorv1.TransactionKind_TRANSACTION_KIND_DEPOSIT, Date: "2026-10-27",
			Postings: []*connectorv1.Posting{
				{Book: cash, Amount: &connectorv1.Money{Amount: "100.00", Currency: "USD"}},
				{Book: equity, Amount: &connectorv1.Money{Amount: "-100.00", Currency: "USD"}},
			},
		}},
		BrokerAccounts: []*connectorv1.BrokerAccount{{Id: "U0000001"}, {Id: "U0000002"}},
		FxRates:        []*connectorv1.FxRate{{Currency: "USD", Date: "2026-10-27", Rate: "4.0000"}},
	}}), nil
}

// viewBearer returns an interceptor that presents token as the view
// plugin's.
func viewBearer(token string) connect.UnaryInterceptorFunc {
	return func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			req.Header().Set("Authorization", "Bearer "+token)
			return next(ctx, req)
		}
	}
}

// TestOnboardingServices runs D-71, D-72 and D-73 through the real core:
// the owner adds a kind and changes sleeves through the SleeveService, a
// broker connector's sync books nothing for its unmapped broker accounts,
// the owner creates the account for one of them through the
// AccountService, and the next sync books it there; the read API lists
// kinds, sleeves, bindings and broker accounts.
func TestOnboardingServices(t *testing.T) {
	t.Parallel()
	dsn := emptyDatabase(t)
	code := enrollCode(t, enrollCmd(t, dsn))
	_, srv, _ := startCoreForUI(t, dsn, func(*serveConfig) {})
	session := passkeyLogin(t, authClient(srv.URL, srv.Client(), uiCaller), passkeyEnroll(t, authClient(srv.URL, srv.Client(), uiCaller), code))
	checkSleeveService(t, srv, session)

	port := serveConnector(t, brokerConnector{})
	cl := fakeCluster(t, port, "ibkr-main")
	out, err := connectorCmd(t, dsn, cl, "sync", "--instance=ibkr-main", "--mode=transactions")
	if err != nil || !strings.Contains(out, "0 transactions") {
		t.Fatalf("sync before the owner mapped anything = %q, %v; want nothing booked", out, err)
	}
	var unmapped viewsv1.ListBrokerAccountsResponse
	getJSON(t, srv, "/v1/broker-accounts?unmapped=true", session, &unmapped)
	if n := len(unmapped.GetBrokerAccounts()); n != 2 || unmapped.GetBrokerAccounts()[0].GetInstitution() != "ibkr" {
		t.Fatalf("unmapped broker accounts = %v, want both reported by ibkr", unmapped.GetBrokerAccounts())
	}
	createIBKRAccount(t, srv, session)
	out, err = connectorCmd(t, dsn, cl, "sync", "--instance=ibkr-main", "--mode=transactions")
	if err != nil || !strings.Contains(out, "1 transactions, 1 new") {
		t.Fatalf("sync after mapping = %q, %v; want the deposit booked", out, err)
	}
	checkMappedAccount(t, srv, session)
}

// createIBKRAccount creates the account IBKR main for the broker account
// U0000001 through the AccountService, an IKZE with a fee rule (D-76,
// D-78) that the core stores with it, and checks it can't be mapped to
// another account.
func createIBKRAccount(t *testing.T, srv *httptest.Server, session string) {
	t.Helper()
	accounts := ingestv1connect.NewAccountServiceClient(srv.Client(), srv.URL, connect.WithInterceptors(viewBearer(uiCaller)))
	created, err := accounts.CreateAccount(t.Context(), connect.NewRequest(&ingestv1.CreateAccountRequest{
		Origin: localUI, SessionToken: session, Name: "IBKR main", Sleeve: "satellite",
		ConnectorInstance: "ibkr-main", BrokerAccount: "U0000001", TaxWrapper: "ikze",
		FeeRule: &ingestv1.FeeRule{Rate: "0.0005", Minimum: "1", MinimumCurrency: "USD"},
	}))
	acc := created.Msg.GetAccount()
	if err != nil || acc.GetId() != "ibkr-main" || acc.GetInstitution() != "ibkr" || acc.GetTaxWrapper() != "ikze" || acc.GetFeeRule().GetMinimumCurrency() != "USD" {
		t.Fatalf("CreateAccount = %v, %v; want ibkr-main at ibkr, an IKZE with a USD minimum", created, err)
	}
	_, err = accounts.CreateAccount(t.Context(), connect.NewRequest(&ingestv1.CreateAccountRequest{
		Origin: localUI, SessionToken: session, Name: "Other", Sleeve: "core", ConnectorInstance: "ibkr-main", BrokerAccount: "U0000001",
	}))
	if connect.CodeOf(err) != connect.CodeFailedPrecondition {
		t.Errorf("mapping a mapped broker account again = %v, want FailedPrecondition", err)
	}
}

// checkSleeveService adds a kind and a sleeve, refuses a role other than
// the kind's without the override, and checks the read API lists them.
func checkSleeveService(t *testing.T, srv *httptest.Server, session string) {
	t.Helper()
	ctx := t.Context()
	sleeves := portfoliov1connect.NewSleeveServiceClient(srv.Client(), srv.URL, connect.WithInterceptors(viewBearer(uiCaller)))
	if _, err := sleeves.AddSleeveKind(ctx, connect.NewRequest(&portfoliov1.AddSleeveKindRequest{
		Origin: localUI, SessionToken: session, Name: "Gold", DefaultRole: portfoliov1.SleeveRole_SLEEVE_ROLE_DEFENSIVE,
	})); err != nil {
		t.Fatalf("AddSleeveKind: %v", err)
	}
	if _, err := sleeves.UpdateSleeve(ctx, connect.NewRequest(&portfoliov1.UpdateSleeveRequest{
		Origin: localUI, SessionToken: session, Id: "satellite", Name: "Satellite", Kind: "satellite",
		Role: portfoliov1.SleeveRole_SLEEVE_ROLE_DEFENSIVE,
	})); connect.CodeOf(err) != connect.CodeInvalidArgument {
		t.Errorf("a role other than the kind's without the override = %v, want InvalidArgument", err)
	}
	added, err := sleeves.AddSleeve(ctx, connect.NewRequest(&portfoliov1.AddSleeveRequest{
		Origin: localUI, SessionToken: session, Name: "Gold bars", Kind: "gold",
	}))
	if err != nil || added.Msg.GetSleeve().GetId() != "gold-bars" || added.Msg.GetSleeve().GetRole() != portfoliov1.SleeveRole_SLEEVE_ROLE_DEFENSIVE {
		t.Fatalf("AddSleeve = %v, %v; want gold-bars, defensive", added, err)
	}
	if _, err := sleeves.AddSleeve(ctx, connect.NewRequest(&portfoliov1.AddSleeveRequest{
		Origin: localUI, SessionToken: "unknown", Name: "X", Kind: "gold",
	})); connect.CodeOf(err) != connect.CodeUnauthenticated {
		t.Errorf("an unknown session = %v, want Unauthenticated", err)
	}
	var kinds viewsv1.ListSleeveKindsResponse
	getJSON(t, srv, "/v1/sleeve-kinds", session, &kinds)
	var listed viewsv1.ListSleevesResponse
	getJSON(t, srv, "/v1/sleeves", session, &listed)
	if len(kinds.GetKinds()) != 7 || len(listed.GetSleeves()) != 4 {
		t.Errorf("read API lists %d kinds and %d sleeves, want the 6 built-in and gold, and bonds, core, gold-bars and satellite",
			len(kinds.GetKinds()), len(listed.GetSleeves()))
	}
}

// checkMappedAccount checks that the deposit is booked in ibkr-main with
// the account in its key, and that the read API lists the binding and
// only U0000002 as unmapped.
func checkMappedAccount(t *testing.T, srv *httptest.Server, session string) {
	t.Helper()
	var txs viewsv1.ListTransactionsResponse
	getJSON(t, srv, "/v1/accounts/ibkr-main/transactions", session, &txs)
	if len(txs.GetTransactions()) != 1 || txs.GetTransactions()[0].GetSource() != "plugin:ibkr-main:cash:ibkr-main" {
		t.Errorf("ibkr-main's transactions = %v, want the deposit keyed plugin:ibkr-main:cash:ibkr-main", txs.GetTransactions())
	}
	var bindings viewsv1.ListConnectorBindingsResponse
	getJSON(t, srv, "/v1/connector-bindings", session, &bindings)
	if b := bindings.GetBindings(); len(b) != 1 || b[0].GetConnectorInstance() != "ibkr-main" || b[0].GetAccountId() != "ibkr-main" {
		t.Errorf("bindings = %v, want ibkr-main bound to ibkr-main", b)
	}
	var unmapped viewsv1.ListBrokerAccountsResponse
	getJSON(t, srv, "/v1/broker-accounts?unmapped=true", session, &unmapped)
	if b := unmapped.GetBrokerAccounts(); len(b) != 1 || b[0].GetId() != "U0000002" {
		t.Errorf("unmapped broker accounts = %v, want only U0000002", b)
	}
}
