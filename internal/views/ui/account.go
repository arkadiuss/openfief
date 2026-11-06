package ui

import (
	"cmp"
	"fmt"
	"net/http"
	"net/url"
	"slices"

	ingestv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/ingest/v1"
	viewsv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/views/v1"
)

// transactionsPageSize is how many transactions one page shows.
const transactionsPageSize = 50

// holdingRow is a holding with its stored valuation on the date, if any.
type holdingRow struct {
	Holding   *viewsv1.Holding
	Valuation *viewsv1.Valuation
}

// transactionsView is one page of an account's or an instrument's
// transactions and the link to the next one.
type transactionsView struct {
	// Base is the page they are on: an account's or an instrument's.
	Base         string
	Transactions []*viewsv1.Transaction
	Next         string
	// ShowAccount adds a column naming each transaction's account, by its
	// name in Accounts, on an instrument's page.
	ShowAccount bool
	Accounts    map[string]string
	// Tab is the tab of the page the transactions are on, if any.
	Tab string
}

// NextLink is the URL of the page that starts at the next page, for
// browsers without htmx.
func (t *transactionsView) NextLink() string {
	q := url.Values{"page_token": {t.Next}}
	if t.Tab != "" {
		q.Set(tabParam, t.Tab)
	}
	return t.Base + "?" + q.Encode()
}

// NextFragment is the URL htmx loads the next page's rows from.
func (t *transactionsView) NextFragment() string {
	return t.Base + "/transactions?" + url.Values{"page_token": {t.Next}}.Encode()
}

// AccountName is the name of the account with ID id, or the ID.
func (t *transactionsView) AccountName(id string) string { return cmp.Or(t.Accounts[id], id) }

// Columns is how many columns a row has.
func (t *transactionsView) Columns() int {
	if t.ShowAccount {
		return transactionColumns + 1
	}
	return transactionColumns
}

// transactionColumns is how many columns an account's transactions have.
const transactionColumns = 5

// accountTransactions is the transactions view of the account with ID id.
func accountTransactions(id string) transactionsView {
	return transactionsView{Base: "/accounts/" + url.PathEscape(id), Tab: tabTransactions}
}

// accountData is what an account's page shows: its header, banner and
// figure cards, then one tab (accountTabs).
type accountData struct {
	Date    string
	Account *viewsv1.Account
	Status  *viewsv1.ValuationStatus
	// Sleeve is the account's sleeve's name; Sync the pill of its
	// connector instances' syncs, nil without a bound instance; Instances
	// the connector instances bound to it.
	Sleeve    string
	Sync      *statusPill
	Instances []string
	// Total is the account's gross, net and liquidation values with cash
	// on Date (D-77, D-106); nil when nothing is valued.
	Total *viewsv1.DateTotal
	// Tab is the tab shown and Tabs every tab; StartParam is the charts'
	// start the owner chose, "" for the default.
	Tab        string
	Tabs       []accountTab
	StartParam string
	Holdings   []holdingRow
	Lots       []*viewsv1.Lot
	Cash       []*viewsv1.Money
	Checks     []*viewsv1.AssertionCheck
	// Checkpoints are the cash checkpoints the owner entered, with their
	// differences (D-158), on the Cash tab.
	Checkpoints []checkpointRow
	// CheckpointForm is the owner's "Enter cash balance" form on the Cash
	// tab; nil for a read session and on the other tabs.
	CheckpointForm *checkpointForm
	// EnterCash is the owner's "Enter cash balance" link in the header,
	// to the Cash tab's form (D-158, D-195); empty for a read session.
	EnterCash string
	// Transactions are a page of the account's transactions.
	Transactions transactionsView
	// Charts are the account's value over time and performance from the
	// charts' start (D-138).
	Charts *accountCharts
	// BondLot is the "Add bond lot" form, shown to the owner on a bond
	// account's page only (D-9); nil elsewhere.
	BondLot *bondLotForm
	// UploadForm is the owner's "Upload statement" form, shown only on an
	// account whose bound connector takes statement files (D-109); nil
	// elsewhere.
	UploadForm *uploadForm
	// Upload is the import the owner's last upload became, to show its
	// counts and ingest problems; nil otherwise.
	Upload *ingestv1.UploadStatementResponse
	// Error is why the core refused what the owner posted.
	Error string
	// Items are the account's own open attention items, for the banner at
	// the top of its page (D-196).
	Items []inboxItem
}

// TabLink is the account's page showing tab, for browsers without htmx.
func (d *accountData) TabLink(tab string) string {
	return "/accounts/" + url.PathEscape(d.Account.GetId()) + "?" + d.tabQuery(tab)
}

// TabFragment is the URL htmx swaps the tabs showing tab in from.
func (d *accountData) TabFragment(tab string) string {
	return "/accounts/" + url.PathEscape(d.Account.GetId()) + "/tab?" + d.tabQuery(tab)
}

// Return is the account's time-weighted return over its charts, from their
// first day to the last, in percent; empty without valued days.
func (d *accountData) Return() string {
	if d.Charts == nil {
		return ""
	}
	return d.Charts.Performance.Return
}

// Crumbs is the account page's breadcrumb.
func (d *accountData) Crumbs() []crumb {
	return []crumb{overviewCrumb(), accountsCrumb(), {Text: d.Account.GetName()}}
}

// HoldingLink is the page of a holding's instrument, reached from this
// account (D-195).
func (d *accountData) HoldingLink(in *viewsv1.Instrument) string {
	return instrumentFrom(in.GetIsin(), d.Account.GetId())
}

// tabQuery is the account page's query that shows tab on the same date
// and with the same charts' start.
func (d *accountData) tabQuery(tab string) string {
	q := url.Values{tabParam: {tab}, dateParam: {d.Date}}
	if d.StartParam != "" {
		q.Set(startParam, d.StartParam)
	}
	return q.Encode()
}

// accountForm is what the owner posted to an account's page and why it was
// refused, to show again; the zero value is a fresh page.
type accountForm struct {
	Lot        *bondLotForm
	Checkpoint *checkpointForm
	Upload     *ingestv1.UploadStatementResponse
	Error      string
	// Tab is the tab the form was on, shown again; empty is the request's.
	Tab string
}

// instrumentKey identifies an instrument by ISIN and venue symbol.
func instrumentKey(i *viewsv1.Instrument) string { return i.GetIsin() + "|" + i.GetSymbol() }

// joinValuations pairs each holding with the stored valuation of the same
// instrument; a holding without one keeps a nil valuation.
func joinValuations(holdings []*viewsv1.Holding, valuations []*viewsv1.Valuation) []holdingRow {
	byInstrument := map[string]*viewsv1.Valuation{}
	for _, v := range valuations {
		byInstrument[instrumentKey(v.GetInstrument())] = v
	}
	rows := make([]holdingRow, 0, len(holdings))
	for _, h := range holdings {
		rows = append(rows, holdingRow{Holding: h, Valuation: byInstrument[instrumentKey(h.GetInstrument())]})
	}
	return rows
}

// accountPage shows one account on a date: its header with its sleeve,
// wrapper and sync, its open items, its figure cards and the tab the
// request names; to the owner, it also offers the "Upload statement"
// form, the "Enter cash balance" form on the Cash tab (D-158), and a bond
// account's page the "Add bond lot" form.
func (s *server) accountPage(w http.ResponseWriter, r *http.Request, sess session) {
	s.renderAccount(w, r, sess, http.StatusOK, accountForm{})
}

// accountTabFragment answers the account's tabs: htmx swaps in the tabs
// showing another one; a browser without htmx gets the whole page.
func (s *server) accountTabFragment(w http.ResponseWriter, r *http.Request, sess session) {
	if !isHTMX(r) {
		s.renderAccount(w, r, sess, http.StatusOK, accountForm{})
		return
	}
	data, err := s.accountView(r, sess, accountForm{})
	if err != nil {
		s.readFailed(w, r, err)
		return
	}
	s.showTemplate(w, r, http.StatusOK, "account", "account-tabs", data)
}

// renderAccount shows the account page with status, and with form: the
// bond lot the owner posted and the core's reason for refusing it.
func (s *server) renderAccount(w http.ResponseWriter, r *http.Request, sess session, code int, form accountForm) {
	data, err := s.accountView(r, sess, form)
	if err != nil {
		s.readFailed(w, r, err)
		return
	}
	s.show(w, r, code, "account", page{Title: data.Account.GetName(), Session: &sess, Data: data})
}

// accountView reads what the account page shows for the request, with
// form as posted; it fails with a *readError.
func (s *server) accountView(r *http.Request, sess session, form accountForm) (*accountData, error) {
	ctx, id := r.Context(), r.PathValue("id")
	acc, err := s.read.GetAccount(ctx, bearer(&viewsv1.GetAccountRequest{AccountId: id}, sess))
	if err != nil {
		return nil, &readError{Method: "GetAccount", Err: err}
	}
	valuation, err := s.read.GetValuationStatus(ctx, bearer(&viewsv1.GetValuationStatusRequest{AccountId: id}, sess))
	if err != nil {
		return nil, &readError{Method: "GetValuationStatus", Err: err}
	}
	q := r.URL.Query()
	date, err := asOfDate(q.Get(dateParam), []*viewsv1.ValuationStatus{valuation.Msg.GetStatus()}, s.today())
	if err != nil {
		return nil, &readError{Status: http.StatusBadRequest, Err: err}
	}
	start, err := parseStart(q.Get(startParam))
	if err != nil {
		return nil, &readError{Status: http.StatusBadRequest, Err: err}
	}
	data := &accountData{
		Date: date.String(), Account: acc.Msg.GetAccount(), Status: valuation.Msg.GetStatus(), Error: form.Error,
		Upload: form.Upload, Items: s.accountItems(r, sess, id), Tab: parseAccountTab(cmp.Or(form.Tab, q.Get(tabParam))), Tabs: accountTabs(),
		StartParam: optionalString(start), Transactions: accountTransactions(id),
	}
	if data.Charts, err = s.accountCharts(r, sess, id, start, date); err != nil {
		return nil, &readError{Method: "account charts", Err: err}
	}
	data.Total = data.Charts.Value.Last
	if method, headErr := s.accountHead(r, sess, data); headErr != nil {
		return nil, &readError{Method: method, Err: headErr}
	}
	if formErr := s.ownerForms(r, sess, data, form.Lot); formErr != nil {
		return nil, &readError{Method: "ListStatementAccounts", Err: formErr}
	}
	if method, tabErr := s.accountTab(r, sess, data); tabErr != nil {
		return nil, &readError{Method: method, Err: tabErr}
	}
	s.checkpointForms(sess, data, form.Checkpoint)
	return data, nil
}

// accountHead reads the account header's sleeve name, connector instances
// and sync pill into data. It returns the method that failed with its
// error.
func (s *server) accountHead(r *http.Request, sess session, data *accountData) (string, error) {
	ctx := r.Context()
	sleeves, err := s.read.ListSleeves(ctx, bearer(&viewsv1.ListSleevesRequest{}, sess))
	if err != nil {
		return "ListSleeves", fmt.Errorf("ListSleeves: %w", err)
	}
	data.Sleeve = data.Account.GetSleeve()
	for _, sl := range sleeves.Msg.GetSleeves() {
		if sl.GetId() == data.Account.GetSleeve() {
			data.Sleeve = sl.GetName()
		}
	}
	bindings, err := s.read.ListConnectorBindings(ctx, bearer(&viewsv1.ListConnectorBindingsRequest{}, sess))
	if err != nil {
		return "ListConnectorBindings", fmt.Errorf("ListConnectorBindings: %w", err)
	}
	for _, b := range bindings.Msg.GetBindings() {
		if b.GetAccountId() == data.Account.GetId() {
			data.Instances = append(data.Instances, b.GetConnectorInstance())
		}
	}
	if len(data.Instances) == 0 {
		return "", nil
	}
	syncs, err := s.read.ListSyncStatuses(ctx, bearer(&viewsv1.ListSyncStatusesRequest{}, sess))
	if err != nil {
		return "ListSyncStatuses", fmt.Errorf("ListSyncStatuses: %w", err)
	}
	data.Sync = syncPill(data.Instances, syncs.Msg.GetStatuses())
	return "", nil
}

// syncPill is the account header's pill of the syncs of instances, the
// connector instances bound to it, by statuses: Sync failing when one
// keeps failing, else Synced with the latest success's day, else none.
func syncPill(instances []string, statuses []*viewsv1.SyncStatus) *statusPill {
	latest := ""
	for _, st := range statuses {
		if !slices.Contains(instances, st.GetConnectorInstance()) {
			continue
		}
		if st.GetKeepsFailing() {
			return &statusPill{"Sync failing · " + st.GetConnectorInstance(), pillDanger}
		}
		latest = max(latest, st.GetLastSuccessAt())
	}
	if latest == "" {
		return nil
	}
	return &statusPill{"Synced " + latest[:min(len(latest), len("2006-01-02"))], pillOK}
}

// ownerForms adds, for the owner, the "Upload statement" form when the
// account's connector takes files (D-109) and a bond account's "Add bond
// lot" form, with lot as posted, to data.
func (s *server) ownerForms(r *http.Request, sess session, data *accountData, lot *bondLotForm) error {
	if !sess.Owner() {
		return nil
	}
	upload, ok, err := s.uploadForm(r, sess, data.Account.GetId())
	if err != nil {
		return err
	}
	if ok {
		data.UploadForm = upload
	}
	if data.Account.GetInstitution() == bondInstitution {
		data.BondLot = cmp.Or(lot, &bondLotForm{})
	}
	return nil
}

// accountTab reads what data's tab shows into data. It returns the method
// that failed with its error.
func (s *server) accountTab(r *http.Request, sess session, data *accountData) (string, error) {
	ctx, id, date := r.Context(), data.Account.GetId(), data.Date
	switch data.Tab {
	case tabHoldings:
		return s.holdings(r, sess, data)
	case tabTransactions:
		page, err := s.transactions(r, sess, id, r.URL.Query().Get("page_token"))
		data.Transactions = page
		return "ListTransactions", err
	case tabLots:
		lots, err := s.read.ListLots(ctx, bearer(&viewsv1.ListLotsRequest{AccountId: id, Date: date}, sess))
		if err != nil {
			return "ListLots", fmt.Errorf("ListLots: %w", err)
		}
		data.Lots = lots.Msg.GetLots()
	case tabCash:
		cash, err := s.read.ListCash(ctx, bearer(&viewsv1.ListCashRequest{AccountId: id, Date: date}, sess))
		if err != nil {
			return "ListCash", fmt.Errorf("ListCash: %w", err)
		}
		data.Cash = cash.Msg.GetBalances()
		checks, err := s.read.ListAssertionChecks(ctx, bearer(&viewsv1.ListAssertionChecksRequest{AccountId: id}, sess))
		if err != nil {
			return "ListAssertionChecks", fmt.Errorf("ListAssertionChecks: %w", err)
		}
		data.Checkpoints = checkpointRows(checks.Msg.GetChecks())
	case tabReconciliation:
		checks, err := s.read.ListAssertionChecks(ctx, bearer(&viewsv1.ListAssertionChecksRequest{AccountId: id}, sess))
		if err != nil {
			return "ListAssertionChecks", fmt.Errorf("ListAssertionChecks: %w", err)
		}
		data.Checks = checks.Msg.GetChecks()
	}
	return "", nil
}

// holdings reads an account's holdings and their stored valuations on
// data.Date into data. It returns the method that failed with its error.
func (s *server) holdings(r *http.Request, sess session, data *accountData) (string, error) {
	ctx, id, date := r.Context(), data.Account.GetId(), data.Date
	holdings, err := s.read.ListHoldings(ctx, bearer(&viewsv1.ListHoldingsRequest{AccountId: id, Date: date}, sess))
	if err != nil {
		return "ListHoldings", fmt.Errorf("ListHoldings: %w", err)
	}
	vals, err := s.read.ListAccountValuations(ctx, bearer(&viewsv1.ListAccountValuationsRequest{AccountId: id, From: date, To: date}, sess))
	if err != nil {
		return "ListAccountValuations", fmt.Errorf("ListAccountValuations: %w", err)
	}
	data.Holdings = joinValuations(holdings.Msg.GetHoldings(), vals.Msg.GetValuations())
	return "", nil
}

// transactions reads one page of an account's transactions.
func (s *server) transactions(r *http.Request, sess session, id, token string) (transactionsView, error) {
	resp, err := s.read.ListTransactions(r.Context(), bearer(&viewsv1.ListTransactionsRequest{
		AccountId: id, PageSize: transactionsPageSize, PageToken: token,
	}, sess))
	if err != nil {
		return transactionsView{}, fmt.Errorf("ListTransactions: %w", err)
	}
	view := accountTransactions(id)
	view.Transactions, view.Next = resp.Msg.GetTransactions(), resp.Msg.GetNextPageToken()
	return view, nil
}

// transactionsFragment answers "More": htmx swaps the next page's rows in
// for the button; a browser without htmx gets the account page at that page.
func (s *server) transactionsFragment(w http.ResponseWriter, r *http.Request, sess session) {
	id, token := r.PathValue("id"), r.URL.Query().Get("page_token")
	if !isHTMX(r) {
		http.Redirect(w, r, (&transactionsView{Base: "/accounts/" + url.PathEscape(id), Tab: tabTransactions, Next: token}).NextLink(), http.StatusSeeOther)
		return
	}
	view, err := s.transactions(r, sess, id, token)
	if err != nil {
		s.coreFailed(w, r, "ListTransactions", err)
		return
	}
	s.showTemplate(w, r, http.StatusOK, "account", "transaction-rows", &view)
}
