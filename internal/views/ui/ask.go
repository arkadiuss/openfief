package ui

import (
	"regexp"
	"strings"
	"time"

	"github.com/arkadiuss/openfief/internal/kernel"
)

// maxQuestionRunes caps a question the Ask box reads; anything longer is
// not one of its commands.
const maxQuestionRunes = 200

// askCommand is one of the commands the Ask box understands in M0 (D-194).
// The box holds no AI model: it matches a question against these, and
// answers from the read API alone.
type askCommand int

// The commands. The zero command is a question the box doesn't understand.
const (
	askUnknown askCommand = iota
	// askNeeds: what needs the owner, the top of the Inbox.
	askNeeds
	// askShow: an account, a sleeve or an instrument, by name.
	askShow
	// askSync: why a connector instance's sync fails.
	askSync
	// askAllocation: the allocation by a dimension.
	askAllocation
	// askPerformance: the time-weighted return over a period.
	askPerformance
	// askValue: the portfolio's value on a date.
	askValue
)

// showKind narrows what a show command names: anything, or an account, a
// sleeve or an instrument.
type showKind int

// The kinds a show command can name.
const (
	showAny showKind = iota
	showAccount
	showSleeve
	showInstrument
)

// askQuery is a question the Ask box understood: its command and what the
// command names.
type askQuery struct {
	Command askCommand
	// Term is what a show or sync command names, as typed.
	Term string
	// Show narrows what a show command names.
	Show showKind
	// Dimension is an allocation command's dimension key (dimensions).
	Dimension string
	// Period is a performance command's period.
	Period askPeriod
	// Date is a value command's date; zero for the latest valued day.
	Date kernel.Date
	// Change is the period a value command asks the change over, instead of
	// the value on a day; nil for the value on Date.
	Change *askPeriod
}

// periodKind is the kind of span a performance command asks about.
type periodKind int

// The periods. The zero period is this year.
const (
	periodYear periodKind = iota
	periodMonth
	periodTwelveMonths
	periodLastYear
	periodAll
	periodSince
	periodCalendarYear
	// periodMonthOf is a calendar month, such as November 2025.
	periodMonthOf
	// periodBetween is two days, both inclusive.
	periodBetween
)

// askPeriod is a performance command's period: a kind, and the date or
// year it names, if any.
type askPeriod struct {
	Kind periodKind
	// Since is a periodSince's first day, and a periodBetween's.
	Since kernel.Date
	// Year is a periodCalendarYear's or a periodMonthOf's year.
	Year int
	// Month is a periodMonthOf's month.
	Month time.Month
	// To is a periodBetween's last day. Its days are as asked: they may be
	// out of order or not valued, which the answer says.
	To kernel.Date
}

// The questions' patterns, matched against the normalized question
// (normalizeQuestion). Each names its parts.
var (
	needsPattern = regexp.MustCompile(`^(what )?(needs (me|you|my attention)|is open|(do i|i) need to (do|check)|should i (do|check))$|^(inbox|todo|to do|open items|what needs me|needs)$`)
	syncPattern  = regexp.MustCompile(`^why (is|does|do|did|are) (the )?(?P<x>.+?)( syncs?)? (fail|failing|keep failing|keeps failing|not syncing|broken)$|^why (?P<y>.+?) (fails|failed|failing)$|^(sync|syncs|sync status) (of |for )?(?P<z>.+)$`)
	allocPattern = regexp.MustCompile(`^(allocation|split|breakdown|exposure)( (by|per|across) (?P<d>.+))?$|^by (?P<e>.+)$`)
	perfPattern  = regexp.MustCompile(`^(performance|return|returns|twr|time-weighted return)( (?P<p>.+))?$`)
	valuePattern = regexp.MustCompile(`^(?P<c>value change|change in value|value changed|net worth change|change in net worth)( (?P<r>.+))?$|` +
		`^(value|worth|total|net worth)( (?P<w>on|at|as of|for|in|during|over|from|between) (?P<d>.+)| (?P<e>today|yesterday|end of last month|end of last year))?$`)
	showPattern  = regexp.MustCompile(`^(show|open|find|go to)( me)?( the| my)? ((?P<k>account|sleeve|instrument) )?(?P<x>.+)$|^(?P<l>account|sleeve|instrument) (?P<y>.+)$`)
	isoDate      = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)
	dottedDate   = regexp.MustCompile(`^(\d{1,2})\.(\d{1,2})\.(\d{4})$`)
	sincePattern = regexp.MustCompile(`^(since|from) (?P<d>.+)$`)
	yearPattern  = regexp.MustCompile(`^(in |for )?(?P<y>\d{4})$`)
)

// normalizeQuestion returns q in lower case with its surrounding space and
// closing punctuation trimmed and every run of white space one space.
func normalizeQuestion(q string) string {
	q = strings.ToLower(strings.Join(strings.Fields(q), " "))
	return strings.TrimRight(q, "?!. ")
}

// parseAsk returns the command question asks for and what it names, today
// being today; ok is false for a question none of the commands matches.
func parseAsk(question string, today kernel.Date) (askQuery, bool) {
	q := normalizeQuestion(question)
	if q == "" || len([]rune(q)) > maxQuestionRunes {
		return askQuery{}, false
	}
	switch {
	case needsPattern.MatchString(q):
		return askQuery{Command: askNeeds}, true
	case syncPattern.MatchString(q):
		return askQuery{Command: askSync, Term: firstGroup(syncPattern, q, "x", "y", "z")}, true
	case allocPattern.MatchString(q):
		return parseAllocation(firstGroup(allocPattern, q, "d", "e"))
	case perfPattern.MatchString(q):
		p, ok := parsePeriod(firstGroup(perfPattern, q, "p"), today)
		return askQuery{Command: askPerformance, Period: p}, ok
	case valuePattern.MatchString(q):
		return parseValue(q, today)
	case showPattern.MatchString(q):
		kind := showKindOf(firstGroup(showPattern, q, "k", "l"))
		return askQuery{Command: askShow, Term: firstGroup(showPattern, q, "x", "y"), Show: kind}, true
	}
	return askQuery{}, false
}

// parseValue returns the value command q is: the value on a day, or its change
// over a month, a year or two days, as "value change in November 2025", "value
// from 2025-11-01 to 2025-11-30" or "value in 2025"; ok is false for a day or a
// period that is none, and for a change without a period.
func parseValue(q string, today kernel.Date) (askQuery, bool) {
	if firstGroup(valuePattern, q, "c") != "" {
		p, ok := parseChange(firstGroup(valuePattern, q, "r"), today)
		return askQuery{Command: askValue, Change: &p}, ok
	}
	w, words := firstGroup(valuePattern, q, "w"), firstGroup(valuePattern, q, "d", "e")
	switch w {
	case "", "on", "at", "as of":
		d, ok := parseAskDate(words, today)
		return askQuery{Command: askValue, Date: d}, ok
	case "for":
		if d, ok := parseAskDate(words, today); ok {
			return askQuery{Command: askValue, Date: d}, true
		}
	}
	if w == "from" || w == "between" {
		words = w + " " + words
	}
	p, ok := parseChange(words, today)
	if !ok {
		return askQuery{Command: askValue}, false
	}
	return askQuery{Command: askValue, Change: &p}, true
}

// firstGroup returns the first of the named groups of p that matched in s,
// or "".
func firstGroup(p *regexp.Regexp, s string, names ...string) string {
	m := p.FindStringSubmatch(s)
	if m == nil {
		return ""
	}
	for _, n := range names {
		if i := p.SubexpIndex(n); i > 0 && m[i] != "" {
			return m[i]
		}
	}
	return ""
}

// showKindOf returns what a show command's word names.
func showKindOf(word string) showKind {
	switch word {
	case accountParam:
		return showAccount
	case dimSleeve:
		return showSleeve
	case kindInstrument:
		return showInstrument
	}
	return showAny
}

// parseAllocation returns the allocation command by the dimension word
// names, asset class when it names none; ok is false for a word that
// names no dimension.
func parseAllocation(word string) (askQuery, bool) {
	dims := map[string]string{
		"": dimAssetClass, "asset class": dimAssetClass, "asset classes": dimAssetClass, "asset": dimAssetClass,
		"class": dimAssetClass, dimCurrency: dimCurrency, "currencies": dimCurrency, dimInstitution: dimInstitution,
		"institutions": dimInstitution, "broker": dimInstitution, "brokers": dimInstitution, dimSleeve: dimSleeve,
		"sleeves": dimSleeve,
	}
	key, ok := dims[strings.TrimPrefix(word, "the ")]
	return askQuery{Command: askAllocation, Dimension: key}, ok
}

// parsePeriod returns the period words name, this year when they name
// none, today being today; ok is false for words that name no period.
func parsePeriod(words string, today kernel.Date) (askPeriod, bool) {
	fixed := map[string]periodKind{
		"": periodYear, labelThisYear: periodYear, "ytd": periodYear, "year to date": periodYear, "year": periodYear,
		labelThisMonth: periodMonth, "month": periodMonth, "mtd": periodMonth,
		"last 12 months": periodTwelveMonths, "12 months": periodTwelveMonths, "1y": periodTwelveMonths,
		labelLastYear: periodLastYear, tabAll: periodAll, "all time": periodAll, "since tracking began": periodAll,
		"since the start": periodAll,
	}
	if k, ok := fixed[words]; ok {
		return askPeriod{Kind: k}, true
	}
	if p, ok := parseBounded(words, today); ok {
		return p, true
	}
	if y := firstGroup(yearPattern, words, "y"); y != "" {
		d, err := kernel.ParseDate(y + "-01-01")
		return askPeriod{Kind: periodCalendarYear, Year: d.Year()}, err == nil && !d.After(today)
	}
	if s := firstGroup(sincePattern, words, "d"); s != "" {
		d, ok := parseAskDate(s, today)
		return askPeriod{Kind: periodSince, Since: d}, ok && !d.IsZero() && !d.After(today)
	}
	return askPeriod{}, false
}

// parseAskDate returns the date words name, today being today: an ISO
// date, a day.month.year one, today, yesterday, or the end of last month
// or last year; nothing names the zero date, the latest valued day. ok is
// false for words that name no date.
func parseAskDate(words string, today kernel.Date) (kernel.Date, bool) {
	switch words {
	case "":
		return kernel.Date{}, true
	case "today":
		return today, true
	case "yesterday":
		return today.AddDays(-1), true
	case "end of last month":
		return monthStart(today).AddDays(-1), true
	case "end of last year":
		return yearStart(today.Year()).AddDays(-1), true
	}
	if isoDate.MatchString(words) {
		d, err := kernel.ParseDate(words)
		return d, err == nil
	}
	if m := dottedDate.FindStringSubmatch(words); m != nil {
		d, err := kernel.ParseDate(m[3] + "-" + pad2(m[2]) + "-" + pad2(m[1]))
		return d, err == nil
	}
	return kernel.Date{}, false
}

// monthStart returns the first day of d's month.
func monthStart(d kernel.Date) kernel.Date { return d.AddDays(1 - d.Day()) }

// yearStart returns 1 January of year; the zero date for a year a date
// can't have.
func yearStart(year int) kernel.Date {
	d, err := kernel.NewDate(year, time.January, 1)
	if err != nil {
		return kernel.Date{}
	}
	return d
}

// pad2 writes a one- or two-digit number with two digits.
func pad2(s string) string {
	if len(s) == 1 {
		return "0" + s
	}
	return s
}

// labelThisYear names the default period.
const labelThisYear = "this year"

// labelThisMonth names the month period.
const labelThisMonth = "this month"

// labelLastYear names the last calendar year period.
const labelLastYear = "last year"

// span is a performance command's days: its first, the zero date for the
// charts' default start, and its last, with the words naming it.
type span struct {
	From, To kernel.Date
	Label    string
}

// periodRange resolves p to the days it covers, both inclusive, to being the
// latest valued day, the one convention of every performance period (D-61,
// D-194): the period's return runs from the end of the day before From to the
// end of To, so From counts in full. This year runs from 1 January and this
// month from its first day, last calendar year is the year before to's, since
// a day runs from that day, a calendar year is clipped at to, and a month or
// two days are as asked. The last 12 months start 364 days before to, so the
// window opens at the end of the day 365 days before it, the same date a year
// back in a common year, as the card used to start its index. Everything has no
// first day here, the zero date: it starts where tracking began, which the
// read API names, so the caller fills it in. It is pure: no I/O, no clock.
func periodRange(p askPeriod, to kernel.Date) span {
	switch p.Kind {
	case periodYear:
		return span{yearStart(to.Year()), to, labelThisYear}
	case periodMonth:
		return span{monthStart(to), to, labelThisMonth}
	case periodTwelveMonths:
		return span{to.AddDays(1 - daysInYear), to, "over the last 12 months"}
	case periodLastYear:
		return span{yearStart(to.Year() - 1), yearStart(to.Year()).AddDays(-1), labelLastYear}
	case periodAll:
		return span{kernel.Date{}, to, "since tracking began"}
	case periodSince:
		return span{p.Since, to, "since " + p.Since.String()}
	case periodCalendarYear:
		end := yearStart(p.Year + 1).AddDays(-1)
		if end.After(to) {
			end = to
		}
		return span{yearStart(p.Year), end, "in " + yearStart(p.Year).String()[:4]}
	case periodMonthOf, periodBetween:
		return boundedSpan(p)
	}
	return span{yearStart(to.Year()), to, labelThisYear}
}

// daysInYear is how many days the last 12 months' opening day lies before its
// last day.
const daysInYear = 365
