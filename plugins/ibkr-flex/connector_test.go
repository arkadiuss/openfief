package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"connectrpc.com/connect"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/arkadiuss/openfief/internal/ingest/ibkrflex"
	"github.com/arkadiuss/openfief/internal/plugins"
	connectorv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/connector/v1"
	"github.com/arkadiuss/openfief/sdk/go/gen/openfief/connector/v1/connectorv1connect"
	"github.com/arkadiuss/openfief/sdk/go/openfief"
)

// Synthetic credentials; not a real Flex token or query.
const (
	syntheticToken   = "synthetic-flex-token-0000"
	syntheticQueryID = "000000"
	coreToken        = "synthetic-core-token"
)

var (
	errIBKRDown = fmt.Errorf("synthetic: %w", ibkrflex.ErrFlexService)
	errOther    = errors.New("synthetic: something else")
)

// statement is the ibkrflex adapter's synthetic Flex statement.
func statement(t *testing.T) []byte {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("..", "..", "internal", "ingest", "ibkrflex", "testdata", "activity.xml"))
	if err != nil {
		t.Fatalf("read statement: %v", err)
	}
	return b
}

// credentials writes the Secret's files into a new directory; an empty value
// leaves its file out.
func credentials(t *testing.T, token, queryID string) string {
	t.Helper()
	dir := t.TempDir()
	for name, v := range map[string]string{tokenFile: token, queryIDFile: queryID} {
		if v == "" {
			continue
		}
		if err := os.WriteFile(filepath.Join(dir, name), []byte(v+"\n"), 0o600); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
	}
	return dir
}

// flexService is a fake Flex Web Service: it answers with body for the
// synthetic credentials and records what it was called with.
type flexService struct {
	body    []byte
	err     error
	token   string
	queryID string
}

func (f *flexService) fetch(_ context.Context, token, queryID string) ([]byte, error) {
	f.token, f.queryID = token, queryID
	return f.body, f.err
}

func newConnector(t *testing.T, f *flexService, secrets string) *connector {
	t.Helper()
	return &connector{
		mapper:  ibkrflex.Mapper{ByBrokerAccount: true},
		fetch:   f.fetch,
		secrets: secrets,
	}
}

func TestSync(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		mode     connectorv1.Mode
		secrets  func(t *testing.T) string
		flex     flexService
		wantCode connect.Code
	}{
		{name: "opening mode", mode: connectorv1.Mode_MODE_OPENING},
		{name: "transactions mode", mode: connectorv1.Mode_MODE_TRANSACTIONS},
		{name: "no mode", mode: connectorv1.Mode_MODE_UNSPECIFIED, wantCode: connect.CodeInvalidArgument},
		{
			name: "no token file", mode: connectorv1.Mode_MODE_OPENING, wantCode: connect.CodeFailedPrecondition,
			secrets: func(t *testing.T) string { t.Helper(); return credentials(t, "", syntheticQueryID) },
		},
		{
			name: "empty query ID", mode: connectorv1.Mode_MODE_OPENING, wantCode: connect.CodeFailedPrecondition,
			secrets: func(t *testing.T) string { t.Helper(); return credentials(t, syntheticToken, " ") },
		},
		{name: "IBKR failing is unavailable", mode: connectorv1.Mode_MODE_OPENING, flex: flexService{err: errIBKRDown}, wantCode: connect.CodeUnavailable},
		{name: "IBKR sending a statement too large", mode: connectorv1.Mode_MODE_OPENING, flex: flexService{err: ibkrflex.ErrTooLarge}, wantCode: connect.CodeInvalidArgument},
		{name: "not a Flex statement", mode: connectorv1.Mode_MODE_OPENING, flex: flexService{body: []byte("<html>")}, wantCode: connect.CodeInvalidArgument},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			f := tt.flex
			if f.body == nil && f.err == nil {
				f.body = statement(t)
			}
			secrets := credentials(t, syntheticToken, syntheticQueryID)
			if tt.secrets != nil {
				secrets = tt.secrets(t)
			}
			resp, err := newConnector(t, &f, secrets).Sync(t.Context(), connect.NewRequest(&connectorv1.SyncRequest{
				BatchId: "b-1", Mode: tt.mode, Watermark: "w-1",
			}))
			checkSync(t, &f, resp, err, tt.wantCode)
		})
	}
}

// checkSync fails t unless the Sync answer is the wanted error code, or,
// with no code wanted, a delivery of the statement for batch b-1 fetched
// with the Secret's credentials. No error may contain the token.
func checkSync(t *testing.T, f *flexService, resp *connect.Response[connectorv1.SyncResponse], err error, wantCode connect.Code) {
	t.Helper()
	if err != nil && strings.Contains(err.Error(), syntheticToken) {
		t.Errorf("error %q contains the token", err)
	}
	switch {
	case wantCode != 0 && connect.CodeOf(err) != wantCode:
		t.Fatalf("Sync error = %v, want code %v", err, wantCode)
	case wantCode != 0:
		return
	case err != nil:
		t.Fatalf("Sync: %v", err)
	}
	checkDelivery(t, f, resp)
}

// checkDelivery fails t unless resp delivers the statement for batch b-1,
// fetched with the Secret's credentials.
func checkDelivery(t *testing.T, f *flexService, resp *connect.Response[connectorv1.SyncResponse]) {
	t.Helper()
	if f.token != syntheticToken || f.queryID != syntheticQueryID {
		t.Errorf("fetched with %q, %q; want the Secret's token and query ID", f.token, f.queryID)
	}
	d := resp.Msg.GetDelivery()
	if d.GetBatchId() != "b-1" || d.GetInstitution() != "ibkr" || len(d.GetTransactions()) == 0 || resp.Msg.GetWatermark() != "w-1" {
		t.Errorf("Sync = batch %q, institution %q, %d transactions, watermark %q; want b-1, ibkr, some, w-1",
			d.GetBatchId(), d.GetInstitution(), len(d.GetTransactions()), resp.Msg.GetWatermark())
	}
	checkBrokerAccounts(t, d)
}

// checkBrokerAccounts fails t unless every entry of d names its IBKR
// account for the core to map (D-72), and the account the synthetic
// statement is of is reported.
func checkBrokerAccounts(t *testing.T, d *connectorv1.Delivery) {
	t.Helper()
	if accounts := d.GetBrokerAccounts(); len(accounts) == 0 || accounts[0].GetId() != "U0000000" {
		t.Errorf("broker_accounts = %v, want the synthetic U0000000 first", accounts)
	}
	for _, tx := range d.GetTransactions() {
		if tx.GetAccountId() != "" || tx.GetBrokerAccount() == "" {
			t.Fatalf("transaction %v: account_id %q, broker_account %q; want only a broker account", tx.GetKey(), tx.GetAccountId(), tx.GetBrokerAccount())
		}
	}
}

// TestSyncModes checks that the request's mode reaches the mapping: only
// opening mode seeds opening lots.
func TestSyncModes(t *testing.T) {
	t.Parallel()
	kinds := map[connectorv1.Mode]int{}
	for _, mode := range []connectorv1.Mode{connectorv1.Mode_MODE_OPENING, connectorv1.Mode_MODE_TRANSACTIONS} {
		c := newConnector(t, &flexService{body: statement(t)}, credentials(t, syntheticToken, syntheticQueryID))
		resp, err := c.Sync(t.Context(), connect.NewRequest(&connectorv1.SyncRequest{BatchId: "b", Mode: mode}))
		if err != nil {
			t.Fatalf("Sync(%s): %v", mode, err)
		}
		for _, tx := range resp.Msg.GetDelivery().GetTransactions() {
			if tx.GetKind() == connectorv1.TransactionKind_TRANSACTION_KIND_OPENING_LOT {
				kinds[mode]++
			}
		}
	}
	if kinds[connectorv1.Mode_MODE_OPENING] == 0 || kinds[connectorv1.Mode_MODE_TRANSACTIONS] != 0 {
		t.Errorf("opening lots per mode = %v, want some in opening mode and none in transactions mode", kinds)
	}
}

type fakeVerifier struct{}

func (fakeVerifier) VerifyCore(_ context.Context, token string) error {
	if token != coreToken {
		return openfief.ErrNotCore
	}
	return nil
}

// importFile serves c with the SDK and streams file to it as the core, in
// chunks of at most 1000 bytes.
func importFile(t *testing.T, c *connector, mode connectorv1.Mode, file []byte) (*connectorv1.ImportFileResponse, error) {
	t.Helper()
	server := openfief.NewConnectorServer("", c, fakeVerifier{})
	srv := httptest.NewUnstartedServer(server.Handler)
	srv.Config.Protocols = server.Protocols
	srv.Start()
	t.Cleanup(srv.Close)
	var p http.Protocols
	p.SetUnencryptedHTTP2(true)
	client := connectorv1connect.NewConnectorServiceClient(&http.Client{Transport: &http.Transport{Protocols: &p}}, srv.URL)
	stream := client.ImportFile(t.Context())
	stream.RequestHeader().Set("Authorization", "Bearer "+coreToken)
	msgs := []*connectorv1.ImportFileRequest{{Part: &connectorv1.ImportFileRequest_Header{
		Header: &connectorv1.FileHeader{BatchId: "b-2", Mode: mode, FileName: "activity.xml", MediaType: "application/xml"},
	}}}
	for len(file) > 0 {
		n := min(len(file), 1000)
		msgs = append(msgs, &connectorv1.ImportFileRequest{Part: &connectorv1.ImportFileRequest_Chunk{Chunk: file[:n]}})
		file = file[n:]
	}
	for _, m := range msgs {
		if err := stream.Send(m); err != nil {
			break // the connector answered early; CloseAndReceive has its answer
		}
	}
	resp, err := stream.CloseAndReceive()
	if err != nil {
		return nil, fmt.Errorf("import file: %w", err)
	}
	return resp.Msg, nil
}

func TestImportFile(t *testing.T) {
	t.Parallel()
	c := newConnector(t, &flexService{}, t.TempDir())
	got, err := importFile(t, c, connectorv1.Mode_MODE_TRANSACTIONS, statement(t))
	if err != nil {
		t.Fatalf("ImportFile: %v", err)
	}
	if d := got.GetDelivery(); d.GetBatchId() != "b-2" || len(d.GetTransactions()) == 0 {
		t.Errorf("ImportFile = batch %q with %d transactions, want b-2 with some", d.GetBatchId(), len(d.GetTransactions()))
	}
	tests := []struct {
		name string
		mode connectorv1.Mode
		file []byte
	}{
		{name: "no mode", mode: connectorv1.Mode_MODE_UNSPECIFIED, file: statement(t)},
		{name: "not a Flex statement", mode: connectorv1.Mode_MODE_OPENING, file: []byte("a,b\n1,2\n")},
	}
	for _, tt := range tests {
		if _, err := importFile(t, c, tt.mode, tt.file); connect.CodeOf(err) != connect.CodeInvalidArgument {
			t.Errorf("%s: ImportFile = %v, want InvalidArgument", tt.name, err)
		}
	}
}

func TestFetchCode(t *testing.T) {
	t.Parallel()
	tests := map[error]connect.Code{
		errIBKRDown:                   connect.CodeUnavailable,
		ibkrflex.ErrStatementNotReady: connect.CodeUnavailable,
		ibkrflex.ErrTooLarge:          connect.CodeInvalidArgument,
		errOther:                      connect.CodeUnavailable,
	}
	for err, want := range tests {
		if got := fetchCode(err); got != want {
			t.Errorf("fetchCode(%v) = %v, want %v", err, got, want)
		}
	}
}

// TestManifest checks that manifest.json, which the image carries as its
// signed label (D-42), is a valid manifest asking for exactly what the
// plugin needs: the connector capability, IBKR's Flex Web Service host and
// the credentials Secret where the plugin reads it.
func TestManifest(t *testing.T) {
	t.Parallel()
	data, err := os.ReadFile("manifest.json")
	if err != nil {
		t.Fatalf("read manifest: %v", err)
	}
	got, err := plugins.ParseManifestLabel(string(data))
	if err != nil {
		t.Fatalf("ParseManifestLabel: %v", err)
	}
	want := plugins.Manifest{
		Capabilities: []plugins.Capability{plugins.CapabilityConnector},
		Egress:       []plugins.EgressHost{{Host: "ndcdyn.interactivebrokers.com", Port: 443}},
		Secrets:      []plugins.SecretMount{{Role: "flex-token", MountPath: flexMount}},
	}
	if diff := cmp.Diff(want, got, cmpopts.EquateEmpty(), cmpopts.IgnoreFields(plugins.Manifest{}, "ConfigSchema")); diff != "" {
		t.Errorf("manifest (-want +got):\n%s", diff)
	}
	if len(got.ConfigSchema) == 0 {
		t.Error("the manifest has no config schema")
	}
}
