package connectrpc

import (
	"context"
	"errors"
	"fmt"

	"connectrpc.com/connect"

	"github.com/arkadiuss/openfief/internal/ingest"
	"github.com/arkadiuss/openfief/internal/ingest/app"
	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
	connectorv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/connector/v1"
)

// errNoListing is an answer with no listing and no candidate.
var errNoListing = errors.New("the answer names no listing")

// maxExchangeRunes is the longest exchange code a quote carries.
const maxExchangeRunes = 16

// LookupQuote asks instance, a market-data connector, for the listing a
// ticker or an ISIN names and its closes in the request's period (D-226). The
// answer is translated and checked here: the connector's text is the source's
// own, so names are cut and cleaned (ingest.CleanName), a close must be a
// positive amount in its listing's currency, dated in order and no later than
// the period's last day, and a quote that doesn't fit is an error wrapping
// ErrInvalidMessage. A connector that can't quote, answering
// UNIMPLEMENTED, wraps app.ErrNoEntries; a query it finds nothing for
// app.ErrQuoteNotFound; one that can't be reached, or has no API key,
// app.ErrConnectorUnavailable; an invalid request app.ErrEntryRefused.
func (c *Client) LookupQuote(ctx context.Context, instance string, req ingest.QuoteRequest) (ingest.QuoteAnswer, error) {
	call, err := c.service(ctx, instance)
	if err != nil {
		return ingest.QuoteAnswer{}, fmt.Errorf("%w: %w", app.ErrConnectorUnavailable, err)
	}
	msg := connect.NewRequest(&connectorv1.LookupQuoteRequest{Query: req.Query, From: req.From.String(), To: req.To.String()})
	msg.Header().Set("Authorization", call.auth)
	resp, err := call.svc.LookupQuote(ctx, msg)
	if err != nil {
		return ingest.QuoteAnswer{}, fmt.Errorf("connector %s: %w", instance, quoteError(err))
	}
	out, err := quoteFrom(resp.Msg, req.To)
	if err != nil {
		return ingest.QuoteAnswer{}, fmt.Errorf("connector %s: %w", instance, err)
	}
	out.Source = instance
	return out, nil
}

// quoteError is entryError for a quote: NOT_FOUND is app.ErrQuoteNotFound and
// FAILED_PRECONDITION, a connector without its API key, is
// app.ErrConnectorUnavailable.
func quoteError(err error) error {
	code := connect.CodeOf(err)
	if code == connect.CodeNotFound {
		return app.ErrQuoteNotFound
	}
	if code == connect.CodeFailedPrecondition {
		return fmt.Errorf("%w: %w", app.ErrConnectorUnavailable, err)
	}
	return entryError(err)
}

// quoteFrom translates a connector's answer, whose last close is no later
// than to.
func quoteFrom(m *connectorv1.LookupQuoteResponse, to kernel.Date) (ingest.QuoteAnswer, error) {
	if m.GetListing() == nil {
		return candidatesFrom(m)
	}
	listing, err := quoteListingFrom("listing", m.GetListing())
	if err != nil {
		return ingest.QuoteAnswer{}, err
	}
	q := ingest.Quote{Listing: listing}
	for i, p := range m.GetHistory() {
		field := fmt.Sprintf("history[%d]", i)
		on, err := dateFrom(field+".date", p.GetDate())
		if err != nil {
			return ingest.QuoteAnswer{}, err
		}
		if !plainDecimal(p.GetClose()) {
			return ingest.QuoteAnswer{}, invalid(field+".close", nil)
		}
		closing, err := kernel.ParseMoney(p.GetClose(), listing.Currency)
		if err != nil {
			return ingest.QuoteAnswer{}, invalid(field+".close", err)
		}
		q.History = append(q.History, ingest.QuotePoint{Date: on, Close: closing})
	}
	if err := q.Validate(to); err != nil {
		return ingest.QuoteAnswer{}, invalid("history", err)
	}
	return ingest.QuoteAnswer{Quote: &q}, nil
}

// candidatesFrom translates the listings a connector offered when the query
// named several; an answer with no listing and no candidate is invalid.
func candidatesFrom(m *connectorv1.LookupQuoteResponse) (ingest.QuoteAnswer, error) {
	var out ingest.QuoteAnswer
	for i, l := range m.GetCandidates() {
		if i == ingest.MaxQuoteCandidates {
			break
		}
		listing, err := quoteListingFrom(fmt.Sprintf("candidates[%d]", i), l)
		if err != nil {
			return ingest.QuoteAnswer{}, err
		}
		out.Candidates = append(out.Candidates, listing)
	}
	if len(out.Candidates) == 0 {
		return ingest.QuoteAnswer{}, invalid("listing", errNoListing)
	}
	return out, nil
}

// quoteListingFrom translates one listing: a ticker's symbol of the allowed
// characters, an ISIN with a valid check digit or none, a name and an
// exchange cleaned as text of the source's, and a currency.
func quoteListingFrom(field string, l *connectorv1.QuoteListing) (ingest.QuoteListing, error) {
	symbol, err := ingest.ParseQuoteSymbol(l.GetSymbol())
	if err != nil {
		return ingest.QuoteListing{}, invalid(field+".symbol", err)
	}
	var isin ledger.ISIN
	if l.GetIsin() != "" {
		if isin, err = ledger.ParseISIN(l.GetIsin()); err != nil {
			return ingest.QuoteListing{}, invalid(field+".isin", err)
		}
	}
	currency, err := kernel.ParseCurrency(l.GetCurrency())
	if err != nil {
		return ingest.QuoteListing{}, invalid(field+".currency", err)
	}
	exchange := ingest.CleanName(l.GetExchange())
	if r := []rune(exchange); len(r) > maxExchangeRunes {
		exchange = string(r[:maxExchangeRunes])
	}
	return ingest.QuoteListing{Symbol: symbol, ISIN: isin, Name: ingest.CleanName(l.GetName()), Exchange: exchange, Currency: currency}, nil
}
