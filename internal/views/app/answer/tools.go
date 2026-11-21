package answer

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/arkadiuss/openfief/internal/kernel"
)

// The data tiers of D-206, as the core's contract spells them. The Ask box
// tags every message it sends a model with one; the core's router decides
// which models may receive it. A tool's Tier is the tier of its card.
const (
	TierT0 = "T0"
	TierTF = "TF"
	TierT1 = "T1"
	TierT2 = "T2"
)

// Tool is one tool the Ask box's model may call (D-203): one of the fixed
// commands' answerers, read-only like them. A tool never changes anything,
// and no tool that does exists: this table is the whole of what a model can
// do.
type Tool struct {
	// Name and Description are what the model is told.
	Name, Description string
	// Schema is a JSON Schema of the tool's arguments.
	Schema string
	// Label names the tool's card for the person, in a notice that the card
	// was withheld.
	Label string
	// Tier is the data tier of what the tool's card carries (D-206).
	Tier string
	// Market marks a tool whose card is public market data fetched through
	// the owner's market-data connector (D-226): unlike the others it reaches
	// outside the owner's record, so a caller offers it only where a connector
	// answers and where its data may go to the model (Tier).
	Market bool
	// Defaults returns what leaving an argument out means, the value it has
	// when the model does not give it, for each argument that has one, latest
	// being the latest valued day. A call that gives an argument its default
	// says what the call without it says: the one source for a harness that
	// compares calls (AskArgDefaults), and the descriptions below say the same
	// in words. Nil when no argument has a default.
	Defaults func(latest kernel.Date) map[string]any
	// ReplacedBy names, for an argument that has a default, the arguments that
	// say the same thing another way: a call that gives one of them has no use
	// for the default, so a harness that fills defaults in leaves it out
	// (AskArgReplacedBy). Nil when none does.
	ReplacedBy map[string][]string
	// query turns the model's arguments, a JSON object text, into the
	// command the answerer runs, today being today.
	query func(args string, today kernel.Date) (Query, error)
}

// Query turns the model's arguments, a JSON object text, into the command the
// answerer runs, today being today.
func (t *Tool) Query(args string, today kernel.Date) (Query, error) {
	return t.query(args, today)
}

// The names of the arguments more than one tool or table spells.
const (
	argPeriod = "period"
	argFrom   = "from"
)

// Tools returns the tools, in the order the model is told them. Every
// tool's card is T0 but the last: it holds whole-portfolio holdings, amounts or
// tax data, or names the owner's accounts and instruments (D-206). The market
// price tool's card holds only public data, so it takes T2 (D-226).
func Tools() []Tool {
	return []Tool{needsTool(), showTool(), syncTool(), allocationTool(), performanceTool(), valueTool(), quoteTool()}
}

func needsTool() Tool {
	return Tool{
		Name:        "what_needs_me",
		Description: "The open attention items: what needs the owner's attention, the top of the Inbox, with how many are open.",
		Schema:      `{"type":"object","properties":{},"additionalProperties":false}`,
		Label:       "What needs you", Tier: TierT0,
		query: func(args string, _ kernel.Date) (Query, error) {
			var a struct{}
			if err := decodeArgs(args, &a); err != nil {
				return Query{}, err
			}
			return Query{Command: CommandNeeds}, nil
		},
	}
}

func showTool() Tool {
	return Tool{
		Name: "show",
		Description: "Shows one account, sleeve or instrument by its name, ticker or ISIN: its value and largest holdings, its accounts, " +
			"or who holds it and its last price.",
		Schema: `{"type":"object","properties":{"name":{"type":"string","description":"The account's, sleeve's or instrument's name, ticker or ISIN."},` +
			`"kind":{"type":"string","enum":["any","account","sleeve","instrument"],"description":"What the name is; any when unsure."}},` +
			`"required":["name"],"additionalProperties":false}`,
		Label: "Account, sleeve or instrument", Tier: TierT0,
		query: showQuery,
	}
}

func syncTool() Tool {
	return Tool{
		Name:        "sync_status",
		Description: "Why a connector's sync fails: its status and latest attempts, each with its error class.",
		Schema: `{"type":"object","properties":{"connector":{"type":"string","description":"The connector instance's name, such as xtb-main."}},` +
			`"required":["connector"],"additionalProperties":false}`,
		Label: "Sync status", Tier: TierT0,
		query: func(args string, _ kernel.Date) (Query, error) {
			var a struct {
				Connector string `json:"connector"`
			}
			if err := decodeArgs(args, &a); err != nil {
				return Query{}, err
			}
			if strings.TrimSpace(a.Connector) == "" {
				return Query{}, errNoName
			}
			return Query{Command: CommandSync, Term: strings.TrimSpace(a.Connector)}, nil
		},
	}
}

func allocationTool() Tool {
	return Tool{
		Name:        "allocation",
		Description: "The portfolio's allocation, gross, on the latest valued day, grouped by asset class, currency, institution or sleeve.",
		Schema: `{"type":"object","properties":{"by":{"type":"string","enum":["asset_class","currency","institution","sleeve"],` +
			`"description":"The grouping; asset_class when omitted."}},"additionalProperties":false}`,
		Label: "Allocation", Tier: TierT0,
		Defaults: func(kernel.Date) map[string]any { return map[string]any{"by": "asset_class"} },
		query:    allocationQuery,
	}
}

func performanceTool() Tool {
	return Tool{
		Name: "performance",
		Description: "Answers questions about return and performance over a period (in Polish, \"stopa zwrotu\" and \"wynik\"): the time-weighted " +
			"return, with the portfolio's value at its start and end and the deposits and withdrawals in it. A period counts its first day " +
			"in full: it runs from the end of the day before it to the end of its last day. Deposits, withdrawals and confirmed transfers " +
			"don't count as return. A period is a calendar month, a calendar year, a first and a last day, a first day onward or a named period.",
		Schema: `{"type":"object","properties":{"period":{"type":"string","enum":["this_year","this_month","last_12_months","last_year","all"],` +
			`"description":"A named period; this_year when omitted. Ignored when from and to, month, year or since is given."},` +
			`"from":{"type":"string","description":"The first day of a period, as YYYY-MM-DD; given with to."},` +
			`"to":{"type":"string","description":"The last day of a period, as YYYY-MM-DD; given with from."},` +
			`"month":{"type":"string","description":"A calendar month, as YYYY-MM."},` +
			`"year":{"type":"integer","description":"A calendar year, such as 2025."},` +
			`"since":{"type":"string","description":"A first day, as YYYY-MM-DD."}},"additionalProperties":false}`,
		Label: "Performance", Tier: TierT0,
		Defaults:   func(kernel.Date) map[string]any { return map[string]any{argPeriod: "this_year"} },
		ReplacedBy: map[string][]string{argPeriod: {argFrom, "to", "month", "year", "since"}},
		query:      performanceQuery,
	}
}

func valueTool() Tool {
	return Tool{
		Name: "value",
		Description: "The portfolio's gross, net and liquidation value on a day, in the reporting currency; the latest valued day when " +
			"no date is given. With a period instead, the three values at its start and end, their change in money and in percent, and " +
			"the deposits and withdrawals in it.",
		Schema: `{"type":"object","properties":{"date":{"type":"string","description":"The day, as YYYY-MM-DD; omitted for the latest valued day. Not given with period."},` +
			`"period":{"type":"object","description":"A period of days, both inclusive, instead of a day.","properties":` +
			`{"from":{"type":"string","description":"The first day, as YYYY-MM-DD."},"to":{"type":"string","description":"The last day, as YYYY-MM-DD."}},` +
			`"required":["from","to"],"additionalProperties":false}},"additionalProperties":false}`,
		Label: "Value", Tier: TierT0,
		Defaults: func(latest kernel.Date) map[string]any {
			if latest.IsZero() {
				return nil
			}
			return map[string]any{"date": latest.String()}
		},
		ReplacedBy: map[string][]string{"date": {argPeriod}},
		query:      valueQuery,
	}
}

// Why a model's arguments to a tool were not accepted. The text of an
// argument is never in the error: it is the person's data (GR-12), and the
// model gets the error back as the tool's result.
var (
	errBadArgs  = errors.New("the arguments are not a JSON object of the tool's schema")
	errNoName   = errors.New("a name is needed")
	errBadEnum  = errors.New("the argument is none of the values the schema lists")
	errToolDate = errors.New("the date is not a valid day, YYYY-MM-DD")
	// errPeriodBounds and errDateAndPeriod refuse a period without both of
	// its days, and a value asked on a day and over a period at once.
	errPeriodBounds  = errors.New("a period needs both a first day, from, and a last day, to")
	errDateAndPeriod = errors.New("a value is asked on a date or over a period, not both")
)

// decodeArgs reads args, a JSON object text, into dst, failing on a field the
// tool does not have. Empty arguments are the empty object.
func decodeArgs(args string, dst any) error {
	if strings.TrimSpace(args) == "" {
		args = "{}"
	}
	dec := json.NewDecoder(strings.NewReader(args))
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		return errBadArgs
	}
	if dec.More() {
		return errBadArgs
	}
	return nil
}

// showQuery maps the show tool's arguments to its command.
func showQuery(args string, _ kernel.Date) (Query, error) {
	var a struct {
		Name string `json:"name"`
		Kind string `json:"kind"`
	}
	if err := decodeArgs(args, &a); err != nil {
		return Query{}, err
	}
	if strings.TrimSpace(a.Name) == "" {
		return Query{}, errNoName
	}
	kinds := map[string]ShowKind{"": ShowAny, "any": ShowAny, wordAccount: ShowAccount, "sleeve": ShowSleeve, "instrument": ShowInstrument}
	kind, ok := kinds[a.Kind]
	if !ok {
		return Query{}, errBadEnum
	}
	return Query{Command: CommandShow, Term: strings.TrimSpace(a.Name), Show: kind}, nil
}

// allocationQuery maps the allocation tool's arguments to its command.
func allocationQuery(args string, _ kernel.Date) (Query, error) {
	var a struct {
		By string `json:"by"`
	}
	if err := decodeArgs(args, &a); err != nil {
		return Query{}, err
	}
	words := map[string]string{"": "", "asset_class": "asset class", "currency": DimCurrency, "institution": DimInstitution, "sleeve": DimSleeve}
	word, ok := words[a.By]
	if !ok {
		return Query{}, errBadEnum
	}
	return parseAllocationWord(word)
}

// parseAllocationWord is parseAllocation as an error.
func parseAllocationWord(word string) (Query, error) {
	q, ok := parseAllocation(word)
	if !ok {
		return Query{}, errBadEnum
	}
	return q, nil
}

// performanceArgs are the performance tool's arguments.
type performanceArgs struct {
	Period string `json:"period"`
	Year   int    `json:"year"`
	Since  string `json:"since"`
	Month  string `json:"month"`
	From   string `json:"from"`
	To     string `json:"to"`
}

// performanceQuery maps the performance tool's arguments to its command. Two
// days, a month, a year or a first day name the period, in that order; else
// the period's own word.
func performanceQuery(args string, today kernel.Date) (Query, error) {
	var a performanceArgs
	if err := decodeArgs(args, &a); err != nil {
		return Query{}, err
	}
	p, err := a.period(today)
	if err != nil {
		return Query{}, err
	}
	return Query{Command: CommandPerformance, Period: p}, nil
}

// period returns the period the arguments name, today being today.
func (a *performanceArgs) period(today kernel.Date) (Period, error) {
	switch {
	case a.From != "" || a.To != "":
		return boundedArgs(a.From, a.To, today)
	case a.Month != "":
		p, ok := parseBounded(strings.ToLower(strings.TrimSpace(a.Month)), today)
		if !ok || p.Kind != PeriodMonthOf {
			return Period{}, errToolDate
		}
		return p, nil
	}
	word, err := a.word()
	if err != nil {
		return Period{}, err
	}
	p, valid := parsePeriod(word, today)
	if !valid {
		return Period{}, errToolDate
	}
	return p, nil
}

// word returns the words of the year, the first day or the named period the
// arguments give, in that order.
func (a *performanceArgs) word() (string, error) {
	switch {
	case a.Year != 0:
		return strconv.Itoa(a.Year), nil
	case a.Since != "":
		return "since " + a.Since, nil
	}
	words := map[string]string{
		"": "", "this_year": labelThisYear, "this_month": labelThisMonth, "last_12_months": "last 12 months", "last_year": labelLastYear, wordAll: wordAll,
	}
	word, ok := words[a.Period]
	if !ok {
		return "", errBadEnum
	}
	return word, nil
}

// boundedArgs returns the period of two days a tool's arguments give, today
// being today. The days are not put in order or checked against the valued
// days here: the answer says when they can't be answered.
func boundedArgs(from, to string, today kernel.Date) (Period, error) {
	if from == "" || to == "" {
		return Period{}, errPeriodBounds
	}
	first, okFrom := parseAskDate(from, today)
	last, okTo := parseAskDate(to, today)
	if !okFrom || !okTo || first.IsZero() || last.IsZero() {
		return Period{}, errToolDate
	}
	return Period{Kind: PeriodBetween, Since: first, To: last}, nil
}

// valueQuery maps the value tool's arguments to its command: the value on a
// day, or its change over a period of two days.
func valueQuery(args string, today kernel.Date) (Query, error) {
	var a struct {
		Date   string `json:"date"`
		Period *struct {
			From string `json:"from"`
			To   string `json:"to"`
		} `json:"period"`
	}
	if err := decodeArgs(args, &a); err != nil {
		return Query{}, err
	}
	if a.Period != nil {
		if a.Date != "" {
			return Query{}, errDateAndPeriod
		}
		p, err := boundedArgs(a.Period.From, a.Period.To, today)
		if err != nil {
			return Query{}, err
		}
		return Query{Command: CommandValue, Change: &p}, nil
	}
	d, ok := parseAskDate(a.Date, today)
	if !ok || d.After(today) {
		return Query{}, errToolDate
	}
	return Query{Command: CommandValue, Date: d}, nil
}

// toolResult is a card as the model receives it: the card's data, with
// nothing the page needs to draw it, such as links, or the question.
type toolResult struct {
	// ID is what the model names the card by in its answer.
	ID    string          `json:"id"`
	Title string          `json:"title"`
	Text  string          `json:"text,omitempty"`
	Rows  []toolResultRow `json:"rows,omitempty"`
}

// toolResultRow is a row of a card as the model receives it.
type toolResultRow struct {
	Label  string `json:"label"`
	Detail string `json:"detail,omitempty"`
	Figure string `json:"figure,omitempty"`
}

// CardResult returns card as the JSON text of the tool's result, with the id
// the model names it by.
func CardResult(card *Card, id string) string {
	res := toolResult{ID: id, Title: card.Title, Text: card.Text}
	for _, r := range card.Rows {
		res.Rows = append(res.Rows, toolResultRow{Label: r.Label, Detail: r.Detail, Figure: r.Figure})
	}
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(res); err != nil { // strings and slices of strings always encode
		return `{"title":"the card could not be written"}`
	}
	return strings.TrimSpace(buf.String())
}

// ToolFailure is the text of a tool's result that holds no data: a refusal or
// a failure, which the model reads and the person never needs to.
func ToolFailure(format string, args ...any) string {
	out, err := json.Marshal(map[string]string{"error": fmt.Sprintf(format, args...)})
	if err != nil { // a map of strings always encodes
		return `{"error":"failed"}`
	}
	return string(out)
}

// CardID is the id a model reads on the n-th card (counting from one) its
// tools produced.
func CardID(n int) string { return "c" + strconv.Itoa(n) }
