package connectapi

import (
	"context"
	"errors"

	"connectrpc.com/connect"

	"github.com/arkadiuss/openfief/internal/ingest"
	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/views/app"
	viewsv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/views/v1"
)

// quoteTier is the data tier of a market quote (D-206): public data.
const quoteTier = "T2"

// What clients see of a failed market quote: fixed texts, since an error's own
// text can quote what was asked for or what the provider said.
var (
	errQuotesUnavailable = errors.New("market quotes are not available: no market-data connector is installed and approved, or this install answers from its synthetic data alone")
	errNoQuote           = errors.New("the market-data source has no such instrument")
	errQuoteSourceDown   = errors.New("the market-data source can't answer now")
	errQuotesForbidden   = errors.New("this caller may not read market quotes")
	errQuoteRateLimited  = errors.New("too many market quotes were asked in a minute")
	errBadQuoteQuery     = errors.New("the query is not a ticker or an ISIN of at most 32 letters, digits, dots, dashes and underscores")
	errBadQuotePeriod    = errors.New("the period ends before it starts, ends after today or is longer than 366 days")
)

// quoteError returns the Connect error err stands for when it is one of a
// market quote's, and nil for any other: unavailable quotes are a failed
// precondition, an unknown instrument not found, a source that is down
// unavailable, a minute's lookups spent resource exhausted and an invalid
// query or period invalid argument.
func quoteError(err error) *connect.Error {
	switch {
	case errors.Is(err, app.ErrQuotesUnavailable):
		return connect.NewError(connect.CodeFailedPrecondition, errQuotesUnavailable)
	case errors.Is(err, app.ErrQuoteNotFound):
		return connect.NewError(connect.CodeNotFound, errNoQuote)
	case errors.Is(err, app.ErrQuoteSourceDown):
		return connect.NewError(connect.CodeUnavailable, errQuoteSourceDown)
	case errors.Is(err, app.ErrQuotesForbidden):
		return connect.NewError(connect.CodePermissionDenied, errQuotesForbidden)
	case errors.Is(err, app.ErrQuoteRateLimited):
		return connect.NewError(connect.CodeResourceExhausted, errQuoteRateLimited)
	case errors.Is(err, ingest.ErrInvalidQuoteQuery):
		return connect.NewError(connect.CodeInvalidArgument, errBadQuoteQuery)
	case errors.Is(err, ingest.ErrInvalidQuotePeriod):
		return connect.NewError(connect.CodeInvalidArgument, errBadQuotePeriod)
	}
	return nil
}

// GetMarketQuote answers GET /v1/market-quotes/{query} (D-226): the latest
// price and history of one instrument from the owner's market-data connector,
// public data of tier T2 that names its source.
func (s *readService) GetMarketQuote(ctx context.Context, req *connect.Request[viewsv1.GetMarketQuoteRequest]) (*connect.Response[viewsv1.GetMarketQuoteResponse], error) {
	from, err := optionalDate(req.Msg.GetFrom())
	if err != nil {
		return nil, toConnectError(ctx, s.log, err)
	}
	to, err := optionalDate(req.Msg.GetTo())
	if err != nil {
		return nil, toConnectError(ctx, s.log, err)
	}
	answer, err := s.views.MarketQuote(ctx, req.Msg.GetQuery(), from, to)
	if err != nil {
		return nil, toConnectError(ctx, s.log, err)
	}
	out := &viewsv1.GetMarketQuoteResponse{Source: answer.Source, DataTier: quoteTier}
	for i := range answer.Candidates {
		out.Candidates = append(out.Candidates, pbQuoteListing(&answer.Candidates[i]))
	}
	if q := answer.Quote; q != nil {
		out.Quote = pbMarketQuote(q, answer.Source)
	}
	return connect.NewResponse(out), nil
}

// optionalDate parses a request's date, or returns the zero date for none.
func optionalDate(s string) (kernel.Date, error) {
	if s == "" {
		return kernel.Date{}, nil
	}
	return parseDate(s)
}

// pbQuoteListing converts a listing a market-data source named.
func pbQuoteListing(l *ingest.QuoteListing) *viewsv1.MarketQuoteListing {
	return &viewsv1.MarketQuoteListing{
		Instrument: &viewsv1.Instrument{Isin: l.ISIN.String(), Symbol: l.Symbol, Name: l.Name},
		Exchange:   l.Exchange, Currency: l.Currency.String(),
	}
}

// pbMarketQuote converts a quote that source answered, with its latest close
// and day.
func pbMarketQuote(q *ingest.Quote, source string) *viewsv1.MarketQuote {
	out := &viewsv1.MarketQuote{Listing: pbQuoteListing(&q.Listing), Source: source, DataTier: quoteTier}
	for _, p := range q.History {
		out.History = append(out.History, &viewsv1.MarketQuotePoint{Date: p.Date.String(), Close: pbMoney(p.Close)})
	}
	if latest, ok := q.Latest(); ok {
		out.Latest, out.Date = pbMoney(latest.Close), latest.Date.String()
	}
	return out
}
