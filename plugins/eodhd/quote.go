package main

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"unicode"
	"unicode/utf8"

	"connectrpc.com/connect"

	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/valuation/eodhd"
	connectorv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/connector/v1"
)

// Limits of a quote: the longest query and period the plugin takes, how far
// before a period it looks for the latest close when the period holds none,
// and how long a name it passes on.
const (
	maxQuoteQuery     = 32
	maxQuoteDays      = 366
	quoteLookbackDays = 14
	maxQuoteName      = 200
)

// quoteQuery is what a ticker or an ISIN looks like: letters, digits, dots,
// dashes and underscores.
var quoteQuery = regexp.MustCompile(`^[A-Za-z0-9._-]{1,32}$`)

// errNoCloses reports a listing the source has no close of.
var errNoCloses = errors.New("no closes")

// errBadQuote reports a quote request with a query or a period the plugin
// doesn't take.
var errBadQuote = errors.New("invalid quote request")

// LookupQuote finds the one listing a ticker or an ISIN names and returns its
// latest end-of-day close with the closes of the period, from EODHD on the
// owner's key (D-226). Nothing is kept: the answer is public market data for
// the person who asked. A query naming several ISINs gets their listings and
// no quote; a query naming none is NotFound; an invalid query or period is
// InvalidArgument, a missing API key FailedPrecondition, and EODHD failing
// Unavailable.
func (c *connector) LookupQuote(ctx context.Context, req *connect.Request[connectorv1.LookupQuoteRequest]) (*connect.Response[connectorv1.LookupQuoteResponse], error) {
	asked, err := parseQuote(req.Msg, c.today())
	if err != nil {
		return nil, status(connect.CodeInvalidArgument, err)
	}
	token, err := readSecretFile(c.secrets, apiTokenFile)
	if err != nil {
		return nil, status(connect.CodeFailedPrecondition, err)
	}
	src := c.source(token)
	listings, err := src.Find(ctx, asked.query)
	if err != nil {
		return nil, status(connect.CodeUnavailable, fmt.Errorf("eodhd: %w", err))
	}
	match, err := eodhd.MatchQuery(listings, asked.query)
	switch {
	case errors.Is(err, eodhd.ErrAmbiguous):
		out := &connectorv1.LookupQuoteResponse{}
		for i := range match.Candidates {
			out.Candidates = append(out.Candidates, quoteListing(&match.Candidates[i]))
		}
		return connect.NewResponse(out), nil
	case err != nil:
		return nil, status(connect.CodeNotFound, fmt.Errorf("eodhd: %w", err))
	}
	listing := match.Listing
	closes, err := c.quoteCloses(ctx, src, &listing, asked.period.From, asked.period.To)
	if err != nil {
		return nil, err
	}
	out := &connectorv1.LookupQuoteResponse{Listing: quoteListing(&listing)}
	for _, cl := range closes {
		m, qErr := listing.Quote(cl.Close)
		if qErr != nil {
			return nil, status(connect.CodeUnavailable, fmt.Errorf("eodhd: %w", qErr))
		}
		out.Listing.Currency = m.Currency().String()
		out.History = append(out.History, &connectorv1.QuotePoint{Date: cl.Date.String(), Close: m.Amount().String()})
	}
	return connect.NewResponse(out), nil
}

// quoteCloses returns listing's closes over the days from through to; a period with none,
// such as a holiday week, is answered with the closes of the quoteLookbackDays
// before it, so the quote always has a latest close when the listing has any.
// A listing with no close at all is NotFound.
func (c *connector) quoteCloses(ctx context.Context, src marketData, listing *eodhd.Listing, from, to kernel.Date) ([]eodhd.Close, error) {
	closes, err := src.Closes(ctx, listing.QuoteTicker(), from, to)
	if err == nil && len(closes) == 0 {
		closes, err = src.Closes(ctx, listing.QuoteTicker(), from.AddDays(-quoteLookbackDays), to)
		if err == nil && len(closes) > 0 {
			closes = closes[len(closes)-1:]
		}
	}
	switch {
	case errors.Is(err, eodhd.ErrNotFound) || (err == nil && len(closes) == 0):
		return nil, status(connect.CodeNotFound, fmt.Errorf("eodhd: %w of %s", errNoCloses, listing.QuoteTicker()))
	case err != nil:
		return nil, status(connect.CodeUnavailable, fmt.Errorf("eodhd: %w", err))
	}
	return closes, nil
}

// quoteAsked is a checked quote request.
type quoteAsked struct {
	query  string
	period quotePeriod
}

// quotePeriod is the first and last day of the closes asked for.
type quotePeriod struct {
	From, To kernel.Date
}

// parseQuote checks a quote request on today: a query of the allowed
// characters and a period of at most maxQuoteDays days that ends today or
// earlier.
func parseQuote(r *connectorv1.LookupQuoteRequest, today kernel.Date) (quoteAsked, error) {
	if !quoteQuery.MatchString(r.GetQuery()) {
		return quoteAsked{}, fmt.Errorf("%w: the query is not a ticker or an ISIN", errBadQuote)
	}
	from, fromErr := kernel.ParseDate(r.GetFrom())
	to, toErr := kernel.ParseDate(r.GetTo())
	switch {
	case fromErr != nil || toErr != nil:
		return quoteAsked{}, fmt.Errorf("%w: the period is not two days", errBadQuote)
	case from.After(to) || to.After(today) || from.DaysUntil(to) >= maxQuoteDays:
		return quoteAsked{}, fmt.Errorf("%w: the period is out of order, in the future or over %d days", errBadQuote, maxQuoteDays)
	}
	return quoteAsked{query: r.GetQuery(), period: quotePeriod{From: from, To: to}}, nil
}

// quoteListing translates a listing. The name is the source's own text, so it
// is cut to maxQuoteName characters and loses its control characters here; the
// core checks it again.
func quoteListing(l *eodhd.Listing) *connectorv1.QuoteListing {
	currency := l.Currency
	if currency == "GBX" { // closes in pence are delivered in pounds
		currency = "GBP"
	}
	return &connectorv1.QuoteListing{
		Symbol: l.Code, Isin: l.ISIN.String(), Name: cleanName(l.Name), Exchange: l.Exchange, Currency: currency,
	}
}

// cleanName returns name without control characters and at most maxQuoteName
// characters.
func cleanName(name string) string {
	out := make([]rune, 0, utf8.RuneCountInString(name))
	for _, r := range name {
		if unicode.IsControl(r) {
			r = ' '
		}
		if len(out) == maxQuoteName {
			break
		}
		out = append(out, r)
	}
	return string(out)
}
