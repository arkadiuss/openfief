package ui

import (
	"context"
	"errors"
	"net/http"
	"net/url"
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
}

func newFakeAttention() *fakeAttention {
	raised := []*viewsv1.AttentionItemChange{{
		Seq: 1, State: viewsv1.AttentionItemState_ATTENTION_ITEM_STATE_OPEN,
		Reason: viewsv1.AttentionItemReason_ATTENTION_ITEM_REASON_RAISED, At: "2026-03-02T10:00:00Z",
	}}
	return &fakeAttention{items: []*viewsv1.AttentionItem{
		{
			Id: 1, Kind: "broker-account-unmapped", Subject: "flex-main/U0000001", Detail: "last seen 2026-03-02",
			State: viewsv1.AttentionItemState_ATTENTION_ITEM_STATE_OPEN, History: raised,
		},
		{
			Id: 2, Kind: "import-problems", Subject: "b-7", Detail: "flex-main: 2 rows left out", OwnerResolves: true,
			State: viewsv1.AttentionItemState_ATTENTION_ITEM_STATE_OPEN, History: raised,
		},
	}}
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

// ListAttentionItems lists the fake's items.
func (f *fakeRead) ListAttentionItems(_ context.Context, r *connect.Request[viewsv1.ListAttentionItemsRequest]) (*connect.Response[viewsv1.ListAttentionItemsResponse], error) {
	if err := f.check(r.Header()); err != nil {
		return nil, err
	}
	f.attention.mu.Lock()
	defer f.attention.mu.Unlock()
	if f.attention.unavailable {
		return nil, codeErr(connect.CodeUnavailable, nil)
	}
	out := &viewsv1.ListAttentionItemsResponse{Items: f.attention.items}
	for _, it := range f.attention.items {
		if it.GetState() == viewsv1.AttentionItemState_ATTENTION_ITEM_STATE_OPEN {
			out.Open++
		}
	}
	return connect.NewResponse(out), nil
}

// itemPost posts item id to path.
func itemPost(path string, id int) *http.Request {
	r := request(http.MethodPost, path, url.Values{"item": {strconv.Itoa(id)}}.Encode())
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return r
}

// TestAttentionPage checks the attention page (D-130): both items in one
// place with links to where the owner acts, the open count in the
// navigation and on the overview, and the owner acknowledging one and
// resolving the other.
func TestAttentionPage(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	rec := h.serve(t, withSession(request(http.MethodGet, "/attention", ""), ownerSession))
	wantStatus(t, rec, http.StatusOK)
	wantBody(t, rec, "Unmapped broker account", `href="/accounts"`, "flex-main/U0000001", "Rows an import left out", "b-7",
		`action="/attention/acknowledge"`, `action="/attention/resolve"`, `name="item" value="2"`, "2 open", "History: 1 change(s)")
	wantNoBody(t, rec, "<script>", "style=")

	rec = h.serve(t, withSession(request(http.MethodGet, "/", ""), ownerSession))
	wantBody(t, rec, `<a href="/attention">2 attention item(s)</a>`)

	for _, step := range []struct {
		path string
		id   int
	}{{"/attention/acknowledge", 1}, {"/attention/resolve", 2}} {
		rec = h.serve(t, withSession(itemPost(step.path, step.id), ownerSession))
		wantStatus(t, rec, http.StatusSeeOther)
		if got := rec.Header().Get("Location"); got != "/attention" {
			t.Errorf("POST %s: Location = %q, want /attention", step.path, got)
		}
	}
	rec = h.serve(t, withSession(request(http.MethodGet, "/attention", ""), ownerSession))
	wantBody(t, rec, "Nothing needs you.", "Acknowledged", "1 resolved item(s)")
	wantNoBody(t, rec, " open</span>")
}

// TestAttentionRefusals checks that a read session sees the items without
// the owner's buttons and can't act, that the core's refusal shows on the
// page, that a malformed form is refused before the core, and that a core
// that can't count the items doesn't break other pages.
func TestAttentionRefusals(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	rec := h.serve(t, withSession(request(http.MethodGet, "/attention", ""), readSession))
	wantStatus(t, rec, http.StatusOK)
	wantBody(t, rec, "flex-main/U0000001")
	wantNoBody(t, rec, `action="/attention/acknowledge"`, `action="/attention/resolve"`)

	rec = h.serve(t, withSession(itemPost("/attention/acknowledge", 1), readSession))
	wantStatus(t, rec, http.StatusForbidden)

	rec = h.serve(t, withSession(itemPost("/attention/resolve", 1), ownerSession))
	wantStatus(t, rec, http.StatusBadRequest)
	wantBody(t, rec, "The core refused this: "+errClearsOnItsOwn.Error())

	rec = h.serve(t, withSession(itemPost("/attention/acknowledge", 0), ownerSession))
	wantStatus(t, rec, http.StatusBadRequest)

	h.attention.mu.Lock()
	h.attention.unavailable = true
	h.attention.mu.Unlock()
	rec = h.serve(t, withSession(request(http.MethodGet, "/transfers", ""), ownerSession))
	wantStatus(t, rec, http.StatusOK)
	wantNoBody(t, rec, " open</span>")
}

// TestBrokenTransfer checks that a confirmed transfer a correction broke
// (D-131) is listed apart, with the owner's button to reject it.
func TestBrokenTransfer(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	h.transfers.mu.Lock()
	h.transfers.decision, h.transfers.broken = ledgerv1.TransferDecision_TRANSFER_DECISION_CONFIRMED, true
	h.transfers.mu.Unlock()
	rec := h.serve(t, withSession(request(http.MethodGet, "/transfers", ""), ownerSession))
	wantStatus(t, rec, http.StatusOK)
	wantBody(t, rec, "Broken by a correction", "2 499,50 PLN", "Reject the pair", "No confirmed transfers yet.")
	rec = h.serve(t, withSession(transferPost("/transfers/reject", "w1"), ownerSession))
	wantStatus(t, rec, http.StatusSeeOther)
}

// TestAttentionKindOf checks how the attention page names each of the
// core's kinds and links where the owner acts on it, an instrument to
// classify (D-132) included, and that an unknown kind shows its own name
// without a link.
func TestAttentionKindOf(t *testing.T) {
	t.Parallel()
	tests := []struct {
		kind, subject string
		want          attentionKind
	}{
		{"cash-checkpoint-mismatch", "acc/1/2026-03-31/PLN", attentionKind{"Cash checkpoint differs", "/accounts/acc%2F1#cash-checkpoints"}},
		{"cash-checkpoint-mismatch", "malformed", attentionKind{Label: "Cash checkpoint differs"}},
		{"balance-assertion-flagged", "", attentionKind{"Flagged balance assertion", "/assertions"}},
		{"price-check-flagged", "", attentionKind{"Flagged price check", "/price-checks"}},
		{"import-problems", "", attentionKind{"Rows an import left out", "/accounts"}},
		{"broker-account-unmapped", "", attentionKind{"Unmapped broker account", "/accounts"}},
		{"ticker-unconfirmed", "", attentionKind{"Ticker to confirm", "/tickers"}},
		{"setting-missing", "", attentionKind{"Missing setting", "/settings"}},
		{"transfer-broken", "", attentionKind{"Transfer broken by a correction", "/transfers"}},
		{"instrument-unclassified", "", attentionKind{"Instrument to classify", "/settings"}},
		{"market-data-gap", "", attentionKind{"Market data stopped arriving", "/settings#series-sources-title"}},
		{"cross-check-prices-undeclared", "flex-main", attentionKind{"Undeclared cross-check prices dropped", "/plugins"}},
		{"some-later-check", "", attentionKind{Label: "some-later-check"}},
	}
	for _, tt := range tests {
		t.Run(tt.kind+" "+tt.subject, func(t *testing.T) {
			t.Parallel()
			if diff := cmp.Diff(tt.want, attentionKindOf(tt.kind, tt.subject)); diff != "" {
				t.Errorf("attentionKindOf(%q, %q) (-want +got):\n%s", tt.kind, tt.subject, diff)
			}
		})
	}
}
