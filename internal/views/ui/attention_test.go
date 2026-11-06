package ui

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"slices"
	"strconv"
	"sync"
	"testing"

	"connectrpc.com/connect"
	"github.com/google/go-cmp/cmp"

	attentionv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/attention/v1"
	"github.com/arkadiuss/openfief/sdk/go/gen/openfief/attention/v1/attentionv1connect"
	ledgerv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/ledger/v1"
	viewsv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/views/v1"
)

// errClearsOnItsOwn is the fake core's reason for refusing to resolve an
// item that resolves when its check clears.
var errClearsOnItsOwn = errors.New("this kind of attention item resolves when its check clears")

// fakeAttention is an in-memory AttentionService over two synthetic open
// items: 1, an unmapped broker account, and 2, an import's left-out rows,
// which the owner may resolve.
type fakeAttention struct {
	attentionv1connect.UnimplementedAttentionServiceHandler

	mu    sync.Mutex
	items []*viewsv1.AttentionItem
	// unavailable makes ListAttentionItems fail, as a core that can't read
	// its database does.
	unavailable bool
	// asked are the ListAttentionItems requests, in order.
	asked []*viewsv1.ListAttentionItemsRequest
}

func (f *fakeAttention) AcknowledgeAttentionItem(_ context.Context, r *connect.Request[attentionv1.AcknowledgeAttentionItemRequest]) (*connect.Response[attentionv1.AcknowledgeAttentionItemResponse], error) {
	m := r.Msg
	if err := f.act(m.GetSessionToken(), m.GetItemId(), viewsv1.AttentionItemState_ATTENTION_ITEM_STATE_ACKNOWLEDGED); err != nil {
		return nil, err
	}
	return connect.NewResponse(&attentionv1.AcknowledgeAttentionItemResponse{}), nil
}

func (f *fakeAttention) ResolveAttentionItem(_ context.Context, r *connect.Request[attentionv1.ResolveAttentionItemRequest]) (*connect.Response[attentionv1.ResolveAttentionItemResponse], error) {
	m := r.Msg
	if err := f.act(m.GetSessionToken(), m.GetItemId(), viewsv1.AttentionItemState_ATTENTION_ITEM_STATE_RESOLVED); err != nil {
		return nil, err
	}
	return connect.NewResponse(&attentionv1.ResolveAttentionItemResponse{}), nil
}

// add adds items to the fake's, raised at a synthetic time.
func (f *fakeAttention) add(items ...*viewsv1.AttentionItem) {
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, it := range items {
		it.History = []*viewsv1.AttentionItemChange{{
			Seq: 1, State: it.GetState(), Reason: viewsv1.AttentionItemReason_ATTENTION_ITEM_REASON_RAISED, At: "2026-03-02T10:00:00Z",
		}}
		f.items = append(f.items, it)
	}
}

func newFakeAttention() *fakeAttention {
	raised := []*viewsv1.AttentionItemChange{{
		Seq: 1, State: viewsv1.AttentionItemState_ATTENTION_ITEM_STATE_OPEN,
		Reason: viewsv1.AttentionItemReason_ATTENTION_ITEM_REASON_RAISED, At: "2026-03-02T10:00:00Z",
	}}
	return &fakeAttention{items: []*viewsv1.AttentionItem{
		{
			Id: 1, Kind: "broker-account-unmapped", Subject: "flex-main", Detail: "last seen 2026-03-02",
			State: viewsv1.AttentionItemState_ATTENTION_ITEM_STATE_OPEN, History: raised,
			Group: viewsv1.AttentionItemGroup_ATTENTION_ITEM_GROUP_DATA_COMING_IN, ConnectorInstance: "flex-main",
		},
		{
			Id: 2, Kind: "import-problems", Subject: "b-7", Detail: "flex-main: 2 rows left out", OwnerResolves: true,
			State: viewsv1.AttentionItemState_ATTENTION_ITEM_STATE_OPEN, History: raised,
			Group: viewsv1.AttentionItemGroup_ATTENTION_ITEM_GROUP_DATA_COMING_IN, ConnectorInstance: "flex-main",
		},
	}}
}

// act moves item id to state for the owner.
func (f *fakeAttention) act(session string, id int64, state viewsv1.AttentionItemState) error {
	if session != ownerSession {
		return codeErr(connect.CodePermissionDenied, nil)
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, it := range f.items {
		if it.GetId() != id {
			continue
		}
		if state == viewsv1.AttentionItemState_ATTENTION_ITEM_STATE_RESOLVED && !it.GetOwnerResolves() {
			return codeErr(connect.CodeFailedPrecondition, errClearsOnItsOwn)
		}
		it.State = state
		return nil
	}
	return codeErr(connect.CodeNotFound, nil)
}

// ListAttentionItems lists the fake's items in the requested view and,
// with an account, those tied to it, counting each state over the account
// filter, as the core does.
// itemSelected reports whether req's account and ISIN filters select it.
func itemSelected(it *viewsv1.AttentionItem, req *viewsv1.ListAttentionItemsRequest) bool {
	if a := req.GetAccountId(); a != "" && !slices.Contains(it.GetAccountIds(), a) {
		return false
	}
	return req.GetIsin() == "" || it.GetIsin() == req.GetIsin()
}

func (f *fakeRead) ListAttentionItems(_ context.Context, r *connect.Request[viewsv1.ListAttentionItemsRequest]) (*connect.Response[viewsv1.ListAttentionItemsResponse], error) {
	if err := f.check(r.Header()); err != nil {
		return nil, err
	}
	f.attention.mu.Lock()
	defer f.attention.mu.Unlock()
	if f.attention.unavailable {
		return nil, codeErr(connect.CodeUnavailable, nil)
	}
	f.attention.asked = append(f.attention.asked, r.Msg)
	out := &viewsv1.ListAttentionItemsResponse{}
	counts := map[viewsv1.AttentionView]int32{}
	for _, it := range f.attention.items {
		if !itemSelected(it, r.Msg) {
			continue
		}
		if it.GetKind() == kindTransferSuggested && f.transfers.decided() {
			// The core's check clears the item once the pair is decided.
			it.State = viewsv1.AttentionItemState_ATTENTION_ITEM_STATE_RESOLVED
		}
		view := viewOfState(it.GetState())
		counts[view]++
		if r.Msg.GetView() == viewsv1.AttentionView_ATTENTION_VIEW_UNSPECIFIED || r.Msg.GetView() == view {
			out.Items = append(out.Items, it)
		}
	}
	out.Open, out.Acknowledged = counts[viewsv1.AttentionView_ATTENTION_VIEW_OPEN], counts[viewsv1.AttentionView_ATTENTION_VIEW_ACKNOWLEDGED]
	out.Done = counts[viewsv1.AttentionView_ATTENTION_VIEW_DONE]
	return connect.NewResponse(out), nil
}

// viewOfState returns the view an item in state is listed in.
func viewOfState(state viewsv1.AttentionItemState) viewsv1.AttentionView {
	switch state {
	case viewsv1.AttentionItemState_ATTENTION_ITEM_STATE_OPEN:
		return viewsv1.AttentionView_ATTENTION_VIEW_OPEN
	case viewsv1.AttentionItemState_ATTENTION_ITEM_STATE_ACKNOWLEDGED:
		return viewsv1.AttentionView_ATTENTION_VIEW_ACKNOWLEDGED
	case viewsv1.AttentionItemState_ATTENTION_ITEM_STATE_RESOLVED, viewsv1.AttentionItemState_ATTENTION_ITEM_STATE_UNSPECIFIED:
	}
	return viewsv1.AttentionView_ATTENTION_VIEW_DONE
}

// itemPost posts item id to path.
func itemPost(path string, id int) *http.Request {
	r := request(http.MethodPost, path, url.Values{"item": {strconv.Itoa(id)}}.Encode())
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return r
}

// TestInboxSettlesAttentionItems checks the Inbox (D-196): both items in
// the data coming in group with the instance's accounts page linked, the
// open count in the navigation and the views' counts, the overview's link,
// and the owner acknowledging one and resolving the other, each back to
// the Inbox, where they move to their views.
func TestInboxSettlesAttentionItems(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	rec := h.serve(t, withSession(request(http.MethodGet, "/inbox", ""), ownerSession))
	wantStatus(t, rec, http.StatusOK)
	wantBody(t, rec, "<title>Inbox", `<a href="/inbox" aria-current="page">Inbox&nbsp;<span class="pill pill-gold" title="2 open">2`,
		"Data coming in · 2", "Map a broker account flex-main reported", "An import left rows out", `href="/accounts"`,
		`action="/inbox/acknowledge"`, `action="/inbox/resolve"`, `name="item" value="2"`, "Open <span class=\"count\">2</span>",
		"History: 1 change(s)", "Numbers to check · 0", "Decisions · 0")
	wantNoBody(t, rec, "<script>", "style=")

	rec = h.serve(t, withSession(request(http.MethodGet, "/", ""), ownerSession))
	wantBody(t, rec, `<a href="/inbox">2 attention item(s)</a>`)

	for _, step := range []struct {
		path string
		id   int
	}{{"/inbox/acknowledge", 1}, {"/inbox/resolve", 2}} {
		rec = h.serve(t, withSession(itemPost(step.path, step.id), ownerSession))
		wantStatus(t, rec, http.StatusSeeOther)
		if got := rec.Header().Get("Location"); got != "/inbox" {
			t.Errorf("POST %s: Location = %q, want /inbox", step.path, got)
		}
	}
	rec = h.serve(t, withSession(request(http.MethodGet, "/inbox", ""), ownerSession))
	wantBody(t, rec, "Data coming in · 0", "Acknowledged <span class=\"count\">1</span>", "Done <span class=\"count\">1</span>")
	wantNoBody(t, rec, "pill-gold")
	rec = h.serve(t, withSession(request(http.MethodGet, "/inbox?view=acknowledged", ""), ownerSession))
	wantBody(t, rec, "Map a broker account flex-main reported", `<span class="pill pill-neutral">Acknowledged</span>`)
	wantNoBody(t, rec, "An import left rows out")
	rec = h.serve(t, withSession(request(http.MethodGet, "/inbox?view=done", ""), ownerSession))
	wantBody(t, rec, "An import left rows out", `<span class="pill pill-ok">Done</span>`)
}

// TestInboxRefusals checks that a read session sees the items without the
// owner's buttons and can't act, that the core's refusal shows in the
// Inbox, that a malformed form is refused before the core, and that a core
// that can't count the items doesn't break other pages.
func TestInboxRefusals(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	rec := h.serve(t, withSession(request(http.MethodGet, "/inbox", ""), readSession))
	wantStatus(t, rec, http.StatusOK)
	wantBody(t, rec, "Map a broker account flex-main reported")
	wantNoBody(t, rec, `action="/inbox/acknowledge"`, `action="/inbox/resolve"`, `action="/accounts/bind"`)

	rec = h.serve(t, withSession(itemPost("/inbox/acknowledge", 1), readSession))
	wantStatus(t, rec, http.StatusForbidden)

	rec = h.serve(t, withSession(itemPost("/inbox/resolve", 1), ownerSession))
	wantStatus(t, rec, http.StatusBadRequest)
	wantBody(t, rec, "The core refused this: "+errClearsOnItsOwn.Error())

	rec = h.serve(t, withSession(itemPost("/inbox/acknowledge", 0), ownerSession))
	wantStatus(t, rec, http.StatusBadRequest)

	h.attention.mu.Lock()
	h.attention.unavailable = true
	h.attention.mu.Unlock()
	rec = h.serve(t, withSession(request(http.MethodGet, "/accounts", ""), ownerSession))
	wantStatus(t, rec, http.StatusOK)
	wantNoBody(t, rec, "pill-gold")
}

// TestBrokenTransfer checks that a confirmed transfer a correction broke
// (D-131) shows in the Inbox with both legs' accounts linked and the
// owner's button to reject it.
func TestBrokenTransfer(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	h.transfers.mu.Lock()
	h.transfers.decision, h.transfers.broken = ledgerv1.TransferDecision_TRANSFER_DECISION_CONFIRMED, true
	h.transfers.mu.Unlock()
	h.attention.add(&viewsv1.AttentionItem{
		Id: 7, Kind: "transfer-broken", Subject: "test/w1 -> test/d1", Detail: "its legs no longer match",
		State: viewsv1.AttentionItemState_ATTENTION_ITEM_STATE_OPEN, Group: viewsv1.AttentionItemGroup_ATTENTION_ITEM_GROUP_DECISIONS,
		AccountId: "acc-1", AccountIds: []string{"acc-1", "acc-2"},
	})
	rec := h.serve(t, withSession(request(http.MethodGet, "/inbox", ""), ownerSession))
	wantStatus(t, rec, http.StatusOK)
	wantBody(t, rec, "A confirmed transfer no longer matches", "2\u202f499,50 PLN", "Reject the pair",
		`<a href="/accounts/acc-1">Open account acc-1</a>`, `<a href="/accounts/acc-2">Open account acc-2</a>`)
	rec = h.serve(t, withSession(transferPost("/inbox/transfers/reject", "w1"), ownerSession))
	wantStatus(t, rec, http.StatusSeeOther)
}

// TestAttentionKindOf checks how the Inbox names each of the
// core's kinds and links where the owner acts on it, an instrument to
// classify (D-132) included, and that an unknown kind shows its own name
// without a link.
func TestAttentionKindOf(t *testing.T) {
	t.Parallel()
	tests := []struct {
		kind string
		want attentionKind
	}{
		{"balance-assertion-flagged", attentionKind{Label: "Flagged balance assertion"}},
		{"price-check-flagged", attentionKind{Label: "Flagged price check"}},
		{"import-problems", attentionKind{"Rows an import left out", "/accounts", "Open the accounts"}},
		{"broker-account-unmapped", attentionKind{"Unmapped broker account", "/accounts", "Open the accounts"}},
		{"ticker-unconfirmed", attentionKind{Label: "Ticker to confirm"}},
		{"setting-missing", attentionKind{"Missing setting", "/settings", "Open the settings"}},
		{"transfer-broken", attentionKind{Label: "Transfer broken by a correction"}},
		{"instrument-unclassified", attentionKind{"Instrument to classify", "/settings", "Classify it in the settings"}},
		{"market-data-gap", attentionKind{"Market data stopped arriving", "/settings#series-sources-title", "Open the sources"}},
		{"connector-sync-failing", attentionKind{"Sync keeps failing", "/plugins", "Open the plugins"}},
		{"cash-checkpoint-mismatch", attentionKind{Label: "Cash balance differs"}},
		{"some-later-check", attentionKind{Label: "some-later-check"}},
	}
	for _, tt := range tests {
		t.Run(tt.kind, func(t *testing.T) {
			t.Parallel()
			if diff := cmp.Diff(tt.want, attentionKindOf(tt.kind)); diff != "" {
				t.Errorf("attentionKindOf(%q) (-want +got):\n%s", tt.kind, diff)
			}
		})
	}
}
