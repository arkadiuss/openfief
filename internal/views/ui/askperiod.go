package ui

import (
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"connectrpc.com/connect"

	"github.com/arkadiuss/openfief/internal/kernel"
	viewsv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/views/v1"
)

// A bounded period is a calendar month, a calendar year or two days: the
// performance and value commands answer for it from the read API's
// GetPeriodChange, which computes every figure of the card (D-194, D-203).

// periodPlaces is how many decimal places a percentage of the core shows: the
// two it gives.
const periodPlaces = 2

// monthNamed returns the month an English name, in full or its first three
// letters (and "sept" for September), is; ok is false for any other word.
func monthNamed(name string) (time.Month, bool) {
	for m := time.January; m <= time.December; m++ {
		full := strings.ToLower(m.String())
		if name == full || name == full[:3] || (m == time.September && name == "sept") {
			return m, true
		}
	}
	return 0, false
}

// The patterns of a bounded period's words, matched against what follows the
// command, such as "in november 2025" or "from 2025-11-01 to 2025-11-30".
var (
	monthYearPattern = regexp.MustCompile(`^(in |for |during )?(?P<m>[a-z]+) (?P<y>\d{4})$`)
	isoMonthPattern  = regexp.MustCompile(`^(in |for |during )?(?P<y>\d{4})-(?P<m>\d{2})$`)
	betweenPattern   = regexp.MustCompile(`^(from |between )?(?P<a>\S+) (to|and|until|through) (?P<b>\S+)$`)
)

// parseBounded returns the bounded period words name: a month, as "november
// 2025" or "2025-11", optionally after in, for or during, or two days, as
// "from 2025-11-01 to 2025-11-30" or "between 1.11.2025 and 30.11.2025". ok is
// false for words that name none, and for a month or a day that doesn't exist.
// The days may be dotted dates, today or yesterday, today being today, and
// need not be in order or valued: the answer says so.
func parseBounded(words string, today kernel.Date) (askPeriod, bool) {
	if m := betweenPattern.FindStringSubmatch(words); m != nil {
		from, okFrom := parseAskDate(m[betweenPattern.SubexpIndex("a")], today)
		to, okTo := parseAskDate(m[betweenPattern.SubexpIndex("b")], today)
		if !okFrom || !okTo || from.IsZero() || to.IsZero() {
			return askPeriod{}, false
		}
		return askPeriod{Kind: periodBetween, Since: from, To: to}, true
	}
	if y, m := firstGroup(isoMonthPattern, words, "y"), firstGroup(isoMonthPattern, words, "m"); y != "" {
		return monthOf(y, m)
	}
	if y, name := firstGroup(monthYearPattern, words, "y"), firstGroup(monthYearPattern, words, "m"); y != "" {
		if month, ok := monthNamed(name); ok {
			return monthOf(y, strconv.Itoa(int(month)))
		}
	}
	return askPeriod{}, false
}

// monthOf returns the period of the month, a number as text, of the year, as
// text; ok is false when either is not a real month.
func monthOf(year, month string) (askPeriod, bool) {
	n, err := strconv.Atoi(month)
	if err != nil || n < 1 || n > 12 {
		return askPeriod{}, false
	}
	first, err := kernel.ParseDate(year + "-" + pad2(strconv.Itoa(n)) + "-01")
	if err != nil {
		return askPeriod{}, false
	}
	return askPeriod{Kind: periodMonthOf, Year: first.Year(), Month: first.Month()}, true
}

// parseChange returns the period a value command's change is asked over: a
// bounded period, or a calendar year; ok is false for words that name none.
func parseChange(words string, today kernel.Date) (askPeriod, bool) {
	words = strings.TrimSpace(words)
	if p, ok := parseBounded(words, today); ok {
		return p, true
	}
	if y := firstGroup(yearPattern, words, "y"); y != "" {
		d, err := kernel.ParseDate(y + "-01-01")
		return askPeriod{Kind: periodCalendarYear, Year: d.Year()}, err == nil
	}
	return askPeriod{}, false
}

// boundedSpan returns the days of a bounded period or a calendar year, none of
// them cut at the latest valued day: a period that reaches past it is refused,
// never shortened.
func boundedSpan(p askPeriod) span {
	switch p.Kind {
	case periodMonthOf:
		first := firstOfMonth(p.Year, p.Month)
		return span{first, monthEnd(p.Year, p.Month), "in " + p.Month.String() + " " + strconv.Itoa(p.Year)}
	case periodBetween:
		return span{p.Since, p.To, "from " + p.Since.String() + " to " + p.To.String()}
	case periodCalendarYear:
		return span{yearStart(p.Year), yearStart(p.Year + 1).AddDays(-1), "in " + strconv.Itoa(p.Year)}
	case periodYear, periodMonth, periodTwelveMonths, periodLastYear, periodAll, periodSince:
	}
	return span{}
}

// firstOfMonth returns the first day of the month of the year.
func firstOfMonth(year int, month time.Month) kernel.Date {
	d, err := kernel.NewDate(year, month, 1)
	if err != nil {
		return kernel.Date{}
	}
	return d
}

// monthEnd returns the last day of the month of the year.
func monthEnd(year int, month time.Month) kernel.Date {
	if month == time.December {
		return yearStart(year + 1).AddDays(-1)
	}
	return firstOfMonth(year, month+1).AddDays(-1)
}

// periodProblem says why the days of s can't be answered, latest being the
// latest day any account is valued through, or "" when they can: the period
// must end on or after the day it starts and on or before the latest valued
// day. Nothing is guessed: no period is shortened or turned round.
func periodProblem(s span, latest kernel.Date) string {
	switch {
	case s.From.IsZero() || s.To.IsZero():
		return "The period needs a first and a last day."
	case s.To.Before(s.From):
		return "The period ends on " + s.To.String() + ", before it starts on " + s.From.String() + "."
	case s.To.After(latest):
		return "The period ends on " + s.To.String() + ", but your accounts are valued only through " + latest.String() + "."
	}
	return ""
}

// problemCard is the card of a period that can't be answered, with why.
func problemCard(problem string) answerCard {
	return answerCard{Title: "That period can't be answered", Text: problem}
}

// periodChange reads how the value changed over s, from the read API, or gives
// the card that says why it can't: a period the box refuses, or one the core
// does not value (failed precondition) or does not accept (invalid argument).
// Exactly one of the response and the card is set unless the read fails.
func (a commandAnswerer) periodChange(r *http.Request, sess session, s span, latest kernel.Date) (*viewsv1.GetPeriodChangeResponse, *answerCard, error) {
	if problem := periodProblem(s, latest); problem != "" {
		card := problemCard(problem)
		return nil, &card, nil
	}
	resp, err := a.s.read.GetPeriodChange(r.Context(), bearer(&viewsv1.GetPeriodChangeRequest{From: s.From.String(), To: s.To.String()}, sess))
	if err != nil {
		if problem := coreRefusal(err, s); problem != "" {
			card := problemCard(problem)
			return nil, &card, nil
		}
		return nil, nil, fmt.Errorf("GetPeriodChange: %w", err)
	}
	return resp.Msg, nil, nil
}

// coreRefusal says why the core would not answer the period s: it does not
// value all of it (failed precondition) or does not accept it (invalid
// argument). It is "" for any other error, which is a failure to read.
func coreRefusal(err error, s span) string {
	days := s.From.String() + " to " + s.To.String()
	code := connect.CodeOf(err)
	if code == connect.CodeFailedPrecondition {
		return "Your accounts are not valued for all of " + days + " yet: a day lacks a valuation or an exchange rate."
	}
	if code == connect.CodeInvalidArgument {
		return "The read API doesn't accept " + days + " as a period; it covers at most about ten years."
	}
	return ""
}

// periodPerformance answers the time-weighted return over the bounded period
// s, with the portfolio's value at its start and end and the deposits and
// withdrawals in it, every figure the core's.
func (a commandAnswerer) periodPerformance(r *http.Request, sess session, s span, latest kernel.Date) (answerCard, error) {
	m, problem, err := a.periodChange(r, sess, s, latest)
	if problem != nil || err != nil {
		if problem == nil {
			return answerCard{}, err
		}
		return *problem, nil
	}
	f := requestFormat(r)
	return answerCard{
		Title: "Time-weighted return " + s.Label + ": " + percentText(m.GetReturnPercent(), f),
		Text: "From the end of " + m.GetOpening() + " to the end of " + m.GetTo() + ", gross. Deposits, withdrawals and the transfers you confirmed don't count as return." +
			pendingText(m),
		Rows: []answerRow{
			{Label: "Start", Detail: "end of " + m.GetOpening(), Figure: f.money(m.GetGross().GetStartReporting())},
			{Label: "End", Detail: m.GetTo(), Figure: f.money(m.GetGross().GetEndReporting())},
			{Label: "Deposits", Detail: "paid in during the period", Figure: f.money(m.GetDepositsReporting())},
			{Label: "Withdrawals", Detail: "taken out during the period", Figure: f.money(m.GetWithdrawalsReporting())},
		},
		Link: &answerLink{Href: "/?" + url.Values{overParam: {overPerformance}, startParam: {m.GetOpening()}}.Encode() + "#over-time", Text: "Show the performance chart"},
	}, nil
}

// periodValue answers how the value changed over the period s on each of
// the three bases (D-77), the change in money and in percent, and the net
// deposits in it, every figure the core's.
func (a commandAnswerer) periodValue(r *http.Request, sess session, s span, latest kernel.Date) (answerCard, error) {
	m, problem, err := a.periodChange(r, sess, s, latest)
	if problem != nil || err != nil {
		if problem == nil {
			return answerCard{}, err
		}
		return *problem, nil
	}
	f := requestFormat(r)
	gross := m.GetGross()
	return answerCard{
		Title: "Value " + moved(gross.GetChangeReporting(), f) + " " + s.Label + ", gross",
		Text: "The value at the end of " + m.GetOpening() + " and at the end of " + m.GetTo() + " on each basis, in the reporting currency. " +
			"Net deposits are what was paid in less what was taken out, and are not a gain." + pendingText(m),
		Rows: []answerRow{
			changeRow("Gross", "", gross, f),
			changeRow("Net", "after the taxes due at the planned exit", m.GetNet(), f),
			changeRow("Liquidation", "if everything were sold", m.GetLiquidation(), f),
			{
				Label:  "Net deposits",
				Detail: "paid in " + f.money(m.GetDepositsReporting()) + ", taken out " + f.money(m.GetWithdrawalsReporting()),
				Figure: f.signed(m.GetNetDepositsReporting()),
			},
		},
		Link: &answerLink{Href: "/?" + url.Values{dateParam: {m.GetTo()}}.Encode(), Text: "Open the overview on " + m.GetTo()},
	}, nil
}

// changeRow is the row of one basis: its value at the start and at the end,
// and its change in money and in percent. A basis the core leaves out is
// missing.
func changeRow(label, what string, c *viewsv1.ValueChange, f numberFormat) answerRow {
	if c == nil {
		return answerRow{Label: label, Detail: what, Figure: reasonMissing}
	}
	figure := f.signed(c.GetChangeReporting())
	if c.GetChangePercent() != "" {
		figure += " (" + percentText(c.GetChangePercent(), f) + ")"
	}
	detail := f.money(c.GetStartReporting()) + " to " + f.money(c.GetEndReporting())
	if what != "" {
		detail = what + ": " + detail
	}
	return answerRow{Label: label, Detail: detail, Figure: figure}
}

// moved says whether a change of money rose, fell or stayed, with its size:
// "rose by 1 966,85 PLN".
func moved(change *viewsv1.Money, f numberFormat) string {
	amount := change.GetAmount()
	d, err := kernel.ParseDecimal(amount)
	if err != nil || d.IsZero() {
		return "did not change"
	}
	size := &viewsv1.Money{Amount: strings.TrimPrefix(amount, "-"), Currency: change.GetCurrency()}
	if d.Sign() < 0 {
		return "fell by " + f.money(size)
	}
	return "rose by " + f.money(size)
}

// pendingText is the sentence a card adds while a revaluation could change it.
func pendingText(m *viewsv1.GetPeriodChangeResponse) string {
	if m.GetRevaluationPending() {
		return " A revaluation is pending, so these figures may change."
	}
	return ""
}

// percentText writes a percentage of the core, signed and with its two
// places, in the number format f: "+3,45%". Anything that isn't a decimal is
// returned as is; an empty percentage is empty.
func percentText(s string, f numberFormat) string {
	if s == "" {
		return ""
	}
	if _, err := kernel.ParseDecimal(s); err != nil {
		return s
	}
	whole, frac, _ := strings.Cut(s, ".")
	for len(frac) < periodPlaces {
		frac += "0"
	}
	sign := "+"
	if strings.HasPrefix(whole, "-") {
		sign = ""
	}
	return sign + f.amount(whole+"."+frac) + "%"
}
