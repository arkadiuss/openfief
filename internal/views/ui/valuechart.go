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
// portfolio's money, cash included, per day from go-live, as one of the
// three values, gross unless the switch picks another, or all three lines
// (D-77).
type valueChartView struct {
	From, To string
	Kind     valueKind
	Kinds    []valueKind
	Currency string
	Points   int
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
	return "/?" + url.Values{chartParam: {k.Key}, dateParam: {v.To}}.Encode()
}

// FragmentLink is the URL htmx swaps the chart of value k in from.
func (v *valueChartView) FragmentLink(k valueKind) string { return "/value-chart?" + v.query(k) }

// query is the overview's query that draws value k over the same days.
func (v *valueChartView) query(k valueKind) string {
	return url.Values{chartParam: {k.Key}, "from": {v.From}, dateParam: {v.To}}.Encode()
}

// buildValueChart turns ListValuationTotals' answer into the chart of value
// k, its numbers written in f; to is the last day asked for.
func buildValueChart(resp *viewsv1.ListValuationTotalsResponse, k valueKind, f numberFormat, from, to string) *valueChartView {
	v := &valueChartView{From: from, To: to, Kind: k, Kinds: chartKinds(), Currency: resp.GetReportingCurrency()}
	kinds := []valueKind{k}
	if k == allValues() {
		kinds = valueKinds()
	}
	chart := chartData{
		Kind: "line", Label: kinds[0].Label + " value with cash", Currency: resp.GetReportingCurrency(),
		Group: f.Group, Decimal: f.Decimal,
	}
	for _, more := range kinds[1:] {
		chart.Series = append(chart.Series, chartSeries{Label: more.Label + " value with cash"})
	}
	last := kinds[len(kinds)-1]
	for _, t := range resp.GetTotals() {
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
// maxChartDays back, and draws value k.
func (s *server) valueChart(r *http.Request, sess session, from, date kernel.Date, k valueKind) (*valueChartView, error) {
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
	return buildValueChart(resp.Msg, k, requestFormat(r), from.String(), date.String()), nil
}

// valueChartFragment answers the value-over-time chart's switch: htmx
// swaps in the chart of another value over the same days; a browser
// without htmx gets the whole overview instead.
func (s *server) valueChartFragment(w http.ResponseWriter, r *http.Request, sess session) {
	q := r.URL.Query()
	if !isHTMX(r) {
		http.Redirect(w, r, "/?"+url.Values{chartParam: {q.Get(chartParam)}, dateParam: {q.Get(dateParam)}}.Encode(), http.StatusSeeOther)
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
	view, err := s.valueChart(r, sess, from, date, parseChartKind(q.Get(chartParam)))
	if err != nil {
		s.coreFailed(w, r, "ListValuationTotals", err)
		return
	}
	s.showTemplate(w, r, http.StatusOK, "overview", "value-chart", view)
}
