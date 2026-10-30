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

// The query parameters of the overview: the date a page shows, the value
// the allocation splits and the value the value-over-time chart draws.
const (
	dateParam  = "date"
	valueParam = "value"
	chartParam = "chart"
)

// errBadDate reports a date query parameter that is not YYYY-MM-DD.
var errBadDate = errors.New("date must be YYYY-MM-DD")

// accountRow is one account on the overview: its valuation status (D-36),
// its gross, net and liquidation values with cash on the date in the
// reporting currency (D-77, D-106), and how many of its balance
// assertions are flagged.
type accountRow struct {
	Account *viewsv1.Account
	Status  *viewsv1.ValuationStatus
	// Total is the account's total on the date; nil when nothing is valued.
	Total   *viewsv1.DateTotal
	Flagged int
}

// allocationRow is one group of the allocation with its share of the total.
type allocationRow struct {
	Key   string
	Value *viewsv1.Money
	Share string
}

// allocationView is the allocation chart and table for one dimension (D-30).
// It splits one of the three values, gross unless the switch picks
// another (D-77).
type allocationView struct {
	Date       string
	Dimension  dimension
	Dimensions []dimension
	Value      valueKind
	Values     []valueKind
	Total      *viewsv1.Money
	Groups     []allocationRow
	Unvalued   []*viewsv1.UnvaluedHolding
	Chart      string
}

// Link is the URL that shows the allocation by d on the same date.
func (a *allocationView) Link(d dimension) string { return "/?" + a.query(d, a.Value) }

// FragmentLink is Link's htmx counterpart, which returns only the fragment.
func (a *allocationView) FragmentLink(d dimension) string {
	return "/allocation?" + a.query(d, a.Value)
}

// ValueLink is the URL that shows the allocation of value k by the same
// dimension on the same date.
func (a *allocationView) ValueLink(k valueKind) string { return "/?" + a.query(a.Dimension, k) }

// ValueFragmentLink is ValueLink's htmx counterpart.
func (a *allocationView) ValueFragmentLink(k valueKind) string {
	return "/allocation?" + a.query(a.Dimension, k)
}

// query is the allocation's query on the same date, by d, of value k.
func (a *allocationView) query(d dimension, k valueKind) string {
	return url.Values{"by": {d.Key}, valueParam: {k.Key}, dateParam: {a.Date}}.Encode()
}

// overviewData is what the overview shows.
type overviewData struct {
	Date     string
	Accounts []accountRow
	// Total is the portfolio's gross, net and liquidation values with cash
	// on Date; nil when the date has no total.
	Total   *viewsv1.DateTotal
	Flagged int
	// FlaggedPrices counts the flagged price checks (D-22).
	FlaggedPrices int
	Allocation    *allocationView
	// Value is the value-over-time chart (D-106), History the performance
	// chart.
	Value   *valueChartView
	History performanceView
}

// chartData is the data a chart canvas carries for app.js: amounts stay
// exact decimal strings, which the script turns into numbers only to draw.
type chartData struct {
	Kind string `json:"kind"`
	// Group and Decimal are the number format's separators, which the
	// axes and tooltips write numbers with (D-113).
	Group    string   `json:"group"`
	Decimal  string   `json:"decimal"`
	Label    string   `json:"label"`
	Currency string   `json:"currency"`
	Labels   []string `json:"labels"`
	Values   []string `json:"values"`
	// Series are more lines on a line chart, such as the model portfolio's.
	Series []chartSeries `json:"series,omitempty"`
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

// buildAllocation turns GetAllocation's answer into the chart and table,
// the chart writing numbers in f.
func buildAllocation(resp *viewsv1.GetAllocationResponse, d dimension, k valueKind, f numberFormat) *allocationView {
	total := resp.GetTotalReporting()
	v := &allocationView{
		Date: resp.GetDate(), Dimension: d, Dimensions: dimensions(), Value: k, Values: valueKinds(),
		Total: total, Unvalued: resp.GetUnvalued(),
	}
	chart := chartData{
		Kind: "bar", Label: k.Label + " value by " + d.Label, Currency: total.GetCurrency(),
		Group: f.Group, Decimal: f.Decimal,
	}
	for _, g := range resp.GetGroups() {
		value := g.GetValueReporting()
		v.Groups = append(v.Groups, allocationRow{Key: g.GetKey(), Value: value, Share: share(value.GetAmount(), total.GetAmount())})
		chart.Labels = append(chart.Labels, g.GetKey())
		chart.Values = append(chart.Values, value.GetAmount())
	}
	v.Chart = chart.encode()
	return v
}

// overview shows the whole portfolio on one date: the allocation's total
// with how many balance assertions and price checks are flagged, every
// account with its valuation status, the allocation chart and the
// value over time as the time-weighted return, with the model portfolio
// while it is on.
func (s *server) overview(w http.ResponseWriter, r *http.Request, sess session) {
	ctx := r.Context()
	accounts, err := s.read.ListAccounts(ctx, bearer(&viewsv1.ListAccountsRequest{}, sess))
	if err != nil {
		s.coreFailed(w, r, "ListAccounts", err)
		return
	}
	if firstRun(sess, len(accounts.Msg.GetAccounts())) {
		http.Redirect(w, r, setupSleeves, http.StatusSeeOther)
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
	if method, chartErr := s.charts(r, sess, date, &data); chartErr != nil {
		s.coreFailed(w, r, method, chartErr)
		return
	}
	if data.Accounts, err = s.accountRows(r, sess, date, accounts.Msg.GetAccounts(), statuses.Msg.GetStatuses()); err != nil {
		s.coreFailed(w, r, "account rows", err)
		return
	}
	for _, a := range data.Accounts {
		data.Flagged += a.Flagged
	}
	if data.FlaggedPrices, err = s.flaggedPriceChecks(r, sess); err != nil {
		s.coreFailed(w, r, "ListPriceChecks", err)
		return
	}
	s.show(w, r, http.StatusOK, "overview", page{Title: "Overview", Session: &sess, Data: data})
}

// charts reads the overview's charts on date into data: the allocation by
// the dimension and of the value the query picks, the performance from
// go-live, and the value over time over the same days with the portfolio's
// total on date. It returns the method that failed with its error.
func (s *server) charts(r *http.Request, sess session, date kernel.Date, data *overviewData) (string, error) {
	q := r.URL.Query()
	var err error
	if data.Allocation, err = s.allocation(r, sess, date, parseDimension(q.Get("by")), parseValueKind(q.Get(valueParam))); err != nil {
		return "GetAllocation", err
	}
	if data.History, err = s.performance(r, sess, date); err != nil {
		return "GetPerformance", err
	}
	from := date
	if start, startErr := kernel.ParseDate(data.History.From); startErr == nil {
		from = start
	}
	if data.Value, err = s.valueChart(r, sess, from, date, parseChartKind(q.Get(chartParam))); err != nil {
		return "ListValuationTotals", err
	}
	data.Total = data.Value.Last
	return "", nil
}

// allocation reads the allocation of value k on date by d.
func (s *server) allocation(r *http.Request, sess session, date kernel.Date, d dimension, k valueKind) (*allocationView, error) {
	resp, err := s.read.GetAllocation(r.Context(), bearer(&viewsv1.GetAllocationRequest{Date: date.String(), By: d.Enum, Value: k.Enum}, sess))
	if err != nil {
		return nil, fmt.Errorf("GetAllocation: %w", err)
	}
	return buildAllocation(resp.Msg, d, k, requestFormat(r)), nil
}

// accountRows reads, for every account, its total on date and
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
			row.Total = totals[len(totals)-1]
		}
		rows = append(rows, row)
	}
	return rows, nil
}

// allocationFragment answers the allocation's switches: htmx swaps in the
// allocation for another dimension or value; a browser without htmx gets the whole
// overview instead.
func (s *server) allocationFragment(w http.ResponseWriter, r *http.Request, sess session) {
	q := r.URL.Query()
	if !isHTMX(r) {
		back := url.Values{"by": {q.Get("by")}, dateParam: {q.Get(dateParam)}}
		if v := q.Get(valueParam); v != "" {
			back.Set(valueParam, v)
		}
		http.Redirect(w, r, "/?"+back.Encode(), http.StatusSeeOther)
		return
	}
	date, err := asOfDate(q.Get(dateParam), nil, s.today())
	if err != nil {
		s.errorPage(w, r, http.StatusBadRequest, err.Error())
		return
	}
	view, err := s.allocation(r, sess, date, parseDimension(q.Get("by")), parseValueKind(q.Get(valueParam)))
	if err != nil {
		s.coreFailed(w, r, "GetAllocation", err)
		return
	}
	s.showTemplate(w, r, http.StatusOK, "overview", "allocation", view)
}
