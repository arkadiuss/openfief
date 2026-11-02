package ui

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"connectrpc.com/connect"

	"github.com/arkadiuss/openfief/internal/kernel"
	valuationv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/valuation/v1"
	viewsv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/views/v1"
)

// startParam is the query parameter that names a charts' start to show
// (D-138) without saving it: a link can preview a start, while the owner's
// own choice is a stored setting (D-167).
const startParam = "start"

// chartStartView is the control that picks where a page's time charts
// start (D-138): the portfolio's on the overview, an account's on its page.
// The owner's choice is a setting the core stores (D-167), which the date
// field saves and going back to the default clears; a read session, such
// as the demo's, picks a start in the page's URL instead, which a link can
// also carry to preview a start without saving it.
type chartStartView struct {
	// Action is the page the form shows again, Date the date it shows.
	Action, Date string
	// From is the day the charts start; Default is the default start,
	// where tracking began, which the charts mark.
	From, Default string
	// Stored is the start the owner saved (D-167); empty for none.
	Stored string
	// Preview reports a start the page's URL names, shown without saving.
	Preview bool
	// Chosen reports charts not on their default start: saved or previewed.
	Chosen bool
	// Account reports an account's charts, whose default is its first
	// day, rather than the portfolio's; AccountID is its ID.
	Account   bool
	AccountID string
	// Owner reports a session that may save the start.
	Owner bool
	// DefaultLink is the page with the charts on their default start:
	// without a start in its URL, or, over a saved one, previewing the
	// default. SavedLink is the page on the start the owner saved, or on
	// the default without one: its URL without a start.
	DefaultLink, SavedLink string
}

// newChartStart returns the start control of the page action on date for
// sess, whose charts start on from, by default on def, with the start the
// owner saved and the one the URL previews, zero for none.
func newChartStart(sess session, action string, date kernel.Date, from, def, stored string, preview kernel.Date) chartStartView {
	saved := url.Values{dateParam: {date.String()}}
	toDefault := url.Values{dateParam: {date.String()}}
	if stored != "" {
		toDefault.Set(startParam, def)
	}
	return chartStartView{
		Action: action, Date: date.String(), From: from, Default: def, Stored: stored,
		Preview: !preview.IsZero(), Chosen: stored != "" || !preview.IsZero(), Owner: sess.Owner(),
		DefaultLink: action + "?" + toDefault.Encode(), SavedLink: action + "?" + saved.Encode(),
	}
}

// parseStart returns the start a query value names, or the zero date for
// an empty one, which means the saved start or the default.
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

// setChartStart saves the posted start as where the portfolio's charts, or
// the posted account's, start (SetChartStart, D-167), or clears it when
// the owner goes back to the default, and shows the page again.
func (s *server) setChartStart(w http.ResponseWriter, r *http.Request, sess session) {
	account := r.PostFormValue("account")
	start := strings.TrimSpace(r.PostFormValue("start"))
	if r.PostFormValue("clear") != "" {
		start = ""
	} else if _, err := kernel.ParseDate(start); err != nil {
		s.errorPage(w, r, http.StatusBadRequest, fmt.Sprintf("%v: start %q", errBadDate, start))
		return
	}
	_, err := s.settings.SetChartStart(r.Context(), connect.NewRequest(&valuationv1.SetChartStartRequest{
		Origin: sess.Origin, SessionToken: sess.Token, AccountId: account, Start: start,
	}))
	if err != nil {
		s.coreFailed(w, r, "SetChartStart", err)
		return
	}
	back := "/"
	if account != "" {
		back = "/accounts/" + url.PathEscape(account)
	}
	q := url.Values{}
	if d, dateErr := kernel.ParseDate(r.PostFormValue(dateParam)); dateErr == nil {
		q.Set(dateParam, d.String())
	}
	if len(q) > 0 {
		back += "?" + q.Encode()
	}
	http.Redirect(w, r, back, http.StatusSeeOther)
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
// from start, or, when start is zero, from the start the owner saved or
// else its first day, up to date.
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
	out.Start = newChartStart(sess, "/accounts/"+url.PathEscape(id), date, from.String(), perf.Msg.GetDefaultFrom(), perf.Msg.GetChosenFrom(), start)
	out.Start.Account, out.Start.AccountID = true, id
	return out, nil
}
