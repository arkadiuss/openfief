package main

import (
	"context"
	"fmt"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/google/go-cmp/cmp"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/protobuf/testing/protocmp"

	"github.com/arkadiuss/openfief/internal/demo"
	"github.com/arkadiuss/openfief/internal/ingest"
	"github.com/arkadiuss/openfief/internal/ingest/connectrpc"
	"github.com/arkadiuss/openfief/internal/ingest/delivery"
	ingestpg "github.com/arkadiuss/openfief/internal/ingest/postgres"
	"github.com/arkadiuss/openfief/internal/ingest/xtb"
	"github.com/arkadiuss/openfief/internal/ingest/xtb/xtbtest"
	connectorv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/connector/v1"
	"github.com/arkadiuss/openfief/sdk/go/gen/openfief/connector/v1/connectorv1connect"
	ingestv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/ingest/v1"
	"github.com/arkadiuss/openfief/sdk/go/gen/openfief/ingest/v1/ingestv1connect"
	viewsv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/views/v1"
	"github.com/arkadiuss/openfief/sdk/go/openfief"
)

// sendingBank is the made-up bank a synthetic deposit's comment names.
const sendingBank = "Bank Testowy"

// xtbExport is a synthetic XTB export: a deposit, buys of TESTA.PL and
// TESTB.UK, a sale of TESTA.PL, a dividend of TESTB.UK with tax withheld,
// a withdrawal whose comment names the account, and the 8 units of
// TESTA.PL still open on the report's date, 2026-06-30 (D-155).
func xtbExport() []byte {
	day := func(m time.Month, d int) time.Time { return time.Date(2026, m, d, 10, 0, 0, 0, time.UTC) }
	e := xtbtest.Export{
		Account: xtbtest.Account, Currency: "PLN", From: day(1, 1), To: day(6, 30),
		Cash: [][]xtbtest.Cell{
			xtbtest.CashRow("Withdrawal", "", "", "", day(5, 4), "-100", "9000000006", "Wyplata na rachunek "+xtbtest.Account),
			xtbtest.CashRow("Withholding tax", "Test Beta PLC", "TESTB.UK", "ETF", day(4, 1), "-1.5", "9000000005", "TESTB.UK GBP WHT 15%"),
			xtbtest.CashRow("Dividend", "Test Beta PLC", "TESTB.UK", "ETF", day(4, 1), "10", "9000000004", "TESTB.UK GBP 0.5000/ SHR"),
			xtbtest.CashRow("Stock sell", "Test Alpha SA", "TESTA.PL", "STOCK", day(3, 2), "250", "9000000003", "CLOSE BUY 2/10 @ 125"),
			xtbtest.CashRow("Stock purchase", "Test Beta PLC", "TESTB.UK", "ETF", day(2, 3), "-400.00000000000006", "9000000008", "OPEN BUY 20/20 @ 4.80"),
			xtbtest.CashRow("Stock purchase", "Test Alpha SA", "TESTA.PL", "STOCK", day(2, 2), "-1000", "9000000002", "OPEN BUY 10/10 @ 100"),
			xtbtest.CashRow("Deposit", "", "", "", day(1, 5), "5000", "9000000001", "Przelew z "+sendingBank+" S.A."),
		},
		Open: [][]xtbtest.Cell{
			xtbtest.OpenRow(&xtbtest.Position{Name: "Test Alpha SA", Ticker: "TESTA.PL", Category: "STOCK", Volume: "8", OpenPrice: "100", Opened: day(2, 2)}),
		},
	}
	return e.Bytes()
}

// xtbConnector maps uploaded exports as the xtb plugin does, and, as a
// market-data connector would, proposes an ISIN for TESTA.PL only.
type xtbConnector struct {
	connectorv1connect.UnimplementedConnectorServiceHandler
}

func (xtbConnector) ImportFile(_ context.Context, stream *connect.ClientStream[connectorv1.ImportFileRequest]) (*connect.Response[connectorv1.ImportFileResponse], error) {
	header, data, err := openfief.ReceiveFile(stream, xtb.MaxFileBytes)
	if err != nil {
		return nil, refused(err)
	}
	mode, err := delivery.Mode(header.GetMode())
	if err != nil {
		return nil, refused(err)
	}
	cash, err := delivery.OpeningCash(header)
	if err != nil {
		return nil, refused(err)
	}
	m, err := xtb.Mapper{}.Map(data, header.GetFileName(), mode, cash)
	if err != nil {
		return nil, refused(err)
	}
	d, err := delivery.FromMapping(header.GetBatchId(), "xtb", &m)
	if err != nil {
		return nil, refused(err)
	}
	return connect.NewResponse(&connectorv1.ImportFileResponse{Delivery: d}), nil
}

func (xtbConnector) LookupInstruments(_ context.Context, req *connect.Request[connectorv1.LookupInstrumentsRequest]) (*connect.Response[connectorv1.LookupInstrumentsResponse], error) {
	out := &connectorv1.LookupInstrumentsResponse{}
	for _, t := range req.Msg.GetTickers() {
		if t.GetCode() == "TESTA" && t.GetCountry() == "PL" {
			out.Proposals = append(out.Proposals, &connectorv1.InstrumentProposal{Symbol: t.GetSymbol(), Isin: "PLTESTA00002", Name: "TEST ALPHA"})
		}
	}
	return connect.NewResponse(out), nil
}

// xtbDiscovery is the cluster as discovery sees it: xtb-main takes
// statement files, and eodhd-main is a market-data connector its grant
// approves as one (D-123).
type xtbDiscovery struct{}

func (xtbDiscovery) ConnectorInstances(context.Context) ([]string, error) {
	return []string{"eodhd-main", "xtb-main"}, nil
}

func (xtbDiscovery) Declarations(_ context.Context, instance string) (ingest.Declarations, error) {
	return ingest.Declarations{StatementFiles: instance == "xtb-main", MarketData: instance == "eodhd-main"}, nil
}

// tickerAPI is a generated TickerService client.
type tickerAPI struct {
	ingestv1connect.TickerServiceClient
}

// TestXTBUpload checks D-116 through the real core and Postgres: an XTB
// export's deposits and withdrawals are booked at once, its stock rows wait
// as ingest problems while their tickers' ISINs aren't confirmed, the
// market-data connector's proposal is offered, the owner's confirmation,
// proposed or typed, books the waiting rows, a later upload books them at
// once, and nothing of a transfer's comment is stored.
func TestXTBUpload(t *testing.T) {
	t.Parallel()
	dsn := emptyDatabase(t)
	code := enrollCode(t, enrollCmd(t, dsn))
	c, srv, pool := startCoreForUI(t, dsn, func(*serveConfig) {})
	ac := authClient(srv.URL, srv.Client(), uiCaller)
	session := passkeyLogin(t, ac, passkeyEnroll(t, ac, code))
	saveDemoRegistry(t, dsn)
	if _, err := connectorCmd(t, dsn, nil, "bind", "--instance=xtb-main", "--account="+string(demo.XTBAccount)); err != nil {
		t.Fatalf("connector bind: %v", err)
	}
	port := serveConnector(t, xtbConnector{})
	c.deps.entries.set(connectrpc.NewClient(connectrpc.H2CClient(), func(string) (string, error) {
		return fmt.Sprintf("http://127.0.0.1:%d", port), nil
	}, staticTokens{}))
	c.deps.entries.setDiscovery(xtbDiscovery{})
	api := statementAPI{ingestv1connect.NewStatementServiceClient(srv.Client(), srv.URL)}
	tickers := tickerAPI{ingestv1connect.NewTickerServiceClient(srv.Client(), srv.URL)}
	checkFirstXTBUploads(t, api, pool, session)
	checkPendingTickers(t, tickers, session)
	checkConfirmTickers(t, tickers, session)
	again, err := uploadStatement(t, api, session, string(demo.XTBAccount), xtbExport())
	if err != nil || again.GetUnchanged() != 6 || again.GetNew() != 0 || again.GetProblemsTotal() != 0 {
		t.Errorf("uploading the export again = %v, %v; want all 6 rows booked and unchanged", again, err)
	}
	checkXTBBooked(t, srv, session, "PLTESTA00002")
	checkHoldingAsserted(t, srv, session, "PLTESTA00002")
	checkCorrectTicker(t, tickers, session)
	checkXTBBooked(t, srv, session, "PLTESTC00008")
	checkHoldingAsserted(t, srv, session, "PLTESTC00008")
	checkNoTransferComments(t, pool)
	if _, err := uploadStatementIn(t, api, session, string(demo.XTBAccount), ingestv1.Mode_MODE_OPENING, xtbExport()); connect.CodeOf(err) != connect.CodeFailedPrecondition ||
		!strings.Contains(err.Error(), "whole history") {
		t.Errorf("an opening upload of an account started from its whole history = %v, want FailedPrecondition", err)
	}
}

// checkFirstXTBUploads uploads the export for the XTB account while its
// broker account isn't mapped, which books nothing and names it unmapped,
// then maps it and uploads again, which books the deposit and the
// withdrawal and holds the four stock rows for their tickers.
func checkFirstXTBUploads(t *testing.T, api statementAPI, pool *pgxpool.Pool, session string) {
	t.Helper()
	first, err := uploadStatement(t, api, session, string(demo.XTBAccount), xtbExport())
	if err != nil || first.GetNew() != 0 || !slicesEqual(first.GetUnmappedBrokerAccounts(), []string{xtbtest.Account}) {
		t.Fatalf("an upload of an unmapped XTB account = %v, %v; want nothing booked and %s unmapped", first, err, xtbtest.Account)
	}
	if mapErr := ingestpg.New(pool).MapBrokerAccount(t.Context(), "xtb-main", xtbtest.Account, demo.XTBAccount, time.Now()); mapErr != nil {
		t.Fatalf("MapBrokerAccount: %v", mapErr)
	}
	second, err := uploadStatement(t, api, session, string(demo.XTBAccount), xtbExport())
	if err != nil || second.GetNew() != 2 || second.GetProblemsTotal() != 4 {
		t.Fatalf("an upload with no ticker confirmed = %v, %v; want the deposit and withdrawal booked and 4 rows waiting", second, err)
	}
}

// slicesEqual reports whether a and b hold the same strings in order.
func slicesEqual(a, b []string) bool { return strings.Join(a, "\n") == strings.Join(b, "\n") }

// checkPendingTickers checks the tickers waiting for the owner: TESTA.PL
// with the market-data connector's proposal, TESTB.UK with none, on
// London's exchange, whose trading currencies the owner picks from
// (D-170).
func checkPendingTickers(t *testing.T, tickers tickerAPI, session string) {
	t.Helper()
	req := connect.NewRequest(&ingestv1.ListPendingTickersRequest{Origin: localUI, SessionToken: session})
	req.Header().Set("Authorization", "Bearer "+uiCaller)
	got, err := tickers.ListPendingTickers(t.Context(), req)
	if err != nil {
		t.Fatalf("ListPendingTickers: %v", err)
	}
	want := []*ingestv1.PendingTicker{
		{Instance: "xtb-main", Symbol: "TESTA.PL", Country: "PL", Name: "Test Alpha SA", ProposedIsin: "PLTESTA00002", ProposedName: "TEST ALPHA", Waiting: 2},
		{Instance: "xtb-main", Symbol: "TESTB.UK", Country: "GB", Name: "Test Beta PLC", Waiting: 2, CurrencyChoices: []string{"GBX", "GBP", "USD"}},
	}
	if diff := cmp.Diff(want, got.Msg.GetTickers(), protocmp.Transform()); diff != "" {
		t.Errorf("pending tickers (-want +got):\n%s", diff)
	}
}

// confirm confirms isin for symbol of xtb-main as the owner's UI, picking
// no trading currency.
func confirm(t *testing.T, tickers tickerAPI, session, symbol, isin string) (*ingestv1.ConfirmTickerResponse, error) {
	t.Helper()
	return confirmPicking(t, tickers, session, symbol, isin, "")
}

// confirmPicking confirms isin for symbol of xtb-main as the owner's UI,
// picking the trading currency code currency, "" for none (D-170).
func confirmPicking(t *testing.T, tickers tickerAPI, session, symbol, isin, currency string) (*ingestv1.ConfirmTickerResponse, error) {
	t.Helper()
	req := connect.NewRequest(&ingestv1.ConfirmTickerRequest{
		Origin: localUI, SessionToken: session, Instance: "xtb-main", Symbol: symbol, Isin: isin, Currency: currency,
	})
	req.Header().Set("Authorization", "Bearer "+uiCaller)
	resp, err := tickers.ConfirmTicker(t.Context(), req)
	if err != nil {
		return nil, fmt.Errorf("confirm: %w", err)
	}
	return resp.Msg, nil
}

// checkConfirmTickers confirms the proposal for TESTA.PL and a typed ISIN
// for TESTB.UK, whose London listing names no trading currency, so it is
// refused, naming the choices, until the owner picks one (D-170), each
// booking its two waiting rows, and checks the refusals.
func checkConfirmTickers(t *testing.T, tickers tickerAPI, session string) {
	t.Helper()
	_, err := confirm(t, tickers, session, "TESTB.UK", "GBTESTB00000")
	if connect.CodeOf(err) != connect.CodeInvalidArgument || !strings.Contains(err.Error(), "GBX, GBP, USD") {
		t.Errorf("ConfirmTicker(TESTB.UK) picking no currency = %v, want InvalidArgument naming GBX, GBP, USD", err)
	}
	for symbol, pick := range map[string][2]string{"TESTA.PL": {"PLTESTA00002", ""}, "TESTB.UK": {"GBTESTB00000", "GBX"}} {
		got, err := confirmPicking(t, tickers, session, symbol, pick[0], pick[1])
		if err != nil || got.GetBooked() != 2 {
			t.Errorf("ConfirmTicker(%s) = %v, %v; want its 2 rows booked", symbol, got, err)
		}
	}
	refusals := []struct {
		symbol, isin, currency string
		want                   connect.Code
	}{
		{symbol: "TESTB.UK", isin: "PLTESTC00008", want: connect.CodeFailedPrecondition},
		{symbol: "TESTB.UK", isin: "GBTESTB00000", currency: "USD", want: connect.CodeFailedPrecondition},
		{symbol: "TESTZ.PL", isin: "PLTESTC00008", want: connect.CodeNotFound},
		{symbol: "TESTA.PL", isin: "PLTESTA00001", want: connect.CodeInvalidArgument},
	}
	for _, r := range refusals {
		if _, err := confirmPicking(t, tickers, session, r.symbol, r.isin, r.currency); connect.CodeOf(err) != r.want {
			t.Errorf("ConfirmTicker(%s, %s, %q) = %v, want %v", r.symbol, r.isin, r.currency, err, r.want)
		}
	}
	if _, err := confirm(t, tickers, "ofsess_nope_nope", "TESTA.PL", "PLTESTA00002"); connect.CodeOf(err) != connect.CodeUnauthenticated {
		t.Errorf("ConfirmTicker from an unknown session = %v, want Unauthenticated", err)
	}
}

// checkXTBBooked checks through the read API that the XTB account holds
// the six rows, keyed by XTB's operation IDs and scoped to xtb-main, its
// instruments with the confirmed ISINs, TESTA.PL's isinA, and XTB's
// symbols.
func checkXTBBooked(t *testing.T, srv *httptest.Server, session, isinA string) {
	t.Helper()
	var got viewsv1.ListTransactionsResponse
	getJSON(t, srv, "/v1/accounts/"+string(demo.XTBAccount)+"/transactions?page_size=100", session, &got)
	instruments := map[string]bool{}
	for _, tx := range got.GetTransactions() {
		if tx.GetSource() != "plugin:xtb-main:xtb:"+string(demo.XTBAccount) || strings.Contains(tx.GetDescription(), sendingBank) {
			t.Errorf("transaction %s/%s %q is not scoped to xtb-main or keeps a comment", tx.GetSource(), tx.GetExternalId(), tx.GetDescription())
		}
		for _, p := range tx.GetPostings() {
			if in := p.GetBook().GetInstrument(); in != nil {
				instruments[in.GetIsin()+"/"+in.GetSymbol()] = true
			}
		}
	}
	want := map[string]bool{isinA + "/TESTA.PL": true, "GBTESTB00000/TESTB.UK": true}
	if len(got.GetTransactions()) != 6 || !cmp.Equal(want, instruments) {
		t.Errorf("the XTB account holds %d transactions of %v, want 6 of %v", len(got.GetTransactions()), instruments, want)
	}
}

// checkHoldingAsserted checks through the read API that the export's open
// positions asserted the account's 8 units of TESTA.PL on 2026-06-30 as a
// holding of isinA, matched, and no other holding (D-155, D-153).
func checkHoldingAsserted(t *testing.T, srv *httptest.Server, session, isinA string) {
	t.Helper()
	var got viewsv1.ListAssertionChecksResponse
	getJSON(t, srv, "/v1/accounts/"+string(demo.XTBAccount)+"/balance-assertions", session, &got)
	var holdings []string
	for _, c := range got.GetChecks() {
		if in := c.GetBook().GetInstrument(); in != nil {
			holdings = append(holdings, c.GetDate()+" "+in.GetIsin()+"/"+in.GetSymbol()+" "+c.GetReported()+" "+c.GetStatus().String())
		}
	}
	want := []string{"2026-06-30 " + isinA + "/TESTA.PL 8 ASSERTION_STATUS_MATCHED"}
	if diff := cmp.Diff(want, holdings); diff != "" {
		t.Errorf("holding assertions (-want +got):\n%s", diff)
	}
}

// checkCorrectTicker lists the confirmed tickers, corrects TESTA.PL to
// PLTESTC00008, which books its three rows anew (D-153), and checks the
// refusals: an unknown ticker and an unknown session.
func checkCorrectTicker(t *testing.T, tickers tickerAPI, session string) {
	t.Helper()
	list := connect.NewRequest(&ingestv1.ListConfirmedTickersRequest{Origin: localUI, SessionToken: session})
	list.Header().Set("Authorization", "Bearer "+uiCaller)
	confirmed, err := tickers.ListConfirmedTickers(t.Context(), list)
	if err != nil || len(confirmed.Msg.GetTickers()) != 2 || confirmed.Msg.GetTickers()[0].GetIsin() != "PLTESTA00002" ||
		confirmed.Msg.GetTickers()[0].GetBooked() != 2 {
		t.Fatalf("ListConfirmedTickers = %v, %v; want TESTA.PL with its 2 rows booked and TESTB.UK", confirmed, err)
	}
	correct := func(sess, symbol string) (*ingestv1.CorrectTickerResponse, error) {
		req := connect.NewRequest(&ingestv1.CorrectTickerRequest{Origin: localUI, SessionToken: sess, Instance: "xtb-main", Symbol: symbol, Isin: "PLTESTC00008"})
		req.Header().Set("Authorization", "Bearer "+uiCaller)
		resp, callErr := tickers.CorrectTicker(t.Context(), req)
		if callErr != nil {
			return nil, fmt.Errorf("correct: %w", callErr)
		}
		return resp.Msg, nil
	}
	got, err := correct(session, "TESTA.PL")
	if err != nil || got.GetBooked() != 2 {
		t.Errorf("CorrectTicker(TESTA.PL) = %v, %v; want its 2 rows booked anew", got, err)
	}
	if _, err := correct(session, "TESTZ.PL"); connect.CodeOf(err) != connect.CodeNotFound {
		t.Errorf("CorrectTicker of an unknown ticker = %v, want NotFound", err)
	}
	if _, err := correct("ofsess_nope_nope", "TESTA.PL"); connect.CodeOf(err) != connect.CodeUnauthenticated {
		t.Errorf("CorrectTicker from an unknown session = %v, want Unauthenticated", err)
	}
}

// checkNoTransferComments checks that no text column of the ingest or
// ledger schema holds anything of a deposit's or withdrawal's comment.
func checkNoTransferComments(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()
	rows, err := pool.Query(t.Context(), `SELECT table_schema, table_name, column_name FROM information_schema.columns
		WHERE table_schema IN ('ingest', 'ledger') AND data_type IN ('text', 'character varying', 'jsonb')`)
	if err != nil {
		t.Fatalf("listing columns: %v", err)
	}
	type column struct{ schema, table, name string }
	var cols []column
	for rows.Next() {
		var c column
		if err := rows.Scan(&c.schema, &c.table, &c.name); err != nil {
			t.Fatalf("scanning a column: %v", err)
		}
		cols = append(cols, c)
	}
	rows.Close()
	for _, c := range cols {
		var n int
		q := fmt.Sprintf(`SELECT count(*) FROM %q.%q WHERE %q::text LIKE '%%' || $1 || '%%' OR %q::text LIKE '%%Wyplata%%'`,
			c.schema, c.table, c.name, c.name)
		if err := pool.QueryRow(t.Context(), q, sendingBank).Scan(&n); err != nil {
			t.Fatalf("searching %s.%s.%s: %v", c.schema, c.table, c.name, err)
		}
		if n != 0 {
			t.Errorf("%s.%s.%s holds a transfer's comment in %d rows", c.schema, c.table, c.name, n)
		}
	}
}
