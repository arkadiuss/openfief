package ui

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/arkadiuss/openfief/internal/kernel"
	viewsv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/views/v1"
)

// startParam is the query parameter that names the charts' start the owner
// chose (D-138); without it the charts start on the default.
const startParam = "start"

// chartStartView is the control that picks where a page's time charts
// start (D-138): the portfolio's on the overview, an account's on its page.
// It is a view parameter, kept in the page's URL, not a stored setting.
type chartStartView struct {
	// Action is the page the form shows again, Date the date it shows.
	Action, Date string
	// From is the day the charts start; Default is the default start,
	// where tracking began, which the charts mark.
	From, Default string
	// Chosen reports a start the owner chose rather than the default.
	Chosen bool
	// Account reports an account's charts, whose default is its first
	// day, rather than the portfolio's.
	Account bool
}

// DefaultLink is the page with the charts back on their default start.
func (c chartStartView) DefaultLink() string {
	return c.Action + "?" + url.Values{dateParam: {c.Date}}.Encode()
}

// parseStart returns the start a query value names, or the zero date for
// an empty one, which means the default.
func parseStart(q string) (kernel.Date, error) {
	if q == "" {
		return kernel.Date{}, nil
	}
	d, err := kernel.ParseDate(q)
	if err != nil {
		return kernel.Date{}, fmt.Errorf("%w: start %q", errBadDate, q)
	}
	return d, nil
}

// optionalString returns d as YYYY-MM-DD, or "" for the zero date.
func optionalString(d kernel.Date) string {
	if d.IsZero() {
		return ""
	}
	return d.String()
}

// accountCharts is an account's time charts (D-138): its value over time
// and its performance from the start the owner chose, by default its own
// first day.
type accountCharts struct {
	Start       chartStartView
	Value       *valueChartView
	Performance performanceView
}

// accountCharts reads the account id's time-weighted return and totals
// from start, or from its first day when start is zero, up to date.
func (s *server) accountCharts(r *http.Request, sess session, id string, start, date kernel.Date) (*accountCharts, error) {
	perf, err := s.read.GetPerformance(r.Context(), bearer(&viewsv1.GetPerformanceRequest{
		AccountId: id, From: optionalString(start), To: date.String(),
	}, sess))
	if err != nil {
		return nil, fmt.Errorf("GetPerformance: %w", err)
	}
	f := requestFormat(r)
	out := &accountCharts{Performance: buildLine(perf.Msg, nil, f, "Account (time-weighted return)")}
	from := date
	if first, parseErr := kernel.ParseDate(perf.Msg.GetFrom()); parseErr == nil && !first.After(date) {
		from = first
	}
	if earliest := date.AddDays(-maxChartDays); from.Before(earliest) {
		from = earliest
	}
	vals, err := s.read.ListAccountValuations(r.Context(), bearer(&viewsv1.ListAccountValuationsRequest{
		AccountId: id, From: from.String(), To: date.String(),
	}, sess))
	if err != nil {
		return nil, fmt.Errorf("ListAccountValuations: %w", err)
	}
	out.Value = buildValueChart(vals.Msg.GetTotals(), vals.Msg.GetReportingCurrency(), allValues(), f, from.String(), date.String(), perf.Msg.GetDefaultFrom())
	out.Start = chartStartView{
		Action: "/accounts/" + url.PathEscape(id), Date: date.String(),
		From: from.String(), Default: perf.Msg.GetDefaultFrom(), Chosen: !start.IsZero(), Account: true,
	}
	return out, nil
}
