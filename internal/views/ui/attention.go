package ui

import (
	"net/http"
	"strconv"

	"connectrpc.com/connect"

	attentionv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/attention/v1"
	viewsv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/views/v1"
)

// attentionKind is how the attention page names a kind of item and where
// the owner acts on it.
type attentionKind struct {
	Label string
	Page  string
}

// attentionKindOf names kind, one of the core's kinds of attention items
// (D-130, D-131, D-132, D-94, D-179); a kind it doesn't know shows its own name and no page.
func attentionKindOf(kind string) attentionKind {
	kinds := map[string]attentionKind{
		"balance-assertion-flagged":             {"Flagged balance assertion", "/assertions"},
		"price-check-flagged":                   {"Flagged price check", "/price-checks"},
		"import-problems":                       {"Rows an import left out", "/accounts"},
		"broker-account-unmapped":               {"Unmapped broker account", "/accounts"},
		"ticker-unconfirmed":                    {"Ticker to confirm", "/tickers"},
		"setting-missing":                       {"Missing setting", "/settings"},
		"transfer-broken":                       {"Transfer broken by a correction", "/transfers"},
		"instrument-unclassified":               {"Instrument to classify", "/settings"},
		"market-data-gap":                       {"Market data stopped arriving", "/settings#series-sources-title"},
		"bond-terms-or-index-levels-undeclared": {"Bond terms, index levels or readings dropped", "/plugins"},
	}
	if k, ok := kinds[kind]; ok {
		return k
	}
	return attentionKind{Label: kind}
}

// attentionRow is one item as the attention page shows it, with the
// owner's actions it offers.
type attentionRow struct {
	Item        *viewsv1.AttentionItem
	Kind        attentionKind
	Acknowledge bool
	Resolve     bool
}

// attentionView is what the attention page shows: the items by state, and
// the core's reason for refusing the owner's last action.
type attentionView struct {
	Open, Acknowledged, Resolved []attentionRow
	Error                        string
}

// attentionPage lists every attention item in one place (D-130): open
// ones first, then acknowledged ones, then resolved ones, each with its
// history and a link to the page where the owner acts on it. Every
// session sees them; only the owner's gets the buttons.
func (s *server) attentionPage(w http.ResponseWriter, r *http.Request, sess session) {
	s.renderAttention(w, r, sess, http.StatusOK, "")
}

// renderAttention shows the attention page with status and the core's
// reason for a refused action, if any.
func (s *server) renderAttention(w http.ResponseWriter, r *http.Request, sess session, status int, refusal string) {
	resp, err := s.read.ListAttentionItems(r.Context(), bearer(&viewsv1.ListAttentionItemsRequest{}, sess))
	if err != nil {
		s.coreFailed(w, r, "ListAttentionItems", err)
		return
	}
	view := attentionView{Error: refusal}
	owner := sess.Owner()
	for _, it := range resp.Msg.GetItems() {
		row := attentionRow{Item: it, Kind: attentionKindOf(it.GetKind())}
		switch it.GetState() {
		case viewsv1.AttentionItemState_ATTENTION_ITEM_STATE_OPEN:
			row.Acknowledge, row.Resolve = owner, owner && it.GetOwnerResolves()
			view.Open = append(view.Open, row)
		case viewsv1.AttentionItemState_ATTENTION_ITEM_STATE_ACKNOWLEDGED:
			row.Resolve = owner && it.GetOwnerResolves()
			view.Acknowledged = append(view.Acknowledged, row)
		case viewsv1.AttentionItemState_ATTENTION_ITEM_STATE_RESOLVED, viewsv1.AttentionItemState_ATTENTION_ITEM_STATE_UNSPECIFIED:
			view.Resolved = append(view.Resolved, row)
		}
	}
	s.show(w, r, status, "attention", page{Title: "Attention", Session: &sess, Data: view})
}

// acknowledgeAttentionItem hands the owner's acknowledgement of the posted
// item to the core and shows the page again.
func (s *server) acknowledgeAttentionItem(w http.ResponseWriter, r *http.Request, sess session) {
	id, ok := postedItem(r)
	if !ok {
		s.errorPage(w, r, http.StatusBadRequest, "The form is malformed.")
		return
	}
	_, err := s.attention.AcknowledgeAttentionItem(r.Context(), connect.NewRequest(&attentionv1.AcknowledgeAttentionItemRequest{
		Origin: sess.Origin, SessionToken: sess.Token, ItemId: id,
	}))
	s.afterAttentionForm(w, r, sess, "AcknowledgeAttentionItem", err)
}

// resolveAttentionItem hands the owner's resolution of the posted item to
// the core and shows the page again.
func (s *server) resolveAttentionItem(w http.ResponseWriter, r *http.Request, sess session) {
	id, ok := postedItem(r)
	if !ok {
		s.errorPage(w, r, http.StatusBadRequest, "The form is malformed.")
		return
	}
	_, err := s.attention.ResolveAttentionItem(r.Context(), connect.NewRequest(&attentionv1.ResolveAttentionItemRequest{
		Origin: sess.Origin, SessionToken: sess.Token, ItemId: id,
	}))
	s.afterAttentionForm(w, r, sess, "ResolveAttentionItem", err)
}

// postedItem reads the posted item ID; ok is false for a malformed form.
func postedItem(r *http.Request) (int64, bool) {
	if err := r.ParseForm(); err != nil {
		return 0, false
	}
	id, err := strconv.ParseInt(r.PostForm.Get("item"), 10, 64)
	return id, err == nil && id > 0
}

// afterAttentionForm answers a posted attention form: back to the page
// when the core took it, the page with the core's reason when it refused
// it, and any other failure like any failed call to the core.
func (s *server) afterAttentionForm(w http.ResponseWriter, r *http.Request, sess session, name string, err error) {
	if err == nil {
		http.Redirect(w, r, "/attention", http.StatusSeeOther)
		return
	}
	code := connect.CodeOf(err)
	if status, refused := refusalStatus(code); refused && code != connect.CodePermissionDenied {
		s.renderAttention(w, r, sess, status, "The core refused this: "+connectMessage(err))
		return
	}
	s.coreFailed(w, r, name, err)
}

// openAttention returns how many attention items are open, for the
// navigation and the overview; a core that can't tell counts none, and
// the page still renders.
func (s *server) openAttention(r *http.Request, sess session) int {
	resp, err := s.read.ListAttentionItems(r.Context(), bearer(&viewsv1.ListAttentionItemsRequest{}, sess))
	if err != nil {
		s.log.WarnContext(r.Context(), "counting attention items", "code", connect.CodeOf(err).String())
		return 0
	}
	return int(resp.Msg.GetOpen())
}
