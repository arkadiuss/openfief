// Package chartcheck works out, for the end-to-end suite's check of D-200's
// criterion 8, what the built-in UI's charts must draw when it draws them
// from the read API alone, and which calls the UI made to the core while it
// drew them. From read API answers it builds each chart's expected days and
// lines: the value over time with its three values (D-77, D-106), the
// time-weighted return (D-61), an instrument's prices per listing (D-195) and
// the allocation's bars (D-30). It reads the amounts the page wrote in the
// owner's number format back as the read API's exact decimals (D-113), and
// picks out of the core's JSON log the calls a plugin made outside the read
// API in a span of time. It only reads messages, strings and log bytes; the
// suite (test/e2e, charts_test.go) loads the pages and makes the calls. It
// is never released.
package chartcheck

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/arkadiuss/openfief/internal/kernel"
	viewsv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/views/v1"
)

// The UI's limits on what a chart asks the read API for: the value chart's
// longest span and the instrument chart's days of prices.
const (
	MaxChartDays        = 3650
	InstrumentPriceDays = 365
)

// The share of a bar: percent with one decimal place.
const (
	sharePlaces = 1
	hundred     = 100
)

// The core's log messages for the calls this package counts: one per read
// API request, and one per call of a service only plugins reach, which
// names its procedure and the calling plugin.
const (
	readRequestMsg = "read API request"
	pluginCallMsg  = "auth call"
)

// narrowSpace is how the UI writes a space group separator.
const narrowSpace = " "

var (
	// ErrBadDate reports a date that is not YYYY-MM-DD.
	ErrBadDate = errors.New("bad date")
	// ErrBadAmount reports an amount the page wrote that is no decimal in
	// the number format, followed by a currency.
	ErrBadAmount = errors.New("bad amount")
	// ErrBadLog reports a log line that is JSON but has no usable time.
	ErrBadLog = errors.New("bad log line")
)

// Chart is what a line chart draws: its days, and per line the value on
// each day, "" for a day the line has none, in the order of the lines.
type Chart struct {
	Labels []string   `json:"labels"`
	Series [][]string `json:"series"`
}

// Bar is one bar of the allocation as the page shows it: the group's key,
// its share of the total in percent, the bar's width and the group's value
// as "amount currency", the amount as the read API's exact decimal.
type Bar struct {
	Key   string `json:"key"`
	Share string `json:"share"`
	Width string `json:"width"`
	Value string `json:"value"`
}

// ValueKinds are the value chart's three lines in the order the UI draws
// them when it draws all three: gross, net, liquidation (D-77).
func ValueKinds() []viewsv1.ValueKind {
	return []viewsv1.ValueKind{
		viewsv1.ValueKind_VALUE_KIND_GROSS, viewsv1.ValueKind_VALUE_KIND_NET, viewsv1.ValueKind_VALUE_KIND_LIQUIDATION,
	}
}

// ChartFrom returns the first day the UI asks the read API's totals for,
// on a value chart up to date: the performance's first day perfFrom, as
// GetPerformance answered it, unless it is empty, no date or after date,
// when it is date itself; never more than MaxChartDays before date.
func ChartFrom(perfFrom, date string) (string, error) {
	to, err := kernel.ParseDate(date)
	if err != nil {
		return "", fmt.Errorf("%w: %q: %w", ErrBadDate, date, err)
	}
	from := to
	if first, parseErr := kernel.ParseDate(perfFrom); parseErr == nil && !first.After(to) {
		from = first
	}
	if earliest := to.AddDays(-MaxChartDays); from.Before(earliest) {
		from = earliest
	}
	return from.String(), nil
}

// DaysBefore returns the day n days before date.
func DaysBefore(date string, n int) (string, error) {
	d, err := kernel.ParseDate(date)
	if err != nil {
		return "", fmt.Errorf("%w: %q: %w", ErrBadDate, date, err)
	}
	return d.AddDays(-n).String(), nil
}

// ValueChart is the value chart of totals, the answer of
// ListValuationTotals or ListAccountValuations: a day per total and a line
// per kind, each day's amount of that kind, "" where the total has none.
func ValueChart(totals []*viewsv1.DateTotal, kinds []viewsv1.ValueKind) Chart {
	c := Chart{Labels: []string{}, Series: make([][]string, len(kinds))}
	for i := range kinds {
		c.Series[i] = []string{}
	}
	for _, t := range totals {
		c.Labels = append(c.Labels, t.GetDate())
		for i, k := range kinds {
			c.Series[i] = append(c.Series[i], totalOf(t, k).GetAmount())
		}
	}
	return c
}

// totalOf returns t's money of kind k; gross for an unspecified kind.
func totalOf(t *viewsv1.DateTotal, k viewsv1.ValueKind) *viewsv1.Money {
	switch k {
	case viewsv1.ValueKind_VALUE_KIND_NET:
		return t.GetNetReporting()
	case viewsv1.ValueKind_VALUE_KIND_LIQUIDATION:
		return t.GetLiquidationReporting()
	case viewsv1.ValueKind_VALUE_KIND_GROSS, viewsv1.ValueKind_VALUE_KIND_UNSPECIFIED:
	}
	return t.GetValueReporting()
}

// PerformanceChart is the time-weighted return chart of GetPerformance's
// answer without the model portfolio: a day per index, one line.
func PerformanceChart(p *viewsv1.GetPerformanceResponse) Chart {
	c := Chart{Labels: []string{}, Series: [][]string{{}}}
	for _, d := range p.GetDays() {
		c.Labels = append(c.Labels, d.GetDate())
		c.Series[0] = append(c.Series[0], d.GetIndex())
	}
	return c
}

// PriceCharts are an instrument's price charts from ListInstrumentPrices'
// answer, one per listing in the order the listings first appear: the
// days any source has a close, sorted, and a line per source, the primary
// first, then the others by name, "" on a day the source has none. Without
// a source marked primary, the first by name is.
func PriceCharts(prices []*viewsv1.InstrumentPrice) []Chart {
	bySymbol := map[string][]*viewsv1.InstrumentPrice{}
	var symbols []string
	for _, p := range prices {
		s := p.GetInstrument().GetSymbol()
		if _, ok := bySymbol[s]; !ok {
			symbols = append(symbols, s)
		}
		bySymbol[s] = append(bySymbol[s], p)
	}
	out := make([]Chart, 0, len(symbols))
	for _, s := range symbols {
		out = append(out, listingChart(bySymbol[s]))
	}
	return out
}

// listingChart is the price chart of one listing's prices.
func listingChart(prices []*viewsv1.InstrumentPrice) Chart {
	closes := map[string]map[string]string{}
	var dates, sources []string
	primary := ""
	for _, p := range prices {
		src := p.GetSource()
		if closes[src] == nil {
			closes[src] = map[string]string{}
			sources = append(sources, src)
		}
		closes[src][p.GetDate()] = p.GetClose().GetAmount()
		dates = append(dates, p.GetDate())
		if p.GetPrimary() {
			primary = src
		}
	}
	slices.Sort(dates)
	dates = slices.Compact(dates)
	slices.Sort(sources)
	if primary == "" && len(sources) > 0 {
		primary = sources[0]
	}
	ordered := []string{primary}
	for _, s := range sources {
		if s != primary {
			ordered = append(ordered, s)
		}
	}
	c := Chart{Labels: append([]string{}, dates...), Series: make([][]string, 0, len(ordered))}
	for _, src := range ordered {
		line := make([]string, len(dates))
		for i, d := range dates {
			line[i] = closes[src][d]
		}
		c.Series = append(c.Series, line)
	}
	return c
}

// Allocation is the allocation's bars from GetAllocation's answer: a bar
// per group in the answer's order, its value in the reporting currency,
// its share of the total in percent, rounded half-even to one place, and
// its width the share's number. A share of a zero total is empty.
func Allocation(resp *viewsv1.GetAllocationResponse) ([]Bar, error) {
	total := resp.GetTotalReporting().GetAmount()
	bars := make([]Bar, 0, len(resp.GetGroups()))
	for _, g := range resp.GetGroups() {
		v := g.GetValueReporting()
		share, err := sharePercent(v.GetAmount(), total)
		if err != nil {
			return nil, fmt.Errorf("group %s: %w", g.GetKey(), err)
		}
		bar := Bar{Key: g.GetKey(), Value: v.GetAmount() + " " + v.GetCurrency(), Width: share}
		if share != "" {
			bar.Share = share + "%"
		}
		bars = append(bars, bar)
	}
	return bars, nil
}

// sharePercent returns part as a percentage of total, rounded half-even to
// one place, or "" when total is zero.
func sharePercent(part, total string) (string, error) {
	p, err := kernel.ParseDecimal(part)
	if err != nil {
		return "", fmt.Errorf("%w: %q: %w", ErrBadAmount, part, err)
	}
	t, err := kernel.ParseDecimal(total)
	if err != nil {
		return "", fmt.Errorf("%w: total %q: %w", ErrBadAmount, total, err)
	}
	if t.IsZero() {
		return "", nil
	}
	q, err := p.Mul(kernel.NewDecimal(hundred, 0)).Quo(t, sharePlaces, kernel.RoundHalfEven)
	if err != nil {
		return "", fmt.Errorf("share of %s in %s: %w", part, total, err)
	}
	return q.String(), nil
}

// PlainMoney reads money the page wrote in number format f, such as
// "1 234,50 PLN" with a space group and a decimal comma, back as the read
// API sends it, "1234.50 PLN": its amount's group separators dropped, its
// decimal separator a point, its currency kept. A space group is written
// as the narrow no-break space (D-113).
func PlainMoney(shown string, f *viewsv1.NumberFormat) (string, error) {
	amount, currency, ok := strings.Cut(strings.TrimSpace(shown), " ")
	if !ok || currency == "" || strings.Contains(currency, " ") {
		return "", fmt.Errorf("%w: %q has no amount and currency", ErrBadAmount, shown)
	}
	group, decimal := f.GetGroupSeparator(), f.GetDecimalSeparator()
	if decimal == "" {
		group, decimal = narrowSpace, "."
	}
	if group == " " {
		group = narrowSpace
	}
	if group != "" {
		amount = strings.ReplaceAll(amount, group, "")
	}
	amount = strings.Replace(amount, decimal, ".", 1)
	if _, err := kernel.ParseDecimal(amount); err != nil {
		return "", fmt.Errorf("%w: %q: %w", ErrBadAmount, shown, err)
	}
	return amount + " " + currency, nil
}

// Call is one line of the core's log this package counts: when it was
// logged, its message, and for a plugin's call its procedure and plugin.
type Call struct {
	Time      time.Time
	Msg       string
	Procedure string
	Plugin    string
}

// logLine is the part of the core's JSON log line Call reads.
type logLine struct {
	Time      string `json:"time"`
	Msg       string `json:"msg"`
	Procedure string `json:"procedure"`
	Plugin    string `json:"plugin"`
}

// Calls reads the core's log, one JSON object per line as log/slog's
// JSON handler writes it, and returns its read API requests and plugin
// calls in order. Lines that are no JSON object, such as a runtime's
// panic, and other messages are skipped.
func Calls(logs []byte) ([]Call, error) {
	var calls []Call
	s := bufio.NewScanner(bytes.NewReader(logs))
	s.Buffer(make([]byte, 0, bufio.MaxScanTokenSize), len(logs)+1)
	for s.Scan() {
		var l logLine
		if json.Unmarshal(s.Bytes(), &l) != nil || (l.Msg != readRequestMsg && l.Msg != pluginCallMsg) {
			continue
		}
		at, err := time.Parse(time.RFC3339Nano, l.Time)
		if err != nil {
			return nil, fmt.Errorf("%w: %q: %w", ErrBadLog, s.Text(), err)
		}
		calls = append(calls, Call{Time: at, Msg: l.Msg, Procedure: l.Procedure, Plugin: l.Plugin})
	}
	if err := s.Err(); err != nil {
		return nil, fmt.Errorf("reading the core's log: %w", err)
	}
	return calls, nil
}

// Window is a span of time, both ends included.
type Window struct {
	From, To time.Time
}

// contains reports whether at is in w.
func (w Window) contains(at time.Time) bool { return !at.Before(w.From) && !at.After(w.To) }

// PluginCalls returns the procedures plugin called in w outside the read
// API, sorted, each once.
func PluginCalls(calls []Call, plugin string, w Window) []string {
	var procs []string
	for _, c := range calls {
		if c.Msg == pluginCallMsg && c.Plugin == plugin && w.contains(c.Time) {
			procs = append(procs, c.Procedure)
		}
	}
	slices.Sort(procs)
	return slices.Compact(procs)
}

// ReadRequests counts the read API requests in w, from any caller.
func ReadRequests(calls []Call, w Window) int {
	n := 0
	for _, c := range calls {
		if c.Msg == readRequestMsg && w.contains(c.Time) {
			n++
		}
	}
	return n
}

// Disallowed returns the procedures that allowed doesn't cover, in order.
// An entry of allowed that ends with "/" covers every procedure of that
// service, any other only the procedure it names.
func Disallowed(procedures, allowed []string) []string {
	var out []string
	for _, p := range procedures {
		ok := slices.ContainsFunc(allowed, func(a string) bool {
			return p == a || (strings.HasSuffix(a, "/") && strings.HasPrefix(p, a))
		})
		if !ok {
			out = append(out, p)
		}
	}
	return out
}
