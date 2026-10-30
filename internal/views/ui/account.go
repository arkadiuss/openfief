package ui

import (
	"cmp"
	"fmt"
	"net/http"
	"net/url"

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

// transactionsView is one page of an account's transactions and the link
// to the next one.
type transactionsView struct {
	AccountID    string
	Transactions []*viewsv1.Transaction
	Next         string
}

// NextLink is the URL of the account page that starts at the next page,
// for browsers without htmx.
func (t transactionsView) NextLink() string {
	return "/accounts/" + url.PathEscape(t.AccountID) + "?" + url.Values{"page_token": {t.Next}}.Encode()
}

// NextFragment is the URL htmx loads the next page's rows from.
func (t transactionsView) NextFragment() string {
	return "/accounts/" + url.PathEscape(t.AccountID) + "/transactions?" + url.Values{"page_token": {t.Next}}.Encode()
}

// accountData is what an account's page shows.
type accountData struct {
	Date    string
	Account *viewsv1.Account
	Status  *viewsv1.ValuationStatus
	// Total is the account's gross, net and liquidation values with cash
	// on Date (D-77, D-106); nil when nothing is valued.
	Total        *viewsv1.DateTotal
	Holdings     []holdingRow
	Lots         []*viewsv1.Lot
	Cash         []*viewsv1.Money
	Checks       []*viewsv1.AssertionCheck
	Transactions transactionsView
	// BondLot is the "Add bond lot" form, shown to the owner on a bond
	// account's page only (D-9); nil elsewhere.
	BondLot *bondLotForm
	// UploadForm shows the owner the "Upload statement" form (D-49).
	UploadForm bool
	// Upload is the import the owner's last upload became, to show its
	// counts and ingest problems; nil otherwise.
	Upload *ingestv1.UploadStatementResponse
	// Error is why the core refused what the owner posted.
	Error string
}

// accountForm is what the owner posted to an account's page and why it was
// refused, to show again; the zero value is a fresh page.
type accountForm struct {
	Lot    *bondLotForm
	Upload *ingestv1.UploadStatementResponse
	Error  string
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

// accountPage shows one account on a date: its valuation status, holdings
// with their values, open lots, cash, balance assertion checks, and a page
// of its transactions; to the owner, it also offers the "Upload statement"
// form, and a bond account's page the "Add bond lot" form.
func (s *server) accountPage(w http.ResponseWriter, r *http.Request, sess session) {
	s.renderAccount(w, r, sess, http.StatusOK, accountForm{})
}

// renderAccount shows the account page with status, and with form: the
// bond lot the owner posted and the core's reason for refusing it.
func (s *server) renderAccount(w http.ResponseWriter, r *http.Request, sess session, code int, form accountForm) {
	ctx, id := r.Context(), r.PathValue("id")
	acc, err := s.read.GetAccount(ctx, bearer(&viewsv1.GetAccountRequest{AccountId: id}, sess))
	if err != nil {
		s.coreFailed(w, r, "GetAccount", err)
		return
	}
	status, err := s.read.GetValuationStatus(ctx, bearer(&viewsv1.GetValuationStatusRequest{AccountId: id}, sess))
	if err != nil {
		s.coreFailed(w, r, "GetValuationStatus", err)
		return
	}
	date, err := asOfDate(r.URL.Query().Get(dateParam), []*viewsv1.ValuationStatus{status.Msg.GetStatus()}, s.today())
	if err != nil {
		s.errorPage(w, r, http.StatusBadRequest, err.Error())
		return
	}
	data := accountData{
		Date: date.String(), Account: acc.Msg.GetAccount(), Status: status.Msg.GetStatus(), Error: form.Error,
		UploadForm: sess.Owner(), Upload: form.Upload,
	}
	if sess.Owner() && data.Account.GetInstitution() == bondInstitution {
		data.BondLot = cmp.Or(form.Lot, &bondLotForm{})
	}
	if method, posErr := s.positions(r, sess, &data); posErr != nil {
		s.coreFailed(w, r, method, posErr)
		return
	}
	if data.Transactions, err = s.transactions(r, sess, id, r.URL.Query().Get("page_token")); err != nil {
		s.coreFailed(w, r, "ListTransactions", err)
		return
	}
	s.show(w, r, code, "account", page{Title: data.Account.GetName(), Session: &sess, Data: data})
}

// positions reads an account's holdings, valuations, lots, cash and
// balance assertion checks on data.Date into data. It returns the method
// that failed with its error.
func (s *server) positions(r *http.Request, sess session, data *accountData) (string, error) {
	ctx, id, date := r.Context(), data.Account.GetId(), data.Date
	holdings, err := s.read.ListHoldings(ctx, bearer(&viewsv1.ListHoldingsRequest{AccountId: id, Date: date}, sess))
	if err != nil {
		return "ListHoldings", fmt.Errorf("ListHoldings: %w", err)
	}
	vals, err := s.read.ListAccountValuations(ctx, bearer(&viewsv1.ListAccountValuationsRequest{AccountId: id, From: date, To: date}, sess))
	if err != nil {
		return "ListAccountValuations", fmt.Errorf("ListAccountValuations: %w", err)
	}
	lots, err := s.read.ListLots(ctx, bearer(&viewsv1.ListLotsRequest{AccountId: id, Date: date}, sess))
	if err != nil {
		return "ListLots", fmt.Errorf("ListLots: %w", err)
	}
	cash, err := s.read.ListCash(ctx, bearer(&viewsv1.ListCashRequest{AccountId: id, Date: date}, sess))
	if err != nil {
		return "ListCash", fmt.Errorf("ListCash: %w", err)
	}
	checks, err := s.read.ListAssertionChecks(ctx, bearer(&viewsv1.ListAssertionChecksRequest{AccountId: id}, sess))
	if err != nil {
		return "ListAssertionChecks", fmt.Errorf("ListAssertionChecks: %w", err)
	}
	data.Holdings = joinValuations(holdings.Msg.GetHoldings(), vals.Msg.GetValuations())
	if totals := vals.Msg.GetTotals(); len(totals) > 0 {
		data.Total = totals[len(totals)-1]
	}
	data.Lots, data.Cash, data.Checks = lots.Msg.GetLots(), cash.Msg.GetBalances(), checks.Msg.GetChecks()
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
	return transactionsView{AccountID: id, Transactions: resp.Msg.GetTransactions(), Next: resp.Msg.GetNextPageToken()}, nil
}

// transactionsFragment answers "More": htmx swaps the next page's rows in
// for the button; a browser without htmx gets the account page at that page.
func (s *server) transactionsFragment(w http.ResponseWriter, r *http.Request, sess session) {
	id, token := r.PathValue("id"), r.URL.Query().Get("page_token")
	if !isHTMX(r) {
		http.Redirect(w, r, transactionsView{AccountID: id, Next: token}.NextLink(), http.StatusSeeOther)
		return
	}
	view, err := s.transactions(r, sess, id, token)
	if err != nil {
		s.coreFailed(w, r, "ListTransactions", err)
		return
	}
	s.showTemplate(w, r, http.StatusOK, "account", "transaction-rows", view)
}

// flaggedAccount is an account with its flagged balance assertion checks.
type flaggedAccount struct {
	Account *viewsv1.Account
	Checks  []*viewsv1.AssertionCheck
}

// assertionsPage lists every flagged balance assertion check, per account:
// each is an error for the owner to resolve (D-18).
func (s *server) assertionsPage(w http.ResponseWriter, r *http.Request, sess session) {
	accounts, err := s.read.ListAccounts(r.Context(), bearer(&viewsv1.ListAccountsRequest{}, sess))
	if err != nil {
		s.coreFailed(w, r, "ListAccounts", err)
		return
	}
	var flagged []flaggedAccount
	for _, a := range accounts.Msg.GetAccounts() {
		checks, err := s.read.ListAssertionChecks(r.Context(), bearer(&viewsv1.ListAssertionChecksRequest{
			AccountId: a.GetId(), Status: viewsv1.AssertionStatus_ASSERTION_STATUS_FLAGGED,
		}, sess))
		if err != nil {
			s.coreFailed(w, r, "ListAssertionChecks", err)
			return
		}
		if len(checks.Msg.GetChecks()) > 0 {
			flagged = append(flagged, flaggedAccount{Account: a, Checks: checks.Msg.GetChecks()})
		}
	}
	s.show(w, r, http.StatusOK, "assertions", page{Title: "Flagged balance assertions", Session: &sess, Data: flagged})
}
