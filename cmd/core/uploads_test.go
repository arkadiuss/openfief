package main

import (
	"bytes"
	"context"
	"fmt"
	"net/http/httptest"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"connectrpc.com/connect"
	"github.com/google/go-cmp/cmp"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/protobuf/testing/protocmp"

	"github.com/arkadiuss/openfief/internal/demo"
	"github.com/arkadiuss/openfief/internal/ingest"
	"github.com/arkadiuss/openfief/internal/ingest/connectrpc"
	"github.com/arkadiuss/openfief/internal/ingest/delivery"
	"github.com/arkadiuss/openfief/internal/ingest/ibkrflex"
	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
	connectorv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/connector/v1"
	"github.com/arkadiuss/openfief/sdk/go/gen/openfief/connector/v1/connectorv1connect"
	ingestv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/ingest/v1"
	"github.com/arkadiuss/openfief/sdk/go/gen/openfief/ingest/v1/ingestv1connect"
	viewsv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/views/v1"
	"github.com/arkadiuss/openfief/sdk/go/openfief"
)

// flexConnector maps uploaded Flex statements as the ibkr-flex plugin does,
// with the synthetic IBKR account U0000000 feeding the demo's IBKR account.
type flexConnector struct {
	connectorv1connect.UnimplementedConnectorServiceHandler
}

func (flexConnector) ImportFile(_ context.Context, stream *connect.ClientStream[connectorv1.ImportFileRequest]) (*connect.Response[connectorv1.ImportFileResponse], error) {
	header, data, err := openfief.ReceiveFile(stream, 64<<20)
	if err != nil {
		return nil, refused(err)
	}
	mode, err := delivery.Mode(header.GetMode())
	if err != nil {
		return nil, refused(err)
	}
	m, err := ibkrflex.Mapper{Accounts: map[string]ledger.AccountID{"U0000000": demo.IBKRAccount}}.Map(bytes.NewReader(data), mode)
	if err != nil {
		return nil, refused(err)
	}
	d, err := delivery.FromMapping(header.GetBatchId(), "ibkr", &m)
	if err != nil {
		return nil, refused(err)
	}
	return connect.NewResponse(&connectorv1.ImportFileResponse{Delivery: d}), nil
}

// statementAPI is a generated StatementService client.
type statementAPI struct {
	ingestv1connect.StatementServiceClient
}

// uploadStatement uploads file for account as the view plugin, in
// transactions mode, in 64 KiB chunks, over HTTP/1.1 as the UI does.
func uploadStatement(t *testing.T, c statementAPI, session, account string, file []byte) (*ingestv1.UploadStatementResponse, error) {
	t.Helper()
	return uploadStatementIn(t, c, session, account, ingestv1.Mode_MODE_TRANSACTIONS, file)
}

// uploadStatementIn uploads file for account like uploadStatement, with
// mode as the owner's override; unspecified lets the core choose.
func uploadStatementIn(t *testing.T, c statementAPI, session, account string, mode ingestv1.Mode, file []byte) (*ingestv1.UploadStatementResponse, error) {
	t.Helper()
	stream := c.UploadStatement(t.Context())
	stream.RequestHeader().Set("Authorization", "Bearer "+uiCaller)
	parts := []*ingestv1.UploadStatementRequest{{Part: &ingestv1.UploadStatementRequest_Header{Header: &ingestv1.StatementHeader{
		Origin: localUI, SessionToken: session, AccountId: account, FileName: "statement.xml", Mode: mode,
	}}}}
	for len(file) > 0 {
		n := min(len(file), 64<<10)
		parts = append(parts, &ingestv1.UploadStatementRequest{Part: &ingestv1.UploadStatementRequest_Chunk{Chunk: file[:n]}})
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

// flexStatement is the ibkrflex adapter's synthetic activity statement.
func flexStatement(t *testing.T) []byte {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("..", "..", "internal", "ingest", "ibkrflex", "testdata", "activity.xml"))
	if err != nil {
		t.Fatalf("read statement: %v", err)
	}
	return b
}

// openingStatement is the ibkrflex adapter's synthetic opening statement.
func openingStatement(t *testing.T) []byte {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("..", "..", "internal", "ingest", "ibkrflex", "testdata", "opening.xml"))
	if err != nil {
		t.Fatalf("read statement: %v", err)
	}
	return b
}

// flexDiscovery is the cluster as discovery sees it: flex-main runs and
// declares statement files and cross-check prices (D-109, D-112).
type flexDiscovery struct{}

func (flexDiscovery) ConnectorInstances(context.Context) ([]string, error) {
	return []string{"flex-main"}, nil
}

func (flexDiscovery) Declarations(context.Context, string) (ingest.Declarations, error) {
	return ingest.Declarations{StatementFiles: true, CrossCheckPrices: true}, nil
}

// statementAccounts lists the accounts that take files, as the UI asks.
func statementAccounts(t *testing.T, c statementAPI, session string) []*ingestv1.StatementAccount {
	t.Helper()
	req := connect.NewRequest(&ingestv1.ListStatementAccountsRequest{Origin: localUI, SessionToken: session})
	req.Header().Set("Authorization", "Bearer "+uiCaller)
	got, err := c.ListStatementAccounts(t.Context(), req)
	if err != nil {
		t.Fatalf("ListStatementAccounts: %v", err)
	}
	return got.Msg.GetAccounts()
}

// checkOpeningUpload checks that the IBKR account takes files in opening
// mode while it isn't seeded, that an upload with no mode seeds it from
// the opening statement, and that it then takes transactions, uploads and
// scheduled syncs alike (D-109, D-95).
func checkOpeningUpload(t *testing.T, c *core, api statementAPI, session string) {
	t.Helper()
	want := []*ingestv1.StatementAccount{{AccountId: string(demo.IBKRAccount), Instance: "flex-main", Mode: ingestv1.Mode_MODE_OPENING}}
	if diff := cmp.Diff(want, statementAccounts(t, api, session), protocmp.Transform()); diff != "" {
		t.Errorf("accounts that take files before seeding (-want +got):\n%s", diff)
	}
	got, err := uploadStatementIn(t, api, session, string(demo.IBKRAccount), ingestv1.Mode_MODE_UNSPECIFIED, openingStatement(t))
	if err != nil || got.GetMode() != ingestv1.Mode_MODE_OPENING || got.GetNew() == 0 ||
		!slices.Equal(got.GetAccountIds(), []string{string(demo.IBKRAccount)}) {
		t.Fatalf("an upload with no mode = %v, %v; want an opening batch that seeds the IBKR account", got, err)
	}
	want[0].Mode = ingestv1.Mode_MODE_TRANSACTIONS
	if diff := cmp.Diff(want, statementAccounts(t, api, session), protocmp.Transform()); diff != "" {
		t.Errorf("accounts that take files after seeding (-want +got):\n%s", diff)
	}
	goLive, seeded, err := c.deps.ledger.GoLive(t.Context(), demo.IBKRAccount)
	if err != nil || !seeded {
		t.Fatalf("GoLive = %v, %v; want the account seeded", seeded, err)
	}
	bound := []ledger.AccountID{demo.IBKRAccount}
	plan := ingest.PlanSync(bound, map[ledger.AccountID]kernel.Date{demo.IBKRAccount: goLive})
	if !slices.Equal(plan, []ingest.Mode{ingest.ModeTransactions}) {
		t.Errorf("the scheduled sync after an opening upload plans %v, want transactions only (D-109)", plan)
	}
}

func TestStatementService(t *testing.T) {
	t.Parallel()
	dsn := emptyDatabase(t)
	code := enrollCode(t, enrollCmd(t, dsn))
	c, srv, pool := startCoreForUI(t, dsn, func(*serveConfig) {})
	ac := authClient(srv.URL, srv.Client(), uiCaller)
	session := passkeyLogin(t, ac, passkeyEnroll(t, ac, code))
	saveDemoRegistry(t, dsn)
	if _, err := connectorCmd(t, dsn, nil, "bind", "--instance=flex-main", "--account="+string(demo.IBKRAccount)); err != nil {
		t.Fatalf("connector bind: %v", err)
	}
	api := statementAPI{ingestv1connect.NewStatementServiceClient(srv.Client(), srv.URL)}
	statement := flexStatement(t)
	if _, err := uploadStatement(t, api, session, string(demo.IBKRAccount), statement); connect.CodeOf(err) != connect.CodeUnavailable {
		t.Fatalf("UploadStatement before the core found any connector = %v, want Unavailable", err)
	}
	if got := statementAccounts(t, api, session); len(got) != 0 {
		t.Errorf("accounts that take files before the core found any connector = %v, want none", got)
	}
	port := serveConnector(t, flexConnector{})
	c.deps.entries.set(connectrpc.NewClient(connectrpc.H2CClient(), func(string) (string, error) {
		return fmt.Sprintf("http://127.0.0.1:%d", port), nil
	}, staticTokens{}))
	c.deps.entries.setDiscovery(flexDiscovery{})
	checkOpeningUpload(t, c, api, session)

	checkUploadRecorded(t, pool, checkUploadTwice(t, api, session, statement)...)
	checkUploadRefusals(t, api, session, statement)
	checkUploadScoped(t, srv, session)
}

// checkUploadTwice uploads statement twice and checks the counts: all new,
// then all unchanged (D-18). It returns the two batch IDs.
func checkUploadTwice(t *testing.T, api statementAPI, session string, statement []byte) []string {
	t.Helper()
	got, err := uploadStatement(t, api, session, string(demo.IBKRAccount), statement)
	if err != nil {
		t.Fatalf("UploadStatement: %v", err)
	}
	// activity.transactions.golden: ten transactions, four assertions,
	// thirteen ingest problems (one is the LOT rows' disagreeing mark prices, D-22).
	counts := []int{int(got.GetTransactions()), int(got.GetNew()), int(got.GetAssertions()), int(got.GetProblemsTotal()), len(got.GetProblems())}
	if diff := cmp.Diff([]int{10, 10, 4, 13, 13}, counts); diff != "" || got.GetInstitution() != "ibkr" {
		t.Errorf("UploadStatement of %s: transactions, new, assertions, problems, listed (-want +got):\n%s", got.GetInstitution(), diff)
	}
	again, err := uploadStatement(t, api, session, string(demo.IBKRAccount), statement)
	if err != nil || again.GetUnchanged() != 10 || again.GetNew() != 0 {
		t.Errorf("uploading the statement again = %v, %v; want all 10 unchanged (D-18)", again, err)
	}
	return []string{got.GetBatchId(), again.GetBatchId()}
}

// checkUploadRecorded checks that both uploads are in the batch history as
// completed upload batches of flex-main.
func checkUploadRecorded(t *testing.T, pool *pgxpool.Pool, ids ...string) {
	t.Helper()
	var n int
	if err := pool.QueryRow(t.Context(), `
		SELECT count(*) FROM ingest.import_batches
		WHERE id = ANY($1) AND instance = 'flex-main' AND source = 'upload' AND state = 'completed' AND transactions = 10`,
		ids).Scan(&n); err != nil {
		t.Fatalf("loading the batch history: %v", err)
	}
	if n != len(ids) {
		t.Errorf("%d of the batches %v are recorded as completed uploads of flex-main, want all", n, ids)
	}
}

// checkUploadRefusals checks that the core refuses an upload from an
// unknown session, for an account no connector is bound to, and of a file
// the connector can't read.
func checkUploadRefusals(t *testing.T, api statementAPI, session string, statement []byte) {
	t.Helper()
	if _, err := uploadStatement(t, api, "ofsess_nope_nope", string(demo.IBKRAccount), statement); connect.CodeOf(err) != connect.CodeUnauthenticated {
		t.Errorf("an unknown session: %v, want Unauthenticated", err)
	}
	if _, err := uploadStatement(t, api, session, string(demo.BondsAccount), statement); connect.CodeOf(err) != connect.CodeFailedPrecondition {
		t.Errorf("an account without a connector: %v, want FailedPrecondition", err)
	}
	if _, err := uploadStatement(t, api, session, string(demo.IBKRAccount), []byte("date,amount\n")); connect.CodeOf(err) != connect.CodeInvalidArgument {
		t.Errorf("a file that isn't a Flex statement: %v, want InvalidArgument", err)
	}
}

// checkUploadScoped checks through the read API that the uploaded deposit
// is in the IBKR account with its key scoped to the instance (D-48).
func checkUploadScoped(t *testing.T, srv *httptest.Server, session string) {
	t.Helper()
	var got viewsv1.ListTransactionsResponse
	getJSON(t, srv, "/v1/accounts/"+string(demo.IBKRAccount)+"/transactions?page_size=100", session, &got)
	for _, tx := range got.GetTransactions() {
		if tx.GetExternalId() == "8000101" {
			if tx.GetSource() != "plugin:flex-main:ibkr-flex:"+string(demo.IBKRAccount) {
				t.Errorf("the deposit's key source is %q, want it scoped to flex-main", tx.GetSource())
			}
			return
		}
	}
	t.Errorf("the uploaded deposit 8000101 is not among the account's %d transactions", len(got.GetTransactions()))
}
