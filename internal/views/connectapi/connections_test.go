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
}

func TestRESTConnections(t *testing.T) {
	t.Parallel()
	c := loadContract(t)
	a := api{handler: connectionsAPI(t)}
	for path, want := range map[string]string{
		"/v1/sleeve-kinds":       `"defaultRole":"SLEEVE_ROLE_OFFENSIVE"`,
		"/v1/connector-bindings": `"connectorInstance":"ibkr-main"`,
		"/v1/sleeves":            `"kind"`,
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

// TestReadAPIServesNoBrokerAccounts checks D-121: the deprecated
// ListBrokerAccounts answers UNIMPLEMENTED, over REST and Connect alike:
// the read API has no way to the broker accounts at all.
func TestReadAPIServesNoBrokerAccounts(t *testing.T) {
	t.Parallel()
	a := api{handler: connectionsAPI(t)}
	for _, path := range []string{"/v1/broker-accounts", "/v1/broker-accounts?unmapped=true"} {
		_, rec := a.serve(t, http.MethodGet, path, goodToken)
		if rec.Code != http.StatusNotImplemented || bytes.Contains(rec.Body.Bytes(), []byte("U000000")) {
			t.Errorf("GET %s = %d %s, want 501 naming no broker account", path, rec.Code, rec.Body)
		}
	}
	_, rec := a.serve(t, http.MethodPost, "/openfief.views.v1.ReadService/ListBrokerAccounts", goodToken)
	if rec.Code == http.StatusOK || bytes.Contains(rec.Body.Bytes(), []byte("U000000")) {
		t.Errorf("Connect ListBrokerAccounts = %d %s, want a refusal naming no broker account", rec.Code, rec.Body)
	}
}
