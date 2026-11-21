package answer

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"connectrpc.com/connect"

	"github.com/arkadiuss/openfief/internal/kernel"
	viewsv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/views/v1"
)

// QuoteReader reads a market quote: the read API's GetMarketQuote, which asks
// the owner's market-data connector (D-226).
type QuoteReader interface {
	GetMarketQuote(ctx context.Context, req *connect.Request[viewsv1.GetMarketQuoteRequest]) (*connect.Response[viewsv1.GetMarketQuoteResponse], error)
}

// QuoteSpan is the days of a market quote's history, both inclusive: a zero
// day is left to the read API, which defaults to the last 30 days up to today.
type QuoteSpan struct {
	From, To kernel.Date
}

// quoteQuery is what a ticker or an ISIN looks like, as the read API takes it.
// Letters and digits with single dots or dashes between them, at most
// maxQuoteQueryLen: no leading or trailing dot or dash, no "..", no slash and
// no underscore.
var quoteQuery = regexp.MustCompile(`^[A-Za-z0-9]+(?:[.-][A-Za-z0-9]+)*$`)

// maxQuoteQueryLen is the longest query the read API takes.
const maxQuoteQueryLen = 32

// The names of the market price tool's periods, by how many days up to today
// each covers, and the one it has when the model gives none.
const (
	toolQuote     = "market_price"
	daysInWeek    = 7
	daysInMonth   = 30
	daysInQuarter = 90
	periodWeek    = "week"
	periodMonth   = "month"
	periodQuarter = "quarter"
	periodYear    = "year"
	argQuery      = "query"
)

// quotePeriodDays are the named periods' lengths in days.
func quotePeriodDays() map[string]int {
	return map[string]int{periodWeek: daysInWeek, periodMonth: daysInMonth, periodQuarter: daysInQuarter, periodYear: daysInYear}
}

// errQuoteQuery refuses a query that is no ticker or ISIN.
var errQuoteQuery = errors.New("the query is a ticker or an ISIN: at most 32 letters and digits, with single dots or dashes between them")

// quoteTool is the market price tool (D-226): the latest end-of-day price of
// any instrument, held or not, with its closes over a period, fetched through
// the owner's market-data connector. Its card is public data (T2, D-206) and
// names its source and the day of its latest close; unlike every other tool it
// reaches outside the owner's record, so the Ask box offers it only on an
// install with a connector, to a route whose models may receive T2.
func quoteTool() Tool {
	return Tool{
		Name: toolQuote,
		Description: "The latest end-of-day price of an instrument, held or not, with its closes over a period, from the owner's market-data " +
			"connector: public data, not the owner's record. Look an instrument up by its ticker, such as NVDA, or its ISIN. It gives the " +
			"price, its currency, the day of the close and the source, and the first, highest, lowest and latest close of the period. " +
			"When a ticker names several instruments it lists them with their ISINs, and you ask again with one's ISIN. Use it only for " +
			"the price of an instrument; the owner's own holdings, their values and a held instrument's last stored price come from the other tools.",
		Schema: `{"type":"object","properties":{"query":{"type":"string","description":"The instrument's ticker, such as NVDA, or its ISIN."},` +
			`"period":{"type":"string","enum":["week","month","quarter","year"],"description":"The days of history up to today: 7, 30, 90 or 365; month when omitted. ` +
			`Ignored when from or to is given."},` +
			`"from":{"type":"string","description":"The first day of the history, as YYYY-MM-DD."},` +
			`"to":{"type":"string","description":"The last day of the history, as YYYY-MM-DD; today at the latest."}},` +
			`"required":["query"],"additionalProperties":false}`,
		Label: "Market price", Tier: TierT2, Market: true,
		Defaults:   func(kernel.Date) map[string]any { return map[string]any{argPeriod: periodMonth} },
		ReplacedBy: map[string][]string{argPeriod: {argFrom, "to"}},
		query:      quoteArgs,
	}
}

// quoteArgs maps the market price tool's arguments to its command, today being
// today: two days, or one, when it gives them, else the days its period names.
func quoteArgs(args string, today kernel.Date) (Query, error) {
	var a struct {
		Query  string `json:"query"`
		Period string `json:"period"`
		From   string `json:"from"`
		To     string `json:"to"`
	}
	if err := decodeArgs(args, &a); err != nil {
		return Query{}, err
	}
	term := strings.TrimSpace(a.Query)
	if term == "" {
		return Query{}, errNoName
	}
	if len(term) > maxQuoteQueryLen || !quoteQuery.MatchString(term) {
		return Query{}, errQuoteQuery
	}
	span, err := quoteSpan(a.Period, a.From, a.To, today)
	if err != nil {
		return Query{}, err
	}
	return Query{Command: CommandQuote, Term: term, Quote: span}, nil
}

// quoteSpan returns the days a market price call asks for.
func quoteSpan(period, from, to string, today kernel.Date) (QuoteSpan, error) {
	if from != "" || to != "" {
		return explicitSpan(from, to, today)
	}
	if period == "" {
		period = periodMonth
	}
	n, ok := quotePeriodDays()[period]
	if !ok {
		return QuoteSpan{}, errBadEnum
	}
	return QuoteSpan{From: today.AddDays(1 - n), To: today}, nil
}

// explicitSpan returns the days a call gives, either of which may be left out.
func explicitSpan(from, to string, today kernel.Date) (QuoteSpan, error) {
	first, okFrom := parseAskDate(from, today)
	last, okTo := parseAskDate(to, today)
	if !okFrom || !okTo || (from != "" && first.IsZero()) || (to != "" && last.IsZero()) {
		return QuoteSpan{}, errToolDate
	}
	return QuoteSpan{From: first, To: last}, nil
}

// quote answers a market price command: the instrument its term names, its
// latest close and the closes of its span, asked of the read API, which asks
// the owner's market-data connector (D-226). A failure the person or the model
// can act on, such as no market-data connector, a ticker nothing is listed
// under or a source that did not answer, is a card that says so; any other is
// an error.
func (cm *Commands) quote(ctx context.Context, c Caller, q *Query) (Card, error) {
	req := &viewsv1.GetMarketQuoteRequest{Query: q.Term, From: OptionalString(q.Quote.From), To: OptionalString(q.Quote.To)}
	resp, err := cm.read.GetMarketQuote(ctx, bearer(req, c))
	if err != nil {
		if card, ok := quoteFailure(q.Term, err); ok {
			return card, nil
		}
		return Card{}, fmt.Errorf("GetMarketQuote: %w", err)
	}
	if mq := resp.Msg.GetQuote(); mq != nil {
		return quoteCard(mq, c.Format), nil
	}
	return quoteChoices(q.Term, resp.Msg), nil
}

// quoteFailure turns the read API's refusal of a market quote into the card
// that says why, and false for a refusal that is an error: the token, the
// arguments the tool checked, or a failure of the core's own.
func quoteFailure(term string, err error) (Card, bool) {
	code := connect.CodeOf(err)
	if code == connect.CodeNotFound {
		return Card{
			Title: "No instrument called “" + term + "” at the market-data source",
			Text:  "The owner's market-data connector lists nothing under that ticker or ISIN. Try its ISIN, or another ticker.",
		}, true
	}
	if code == connect.CodeFailedPrecondition {
		return Card{Title: "Market prices are not available here", Text: "No market-data connector is installed and approved on this install, or it answers from its own data alone."}, true
	}
	if code == connect.CodeUnavailable || code == connect.CodeDeadlineExceeded {
		return Card{Title: "The market-data source did not answer", Text: "The owner's market-data connector or its provider could not be reached. Try again in a moment."}, true
	}
	if code == connect.CodeResourceExhausted {
		return Card{Title: "Too many price lookups in a minute", Text: "The owner's market-data key is spared: try again in a minute."}, true
	}
	return Card{}, false
}

// quoteCard is the price card of a quote: the instrument, its latest price,
// the day of its close and the source; rows hold the first, highest, lowest and
// latest close of the period. It is public market data (T2), which the text
// says, with whose connector it came from.
func quoteCard(q *viewsv1.MarketQuote, f NumberFormat) Card {
	in := q.GetListing().GetInstrument()
	name := strings.Join(nonEmpty(in.GetName(), in.GetSymbol()), " · ")
	card := Card{
		Title: name + ": " + f.Money(q.GetLatest()),
		Text: "Latest close on " + q.GetDate() + " from " + q.GetSource() + ", the owner's market-data connector. " +
			strings.Join(nonEmpty(isinWords(in.GetIsin()), q.GetListing().GetExchange()), ", ") + ". Public market data, not the owner's record.",
	}
	history := q.GetHistory()
	if len(history) == 0 {
		return card
	}
	high, low := history[0], history[0]
	for _, p := range history {
		if decimalOf(p.GetClose()).Cmp(decimalOf(high.GetClose())) > 0 {
			high = p
		}
		if decimalOf(p.GetClose()).Cmp(decimalOf(low.GetClose())) < 0 {
			low = p
		}
	}
	row := func(label string, p *viewsv1.MarketQuotePoint) Row {
		return Row{Label: label, Detail: p.GetDate(), Figure: f.Money(p.GetClose())}
	}
	card.Rows = []Row{
		row("First close of the period", history[0]), row("Highest close", high), row("Lowest close", low), row("Latest close", history[len(history)-1]),
	}
	return card
}

// isinWords writes an ISIN for a card's text, or nothing for none.
func isinWords(isin string) string {
	if isin == "" {
		return ""
	}
	return "ISIN " + isin
}

// decimalOf is the decimal of a money, zero when it isn't one.
func decimalOf(m *viewsv1.Money) kernel.Decimal {
	d, err := kernel.ParseDecimal(m.GetAmount())
	if err != nil {
		return kernel.Decimal{}
	}
	return d
}

// quoteChoices is the card of a ticker that names several instruments: they
// are listed with their ISINs, none is priced, and the model asks again with
// one's ISIN, never guessing which the person means.
func quoteChoices(term string, r *viewsv1.GetMarketQuoteResponse) Card {
	card := Card{
		Title: "“" + term + "” names " + Plural(len(r.GetCandidates()), "instrument", "instruments"),
		Text:  "None was priced. Ask which one is meant, and look it up again by its ISIN.",
	}
	for i, l := range r.GetCandidates() {
		if i == answerRows {
			break
		}
		in := l.GetInstrument()
		card.Rows = append(card.Rows, Row{
			Label: strings.Join(nonEmpty(in.GetName(), in.GetSymbol()), " · "), Detail: strings.Join(nonEmpty(isinWords(in.GetIsin()), l.GetExchange()), ", "),
		})
	}
	return card
}
