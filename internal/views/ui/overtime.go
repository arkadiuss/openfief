package ui

import (
	"net/http"
	"net/url"
	"strconv"

	"github.com/arkadiuss/openfief/internal/kernel"
	viewsv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/views/v1"
)

// The chart card's tabs (D-193): the value, the performance, the
// allocation and the three values (D-77).
const (
	tabValue       = "value"
	tabPerformance = overPerformance
	tabAllocation  = overAllocation
	tabAll         = "all"
)

// overTab is one tab of the chart card.
type overTab struct {
	Key, Label string
}

// overTabs are the chart card's tabs, in order.
func overTabs() []overTab {
	return []overTab{{tabValue, "Value"}, {tabPerformance, "Performance"}, {tabAllocation, "Allocation"}, {tabAll, "Gross · net · liquidation"}}
}

// overTimeView is the overview's one chart card (D-193): the value over
// time (D-106) with its switch of the three values (D-77), or the
// performance (D-61), from the charts' start (D-138), or the allocation
// (D-30) by the dimension and of the value its switches pick.
type overTimeView struct {
	Tab  string
	Tabs []overTab
	// Date is the last day drawn; StartParam the start the owner chose,
	// "" for the default.
	Date, StartParam string
	Value            *valueChartView
	History          performanceView
	Start            chartStartView
	// Allocation is the allocation the card's allocation tab shows.
	Allocation *allocationView
}

// Link is the overview showing tab, for browsers without htmx.
func (o *overTimeView) Link(tab string) string { return "/?" + o.query(tab) }

// FragmentLink is the URL htmx swaps the card showing tab in from.
func (o *overTimeView) FragmentLink(tab string) string { return "/over-time?" + o.query(tab) }

// query is the overview's query that shows tab on the same days.
func (o *overTimeView) query(tab string) string {
	q := url.Values{dateParam: {o.Date}}
	switch tab {
	case tabPerformance:
		q.Set(overParam, overPerformance)
	case tabAllocation:
		q.Set(overParam, overAllocation)
		if a := o.Allocation; a != nil {
			q.Set("by", a.Dimension.Key)
			q.Set(valueParam, a.Value.Key)
		}
	case tabAll:
		q.Set(chartParam, allValues().Key)
	}
	if o.StartParam != "" {
		q.Set(startParam, o.StartParam)
	}
	return q.Encode()
}

// overTime reads the chart card on date: the performance from the charts'
// start, the one the owner chose in start or else the default (D-138),
// the value over time over the same days, of the value the query's chart
// parameter picks, and the allocation by the dimension and of the value
// the query picks; the query's over parameter picks the tab.
func (s *server) overTime(r *http.Request, sess session, date, start kernel.Date) (*overTimeView, error) {
	q := r.URL.Query()
	history, err := s.performance(r, sess, start, date)
	if err != nil {
		return nil, err
	}
	from := date
	if first, startErr := kernel.ParseDate(history.From); startErr == nil && !first.After(date) {
		from = first
	}
	kind := parseChartKind(q.Get(chartParam))
	value, err := s.valueChart(r, sess, from, date, kind, optionalString(start), history.Default)
	if err != nil {
		return nil, err
	}
	alloc, err := s.allocation(r, sess, date, parseDimension(q.Get("by")), parseValueKind(q.Get(valueParam)))
	if err != nil {
		return nil, err
	}
	o := &overTimeView{
		Tab: tabValue, Tabs: overTabs(), Date: date.String(), StartParam: optionalString(start), Value: value, History: history,
		Start:      chartStartView{Action: "/", Date: date.String(), From: value.From, Default: history.Default, Chosen: !start.IsZero()},
		Allocation: alloc,
	}
	switch {
	case q.Get(overParam) == overPerformance:
		o.Tab = tabPerformance
	case q.Get(overParam) == overAllocation:
		o.Tab = tabAllocation
	case kind == allValues():
		o.Tab = tabAll
	}
	return o, nil
}

// overTimeFragment answers the chart card's tabs: htmx swaps in the card
// showing another tab; a browser without htmx gets the whole overview.
func (s *server) overTimeFragment(w http.ResponseWriter, r *http.Request, sess session) {
	q := r.URL.Query()
	if !isHTMX(r) {
		back := url.Values{}
		for _, k := range []string{dateParam, startParam, overParam, chartParam, "by", valueParam} {
			if v := q.Get(k); v != "" {
				back.Set(k, v)
			}
		}
		http.Redirect(w, r, "/?"+back.Encode(), http.StatusSeeOther)
		return
	}
	date, err := asOfDate(q.Get(dateParam), nil, s.today())
	if err != nil {
		s.errorPage(w, r, http.StatusBadRequest, err.Error())
		return
	}
	start, err := parseStart(q.Get(startParam))
	if err != nil {
		s.errorPage(w, r, http.StatusBadRequest, err.Error())
		return
	}
	view, err := s.overTime(r, sess, date, start)
	if err != nil {
		s.coreFailed(w, r, "over time", err)
		return
	}
	s.showTemplate(w, r, http.StatusOK, "overview", "over-time", view)
}

// askParam is the query parameter that carries a question to the Ask box
// (D-194). It is never logged: the UI logs a failed call by its method and
// code, and the core's front door logs the path alone.
const askParam = "q"

// askFragment answers a question asked in the Ask box (D-194): htmx swaps
// in its answer card; a browser without htmx gets the overview with the
// answer. It reads the read API only and changes nothing.
func (s *server) askFragment(w http.ResponseWriter, r *http.Request, sess session) {
	question := r.URL.Query().Get(askParam)
	if !isHTMX(r) {
		http.Redirect(w, r, "/?"+url.Values{askParam: {question}}.Encode()+"#ask", http.StatusSeeOther)
		return
	}
	if question == "" {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	card, err := s.ask.answer(r, sess, question)
	if err != nil {
		s.coreFailed(w, r, "ask", err)
		return
	}
	s.showTemplate(w, r, http.StatusOK, "overview", "ask-answer", card)
}

// askHref is the query string that asks q.
func askHref(q string) string { return url.Values{askParam: {q}}.Encode() }

// askLink is the overview answering q, for a link without htmx.
func askLink(q string) string { return "/?" + askHref(q) + "#ask" }

// askGet is the Ask box's fragment answering q, for htmx.
func askGet(q string) string { return "/ask?" + askHref(q) }

// needsAction is the action Needs you offers for item it (D-193): a
// failing sync asks why, every other kind opens the item in the Inbox,
// where it is settled in place.
func needsAction(it *viewsv1.AttentionItem) inboxLink {
	href := inboxPath + "#item-" + strconv.FormatInt(it.GetId(), 10)
	switch it.GetKind() {
	case kindSyncFailing:
		return inboxLink{Href: askLink("Why is " + it.GetConnectorInstance() + " failing?"), Text: "Why?"}
	case kindTicker:
		return inboxLink{Href: href, Text: "Confirm"}
	case kindPriceCheck, kindAssertion:
		return inboxLink{Href: href, Text: "Review"}
	case kindTransferSuggested:
		return inboxLink{Href: href, Text: "Decide"}
	}
	return inboxLink{Href: href, Text: "Open"}
}
