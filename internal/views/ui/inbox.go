package ui

import (
	"fmt"
	"net/http"
	"net/url"
	"slices"
	"strings"

	"connectrpc.com/connect"

	ingestv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/ingest/v1"
	viewsv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/views/v1"
)

// inboxPath is the Inbox: every check that needs the owner, in one place
// (D-192, D-196).
const inboxPath = "/inbox"

// syncTimelineSize is how many of a failing sync's latest attempts its
// item shows.
const syncTimelineSize = 10

// labelDone names the done view and a done item's pill.
const labelDone = "Done"

// The Inbox's views (D-196), as the view parameter names them.
const (
	viewOpen         = "open"
	viewAcknowledged = "acknowledged"
	viewDone         = "done"
)

// removedCheckPages are the pages of checks D-192 removed: each one's
// path now redirects to the Inbox, which holds what it showed.
func removedCheckPages() []string {
	return []string{"/attention", "/tickers", "/assertions", "/price-checks", "/transfers"}
}

// inboxRoutes maps the Inbox, the owner's actions on its items, and the
// removed check pages, which redirect to it (D-192).
func (s *server) inboxRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET "+inboxPath, s.withSession(s.inboxPage))
	mux.HandleFunc("POST /inbox/acknowledge", s.withOwner(s.acknowledgeAttentionItem))
	mux.HandleFunc("POST /inbox/resolve", s.withOwner(s.resolveAttentionItem))
	mux.HandleFunc("POST /inbox/tickers/confirm", s.withOwner(s.confirmTicker))
	mux.HandleFunc("POST /inbox/price-checks/acknowledge", s.withOwner(s.acknowledgePriceCheck))
	mux.HandleFunc("POST /inbox/price-checks/threshold", s.withOwner(s.setPriceCheckThreshold))
	mux.HandleFunc("POST /inbox/transfers/confirm", s.withOwner(s.confirmTransfer))
	mux.HandleFunc("POST /inbox/transfers/reject", s.withOwner(s.rejectTransfer))
	mux.HandleFunc("POST /inbox/retry-sync", s.withOwner(s.retrySync))
	for _, p := range removedCheckPages() {
		mux.HandleFunc("GET "+p, toInbox)
	}
}

// toInbox sends the browser from a removed check page to the Inbox, for
// good (D-192).
func toInbox(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, inboxPath, http.StatusMovedPermanently)
}

// inboxLink is a link from an item to the page it is about.
type inboxLink struct {
	Href, Text string
}

// syncTimeline is a failing sync's evidence: its status and its latest
// attempts, newest first, and whether the owner may retry it.
type syncTimeline struct {
	Status   *viewsv1.SyncStatus
	Attempts []*viewsv1.SyncAttempt
	Retry    bool
}

// inboxItem is one attention item as the Inbox shows it: a title, its
// links, its pill, its evidence and the actions that settle it in place
// (D-196).
type inboxItem struct {
	Item  *viewsv1.AttentionItem
	Title string
	Links []inboxLink
	// Pill is the item's status pill.
	Pill statusPill
	// Acknowledge and Resolve offer the owner the AttentionService's
	// actions.
	Acknowledge, Resolve bool
	// The evidence and in-place form of the kinds that have them, nil for
	// the others.
	Ticker     *ingestv1.PendingTicker
	PriceCheck *instrumentChecks
	Sync       *syncTimeline
	Assertions []*viewsv1.AssertionCheck
	Transfer   *viewsv1.Transfer
	// Suggestion is a suggested transfer's pair (D-105), which the owner
	// confirms or marks not a transfer in place when Decide; a read
	// session sees it Waiting for the owner.
	Suggestion      *viewsv1.Transfer
	Decide, Waiting bool
	// Brokers are the unmapped broker accounts of an unmapped broker
	// account's instance, and Accounts the accounts to map them to, for
	// the owner's form; empty for a read session.
	Brokers  []*ingestv1.BrokerAccount
	Accounts []*viewsv1.Account
}

// inboxGroup is one of the Inbox's groups (D-196), with its items and,
// in the decisions group of the done view, the transfers decided.
type inboxGroup struct {
	ID, Title, Empty string
	Items            []inboxItem
	// Confirmed and Rejected are the transfers the owner decided (D-105),
	// in the done view; a suggested one is an attention item (D-130).
	Confirmed, Rejected []*viewsv1.Transfer
	// PriceChecks are every instrument's price checks with their history
	// (D-112), in the done view.
	PriceChecks []instrumentChecks
}

// Count is how many items the group lists.
func (g *inboxGroup) Count() int { return len(g.Items) }

// inboxFilter is one of the Inbox's views as its segmented filter shows it.
type inboxFilter struct {
	Key, Label string
	Count      int
	Current    bool
}

// inboxView is what the Inbox shows.
type inboxView struct {
	View    string
	Filters []inboxFilter
	Groups  []*inboxGroup
	Owner   bool
	// Notice is what the owner's last action did; Error why the core
	// refused it.
	Notice, Error string
	// Calm is an open view with nothing open: Kasztan asleep and one
	// line.
	Calm bool
	// Ask are the side column's Ask suggestions about the open items
	// (D-194), in the open view.
	Ask []string
}

// inboxAsk returns the Ask box's questions about items, the open ones
// (D-194): what needs the owner, why each failing sync fails and each
// account an item is tied to, at most maxSuggestions, each once.
func inboxAsk(items []*viewsv1.AttentionItem) []string {
	out := []string{"What needs me?"}
	add := func(q string) {
		if !slices.Contains(out, q) && len(out) < maxSuggestions {
			out = append(out, q)
		}
	}
	for _, it := range items {
		if it.GetKind() == kindSyncFailing && it.GetConnectorInstance() != "" {
			add("Why is " + it.GetConnectorInstance() + " failing?")
		}
	}
	for _, it := range items {
		for _, a := range it.GetAccountIds() {
			add("Show account " + a)
		}
	}
	return out
}

// inboxPage shows the Inbox in the view the request names, open by
// default (D-196).
func (s *server) inboxPage(w http.ResponseWriter, r *http.Request, sess session) {
	s.renderInbox(w, r, sess, http.StatusOK, inboxNotice(r.URL.Query()), "")
}

// inboxNotice is what the owner's last action did, read from the
// redirect's query.
func inboxNotice(q url.Values) string {
	switch {
	case q.Get("booked") != "":
		return "Confirmed: " + digitsOnly(q.Get("booked")) + " waiting row(s) booked."
	case q.Get("retried") != "":
		return "A sync of " + q.Get("retried") + " has started. Its attempt shows in the timeline once it ends."
	}
	return ""
}

// digitsOnly keeps the digits of s, so a crafted query writes nothing else
// into the page.
func digitsOnly(s string) string {
	return strings.Map(func(r rune) rune {
		if r >= '0' && r <= '9' {
			return r
		}
		return -1
	}, s)
}

// viewOf returns the Inbox view a request names and its read API value.
func viewOf(v string) (string, viewsv1.AttentionView) {
	switch v {
	case viewAcknowledged:
		return viewAcknowledged, viewsv1.AttentionView_ATTENTION_VIEW_ACKNOWLEDGED
	case viewDone:
		return viewDone, viewsv1.AttentionView_ATTENTION_VIEW_DONE
	}
	return viewOpen, viewsv1.AttentionView_ATTENTION_VIEW_OPEN
}

// renderInbox shows the Inbox with status, a notice of the owner's last
// action and the core's reason for refusing it, if any.
func (s *server) renderInbox(w http.ResponseWriter, r *http.Request, sess session, status int, notice, refusal string) {
	key, view := viewOf(r.URL.Query().Get("view"))
	resp, err := s.read.ListAttentionItems(r.Context(), bearer(&viewsv1.ListAttentionItemsRequest{View: view}, sess))
	if err != nil {
		s.coreFailed(w, r, "ListAttentionItems", err)
		return
	}
	ev, err := s.inboxEvidence(r, sess, key, resp.Msg.GetItems())
	if err != nil {
		s.coreFailed(w, r, "inbox evidence", err)
		return
	}
	data := inboxView{View: key, Owner: sess.Owner(), Notice: notice, Error: refusal, Groups: newInboxGroups()}
	for _, f := range []inboxFilter{
		{Key: viewOpen, Label: "Open", Count: int(resp.Msg.GetOpen())},
		{Key: viewAcknowledged, Label: "Acknowledged", Count: int(resp.Msg.GetAcknowledged())},
		{Key: viewDone, Label: labelDone, Count: int(resp.Msg.GetDone())},
	} {
		f.Current = f.Key == key
		data.Filters = append(data.Filters, f)
	}
	seenInstances := map[string]bool{}
	for _, it := range resp.Msg.GetItems() {
		row := ev.item(it, sess.Owner(), seenInstances)
		g := data.Groups[groupIndex(it.GetGroup())]
		g.Items = append(g.Items, row)
	}
	ev.extras(&data, key)
	if key == viewOpen {
		data.Ask = inboxAsk(resp.Msg.GetItems())
	}
	data.Calm = key == viewOpen && resp.Msg.GetOpen() == 0
	s.show(w, r, status, "inbox", page{Title: "Inbox", Session: &sess, Data: data})
}

// Indexes of the Inbox's groups, in the order it shows them.
const (
	groupDataComingIn = iota
	groupNumbersToCheck
	groupDecisions
)

// newInboxGroups returns the Inbox's three groups, empty (D-196).
func newInboxGroups() []*inboxGroup {
	return []*inboxGroup{
		{ID: "data-coming-in", Title: "Data coming in", Empty: "Every ticker is confirmed, every broker account mapped, and every sync and import went through."},
		{ID: "numbers-to-check", Title: "Numbers to check", Empty: "Every balance matches and no price differs beyond its threshold."},
		{ID: "decisions", Title: "Decisions", Empty: "No transfers to confirm and no settings missing."},
	}
}

// groupIndex returns the index of the group items of g are shown in; an
// item without a group is a decision, as a kind the core doesn't know is.
func groupIndex(g viewsv1.AttentionItemGroup) int {
	switch g {
	case viewsv1.AttentionItemGroup_ATTENTION_ITEM_GROUP_DATA_COMING_IN:
		return groupDataComingIn
	case viewsv1.AttentionItemGroup_ATTENTION_ITEM_GROUP_NUMBERS_TO_CHECK:
		return groupNumbersToCheck
	case viewsv1.AttentionItemGroup_ATTENTION_ITEM_GROUP_DECISIONS, viewsv1.AttentionItemGroup_ATTENTION_ITEM_GROUP_UNSPECIFIED:
	}
	return groupDecisions
}

// inboxEvidence is what the Inbox read besides the items, to show their
// evidence and settle them in place.
type inboxEvidence struct {
	tickers   []*ingestv1.PendingTicker
	checks    []instrumentChecks
	syncs     map[string]*syncTimeline
	asserts   map[string][]*viewsv1.AssertionCheck
	transfers *viewsv1.ListTransfersResponse
	brokers   []*ingestv1.BrokerAccount
	accounts  []*viewsv1.Account
}

// inboxEvidence reads what the items' evidence and forms need: only what
// some item, or the view's extra lists, asks for.
func (s *server) inboxEvidence(r *http.Request, sess session, view string, items []*viewsv1.AttentionItem) (*inboxEvidence, error) {
	ev := &inboxEvidence{syncs: map[string]*syncTimeline{}, asserts: map[string][]*viewsv1.AssertionCheck{}}
	for _, st := range s.evidenceSteps(r, sess, view, items, ev) {
		if !st.needed {
			continue
		}
		if err := st.read(); err != nil {
			return nil, err
		}
	}
	for _, it := range items {
		if err := s.itemEvidence(r, sess, it, ev); err != nil {
			return nil, err
		}
	}
	return ev, nil
}

// evidenceStep is one read of the Inbox's evidence, and whether the
// items or the view need it.
type evidenceStep struct {
	needed bool
	read   func() error
}

// evidenceSteps returns the reads the Inbox may need for items in view,
// into ev: pending tickers and unmapped broker accounts for the owner,
// price checks for a flagged one or the done view's list, and transfers
// for a broken or suggested one or the done view's decided transfers.
func (s *server) evidenceSteps(r *http.Request, sess session, view string, items []*viewsv1.AttentionItem, ev *inboxEvidence) []evidenceStep {
	kinds := map[string]bool{}
	for _, it := range items {
		kinds[it.GetKind()] = true
	}
	owner := sess.Owner()
	return []evidenceStep{
		{kinds[kindTicker] && owner, func() error { return s.tickerEvidence(r, sess, ev) }},
		{kinds[kindPriceCheck] || view == viewDone, func() error { return s.priceCheckEvidence(r, sess, ev) }},
		{kinds[kindTransferBroken] || kinds[kindTransferSuggested] || view == viewDone, func() error { return s.transferEvidence(r, sess, ev) }},
		{kinds[kindBrokerUnmapped] && owner, func() error { return s.unmappedEvidence(r, sess, ev) }},
	}
}

// tickerEvidence reads the tickers waiting for the owner (D-116).
func (s *server) tickerEvidence(r *http.Request, sess session, ev *inboxEvidence) error {
	tickers, err := s.pendingTickers(r, sess)
	ev.tickers = tickers
	return err
}

// priceCheckEvidence reads every instrument's price checks (D-112).
func (s *server) priceCheckEvidence(r *http.Request, sess session, ev *inboxEvidence) error {
	resp, err := s.read.ListPriceChecks(r.Context(), bearer(&viewsv1.ListPriceChecksRequest{}, sess))
	if err != nil {
		return fmt.Errorf("ListPriceChecks: %w", err)
	}
	ev.checks = latestPerInstrument(resp.Msg.GetChecks(), resp.Msg.GetThreshold(), sess.Owner())
	return nil
}

// transferEvidence reads the transfers between the owner's accounts
// (D-105, D-131).
func (s *server) transferEvidence(r *http.Request, sess session, ev *inboxEvidence) error {
	resp, err := s.read.ListTransfers(r.Context(), bearer(&viewsv1.ListTransfersRequest{}, sess))
	if err != nil {
		return fmt.Errorf("ListTransfers: %w", err)
	}
	ev.transfers = resp.Msg
	return nil
}

// itemEvidence reads a failing sync's timeline and a flagged balance
// assertion's checks, once per instance or account.
func (s *server) itemEvidence(r *http.Request, sess session, it *viewsv1.AttentionItem, ev *inboxEvidence) error {
	switch it.GetKind() {
	case kindSyncFailing:
		return s.syncEvidence(r, sess, it.GetConnectorInstance(), ev)
	case kindAssertion:
		return s.assertionEvidence(r, sess, it.GetAccountId(), ev)
	}
	return nil
}

// syncEvidence reads the timeline of instance's syncs (D-196); an instance
// never synced has none.
func (s *server) syncEvidence(r *http.Request, sess session, instance string, ev *inboxEvidence) error {
	if _, ok := ev.syncs[instance]; ok || instance == "" {
		return nil
	}
	resp, err := s.read.ListSyncAttempts(r.Context(), bearer(&viewsv1.ListSyncAttemptsRequest{
		ConnectorInstance: instance, PageSize: syncTimelineSize,
	}, sess))
	if connect.CodeOf(err) == connect.CodeNotFound {
		return nil
	}
	if err != nil {
		return fmt.Errorf("ListSyncAttempts: %w", err)
	}
	ev.syncs[instance] = &syncTimeline{Status: resp.Msg.GetStatus(), Attempts: resp.Msg.GetAttempts()}
	return nil
}

// assertionEvidence reads account's flagged balance assertion checks.
func (s *server) assertionEvidence(r *http.Request, sess session, account string, ev *inboxEvidence) error {
	if _, ok := ev.asserts[account]; ok || account == "" {
		return nil
	}
	resp, err := s.read.ListAssertionChecks(r.Context(), bearer(&viewsv1.ListAssertionChecksRequest{
		AccountId: account, Status: viewsv1.AssertionStatus_ASSERTION_STATUS_FLAGGED,
	}, sess))
	if err != nil {
		return fmt.Errorf("ListAssertionChecks: %w", err)
	}
	ev.asserts[account] = resp.Msg.GetChecks()
	return nil
}

// pendingTickers lists the tickers waiting for the owner (D-116).
func (s *server) pendingTickers(r *http.Request, sess session) ([]*ingestv1.PendingTicker, error) {
	resp, err := s.tickers.ListPendingTickers(r.Context(), connect.NewRequest(&ingestv1.ListPendingTickersRequest{
		Origin: sess.Origin, SessionToken: sess.Token,
	}))
	if err != nil {
		return nil, fmt.Errorf("ListPendingTickers: %w", err)
	}
	return resp.Msg.GetTickers(), nil
}

// unmappedEvidence reads the owner's unmapped broker accounts and the
// accounts to map them to (D-72, D-121).
func (s *server) unmappedEvidence(r *http.Request, sess session, ev *inboxEvidence) error {
	brokers, err := s.brokerAccounts(r, sess)
	if err != nil {
		return err
	}
	for _, b := range brokers {
		if b.GetAccountId() == "" {
			ev.brokers = append(ev.brokers, b)
		}
	}
	resp, err := s.read.ListAccounts(r.Context(), bearer(&viewsv1.ListAccountsRequest{}, sess))
	if err != nil {
		return fmt.Errorf("ListAccounts: %w", err)
	}
	ev.accounts = resp.Msg.GetAccounts()
	return nil
}

// item returns it as the Inbox shows it to a session that is the owner's
// or not; seen holds the instances whose unmapped broker accounts an item
// already offers to map.
func (ev *inboxEvidence) item(it *viewsv1.AttentionItem, owner bool, seen map[string]bool) inboxItem {
	row := inboxItem{Item: it, Title: itemTitle(it), Links: itemLinks(it)}
	row.Pill = itemPill(it)
	resolved := it.GetState() == viewsv1.AttentionItemState_ATTENTION_ITEM_STATE_RESOLVED
	row.Acknowledge = owner && it.GetState() == viewsv1.AttentionItemState_ATTENTION_ITEM_STATE_OPEN
	row.Resolve = owner && !resolved && it.GetOwnerResolves()
	ev.evidence(&row, owner && !resolved, seen)
	row.Waiting = !owner && !resolved && row.Suggestion != nil
	return row
}

// evidence adds to row its kind's evidence and the form that settles it in
// place; act offers the owner's forms.
func (ev *inboxEvidence) evidence(row *inboxItem, act bool, seen map[string]bool) {
	it := row.Item
	switch it.GetKind() {
	case kindTicker:
		row.Ticker = ev.ticker(it.GetSubject())
	case kindPriceCheck:
		row.PriceCheck = ev.priceChecks(it.GetSubject())
		// The price check's own acknowledgement settles it (D-112).
		row.Acknowledge = row.Acknowledge && row.PriceCheck == nil
	case kindSyncFailing:
		row.Sync = ev.syncOf(it.GetConnectorInstance(), act)
	case kindAssertion:
		row.Assertions = assertionsOn(ev.asserts[it.GetAccountId()], it.GetSubject())
	case kindTransferBroken:
		row.Transfer = ev.brokenTransfer(it.GetSubject())
	case kindTransferSuggested:
		row.Suggestion = pairNamed(ev.transfers.GetSuggested(), it.GetSubject())
		row.Decide = act && row.Suggestion != nil
	case kindBrokerUnmapped:
		ev.mapForm(row, act, seen)
	}
}

// syncOf returns the timeline of instance's syncs, with the owner's "Retry
// now" when act, or nil when it has none.
func (ev *inboxEvidence) syncOf(instance string, act bool) *syncTimeline {
	t, ok := ev.syncs[instance]
	if !ok {
		return nil
	}
	c := *t
	c.Retry = act
	return &c
}

// mapForm adds to an unmapped broker account's row the form that maps its
// instance's unmapped broker accounts, when act, once per instance: seen
// holds the instances an earlier row offers it for.
func (ev *inboxEvidence) mapForm(row *inboxItem, act bool, seen map[string]bool) {
	instance := row.Item.GetConnectorInstance()
	if !act || seen[instance] {
		return
	}
	seen[instance] = true
	row.Brokers, row.Accounts = brokersOf(ev.brokers, instance), ev.accounts
}

// extras adds what the done view lists besides items: the decided
// transfers and every price check with its history, so nothing the
// removed pages showed is lost (D-192).
func (ev *inboxEvidence) extras(data *inboxView, view string) {
	if view != viewDone {
		return
	}
	decisions := data.Groups[groupDecisions]
	decisions.Confirmed, decisions.Rejected = ev.transfers.GetConfirmed(), ev.transfers.GetRejected()
	data.Groups[groupNumbersToCheck].PriceChecks = ev.checks
}

// ticker returns the pending ticker an item's subject, "<instance>/<symbol>",
// names, or nil.
func (ev *inboxEvidence) ticker(subject string) *ingestv1.PendingTicker {
	for _, t := range ev.tickers {
		if t.GetInstance()+"/"+t.GetSymbol() == subject {
			return t
		}
	}
	return nil
}

// priceChecks returns the price checks of the instrument a price check
// item's subject, "<ISIN>/<symbol>", names, or nil.
func (ev *inboxEvidence) priceChecks(subject string) *instrumentChecks {
	isin, symbol, _ := strings.Cut(subject, "/")
	for i := range ev.checks {
		in := ev.checks[i].Instrument
		if in.GetIsin() == isin && (symbol == "" || in.GetSymbol() == symbol) {
			return &ev.checks[i]
		}
	}
	return nil
}

// brokenTransfer returns the broken pair a transfer-broken item's subject,
// "<withdrawal> -> <deposit>", names, or nil.
func (ev *inboxEvidence) brokenTransfer(subject string) *viewsv1.Transfer {
	return pairNamed(ev.transfers.GetBroken(), subject)
}

// pairNamed returns the transfer among transfers whose pair a transfer
// item's subject, "<withdrawal> -> <deposit>", names, or nil.
func pairNamed(transfers []*viewsv1.Transfer, subject string) *viewsv1.Transfer {
	for _, t := range transfers {
		if t.GetWithdrawalSource()+"/"+t.GetWithdrawalExternalId()+" -> "+t.GetDepositSource()+"/"+t.GetDepositExternalId() == subject {
			return t
		}
	}
	return nil
}

// subjectDateParts is how many parts a balance assertion item's subject
// has up to its date: the account and the date.
const subjectDateParts = 2

// assertionsOn returns the flagged checks among checks on the date a
// balance assertion item's subject, "<account>/<date>/...", names.
func assertionsOn(checks []*viewsv1.AssertionCheck, subject string) []*viewsv1.AssertionCheck {
	parts := strings.SplitN(subject, "/", subjectDateParts+1)
	if len(parts) < subjectDateParts {
		return nil
	}
	var out []*viewsv1.AssertionCheck
	for _, c := range checks {
		if c.GetDate() == parts[1] {
			out = append(out, c)
		}
	}
	return out
}

// brokersOf returns those of brokers instance reported.
func brokersOf(brokers []*ingestv1.BrokerAccount, instance string) []*ingestv1.BrokerAccount {
	var out []*ingestv1.BrokerAccount
	for _, b := range brokers {
		if b.GetConnectorInstance() == instance {
			out = append(out, b)
		}
	}
	return out
}

// fixedTitles are the titles of the kinds whose title names nothing of
// their item.
func fixedTitles() map[string]string {
	return map[string]string{
		kindImportProblems:    "An import left rows out",
		kindTransferBroken:    "A confirmed transfer no longer matches",
		kindTransferSuggested: "Confirm a transfer between your accounts",
	}
}

// itemTitle says in a few words what an item asks of the owner.
func itemTitle(it *viewsv1.AttentionItem) string {
	if title, ok := fixedTitles()[it.GetKind()]; ok {
		return title
	}
	subject := it.GetSubject()
	switch it.GetKind() {
	case kindTicker:
		_, symbol, _ := strings.Cut(subject, "/")
		return "Confirm the ISIN of " + symbol
	case kindBrokerUnmapped:
		return "Map a broker account " + it.GetConnectorInstance() + " reported"
	case kindSyncFailing:
		return "The " + it.GetConnectorInstance() + " sync keeps failing"
	case kindPriceCheck:
		return "The price of " + subject + " differs"
	case kindAssertion:
		return "A balance in " + it.GetAccountId() + " doesn't match"
	case kindSettingMissing:
		return "Set " + subject
	case kindUnclassified:
		return "Classify " + subject + " for the yearly tax"
	}
	return attentionKindOf(it.GetKind()).Label
}

// itemLinks returns the pages an item links to: its accounts, and the page
// where a kind not settled in place is settled.
func itemLinks(it *viewsv1.AttentionItem) []inboxLink {
	var out []inboxLink
	accounts := it.GetAccountIds()
	if len(accounts) == 0 && it.GetAccountId() != "" {
		accounts = []string{it.GetAccountId()}
	}
	for _, a := range accounts {
		out = append(out, inboxLink{Href: "/accounts/" + url.PathEscape(a), Text: "Open account " + a})
	}
	if page := attentionKindOf(it.GetKind()).Page; page != "" {
		out = append(out, inboxLink{Href: page, Text: attentionKindOf(it.GetKind()).Action})
	}
	return out
}

// statusPill is a status pill: its word and its class.
type statusPill struct {
	Word, Class string
}

// itemPill returns an item's status pill.
func itemPill(it *viewsv1.AttentionItem) statusPill {
	switch it.GetState() {
	case viewsv1.AttentionItemState_ATTENTION_ITEM_STATE_ACKNOWLEDGED:
		return statusPill{"Acknowledged", "pill-neutral"}
	case viewsv1.AttentionItemState_ATTENTION_ITEM_STATE_RESOLVED:
		return statusPill{labelDone, "pill-ok"}
	case viewsv1.AttentionItemState_ATTENTION_ITEM_STATE_OPEN, viewsv1.AttentionItemState_ATTENTION_ITEM_STATE_UNSPECIFIED:
	}
	if it.GetKind() == kindSyncFailing {
		return statusPill{"Failing", "pill-danger"}
	}
	return statusPill{"Needs you", "pill-warn"}
}

// retrySync hands the owner's "Retry now" on a failing sync to the core
// (RetrySync, D-196), then back to the Inbox saying the sync started, or
// with the core's reason for refusing it, such as a sync that runs now.
func (s *server) retrySync(w http.ResponseWriter, r *http.Request, sess session) {
	instance := strings.TrimSpace(r.PostFormValue("instance"))
	if instance == "" || len(instance) > maxInstanceChars {
		s.errorPage(w, r, http.StatusBadRequest, "The form is malformed.")
		return
	}
	_, err := s.syncs.RetrySync(r.Context(), connect.NewRequest(&ingestv1.RetrySyncRequest{
		Origin: sess.Origin, SessionToken: sess.Token, Instance: instance,
	}))
	if err == nil {
		http.Redirect(w, r, inboxPath+"?"+url.Values{"retried": {instance}}.Encode(), http.StatusSeeOther)
		return
	}
	s.afterInboxForm(w, r, sess, "RetrySync", err)
}

// maxInstanceChars caps a posted connector instance name, as the core's
// names are (56 characters).
const maxInstanceChars = 56

// afterInboxForm answers a posted Inbox form the core didn't take: the
// Inbox with the core's reason when it refused it, and any other failure
// like any failed call to the core.
func (s *server) afterInboxForm(w http.ResponseWriter, r *http.Request, sess session, name string, err error) {
	code := connect.CodeOf(err)
	if status, refused := refusalStatus(code); refused && code != connect.CodePermissionDenied {
		s.renderInbox(w, r, sess, status, "", "The core refused this: "+connectMessage(err))
		return
	}
	if code == connect.CodeUnavailable || code == connect.CodeFailedPrecondition {
		s.renderInbox(w, r, sess, http.StatusServiceUnavailable, "", "The core refused this: "+connectMessage(err))
		return
	}
	s.coreFailed(w, r, name, err)
}

// accountItems lists the open attention items tied to an account, for the
// banner on its page (D-196); a core that can't list them shows none, and
// the page still renders.
func (s *server) accountItems(r *http.Request, sess session, account string) []inboxItem {
	resp, err := s.read.ListAttentionItems(r.Context(), bearer(&viewsv1.ListAttentionItemsRequest{
		View: viewsv1.AttentionView_ATTENTION_VIEW_OPEN, AccountId: account,
	}, sess))
	if err != nil {
		s.log.WarnContext(r.Context(), "listing an account's attention items", "code", connect.CodeOf(err).String())
		return nil
	}
	out := make([]inboxItem, 0, len(resp.Msg.GetItems()))
	for _, it := range resp.Msg.GetItems() {
		out = append(out, inboxItem{Item: it, Title: itemTitle(it)})
	}
	return out
}

// backToInbox returns where a posted Inbox form goes once the core took
// it: the Inbox in the view the form names, open by default.
func backToInbox(r *http.Request) string {
	switch r.PostFormValue("view") {
	case viewAcknowledged:
		return inboxPath + "?view=" + viewAcknowledged
	case viewDone:
		return inboxPath + "?view=" + viewDone
	}
	return inboxPath
}
