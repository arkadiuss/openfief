package ui

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/arkadiuss/openfief/internal/kernel"
	viewsv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/views/v1"
)

// historyDays is how far back the value-over-time chart asks for: 3650 of
// the 3700 days one ListValuationTotals call may span. The series starts at
// the first date with stored valuations, go-live (D-23), when that is later.
const historyDays = 3650

// dateParam is the query parameter that picks the date a page shows.
const dateParam = "date"

// errBadDate reports a date query parameter that is not YYYY-MM-DD.
var errBadDate = errors.New("date must be YYYY-MM-DD")

// accountRow is one account on the overview: its valuation status (D-36),
// the value of its holdings on the date in the reporting currency, and how
// many of its balance assertions are flagged.
type accountRow struct {
	Account  *viewsv1.Account
	Status   *viewsv1.ValuationStatus
	Holdings *viewsv1.Money
	Flagged  int
}

// allocationRow is one group of the allocation with its share of the total.
type allocationRow struct {
	Key   string
	Value *viewsv1.Money
	Share string
}

// allocationView is the allocation chart and table for one dimension (D-30).
type allocationView struct {
	Date       string
	Dimension  dimension
	Dimensions []dimension
	Total      *viewsv1.Money
	Groups     []allocationRow
	Unvalued   []*viewsv1.UnvaluedHolding
	Chart      string
}

// Link is the URL that shows the allocation by d on the same date.
func (a *allocationView) Link(d dimension) string {
	return "/?" + url.Values{"by": {d.Key}, dateParam: {a.Date}}.Encode()
}

// FragmentLink is Link's htmx counterpart, which returns only the fragment.
func (a *allocationView) FragmentLink(d dimension) string {
	return "/allocation?" + url.Values{"by": {d.Key}, dateParam: {a.Date}}.Encode()
}

// historyView is the value-over-time chart (D-30): the portfolio's stored
// holding valuations per date in the reporting currency.
type historyView struct {
	From     string
	To       string
	Currency string
	Points   int
	Chart    string
}

// overviewData is what the overview shows.
type overviewData struct {
	Date       string
	Accounts   []accountRow
	Flagged    int
	Allocation *allocationView
	History    historyView
}

// chartData is the data a chart canvas carries for app.js: amounts stay
// exact decimal strings, which the script turns into numbers only to draw.
type chartData struct {
	Kind     string   `json:"kind"`
	Label    string   `json:"label"`
	Currency string   `json:"currency"`
	Labels   []string `json:"labels"`
	Values   []string `json:"values"`
}

// encode returns the chart data as JSON for a data attribute.
func (c *chartData) encode() string {
	b, err := json.Marshal(c)
	if err != nil { // strings only: it always encodes
		return "{}"
	}
	return string(b)
}

// asOfDate is the date a page shows: the date query parameter when given,
// otherwise the latest date any account is valued through, otherwise today.
func asOfDate(query string, statuses []*viewsv1.ValuationStatus, today kernel.Date) (kernel.Date, error) {
	if query != "" {
		d, err := kernel.ParseDate(query)
		if err != nil {
			return kernel.Date{}, fmt.Errorf("%w: %w", errBadDate, err)
		}
		return d, nil
	}
	var latest kernel.Date
	for _, s := range statuses {
		d, err := kernel.ParseDate(s.GetValuedThrough())
		if err == nil && d.After(latest) {
			latest = d
		}
	}
	if latest.IsZero() {
		return today, nil
	}
	return latest, nil
}

// buildAllocation turns GetAllocation's answer into the chart and table.
func buildAllocation(resp *viewsv1.GetAllocationResponse, d dimension) *allocationView {
	total := resp.GetTotalReporting()
	v := &allocationView{
		Date: resp.GetDate(), Dimension: d, Dimensions: dimensions(),
		Total: total, Unvalued: resp.GetUnvalued(),
	}
	chart := chartData{Kind: "bar", Label: "Value by " + d.Label, Currency: total.GetCurrency()}
	for _, g := range resp.GetGroups() {
		value := g.GetValueReporting()
		v.Groups = append(v.Groups, allocationRow{Key: g.GetKey(), Value: value, Share: share(value.GetAmount(), total.GetAmount())})
		chart.Labels = append(chart.Labels, g.GetKey())
		chart.Values = append(chart.Values, value.GetAmount())
	}
	v.Chart = chart.encode()
	return v
}

// buildHistory turns ListValuationTotals' answer into the value-over-time
// chart: one point per date with stored valuations.
func buildHistory(resp *viewsv1.ListValuationTotalsResponse) historyView {
	chart := chartData{Kind: "line", Label: "Holdings value", Currency: resp.GetReportingCurrency()}
	for _, t := range resp.GetTotals() {
		chart.Labels = append(chart.Labels, t.GetDate())
		chart.Values = append(chart.Values, t.GetValueReporting().GetAmount())
	}
	v := historyView{Currency: resp.GetReportingCurrency(), Points: len(chart.Labels), Chart: chart.encode()}
	if v.Points > 0 {
		v.From, v.To = chart.Labels[0], chart.Labels[v.Points-1]
	}
	return v
}

// overview shows the whole portfolio on one date: the allocation's total,
// every account with its valuation status, the allocation chart and the
// value over time.
func (s *server) overview(w http.ResponseWriter, r *http.Request, sess session) {
	ctx := r.Context()
	accounts, err := s.read.ListAccounts(ctx, bearer(&viewsv1.ListAccountsRequest{}, sess))
	if err != nil {
		s.coreFailed(w, r, "ListAccounts", err)
		return
	}
	statuses, err := s.read.ListValuationStatuses(ctx, bearer(&viewsv1.ListValuationStatusesRequest{}, sess))
	if err != nil {
		s.coreFailed(w, r, "ListValuationStatuses", err)
		return
	}
	date, err := asOfDate(r.URL.Query().Get(dateParam), statuses.Msg.GetStatuses(), s.today())
	if err != nil {
		s.errorPage(w, r, http.StatusBadRequest, err.Error())
		return
	}
	data := overviewData{Date: date.String()}
	if data.Allocation, err = s.allocation(r, sess, date, parseDimension(r.URL.Query().Get("by"))); err != nil {
		s.coreFailed(w, r, "GetAllocation", err)
		return
	}
	if data.History, err = s.history(r, sess, date); err != nil {
		s.coreFailed(w, r, "ListValuationTotals", err)
		return
	}
	if data.Accounts, err = s.accountRows(r, sess, date, accounts.Msg.GetAccounts(), statuses.Msg.GetStatuses()); err != nil {
		s.coreFailed(w, r, "account rows", err)
		return
	}
	for _, a := range data.Accounts {
		data.Flagged += a.Flagged
	}
	s.show(w, r, http.StatusOK, "overview", page{Title: "Overview", Session: &sess, Data: data})
}

// allocation reads the allocation on date by d.
func (s *server) allocation(r *http.Request, sess session, date kernel.Date, d dimension) (*allocationView, error) {
	resp, err := s.read.GetAllocation(r.Context(), bearer(&viewsv1.GetAllocationRequest{Date: date.String(), By: d.Enum}, sess))
	if err != nil {
		return nil, fmt.Errorf("GetAllocation: %w", err)
	}
	return buildAllocation(resp.Msg, d), nil
}

// history reads the portfolio's value per date up to date.
func (s *server) history(r *http.Request, sess session, date kernel.Date) (historyView, error) {
	resp, err := s.read.ListValuationTotals(r.Context(), bearer(&viewsv1.ListValuationTotalsRequest{
		From: date.AddDays(-historyDays).String(), To: date.String(),
	}, sess))
	if err != nil {
		return historyView{}, fmt.Errorf("ListValuationTotals: %w", err)
	}
	return buildHistory(resp.Msg), nil
}

// accountRows reads, for every account, its holdings' value on date and
// its flagged balance assertions, and pairs it with its valuation status.
func (s *server) accountRows(r *http.Request, sess session, date kernel.Date, accounts []*viewsv1.Account, statuses []*viewsv1.ValuationStatus) ([]accountRow, error) {
	byID := map[string]*viewsv1.ValuationStatus{}
	for _, st := range statuses {
		byID[st.GetAccountId()] = st
	}
	rows := make([]accountRow, 0, len(accounts))
	for _, a := range accounts {
		vals, err := s.read.ListAccountValuations(r.Context(), bearer(&viewsv1.ListAccountValuationsRequest{
			AccountId: a.GetId(), From: date.String(), To: date.String(),
		}, sess))
		if err != nil {
			return nil, fmt.Errorf("ListAccountValuations: %w", err)
		}
		flagged, err := s.read.ListAssertionChecks(r.Context(), bearer(&viewsv1.ListAssertionChecksRequest{
			AccountId: a.GetId(), Status: viewsv1.AssertionStatus_ASSERTION_STATUS_FLAGGED,
		}, sess))
		if err != nil {
			return nil, fmt.Errorf("ListAssertionChecks: %w", err)
		}
		row := accountRow{Account: a, Status: byID[a.GetId()], Flagged: len(flagged.Msg.GetChecks())}
		if totals := vals.Msg.GetTotals(); len(totals) > 0 {
			row.Holdings = totals[len(totals)-1].GetValueReporting()
		}
		rows = append(rows, row)
	}
	return rows, nil
}

// allocationFragment answers the allocation switch: htmx swaps in the
// allocation for another dimension; a browser without htmx gets the whole
// overview instead.
func (s *server) allocationFragment(w http.ResponseWriter, r *http.Request, sess session) {
	q := r.URL.Query()
	if !isHTMX(r) {
		http.Redirect(w, r, "/?"+url.Values{"by": {q.Get("by")}, dateParam: {q.Get(dateParam)}}.Encode(), http.StatusSeeOther)
		return
	}
	date, err := asOfDate(q.Get(dateParam), nil, s.today())
	if err != nil {
		s.errorPage(w, r, http.StatusBadRequest, err.Error())
		return
	}
	view, err := s.allocation(r, sess, date, parseDimension(q.Get("by")))
	if err != nil {
		s.coreFailed(w, r, "GetAllocation", err)
		return
	}
	s.showTemplate(w, r, http.StatusOK, "overview", "allocation", view)
}
