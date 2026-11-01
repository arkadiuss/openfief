package ui

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/arkadiuss/openfief/internal/kernel"
	viewsv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/views/v1"
)

// maxChartDays is the longest span the value-over-time chart asks for,
// inside the read API's range limit of 3700 days.
const maxChartDays = 3650

// allValues is the value chart's switch position that draws all three
// lines (D-77).
func allValues() valueKind { return valueKind{Key: "all", Label: "All three"} }

// chartKinds are the value chart's switch positions: the three values,
// gross first, then all three lines.
func chartKinds() []valueKind { return append(valueKinds(), allValues()) }

// parseChartKind returns the switch position a query value names, and
// gross for an empty or unknown one.
func parseChartKind(key string) valueKind {
	if key == allValues().Key {
		return allValues()
	}
	return parseValueKind(key)
}

// valueChartView is the value-over-time chart (D-30, D-106): the
// portfolio's or an account's money, cash included, per day from the
// charts' start (D-138), as one of the three values, gross unless the
// switch picks another, or all three lines (D-77).
type valueChartView struct {
	From, To string
	// Start is the start the owner chose, "" for the default; Mark is
	// where tracking began, which the chart marks (D-138).
	Start, Mark string
	Kind        valueKind
	Kinds       []valueKind
	Currency    string
	Points      int
	// Gaps counts the days the total has no value of Kind, which the line
	// skips, such as liquidation before the marginal rate is set; with all
	// three lines, the days without a liquidation value.
	Gaps  int
	Chart string
	// Last is the total on To; nil when To has none.
	Last *viewsv1.DateTotal
}

// Link is the overview drawing value k, for browsers without htmx.
func (v *valueChartView) Link(k valueKind) string {
	q := url.Values{chartParam: {k.Key}, dateParam: {v.To}}
	if v.Start != "" {
		q.Set(startParam, v.Start)
	}
	return "/?" + q.Encode()
}

// FragmentLink is the URL htmx swaps the chart of value k in from.
func (v *valueChartView) FragmentLink(k valueKind) string { return "/value-chart?" + v.query(k) }

// query is the overview's query that draws value k over the same days.
func (v *valueChartView) query(k valueKind) string {
	q := url.Values{chartParam: {k.Key}, "from": {v.From}, dateParam: {v.To}}
	if v.Start != "" {
		q.Set(startParam, v.Start)
	}
	if v.Mark != "" {
		q.Set(markParam, v.Mark)
	}
	return q.Encode()
}

// markParam is the value chart fragment's query parameter that carries
// where tracking began, the day the chart marks (D-138).
const markParam = "mark"

// buildValueChart turns the totals ListValuationTotals or
// ListAccountValuations answer, in currency, into the chart of value k,
// its numbers written in f, marking the day mark; to is the last day asked
// for.
func buildValueChart(totals []*viewsv1.DateTotal, currency string, k valueKind, f numberFormat, from, to, mark string) *valueChartView {
	v := &valueChartView{From: from, To: to, Mark: mark, Kind: k, Kinds: chartKinds(), Currency: currency}
	kinds := []valueKind{k}
	if k == allValues() {
		kinds = valueKinds()
	}
	chart := chartData{
		Kind: "line", Label: kinds[0].Label + " value with cash", Currency: currency,
		Group: f.Group, Decimal: f.Decimal, Mark: mark,
	}
	for _, more := range kinds[1:] {
		chart.Series = append(chart.Series, chartSeries{Label: more.Label + " value with cash"})
	}
	last := kinds[len(kinds)-1]
	for _, t := range totals {
		chart.Labels = append(chart.Labels, t.GetDate())
		if totalOf(t, last) == nil {
			v.Gaps++
		}
		chart.Values = append(chart.Values, totalOf(t, kinds[0]).GetAmount())
		for i, more := range kinds[1:] {
			chart.Series[i].Values = append(chart.Series[i].Values, totalOf(t, more).GetAmount())
		}
		if t.GetDate() == to {
			v.Last = t
		}
	}
	v.Points = len(chart.Labels)
	v.Chart = chart.encode()
	return v
}

// valueChart reads the portfolio's totals between the dates given, at most
// maxChartDays back, and draws value k, marking the day mark; start is the
// start the owner chose, "" for the default.
func (s *server) valueChart(r *http.Request, sess session, from, date kernel.Date, k valueKind, start, mark string) (*valueChartView, error) {
	if earliest := date.AddDays(-maxChartDays); from.Before(earliest) {
		from = earliest
	}
	if from.After(date) {
		from = date
	}
	resp, err := s.read.ListValuationTotals(r.Context(), bearer(&viewsv1.ListValuationTotalsRequest{
		From: from.String(), To: date.String(),
	}, sess))
	if err != nil {
		return nil, fmt.Errorf("ListValuationTotals: %w", err)
	}
	v := buildValueChart(resp.Msg.GetTotals(), resp.Msg.GetReportingCurrency(), k, requestFormat(r), from.String(), date.String(), mark)
	v.Start = start
	return v, nil
}

// valueChartFragment answers the value-over-time chart's switch: htmx
// swaps in the chart of another value over the same days; a browser
// without htmx gets the whole overview instead.
func (s *server) valueChartFragment(w http.ResponseWriter, r *http.Request, sess session) {
	q := r.URL.Query()
	if !isHTMX(r) {
		back := url.Values{chartParam: {q.Get(chartParam)}, dateParam: {q.Get(dateParam)}}
		if start := q.Get(startParam); start != "" {
			back.Set(startParam, start)
		}
		http.Redirect(w, r, "/?"+back.Encode(), http.StatusSeeOther)
		return
	}
	date, err := asOfDate(q.Get(dateParam), nil, s.today())
	if err != nil {
		s.errorPage(w, r, http.StatusBadRequest, err.Error())
		return
	}
	from, err := asOfDate(q.Get("from"), nil, date)
	if err != nil {
		s.errorPage(w, r, http.StatusBadRequest, err.Error())
		return
	}
	mark, err := parseStart(q.Get(markParam))
	if err != nil {
		s.errorPage(w, r, http.StatusBadRequest, err.Error())
		return
	}
	view, err := s.valueChart(r, sess, from, date, parseChartKind(q.Get(chartParam)), q.Get(startParam), optionalString(mark))
	if err != nil {
		s.coreFailed(w, r, "ListValuationTotals", err)
		return
	}
	s.showTemplate(w, r, http.StatusOK, "overview", "value-chart", view)
}
