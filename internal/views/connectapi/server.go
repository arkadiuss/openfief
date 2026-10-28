package connectapi

import (
	"context"
	"fmt"
	"log/slog"

	"connectrpc.com/connect"

	"github.com/arkadiuss/openfief/internal/ledger"
	ledgerapp "github.com/arkadiuss/openfief/internal/ledger/app"
	"github.com/arkadiuss/openfief/internal/views"
	"github.com/arkadiuss/openfief/internal/views/app"
	viewsv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/views/v1"
	"github.com/arkadiuss/openfief/sdk/go/gen/openfief/views/v1/viewsv1connect"
)

// readService implements the generated ReadServiceHandler over the views use
// cases. Every error goes through toConnectError.
type readService struct {
	views *app.Views
	log   *slog.Logger
}

var _ viewsv1connect.ReadServiceHandler = (*readService)(nil)

// ListSleeves answers GET /v1/sleeves.
func (s *readService) ListSleeves(ctx context.Context, _ *connect.Request[viewsv1.ListSleevesRequest]) (*connect.Response[viewsv1.ListSleevesResponse], error) {
	sleeves, err := s.views.Sleeves(ctx)
	if err != nil {
		return nil, toConnectError(ctx, s.log, err)
	}
	out := &viewsv1.ListSleevesResponse{Sleeves: make([]*viewsv1.Sleeve, 0, len(sleeves))}
	for i := range sleeves {
		out.Sleeves = append(out.Sleeves, pbSleeve(&sleeves[i]))
	}
	return connect.NewResponse(out), nil
}

// ListAccounts answers GET /v1/accounts.
func (s *readService) ListAccounts(ctx context.Context, _ *connect.Request[viewsv1.ListAccountsRequest]) (*connect.Response[viewsv1.ListAccountsResponse], error) {
	accounts, err := s.views.Accounts(ctx)
	if err != nil {
		return nil, toConnectError(ctx, s.log, err)
	}
	out := &viewsv1.ListAccountsResponse{Accounts: make([]*viewsv1.Account, 0, len(accounts))}
	for i := range accounts {
		out.Accounts = append(out.Accounts, pbAccount(&accounts[i]))
	}
	return connect.NewResponse(out), nil
}

// GetAccount answers GET /v1/accounts/{account_id}.
func (s *readService) GetAccount(ctx context.Context, req *connect.Request[viewsv1.GetAccountRequest]) (*connect.Response[viewsv1.GetAccountResponse], error) {
	a, err := s.views.Account(ctx, ledger.AccountID(req.Msg.GetAccountId()))
	if err != nil {
		return nil, toConnectError(ctx, s.log, err)
	}
	return connect.NewResponse(&viewsv1.GetAccountResponse{Account: pbAccount(&a)}), nil
}

// ListTransactions answers GET /v1/accounts/{account_id}/transactions.
func (s *readService) ListTransactions(ctx context.Context, req *connect.Request[viewsv1.ListTransactionsRequest]) (*connect.Response[viewsv1.ListTransactionsResponse], error) {
	size := int(req.Msg.GetPageSize())
	if size == 0 {
		size = views.DefaultPageSize
	}
	var after *views.TransactionCursor
	if token := req.Msg.GetPageToken(); token != "" {
		c, err := views.ParseTransactionCursor(token)
		if err != nil {
			return nil, toConnectError(ctx, s.log, err)
		}
		after = &c
	}
	page, err := s.views.Transactions(ctx, ledger.AccountID(req.Msg.GetAccountId()), after, size)
	if err != nil {
		return nil, toConnectError(ctx, s.log, err)
	}
	out := &viewsv1.ListTransactionsResponse{Transactions: make([]*viewsv1.Transaction, 0, len(page.Transactions))}
	for i := range page.Transactions {
		out.Transactions = append(out.Transactions, pbTransaction(&page.Transactions[i]))
	}
	if page.Next != nil {
		out.NextPageToken = page.Next.Encode()
	}
	return connect.NewResponse(out), nil
}

// ListHoldings answers GET /v1/accounts/{account_id}/holdings.
func (s *readService) ListHoldings(ctx context.Context, req *connect.Request[viewsv1.ListHoldingsRequest]) (*connect.Response[viewsv1.ListHoldingsResponse], error) {
	on, err := parseDate(req.Msg.GetDate())
	if err != nil {
		return nil, toConnectError(ctx, s.log, err)
	}
	holdings, err := s.views.Holdings(ctx, ledger.AccountID(req.Msg.GetAccountId()), on)
	if err != nil {
		return nil, toConnectError(ctx, s.log, err)
	}
	out := &viewsv1.ListHoldingsResponse{AccountId: req.Msg.GetAccountId(), Date: pbDate(on), Holdings: make([]*viewsv1.Holding, 0, len(holdings))}
	for i := range holdings {
		out.Holdings = append(out.Holdings, &viewsv1.Holding{Instrument: pbInstrument(holdings[i].Instrument), Quantity: holdings[i].Quantity.String()})
	}
	return connect.NewResponse(out), nil
}

// ListLots answers GET /v1/accounts/{account_id}/lots.
func (s *readService) ListLots(ctx context.Context, req *connect.Request[viewsv1.ListLotsRequest]) (*connect.Response[viewsv1.ListLotsResponse], error) {
	on, err := parseDate(req.Msg.GetDate())
	if err != nil {
		return nil, toConnectError(ctx, s.log, err)
	}
	lots, err := s.views.Lots(ctx, ledger.AccountID(req.Msg.GetAccountId()), on)
	if err != nil {
		return nil, toConnectError(ctx, s.log, err)
	}
	out := &viewsv1.ListLotsResponse{AccountId: req.Msg.GetAccountId(), Date: pbDate(on), Lots: make([]*viewsv1.Lot, 0, len(lots))}
	for i := range lots {
		out.Lots = append(out.Lots, pbLot(&lots[i]))
	}
	return connect.NewResponse(out), nil
}

// ListCash answers GET /v1/accounts/{account_id}/cash.
func (s *readService) ListCash(ctx context.Context, req *connect.Request[viewsv1.ListCashRequest]) (*connect.Response[viewsv1.ListCashResponse], error) {
	on, err := parseDate(req.Msg.GetDate())
	if err != nil {
		return nil, toConnectError(ctx, s.log, err)
	}
	cash, err := s.views.Cash(ctx, ledger.AccountID(req.Msg.GetAccountId()), on)
	if err != nil {
		return nil, toConnectError(ctx, s.log, err)
	}
	return connect.NewResponse(&viewsv1.ListCashResponse{AccountId: req.Msg.GetAccountId(), Date: pbDate(on), Balances: pbMonies(cash)}), nil
}

// ListAssertionChecks answers GET /v1/accounts/{account_id}/balance-assertions.
func (s *readService) ListAssertionChecks(ctx context.Context, req *connect.Request[viewsv1.ListAssertionChecksRequest]) (*connect.Response[viewsv1.ListAssertionChecksResponse], error) {
	status, err := assertionStatus(req.Msg.GetStatus())
	if err != nil {
		return nil, toConnectError(ctx, s.log, err)
	}
	checks, err := s.views.AssertionChecks(ctx, ledger.AccountID(req.Msg.GetAccountId()), status)
	if err != nil {
		return nil, toConnectError(ctx, s.log, err)
	}
	out := &viewsv1.ListAssertionChecksResponse{AccountId: req.Msg.GetAccountId(), Checks: make([]*viewsv1.AssertionCheck, 0, len(checks))}
	for i := range checks {
		out.Checks = append(out.Checks, pbCheck(&checks[i]))
	}
	return connect.NewResponse(out), nil
}

// assertionStatus converts a requested status filter; unspecified means
// every status.
func assertionStatus(s viewsv1.AssertionStatus) (ledgerapp.AssertionStatus, error) {
	switch s {
	case viewsv1.AssertionStatus_ASSERTION_STATUS_UNSPECIFIED:
		return 0, nil
	case viewsv1.AssertionStatus_ASSERTION_STATUS_MATCHED:
		return ledgerapp.AssertionMatched, nil
	case viewsv1.AssertionStatus_ASSERTION_STATUS_FLAGGED:
		return ledgerapp.AssertionFlagged, nil
	}
	return 0, fmt.Errorf("%w: assertion status %d", errBadParameter, s)
}

// ListAccountValuations answers GET /v1/accounts/{account_id}/valuations.
func (s *readService) ListAccountValuations(ctx context.Context, req *connect.Request[viewsv1.ListAccountValuationsRequest]) (*connect.Response[viewsv1.ListAccountValuationsResponse], error) {
	from, err := parseDate(req.Msg.GetFrom())
	if err != nil {
		return nil, toConnectError(ctx, s.log, err)
	}
	to, err := parseDate(req.Msg.GetTo())
	if err != nil {
		return nil, toConnectError(ctx, s.log, err)
	}
	got, err := s.views.AccountValuations(ctx, ledger.AccountID(req.Msg.GetAccountId()), from, to)
	if err != nil {
		return nil, toConnectError(ctx, s.log, err)
	}
	out := &viewsv1.ListAccountValuationsResponse{
		AccountId: req.Msg.GetAccountId(), From: pbDate(from), To: pbDate(to), ReportingCurrency: s.views.ReportingCurrency().String(),
		Valuations: make([]*viewsv1.Valuation, 0, len(got.Valuations)), Totals: pbTotals(got.Totals),
	}
	for i := range got.Valuations {
		out.Valuations = append(out.Valuations, pbValuation(&got.Valuations[i]))
	}
	return connect.NewResponse(out), nil
}

// ListValuationTotals answers GET /v1/valuations.
func (s *readService) ListValuationTotals(ctx context.Context, req *connect.Request[viewsv1.ListValuationTotalsRequest]) (*connect.Response[viewsv1.ListValuationTotalsResponse], error) {
	from, err := parseDate(req.Msg.GetFrom())
	if err != nil {
		return nil, toConnectError(ctx, s.log, err)
	}
	to, err := parseDate(req.Msg.GetTo())
	if err != nil {
		return nil, toConnectError(ctx, s.log, err)
	}
	totals, err := s.views.ValuationTotals(ctx, from, to)
	if err != nil {
		return nil, toConnectError(ctx, s.log, err)
	}
	return connect.NewResponse(&viewsv1.ListValuationTotalsResponse{
		From: pbDate(from), To: pbDate(to), Totals: pbTotals(totals), ReportingCurrency: s.views.ReportingCurrency().String(),
	}), nil
}

// GetValuationStatus answers GET /v1/accounts/{account_id}/valuation-status.
func (s *readService) GetValuationStatus(ctx context.Context, req *connect.Request[viewsv1.GetValuationStatusRequest]) (*connect.Response[viewsv1.GetValuationStatusResponse], error) {
	st, err := s.views.ValuationStatus(ctx, ledger.AccountID(req.Msg.GetAccountId()))
	if err != nil {
		return nil, toConnectError(ctx, s.log, err)
	}
	return connect.NewResponse(&viewsv1.GetValuationStatusResponse{Status: pbStatus(&st)}), nil
}

// ListValuationStatuses answers GET /v1/valuation-status.
func (s *readService) ListValuationStatuses(ctx context.Context, _ *connect.Request[viewsv1.ListValuationStatusesRequest]) (*connect.Response[viewsv1.ListValuationStatusesResponse], error) {
	statuses, err := s.views.ValuationStatuses(ctx)
	if err != nil {
		return nil, toConnectError(ctx, s.log, err)
	}
	out := &viewsv1.ListValuationStatusesResponse{Statuses: make([]*viewsv1.ValuationStatus, 0, len(statuses))}
	for i := range statuses {
		out.Statuses = append(out.Statuses, pbStatus(&statuses[i]))
	}
	return connect.NewResponse(out), nil
}

// GetAllocation answers GET /v1/allocation.
func (s *readService) GetAllocation(ctx context.Context, req *connect.Request[viewsv1.GetAllocationRequest]) (*connect.Response[viewsv1.GetAllocationResponse], error) {
	on, err := parseDate(req.Msg.GetDate())
	if err != nil {
		return nil, toConnectError(ctx, s.log, err)
	}
	by := views.Dimension(req.Msg.GetBy())
	if _, dimErr := views.ParseDimension(by.String()); dimErr != nil {
		return nil, toConnectError(ctx, s.log, dimErr)
	}
	a, err := s.views.Allocation(ctx, on, by)
	if err != nil {
		return nil, toConnectError(ctx, s.log, err)
	}
	return connect.NewResponse(pbAllocation(&a)), nil
}
