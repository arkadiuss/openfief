package connectapi_test

import (
	"bytes"
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/arkadiuss/openfief/internal/ingest"
	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/views/app"
	"github.com/arkadiuss/openfief/internal/views/connectapi"
	viewsv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/views/v1"
	"github.com/arkadiuss/openfief/sdk/go/gen/openfief/views/v1/viewsv1connect"
)

// fakeConnections holds one binding and two synthetic broker accounts of
// ibkr-main, U0000001 mapped and U0000002 not.
type fakeConnections struct {
	seen kernel.Date
}

func (fakeConnections) Bindings(context.Context) ([]ingest.Binding, error) {
	return []ingest.Binding{{Instance: "ibkr-main", Account: "acc-1"}}, nil
}

func (f fakeConnections) AllBrokerAccounts(context.Context) ([]ingest.BrokerAccount, error) {
	return []ingest.BrokerAccount{
		{Instance: "ibkr-main", ID: "U0000001", Institution: "ibkr", Account: "acc-1", LastSeen: f.seen, InLastDelivery: true},
		{Instance: "ibkr-main", ID: "U0000002", Institution: "ibkr", LastSeen: f.seen},
	}, nil
}

// connectionsAPI is the read API over the synthetic ledger and connections.
func connectionsAPI(t *testing.T) http.Handler {
	t.Helper()
	h, err := connectapi.New(newViews(t, newLedger(t), app.WithConnections(fakeConnections{seen: date(t, "2026-03-02")})), fakeVerifier{},
		connectapi.WithLogger(slog.New(slog.NewJSONHandler(&bytes.Buffer{}, nil))),
		connectapi.WithRequestIDs(func() string { return "req-1" }))
	if err != nil {
		t.Fatalf("connectapi.New: %v", err)
	}
	return h
}

func TestConnectListConnections(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(connectionsAPI(t))
	t.Cleanup(srv.Close)
	c := viewsv1connect.NewReadServiceClient(srv.Client(), srv.URL)
	ctx := context.Background()
	kinds, err := c.ListSleeveKinds(ctx, authed(goodToken, &viewsv1.ListSleeveKindsRequest{}))
	if err != nil || len(kinds.Msg.GetKinds()) != 6 || kinds.Msg.GetKinds()[0].GetDefaultRole() != viewsv1.SleeveRole_SLEEVE_ROLE_OFFENSIVE {
		t.Errorf("ListSleeveKinds = %v, %v; want the six built-in kinds, agent-pool offensive first", kinds, err)
	}
	checkConnectionLists(t, c)
}

// checkConnectionLists checks the bindings and broker accounts c lists.
func checkConnectionLists(t *testing.T, c viewsv1connect.ReadServiceClient) {
	t.Helper()
	ctx := context.Background()
	bindings, err := c.ListConnectorBindings(ctx, authed(goodToken, &viewsv1.ListConnectorBindingsRequest{}))
	if b := bindings.Msg.GetBindings(); err != nil || len(b) != 1 || b[0].GetConnectorInstance() != "ibkr-main" || b[0].GetAccountId() != "acc-1" {
		t.Errorf("ListConnectorBindings = %v, %v", bindings, err)
	}
	checkBrokerAccountList(t, c)
}

// checkBrokerAccountList checks that c lists only the unmapped broker
// account on request, and nothing without a token.
func checkBrokerAccountList(t *testing.T, c viewsv1connect.ReadServiceClient) {
	t.Helper()
	ctx := context.Background()
	unmapped, err := c.ListBrokerAccounts(ctx, authed(goodToken, &viewsv1.ListBrokerAccountsRequest{Unmapped: true}))
	if b := unmapped.Msg.GetBrokerAccounts(); err != nil || len(b) != 1 || b[0].GetId() != "U0000002" || b[0].GetAccountId() != "" ||
		b[0].GetLastSeen() != "2026-03-02" || b[0].GetInLastDelivery() {
		t.Errorf("ListBrokerAccounts(unmapped) = %v, %v; want only U0000002, unmapped", unmapped, err)
	}
	if _, err := c.ListBrokerAccounts(ctx, authed("", &viewsv1.ListBrokerAccountsRequest{})); err == nil {
		t.Error("ListBrokerAccounts without a token answered")
	}
}

func TestRESTConnections(t *testing.T) {
	t.Parallel()
	c := loadContract(t)
	a := api{handler: connectionsAPI(t)}
	for path, want := range map[string]string{
		"/v1/sleeve-kinds":                  `"defaultRole":"SLEEVE_ROLE_OFFENSIVE"`,
		"/v1/connector-bindings":            `"connectorInstance":"ibkr-main"`,
		"/v1/broker-accounts":               `"accountId":"acc-1"`,
		"/v1/broker-accounts?unmapped=true": `"id":"U0000002"`,
		"/v1/sleeves":                       `"kind"`,
	} {
		req, rec := a.serve(t, http.MethodGet, path, goodToken)
		if rec.Code != http.StatusOK {
			t.Fatalf("GET %s = %d\n%s", path, rec.Code, rec.Body)
		}
		c.check(t, req, rec)
		if !bytes.Contains(rec.Body.Bytes(), []byte(want)) {
			t.Errorf("GET %s = %s, want %s", path, rec.Body, want)
		}
	}
}
