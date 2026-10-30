package ghostfolio_test

import (
	"context"
	"fmt"
	"slices"
	"strconv"

	"connectrpc.com/connect"

	viewsv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/views/v1"
)

// fakeRecord serves a synthetic record the way the read API does: accounts,
// pages of transactions, valuation statuses and valuations.
type fakeRecord struct {
	accounts     []*viewsv1.Account
	transactions map[string][]*viewsv1.Transaction
	through      map[string]string
	valuations   map[string][]*viewsv1.Valuation
	pageSize     int
}

func (f *fakeRecord) ListAccounts(context.Context, *connect.Request[viewsv1.ListAccountsRequest]) (*connect.Response[viewsv1.ListAccountsResponse], error) {
	return connect.NewResponse(&viewsv1.ListAccountsResponse{Accounts: f.accounts}), nil
}

// ListTransactions serves pages of f.pageSize, the page token being the
// index of the first transaction of the page.
func (f *fakeRecord) ListTransactions(_ context.Context, req *connect.Request[viewsv1.ListTransactionsRequest]) (*connect.Response[viewsv1.ListTransactionsResponse], error) {
	all := f.transactions[req.Msg.GetAccountId()]
	start := 0
	if tok := req.Msg.GetPageToken(); tok != "" {
		var err error
		if start, err = strconv.Atoi(tok); err != nil {
			return nil, fmt.Errorf("page token: %w", err)
		}
	}
	end := min(start+f.pageSize, len(all))
	resp := &viewsv1.ListTransactionsResponse{Transactions: all[start:end]}
	if end < len(all) {
		resp.NextPageToken = strconv.Itoa(end)
	}
	return connect.NewResponse(resp), nil
}

func (f *fakeRecord) GetValuationStatus(_ context.Context, req *connect.Request[viewsv1.GetValuationStatusRequest]) (*connect.Response[viewsv1.GetValuationStatusResponse], error) {
	id := req.Msg.GetAccountId()
	return connect.NewResponse(&viewsv1.GetValuationStatusResponse{Status: &viewsv1.ValuationStatus{AccountId: id, ValuedThrough: f.through[id]}}), nil
}

// ListAccountValuations serves the valuations dated from req's from
// through its to.
func (f *fakeRecord) ListAccountValuations(_ context.Context, req *connect.Request[viewsv1.ListAccountValuationsRequest]) (*connect.Response[viewsv1.ListAccountValuationsResponse], error) {
	var out []*viewsv1.Valuation
	for _, v := range f.valuations[req.Msg.GetAccountId()] {
		if v.GetDate() >= req.Msg.GetFrom() && v.GetDate() <= req.Msg.GetTo() {
			out = append(out, v)
		}
	}
	return connect.NewResponse(&viewsv1.ListAccountValuationsResponse{AccountId: req.Msg.GetAccountId(), Valuations: out}), nil
}

// Read API message builders for the synthetic record.
func pbMoney(amount, currency string) *viewsv1.Money {
	return &viewsv1.Money{Amount: amount, Currency: currency}
}

// pbBond is a posting of quantity units of the synthetic bond for amount
// PLN.
func pbBond(amount, quantity string) *viewsv1.Posting {
	return &viewsv1.Posting{
		Book:   &viewsv1.Book{Kind: viewsv1.BookKind_BOOK_KIND_HOLDING, Instrument: &viewsv1.Instrument{Isin: isinPL, Symbol: "EDO0136"}},
		Amount: pbMoney(amount, "PLN"), Quantity: quantity,
	}
}

// pbBook is a posting of amount PLN to a ledger account of kind.
func pbBook(kind viewsv1.BookKind, amount string) *viewsv1.Posting {
	return &viewsv1.Posting{Book: &viewsv1.Book{Kind: kind, Currency: "PLN"}, Amount: pbMoney(amount, "PLN")}
}

// The synthetic record's accounts.
const (
	bondAccount = "acct-bonds"
	cashAccount = "acct-cash"
)

// syntheticRecord is a record of two accounts: a bond account with an
// opening lot, a buy and a fee of its own, valued daily, and a cash
// account with a deposit only.
func syntheticRecord() *fakeRecord {
	cash := viewsv1.BookKind_BOOK_KIND_CASH
	equity := viewsv1.BookKind_BOOK_KIND_EQUITY
	return &fakeRecord{
		pageSize: 2,
		accounts: []*viewsv1.Account{
			{Id: bondAccount, Institution: "bond-registry", Name: "Synthetic bonds", Sleeve: "core"},
			{Id: cashAccount, Institution: "bank", Name: "Synthetic cash", Sleeve: "core"},
		},
		transactions: map[string][]*viewsv1.Transaction{
			bondAccount: {
				{
					Source: "test/bonds", ExternalId: "open-1", Kind: viewsv1.TransactionKind_TRANSACTION_KIND_OPENING_LOT,
					Date: "2025-01-02", LotOpened: "2024-11-04", Postings: []*viewsv1.Posting{
						pbBond("1000.00", "10"), pbBook(equity, "-1000.00"),
					},
				},
				{
					Source: "test/bonds", ExternalId: "buy-1", Kind: viewsv1.TransactionKind_TRANSACTION_KIND_BUY,
					Date: "2025-01-03", Postings: []*viewsv1.Posting{
						pbBond("500.00", "5"), pbBook(cash, "-500.00"),
					},
				},
				{
					Source: "test/bonds", ExternalId: "fee-1", Kind: viewsv1.TransactionKind_TRANSACTION_KIND_FEE,
					Date: "2025-01-06", Postings: []*viewsv1.Posting{
						pbBook(viewsv1.BookKind_BOOK_KIND_FEE, "2.00"), pbBook(cash, "-2.00"),
					},
				},
			},
			cashAccount: {
				{
					Source: "test/cash", ExternalId: "dep-1", Kind: viewsv1.TransactionKind_TRANSACTION_KIND_DEPOSIT,
					Date: "2025-01-02", Postings: []*viewsv1.Posting{pbBook(cash, "100.00"), pbBook(equity, "-100.00")},
				},
			},
		},
		through: map[string]string{bondAccount: "2025-01-06", cashAccount: "2025-01-06"},
		valuations: map[string][]*viewsv1.Valuation{
			bondAccount: {
				bondValuation("2025-01-02", "1004.10"),
				bondValuation("2025-01-03", "1506.75"),
				bondValuation("2025-01-06", "1507.95"),
			},
		},
	}
}

// bondValuation is the stored value of the synthetic bond on day.
func bondValuation(day, value string) *viewsv1.Valuation {
	return &viewsv1.Valuation{
		Date: day, Instrument: &viewsv1.Instrument{Isin: isinPL, Symbol: "EDO0136"},
		Basis: viewsv1.ValuationBasis_VALUATION_BASIS_GROSS_ACCRUED, Source: "test", Value: pbMoney(value, "PLN"),
	}
}

// withoutBond removes the bond account's transaction whose external ID is
// ext from f.
func (f *fakeRecord) withoutBond(ext string) {
	f.transactions[bondAccount] = slices.DeleteFunc(slices.Clone(f.transactions[bondAccount]), func(t *viewsv1.Transaction) bool {
		return t.GetExternalId() == ext
	})
}
