package ui

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"sync"
	"testing"

	"connectrpc.com/connect"
	"github.com/google/go-cmp/cmp"
	"google.golang.org/protobuf/testing/protocmp"

	ingestv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/ingest/v1"
	"github.com/arkadiuss/openfief/sdk/go/gen/openfief/ingest/v1/ingestv1connect"
)

// Refusals of the fake core.
var (
	// errConfirmed is the fake core's refusal of a second, different ISIN.
	errConfirmed = errors.New("the ticker's ISIN is already confirmed")
	// errPickCurrency is its refusal of TESTB.UK without a trading
	// currency (D-170).
	errPickCurrency = errors.New("pick TESTB.UK's trading currency, one of GBX, GBP, USD")
)

// fakeTickers is an in-memory TickerService: TESTA.PL with a proposal,
// TESTB.UK without one and TESTE.UK with a proposal quoted in pence wait
// for the owner, both on London's exchange, which quotes in more than one
// currency; it wants an owner-scoped session, refuses another ISIN for
// TESTC.PL as confirmed and TESTB.UK without a trading currency, and
// records the confirmations it took.
type fakeTickers struct {
	ingestv1connect.UnimplementedTickerServiceHandler

	mu        sync.Mutex
	confirmed []*ingestv1.ConfirmTickerRequest
	corrected []*ingestv1.CorrectTickerRequest
}

// errNotConfirmed is the fake core's refusal of a correction of a ticker
// never confirmed.
var errNotConfirmed = errors.New("the ticker's ISIN is not confirmed")

func (f *fakeTickers) ListConfirmedTickers(_ context.Context, r *connect.Request[ingestv1.ListConfirmedTickersRequest]) (*connect.Response[ingestv1.ListConfirmedTickersResponse], error) {
	if err := owner(r.Msg.GetSessionToken()); err != nil {
		return nil, err
	}
	return connect.NewResponse(&ingestv1.ListConfirmedTickersResponse{Tickers: []*ingestv1.ConfirmedTicker{
		{Institution: "xtb", Instance: "xtb-main", Symbol: "TESTD.PL", Country: "PL", Name: "Test Delta SA", Isin: "PLTESTD00004", TradingCurrency: "PLN", Booked: 3},
		{
			Institution: "xtb", Instance: "xtb-main", Symbol: "TESTF.UK", Country: "GB", Name: "Test Phi PLC", Isin: "GBTESTF00009",
			TradingCurrency: "GBX", CurrencyChoices: londonCurrencies(), Booked: 1,
		},
	}}), nil
}

func (f *fakeTickers) CorrectTicker(_ context.Context, r *connect.Request[ingestv1.CorrectTickerRequest]) (*connect.Response[ingestv1.CorrectTickerResponse], error) {
	if err := owner(r.Msg.GetSessionToken()); err != nil {
		return nil, err
	}
	if r.Msg.GetSymbol() != "TESTD.PL" && r.Msg.GetSymbol() != "TESTF.UK" {
		return nil, codeErr(connect.CodeFailedPrecondition, errNotConfirmed)
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	f.corrected = append(f.corrected, r.Msg)
	return connect.NewResponse(&ingestv1.CorrectTickerResponse{Booked: 3}), nil
}

func (f *fakeTickers) ListPendingTickers(_ context.Context, r *connect.Request[ingestv1.ListPendingTickersRequest]) (*connect.Response[ingestv1.ListPendingTickersResponse], error) {
	if err := owner(r.Msg.GetSessionToken()); err != nil {
		return nil, err
	}
	return connect.NewResponse(&ingestv1.ListPendingTickersResponse{Tickers: []*ingestv1.PendingTicker{
		{Instance: "xtb-main", Symbol: "TESTA.PL", Country: "PL", Name: "Test Alpha SA", ProposedIsin: "PLTESTA00002", ProposedName: "TEST ALPHA", Waiting: 2},
		{Instance: "xtb-main", Symbol: "TESTB.UK", Country: "GB", Name: "Test Beta <PLC>", Waiting: 1, CurrencyChoices: londonCurrencies()},
		{
			Instance: "xtb-main", Symbol: "TESTE.UK", Country: "GB", Name: "Test Epsilon PLC", ProposedIsin: "GBTESTE00006",
			ProposedName: "TEST EPSILON", ProposedCurrency: "GBX", Waiting: 1, CurrencyChoices: londonCurrencies(),
		},
	}}), nil
}

func (f *fakeTickers) ConfirmTicker(_ context.Context, r *connect.Request[ingestv1.ConfirmTickerRequest]) (*connect.Response[ingestv1.ConfirmTickerResponse], error) {
	if err := owner(r.Msg.GetSessionToken()); err != nil {
		return nil, err
	}
	switch {
	case r.Msg.GetSymbol() == "TESTC.PL":
		return nil, codeErr(connect.CodeFailedPrecondition, errConfirmed)
	case r.Msg.GetSymbol() == "TESTB.UK" && r.Msg.GetCurrency() == "":
		return nil, codeErr(connect.CodeInvalidArgument, errPickCurrency)
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	f.confirmed = append(f.confirmed, r.Msg)
	return connect.NewResponse(&ingestv1.ConfirmTickerResponse{Booked: 2}), nil
}

// londonCurrencies are the trading currencies London's exchange quotes in.
func londonCurrencies() []string { return []string{"GBX", "GBP", "USD"} }

// xtbUploadResponse is the fake core's answer to an XTB export: two rows
// wait for their ticker's ISIN (D-116).
func xtbUploadResponse() *ingestv1.UploadStatementResponse {
	return &ingestv1.UploadStatementResponse{
		Mode: ingestv1.Mode_MODE_TRANSACTIONS, AccountIds: []string{accountID}, BatchId: "b-8", Institution: "xtb",
		Transactions: 1, New: 1, ProblemsTotal: 3, Problems: []*ingestv1.Problem{
			{Kind: ingestv1.ProblemKind_PROBLEM_KIND_UNSUPPORTED, Ref: "Cash Operations#7[ID=9000000002]", Reason: "waits for the owner to confirm the ISIN of TESTA.PL"},
			{Kind: ingestv1.ProblemKind_PROBLEM_KIND_UNSUPPORTED, Ref: "Cash Operations#8[ID=9000000003]", Reason: "waits for the owner to confirm the ISIN of TESTB.UK"},
			{Kind: ingestv1.ProblemKind_PROBLEM_KIND_UNSUPPORTED, Ref: "Open Positions#11", Reason: `open CFD position (category "CMD") is not supported (D-24)`},
		},
	}
}

// TestTickersPage checks the owner's tickers page (D-116): each ticker
// waiting with its name, instance and rows, the proposal filled in to
// confirm, a field to type the ISIN when there is none, a choice of
// trading currency on an exchange that quotes in more than one, required
// when the proposal names none (D-170), and the link in the menu; a read
// session gets no page.
func TestTickersPage(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	rec := h.serve(t, withSession(request(http.MethodGet, "/tickers", ""), ownerSession))
	wantStatus(t, rec, http.StatusOK)
	wantBody(t, rec, "<title>Tickers to confirm", `<a href="/tickers" aria-current="page">Tickers</a>`, "TESTA.PL", "Test Alpha SA",
		`value="PLTESTA00002"`, "Proposed: TEST ALPHA", ">Confirm</button>", "TESTB.UK", "Test Beta &lt;PLC&gt;",
		"No proposal: the lookup found none.", ">Save</button>", `action="/tickers/confirm"`, `<td class="num">2</td>`,
		`<select id="currency-xtb-main-TESTB.UK" name="currency" aria-describedby="currency-xtb-main-TESTB.UK-help" required>`,
		`<option value="" selected disabled>Trading currency</option>`, `<option value="GBX">GBX</option>`,
		`<option value="GBP">GBP</option>`, `<option value="USD">USD</option>`,
		`<select id="currency-xtb-main-TESTE.UK" name="currency" aria-describedby="currency-xtb-main-TESTE.UK-help">`,
		`<option value="" selected>The listing's: GBX</option>`, "The proposed listing quotes in GBX.")
	wantNoBody(t, rec, "<script>", "style=", "Test Beta <PLC>", `id="currency-xtb-main-TESTA.PL"`)
	read := h.serve(t, withSession(request(http.MethodGet, "/tickers", ""), readSession))
	if read.Code == http.StatusOK {
		t.Errorf("GET /tickers for a read session = %d, want no page", read.Code)
	}
	overview := h.serve(t, withSession(request(http.MethodGet, "/price-checks", ""), readSession))
	wantNoBody(t, overview, `href="/tickers"`)
}

// TestConfirmTicker checks that the owner's ISIN, trimmed and in upper
// case, reaches the core with the ticker, instance and the trading
// currency picked, that the page then says how many rows were booked, and
// that the core's refusals, such as one naming the currencies to pick from
// (D-170), and an overlong ISIN or currency are shown on the page.
func TestConfirmTicker(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	rec := h.serve(t, withSession(form("/tickers/confirm", "instance=xtb-main&symbol=TESTB.UK&isin=+gbtestb00000+&currency=GBX"), ownerSession))
	wantStatus(t, rec, http.StatusSeeOther)
	if loc := rec.Header().Get("Location"); loc != "/tickers?booked=2" {
		t.Errorf("redirect to %q, want /tickers?booked=2", loc)
	}
	h.tickers.mu.Lock()
	got := h.tickers.confirmed
	h.tickers.mu.Unlock()
	want := []*ingestv1.ConfirmTickerRequest{{Origin: testOrigin, SessionToken: ownerSession, Instance: "xtb-main", Symbol: "TESTB.UK", Isin: "GBTESTB00000", Currency: "GBX"}}
	if diff := cmp.Diff(want, got, protocmp.Transform()); diff != "" {
		t.Errorf("confirmations handed to the core (-want +got):\n%s", diff)
	}
	booked := h.serve(t, withSession(request(http.MethodGet, "/tickers?booked=2", ""), ownerSession))
	wantBody(t, booked, "Confirmed: 2 waiting row(s) booked.")
	forged := h.serve(t, withSession(request(http.MethodGet, "/tickers?booked=all", ""), ownerSession))
	wantNoBody(t, forged, "Confirmed:")
	refused := h.serve(t, withSession(form("/tickers/confirm", "instance=xtb-main&symbol=TESTC.PL&isin=PLTESTC00008"), ownerSession))
	wantStatus(t, refused, http.StatusBadRequest)
	wantBody(t, refused, "The core refused this: the ticker&#39;s ISIN is already confirmed")
	unpicked := h.serve(t, withSession(form("/tickers/confirm", "instance=xtb-main&symbol=TESTB.UK&isin=GBTESTB00000"), ownerSession))
	wantStatus(t, unpicked, http.StatusBadRequest)
	wantBody(t, unpicked, "The core refused this: pick TESTB.UK&#39;s trading currency, one of GBX, GBP, USD")
	longCurrency := h.serve(t, withSession(form("/tickers/confirm", "instance=xtb-main&symbol=TESTB.UK&isin=GBTESTB00000&currency=POUNDS"), ownerSession))
	wantStatus(t, longCurrency, http.StatusBadRequest)
	wantBody(t, longCurrency, "A trading currency has 3 letters")
	long := h.serve(t, withSession(form("/tickers/confirm", "instance=xtb-main&symbol=TESTA.PL&isin="+strings.Repeat("X", 13)), ownerSession))
	wantStatus(t, long, http.StatusBadRequest)
	wantBody(t, long, "An ISIN has 12 characters")
}

// TestUploadWaitingForTickers checks that an upload whose rows wait for
// their tickers' ISINs links the owner to the tickers page.
func TestUploadWaitingForTickers(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	rec := h.serve(t, withSession(uploadRequest(modePart("transactions"), filePart("xtb.xlsx", strings.NewReader("PK\x03\x04"))), ownerSession))
	wantStatus(t, rec, http.StatusOK)
	wantBody(t, rec, "Import batch b-8 from xtb", `2 of the rows wait for their ticker's ISIN: <a href="/tickers">confirm the tickers</a>`,
		"waits for the owner to confirm the ISIN of TESTA.PL")
}

// TestCorrectTicker checks D-153 on the tickers page: the confirmed
// tickers are listed with their institution, ISIN, trading currency and
// rows booked, each with a form to correct the ISIN, and a choice of
// currency on London; the owner's correction, trimmed and in upper case,
// reaches the core, the page then says how many rows were booked anew,
// and the core's refusal is shown.
func TestCorrectTicker(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	page := h.serve(t, withSession(request(http.MethodGet, "/tickers", ""), ownerSession))
	wantStatus(t, page, http.StatusOK)
	wantBody(t, page, "Confirmed tickers", "TESTD.PL", "Test Delta SA", `<td class="mono">PLTESTD00004</td>`, `action="/tickers/correct"`,
		`value="PLTESTD00004"`, `aria-label="Correct the ISIN of TESTD.PL"`, `<td class="num">3</td>`,
		`<select id="correct-currency-xtb-TESTF.UK" name="currency">`, `<option value="" selected>The listing's: GBX</option>`,
		"superseded and kept in the log")
	wantNoBody(t, page, `id="correct-currency-xtb-TESTD.PL"`)
	rec := h.serve(t, withSession(form("/tickers/correct", "instance=xtb-main&symbol=TESTD.PL&isin=+pltestc00008+"), ownerSession))
	wantStatus(t, rec, http.StatusSeeOther)
	if loc := rec.Header().Get("Location"); loc != "/tickers?corrected=3" {
		t.Errorf("redirect to %q, want /tickers?corrected=3", loc)
	}
	h.tickers.mu.Lock()
	got := h.tickers.corrected
	h.tickers.mu.Unlock()
	want := []*ingestv1.CorrectTickerRequest{{Origin: testOrigin, SessionToken: ownerSession, Instance: "xtb-main", Symbol: "TESTD.PL", Isin: "PLTESTC00008"}}
	if diff := cmp.Diff(want, got, protocmp.Transform()); diff != "" {
		t.Errorf("corrections handed to the core (-want +got):\n%s", diff)
	}
	corrected := h.serve(t, withSession(request(http.MethodGet, "/tickers?corrected=3", ""), ownerSession))
	wantBody(t, corrected, "Corrected: 3 row(s) booked anew.")
	refused := h.serve(t, withSession(form("/tickers/correct", "instance=xtb-main&symbol=TESTA.PL&isin=PLTESTC00008"), ownerSession))
	wantStatus(t, refused, http.StatusBadRequest)
	wantBody(t, refused, "The core refused this: the ticker&#39;s ISIN is not confirmed")
	long := h.serve(t, withSession(form("/tickers/correct", "instance=xtb-main&symbol=TESTD.PL&isin="+strings.Repeat("X", 13)), ownerSession))
	wantStatus(t, long, http.StatusBadRequest)
	if read := h.serve(t, withSession(form("/tickers/correct", "instance=xtb-main&symbol=TESTD.PL&isin=PLTESTC00008"), readSession)); read.Code != http.StatusForbidden {
		t.Errorf("a read session's correction = %d, want 403", read.Code)
	}
}
