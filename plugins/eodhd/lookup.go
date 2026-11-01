package main

import (
	"context"
	"errors"
	"fmt"

	"connectrpc.com/connect"

	"github.com/arkadiuss/openfief/internal/valuation/eodhd"
	connectorv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/connector/v1"
)

// maxLookupTickers caps the tickers one lookup asks EODHD about.
const maxLookupTickers = 100

// errTooManyTickers reports a lookup of more than maxLookupTickers.
var errTooManyTickers = errors.New("too many tickers")

// LookupInstruments proposes the ISIN of each ticker a broker named without
// one (D-116): it searches EODHD for the ticker's code on the supported
// exchange of its market's country and proposes the one ISIN listed as
// that code there, with the currency the listing trades in (D-159). A
// ticker of a country with no supported exchange, or
// with no single ISIN, gets no proposal. More than 100 tickers is
// InvalidArgument, a missing API key FailedPrecondition, and EODHD failing
// Unavailable.
func (c *connector) LookupInstruments(ctx context.Context, req *connect.Request[connectorv1.LookupInstrumentsRequest]) (*connect.Response[connectorv1.LookupInstrumentsResponse], error) {
	tickers := req.Msg.GetTickers()
	if len(tickers) > maxLookupTickers {
		return nil, status(connect.CodeInvalidArgument, fmt.Errorf("%w: more than %d", errTooManyTickers, maxLookupTickers))
	}
	token, err := readSecretFile(c.secrets, apiTokenFile)
	if err != nil {
		return nil, status(connect.CodeFailedPrecondition, err)
	}
	src := c.source(token)
	out := &connectorv1.LookupInstrumentsResponse{}
	for _, t := range tickers {
		exchange := eodhd.CountryExchange(t.GetCountry())
		if exchange == "" || t.GetCode() == "" {
			continue
		}
		listings, err := src.Search(ctx, t.GetCode(), exchange)
		if err != nil {
			return nil, status(connect.CodeUnavailable, fmt.Errorf("eodhd: %w", err))
		}
		if l, ok := eodhd.FindByCode(listings, t.GetCode(), exchange); ok {
			out.Proposals = append(out.Proposals, &connectorv1.InstrumentProposal{
				Symbol: t.GetSymbol(), Isin: l.ISIN.String(), Name: l.Name, Currency: l.Currency,
			})
		}
	}
	return connect.NewResponse(out), nil
}
