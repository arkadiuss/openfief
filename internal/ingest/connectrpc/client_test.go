package connectrpc_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"sync"
	"testing"

	"connectrpc.com/connect"

	"github.com/arkadiuss/openfief/internal/ingest"
	"github.com/arkadiuss/openfief/internal/ingest/app"
	"github.com/arkadiuss/openfief/internal/ingest/connectrpc"
	"github.com/arkadiuss/openfief/internal/ledger"
	connectorv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/connector/v1"
	"github.com/arkadiuss/openfief/sdk/go/gen/openfief/connector/v1/connectorv1connect"
)

var (
	errNoToken   = errors.New("token request failed")
	errNoAddress = errors.New("no such instance")
)

// fakeTokens issues "synthetic-core-<instance>", or fails for "broken".
type fakeTokens struct{}

func (fakeTokens) TokenFor(_ context.Context, instance string) (string, error) {
	if instance == "broken" {
		return "", errNoToken
	}
	return "synthetic-core-" + instance, nil
}

// fakeConnector records what the core sent and delivers one deposit for the
// batch it was asked for; wrongBatch makes it answer for another batch.
type fakeConnector struct {
	connectorv1connect.UnimplementedConnectorServiceHandler

	mu         sync.Mutex
	wrongBatch bool
	auths      []string
	modes      []connectorv1.Mode
	watermark  string
	header     *connectorv1.FileHeader
	file       []byte
	chunks     int
}

func (f *fakeConnector) Sync(_ context.Context, req *connect.Request[connectorv1.SyncRequest]) (*connect.Response[connectorv1.SyncResponse], error) {
	f.record(req.Header().Get("Authorization"), req.Msg.GetMode())
	f.mu.Lock()
	f.watermark = req.Msg.GetWatermark()
	f.mu.Unlock()
	return connect.NewResponse(&connectorv1.SyncResponse{Delivery: f.delivery(req.Msg.GetBatchId()), Watermark: "w-2"}), nil
}

func (f *fakeConnector) ImportFile(_ context.Context, stream *connect.ClientStream[connectorv1.ImportFileRequest]) (*connect.Response[connectorv1.ImportFileResponse], error) {
	var header *connectorv1.FileHeader
	var file []byte
	chunks := 0
	for stream.Receive() {
		if h := stream.Msg().GetHeader(); h != nil {
			header = h
			continue
		}
		chunks++
		file = append(file, stream.Msg().GetChunk()...)
	}
	if err := stream.Err(); err != nil {
		return nil, fmt.Errorf("fake connector: %w", err)
	}
	f.record(stream.RequestHeader().Get("Authorization"), header.GetMode())
	f.mu.Lock()
	f.header, f.file, f.chunks = header, file, chunks
	f.mu.Unlock()
	return connect.NewResponse(&connectorv1.ImportFileResponse{Delivery: f.delivery(header.GetBatchId())}), nil
}

func (f *fakeConnector) record(auth string, mode connectorv1.Mode) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.auths = append(f.auths, auth)
	f.modes = append(f.modes, mode)
}

func (f *fakeConnector) delivery(id string) *connectorv1.Delivery {
	if f.wrongBatch {
		id = "b-other"
	}
	cash := &connectorv1.Book{Kind: connectorv1.BookKind_BOOK_KIND_CASH, Currency: "PLN"}
	equity := &connectorv1.Book{Kind: connectorv1.BookKind_BOOK_KIND_EQUITY, Currency: "PLN"}
	return &connectorv1.Delivery{
		BatchId: id, Institution: "xtb",
		Transactions: []*connectorv1.Transaction{{
			Key:       &connectorv1.TransactionKey{Source: "acc-1", ExternalId: "D1"},
			AccountId: "acc-1", Kind: connectorv1.TransactionKind_TRANSACTION_KIND_DEPOSIT, Date: "2026-03-02",
			Postings: []*connectorv1.Posting{
				{Book: cash, Amount: &connectorv1.Money{Amount: "1000.00", Currency: "PLN"}},
				{Book: equity, Amount: &connectorv1.Money{Amount: "-1000.00", Currency: "PLN"}},
			},
		}},
	}
}

// failingConnector refuses every call.
type failingConnector struct {
	connectorv1connect.UnimplementedConnectorServiceHandler
}

// newClient serves conn over HTTP/2 without TLS and returns a core client
// that reaches every instance but "nowhere" there.
func newClient(t *testing.T, conn connectorv1connect.ConnectorServiceHandler) *connectrpc.Client {
	t.Helper()
	mux := http.NewServeMux()
	mux.Handle(connectorv1connect.NewConnectorServiceHandler(conn))
	srv := httptest.NewUnstartedServer(mux)
	var p http.Protocols
	p.SetHTTP1(true)
	p.SetUnencryptedHTTP2(true)
	srv.Config.Protocols = &p
	srv.Start()
	t.Cleanup(srv.Close)
	baseURL := func(instance string) (string, error) {
		if instance == "nowhere" {
			return "", errNoAddress
		}
		return srv.URL, nil
	}
	return connectrpc.NewClient(connectrpc.H2CClient(), baseURL, fakeTokens{})
}

func TestClientSync(t *testing.T) {
	t.Parallel()
	conn := &fakeConnector{}
	c := newClient(t, conn)
	got, err := c.Sync(t.Context(), "xtb-main", "b-1", ingest.ModeOpening, "w-1")
	if err != nil {
		t.Fatalf("Sync: %v", err)
	}
	if got.Institution != "xtb" || got.Watermark != "w-2" || len(got.Mapping.Transactions) != 1 {
		t.Errorf("Sync = %+v, want one xtb transaction and watermark w-2", got)
	}
	if k := got.Mapping.Transactions[0].Key; k != (ledger.TransactionKey{Source: "acc-1", ExternalID: "D1"}) {
		t.Errorf("key = %v; the client must not scope keys, the use case does", k)
	}
	if conn.auths[0] != "Bearer synthetic-core-xtb-main" || conn.modes[0] != connectorv1.Mode_MODE_OPENING || conn.watermark != "w-1" {
		t.Errorf("connector got auth %q, mode %v, watermark %q", conn.auths[0], conn.modes[0], conn.watermark)
	}
}

func TestClientImportFile(t *testing.T) {
	t.Parallel()
	conn := &fakeConnector{}
	c := newClient(t, conn)
	contents := bytes.Repeat([]byte("synthetic,row\n"), connectrpc.ChunkSize/7)
	file := app.NamedFile{Name: "statement.csv", MediaType: "text/csv", Contents: bytes.NewReader(contents)}
	got, err := c.ImportFile(t.Context(), "xtb-main", "b-2", ingest.ModeTransactions, file)
	if err != nil {
		t.Fatalf("ImportFile: %v", err)
	}
	if got.Institution != "xtb" || len(got.Mapping.Transactions) != 1 {
		t.Errorf("ImportFile = %+v, want one xtb transaction", got)
	}
	h := conn.header
	gotHeader := []string{h.GetBatchId(), h.GetFileName(), h.GetMediaType(), h.GetMode().String()}
	if want := []string{"b-2", "statement.csv", "text/csv", "MODE_TRANSACTIONS"}; !slices.Equal(gotHeader, want) {
		t.Errorf("header = %q, want %q", gotHeader, want)
	}
	if !bytes.Equal(conn.file, contents) || conn.chunks < 2 {
		t.Errorf("connector got %d bytes in %d chunks, want %d bytes in several", len(conn.file), conn.chunks, len(contents))
	}
	if conn.auths[0] != "Bearer synthetic-core-xtb-main" {
		t.Errorf("connector got auth %q", conn.auths[0])
	}
}

func TestClientFailures(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		conn     connectorv1connect.ConnectorServiceHandler
		instance string
		mode     ingest.Mode
		wantErr  error
		wantCode connect.Code
	}{
		{name: "no address", conn: &fakeConnector{}, instance: "nowhere", mode: ingest.ModeOpening, wantErr: errNoAddress},
		{name: "no token", conn: &fakeConnector{}, instance: "broken", mode: ingest.ModeOpening, wantErr: errNoToken},
		{name: "no mode", conn: &fakeConnector{}, instance: "xtb-main", wantErr: ingest.ErrInvalidBatch},
		{name: "another batch delivered", conn: &fakeConnector{wrongBatch: true}, instance: "xtb-main", mode: ingest.ModeOpening, wantErr: connectrpc.ErrInvalidMessage},
		{name: "connector refuses", conn: failingConnector{}, instance: "xtb-main", mode: ingest.ModeOpening, wantCode: connect.CodeUnimplemented},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			c := newClient(t, tt.conn)
			_, syncErr := c.Sync(t.Context(), tt.instance, "b-1", tt.mode, "")
			file := app.NamedFile{Name: "f.csv", Contents: strings.NewReader("synthetic")}
			_, importErr := c.ImportFile(t.Context(), tt.instance, "b-1", tt.mode, file)
			for _, err := range []error{syncErr, importErr} {
				if tt.wantErr != nil && !errors.Is(err, tt.wantErr) {
					t.Errorf("error = %v, want %v", err, tt.wantErr)
				}
				if tt.wantCode != 0 && connect.CodeOf(err) != tt.wantCode {
					t.Errorf("code = %v, want %v", connect.CodeOf(err), tt.wantCode)
				}
			}
		})
	}
}
