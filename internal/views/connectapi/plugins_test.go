package connectapi_test

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"connectrpc.com/connect"
	"github.com/google/go-cmp/cmp"
	"google.golang.org/protobuf/testing/protocmp"

	"github.com/arkadiuss/openfief/internal/ingest"
	"github.com/arkadiuss/openfief/internal/views"
	"github.com/arkadiuss/openfief/internal/views/app"
	"github.com/arkadiuss/openfief/internal/views/connectapi"
	viewsv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/views/v1"
	"github.com/arkadiuss/openfief/sdk/go/gen/openfief/views/v1/viewsv1connect"
)

// pluginImage is a synthetic image reference; not of any real image.
const pluginImage = "registry.example.com/openfief/ibkr-flex@sha256:a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1"

// fakePlugins answers the installed plugins from fixed records.
type fakePlugins []views.PluginRecord

func (f fakePlugins) InstalledPlugins(context.Context) ([]views.PluginRecord, error) { return f, nil }

// installed are two plugins: ibkr-main, signed by OpenFief's CI and
// running, and ui, refused. ibkr-old, which brought transactions in, is no
// longer installed.
func installed() fakePlugins {
	return fakePlugins{
		{
			Instance: "ui", Image: pluginImage, Requested: []string{"view"},
			Ready:     views.Condition{Status: "False", Reason: "SignatureNotVerified"},
			Signature: views.Condition{Status: "False", Reason: "Unsigned"},
		},
		{
			Instance: "ibkr-main", Image: pluginImage, Requested: []string{"connector"}, Effective: []string{"connector"},
			Ready:     views.Condition{Status: "True", Reason: "Running"},
			Signature: views.Condition{Status: "True", Reason: "SignedByTrustedSigner"},
			Trust:     "Unverified", SignerKind: "OpenFiefCI", Signer: "openfief-ci",
		},
	}
}

// pluginAPI is the read API over the synthetic ledger, whose buy came from
// the removed instance ibkr-old and opening lot from ibkr-main, reading the
// installed plugins from plugins, or none when plugins is nil.
func pluginAPI(t *testing.T, plugins app.Plugins) http.Handler {
	t.Helper()
	l := newLedger(t)
	for i := range l.txs {
		instance := "ibkr-main"
		if l.txs[i].Key.ExternalID == "buy-1" {
			instance = "ibkr-old"
		}
		l.txs[i].Key.Source = ingest.ConnectorKeySource(instance, "ibkr")
	}
	var opts []app.Option
	if plugins != nil {
		opts = append(opts, app.WithPlugins(plugins))
	}
	h, err := connectapi.New(newViews(t, l, opts...), fakeVerifier{},
		connectapi.WithLogger(slog.New(slog.NewJSONHandler(&bytes.Buffer{}, nil))),
		connectapi.WithRequestIDs(func() string { return "req-1" }))
	if err != nil {
		t.Fatalf("connectapi.New: %v", err)
	}
	return h
}

func TestConnectListPlugins(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(pluginAPI(t, installed()))
	t.Cleanup(srv.Close)
	c := viewsv1connect.NewReadServiceClient(srv.Client(), srv.URL)
	got, err := c.ListPlugins(context.Background(), authed(goodToken, &viewsv1.ListPluginsRequest{}))
	if err != nil {
		t.Fatalf("ListPlugins: %v", err)
	}
	want := &viewsv1.ListPluginsResponse{Plugins: []*viewsv1.Plugin{
		{
			Instance: "ibkr-main", Image: pluginImage, ImageDigest: views.ImageDigest(pluginImage),
			Capabilities: []string{"connector"}, RequestedCapabilities: []string{"connector"},
			Trust: viewsv1.SignatureTrust_SIGNATURE_TRUST_OPENFIEF_CI, Signer: "openfief-ci",
			TrustReason: "SignedByTrustedSigner", Phase: "Running", Ready: true,
		},
		{
			Instance: "ui", Image: pluginImage, ImageDigest: views.ImageDigest(pluginImage),
			RequestedCapabilities: []string{"view"}, Trust: viewsv1.SignatureTrust_SIGNATURE_TRUST_REFUSED,
			TrustReason: "Unsigned", Phase: "SignatureNotVerified",
		},
	}}
	if diff := cmp.Diff(want, got.Msg, protocmp.Transform()); diff != "" {
		t.Errorf("ListPlugins (-want +got):\n%s", diff)
	}
	_, unauthenticated := c.ListPlugins(context.Background(), authed("", &viewsv1.ListPluginsRequest{}))
	if connect.CodeOf(unauthenticated) != connect.CodeUnauthenticated {
		t.Errorf("ListPlugins without a token = %v, want unauthenticated", unauthenticated)
	}
}

func TestConnectTransactionOrigins(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		plugins app.Plugins
		want    map[string][2]any // external ID -> instance, removed
	}{
		{"the removed instance is marked", installed(), map[string][2]any{"buy-1": {"ibkr-old", true}, "open-1": {"ibkr-main", false}}},
		{"nothing is marked without the plugin list", nil, map[string][2]any{"buy-1": {"ibkr-old", false}, "open-1": {"ibkr-main", false}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			srv := httptest.NewServer(pluginAPI(t, tt.plugins))
			t.Cleanup(srv.Close)
			c := viewsv1connect.NewReadServiceClient(srv.Client(), srv.URL)
			page, err := c.ListTransactions(context.Background(), authed(goodToken, &viewsv1.ListTransactionsRequest{AccountId: "acc-1"}))
			if err != nil {
				t.Fatalf("ListTransactions: %v", err)
			}
			got := map[string][2]any{}
			for _, tx := range page.Msg.GetTransactions() {
				got[tx.GetExternalId()] = [2]any{tx.GetConnectorInstance(), tx.GetConnectorInstanceRemoved()}
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("origins (-want +got):\n%s", diff)
			}
		})
	}
}

func TestRESTPlugins(t *testing.T) {
	t.Parallel()
	c := loadContract(t)
	a := api{handler: pluginAPI(t, installed())}
	req, rec := a.serve(t, http.MethodGet, "/v1/plugins", goodToken)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET /v1/plugins = %d\n%s", rec.Code, rec.Body)
	}
	c.check(t, req, rec)
	var body struct {
		Plugins []struct {
			Instance    string `json:"instance"`
			Trust       string `json:"trust"`
			Signer      string `json:"signer"`
			ImageDigest string `json:"imageDigest"`
		} `json:"plugins"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decoding %s: %v", rec.Body, err)
	}
	if len(body.Plugins) != 2 || body.Plugins[0].Trust != "SIGNATURE_TRUST_OPENFIEF_CI" || body.Plugins[0].Signer != "openfief-ci" ||
		body.Plugins[0].ImageDigest == "" || body.Plugins[1].Trust != "SIGNATURE_TRUST_REFUSED" {
		t.Errorf("GET /v1/plugins = %s, want ibkr-main signed by OpenFief's CI, then ui refused", rec.Body)
	}
}

func TestRESTTransactionOrigins(t *testing.T) {
	t.Parallel()
	c := loadContract(t)
	a := api{handler: pluginAPI(t, installed())}
	req, rec := a.serve(t, http.MethodGet, "/v1/accounts/acc-1/transactions", goodToken)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET transactions = %d\n%s", rec.Code, rec.Body)
	}
	c.check(t, req, rec)
	if !bytes.Contains(rec.Body.Bytes(), []byte(`"connectorInstance":"ibkr-old"`)) ||
		!bytes.Contains(rec.Body.Bytes(), []byte(`"connectorInstanceRemoved":true`)) {
		t.Errorf("transactions = %s, want ibkr-old marked removed", rec.Body)
	}
}

func TestRESTPluginsUnavailable(t *testing.T) {
	t.Parallel()
	a := api{handler: pluginAPI(t, nil)}
	_, rec := a.serve(t, http.MethodGet, "/v1/plugins", goodToken)
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("GET /v1/plugins without the plugin list = %d, want 503\n%s", rec.Code, rec.Body)
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte("the installed plugins can't be read now")) {
		t.Errorf("body = %s, want the fixed text", rec.Body)
	}
}

func TestPluginEnumsMatch(t *testing.T) {
	t.Parallel()
	for tr := views.TrustVerified; tr <= views.TrustNotChecked; tr++ {
		checkEnum(t, viewsv1.SignatureTrust(0).Descriptor(), "SIGNATURE_TRUST_", int(tr), tr.String())
	}
}
