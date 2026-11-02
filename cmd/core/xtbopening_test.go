package main

import (
	"context"
	"fmt"
	"net/http/httptest"
	"sort"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/google/go-cmp/cmp"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/arkadiuss/openfief/internal/demo"
	"github.com/arkadiuss/openfief/internal/ingest/connectrpc"
	ingestpg "github.com/arkadiuss/openfief/internal/ingest/postgres"
	"github.com/arkadiuss/openfief/internal/ingest/xtb/xtbtest"
	connectorv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/connector/v1"
	ingestv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/ingest/v1"
	"github.com/arkadiuss/openfief/sdk/go/gen/openfief/ingest/v1/ingestv1connect"
	viewsv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/views/v1"
)

// xtbOpeningExport is a synthetic current XTB export of a PLN account
// whose open positions are 7 TESTA.PL at 123.45 and 5 TESTB.UK at 412.55,
// a London price in pence, reported on 2026-06-30.
func xtbOpeningExport() []byte {
	day := func(m time.Month, d int) time.Time { return time.Date(2026, m, d, 10, 0, 0, 0, time.UTC) }
	e := xtbtest.Export{
		Account: xtbtest.Account, Currency: "PLN", From: day(6, 1), To: day(6, 30),
		Open: [][]xtbtest.Cell{
			xtbtest.OpenRow(&xtbtest.Position{
				Name: "Test Alpha SA", Ticker: "TESTA.PL", Category: "STOCK", Volume: "7", OpenPrice: "123.45", Opened: day(1, 5),
			}),
			xtbtest.OpenRow(&xtbtest.Position{
				Name: "Test Beta PLC", Ticker: "TESTB.UK", Category: "ETF", Volume: "5", OpenPrice: "412.55", Opened: day(2, 3),
			}),
		},
	}
	return e.Bytes()
}

// xtbListingConnector is xtbConnector whose market-data side proposes
// TESTA.PL on a listing that trades in PLN.
type xtbListingConnector struct {
	xtbConnector
}

func (xtbListingConnector) LookupInstruments(_ context.Context, req *connect.Request[connectorv1.LookupInstrumentsRequest]) (*connect.Response[connectorv1.LookupInstrumentsResponse], error) {
	out := &connectorv1.LookupInstrumentsResponse{}
	for _, t := range req.Msg.GetTickers() {
		if t.GetCode() == "TESTA" {
			out.Proposals = append(out.Proposals, &connectorv1.InstrumentProposal{
				Symbol: t.GetSymbol(), Isin: "PLTESTA00002", Name: "TEST ALPHA", Currency: "PLN",
			})
		}
	}
	return connect.NewResponse(out), nil
}

// TestXTBOpeningUpload checks D-159, D-160 and D-170 through the real
// core and Postgres: an opening upload whose positions all wait for their
// tickers' ISINs starts the account at once, so the next upload is a
// transactions batch; each confirmation adds its position to the opening on
// the report's date, its cost in the trading currency of the confirmed
// listing, or the one the owner picks for London, whose exchange quotes in
// more than one and is refused without a pick, pence converted to pounds.
func TestXTBOpeningUpload(t *testing.T) {
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
	port := serveConnector(t, xtbListingConnector{})
	c.deps.entries.set(connectrpc.NewClient(connectrpc.H2CClient(), func(string) (string, error) {
		return fmt.Sprintf("http://127.0.0.1:%d", port), nil
	}, staticTokens{}))
	c.deps.entries.setDiscovery(xtbDiscovery{})
	api := statementAPI{ingestv1connect.NewStatementServiceClient(srv.Client(), srv.URL)}
	tickers := tickerAPI{ingestv1connect.NewTickerServiceClient(srv.Client(), srv.URL)}
	checkOpeningStartsAccount(t, api, pool, session)
	if _, confirmErr := confirm(t, tickers, session, "TESTB.UK", "GBTESTB00000"); connect.CodeOf(confirmErr) != connect.CodeInvalidArgument {
		t.Fatalf("ConfirmTicker(TESTB.UK) picking no trading currency = %v, want InvalidArgument (D-170)", confirmErr)
	}
	for symbol, pick := range map[string][2]string{"TESTA.PL": {"PLTESTA00002", ""}, "TESTB.UK": {"GBTESTB00000", "GBX"}} {
		if _, confirmErr := confirmPicking(t, tickers, session, symbol, pick[0], pick[1]); confirmErr != nil {
			t.Fatalf("ConfirmTicker(%s): %v", symbol, confirmErr)
		}
	}
	checkOpeningLotCosts(t, srv, session)
}

// checkOpeningStartsAccount uploads the opening export, first while its
// broker account isn't mapped, which starts nothing, then mapped, which
// books nothing, every position waiting for its ticker, and starts the
// account, so the next upload is a transactions batch (D-160).
func checkOpeningStartsAccount(t *testing.T, api statementAPI, pool *pgxpool.Pool, session string) {
	t.Helper()
	if _, err := uploadStatementIn(t, api, session, string(demo.XTBAccount), ingestv1.Mode_MODE_UNSPECIFIED, xtbOpeningExport()); err != nil {
		t.Fatalf("first upload: %v", err)
	}
	if mapErr := ingestpg.New(pool).MapBrokerAccount(t.Context(), "xtb-main", xtbtest.Account, demo.XTBAccount, time.Now()); mapErr != nil {
		t.Fatalf("MapBrokerAccount: %v", mapErr)
	}
	opening, err := uploadStatementIn(t, api, session, string(demo.XTBAccount), ingestv1.Mode_MODE_UNSPECIFIED, xtbOpeningExport())
	if err != nil || opening.GetMode() != ingestv1.Mode_MODE_OPENING || opening.GetNew() != 0 || opening.GetProblemsTotal() != 2 {
		t.Fatalf("the opening upload = %v, %v; want an opening batch with both positions waiting", opening, err)
	}
	later, err := uploadStatementIn(t, api, session, string(demo.XTBAccount), ingestv1.Mode_MODE_UNSPECIFIED, xtbExport())
	if err != nil || later.GetMode() != ingestv1.Mode_MODE_TRANSACTIONS {
		t.Fatalf("a later upload = %v, %v; want a transactions batch: the opening upload started the account", later, err)
	}
}

// checkOpeningLotCosts checks through the read API that the XTB account's
// opening lots are dated on the report's date and cost 864.15 PLN for
// TESTA.PL, on its PLN listing, and 20.63 GBP for TESTB.UK, 5 times 412.55
// pence, the trading currency the owner picked, rounded half to even; the rows of the
// later upload, all dated before go-live, are not booked.
func checkOpeningLotCosts(t *testing.T, srv *httptest.Server, session string) {
	t.Helper()
	var got viewsv1.ListTransactionsResponse
	getJSON(t, srv, "/v1/accounts/"+string(demo.XTBAccount)+"/transactions?page_size=100", session, &got)
	var lots []string
	for _, tx := range got.GetTransactions() {
		for _, p := range tx.GetPostings() {
			if in := p.GetBook().GetInstrument(); in != nil {
				lots = append(lots, fmt.Sprintf("%s %s %s %s %s", tx.GetKind(), tx.GetDate(), in.GetSymbol(),
					p.GetAmount().GetAmount(), p.GetAmount().GetCurrency()))
			}
		}
	}
	sort.Strings(lots)
	want := []string{
		"TRANSACTION_KIND_OPENING_LOT 2026-06-30 TESTA.PL 864.15 PLN",
		"TRANSACTION_KIND_OPENING_LOT 2026-06-30 TESTB.UK 20.63 GBP",
	}
	if diff := cmp.Diff(want, lots); diff != "" {
		t.Errorf("the XTB account's instrument postings (-want +got):\n%s", diff)
	}
}
