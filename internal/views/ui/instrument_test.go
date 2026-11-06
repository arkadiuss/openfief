package ui

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"testing"

	"connectrpc.com/connect"
	"github.com/google/go-cmp/cmp"

	ledgerv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/ledger/v1"
	"github.com/arkadiuss/openfief/sdk/go/gen/openfief/ledger/v1/ledgerv1connect"
	viewsv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/views/v1"
)

// errNameRefused is the fake core's refusal of a name holding a control
// character.
var errNameRefused = errors.New("invalid instrument name: control character")

// fakeInstruments is an in-memory InstrumentService (D-197): it wants the
// plugin's token and an owner-scoped session, refuses a name with a tab
// as the core refuses control characters, and records each rename.
type fakeInstruments struct {
	ledgerv1connect.UnimplementedInstrumentServiceHandler

	mu      sync.Mutex
	renamed []string
}

func (f *fakeInstruments) RenameInstrument(_ context.Context, r *connect.Request[ledgerv1.RenameInstrumentRequest]) (*connect.Response[ledgerv1.RenameInstrumentResponse], error) {
	if r.Header().Get("Authorization") != "Bearer "+pluginToken {
		return nil, codeErr(connect.CodePermissionDenied, nil)
	}
	if err := owner(r.Msg.GetSessionToken()); err != nil {
		return nil, err
	}
	if strings.Contains(r.Msg.GetName(), "\t") {
		return nil, codeErr(connect.CodeInvalidArgument, errNameRefused)
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	f.renamed = append(f.renamed, r.Msg.GetIsin()+"="+r.Msg.GetName())
	return connect.NewResponse(&ledgerv1.RenameInstrumentResponse{Name: r.Msg.GetName()}), nil
}

// locked runs fn with the fake's lock held.
func (f *fakeInstruments) locked(fn func()) {
	f.mu.Lock()
	defer f.mu.Unlock()
	fn()
}

// ListInstrumentPrices has the synthetic instrument priced by eodhd, the
// source it is valued at, and cross-checked by ibkr-main on one day.
func (f *fakeRead) ListInstrumentPrices(_ context.Context, r *connect.Request[viewsv1.ListInstrumentPricesRequest]) (*connect.Response[viewsv1.ListInstrumentPricesResponse], error) {
	if err := f.check(r.Header()); err != nil {
		return nil, err
	}
	f.locked(func() { f.dates = append(f.dates, r.Msg.GetFrom()+".."+r.Msg.GetTo()) })
	price := func(day, amount, source string, primary bool) *viewsv1.InstrumentPrice {
		return &viewsv1.InstrumentPrice{
			Instrument: syntheticInstrument(), Date: day, Close: &viewsv1.Money{Amount: amount, Currency: "EUR"}, Source: source, Primary: primary,
		}
	}
	return connect.NewResponse(&viewsv1.ListInstrumentPricesResponse{Isin: r.Msg.GetIsin(), Prices: []*viewsv1.InstrumentPrice{
		price("2026-03-30", "24.50", "eodhd", true), price("2026-03-31", "25.00", "eodhd", true), price("2026-03-31", "25.10", "ibkr-main", false),
	}}), nil
}

// ListInstrumentTransactions lists the synthetic account's buy of the
// synthetic instrument, with a next page on the first.
func (f *fakeRead) ListInstrumentTransactions(_ context.Context, r *connect.Request[viewsv1.ListInstrumentTransactionsRequest]) (*connect.Response[viewsv1.ListInstrumentTransactionsResponse], error) {
	if err := f.check(r.Header()); err != nil {
		return nil, err
	}
	f.locked(func() { f.pageTokens = append(f.pageTokens, r.Msg.GetIsin()+":"+r.Msg.GetPageToken()) })
	resp := &viewsv1.ListInstrumentTransactionsResponse{Transactions: []*viewsv1.Transaction{{
		AccountId: accountID, Source: "synthetic", ExternalId: "b1", Kind: viewsv1.TransactionKind_TRANSACTION_KIND_BUY, Date: "2026-01-02",
		Description: "Synthetic buy",
		Postings: []*viewsv1.Posting{{
			Book: &viewsv1.Book{Kind: viewsv1.BookKind_BOOK_KIND_HOLDING, Instrument: syntheticInstrument()}, Amount: &viewsv1.Money{Amount: "250.00", Currency: "EUR"}, Quantity: "10",
		}},
	}}}
	if r.Msg.GetPageToken() == "" {
		resp.NextPageToken = "page-2"
	}
	return connect.NewResponse(resp), nil
}

// TestInstrumentPage checks an instrument's page (D-195, D-197): its name
// first, then its ticker and ISIN, the breadcrumb through the account it
// was opened from, its latest price and price history with the primary
// source marked, its holdings and lots across accounts, its transactions
// with their accounts and a next page, its price checks with their
// history, its yearly tax class and its open items; the owner gets the
// rename form and the price checks' forms, which return to the page, and
// a read session neither; an unknown ISIN is not found.
func TestInstrumentPage(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	priceCheckItem(h)
	rec := h.serve(t, withSession(request(http.MethodGet, "/instruments/XX0000000001?account="+accountID, ""), ownerSession))
	wantStatus(t, rec, http.StatusOK)
	wantBody(t, rec,
		"<title>Synthetic Fund · OpenFief</title>", "<h1>Synthetic Fund</h1>", `<span class="mono">SYN</span> · ISIN <span class="mono">XX0000000001</span>`,
		`<nav class="breadcrumb" aria-label="Breadcrumb">`, `<a href="/">Overview</a>`, `<a href="/accounts">Accounts</a>`,
		`<a href="/accounts/acc-synthetic-1">Synthetic brokerage</a>`, `<span aria-current="page">Synthetic Fund</span>`,
		"SYN · latest price", "25,00 EUR", "valued at eodhd",
		`&#34;label&#34;:&#34;SYN · eodhd (primary)&#34;`, `&#34;label&#34;:&#34;SYN · ibkr-main&#34;`, "only cross-check it",
		`<a href="/accounts/acc-synthetic-1">Synthetic brokerage</a></td>`, "Synthetic buy",
		`hx-get="/instruments/XX0000000001/transactions?page_token=page-2"`, `href="/instruments/XX0000000001?page_token=page-2"`,
		"2.8232%", "0.2381%", `name="next" value="/instruments/XX0000000001"`, `action="/inbox/price-checks/threshold"`,
		"OKI", "Investment assets", "the jurisdiction pack",
		`id="instrument-items-title"`, `<a href="/inbox#item-30">`,
		`action="/instruments/XX0000000001/name"`, `name="name" value="Synthetic Fund"`, "Save name",
	)
	if got := seen(h.read, func() []string { return h.read.dates }); len(got) == 0 || got[len(got)-1] != "2025-03-31..2026-03-31" {
		t.Errorf("ListInstrumentPrices asked for %v, want the year to 2026-03-31", got)
	}
	rec = h.serve(t, withSession(request(http.MethodGet, "/instruments/XX0000000001", ""), readSession))
	wantStatus(t, rec, http.StatusOK)
	wantBody(t, rec, "<h1>Synthetic Fund</h1>", `<a href="/">Overview</a></li><li><span aria-current="page">Synthetic Fund</span>`)
	wantNoBody(t, rec, "Save name", `action="/inbox/price-checks`)
	rec = h.serve(t, withSession(request(http.MethodGet, "/instruments/XX0000000009", ""), readSession))
	wantStatus(t, rec, http.StatusNotFound)
}

// TestInstrumentTransactionsPaging checks "More" on an instrument's page:
// htmx gets only the next page's rows with their accounts; a browser
// without htmx goes to the page at that page.
func TestInstrumentTransactionsPaging(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	rec := h.serve(t, htmx(withSession(request(http.MethodGet, "/instruments/XX0000000001/transactions?page_token=page-2", ""), readSession)))
	wantStatus(t, rec, http.StatusOK)
	wantBody(t, rec, "Synthetic buy", "Synthetic brokerage")
	wantNoBody(t, rec, "<html", "More transactions")
	if diff := cmp.Diff([]string{"XX0000000001:page-2"}, seen(h.read, func() []string { return h.read.pageTokens })); diff != "" {
		t.Errorf("page tokens (-want +got):\n%s", diff)
	}
	rec = h.serve(t, withSession(request(http.MethodGet, "/instruments/XX0000000001/transactions?page_token=page-2", ""), readSession))
	wantStatus(t, rec, http.StatusSeeOther)
	if loc := rec.Header().Get("Location"); loc != "/instruments/XX0000000001?page_token=page-2" {
		t.Errorf("Location = %q", loc)
	}
}

// TestRenameInstrument checks the owner's rename form (D-197): the name
// goes to the core trimmed with the plugin's token and the owner's session,
// and the page says it is saved; the core's refusal shows its reason with
// the name as posted; an empty or too long name and a malformed form never
// reach the core; a read session is refused.
func TestRenameInstrument(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	rename := func(body, token string) *http.Request {
		return withSession(form("/instruments/XX0000000001/name", body), token)
	}
	rec := h.serve(t, rename(url.Values{"name": {"  Synthetic World Fund "}}.Encode(), ownerSession))
	wantStatus(t, rec, http.StatusSeeOther)
	if loc := rec.Header().Get("Location"); loc != "/instruments/XX0000000001?renamed=1" {
		t.Errorf("Location = %q", loc)
	}
	rec = h.serve(t, withSession(request(http.MethodGet, "/instruments/XX0000000001?renamed=1", ""), ownerSession))
	wantBody(t, rec, "The name is saved.")
	rec = h.serve(t, rename(url.Values{"name": {"Tab\there"}}.Encode(), ownerSession))
	wantStatus(t, rec, http.StatusBadRequest)
	wantBody(t, rec, "The core refused this: invalid instrument name: control character", `value="Tab`)
	for _, body := range []string{"name=+", "name=" + strings.Repeat("x", 201), "name=a&name=b", "other=x"} {
		rec = h.serve(t, rename(body, ownerSession))
		wantStatus(t, rec, http.StatusBadRequest)
	}
	rec = h.serve(t, rename("name=Mine", readSession))
	wantStatus(t, rec, http.StatusForbidden)
	var got []string
	h.names.locked(func() { got = h.names.renamed })
	if diff := cmp.Diff([]string{"XX0000000001=Synthetic World Fund"}, got); diff != "" {
		t.Errorf("renames (-want +got):\n%s", diff)
	}
}

// TestPriceCheckFormsReturnToInstrument checks that a price check form on
// an instrument's page returns there, with a refusal shown on it, and that
// a next field naming another page leads back to the Inbox.
func TestPriceCheckFormsReturnToInstrument(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	ack := url.Values{"isin": {"XX0000000001"}, "symbol": {"SYN"}, "date": {"2026-03-15"}, "source": {"ibkr-main"}, "next": {"/instruments/XX0000000001"}}
	rec := h.serve(t, withSession(form("/inbox/price-checks/acknowledge", ack.Encode()), ownerSession))
	wantStatus(t, rec, http.StatusSeeOther)
	if loc := rec.Header().Get("Location"); loc != "/instruments/XX0000000001" {
		t.Errorf("Location = %q", loc)
	}
	ack.Set("next", "https://elsewhere.example/")
	rec = h.serve(t, withSession(form("/inbox/price-checks/acknowledge", ack.Encode()), ownerSession))
	if loc := rec.Header().Get("Location"); loc != "/inbox" {
		t.Errorf("Location = %q, want the Inbox", loc)
	}
	threshold := url.Values{"isin": {"XX0000000001"}, "symbol": {"SYN"}, "threshold": {"150"}, "next": {"/instruments/XX0000000001"}}
	rec = h.serve(t, withSession(form("/inbox/price-checks/threshold", threshold.Encode()), ownerSession))
	wantStatus(t, rec, http.StatusBadRequest)
	wantBody(t, rec, "<h1>Synthetic Fund</h1>", "the threshold must be a percent")
	threshold.Set("threshold", "100")
	rec = h.serve(t, withSession(form("/inbox/price-checks/threshold", threshold.Encode()), ownerSession))
	wantStatus(t, rec, http.StatusBadRequest)
	threshold.Set("threshold", "")
	rec = h.serve(t, withSession(form("/inbox/price-checks/threshold", threshold.Encode()), ownerSession))
	wantStatus(t, rec, http.StatusSeeOther)
}
