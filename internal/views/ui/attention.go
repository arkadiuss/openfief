package ui

import (
	"net/http"
	"strconv"
	"strings"

	"connectrpc.com/connect"

	attentionv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/attention/v1"
	viewsv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/views/v1"
)

// attentionKind is how the Inbox names a kind of item, and, for a kind
// not settled in place, the page where the owner settles it and the
// link's words (D-196).
type attentionKind struct {
	Label  string
	Page   string
	Action string
}

// The core's kinds of attention items the UI tells apart (D-130, D-131,
// D-132, D-94, D-105, D-158, D-179, D-196).
const (
	kindAssertion         = "balance-assertion-flagged"
	kindPriceCheck        = "price-check-flagged"
	kindImportProblems    = "import-problems"
	kindBrokerUnmapped    = "broker-account-unmapped"
	kindTicker            = "ticker-unconfirmed"
	kindSettingMissing    = "setting-missing"
	kindTransferBroken    = "transfer-broken"
	kindTransferSuggested = "transfer-suggested"
	kindUnclassified      = "instrument-unclassified"
	kindSyncFailing       = "connector-sync-failing"
	kindCashCheckpoint    = "cash-checkpoint-mismatch"
	kindBondFeeAssumed    = "bond-fee-assumed"
	kindBondSeries        = "bond-series-incomplete"
	// kindViewPluginUngranted is a view plugin that can't read anything until
	// the owner grants it accounts (D-196, D-209).
	kindViewPluginUngranted = "view-plugin-ungranted"
)

// openAccounts is the words of an item's link to the accounts page.
const openAccounts = "Open the accounts"

// attentionKindOf names kind, one of the core's kinds of attention items
// (D-130, D-131, D-132, D-94, D-141, D-179, D-196); a kind it doesn't know shows
// its own name and no page. Kinds the Inbox settles in place name no page:
// their item links its account instead.
func attentionKindOf(kind string) attentionKind {
	kinds := map[string]attentionKind{
		kindAssertion:                           {Label: "Flagged balance assertion"},
		kindPriceCheck:                          {Label: "Flagged price check"},
		kindImportProblems:                      {Label: "Rows an import left out", Page: accountsPath, Action: openAccounts},
		kindBrokerUnmapped:                      {Label: "Unmapped broker account", Page: accountsPath, Action: openAccounts},
		kindTicker:                              {Label: "Ticker to confirm"},
		kindSettingMissing:                      {Label: "Missing setting", Page: "/settings", Action: "Open the settings"},
		kindTransferBroken:                      {Label: "Transfer broken by a correction"},
		kindTransferSuggested:                   {Label: "Transfer to confirm"},
		kindUnclassified:                        {Label: "Instrument to classify", Page: "/settings", Action: "Classify it in the settings"},
		"market-data-gap":                       {Label: "Market data stopped arriving", Page: "/settings#series-sources-title", Action: "Open the sources"},
		"bond-terms-or-index-levels-undeclared": {Label: "Bond terms, index levels or readings dropped", Page: pluginsPath, Action: openPlugins},
		kindSyncFailing:                         {Label: "Sync keeps failing", Page: pluginsPath, Action: openPlugins},
		kindCashCheckpoint:                      {Label: "Cash balance differs"},
		kindBondFeeAssumed:                      {Label: "Bond fee to confirm"},
		kindBondSeries:                          {Label: "Bond series still without its ISIN or terms"},
		"model-route-broken":                    {Label: "Model route names a missing provider instance"},
		kindTargetProposal:                      {Label: "Target weights proposed by the maintainer", Page: modelPath, Action: "Open the model portfolio"},
		kindMaintainerFailing:                   {Label: "Maintainer's weekly run failing", Page: modelPath, Action: "Open the model portfolio"},
		kindViewPluginUngranted:                 {Label: "View plugin without access to any account", Page: accessPath + viewPluginsAnchor, Action: "Open Access"},
		"cross-check-prices-undeclared":         {Label: "Undeclared cross-check prices dropped", Page: pluginsPath, Action: openPlugins},
	}
	if k, ok := kinds[kind]; ok {
		return k
	}
	return attentionKind{Label: kind}
}

// attentionItemKind names an item of kind about subject: as its kind does
// (attentionKindOf), but an IKE's missing earlier payments, whose subject
// starts with "earlier-payments/", are entered on the accounts page
// (D-149).
func attentionItemKind(kind, subject string) attentionKind {
	if kind == kindSettingMissing && strings.HasPrefix(subject, "earlier-payments/") {
		return attentionKind{Label: "Missing earlier payments into an IKE", Page: accountsPath, Action: openAccounts}
	}
	return attentionKindOf(kind)
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

// afterAttentionForm answers a posted attention form: back to the Inbox
// when the core took it, the Inbox with the core's reason when it refused
// it, and any other failure like any failed call to the core.
func (s *server) afterAttentionForm(w http.ResponseWriter, r *http.Request, sess session, name string, err error) {
	if err == nil {
		http.Redirect(w, r, inboxPath, http.StatusSeeOther)
		return
	}
	s.afterInboxForm(w, r, sess, name, err)
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
