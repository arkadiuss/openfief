package connectapi_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync"
	"testing"

	"connectrpc.com/connect"
	"github.com/google/go-cmp/cmp"
	"google.golang.org/protobuf/testing/protocmp"

	"github.com/arkadiuss/openfief/internal/ingest"
	"github.com/arkadiuss/openfief/internal/ingest/app"
	"github.com/arkadiuss/openfief/internal/ingest/connectapi"
	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
	ingestv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/ingest/v1"
	"github.com/arkadiuss/openfief/sdk/go/gen/openfief/ingest/v1/ingestv1connect"
)

// flexAccount is the synthetic account statements are uploaded for.
const flexAccount = "acc-flex"

// syntheticStatement is a synthetic statement: an XML declaration, then
// enough rows to need several chunks.
func syntheticStatement() []byte {
	return []byte(`<?xml version="1.0" encoding="UTF-8"?>` + "\n<Statement>" +
		strings.Repeat(`<Row id="synthetic"/>`, 10000) + "</Statement>\n")
}

// fileCaller is a connector that reads each file whole and delivers one
// deposit in its account and problems ingest problems, or fails with err.
type fileCaller struct {
	caller

	account  ledger.AccountID
	problems int
	mu       sync.Mutex
	name     string
	media    string
	file     []byte
	mode     ingest.Mode
}

func (f *fileCaller) ImportFile(_ context.Context, _ string, _ ingest.BatchID, mode ingest.Mode, file app.NamedFile) (app.Delivery, error) {
	b, err := io.ReadAll(file.Contents)
	if err != nil {
		return app.Delivery{}, fmt.Errorf("fake connector: %w", err)
	}
	f.mu.Lock()
	f.name, f.media, f.file, f.mode = file.Name, file.MediaType, b, mode
	f.mu.Unlock()
	if f.err != nil {
		return app.Delivery{}, f.err
	}
	day, err := kernel.ParseDate("2026-03-13")
	if err != nil {
		return app.Delivery{}, fmt.Errorf("fake connector: %w", err)
	}
	currency, err := kernel.ParseCurrency("PLN")
	if err != nil {
		return app.Delivery{}, fmt.Errorf("fake connector: %w", err)
	}
	pln, err := kernel.ParseMoney("100", currency)
	if err != nil {
		return app.Delivery{}, fmt.Errorf("fake connector: %w", err)
	}
	t, err := ledger.NewDeposit(ledger.TransactionHeader{
		Key: ledger.TransactionKey{Source: "flex", ExternalID: "1"}, Account: f.account, Date: day,
	}, pln)
	if err != nil {
		return app.Delivery{}, fmt.Errorf("fake connector: %w", err)
	}
	m := ingest.Mapping{Transactions: []ledger.Transaction{t}}
	for i := range f.problems {
		m.Problems = append(m.Problems, ingest.Skipped(fmt.Sprintf("row %d", i+1), "summary row"))
	}
	return app.Delivery{Institution: "ibkr", Mapping: m}, nil
}

// fileBindings binds flex-main to acc-flex only.
type fileBindings struct{}

func (fileBindings) BoundInstances(_ context.Context, account ledger.AccountID) ([]string, error) {
	if account == flexAccount {
		return []string{"flex-main"}, nil
	}
	return nil, nil
}

func (fileBindings) BoundAccounts(context.Context, string) ([]ledger.AccountID, error) {
	return []ledger.AccountID{flexAccount}, nil
}

// flexInstance runs flex-main, which declares statement files (D-109).
type flexInstance struct{}

func (flexInstance) ConnectorInstances(context.Context) ([]string, error) {
	return []string{"flex-main"}, nil
}

func (flexInstance) Declarations(context.Context, string) (ingest.Declarations, error) {
	return ingest.Declarations{StatementFiles: true}, nil
}

// viewStreams admits streams bearing viewCaller, like cmd/core's
// view-plugin check.
type viewStreams struct{}

func (viewStreams) WrapUnary(next connect.UnaryFunc) connect.UnaryFunc { return next }

func (viewStreams) WrapStreamingClient(next connect.StreamingClientFunc) connect.StreamingClientFunc {
	return next
}

func (viewStreams) WrapStreamingHandler(next connect.StreamingHandlerFunc) connect.StreamingHandlerFunc {
	return func(ctx context.Context, conn connect.StreamingHandlerConn) error {
		if conn.RequestHeader().Get("Authorization") != "Bearer "+viewCaller {
			return connect.NewError(connect.CodePermissionDenied, errNotView)
		}
		return next(ctx, conn)
	}
}

// statementsClient is the generated client, as a struct so helpers return
// no interface.
type statementsClient struct {
	ingestv1connect.StatementServiceClient
}

// newStatementsClient serves the service over c and l, with a 300 000-byte
// limit and real content sniffing, and returns a client of it.
func newStatementsClient(t *testing.T, c *fileCaller, l importer) statementsClient {
	t.Helper()
	day, err := kernel.ParseDate("2026-03-14")
	if err != nil {
		t.Fatalf("ParseDate: %v", err)
	}
	uploads := &app.Uploads{
		Owners: owners{}, Bindings: fileBindings{}, Batches: recorder{}, MaxBytes: 300000, Sniff: http.DetectContentType,
		Connectors: &app.Connectors{
			Caller: c, Bindings: fileBindings{}, Ledger: l, Today: func() kernel.Date { return day }, Declarations: flexInstance{},
		},
		NewBatchID: func() ingest.BatchID { return "b-1" }, Instances: flexInstance{},
	}
	_, handler := connectapi.NewStatements(uploads, viewStreams{}, slog.New(slog.DiscardHandler))
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	return statementsClient{ingestv1connect.NewStatementServiceClient(srv.Client(), srv.URL)}
}

// header is the header of an owner's upload of statement.xml in
// transactions mode.
func header() *ingestv1.StatementHeader {
	return &ingestv1.StatementHeader{
		Origin: local, SessionToken: ownerSession, AccountId: flexAccount, FileName: "C:/exports/statement.xml",
		Mode: ingestv1.Mode_MODE_TRANSACTIONS,
	}
}

// headerPart and chunkPart are the two kinds of upload message.
func headerPart(h *ingestv1.StatementHeader) *ingestv1.UploadStatementRequest {
	return &ingestv1.UploadStatementRequest{Part: &ingestv1.UploadStatementRequest_Header{Header: h}}
}

func chunkPart(b []byte) *ingestv1.UploadStatementRequest {
	return &ingestv1.UploadStatementRequest{Part: &ingestv1.UploadStatementRequest_Chunk{Chunk: b}}
}

// upload sends parts as caller, then file in chunks of 64 KiB, and returns
// the answer.
func upload(t *testing.T, c statementsClient, callerToken string, parts []*ingestv1.UploadStatementRequest, file []byte) (*ingestv1.UploadStatementResponse, error) {
	t.Helper()
	stream := c.UploadStatement(t.Context())
	stream.RequestHeader().Set("Authorization", "Bearer "+callerToken)
	for len(file) > 0 {
		n := min(len(file), 64<<10)
		parts = append(parts, chunkPart(file[:n]))
		file = file[n:]
	}
	for _, p := range parts {
		if stream.Send(p) != nil {
			break // the core answered early; CloseAndReceive has its answer
		}
	}
	resp, err := stream.CloseAndReceive()
	if err != nil {
		return nil, fmt.Errorf("upload: %w", err)
	}
	return resp.Msg, nil
}

func TestUploadStatement(t *testing.T) {
	t.Parallel()
	statement := syntheticStatement()
	c := &fileCaller{account: flexAccount, problems: 105}
	got, err := upload(t, newStatementsClient(t, c, importer{}), viewCaller, []*ingestv1.UploadStatementRequest{headerPart(header())}, statement)
	if err != nil {
		t.Fatalf("UploadStatement: %v", err)
	}
	want := &ingestv1.UploadStatementResponse{
		BatchId: "b-1", Institution: "ibkr", Transactions: 1, New: 1, ProblemsTotal: 105,
		Mode: ingestv1.Mode_MODE_TRANSACTIONS, AccountIds: []string{flexAccount},
	}
	for i := range 100 {
		want.Problems = append(want.Problems, &ingestv1.Problem{
			Kind: ingestv1.ProblemKind_PROBLEM_KIND_SKIPPED, Ref: fmt.Sprintf("row %d", i+1), Reason: "summary row",
		})
	}
	if diff := cmp.Diff(want, got, protocmp.Transform()); diff != "" {
		t.Errorf("UploadStatement (-want +got):\n%s", diff)
	}
	handed := []string{c.name, c.media, c.mode.String(), strconv.FormatBool(bytes.Equal(c.file, statement))}
	if diff := cmp.Diff([]string{"statement.xml", "text/xml; charset=utf-8", "transactions", "true"}, handed); diff != "" {
		t.Errorf("handed to the connector: name, sniffed type, mode, every byte (-want +got):\n%s", diff)
	}
}

// TestListStatementAccounts checks that the owner learns which accounts
// take statement files, with the mode the core chooses for the next upload,
// and that an upload without a mode takes it (D-109).
func TestListStatementAccounts(t *testing.T) {
	t.Parallel()
	c := &fileCaller{account: flexAccount}
	client := newStatementsClient(t, c, importer{})
	got, err := client.ListStatementAccounts(t.Context(), connect.NewRequest(&ingestv1.ListStatementAccountsRequest{Origin: local, SessionToken: ownerSession}))
	if err != nil {
		t.Fatalf("ListStatementAccounts: %v", err)
	}
	want := &ingestv1.ListStatementAccountsResponse{Accounts: []*ingestv1.StatementAccount{
		{AccountId: flexAccount, Instance: "flex-main", Mode: ingestv1.Mode_MODE_OPENING},
	}}
	if diff := cmp.Diff(want, got.Msg, protocmp.Transform()); diff != "" {
		t.Errorf("ListStatementAccounts (-want +got):\n%s", diff)
	}
	_, err = client.ListStatementAccounts(t.Context(), connect.NewRequest(&ingestv1.ListStatementAccountsRequest{Origin: local, SessionToken: "ofsess_unknown"}))
	if connect.CodeOf(err) != connect.CodeUnauthenticated {
		t.Errorf("ListStatementAccounts with an unknown session = %v, want unauthenticated", err)
	}
	h := header()
	h.Mode = ingestv1.Mode_MODE_UNSPECIFIED
	up, err := upload(t, client, viewCaller, []*ingestv1.UploadStatementRequest{headerPart(h)}, syntheticStatement())
	if err != nil || up.GetMode() != ingestv1.Mode_MODE_OPENING || c.mode != ingest.ModeOpening {
		t.Errorf("an upload without a mode = %v, %v, handed %s; want opening, the account not being seeded", up, err, c.mode)
	}
}

func TestUploadStatementRefusals(t *testing.T) {
	t.Parallel()
	statement := syntheticStatement()
	edit := func(change func(h *ingestv1.StatementHeader)) []*ingestv1.UploadStatementRequest {
		h := header()
		change(h)
		return []*ingestv1.UploadStatementRequest{headerPart(h)}
	}
	tests := []struct {
		name     string
		caller   string
		parts    []*ingestv1.UploadStatementRequest
		file     []byte
		callErr  error
		ledger   error
		want     connect.Code
		wantText string
	}{
		{name: "not a view plugin", caller: "nope", want: connect.CodePermissionDenied},
		{name: "no session", parts: edit(func(h *ingestv1.StatementHeader) { h.SessionToken = "ofsess_unknown" }), want: connect.CodeUnauthenticated},
		{name: "a read-scoped session", parts: edit(func(h *ingestv1.StatementHeader) { h.SessionToken = readSession }), want: connect.CodePermissionDenied},
		{name: "no header", parts: []*ingestv1.UploadStatementRequest{}, file: statement, want: connect.CodeInvalidArgument, wantText: "header"},
		{name: "no message at all", parts: []*ingestv1.UploadStatementRequest{}, file: []byte{}, want: connect.CodeInvalidArgument, wantText: "header"},
		{name: "a second header", parts: []*ingestv1.UploadStatementRequest{headerPart(header()), chunkPart([]byte("x")), headerPart(header())}, want: connect.CodeInvalidArgument, wantText: "second header"},
		{name: "an empty file", file: []byte{}, want: connect.CodeInvalidArgument, wantText: "empty"},
		{name: "a file over the limit", file: bytes.Repeat([]byte("x"), 300001), want: connect.CodeResourceExhausted, wantText: "too large"},
		{name: "a message over 1 MiB", parts: []*ingestv1.UploadStatementRequest{headerPart(header()), chunkPart(bytes.Repeat([]byte("x"), 2<<20))}, want: connect.CodeResourceExhausted},
		{name: "an account without a connector", parts: edit(func(h *ingestv1.StatementHeader) { h.AccountId = "acc-other" }), want: connect.CodeFailedPrecondition, wantText: "acc-other"},
		{name: "the connector refuses the file", callErr: fmt.Errorf("%w: synthetic: not a Flex statement", app.ErrEntryRefused), want: connect.CodeInvalidArgument, wantText: "not a Flex statement"},
		{name: "the connector takes no files", callErr: app.ErrNoEntries, want: connect.CodeFailedPrecondition},
		{name: "the connector can't be reached", callErr: app.ErrConnectorUnavailable, want: connect.CodeUnavailable},
		{name: "an opening statement of a seeded account", ledger: &ledger.AlreadySeededError{Account: flexAccount}, want: connect.CodeFailedPrecondition, wantText: "already seeded"},
		{name: "the ledger fails", ledger: errDown, want: connect.CodeInternal},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			parts := tt.parts
			if parts == nil {
				parts = []*ingestv1.UploadStatementRequest{headerPart(header())}
			}
			file := tt.file
			if file == nil {
				file = statement
			}
			c := &fileCaller{account: flexAccount}
			c.err = tt.callErr
			_, err := upload(t, newStatementsClient(t, c, importer{err: tt.ledger}), or(tt.caller, viewCaller), parts, file)
			if connect.CodeOf(err) != tt.want || !strings.Contains(fmt.Sprint(err), tt.wantText) || strings.Contains(fmt.Sprint(err), errDown.Error()) {
				t.Errorf("UploadStatement = %v, want %s with %q and never an internal cause", err, tt.want, tt.wantText)
			}
		})
	}
}
