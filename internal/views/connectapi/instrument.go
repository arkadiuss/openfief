package connectapi

import (
	"context"

	"connectrpc.com/connect"

	"github.com/arkadiuss/openfief/internal/views"
	viewsv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/views/v1"
)

// GetInstrument answers GET /v1/instruments/{isin} (D-195).
func (s *readService) GetInstrument(ctx context.Context, req *connect.Request[viewsv1.GetInstrumentRequest]) (*connect.Response[viewsv1.GetInstrumentResponse], error) {
	in, err := s.views.Instrument(ctx, req.Msg.GetIsin())
	if err != nil {
		return nil, toConnectError(ctx, s.log, err)
	}
	out := &viewsv1.InstrumentDetail{
		Isin: in.ISIN, AssetClass: in.AssetClass.String(),
		Listings:         make([]*viewsv1.InstrumentListing, 0, len(in.Listings)),
		ExemptionClasses: make([]*viewsv1.InstrumentExemptionClass, 0, len(in.Classes)),
	}
	for i := range in.Listings {
		l := &in.Listings[i]
		pl := &viewsv1.InstrumentListing{Instrument: pbInstrument(l.Instrument), PriceSource: l.Source}
		if l.Latest != nil {
			pl.LatestPrice = pbInstrumentPrice(&views.InstrumentPrice{Price: *l.Latest, Primary: true})
		}
		out.Listings = append(out.Listings, pl)
	}
	for _, c := range in.Classes {
		out.ExemptionClasses = append(out.ExemptionClasses, &viewsv1.InstrumentExemptionClass{
			Wrapper: string(c.Wrapper), ExemptionClass: string(c.Class), OwnerChosen: c.OwnerChosen, Known: c.Known,
		})
	}
	return connect.NewResponse(&viewsv1.GetInstrumentResponse{Instrument: out}), nil
}

// pbInstrumentPrice converts one stored price of a listing.
func pbInstrumentPrice(p *views.InstrumentPrice) *viewsv1.InstrumentPrice {
	out := &viewsv1.InstrumentPrice{Date: pbDate(p.Date), Close: pbMoney(p.Close), Source: p.Source, Primary: p.Primary}
	if in, ok := views.ListingOf(p.Instrument); ok {
		out.Instrument = pbInstrument(in)
	}
	return out
}

// ListInstrumentPrices answers GET /v1/instruments/{isin}/prices (D-94,
// D-195).
func (s *readService) ListInstrumentPrices(ctx context.Context, req *connect.Request[viewsv1.ListInstrumentPricesRequest]) (*connect.Response[viewsv1.ListInstrumentPricesResponse], error) {
	from, err := parseDate(req.Msg.GetFrom())
	if err != nil {
		return nil, toConnectError(ctx, s.log, err)
	}
	to, err := parseDate(req.Msg.GetTo())
	if err != nil {
		return nil, toConnectError(ctx, s.log, err)
	}
	prices, err := s.views.InstrumentPrices(ctx, req.Msg.GetIsin(), from, to)
	if err != nil {
		return nil, toConnectError(ctx, s.log, err)
	}
	out := &viewsv1.ListInstrumentPricesResponse{Isin: req.Msg.GetIsin(), Prices: make([]*viewsv1.InstrumentPrice, 0, len(prices))}
	for i := range prices {
		out.Prices = append(out.Prices, pbInstrumentPrice(&prices[i]))
	}
	return connect.NewResponse(out), nil
}

// ListInstrumentHoldings answers GET /v1/instruments/{isin}/holdings
// (D-195).
func (s *readService) ListInstrumentHoldings(ctx context.Context, req *connect.Request[viewsv1.ListInstrumentHoldingsRequest]) (*connect.Response[viewsv1.ListInstrumentHoldingsResponse], error) {
	on, err := parseDate(req.Msg.GetDate())
	if err != nil {
		return nil, toConnectError(ctx, s.log, err)
	}
	holdings, err := s.views.InstrumentHoldings(ctx, req.Msg.GetIsin(), on)
	if err != nil {
		return nil, toConnectError(ctx, s.log, err)
	}
	out := &viewsv1.ListInstrumentHoldingsResponse{
		Isin: req.Msg.GetIsin(), Date: pbDate(on), Holdings: make([]*viewsv1.InstrumentHolding, 0, len(holdings)),
	}
	for i := range holdings {
		h := &holdings[i]
		ph := &viewsv1.InstrumentHolding{
			AccountId: string(h.Account), Instrument: pbInstrument(h.Holding.Instrument), Quantity: h.Holding.Quantity.String(),
			Lots: make([]*viewsv1.Lot, 0, len(h.Lots)),
		}
		for j := range h.Lots {
			ph.Lots = append(ph.Lots, pbLot(&h.Lots[j], nil))
		}
		out.Holdings = append(out.Holdings, ph)
	}
	return connect.NewResponse(out), nil
}

// ListInstrumentTransactions answers GET
// /v1/instruments/{isin}/transactions (D-195).
func (s *readService) ListInstrumentTransactions(ctx context.Context, req *connect.Request[viewsv1.ListInstrumentTransactionsRequest]) (*connect.Response[viewsv1.ListInstrumentTransactionsResponse], error) {
	var after *views.TransactionCursor
	if token := req.Msg.GetPageToken(); token != "" {
		c, err := views.ParseTransactionCursor(token)
		if err != nil {
			return nil, toConnectError(ctx, s.log, err)
		}
		after = &c
	}
	page, err := s.views.InstrumentTransactions(ctx, req.Msg.GetIsin(), after, pageSize(req.Msg.GetPageSize()))
	if err != nil {
		return nil, toConnectError(ctx, s.log, err)
	}
	out := &viewsv1.ListInstrumentTransactionsResponse{Transactions: make([]*viewsv1.Transaction, 0, len(page.Transactions))}
	for i := range page.Transactions {
		var origin views.Origin
		if i < len(page.Origins) {
			origin = page.Origins[i]
		}
		out.Transactions = append(out.Transactions, pbTransaction(&page.Transactions[i], origin))
	}
	if page.Next != nil {
		out.NextPageToken = page.Next.Encode()
	}
	return connect.NewResponse(out), nil
}
