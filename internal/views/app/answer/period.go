package answer

import (
	"context"
	"fmt"
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
func parseBounded(words string, today kernel.Date) (Period, bool) {
	if m := betweenPattern.FindStringSubmatch(words); m != nil {
		from, okFrom := parseAskDate(m[betweenPattern.SubexpIndex("a")], today)
		to, okTo := parseAskDate(m[betweenPattern.SubexpIndex("b")], today)
		if !okFrom || !okTo || from.IsZero() || to.IsZero() {
			return Period{}, false
		}
		return Period{Kind: PeriodBetween, Since: from, To: to}, true
	}
	if y, m := firstGroup(isoMonthPattern, words, "y"), firstGroup(isoMonthPattern, words, "m"); y != "" {
		return monthOf(y, m)
	}
	if y, name := firstGroup(monthYearPattern, words, "y"), firstGroup(monthYearPattern, words, "m"); y != "" {
		if month, ok := monthNamed(name); ok {
			return monthOf(y, strconv.Itoa(int(month)))
		}
	}
	return Period{}, false
}

// monthOf returns the period of the month, a number as text, of the year, as
// text; ok is false when either is not a real month.
func monthOf(year, month string) (Period, bool) {
	n, err := strconv.Atoi(month)
	if err != nil || n < 1 || n > 12 {
		return Period{}, false
	}
	first, err := kernel.ParseDate(year + "-" + pad2(strconv.Itoa(n)) + "-01")
	if err != nil {
		return Period{}, false
	}
	return Period{Kind: PeriodMonthOf, Year: first.Year(), Month: first.Month()}, true
}

// parseChange returns the period a value command's change is asked over: a
// bounded period, or a calendar year; ok is false for words that name none.
func parseChange(words string, today kernel.Date) (Period, bool) {
	words = strings.TrimSpace(words)
	if p, ok := parseBounded(words, today); ok {
		return p, true
	}
	if y := firstGroup(yearPattern, words, "y"); y != "" {
		d, err := kernel.ParseDate(y + "-01-01")
		return Period{Kind: PeriodCalendarYear, Year: d.Year()}, err == nil
	}
	return Period{}, false
}

// boundedSpan returns the days of a bounded period or a calendar year, none of
// them cut at the latest valued day: a period that reaches past it is refused,
// never shortened.
func boundedSpan(p Period) span {
	switch p.Kind {
	case PeriodMonthOf:
		first := firstOfMonth(p.Year, p.Month)
		return span{first, monthEnd(p.Year, p.Month), "in " + p.Month.String() + " " + strconv.Itoa(p.Year)}
	case PeriodBetween:
		return span{p.Since, p.To, "from " + p.Since.String() + " to " + p.To.String()}
	case PeriodCalendarYear:
		return span{YearStart(p.Year), YearStart(p.Year + 1).AddDays(-1), "in " + strconv.Itoa(p.Year)}
	case PeriodYear, PeriodMonth, PeriodTwelveMonths, PeriodLastYear, PeriodAll, PeriodSince:
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
		return YearStart(year + 1).AddDays(-1)
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
func problemCard(problem string) Card {
	return Card{Title: "That period can't be answered", Text: problem}
}

// periodChange reads how the value changed over s, from the read API, or gives
// the card that says why it can't: a period the box refuses, or one the core
// does not value (failed precondition) or does not accept (invalid argument).
// Exactly one of the response and the card is set unless the read fails.
func (cm *Commands) periodChange(ctx context.Context, c Caller, s span, latest kernel.Date) (*viewsv1.GetPeriodChangeResponse, *Card, error) {
	if problem := periodProblem(s, latest); problem != "" {
		card := problemCard(problem)
		return nil, &card, nil
	}
	resp, err := cm.read.GetPeriodChange(ctx, bearer(&viewsv1.GetPeriodChangeRequest{From: s.From.String(), To: s.To.String()}, c))
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
func (cm *Commands) periodPerformance(ctx context.Context, c Caller, s span, latest kernel.Date) (Card, error) {
	m, problem, err := cm.periodChange(ctx, c, s, latest)
	if problem != nil || err != nil {
		if problem == nil {
			return Card{}, err
		}
		return *problem, nil
	}
	f := c.Format
	return Card{
		Title: "Time-weighted return " + s.Label + ": " + percentText(m.GetReturnPercent(), f),
		Text: "From the end of " + m.GetOpening() + " to the end of " + m.GetTo() + ", gross. Deposits, withdrawals and the transfers you confirmed don't count as return." +
			pendingText(m),
		Rows: []Row{
			{Label: "Start", Detail: "end of " + m.GetOpening(), Figure: f.Money(m.GetGross().GetStartReporting())},
			{Label: "End", Detail: m.GetTo(), Figure: f.Money(m.GetGross().GetEndReporting())},
			{Label: "Deposits", Detail: "paid in during the period", Figure: f.Money(m.GetDepositsReporting())},
			{Label: "Withdrawals", Detail: "taken out during the period", Figure: f.Money(m.GetWithdrawalsReporting())},
		},
		Link: &Link{Href: cm.links.PerformanceChart(m.GetOpening()), Text: "Show the performance chart"},
	}, nil
}

// periodValue answers how the value changed over the period s on each of
// the three bases (D-77), the change in money and in percent, and the net
// deposits in it, every figure the core's.
func (cm *Commands) periodValue(ctx context.Context, c Caller, s span, latest kernel.Date) (Card, error) {
	m, problem, err := cm.periodChange(ctx, c, s, latest)
	if problem != nil || err != nil {
		if problem == nil {
			return Card{}, err
		}
		return *problem, nil
	}
	f := c.Format
	gross := m.GetGross()
	return Card{
		Title: "Value " + moved(gross.GetChangeReporting(), f) + " " + s.Label + ", gross",
		Text: "The value at the end of " + m.GetOpening() + " and at the end of " + m.GetTo() + " on each basis, in the reporting currency. " +
			"Net deposits are what was paid in less what was taken out, and are not a gain." + pendingText(m),
		Rows: []Row{
			changeRow("Gross", "", gross, f),
			changeRow("Net", "after the taxes due at the planned exit", m.GetNet(), f),
			changeRow("Liquidation", "if everything were sold", m.GetLiquidation(), f),
			{
				Label:  "Net deposits",
				Detail: "paid in " + f.Money(m.GetDepositsReporting()) + ", taken out " + f.Money(m.GetWithdrawalsReporting()),
				Figure: f.Signed(m.GetNetDepositsReporting()),
			},
		},
		Link: &Link{Href: cm.links.Overview(m.GetTo()), Text: "Open the overview on " + m.GetTo()},
	}, nil
}

// changeRow is the row of one basis: its value at the start and at the end,
// and its change in money and in percent. A basis the core leaves out is
// missing.
func changeRow(label, what string, c *viewsv1.ValueChange, f NumberFormat) Row {
	if c == nil {
		return Row{Label: label, Detail: what, Figure: reasonMissing}
	}
	figure := f.Signed(c.GetChangeReporting())
	if c.GetChangePercent() != "" {
		figure += " (" + percentText(c.GetChangePercent(), f) + ")"
	}
	detail := f.Money(c.GetStartReporting()) + " to " + f.Money(c.GetEndReporting())
	if what != "" {
		detail = what + ": " + detail
	}
	return Row{Label: label, Detail: detail, Figure: figure}
}

// moved says whether a change of money rose, fell or stayed, with its size:
// "rose by 1 966,85 PLN".
func moved(change *viewsv1.Money, f NumberFormat) string {
	amount := change.GetAmount()
	d, err := kernel.ParseDecimal(amount)
	if err != nil || d.IsZero() {
		return "did not change"
	}
	size := &viewsv1.Money{Amount: strings.TrimPrefix(amount, "-"), Currency: change.GetCurrency()}
	if d.Sign() < 0 {
		return "fell by " + f.Money(size)
	}
	return "rose by " + f.Money(size)
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
func percentText(s string, f NumberFormat) string {
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
	return sign + f.Amount(whole+"."+frac) + "%"
}

// PeriodOpening returns the day a period's return opens on, the close of
// the day before its first day (D-223), so the first day counts in full: the
// date the charts' start is set to wherever the period opens on the
// performance chart, from the overview's link or an Ask box card. It is the
// zero date for a period with no first day (everything: the charts' default
// start). It resolves p through periodRange, the one place a period's days
// are decided.
func PeriodOpening(p Period, to kernel.Date) kernel.Date {
	from := periodRange(p, to).From
	if from.IsZero() {
		return from
	}
	return from.AddDays(-1)
}
