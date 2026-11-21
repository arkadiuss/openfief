package ingest

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"unicode"

	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
)

// Errors of a quote lookup (D-226).
var (
	// ErrInvalidQuoteQuery reports a query that is no ticker or ISIN.
	ErrInvalidQuoteQuery = errors.New("a quote query is a ticker or an ISIN")
	// ErrInvalidQuote reports a quote a connector answered that does not hold:
	// no closes or too many, a close out of order, dated after the period, not
	// positive or in another currency than its listing.
	ErrInvalidQuote = errors.New("the quote does not hold")
	// ErrInvalidQuotePeriod reports a period of days that is out of order,
	// ends after today or is too long.
	ErrInvalidQuotePeriod = errors.New("a quote period is at most a year of days, ending today or earlier")
)

// The limits of a quote (D-226): how many days of history, by default and at
// most, how long a query is and how long a name is when the core passes on
// what a market-data source says.
const (
	DefaultQuoteDays = 30
	MaxQuoteDays     = 366
	MaxQuoteQuery    = 32
	// MaxQuoteName is the longest name a quote carries; the source's own text
	// is cut to it.
	MaxQuoteName = 200
	// MaxQuoteCandidates is the most listings a quote that names several
	// carries.
	MaxQuoteCandidates = 10
)

// quoteQuery is what a ticker or an ISIN asked for looks like: letters and
// digits, with single dots or dashes between them, so that no query is a path
// segment such as "." or "..", or holds a slash, when a connector puts it in a
// provider's URL.
func quoteQuery() *regexp.Regexp { return regexp.MustCompile(`^[A-Za-z0-9]+(?:[.-][A-Za-z0-9]+)*$`) }

// quoteSymbol is what a listing's symbol a connector answers with looks like:
// letters, digits, dots, dashes and underscores.
func quoteSymbol() *regexp.Regexp { return regexp.MustCompile(`^[A-Za-z0-9._-]+$`) }

// QuoteRequest asks a market-data connector for one instrument's latest price
// and history (D-226): a ticker or an ISIN, and the days of the history, both
// inclusive.
type QuoteRequest struct {
	Query    string
	From, To kernel.Date
}

// ParseQuoteQuery returns query without its surrounding space, or an error
// wrapping ErrInvalidQuoteQuery when what is left is not 1 to MaxQuoteQuery
// letters and digits with single dots or dashes between them: no leading or
// trailing dot or dash, no "..", no slash and no underscore. It never quotes
// the query.
func ParseQuoteQuery(query string) (string, error) {
	return parseQuoteText(query, quoteQuery())
}

// ParseQuoteSymbol returns the symbol of a listing a connector answered with,
// without its surrounding space, or an error wrapping ErrInvalidQuoteQuery when
// what is left is not 1 to MaxQuoteQuery letters, digits, dots, dashes and
// underscores. A symbol is the provider's own and the core sends it nowhere, so
// it keeps the wider set. It never quotes the symbol.
func ParseQuoteSymbol(symbol string) (string, error) {
	return parseQuoteText(symbol, quoteSymbol())
}

// parseQuoteText trims text and checks it against pattern and MaxQuoteQuery.
func parseQuoteText(text string, pattern *regexp.Regexp) (string, error) {
	q := strings.TrimSpace(text)
	if q == "" || len(q) > MaxQuoteQuery || !pattern.MatchString(q) {
		return "", ErrInvalidQuoteQuery
	}
	return q, nil
}

// Period is the first and the last day of a quote's history.
type Period struct {
	From, To kernel.Date
}

// QuotePeriod returns the first and the last day of a quote's history on
// today: both empty is the last DefaultQuoteDays days ending today; a first
// day alone runs to today and a last day alone covers the DefaultQuoteDays
// before it. A period out of order, ending after today, or of more than
// MaxQuoteDays days wraps ErrInvalidQuotePeriod.
func QuotePeriod(from, to, today kernel.Date) (Period, error) {
	if to.IsZero() {
		to = today
	}
	if from.IsZero() {
		from = to.AddDays(1 - DefaultQuoteDays)
	}
	if from.After(to) || to.After(today) || from.DaysUntil(to) >= MaxQuoteDays {
		return Period{}, ErrInvalidQuotePeriod
	}
	return Period{From: from, To: to}, nil
}

// QuoteListing is one listing a market-data source names. Its Name is the
// source's own text, kept short and free of control characters by CleanName;
// every reader treats it as data and never as an instruction.
type QuoteListing struct {
	Symbol   string
	ISIN     ledger.ISIN
	Name     string
	Exchange string
	Currency kernel.Currency
}

// QuotePoint is one end-of-day close.
type QuotePoint struct {
	Date  kernel.Date
	Close kernel.Money
}

// Quote is one listing's closes of a period, in date order: public market
// data (D-206, tier T2) fetched for the person who asked and never stored.
type Quote struct {
	Listing QuoteListing
	History []QuotePoint
}

// Latest returns the latest close, and false for a quote without any.
func (q *Quote) Latest() (QuotePoint, bool) {
	if len(q.History) == 0 {
		return QuotePoint{}, false
	}
	return q.History[len(q.History)-1], true
}

// QuoteAnswer is what a market-data connector answered: the one listing the
// query names with its closes, or, when it names several, the listings and
// no quote. Source is the connector instance that answered (D-94).
type QuoteAnswer struct {
	Source     string
	Quote      *Quote
	Candidates []QuoteListing
}

// Validate checks a quote the connector answered: a latest close, closes in
// date order on days no later than to, each positive and in the listing's
// currency, at most MaxQuoteDays of them.
func (q *Quote) Validate(to kernel.Date) error {
	if len(q.History) == 0 || len(q.History) > MaxQuoteDays+1 {
		return fmt.Errorf("%w: a quote has 1 to %d closes", ErrInvalidQuote, MaxQuoteDays+1)
	}
	for i, p := range q.History {
		switch {
		case p.Date.IsZero() || p.Date.After(to):
			return fmt.Errorf("%w: a close is dated after the period or not at all", ErrInvalidQuote)
		case i > 0 && !p.Date.After(q.History[i-1].Date):
			return fmt.Errorf("%w: the closes are not in date order", ErrInvalidQuote)
		case p.Close.Amount().Sign() <= 0:
			return fmt.Errorf("%w: a close is not positive", ErrInvalidQuote)
		case !p.Close.Currency().Equal(q.Listing.Currency):
			return fmt.Errorf("%w: a close is in another currency than its listing", ErrInvalidQuote)
		}
	}
	return nil
}

// CleanName returns text, a name a market-data source gave, as one line of at
// most MaxQuoteName characters: control and other characters that print nothing, line breaks included,
// become spaces, and runs of space one. It is how the core defends the Ask
// box's model and the owner against text planted in a source's name: the text
// is kept as data, never trusted as an instruction (R-6).
func CleanName(text string) string {
	var b strings.Builder
	space := true
	n := 0
	for _, r := range text {
		if !unicode.IsPrint(r) {
			r = ' '
		}
		if r == ' ' && space {
			continue
		}
		if n == MaxQuoteName {
			break
		}
		space = r == ' '
		b.WriteRune(r)
		n++
	}
	return strings.TrimSpace(b.String())
}
