package ui

import (
	"net/http"
	"strconv"
	"strings"

	"connectrpc.com/connect"

	"github.com/arkadiuss/openfief/internal/views/app/answer"
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
	kindAssertion         = answer.KindAssertion
	kindPriceCheck        = answer.KindPriceCheck
	kindImportProblems    = answer.KindImportProblems
	kindBrokerUnmapped    = answer.KindBrokerUnmapped
	kindTicker            = answer.KindTicker
	kindSettingMissing    = answer.KindSettingMissing
	kindTransferBroken    = answer.KindTransferBroken
	kindTransferSuggested = answer.KindTransferSuggested
	kindUnclassified      = answer.KindUnclassified
	kindSyncFailing       = answer.KindSyncFailing
	kindCashCheckpoint    = answer.KindCashCheckpoint
	kindBondFeeAssumed    = answer.KindBondFeeAssumed
	kindBondSeries        = answer.KindBondSeries
	// kindViewPluginUngranted is a view plugin that can't read anything until
	// the owner grants it accounts (D-196, D-209).
	kindViewPluginUngranted = answer.KindViewPluginUngranted
	// kindAuditLogBehind is the audit entries waiting in the outbox through
	// several failed attempts to enter the log (D-196, D-211).
	kindAuditLogBehind = answer.KindAuditLogBehind
)

// openAccounts is the words of an item's link to the accounts page.
const openAccounts = "Open the accounts"

// attentionKindOf names kind, one of the core's kinds of attention items
// (D-130, D-131, D-132, D-94, D-141, D-179, D-196): its label is the words the
// Ask box names it by too (answer.KindLabel, which for a kind it doesn't know
// is its own name), and the page where it is settled, none for a kind the Inbox
// settles in place: its item links its account instead.
func attentionKindOf(kind string) attentionKind {
	pages := map[string]attentionKind{
		kindImportProblems:                      {Page: accountsPath, Action: openAccounts},
		kindBrokerUnmapped:                      {Page: accountsPath, Action: openAccounts},
		kindSettingMissing:                      {Page: "/settings", Action: "Open the settings"},
		kindUnclassified:                        {Page: "/settings", Action: "Classify it in the settings"},
		"market-data-gap":                       {Page: "/settings#series-sources-title", Action: "Open the sources"},
		"bond-terms-or-index-levels-undeclared": {Page: pluginsPath, Action: openPlugins},
		kindSyncFailing:                         {Page: pluginsPath, Action: openPlugins},
		"model-route-broken":                    {Page: modelsPath + "#routes-title", Action: "Fix the route"},
		kindTargetProposal:                      {Page: modelPath, Action: "Open the model portfolio"},
		kindMaintainerFailing:                   {Page: modelPath, Action: "Open the model portfolio"},
		kindViewPluginUngranted:                 {Page: accessPath + viewPluginsAnchor, Action: "Open Access"},
		"cross-check-prices-undeclared":         {Page: pluginsPath, Action: openPlugins},
		kindAuditLogBehind:                      {Page: auditPath, Action: "Open the audit log"},
	}
	k := pages[kind]
	k.Label = answer.KindLabel(kind)
	return k
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
