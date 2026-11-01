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

// errConfirmed is the fake core's refusal of a second, different ISIN.
var errConfirmed = errors.New("the ticker's ISIN is already confirmed")

// fakeTickers is an in-memory TickerService: TESTA.PL with a proposal and
// TESTB.UK without one wait for the owner; it wants an owner-scoped
// session, refuses another ISIN for TESTC.PL as confirmed, and records
// the confirmations it took.
type fakeTickers struct {
	ingestv1connect.UnimplementedTickerServiceHandler

	mu        sync.Mutex
	confirmed []*ingestv1.ConfirmTickerRequest
}

func (f *fakeTickers) ListPendingTickers(_ context.Context, r *connect.Request[ingestv1.ListPendingTickersRequest]) (*connect.Response[ingestv1.ListPendingTickersResponse], error) {
	if err := owner(r.Msg.GetSessionToken()); err != nil {
		return nil, err
	}
	return connect.NewResponse(&ingestv1.ListPendingTickersResponse{Tickers: []*ingestv1.PendingTicker{
		{Instance: "xtb-main", Symbol: "TESTA.PL", Country: "PL", Name: "Test Alpha SA", ProposedIsin: "PLTESTA00002", ProposedName: "TEST ALPHA", Waiting: 2},
		{Instance: "xtb-main", Symbol: "TESTB.UK", Country: "GB", Name: "Test Beta <PLC>", Waiting: 1},
	}}), nil
}

func (f *fakeTickers) ConfirmTicker(_ context.Context, r *connect.Request[ingestv1.ConfirmTickerRequest]) (*connect.Response[ingestv1.ConfirmTickerResponse], error) {
	if err := owner(r.Msg.GetSessionToken()); err != nil {
		return nil, err
	}
	if r.Msg.GetSymbol() == "TESTC.PL" {
		return nil, codeErr(connect.CodeFailedPrecondition, errConfirmed)
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	f.confirmed = append(f.confirmed, r.Msg)
	return connect.NewResponse(&ingestv1.ConfirmTickerResponse{Booked: 2}), nil
}

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
// confirm, a field to type the ISIN when there is none, and the link in
// the menu; a read session gets no page.
func TestTickersPage(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	rec := h.serve(t, withSession(request(http.MethodGet, "/tickers", ""), ownerSession))
	wantStatus(t, rec, http.StatusOK)
	wantBody(t, rec, "<title>Tickers to confirm", `<a href="/tickers" aria-current="page">Tickers</a>`, "TESTA.PL", "Test Alpha SA",
		`value="PLTESTA00002"`, "Proposed: TEST ALPHA", ">Confirm</button>", "TESTB.UK", "Test Beta &lt;PLC&gt;",
		"No proposal: the lookup found none.", ">Save</button>", `action="/tickers/confirm"`, `<td class="num">2</td>`)
	wantNoBody(t, rec, "<script>", "style=", "Test Beta <PLC>")
	read := h.serve(t, withSession(request(http.MethodGet, "/tickers", ""), readSession))
	if read.Code == http.StatusOK {
		t.Errorf("GET /tickers for a read session = %d, want no page", read.Code)
	}
	overview := h.serve(t, withSession(request(http.MethodGet, "/price-checks", ""), readSession))
	wantNoBody(t, overview, `href="/tickers"`)
}

// TestConfirmTicker checks that the owner's ISIN, trimmed and in upper
// case, reaches the core with the ticker and instance, that the page then
// says how many rows were booked, and that the core's refusal and an
// overlong ISIN are shown on the page.
func TestConfirmTicker(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	rec := h.serve(t, withSession(form("/tickers/confirm", "instance=xtb-main&symbol=TESTB.UK&isin=+gbtestb00000+"), ownerSession))
	wantStatus(t, rec, http.StatusSeeOther)
	if loc := rec.Header().Get("Location"); loc != "/tickers?booked=2" {
		t.Errorf("redirect to %q, want /tickers?booked=2", loc)
	}
	h.tickers.mu.Lock()
	got := h.tickers.confirmed
	h.tickers.mu.Unlock()
	want := []*ingestv1.ConfirmTickerRequest{{Origin: testOrigin, SessionToken: ownerSession, Instance: "xtb-main", Symbol: "TESTB.UK", Isin: "GBTESTB00000"}}
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
