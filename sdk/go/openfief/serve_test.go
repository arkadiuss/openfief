package openfief_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"connectrpc.com/connect"

	connectorv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/connector/v1"
	"github.com/arkadiuss/openfief/sdk/go/gen/openfief/connector/v1/connectorv1connect"
	"github.com/arkadiuss/openfief/sdk/go/openfief"
)

const coreToken = "synthetic-core-token"

var errNotThisCore = fmt.Errorf("fake verifier: %w", openfief.ErrNotCore)

// fakeVerifier accepts only coreToken.
type fakeVerifier struct{}

func (fakeVerifier) VerifyCore(_ context.Context, token string) error {
	if token != coreToken {
		return errNotThisCore
	}
	return nil
}

// fakeConnector echoes the batch and reports the file it received.
type fakeConnector struct {
	connectorv1connect.UnimplementedConnectorServiceHandler

	mu    sync.Mutex
	calls int
	limit int
}

func (f *fakeConnector) Sync(_ context.Context, req *connect.Request[connectorv1.SyncRequest]) (*connect.Response[connectorv1.SyncResponse], error) {
	f.mu.Lock()
	f.calls++
	f.mu.Unlock()
	return connect.NewResponse(&connectorv1.SyncResponse{
		Delivery:  &connectorv1.Delivery{BatchId: req.Msg.GetBatchId(), Institution: "test"},
		Watermark: req.Msg.GetWatermark() + "+1",
	}), nil
}

func (f *fakeConnector) ImportFile(_ context.Context, stream *connect.ClientStream[connectorv1.ImportFileRequest]) (*connect.Response[connectorv1.ImportFileResponse], error) {
	f.mu.Lock()
	f.calls++
	f.mu.Unlock()
	header, data, err := openfief.ReceiveFile(stream, f.limit)
	if err != nil {
		return nil, fmt.Errorf("fake connector: %w", connect.NewError(connect.CodeInvalidArgument, err))
	}
	return connect.NewResponse(&connectorv1.ImportFileResponse{Delivery: &connectorv1.Delivery{
		BatchId: header.GetBatchId(), Institution: header.GetFileName() + ":" + string(data),
	}}), nil
}

func (f *fakeConnector) callCount() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.calls
}

// testConnector is a served fakeConnector and a client of it.
type testConnector struct {
	*fakeConnector

	client connectorv1connect.ConnectorServiceClient
}

// startConnector serves a fakeConnector with NewConnectorServer's handler
// over HTTP/2 without TLS.
func startConnector(t *testing.T, limit int) testConnector {
	t.Helper()
	conn := &fakeConnector{limit: limit}
	server := openfief.NewConnectorServer("", conn, fakeVerifier{})
	srv := httptest.NewUnstartedServer(server.Handler)
	srv.Config.Protocols = server.Protocols
	srv.Start()
	t.Cleanup(srv.Close)
	var p http.Protocols
	p.SetUnencryptedHTTP2(true)
	client := &http.Client{Transport: &http.Transport{Protocols: &p}}
	return testConnector{fakeConnector: conn, client: connectorv1connect.NewConnectorServiceClient(client, srv.URL)}
}

func TestConnectorServerRequiresCore(t *testing.T) {
	t.Parallel()
	for _, auth := range []string{"", "Bearer synthetic-other-token", "Basic " + coreToken, "Bearer"} {
		conn := startConnector(t, 1)
		req := connect.NewRequest(&connectorv1.SyncRequest{BatchId: "b-1"})
		if auth != "" {
			req.Header().Set("Authorization", auth)
		}
		_, err := conn.client.Sync(t.Context(), req)
		if got := connect.CodeOf(err); got != connect.CodeUnauthenticated {
			t.Errorf("Authorization %q: code = %v, want %v", auth, got, connect.CodeUnauthenticated)
		}
		if conn.callCount() != 0 {
			t.Errorf("Authorization %q reached the connector", auth)
		}
	}
}

func TestConnectorServerSync(t *testing.T) {
	t.Parallel()
	client := startConnector(t, 1).client
	req := connect.NewRequest(&connectorv1.SyncRequest{BatchId: "b-1", Watermark: "w"})
	req.Header().Set("Authorization", "Bearer "+coreToken)
	resp, err := client.Sync(t.Context(), req)
	if err != nil {
		t.Fatalf("Sync: %v", err)
	}
	if resp.Msg.GetDelivery().GetBatchId() != "b-1" || resp.Msg.GetWatermark() != "w+1" {
		t.Errorf("Sync = %v, want batch b-1 and watermark w+1", resp.Msg)
	}
}

// importFile streams a header (unless noHeader) and chunks to client as the
// core.
func importFile(t *testing.T, client connectorv1connect.ConnectorServiceClient, noHeader bool, chunks ...string) (*connectorv1.ImportFileResponse, error) {
	t.Helper()
	stream := client.ImportFile(t.Context())
	stream.RequestHeader().Set("Authorization", "Bearer "+coreToken)
	var msgs []*connectorv1.ImportFileRequest
	if !noHeader {
		msgs = append(msgs, &connectorv1.ImportFileRequest{Part: &connectorv1.ImportFileRequest_Header{
			Header: &connectorv1.FileHeader{BatchId: "b-2", FileName: "statement.csv"},
		}})
	}
	for _, c := range chunks {
		msgs = append(msgs, &connectorv1.ImportFileRequest{Part: &connectorv1.ImportFileRequest_Chunk{Chunk: []byte(c)}})
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

func TestReceiveFile(t *testing.T) {
	t.Parallel()
	client := startConnector(t, 8).client
	got, err := importFile(t, client, false, "a,b", "\n1,2")
	if err != nil {
		t.Fatalf("ImportFile: %v", err)
	}
	if d := got.GetDelivery(); d.GetBatchId() != "b-2" || d.GetInstitution() != "statement.csv:a,b\n1,2" {
		t.Errorf("ImportFile = %v, want batch b-2 with the file reassembled", d)
	}
	tests := []struct {
		name     string
		noHeader bool
		chunks   []string
	}{
		{name: "too large", chunks: []string{"12345", "6789"}},
		{name: "no header", noHeader: true, chunks: []string{"a"}},
		{name: "empty stream", noHeader: true},
	}
	for _, tt := range tests {
		if _, err := importFile(t, client, tt.noHeader, tt.chunks...); connect.CodeOf(err) != connect.CodeInvalidArgument {
			t.Errorf("%s: error = %v, want InvalidArgument", tt.name, err)
		}
	}
}

func TestInstanceAudience(t *testing.T) {
	t.Parallel()
	if got, want := openfief.InstanceAudience("ibkr-main"), "openfief-plugin:ibkr-main"; got != want {
		t.Errorf("InstanceAudience = %q, want %q", got, want)
	}
}
