package main

import (
	"fmt"
	"net/http/httptest"
	"slices"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/google/go-cmp/cmp"

	"github.com/arkadiuss/openfief/internal/demo"
	"github.com/arkadiuss/openfief/internal/ingest/connectrpc"
	ingestpg "github.com/arkadiuss/openfief/internal/ingest/postgres"
	"github.com/arkadiuss/openfief/internal/ingest/xtb/xtbtest"
	ingestv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/ingest/v1"
	"github.com/arkadiuss/openfief/sdk/go/gen/openfief/ingest/v1/ingestv1connect"
	viewsv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/views/v1"
)

// xtbGoLive is when the synthetic current export of
// TestXTBStartFromOpenPositions was generated: the XTB account's go-live.
func xtbGoLive() time.Time { return time.Date(2026, 6, 30, 12, 0, 0, 0, time.UTC) }

// xtbCurrentName is the current export's name, as xStation writes it.
const xtbCurrentName = "PLN_0000000_2026-01-01_2026-06-30.xlsx"

// currentXTBExport is a synthetic current XTB export generated at to: one
// open position of 7 TESTA.PL bought on 2026-01-05 at 100 PLN, and the
// cash operations since the account was opened, plus later ones.
func currentXTBExport(to time.Time, later ...[]xtbtest.Cell) []byte {
	day := func(d int) time.Time { return time.Date(2026, time.January, d, 10, 0, 0, 0, time.UTC) }
	e := xtbtest.Export{
		Account: xtbtest.Account, Currency: "PLN", From: day(1), To: to,
		Cash: append(slices.Clone(later),
			xtbtest.CashRow("Stock purchase", "Test Alpha SA", "TESTA.PL", "STOCK", day(5), "-700", "9000000002", "OPEN BUY 7/7 @ 100"),
			xtbtest.CashRow("Deposit", "", "", "", day(5), "5000", "9000000001", "Przelew z "+sendingBank+" S.A."),
		),
		Open: [][]xtbtest.Cell{xtbtest.OpenRow(&xtbtest.Position{
			Name: "Test Alpha SA", Ticker: "TESTA.PL", Category: "STOCK", Volume: "7", OpenPrice: "100", Opened: day(5), Value: "700",
		})},
	}
	return e.Bytes()
}

// xtbCore is a real core with an owner session and the XTB account bound
// to an in-process xtb connector, its broker account reported by a first
// upload and mapped.
type xtbCore struct {
	srv     *httptest.Server
	api     statementAPI
	tickers tickerAPI
	session string
}

// startXTBCore starts the core over a fresh database for an XTB account,
// uploads the current export once while its broker account is unmapped,
// which books nothing, and maps it.
func startXTBCore(t *testing.T) xtbCore {
	t.Helper()
	dsn := emptyDatabase(t)
	code := enrollCode(t, enrollCmd(t, dsn))
	c, srv, pool := startCoreForUI(t, dsn, func(*serveConfig) {})
	ac := authClient(srv.URL, srv.Client(), uiCaller)
	x := xtbCore{
		srv: srv, session: passkeyLogin(t, ac, passkeyEnroll(t, ac, code)),
		api:     statementAPI{ingestv1connect.NewStatementServiceClient(srv.Client(), srv.URL)},
		tickers: tickerAPI{ingestv1connect.NewTickerServiceClient(srv.Client(), srv.URL)},
	}
	saveDemoRegistry(t, dsn)
	if _, err := connectorCmd(t, dsn, nil, "bind", "--instance=xtb-main", "--account="+string(demo.XTBAccount)); err != nil {
		t.Fatalf("connector bind: %v", err)
	}
	port := serveConnector(t, xtbConnector{})
	c.deps.entries.set(connectrpc.NewClient(connectrpc.H2CClient(), func(string) (string, error) {
		return fmt.Sprintf("http://127.0.0.1:%d", port), nil
	}, staticTokens{}))
	c.deps.entries.setDiscovery(xtbDiscovery{})
	unmapped, err := uploadWithHeader(t, x.api, x.header("", xtbCurrentName), currentXTBExport(xtbGoLive()))
	if err != nil || unmapped.GetNew() != 0 || len(unmapped.GetUnmappedBrokerAccounts()) != 1 {
		t.Fatalf("an upload of an unmapped XTB account = %v, %v; want nothing booked and its broker account unmapped", unmapped, err)
	}
	if mapErr := ingestpg.New(pool).MapBrokerAccount(t.Context(), "xtb-main", xtbtest.Account, demo.XTBAccount, time.Now()); mapErr != nil {
		t.Fatalf("MapBrokerAccount: %v", mapErr)
	}
	return x
}

// header is the owner's upload header for the XTB account with cash
// entered in PLN, none when cash is empty, and the core choosing the mode.
func (x *xtbCore) header(cash, fileName string) *ingestv1.StatementHeader {
	h := &ingestv1.StatementHeader{Origin: localUI, SessionToken: x.session, AccountId: string(demo.XTBAccount), FileName: fileName}
	if cash != "" {
		h.OpeningCash, h.OpeningCashCurrency = cash, "PLN"
	}
	return h
}

// TestXTBStartFromOpenPositions checks D-136 through the real core and
// Postgres: an XTB account starts from a current export's open positions,
// each an opening lot waiting for its ticker's ISIN, and the cash balance
// the owner enters once; uploading it again changes nothing, and cash
// entered again is refused; confirming the ticker adds the lot to the
// seeding; a later export continues the account with what came after its
// go-live.
func TestXTBStartFromOpenPositions(t *testing.T) {
	t.Parallel()
	x := startXTBCore(t)
	file := currentXTBExport(xtbGoLive())
	checkXTBOpening(t, &x, file)
	checkXTBOpeningAgain(t, &x, file)
	if got, err := confirm(t, x.tickers, x.session, "TESTA.PL", "PLTESTA00002"); err != nil || got.GetBooked() != 1 {
		t.Fatalf("ConfirmTicker(TESTA.PL) = %v, %v; want its opening lot booked", got, err)
	}
	later := currentXTBExport(time.Date(2026, 7, 31, 12, 0, 0, 0, time.UTC),
		xtbtest.CashRow("Stock purchase", "Test Alpha SA", "TESTA.PL", "STOCK", time.Date(2026, 7, 10, 9, 0, 0, 0, time.UTC), "-220", "9000000004", "OPEN BUY 2/2 @ 110"),
		xtbtest.CashRow("Deposit", "", "", "", time.Date(2026, 7, 2, 9, 0, 0, 0, time.UTC), "1000", "9000000003", "Przelew z "+sendingBank+" S.A."),
	)
	next, err := uploadWithHeader(t, x.api, x.header("", "PLN_0000000_2026-01-01_2026-07-31.xlsx"), later)
	if err != nil || next.GetMode() != ingestv1.Mode_MODE_TRANSACTIONS || next.GetNew() != 2 || next.GetProblemsTotal() != 2 {
		t.Fatalf("a later export = %v, %v; want transactions, the 2 rows after go-live booked, the 2 before skipped", next, err)
	}
	checkXTBStarted(t, x.srv, x.session)
}

// checkXTBOpening uploads the current export with the core choosing the
// mode and 4300 PLN of opening cash, which books the cash and holds the
// opening lot for its ticker.
func checkXTBOpening(t *testing.T, x *xtbCore, file []byte) {
	t.Helper()
	first, err := uploadWithHeader(t, x.api, x.header("4300.00", xtbCurrentName), file)
	if err != nil || first.GetMode() != ingestv1.Mode_MODE_OPENING || first.GetNew() != 1 || first.GetProblemsTotal() != 1 {
		t.Fatalf("the opening upload = %v, %v; want opening, the cash booked and the lot waiting", first, err)
	}
}

// checkXTBOpeningAgain uploads the same export again: with the core
// choosing, in transactions mode, which skips the rows before go-live and
// books nothing; with cash, which is refused; in opening mode without
// cash, which changes nothing; and in opening mode with other cash, which
// is refused.
func checkXTBOpeningAgain(t *testing.T, x *xtbCore, file []byte) {
	t.Helper()
	again, err := uploadWithHeader(t, x.api, x.header("", xtbCurrentName), file)
	if err != nil || again.GetMode() != ingestv1.Mode_MODE_TRANSACTIONS || again.GetNew() != 0 || again.GetProblemsTotal() != 2 {
		t.Fatalf("the same export again = %v, %v; want transactions, nothing new, the 2 rows before go-live skipped", again, err)
	}
	if _, cashErr := uploadWithHeader(t, x.api, x.header("4300.00", xtbCurrentName), file); connect.CodeOf(cashErr) != connect.CodeInvalidArgument {
		t.Errorf("opening cash with the seeded account's upload = %v, want InvalidArgument", cashErr)
	}
	repeat := x.header("", xtbCurrentName)
	repeat.Mode = ingestv1.Mode_MODE_OPENING
	if same, sameErr := uploadWithHeader(t, x.api, repeat, file); sameErr != nil || same.GetNew() != 0 || same.GetTransactions() != 0 {
		t.Errorf("the same opening upload again = %v, %v; want nothing booked, the account being seeded", same, sameErr)
	}
	other := x.header("4000.00", xtbCurrentName)
	other.Mode = ingestv1.Mode_MODE_OPENING
	if _, otherErr := uploadWithHeader(t, x.api, other, file); connect.CodeOf(otherErr) != connect.CodeFailedPrecondition {
		t.Errorf("an opening upload with cash entered again = %v, want FailedPrecondition", otherErr)
	}
}

// checkXTBStarted checks through the read API that the account holds its
// opening cash and opening lot on its go-live, 2026-06-30, and the deposit
// and buy after it, scoped to xtb-main, and nothing from before it.
func checkXTBStarted(t *testing.T, srv *httptest.Server, session string) {
	t.Helper()
	var got viewsv1.ListTransactionsResponse
	getJSON(t, srv, "/v1/accounts/"+string(demo.XTBAccount)+"/transactions?page_size=100", session, &got)
	kinds := map[string]viewsv1.TransactionKind{}
	for _, tx := range got.GetTransactions() {
		if tx.GetSource() != "plugin:xtb-main:xtb:"+string(demo.XTBAccount) {
			t.Errorf("transaction %s/%s is not scoped to xtb-main", tx.GetSource(), tx.GetExternalId())
		}
		kinds[tx.GetExternalId()] = tx.GetKind()
	}
	want := map[string]viewsv1.TransactionKind{
		"opening-cash:PLN:2026-06-30":               viewsv1.TransactionKind_TRANSACTION_KIND_DEPOSIT,
		"open-position:TESTA.PL@20260105T100000Z#1": viewsv1.TransactionKind_TRANSACTION_KIND_OPENING_LOT,
		"9000000003": viewsv1.TransactionKind_TRANSACTION_KIND_DEPOSIT,
		"9000000004": viewsv1.TransactionKind_TRANSACTION_KIND_BUY,
	}
	if diff := cmp.Diff(want, kinds); diff != "" {
		t.Errorf("the XTB account's transactions (-want +got):\n%s", diff)
	}
}
