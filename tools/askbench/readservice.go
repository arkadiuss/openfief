package main

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sort"

	"connectrpc.com/connect"

	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
	"github.com/arkadiuss/openfief/internal/portfolio"
	viewsv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/views/v1"
	"github.com/arkadiuss/openfief/sdk/go/gen/openfief/views/v1/viewsv1connect"
)

// The connector instances of the bench's seed: IBKR's Flex connector is a
// network connector, and in the seed it fails (A14); XTB has none, since its
// statements are uploaded; the exchange-rate connector works.
const (
	failingInstance = "ibkr-flex-main"
	workingInstance = "nbp-fx-main"
	// syncAttempts is how many attempts the failing instance has behind it.
	syncAttempts = 4
	// stamp is the time of day of every sync in the seed.
	stamp = "T06:00:00Z"
)

// readService is the core's read API over the bench's portfolio: the calls the
// Ask box's tools make, answered from the seed in memory. Every other call is
// unimplemented, so a tool that starts reading more shows up as a failure.
type readService struct {
	viewsv1connect.UnimplementedReadServiceHandler

	seed *Seed
	pf   *Portfolio
}

var _ viewsv1connect.ReadServiceHandler = (*readService)(nil)

// readClient is a client of the read API over the seed. It embeds the
// generated client, so it is one, and talks to the service in memory, with no
// socket.
type readClient struct {
	viewsv1connect.ReadServiceClient
}

// newReadClient returns a client of the read API over seed and its valued days.
func newReadClient(s *Seed, p *Portfolio) readClient {
	_, handler := viewsv1connect.NewReadServiceHandler(&readService{seed: s, pf: p})
	hc := &http.Client{Transport: inMemory{h: handler}}
	return readClient{viewsv1connect.NewReadServiceClient(hc, "http://read.bench.invalid")}
}

// inMemory is an http.RoundTripper that hands each request to a handler.
type inMemory struct {
	h http.Handler
}

// RoundTrip serves req with the handler and returns its response.
func (m inMemory) RoundTrip(req *http.Request) (*http.Response, error) {
	rec := httptest.NewRecorder()
	m.h.ServeHTTP(rec, req)
	return rec.Result(), nil
}

// ListSleeves lists the demo's two sleeves.
func (r *readService) ListSleeves(context.Context, *connect.Request[viewsv1.ListSleevesRequest]) (*connect.Response[viewsv1.ListSleevesResponse], error) {
	out := &viewsv1.ListSleevesResponse{}
	for _, s := range r.seed.Dataset.Sleeves {
		role := viewsv1.SleeveRole_SLEEVE_ROLE_OFFENSIVE
		if s.Role == portfolio.RoleDefensive {
			role = viewsv1.SleeveRole_SLEEVE_ROLE_DEFENSIVE
		}
		out.Sleeves = append(out.Sleeves, &viewsv1.Sleeve{Id: string(s.ID), Name: s.Name, Role: role, Kind: string(s.Kind)})
	}
	return connect.NewResponse(out), nil
}

// ListAccounts lists the demo's three accounts.
func (r *readService) ListAccounts(context.Context, *connect.Request[viewsv1.ListAccountsRequest]) (*connect.Response[viewsv1.ListAccountsResponse], error) {
	out := &viewsv1.ListAccountsResponse{}
	for _, a := range r.seed.Dataset.Accounts {
		acc := &viewsv1.Account{Id: string(a.ID), Institution: a.Institution, Name: a.Name, Sleeve: string(r.seed.Dataset.AccountSleeves[a.ID])}
		if a.ID == "demo-bonds-ike" {
			acc.TaxWrapper = r.seed.Dataset.TaxWrapper.String()
		}
		out.Accounts = append(out.Accounts, acc)
	}
	return connect.NewResponse(out), nil
}

// ListValuationStatuses says every account is valued through the last day.
func (r *readService) ListValuationStatuses(context.Context, *connect.Request[viewsv1.ListValuationStatusesRequest]) (*connect.Response[viewsv1.ListValuationStatusesResponse], error) {
	out := &viewsv1.ListValuationStatusesResponse{}
	last := r.lastDay().String()
	for _, a := range r.seed.Dataset.Accounts {
		out.Statuses = append(out.Statuses, &viewsv1.ValuationStatus{AccountId: string(a.ID), ValuedThrough: last})
	}
	return connect.NewResponse(out), nil
}

// ListHoldings lists an account's holdings on a day.
func (r *readService) ListHoldings(_ context.Context, req *connect.Request[viewsv1.ListHoldingsRequest]) (*connect.Response[viewsv1.ListHoldingsResponse], error) {
	v, err := r.day(req.Msg.GetDate())
	if err != nil {
		return nil, err
	}
	out := &viewsv1.ListHoldingsResponse{AccountId: req.Msg.GetAccountId(), Date: v.Date.String()}
	for i := range v.Holdings {
		if h := &v.Holdings[i]; string(h.Account) == req.Msg.GetAccountId() {
			out.Holdings = append(out.Holdings, &viewsv1.Holding{Instrument: r.instrument(h.ISIN), Quantity: h.Quantity.String()})
		}
	}
	return connect.NewResponse(out), nil
}

// ListAccountValuations returns an account's valuations and total on each day
// of a range, the largest holding first.
func (r *readService) ListAccountValuations(_ context.Context, req *connect.Request[viewsv1.ListAccountValuationsRequest]) (*connect.Response[viewsv1.ListAccountValuationsResponse], error) {
	from, to := req.Msg.GetFrom(), req.Msg.GetTo()
	out := &viewsv1.ListAccountValuationsResponse{AccountId: req.Msg.GetAccountId(), From: from, To: to, ReportingCurrency: reporting}
	for _, v := range r.daysBetween(from, to) {
		var vals []*viewsv1.Valuation
		for i := range v.Holdings {
			if h := &v.Holdings[i]; string(h.Account) == req.Msg.GetAccountId() {
				vals = append(vals, r.valuation(v, h))
			}
		}
		sort.SliceStable(vals, func(i, j int) bool {
			return amountOf(vals[i].GetValueReporting()).Cmp(amountOf(vals[j].GetValueReporting())) > 0
		})
		out.Valuations = append(out.Valuations, vals...)
		out.Totals = append(out.Totals, r.total(v, ledger.AccountID(req.Msg.GetAccountId())))
	}
	return connect.NewResponse(out), nil
}

// ListValuationTotals returns the portfolio's totals on each day of a range.
func (r *readService) ListValuationTotals(_ context.Context, req *connect.Request[viewsv1.ListValuationTotalsRequest]) (*connect.Response[viewsv1.ListValuationTotalsResponse], error) {
	out := &viewsv1.ListValuationTotalsResponse{From: req.Msg.GetFrom(), To: req.Msg.GetTo(), ReportingCurrency: reporting}
	for _, v := range r.daysBetween(req.Msg.GetFrom(), req.Msg.GetTo()) {
		out.Totals = append(out.Totals, r.total(v, ""))
	}
	return connect.NewResponse(out), nil
}

// GetAllocation groups the gross value on a day by a dimension, the largest
// group first.
func (r *readService) GetAllocation(_ context.Context, req *connect.Request[viewsv1.GetAllocationRequest]) (*connect.Response[viewsv1.GetAllocationResponse], error) {
	v, err := r.day(req.Msg.GetDate())
	if err != nil {
		return nil, err
	}
	sums, err := r.groupSums(v, req.Msg.GetBy())
	if err != nil {
		return nil, err
	}
	out := &viewsv1.GetAllocationResponse{Date: v.Date.String(), By: req.Msg.GetBy(), Value: viewsv1.ValueKind_VALUE_KIND_GROSS, TotalReporting: money(v.Gross())}
	keys := make([]string, 0, len(sums))
	for k := range sums {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		if c := sums[keys[i]].Cmp(sums[keys[j]]); c != 0 {
			return c > 0
		}
		return keys[i] < keys[j]
	})
	for _, k := range keys {
		out.Groups = append(out.Groups, &viewsv1.AllocationGroup{Key: k, ValueReporting: money(sums[k])})
	}
	return connect.NewResponse(out), nil
}

// GetPerformance returns the daily value and time-weighted return index over a
// range.
func (r *readService) GetPerformance(_ context.Context, req *connect.Request[viewsv1.GetPerformanceRequest]) (*connect.Response[viewsv1.GetPerformanceResponse], error) {
	out := &viewsv1.GetPerformanceResponse{From: req.Msg.GetFrom(), To: req.Msg.GetTo()}
	for _, v := range r.daysBetween(req.Msg.GetFrom(), req.Msg.GetTo()) {
		out.Days = append(out.Days, &viewsv1.PerformanceDay{
			Date: v.Date.String(), Index: v.Index.String(), ValueReporting: money(v.Gross()), FlowReporting: money(v.Flow),
		})
	}
	return connect.NewResponse(out), nil
}

// ListAttentionItems lists the three open items of the bench's seed: the
// failing sync, a price that differs and a ticker to confirm.
func (r *readService) ListAttentionItems(_ context.Context, req *connect.Request[viewsv1.ListAttentionItemsRequest]) (*connect.Response[viewsv1.ListAttentionItemsResponse], error) {
	if req.Msg.GetView() != viewsv1.AttentionView_ATTENTION_VIEW_OPEN {
		return connect.NewResponse(&viewsv1.ListAttentionItemsResponse{}), nil
	}
	specs := []struct {
		kind, subject, instance string
		group                   viewsv1.AttentionItemGroup
	}{
		{"connector-sync-failing", failingInstance, failingInstance, viewsv1.AttentionItemGroup_ATTENTION_ITEM_GROUP_DATA_COMING_IN},
		{"price-check-flagged", r.symbolISIN("DWLD"), "", viewsv1.AttentionItemGroup_ATTENTION_ITEM_GROUP_NUMBERS_TO_CHECK},
		{"ticker-unconfirmed", r.symbolISIN("DEMG"), "", viewsv1.AttentionItemGroup_ATTENTION_ITEM_GROUP_DECISIONS},
	}
	out := &viewsv1.ListAttentionItemsResponse{}
	for i, s := range specs {
		out.Items = append(out.Items, &viewsv1.AttentionItem{
			Id: int64(i + 1), Kind: s.kind, Subject: s.subject, ConnectorInstance: s.instance, Group: s.group,
			State: viewsv1.AttentionItemState_ATTENTION_ITEM_STATE_OPEN,
		})
		out.Open++
	}
	return connect.NewResponse(out), nil
}

// ListSyncStatuses lists the connector instances' sync statuses.
func (r *readService) ListSyncStatuses(context.Context, *connect.Request[viewsv1.ListSyncStatusesRequest]) (*connect.Response[viewsv1.ListSyncStatusesResponse], error) {
	return connect.NewResponse(&viewsv1.ListSyncStatusesResponse{Statuses: r.statuses()}), nil
}

// ListSyncAttempts lists an instance's latest attempts.
func (r *readService) ListSyncAttempts(_ context.Context, req *connect.Request[viewsv1.ListSyncAttemptsRequest]) (*connect.Response[viewsv1.ListSyncAttemptsResponse], error) {
	today := r.lastDay()
	for _, st := range r.statuses() {
		if st.GetConnectorInstance() != req.Msg.GetConnectorInstance() {
			continue
		}
		out := &viewsv1.ListSyncAttemptsResponse{Status: st}
		for i := range st.GetFailures() {
			out.Attempts = append(out.Attempts, &viewsv1.SyncAttempt{
				ConnectorInstance: st.GetConnectorInstance(), Mode: "scheduled", StartedAt: today.AddDays(-int(i)).String() + stamp,
				Failed: true, ErrorClass: st.GetLastErrorClass(), Failures: st.GetFailures() - i,
			})
		}
		return connect.NewResponse(out), nil
	}
	return nil, fail(connect.CodeNotFound, errSeed)
}

// GetInstrument returns an instrument with its latest price.
func (r *readService) GetInstrument(_ context.Context, req *connect.Request[viewsv1.GetInstrumentRequest]) (*connect.Response[viewsv1.GetInstrumentResponse], error) {
	isin := req.Msg.GetIsin()
	info, ok := r.seed.Instruments[isin]
	if !ok {
		return nil, fail(connect.CodeNotFound, errSeed)
	}
	detail := &viewsv1.InstrumentDetail{Isin: isin, Name: info.Name, AssetClass: info.AssetClass}
	if info.AssetClass != assetBond {
		if err := r.addListing(detail, isin, &info); err != nil {
			return nil, err
		}
	}
	return connect.NewResponse(&viewsv1.GetInstrumentResponse{Instrument: detail}), nil
}

// ListInstrumentHoldings lists who holds an instrument on a day.
func (r *readService) ListInstrumentHoldings(_ context.Context, req *connect.Request[viewsv1.ListInstrumentHoldingsRequest]) (*connect.Response[viewsv1.ListInstrumentHoldingsResponse], error) {
	v, err := r.day(req.Msg.GetDate())
	if err != nil {
		return nil, err
	}
	out := &viewsv1.ListInstrumentHoldingsResponse{Isin: req.Msg.GetIsin(), Date: v.Date.String()}
	for i := range v.Holdings {
		if h := &v.Holdings[i]; h.ISIN == req.Msg.GetIsin() {
			out.Holdings = append(out.Holdings, &viewsv1.InstrumentHolding{AccountId: string(h.Account), Instrument: r.instrument(h.ISIN), Quantity: h.Quantity.String()})
		}
	}
	return connect.NewResponse(out), nil
}

// fail returns a Connect error of code for cause, wrapped as the rest of the
// bench wraps what it gets from another package.
func fail(code connect.Code, cause error) error {
	return fmt.Errorf("the bench's read service: %w", connect.NewError(code, cause))
}

// money returns a Money in PLN.
func money(d kernel.Decimal) *viewsv1.Money {
	return &viewsv1.Money{Amount: d.String(), Currency: reporting}
}

// amountOf reads a Money's amount; an unreadable one is zero.
func amountOf(m *viewsv1.Money) kernel.Decimal {
	d, err := kernel.ParseDecimal(m.GetAmount())
	if err != nil {
		return kernel.Decimal{}
	}
	return d
}

// lastDay returns the last valued day.
func (r *readService) lastDay() kernel.Date { return r.pf.Days[len(r.pf.Days)-1].Date }

// day returns the valued day for a date text; an empty one is the last day.
func (r *readService) day(text string) (*dayValues, error) {
	d := r.lastDay()
	if text != "" {
		parsed, err := kernel.ParseDate(text)
		if err != nil {
			return nil, fail(connect.CodeInvalidArgument, err)
		}
		d = parsed
	}
	v, ok := r.pf.On(d)
	if !ok {
		return nil, fail(connect.CodeNotFound, errSeed)
	}
	return v, nil
}

// instrument returns the read API's instrument for isin.
func (r *readService) instrument(isin string) *viewsv1.Instrument {
	in := r.seed.Instruments[isin]
	return &viewsv1.Instrument{Isin: isin, Symbol: in.Symbol, Name: in.Name}
}

// valuation returns h's valuation on v.
func (r *readService) valuation(v *dayValues, h *holdingValue) *viewsv1.Valuation {
	basis, source := viewsv1.ValuationBasis_VALUATION_BASIS_MARKET, "demo-synthetic"
	if r.seed.Instruments[h.ISIN].AssetClass == assetBond {
		basis, source = viewsv1.ValuationBasis_VALUATION_BASIS_GROSS_ACCRUED, "nominal"
	}
	return &viewsv1.Valuation{
		Date: v.Date.String(), Instrument: r.instrument(h.ISIN), Basis: basis, Source: source,
		Value:          &viewsv1.Money{Amount: h.Native.String(), Currency: h.Currency},
		ValueReporting: money(h.Reporting), NetReporting: money(Net(h.Reporting)), LiquidationReporting: money(Liquidation(h.Reporting)),
	}
}

// total returns the total of one account, or of every account when id is
// empty, on v.
func (r *readService) total(v *dayValues, id ledger.AccountID) *viewsv1.DateTotal {
	var gross, cash kernel.Decimal
	if id == "" {
		gross = v.Gross()
		for _, c := range v.Cash {
			cash = cash.Add(c)
		}
	} else {
		gross, cash = v.AccountGross(id), v.Cash[id]
	}
	return &viewsv1.DateTotal{
		Date: v.Date.String(), ValueReporting: money(gross), NetReporting: money(Net(gross)),
		LiquidationReporting: money(Liquidation(gross)), CashReporting: money(cash),
	}
}

// daysBetween returns the valued days from one date to another, both included;
// an empty first starts at go-live and an empty last ends at the last day.
func (r *readService) daysBetween(from, to string) []*dayValues {
	first, last := r.pf.Days[0].Date, r.lastDay()
	if d, err := kernel.ParseDate(from); err == nil && d.After(first) {
		first = d
	}
	if d, err := kernel.ParseDate(to); err == nil && d.Before(last) {
		last = d
	}
	var out []*dayValues
	for d := first; !d.After(last); d = d.AddDays(1) {
		if v, ok := r.pf.On(d); ok {
			out = append(out, v)
		}
	}
	return out
}

// groupSums sums the gross value of v by dimension: its holdings, and its cash
// by account and currency.
func (r *readService) groupSums(v *dayValues, by viewsv1.AllocationDimension) (map[string]kernel.Decimal, error) {
	sums := map[string]kernel.Decimal{}
	for i := range v.Holdings {
		h := &v.Holdings[i]
		key := r.groupKey(by, h.Account, h.ISIN, h.Currency)
		sums[key] = sums[key].Add(h.Reporting)
	}
	for acc, byCur := range v.CashByCurrency {
		for cur, amount := range byCur {
			rate, err := fxRate(cur)
			if err != nil {
				return nil, fail(connect.CodeInternal, err)
			}
			key := r.groupKey(by, acc, "", cur)
			sums[key] = sums[key].Add(round2(amount.Mul(rate)))
		}
	}
	return sums, nil
}

// groupKey names the group of a holding (or, with no isin, of cash) in acc.
func (r *readService) groupKey(by viewsv1.AllocationDimension, acc ledger.AccountID, isin, currency string) string {
	switch by {
	case viewsv1.AllocationDimension_ALLOCATION_DIMENSION_CURRENCY:
		return currency
	case viewsv1.AllocationDimension_ALLOCATION_DIMENSION_INSTITUTION:
		return r.institution(acc)
	case viewsv1.AllocationDimension_ALLOCATION_DIMENSION_SLEEVE:
		return string(r.seed.Dataset.AccountSleeves[acc])
	case viewsv1.AllocationDimension_ALLOCATION_DIMENSION_UNSPECIFIED, viewsv1.AllocationDimension_ALLOCATION_DIMENSION_ASSET_CLASS:
	}
	if isin == "" {
		return cashAssetName
	}
	return r.seed.Instruments[isin].AssetClass
}

// institution returns the institution of the account with id.
func (r *readService) institution(id ledger.AccountID) string {
	for _, a := range r.seed.Dataset.Accounts {
		if a.ID == id {
			return a.Institution
		}
	}
	return ""
}

// symbolISIN returns "ISIN/SYMBOL" of the instrument with symbol, as the
// subject of an item about it.
func (r *readService) symbolISIN(symbol string) string {
	for isin, in := range r.seed.Instruments {
		if in.Symbol == symbol {
			return isin + "/" + symbol
		}
	}
	return symbol
}

// statuses are the two instances' statuses: IBKR's failing with an
// unavailable provider since the day after the last good sync, the exchange
// rates' working.
func (r *readService) statuses() []*viewsv1.SyncStatus {
	today := r.lastDay()
	return []*viewsv1.SyncStatus{
		{
			ConnectorInstance: failingInstance, LastAttemptAt: today.String() + stamp, LastSuccessAt: today.AddDays(-syncAttempts).String() + stamp,
			Failures: syncAttempts, FailingSince: today.AddDays(1 - syncAttempts).String(), LastErrorClass: viewsv1.SyncErrorClass_SYNC_ERROR_CLASS_UNAVAILABLE,
			NextSyncAt: today.AddDays(1).String() + stamp, KeepsFailing: true,
		},
		{
			ConnectorInstance: workingInstance, LastAttemptAt: today.String() + stamp, LastSuccessAt: today.String() + stamp,
			NextSyncAt: today.AddDays(1).String() + stamp,
		},
	}
}

// addListing gives detail its listing, priced at the last close.
func (r *readService) addListing(detail *viewsv1.InstrumentDetail, isin string, info *holdingInfo) error {
	today := r.lastDay()
	c, found := priceTableOf(r.seed.Dataset).closeOn(isin, today)
	if !found {
		return fail(connect.CodeInternal, errSeed)
	}
	detail.Listings = append(detail.Listings, &viewsv1.InstrumentListing{
		Instrument: r.instrument(isin), PriceSource: "demo-synthetic",
		LatestPrice: &viewsv1.InstrumentPrice{Instrument: r.instrument(isin), Date: today.String(), Close: &viewsv1.Money{Amount: c.String(), Currency: info.Currency}},
	})
	return nil
}
