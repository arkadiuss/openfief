package ui

import (
	"context"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"connectrpc.com/connect"

	viewsv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/views/v1"
)

// ListSyncStatuses lists xtb-main, whose sync keeps failing, and nbp-fx.
func (f *fakeRead) ListSyncStatuses(_ context.Context, r *connect.Request[viewsv1.ListSyncStatusesRequest]) (*connect.Response[viewsv1.ListSyncStatusesResponse], error) {
	if err := f.check(r.Header()); err != nil {
		return nil, err
	}
	var down bool
	f.locked(func() { down = f.syncsUnavailable })
	if down {
		return nil, codeErr(connect.CodeUnavailable, nil)
	}
	return connect.NewResponse(&viewsv1.ListSyncStatusesResponse{Statuses: []*viewsv1.SyncStatus{
		{ConnectorInstance: "nbp-fx", LastAttemptAt: "2026-03-31T18:00:00Z", LastSuccessAt: "2026-03-31T18:00:00Z"},
		{ConnectorInstance: "xtb-main", LastAttemptAt: "2026-03-03T07:00:00Z", Failures: 3, FailingSince: "2026-03-03T06:00:00Z", KeepsFailing: true},
	}}), nil
}

// ListActivity lists a sync into the synthetic account and a coupon.
func (f *fakeRead) ListActivity(_ context.Context, r *connect.Request[viewsv1.ListActivityRequest]) (*connect.Response[viewsv1.ListActivityResponse], error) {
	if err := f.check(r.Header()); err != nil {
		return nil, err
	}
	return connect.NewResponse(&viewsv1.ListActivityResponse{Activities: []*viewsv1.Activity{
		{
			Kind: viewsv1.ActivityKind_ACTIVITY_KIND_SYNC, At: "2026-03-31T18:00:00Z", ConnectorInstance: "ibkr-main",
			Batch: &viewsv1.ActivityBatch{BatchId: "b-9", NewTransactions: 4, AccountIds: []string{accountID}},
		},
		{
			Kind: viewsv1.ActivityKind_ACTIVITY_KIND_COUPON, At: "2026-03-30T08:00:00Z",
			Change: &viewsv1.ActivityChange{AccountId: accountID, Instrument: syntheticInstrument(), Amount: pln("12.50"), Kind: viewsv1.TransactionKind_TRANSACTION_KIND_INTEREST},
		},
	}}), nil
}

// GetInstrument answers the synthetic instrument, and the synthetic
// treasury bonds (bondDetail).
func (f *fakeRead) GetInstrument(_ context.Context, r *connect.Request[viewsv1.GetInstrumentRequest]) (*connect.Response[viewsv1.GetInstrumentResponse], error) {
	if err := f.check(r.Header()); err != nil {
		return nil, err
	}
	in := syntheticInstrument()
	if isin := r.Msg.GetIsin(); isin == assumedBond || isin == pendingBond {
		return connect.NewResponse(&viewsv1.GetInstrumentResponse{Instrument: bondDetail(isin)}), nil
	}
	if r.Msg.GetIsin() != in.GetIsin() {
		return nil, codeErr(connect.CodeNotFound, nil)
	}
	return connect.NewResponse(&viewsv1.GetInstrumentResponse{Instrument: &viewsv1.InstrumentDetail{
		Isin: in.GetIsin(), Name: in.GetName(), AssetClass: "listed",
		ExemptionClasses: []*viewsv1.InstrumentExemptionClass{{Wrapper: "oki", ExemptionClass: "investment", Known: true}},
		Listings:         []*viewsv1.InstrumentListing{{Instrument: in, PriceSource: "eodhd", LatestPrice: &viewsv1.InstrumentPrice{Date: "2026-03-31", Close: &viewsv1.Money{Amount: "25.00", Currency: "EUR"}}}},
	}}), nil
}

// ListInstrumentHoldings has the synthetic account hold ten units.
func (f *fakeRead) ListInstrumentHoldings(_ context.Context, r *connect.Request[viewsv1.ListInstrumentHoldingsRequest]) (*connect.Response[viewsv1.ListInstrumentHoldingsResponse], error) {
	if err := f.check(r.Header()); err != nil {
		return nil, err
	}
	return connect.NewResponse(&viewsv1.ListInstrumentHoldingsResponse{Isin: r.Msg.GetIsin(), Date: r.Msg.GetDate(), Holdings: []*viewsv1.InstrumentHolding{
		{AccountId: accountID, Instrument: syntheticInstrument(), Quantity: "10"},
	}}), nil
}

// TestOverviewHome checks the overview's places (D-193): the status
// sentence with what needs the owner and the failing sync, the Ask box with
// its chips, the value card, the chart card's tabs, Needs you, the
// accounts under their sleeve, the allocation's bars and Recent activity.
func TestOverviewHome(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	rec := h.serve(t, withSession(request(http.MethodGet, "/", ""), ownerSession))
	wantStatus(t, rec, http.StatusOK)
	wantBody(t, rec,
		"Everything is valued through 2026-03-31. 2 things need you. The xtb-main sync keeps failing.",
		`id="ask-q" name="q"`, `<kbd class="key"`, "holds no AI model", `hx-get="/ask?q=What&#43;needs&#43;me%3F"`,
		`href="/?q=Why&#43;is&#43;xtb-main&#43;failing%3F#ask"`,
		`class="card hero"`, `<a class="figure" href="#accounts"`, "1 500,00 PLN",
		`id="over-title">Over time`, `aria-current="page">Value</a>`, `hx-get="/over-time?date=2026-03-31&amp;over=performance"`,
		`id="needs-title">Needs you`, "Open inbox (2)", "Map a broker account flex-main reported",
		`id="sleeve-core"`, `<a class="account-row" href="/accounts/acc-synthetic-1">`, "Valued 2026-03-31", `id="sleeve-core"`,
		`hx-get="/over-time?by=asset-class&amp;date=2026-03-31&amp;over=allocation&amp;value=gross"`,
		`id="activity-title">Recent activity`, "Synced · ibkr-main · 4 new transactions", "Coupon booked",
		`<script src="/static/ask.js" defer></script>`,
	)
	wantNoBody(t, rec, "<script>", "style=")
}

// TestOverTimeTabs checks the chart card's tabs: the performance tab as a
// whole page and as an htmx fragment, and the three values.
func TestOverTimeTabs(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	rec := h.serve(t, withSession(request(http.MethodGet, "/?over=performance", ""), readSession))
	wantStatus(t, rec, http.StatusOK)
	wantBody(t, rec, `aria-current="page">Performance</a>`, `id="history-title">Performance`, "time-weighted return as an index")
	frag := h.serve(t, htmx(withSession(request(http.MethodGet, "/over-time?date=2026-03-31&chart=all", ""), readSession)))
	wantStatus(t, frag, http.StatusOK)
	wantBody(t, frag, `id="over-time"`, `aria-current="page">Gross · net · liquidation</a>`, "Liquidation value with cash")
	wantNoBody(t, frag, "<html")
	alloc := h.serve(t, htmx(withSession(request(http.MethodGet, "/over-time?date=2026-03-31&over=allocation&by=currency", ""), readSession)))
	wantStatus(t, alloc, http.StatusOK)
	wantBody(t, alloc, `aria-current="page">Allocation</a>`, `<section id="allocation"`, `class="bars"`, `class="fill series-1" width="66.7"`,
		`aria-current="true">Currency</a>`, `<a class="bar" href="/holdings?by=currency&amp;date=2026-03-31&amp;key=`)
	wantNoBody(t, alloc, "Charts start on")
	plain := h.serve(t, withSession(request(http.MethodGet, "/over-time?date=2026-03-31&over=performance", ""), readSession))
	wantStatus(t, plain, http.StatusSeeOther)
}

// TestAskBox checks the Ask box (D-194): a few commands answered as cards
// with one Open link, a question it doesn't understand answered with the
// commands it does, the answer on the overview without htmx, and the
// question in no log line.
func TestAskBox(t *testing.T) {
	t.Parallel()
	tests := []struct {
		q    string
		want []string
	}{
		{"What needs me?", []string{"2 things need you", "Open the Inbox (2)", "Map a broker account flex-main reported", "Data coming in"}},
		{"Why is xtb failing?", []string{"The xtb-main sync keeps failing since 2026-03-03T06:00:00Z", "timeout", "Open the failing sync in the Inbox"}},
		{"allocation by currency", []string{"By currency, gross, on 2026-03-31", "66.7%", "Open the allocation"}},
		{"Performance this year", []string{"Time-weighted return this year: &#43;1,5%", "Show the performance chart"}},
		{"value on 2026-03-31", []string{"Value on 2026-03-31: 1 500,00 PLN gross", "Liquidation", "Open the overview on 2026-03-31"}},
		{"show synthetic brokerage", []string{"Synthetic brokerage: 1 000,00 PLN gross", "Open Synthetic brokerage", `href="/accounts/acc-synthetic-1"`}},
		{"show sleeve core", []string{"Core: 1 account", "Open Core on the overview"}},
		{"show instrument SYN", []string{"Synthetic Fund · XX0000000001", `href="/instruments/XX0000000001"`, "Open Synthetic Fund · XX0000000001"}},
		{"show SYN", []string{"Synthetic Fund", "10 units", "last priced 25,00 EUR"}},
		{"buy more bonds", []string{"I can&#39;t answer that yet", "What needs me", "Allocation by currency"}},
	}
	h := newHarness(t)
	for _, tt := range tests {
		t.Run(tt.q, func(t *testing.T) {
			t.Parallel()
			rec := h.serve(t, htmx(withSession(request(http.MethodGet, "/ask?"+url.Values{"q": {tt.q}}.Encode(), ""), readSession)))
			wantStatus(t, rec, http.StatusOK)
			wantBody(t, rec, append([]string{`class="bubble"`, "Answered from your data · no AI model"}, tt.want...)...)
			wantNoBody(t, rec, "<html", "<form")
		})
	}
	page := h.serve(t, withSession(request(http.MethodGet, "/?q=what+needs+me", ""), ownerSession))
	wantStatus(t, page, http.StatusOK)
	wantBody(t, page, `value="what needs me"`, `class="bubble"`, "Open the Inbox (2)")
	plain := h.serve(t, withSession(request(http.MethodGet, "/ask?q=value", ""), ownerSession))
	wantStatus(t, plain, http.StatusSeeOther)
	if got := plain.Header().Get("Location"); got != "/?q=value#ask" {
		t.Errorf("Location = %q, want /?q=value#ask", got)
	}
}

// TestAskNeverLogsTheQuestion checks that a question whose answer needs a
// core that fails is logged by method and code only (GR-12).
func TestAskNeverLogsTheQuestion(t *testing.T) {
	t.Parallel()
	var logged strings.Builder
	h := newHarnessLogging(t, slog.New(slog.NewTextHandler(&logged, nil)))
	h.read.locked(func() { h.read.syncsUnavailable = true })
	rec := h.serve(t, htmx(withSession(request(http.MethodGet, "/ask?q=why+is+secret-question+failing", ""), ownerSession)))
	wantStatus(t, rec, http.StatusBadGateway)
	if strings.Contains(logged.String(), "secret-question") || !strings.Contains(logged.String(), "call to the core failed") {
		t.Errorf("log = %q, want the failure without the question", logged.String())
	}
}

// TestInboxAskColumn checks the Inbox's side column (D-194): questions
// about the open items, each answered on the overview.
func TestInboxAskColumn(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	h.attention.add(&viewsv1.AttentionItem{
		Id: 7, Kind: kindSyncFailing, Subject: "xtb-main", Detail: "failing since 2026-03-03T06:00Z", ConnectorInstance: "xtb-main",
		State: viewsv1.AttentionItemState_ATTENTION_ITEM_STATE_OPEN, Group: viewsv1.AttentionItemGroup_ATTENTION_ITEM_GROUP_DATA_COMING_IN,
	})
	rec := h.serve(t, withSession(request(http.MethodGet, "/inbox", ""), readSession))
	wantStatus(t, rec, http.StatusOK)
	wantBody(t, rec, `id="ask-about-title">Ask about these`, `href="/?q=What&#43;needs&#43;me%3F#ask"`, `href="/?q=Why&#43;is&#43;xtb-main&#43;failing%3F#ask"`)
	done := h.serve(t, withSession(request(http.MethodGet, "/inbox?view=done", ""), readSession))
	wantNoBody(t, done, "Ask about these")
}
