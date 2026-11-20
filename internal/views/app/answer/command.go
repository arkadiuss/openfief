package answer

import (
	"regexp"
	"strings"
	"time"

	"github.com/arkadiuss/openfief/internal/kernel"
)

// maxQuestionRunes caps a question the Ask box reads; anything longer is
// not one of its commands.
const maxQuestionRunes = 200

// Command is one of the commands the Ask box understands in M0 (D-194).
// The box holds no AI model: it matches a question against these, and
// answers from the read API alone.
type Command int

// The commands. The zero command is a question the box doesn't understand.
const (
	CommandUnknown Command = iota
	// CommandNeeds: what needs the owner, the top of the Inbox.
	CommandNeeds
	// CommandShow: an account, a sleeve or an instrument, by name.
	CommandShow
	// CommandSync: why a connector instance's sync fails.
	CommandSync
	// CommandAllocation: the allocation by a dimension.
	CommandAllocation
	// CommandPerformance: the time-weighted return over a period.
	CommandPerformance
	// CommandValue: the portfolio's value on a date.
	CommandValue
)

// ShowKind narrows what a show command names: anything, or an account, a
// sleeve or an instrument.
type ShowKind int

// The kinds a show command can name.
const (
	ShowAny ShowKind = iota
	ShowAccount
	ShowSleeve
	ShowInstrument
)

// Query is a question the Ask box understood: its command and what the
// command names.
type Query struct {
	Command Command
	// Term is what a show or sync command names, as typed.
	Term string
	// Show narrows what a show command names.
	Show ShowKind
	// Dimension is an allocation command's dimension key (dimensions).
	Dimension string
	// Period is a performance command's period.
	Period Period
	// Date is a value command's date; zero for the latest valued day.
	Date kernel.Date
	// Change is the period a value command asks the change over, instead of
	// the value on a day; nil for the value on Date.
	Change *Period
}

// PeriodKind is the kind of span a performance command asks about.
type PeriodKind int

// The periods. The zero period is this year.
const (
	PeriodYear PeriodKind = iota
	PeriodMonth
	PeriodTwelveMonths
	PeriodLastYear
	PeriodAll
	PeriodSince
	PeriodCalendarYear
	// PeriodMonthOf is a calendar month, such as November 2025.
	PeriodMonthOf
	// PeriodBetween is two days, both inclusive.
	PeriodBetween
)

// Period is a performance command's period: a kind, and the date or
// year it names, if any.
type Period struct {
	Kind PeriodKind
	// Since is a PeriodSince's first day, and a PeriodBetween's.
	Since kernel.Date
	// Year is a PeriodCalendarYear's or a PeriodMonthOf's year.
	Year int
	// Month is a PeriodMonthOf's month.
	Month time.Month
	// To is a PeriodBetween's last day. Its days are as asked: they may be
	// out of order or not valued, which the answer says.
	To kernel.Date
}

// The questions' patterns, matched against the normalized question
// (NormalizeQuestion). Each names its parts.
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

// NormalizeQuestion returns q in lower case with its surrounding space and
// closing punctuation trimmed and every run of white space one space.
func NormalizeQuestion(q string) string {
	q = strings.ToLower(strings.Join(strings.Fields(q), " "))
	return strings.TrimRight(q, "?!. ")
}

// Parse returns the command question asks for and what it names, today
// being today; ok is false for a question none of the commands matches.
func Parse(question string, today kernel.Date) (Query, bool) {
	q := NormalizeQuestion(question)
	if q == "" || len([]rune(q)) > maxQuestionRunes {
		return Query{}, false
	}
	switch {
	case needsPattern.MatchString(q):
		return Query{Command: CommandNeeds}, true
	case syncPattern.MatchString(q):
		return Query{Command: CommandSync, Term: firstGroup(syncPattern, q, "x", "y", "z")}, true
	case allocPattern.MatchString(q):
		return parseAllocation(firstGroup(allocPattern, q, "d", "e"))
	case perfPattern.MatchString(q):
		p, ok := parsePeriod(firstGroup(perfPattern, q, "p"), today)
		return Query{Command: CommandPerformance, Period: p}, ok
	case valuePattern.MatchString(q):
		return parseValue(q, today)
	case showPattern.MatchString(q):
		kind := showKindOf(firstGroup(showPattern, q, "k", "l"))
		return Query{Command: CommandShow, Term: firstGroup(showPattern, q, "x", "y"), Show: kind}, true
	}
	return Query{}, false
}

// parseValue returns the value command q is: the value on a day, or its change
// over a month, a year or two days, as "value change in November 2025", "value
// from 2025-11-01 to 2025-11-30" or "value in 2025"; ok is false for a day or a
// period that is none, and for a change without a period.
func parseValue(q string, today kernel.Date) (Query, bool) {
	if firstGroup(valuePattern, q, "c") != "" {
		p, ok := parseChange(firstGroup(valuePattern, q, "r"), today)
		return Query{Command: CommandValue, Change: &p}, ok
	}
	w, words := firstGroup(valuePattern, q, "w"), firstGroup(valuePattern, q, "d", "e")
	switch w {
	case "", "on", "at", "as of":
		d, ok := parseAskDate(words, today)
		return Query{Command: CommandValue, Date: d}, ok
	case "for":
		if d, ok := parseAskDate(words, today); ok {
			return Query{Command: CommandValue, Date: d}, true
		}
	}
	if w == "from" || w == "between" {
		words = w + " " + words
	}
	p, ok := parseChange(words, today)
	if !ok {
		return Query{Command: CommandValue}, false
	}
	return Query{Command: CommandValue, Change: &p}, true
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
func showKindOf(word string) ShowKind {
	switch word {
	case wordAccount:
		return ShowAccount
	case DimSleeve:
		return ShowSleeve
	case wordInstrument:
		return ShowInstrument
	}
	return ShowAny
}

// parseAllocation returns the allocation command by the dimension word
// names, asset class when it names none; ok is false for a word that
// names no dimension.
func parseAllocation(word string) (Query, bool) {
	dims := map[string]string{
		"": DimAssetClass, "asset class": DimAssetClass, "asset classes": DimAssetClass, "asset": DimAssetClass,
		"class": DimAssetClass, DimCurrency: DimCurrency, "currencies": DimCurrency, DimInstitution: DimInstitution,
		"institutions": DimInstitution, "broker": DimInstitution, "brokers": DimInstitution, DimSleeve: DimSleeve,
		"sleeves": DimSleeve,
	}
	key, ok := dims[strings.TrimPrefix(word, "the ")]
	return Query{Command: CommandAllocation, Dimension: key}, ok
}

// parsePeriod returns the period words name, this year when they name
// none, today being today; ok is false for words that name no period.
func parsePeriod(words string, today kernel.Date) (Period, bool) {
	fixed := map[string]PeriodKind{
		"": PeriodYear, labelThisYear: PeriodYear, "ytd": PeriodYear, "year to date": PeriodYear, "year": PeriodYear,
		labelThisMonth: PeriodMonth, "month": PeriodMonth, "mtd": PeriodMonth,
		"last 12 months": PeriodTwelveMonths, "12 months": PeriodTwelveMonths, "1y": PeriodTwelveMonths,
		labelLastYear: PeriodLastYear, wordAll: PeriodAll, "all time": PeriodAll, "since tracking began": PeriodAll,
		"since the start": PeriodAll,
	}
	if k, ok := fixed[words]; ok {
		return Period{Kind: k}, true
	}
	if p, ok := parseBounded(words, today); ok {
		return p, true
	}
	if y := firstGroup(yearPattern, words, "y"); y != "" {
		d, err := kernel.ParseDate(y + "-01-01")
		return Period{Kind: PeriodCalendarYear, Year: d.Year()}, err == nil && !d.After(today)
	}
	if s := firstGroup(sincePattern, words, "d"); s != "" {
		d, ok := parseAskDate(s, today)
		return Period{Kind: PeriodSince, Since: d}, ok && !d.IsZero() && !d.After(today)
	}
	return Period{}, false
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
		return MonthStart(today).AddDays(-1), true
	case "end of last year":
		return YearStart(today.Year()).AddDays(-1), true
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

// MonthStart returns the first day of d's month.
func MonthStart(d kernel.Date) kernel.Date { return d.AddDays(1 - d.Day()) }

// YearStart returns 1 January of year; the zero date for a year a date
// can't have.
func YearStart(year int) kernel.Date {
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
func periodRange(p Period, to kernel.Date) span {
	switch p.Kind {
	case PeriodYear:
		return span{YearStart(to.Year()), to, labelThisYear}
	case PeriodMonth:
		return span{MonthStart(to), to, labelThisMonth}
	case PeriodTwelveMonths:
		return span{to.AddDays(1 - daysInYear), to, "over the last 12 months"}
	case PeriodLastYear:
		return span{YearStart(to.Year() - 1), YearStart(to.Year()).AddDays(-1), labelLastYear}
	case PeriodAll:
		return span{kernel.Date{}, to, "since tracking began"}
	case PeriodSince:
		return span{p.Since, to, "since " + p.Since.String()}
	case PeriodCalendarYear:
		end := YearStart(p.Year + 1).AddDays(-1)
		if end.After(to) {
			end = to
		}
		return span{YearStart(p.Year), end, "in " + YearStart(p.Year).String()[:4]}
	case PeriodMonthOf, PeriodBetween:
		return boundedSpan(p)
	}
	return span{YearStart(to.Year()), to, labelThisYear}
}

// daysInYear is how many days the last 12 months' opening day lies before its
// last day.
const daysInYear = 365

// The words of a command that name what it shows, in a question and in a
// tool's argument.
const (
	wordAccount    = "account"
	wordInstrument = "instrument"
	wordAll        = "all"
	// reasonMissing is what a basis the core leaves out shows.
	reasonMissing = "missing"
)
