package main

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"slices"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/testing/protocmp"

	"github.com/arkadiuss/openfief/internal/ingest/xtb/xtbtest"
	"github.com/arkadiuss/openfief/internal/plugins"
	connectorv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/connector/v1"
	"github.com/arkadiuss/openfief/sdk/go/gen/openfief/connector/v1/connectorv1connect"
	"github.com/arkadiuss/openfief/sdk/go/openfief"
)

// coreToken is the synthetic token the fake core presents.
const coreToken = "synthetic-core-token"

// bankComment is a made-up deposit comment naming a sending bank.
const bankComment = "Przelew z Bank Testowy S.A."

// export is a synthetic export: a deposit, a buy of TESTA.PL and a
// withdrawal whose comment names the account.
func export() *xtbtest.Export {
	day := func(d int) time.Time { return time.Date(2026, 1, d, 10, 0, 0, 0, time.UTC) }
	return &xtbtest.Export{
		Account: xtbtest.Account, Currency: "PLN", From: day(1), To: day(31),
		Cash: [][]xtbtest.Cell{
			xtbtest.CashRow("Withdrawal", "", "", "", day(9), "-100", "9000000003", "Wyplata 0000000"),
			xtbtest.CashRow("Stock purchase", "Test Alpha SA", "TESTA.PL", "STOCK", day(5), "-400", "9000000002", "OPEN BUY 4/4 @ 100"),
			xtbtest.CashRow("Deposit", "", "", "", day(2), "1000", "9000000001", bankComment),
		},
	}
}

type fakeVerifier struct{}

func (fakeVerifier) VerifyCore(_ context.Context, token string) error {
	if token != coreToken {
		return openfief.ErrNotCore
	}
	return nil
}

// fakeCore is a client of a connector served with the SDK, as the core.
type fakeCore struct {
	connectorv1connect.ConnectorServiceClient
}

// client serves a connector with the SDK and returns a client of it that
// authenticates as the core.
func client(t *testing.T) fakeCore {
	t.Helper()
	server := openfief.NewConnectorServer("", &connector{}, fakeVerifier{})
	srv := httptest.NewUnstartedServer(server.Handler)
	srv.Config.Protocols = server.Protocols
	srv.Start()
	t.Cleanup(srv.Close)
	var p http.Protocols
	p.SetUnencryptedHTTP2(true)
	return fakeCore{connectorv1connect.NewConnectorServiceClient(&http.Client{Transport: &http.Transport{Protocols: &p}}, srv.URL)}
}

// importFile streams file, named name, to the connector as the core, in
// chunks of at most 1000 bytes.
func importFile(t *testing.T, mode connectorv1.Mode, name string, file []byte) (*connectorv1.ImportFileResponse, error) {
	t.Helper()
	return importWith(t, &connectorv1.FileHeader{BatchId: "b-1", Mode: mode, FileName: name}, file)
}

// importWith streams file to the connector as the core with header h, in
// chunks of at most 1000 bytes.
func importWith(t *testing.T, h *connectorv1.FileHeader, file []byte) (*connectorv1.ImportFileResponse, error) {
	t.Helper()
	stream := client(t).ImportFile(t.Context())
	stream.RequestHeader().Set("Authorization", "Bearer "+coreToken)
	msgs := []*connectorv1.ImportFileRequest{{Part: &connectorv1.ImportFileRequest_Header{Header: h}}}
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
	e := export()
	got, err := importFile(t, connectorv1.Mode_MODE_TRANSACTIONS, e.FileName(), e.Bytes())
	if err != nil {
		t.Fatalf("ImportFile: %v", err)
	}
	d := got.GetDelivery()
	want := &connectorv1.Delivery{
		BatchId: "b-1", Institution: "xtb",
		BrokerAccounts: []*connectorv1.BrokerAccount{{Id: xtbtest.Account}},
		Tickers:        []*connectorv1.Ticker{{Symbol: "TESTA.PL", Code: "TESTA", Country: "PL", Name: "Test Alpha SA"}},
		TickerTransactions: []*connectorv1.TickerTransaction{{
			Symbol: "TESTA.PL", Key: &connectorv1.TransactionKey{Source: "xtb", ExternalId: "9000000002"},
			BrokerAccount: xtbtest.Account, Kind: connectorv1.TransactionKind_TRANSACTION_KIND_BUY, Date: "2026-01-05",
			Quantity: "4", Amount: &connectorv1.Money{Amount: "400", Currency: "PLN"}, Description: "buy TESTA.PL",
			Ref: "Cash Operations#7[ID=9000000002]",
		}},
	}
	if diff := cmp.Diff(want, d, protocmp.Transform(), protocmp.IgnoreFields(&connectorv1.Delivery{}, "transactions")); diff != "" {
		t.Errorf("delivery (-want +got):\n%s", diff)
	}
	kinds := make([]connectorv1.TransactionKind, 0, len(d.GetTransactions()))
	for _, tx := range d.GetTransactions() {
		if tx.GetBrokerAccount() != xtbtest.Account || tx.GetAccountId() != "" {
			t.Errorf("transaction %s is not named by the broker account alone", tx.GetKey())
		}
		kinds = append(kinds, tx.GetKind())
	}
	if !slices.Equal(kinds, []connectorv1.TransactionKind{connectorv1.TransactionKind_TRANSACTION_KIND_DEPOSIT, connectorv1.TransactionKind_TRANSACTION_KIND_WITHDRAWAL}) {
		t.Errorf("transaction kinds = %v, want a deposit and a withdrawal", kinds)
	}
	raw, err := proto.Marshal(d)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	for _, secret := range []string{bankComment, "Wyplata"} {
		if bytes.Contains(raw, []byte(secret)) {
			t.Errorf("the delivery holds the transfer comment %q", secret)
		}
	}
}

func TestImportFileRefused(t *testing.T) {
	t.Parallel()
	e := export()
	undated := export()
	undated.NoReportDate = true
	tests := []struct {
		name string
		mode connectorv1.Mode
		file []byte
		cash *connectorv1.Money
	}{
		{name: "no mode", mode: connectorv1.Mode_MODE_UNSPECIFIED, file: e.Bytes()},
		{name: "opening mode without the report's date", mode: connectorv1.Mode_MODE_OPENING, file: undated.Bytes()},
		{name: "opening cash in another currency", mode: connectorv1.Mode_MODE_OPENING, file: e.Bytes(), cash: &connectorv1.Money{Amount: "100", Currency: "EUR"}},
		{name: "opening cash that isn't an amount", mode: connectorv1.Mode_MODE_OPENING, file: e.Bytes(), cash: &connectorv1.Money{Amount: "1 000", Currency: "PLN"}},
		{name: "opening cash with transactions", mode: connectorv1.Mode_MODE_TRANSACTIONS, file: e.Bytes(), cash: &connectorv1.Money{Amount: "100", Currency: "PLN"}},
		{name: "not an export", mode: connectorv1.Mode_MODE_TRANSACTIONS, file: []byte("a,b\n1,2\n")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			h := &connectorv1.FileHeader{BatchId: "b-1", Mode: tt.mode, FileName: e.FileName(), OpeningCash: tt.cash}
			if _, err := importWith(t, h, tt.file); connect.CodeOf(err) != connect.CodeInvalidArgument {
				t.Errorf("ImportFile = %v, want InvalidArgument", err)
			}
		})
	}
}

// TestImportFileOpening checks that a current export in opening mode seeds
// the account from its open positions, each an opening lot by ticker, and
// the cash balance the owner entered, on the report's date (D-136).
func TestImportFileOpening(t *testing.T) {
	t.Parallel()
	e := export()
	e.Open = [][]xtbtest.Cell{xtbtest.OpenRow(&xtbtest.Position{
		Name: "Test Alpha SA", Ticker: "TESTA.PL", Category: "STOCK", Volume: "4", OpenPrice: "100",
		Opened: time.Date(2026, 1, 5, 10, 0, 0, 0, time.UTC), Value: "400",
	})}
	h := &connectorv1.FileHeader{
		BatchId: "b-1", Mode: connectorv1.Mode_MODE_OPENING, FileName: e.FileName(),
		OpeningCash: &connectorv1.Money{Amount: "500.00", Currency: "PLN"},
	}
	got, err := importWith(t, h, e.Bytes())
	if err != nil {
		t.Fatalf("ImportFile: %v", err)
	}
	d := got.GetDelivery()
	wantLot := []*connectorv1.TickerTransaction{{
		Symbol: "TESTA.PL", Key: &connectorv1.TransactionKey{Source: "xtb", ExternalId: "open-position:TESTA.PL@20260105T100000Z#1"},
		BrokerAccount: xtbtest.Account, Kind: connectorv1.TransactionKind_TRANSACTION_KIND_OPENING_LOT, Date: "2026-01-31",
		Quantity: "4", Amount: &connectorv1.Money{Amount: "400", Currency: "PLN"}, Description: "opening lot TESTA.PL",
		Ref: "Open Positions#10", LotOpened: "2026-01-05",
	}}
	if diff := cmp.Diff(wantLot, d.GetTickerTransactions(), protocmp.Transform()); diff != "" {
		t.Errorf("opening lots (-want +got):\n%s", diff)
	}
	if len(d.GetTransactions()) != 1 || d.GetTransactions()[0].GetKey().GetExternalId() != "opening-cash:PLN:2026-01-31" {
		t.Errorf("transactions = %v, want only the opening cash on 2026-01-31", d.GetTransactions())
	}
}

// TestSync checks that a sync delivers nothing and keeps the watermark.
func TestSync(t *testing.T) {
	t.Parallel()
	req := connect.NewRequest(&connectorv1.SyncRequest{BatchId: "b-9", Mode: connectorv1.Mode_MODE_TRANSACTIONS, Watermark: "w"})
	req.Header().Set("Authorization", "Bearer "+coreToken)
	resp, err := client(t).Sync(t.Context(), req)
	if err != nil {
		t.Fatalf("Sync: %v", err)
	}
	d := resp.Msg.GetDelivery()
	if resp.Msg.GetWatermark() != "w" || d.GetBatchId() != "b-9" || len(d.GetTransactions())+len(d.GetTickerTransactions()) != 0 {
		t.Errorf("Sync = %v, want an empty delivery of b-9 keeping watermark w", resp.Msg)
	}
}

// TestManifest checks that manifest.json, the image's signed label (D-42),
// asks for nothing but the connector capability and uploaded statements:
// no egress, no Secret.
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
	want := plugins.Manifest{Capabilities: []plugins.Capability{plugins.CapabilityConnector}, StatementFiles: true}
	if diff := cmp.Diff(want, got, cmpopts.EquateEmpty(), cmpopts.IgnoreFields(plugins.Manifest{}, "ConfigSchema")); diff != "" {
		t.Errorf("manifest (-want +got):\n%s", diff)
	}
}
