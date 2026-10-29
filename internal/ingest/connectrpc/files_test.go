package connectrpc_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	"connectrpc.com/connect"

	"github.com/arkadiuss/openfief/internal/ingest"
	"github.com/arkadiuss/openfief/internal/ingest/app"
	"github.com/arkadiuss/openfief/internal/ingest/connectrpc"
	connectorv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/connector/v1"
	"github.com/arkadiuss/openfief/sdk/go/gen/openfief/connector/v1/connectorv1connect"
)

var (
	// errNotAStatement is what fileRefuser answers every file.
	errNotAStatement = errors.New("synthetic: not a statement this connector reads")
	// errBrokenFile is what a brokenFile returns after its bytes.
	errBrokenFile = errors.New("synthetic: the upload broke off")
)

// fileRefuser reads each file whole, then answers with code and err.
type fileRefuser struct {
	connectorv1connect.UnimplementedConnectorServiceHandler

	code connect.Code
	err  error
}

func (f fileRefuser) ImportFile(_ context.Context, stream *connect.ClientStream[connectorv1.ImportFileRequest]) (*connect.Response[connectorv1.ImportFileResponse], error) {
	for stream.Receive() {
		_ = stream.Msg()
	}
	return nil, refuse(f.code, f.err)
}

// brokenFile returns its bytes, then errBrokenFile.
type brokenFile struct {
	r io.Reader
}

func (b *brokenFile) Read(p []byte) (int, error) {
	n, err := b.r.Read(p)
	if err != nil {
		return n, errBrokenFile
	}
	return n, nil
}

func TestClientImportFileRefusals(t *testing.T) {
	t.Parallel()
	file := func() app.NamedFile { return app.NamedFile{Name: "s.xml", Contents: strings.NewReader("synthetic")} }
	ctx := t.Context()
	_, err := newClient(t, fileRefuser{code: connect.CodeInvalidArgument, err: errNotAStatement}).ImportFile(ctx, "flex-main", "b-1", ingest.ModeTransactions, file())
	if !errors.Is(err, app.ErrEntryRefused) || !strings.Contains(err.Error(), errNotAStatement.Error()) {
		t.Errorf("a refused file: %v; want ErrEntryRefused with the connector's reason", err)
	}
	if _, err = newClient(t, failingConnector{}).ImportFile(ctx, "bonds-main", "b-1", ingest.ModeTransactions, file()); !errors.Is(err, app.ErrNoEntries) {
		t.Errorf("a connector without files: %v; want ErrNoEntries", err)
	}
	if _, err = newClient(t, fileRefuser{code: connect.CodeUnavailable, err: errBusy}).ImportFile(ctx, "flex-main", "b-1", ingest.ModeTransactions, file()); !errors.Is(err, app.ErrConnectorUnavailable) {
		t.Errorf("a connector answering UNAVAILABLE: %v; want ErrConnectorUnavailable", err)
	}
	if _, err = newClient(t, &fakeConnector{}).ImportFile(ctx, "nowhere", "b-1", ingest.ModeTransactions, file()); !errors.Is(err, app.ErrConnectorUnavailable) {
		t.Errorf("an instance without an address: %v; want ErrConnectorUnavailable", err)
	}
}

// TestClientImportFileStopsOnReadError checks that a file that breaks off
// after more than one chunk ends the call with the read's error, and the
// connector never sees the end of the file, so it maps nothing.
func TestClientImportFileStopsOnReadError(t *testing.T) {
	t.Parallel()
	conn := &fakeConnector{}
	contents := &brokenFile{r: bytes.NewReader(bytes.Repeat([]byte("x"), 3*connectrpc.ChunkSize))}
	_, err := newClient(t, conn).ImportFile(t.Context(), "flex-main", "b-1", ingest.ModeTransactions, app.NamedFile{Name: "s.xml", Contents: contents})
	if !errors.Is(err, errBrokenFile) {
		t.Errorf("ImportFile = %v, want the read's error", err)
	}
	conn.mu.Lock()
	defer conn.mu.Unlock()
	if conn.header != nil || len(conn.auths) != 0 {
		t.Errorf("the connector completed a call with %d bytes; want the call stopped", len(conn.file))
	}
}
