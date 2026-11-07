package ui

import (
	"cmp"
	"fmt"
	"net/http"
	"net/url"
	"slices"
	"strconv"
	"strings"
	"unicode/utf8"

	"connectrpc.com/connect"

	"github.com/arkadiuss/openfief/internal/kernel"
	ingestv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/ingest/v1"
	ledgerv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/ledger/v1"
	viewsv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/views/v1"
)

// The instrument page's limits: how many days of prices its chart draws,
// and the longest display name the rename form takes, as the core does
// (D-197).
const (
	instrumentPriceDays = 365
	maxInstrumentName   = 200
)

// accountParam names, on an instrument's page, the account the owner came
// from, for the breadcrumb.
const accountParam = "account"

// instrumentPath is the page of the instrument with ISIN isin (D-195).
func instrumentPath(isin string) string { return "/instruments/" + url.PathEscape(isin) }

// instrumentFrom is the page of the instrument with ISIN isin reached from
// the account with ID account, whose breadcrumb names that account.
func instrumentFrom(isin, account string) string {
	return instrumentPath(isin) + "?" + url.Values{accountParam: {account}}.Encode()
}

// crumb is one step of a page's breadcrumb; the last has no link.
type crumb struct {
	Href, Text string
}

// overviewCrumb is a breadcrumb's first step, the overview.
func overviewCrumb() crumb { return crumb{Href: "/", Text: "Overview"} }

// accountsCrumb is a breadcrumb's step to the accounts.
func accountsCrumb() crumb { return crumb{Href: accountsPath, Text: "Accounts"} }

// The status pills' classes for good and bad news, and the chart kind of a
// line over time.
const (
	pillOK     = "pill-ok"
	pillDanger = "pill-danger"
	chartLine  = "line"
)

// priceChart is one listing's price history (D-195): the source it is
// valued at as the first line, marked primary, and every other source's
// prices of it, which only cross-check it (D-22, D-94), as more lines.
type priceChart struct {
	Symbol, Primary, Currency string
	Others                    []string
	From, To                  string
	Points                    int
	Chart                     string
}

// instrumentHoldingRow is one account's holding of one listing of the
// instrument, with its open lots.
type instrumentHoldingRow struct {
	AccountID, Account string
	Holding            *viewsv1.InstrumentHolding
}

// instrumentData is what an instrument's page shows (D-195, D-197): its
// display name, tickers and ISIN, its price history with its sources, its
// holdings and lots across the accounts the caller may view, its
// transactions, its price checks with their history, its exemption classes
// and its open attention items; to the owner, the rename form.
type instrumentData struct {
	Date     string
	Detail   *viewsv1.InstrumentDetail
	Crumbs   []crumb
	Prices   []priceChart
	Holdings []instrumentHoldingRow
	Checks   []instrumentChecks
	Items    []inboxItem
	Owner    bool
	Rename   string
	Error    string
	Renamed  bool
	// Confirmed are, for the owner, the tickers confirmed with this ISIN,
	// each with its correction form (D-153); Corrected is how many rows
	// the last correction booked anew, "" for none.
	Confirmed []*ingestv1.ConfirmedTicker
	Corrected string
	Txs       transactionsView
	Accounts  map[string]string
	// Terms are a treasury bond series' stored terms with where each came
	// from (D-97), nil for none; FeeSet says the owner's fee was saved.
	Terms  *seriesTerms
	FeeSet bool
}

// Bond reports whether the instrument is a treasury bond series, whose
// page shows its series' terms.
func (d *instrumentData) Bond() bool { return d.Detail.GetAssetClass() == assetTreasuryBond }

// Name is the instrument's display name, or its ISIN when it has none
// (D-197).
func (d *instrumentData) Name() string {
	if n := d.Detail.GetName(); n != "" {
		return n
	}
	return d.Detail.GetIsin()
}

// Tickers are the symbols the instrument is listed under, in order.
func (d *instrumentData) Tickers() []string {
	out := make([]string, 0, len(d.Detail.GetListings()))
	for _, l := range d.Detail.GetListings() {
		out = append(out, l.GetInstrument().GetSymbol())
	}
	return out
}

// instrumentForm is what the owner posted to an instrument's page and why
// the core refused it; the zero value is a fresh page.
type instrumentForm struct {
	Name      string
	Error     string
	Renamed   bool
	Corrected string
	FeeSet    bool
}

// priceCharts groups an instrument's stored prices by listing (D-195),
// each listing's primary source first, then its other sources in name
// order, their values aligned by date; a day a source has no price is a
// gap. Numbers are written in f.
func priceCharts(prices []*viewsv1.InstrumentPrice, f numberFormat) []priceChart {
	bySymbol := map[string][]*viewsv1.InstrumentPrice{}
	var symbols []string
	for _, p := range prices {
		s := p.GetInstrument().GetSymbol()
		if _, ok := bySymbol[s]; !ok {
			symbols = append(symbols, s)
		}
		bySymbol[s] = append(bySymbol[s], p)
	}
	out := make([]priceChart, 0, len(symbols))
	for _, s := range symbols {
		out = append(out, listingChart(s, bySymbol[s], f))
	}
	return out
}

// listingChart is the price chart of one listing, symbol, from its prices.
func listingChart(symbol string, prices []*viewsv1.InstrumentPrice, f numberFormat) priceChart {
	c := priceChart{Symbol: symbol}
	values := map[string]map[string]string{}
	var dates, sources []string
	for _, p := range prices {
		src := p.GetSource()
		if values[src] == nil {
			values[src] = map[string]string{}
			sources = append(sources, src)
		}
		values[src][p.GetDate()] = p.GetClose().GetAmount()
		dates = append(dates, p.GetDate())
		if p.GetPrimary() {
			c.Primary, c.Currency = src, p.GetClose().GetCurrency()
		}
	}
	slices.Sort(dates)
	dates = slices.Compact(dates)
	slices.Sort(sources)
	if c.Primary == "" && len(sources) > 0 {
		c.Primary, c.Currency = sources[0], prices[0].GetClose().GetCurrency()
	}
	line := func(src string) []string {
		out := make([]string, len(dates))
		for i, d := range dates {
			out[i] = values[src][d]
		}
		return out
	}
	chart := chartData{
		Kind: chartLine, Label: symbol + " · " + c.Primary + " (primary)", Currency: c.Currency, Labels: dates,
		Values: line(c.Primary), Group: f.Group, Decimal: f.Decimal,
	}
	for _, src := range sources {
		if src != c.Primary {
			c.Others = append(c.Others, src)
			chart.Series = append(chart.Series, chartSeries{Label: symbol + " · " + src, Values: line(src)})
		}
	}
	c.Points = len(dates)
	if c.Points > 0 {
		c.From, c.To = dates[0], dates[c.Points-1]
	}
	c.Chart = chart.encode()
	return c
}

// instrumentHoldings pairs each holding with its account's name in names.
func instrumentHoldings(holdings []*viewsv1.InstrumentHolding, names map[string]string) []instrumentHoldingRow {
	out := make([]instrumentHoldingRow, 0, len(holdings))
	for _, h := range holdings {
		id := h.GetAccountId()
		name := names[id]
		if name == "" {
			name = id
		}
		out = append(out, instrumentHoldingRow{AccountID: id, Account: name, Holding: h})
	}
	return out
}

// instrumentCrumbs is an instrument's breadcrumb: the overview, then the
// accounts and the account the owner came from, when known, then the
// instrument's name.
func instrumentCrumbs(name, account string, names map[string]string) []crumb {
	out := []crumb{overviewCrumb()}
	if n, ok := names[account]; ok {
		out = append(out, accountsCrumb(), crumb{Href: "/accounts/" + url.PathEscape(account), Text: n})
	}
	return append(out, crumb{Text: name})
}

// instrumentPage shows one instrument (D-195).
func (s *server) instrumentPage(w http.ResponseWriter, r *http.Request, sess session) {
	form := instrumentForm{Renamed: r.URL.Query().Has("renamed"), FeeSet: r.URL.Query().Has("fee")}
	if n, err := strconv.Atoi(r.URL.Query().Get("corrected")); err == nil && n >= 0 {
		form.Corrected = strconv.Itoa(n)
	}
	s.renderInstrument(w, r, sess, r.PathValue("isin"), http.StatusOK, form)
}

// renderInstrument shows the page of the instrument with ISIN isin with
// status, and with form: the name the owner posted and the core's reason
// for refusing what was posted.
func (s *server) renderInstrument(w http.ResponseWriter, r *http.Request, sess session, isin string, code int, form instrumentForm) {
	data, method, err := s.instrumentView(r, sess, isin, form)
	switch {
	case err != nil && method == "":
		s.errorPage(w, r, http.StatusBadRequest, err.Error())
	case err != nil:
		s.coreFailed(w, r, method, err)
	default:
		s.show(w, r, code, "instrument", page{Title: data.Name(), Session: &sess, Data: data})
	}
}

// instrumentView reads what the page of the instrument with ISIN isin
// shows, with form as posted. On failure it returns the method that failed
// with its error, or no method for a malformed query.
func (s *server) instrumentView(r *http.Request, sess session, isin string, form instrumentForm) (*instrumentData, string, error) {
	in, err := s.read.GetInstrument(r.Context(), bearer(&viewsv1.GetInstrumentRequest{Isin: isin}, sess))
	if err != nil {
		return nil, "GetInstrument", fmt.Errorf("GetInstrument: %w", err)
	}
	statuses, err := s.read.ListValuationStatuses(r.Context(), bearer(&viewsv1.ListValuationStatusesRequest{}, sess))
	if err != nil {
		return nil, "ListValuationStatuses", fmt.Errorf("ListValuationStatuses: %w", err)
	}
	date, err := asOfDate(r.URL.Query().Get(dateParam), statuses.Msg.GetStatuses(), s.today())
	if err != nil {
		return nil, "", err
	}
	data := &instrumentData{
		Date: date.String(), Detail: in.Msg.GetInstrument(), Owner: sess.Owner(), Error: form.Error, Renamed: form.Renamed, Corrected: form.Corrected,
		Rename: cmp.Or(form.Name, in.Msg.GetInstrument().GetName()), Items: s.instrumentItems(r, sess, isin),
		Terms: seriesTermsView(in.Msg.GetInstrument().GetBondTerms(), requestFormat(r)), FeeSet: form.FeeSet,
	}
	if method, readErr := s.instrumentRecord(r, sess, date, data); readErr != nil {
		return nil, method, readErr
	}
	if data.Confirmed, err = s.confirmedTickers(r, sess, isin); err != nil {
		return nil, "ListConfirmedTickers", err
	}
	data.Crumbs = instrumentCrumbs(data.Name(), r.URL.Query().Get(accountParam), data.Accounts)
	return data, "", nil
}

// checksReturnTo sets next, the page the owner's price check forms return to, on
// checks and each of their rows.
func checksReturnTo(checks []instrumentChecks, next string) {
	for i := range checks {
		c := &checks[i]
		c.Next, c.Latest.Next = next, next
		for j := range c.History {
			c.History[j].Next = next
		}
	}
}

// afterInstrumentForm answers a form on the page of the instrument with
// ISIN isin that the core refused: the page with the core's reason, or
// any other failure like any failed call to the core.
func (s *server) afterInstrumentForm(w http.ResponseWriter, r *http.Request, sess session, isin, name string, err error) {
	code := connect.CodeOf(err)
	if status, refused := refusalStatus(code); refused && code != connect.CodePermissionDenied && code != connect.CodeNotFound {
		s.renderInstrument(w, r, sess, isin, status, instrumentForm{Name: r.PostFormValue("name"), Error: "The core refused this: " + connectMessage(err)})
		return
	}
	s.coreFailed(w, r, name, err)
}

// instrumentRecord reads what the instrument page shows of the record on
// date into data: the account names, the prices, the holdings, a page of
// transactions and the price checks. It returns the method that failed
// with its error.
func (s *server) instrumentRecord(r *http.Request, sess session, date kernel.Date, data *instrumentData) (string, error) {
	ctx, isin, f := r.Context(), data.Detail.GetIsin(), requestFormat(r)
	accounts, err := s.read.ListAccounts(ctx, bearer(&viewsv1.ListAccountsRequest{}, sess))
	if err != nil {
		return "ListAccounts", fmt.Errorf("ListAccounts: %w", err)
	}
	data.Accounts = map[string]string{}
	for _, a := range accounts.Msg.GetAccounts() {
		data.Accounts[a.GetId()] = a.GetName()
	}
	prices, err := s.read.ListInstrumentPrices(ctx, bearer(&viewsv1.ListInstrumentPricesRequest{
		Isin: isin, From: date.AddDays(-instrumentPriceDays).String(), To: date.String(),
	}, sess))
	if err != nil {
		return "ListInstrumentPrices", fmt.Errorf("ListInstrumentPrices: %w", err)
	}
	data.Prices = priceCharts(prices.Msg.GetPrices(), f)
	held, err := s.read.ListInstrumentHoldings(ctx, bearer(&viewsv1.ListInstrumentHoldingsRequest{Isin: isin, Date: date.String()}, sess))
	if err != nil {
		return "ListInstrumentHoldings", fmt.Errorf("ListInstrumentHoldings: %w", err)
	}
	data.Holdings = instrumentHoldings(held.Msg.GetHoldings(), data.Accounts)
	if data.Txs, err = s.instrumentTransactions(r, sess, isin, r.URL.Query().Get("page_token")); err != nil {
		return "ListInstrumentTransactions", err
	}
	data.Txs.Accounts = data.Accounts
	checks, err := s.read.ListPriceChecks(ctx, bearer(&viewsv1.ListPriceChecksRequest{Isin: isin}, sess))
	if err != nil {
		return "ListPriceChecks", fmt.Errorf("ListPriceChecks: %w", err)
	}
	data.Checks = latestPerInstrument(checks.Msg.GetChecks(), checks.Msg.GetThreshold(), sess.Owner())
	checksReturnTo(data.Checks, instrumentPath(isin))
	return "", nil
}

// instrumentItems lists the open attention items about the instrument with
// ISIN isin, for the banner on its page (D-196); a core that can't list
// them shows none, and the page still renders.
func (s *server) instrumentItems(r *http.Request, sess session, isin string) []inboxItem {
	resp, err := s.read.ListAttentionItems(r.Context(), bearer(&viewsv1.ListAttentionItemsRequest{
		View: viewsv1.AttentionView_ATTENTION_VIEW_OPEN, Isin: isin,
	}, sess))
	if err != nil {
		s.log.WarnContext(r.Context(), "listing an instrument's attention items", "code", connect.CodeOf(err).String())
		return nil
	}
	out := make([]inboxItem, 0, len(resp.Msg.GetItems()))
	for _, it := range resp.Msg.GetItems() {
		out = append(out, inboxItem{Item: it, Title: itemTitle(it), Pill: itemPill(it)})
	}
	return out
}

// instrumentTransactions reads one page of the instrument's transactions
// across the accounts the caller may view.
func (s *server) instrumentTransactions(r *http.Request, sess session, isin, token string) (transactionsView, error) {
	resp, err := s.read.ListInstrumentTransactions(r.Context(), bearer(&viewsv1.ListInstrumentTransactionsRequest{
		Isin: isin, PageSize: transactionsPageSize, PageToken: token,
	}, sess))
	if err != nil {
		return transactionsView{}, fmt.Errorf("ListInstrumentTransactions: %w", err)
	}
	return transactionsView{
		Base: instrumentPath(isin), Transactions: resp.Msg.GetTransactions(), Next: resp.Msg.GetNextPageToken(), ShowAccount: true,
	}, nil
}

// instrumentTransactionsFragment answers "More" on an instrument's page:
// htmx swaps the next page's rows in; a browser without htmx gets the
// instrument page at that page.
func (s *server) instrumentTransactionsFragment(w http.ResponseWriter, r *http.Request, sess session) {
	isin, token := r.PathValue("isin"), r.URL.Query().Get("page_token")
	if !isHTMX(r) {
		http.Redirect(w, r, (&transactionsView{Base: instrumentPath(isin), Next: token}).NextLink(), http.StatusSeeOther)
		return
	}
	view, err := s.instrumentTransactions(r, sess, isin, token)
	if err != nil {
		s.coreFailed(w, r, "ListInstrumentTransactions", err)
		return
	}
	accounts, err := s.read.ListAccounts(r.Context(), bearer(&viewsv1.ListAccountsRequest{}, sess))
	if err != nil {
		s.coreFailed(w, r, "ListAccounts", err)
		return
	}
	view.Accounts = map[string]string{}
	for _, a := range accounts.Msg.GetAccounts() {
		view.Accounts[a.GetId()] = a.GetName()
	}
	s.showTemplate(w, r, http.StatusOK, "instrument", "transaction-rows", &view)
}

// renameInstrument hands the owner's display name for the instrument to
// the core (D-197), then shows its page again with the stored name; a
// refusal shows the core's reason with the name as posted.
func (s *server) renameInstrument(w http.ResponseWriter, r *http.Request, sess session) {
	if err := r.ParseForm(); err != nil {
		s.errorPage(w, r, http.StatusBadRequest, "The form is malformed.")
		return
	}
	fields, ok := postedOnce(r, "name")
	if !ok {
		s.errorPage(w, r, http.StatusBadRequest, "The form is malformed.")
		return
	}
	name, isin := strings.TrimSpace(fields[0]), r.PathValue("isin")
	if name == "" || utf8.RuneCountInString(name) > maxInstrumentName {
		s.renderInstrument(w, r, sess, isin, http.StatusBadRequest, instrumentForm{Name: name, Error: "A name must be 1 to 200 characters."})
		return
	}
	_, err := s.instruments.RenameInstrument(r.Context(), connect.NewRequest(&ledgerv1.RenameInstrumentRequest{
		Origin: sess.Origin, SessionToken: sess.Token, Isin: isin, Name: name,
	}))
	if err == nil {
		http.Redirect(w, r, "/instruments/"+url.PathEscape(isin)+"?renamed=1", http.StatusSeeOther)
		return
	}
	s.afterInstrumentForm(w, r, sess, isin, "RenameInstrument", err)
}
