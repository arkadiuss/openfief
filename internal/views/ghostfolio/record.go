package ghostfolio

import (
	"context"
	"errors"
	"fmt"

	"connectrpc.com/connect"

	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
	"github.com/arkadiuss/openfief/internal/views"
	viewsv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/views/v1"
)

// Record is the part of the read API the view reads (D-6): accounts, their
// transactions, and the stored valuations of their bonds. The generated
// viewsv1connect.ReadServiceClient implements it.
type Record interface {
	ListAccounts(ctx context.Context, req *connect.Request[viewsv1.ListAccountsRequest]) (*connect.Response[viewsv1.ListAccountsResponse], error)
	ListTransactions(ctx context.Context, req *connect.Request[viewsv1.ListTransactionsRequest]) (*connect.Response[viewsv1.ListTransactionsResponse], error)
	GetValuationStatus(ctx context.Context, req *connect.Request[viewsv1.GetValuationStatusRequest]) (*connect.Response[viewsv1.GetValuationStatusResponse], error)
	ListAccountValuations(ctx context.Context, req *connect.Request[viewsv1.ListAccountValuationsRequest]) (*connect.Response[viewsv1.ListAccountValuationsResponse], error)
}

// ErrBadRecord reports a read API answer the view can't read.
var ErrBadRecord = errors.New("unreadable read API answer")

// transactionPage is how many transactions ReadRecord asks for at a time,
// the read API's largest page.
const transactionPage = 500

// ReadRecord reads every account of the record with its current
// transactions and, when it ever held a bond, its stored valuations from
// its first transaction through the date it is valued through.
func ReadRecord(ctx context.Context, r Record) ([]SourceAccount, error) {
	resp, err := r.ListAccounts(ctx, connect.NewRequest(&viewsv1.ListAccountsRequest{}))
	if err != nil {
		return nil, fmt.Errorf("listing accounts: %w", err)
	}
	out := make([]SourceAccount, 0, len(resp.Msg.GetAccounts()))
	for _, a := range resp.Msg.GetAccounts() {
		sa := SourceAccount{ID: ledger.AccountID(a.GetId()), Name: a.GetName()}
		if sa.Transactions, err = readTransactions(ctx, r, a.GetId()); err != nil {
			return nil, err
		}
		if sa.Valuations, err = readBondValuations(ctx, r, &sa); err != nil {
			return nil, err
		}
		out = append(out, sa)
	}
	return out, nil
}

// readTransactions reads every page of an account's transactions.
func readTransactions(ctx context.Context, r Record, account string) ([]ledger.Transaction, error) {
	var out []ledger.Transaction
	token := ""
	for {
		resp, err := r.ListTransactions(ctx, connect.NewRequest(&viewsv1.ListTransactionsRequest{
			AccountId: account, PageSize: transactionPage, PageToken: token,
		}))
		if err != nil {
			return nil, fmt.Errorf("listing transactions: %w", err)
		}
		for _, t := range resp.Msg.GetTransactions() {
			lt, err := transactionOf(account, t)
			if err != nil {
				return nil, err
			}
			out = append(out, lt)
		}
		if token = resp.Msg.GetNextPageToken(); token == "" {
			return out, nil
		}
	}
}

// readBondValuations reads a's stored valuations of bonds, from its first
// transaction through its valued-through date, in ranges the read API
// accepts. An account that never held a bond, or was never valued, has
// none.
func readBondValuations(ctx context.Context, r Record, a *SourceAccount) ([]Valuation, error) {
	first := firstBondDate(a.Transactions)
	if first.IsZero() {
		return nil, nil
	}
	st, err := r.GetValuationStatus(ctx, connect.NewRequest(&viewsv1.GetValuationStatusRequest{AccountId: string(a.ID)}))
	if err != nil {
		return nil, fmt.Errorf("reading the valuation status: %w", err)
	}
	through := st.Msg.GetStatus().GetValuedThrough()
	if through == "" {
		return nil, nil
	}
	last, err := kernel.ParseDate(through)
	if err != nil {
		return nil, fmt.Errorf("%w: valued through: %w", ErrBadRecord, err)
	}
	var out []Valuation
	for from := first; !from.After(last); from = from.AddDays(views.MaxRangeDays + 1) {
		to := from.AddDays(views.MaxRangeDays)
		if to.After(last) {
			to = last
		}
		vals, err := readValuations(ctx, r, a.ID, from, to)
		if err != nil {
			return nil, err
		}
		out = append(out, vals...)
	}
	return out, nil
}

// firstBondDate returns the date of the first of ts that posts to a bond's
// holding, or the zero date when none does.
func firstBondDate(ts []ledger.Transaction) kernel.Date {
	var first kernel.Date
	for i := range ts {
		t := &ts[i]
		if (first.IsZero() || t.Date.Before(first)) && postsToBond(t) {
			first = t.Date
		}
	}
	return first
}

// postsToBond reports whether t posts to a bond's holding.
func postsToBond(t *ledger.Transaction) bool {
	for i := range t.Postings {
		p := &t.Postings[i]
		if p.Book.Kind == ledger.BookHolding && views.AssetClassOf(p.Book.Instrument) == views.AssetTreasuryBond {
			return true
		}
	}
	return false
}

// readValuations reads account's stored valuations of bonds dated from
// through to.
func readValuations(ctx context.Context, r Record, account ledger.AccountID, from, to kernel.Date) ([]Valuation, error) {
	resp, err := r.ListAccountValuations(ctx, connect.NewRequest(&viewsv1.ListAccountValuationsRequest{
		AccountId: string(account), From: from.String(), To: to.String(),
	}))
	if err != nil {
		return nil, fmt.Errorf("listing valuations: %w", err)
	}
	var out []Valuation
	for _, v := range resp.Msg.GetValuations() {
		val, err := valuationOf(v)
		if err != nil {
			return nil, err
		}
		if views.AssetClassOf(val.Instrument) == views.AssetTreasuryBond {
			out = append(out, val)
		}
	}
	return out, nil
}

// transactionOf converts a read API transaction of account.
func transactionOf(account string, t *viewsv1.Transaction) (ledger.Transaction, error) {
	lt := ledger.Transaction{
		Key:         ledger.TransactionKey{Source: t.GetSource(), ExternalID: t.GetExternalId()},
		Account:     ledger.AccountID(account),
		Description: t.GetDescription(),
	}
	var err error
	if lt.Kind, err = kindOf(t.GetKind()); err != nil {
		return ledger.Transaction{}, err
	}
	if lt.Date, err = kernel.ParseDate(t.GetDate()); err != nil {
		return ledger.Transaction{}, fmt.Errorf("%w: transaction date: %w", ErrBadRecord, err)
	}
	if t.GetLotOpened() != "" {
		if lt.LotOpened, err = kernel.ParseDate(t.GetLotOpened()); err != nil {
			return ledger.Transaction{}, fmt.Errorf("%w: lot opened: %w", ErrBadRecord, err)
		}
	}
	for _, p := range t.GetPostings() {
		lp, err := postingOf(p)
		if err != nil {
			return ledger.Transaction{}, err
		}
		lt.Postings = append(lt.Postings, lp)
	}
	return lt, nil
}

// postingOf converts a read API posting.
func postingOf(p *viewsv1.Posting) (ledger.Posting, error) {
	amount, err := moneyOf(p.GetAmount())
	if err != nil {
		return ledger.Posting{}, err
	}
	lp := ledger.Posting{Amount: amount}
	if lp.Book.Kind, err = bookKindOf(p.GetBook().GetKind()); err != nil {
		return ledger.Posting{}, err
	}
	if lp.Book.Kind == ledger.BookHolding {
		if lp.Book.Instrument, err = instrumentOf(p.GetBook().GetInstrument()); err != nil {
			return ledger.Posting{}, err
		}
		if lp.Quantity, err = kernel.ParseQuantity(p.GetQuantity()); err != nil {
			return ledger.Posting{}, fmt.Errorf("%w: quantity: %w", ErrBadRecord, err)
		}
		return lp, nil
	}
	if lp.Book.Currency, err = kernel.ParseCurrency(p.GetBook().GetCurrency()); err != nil {
		return ledger.Posting{}, fmt.Errorf("%w: ledger account currency: %w", ErrBadRecord, err)
	}
	return lp, nil
}

// valuationOf converts a read API valuation.
func valuationOf(v *viewsv1.Valuation) (Valuation, error) {
	d, err := kernel.ParseDate(v.GetDate())
	if err != nil {
		return Valuation{}, fmt.Errorf("%w: valuation date: %w", ErrBadRecord, err)
	}
	in, err := instrumentOf(v.GetInstrument())
	if err != nil {
		return Valuation{}, err
	}
	m, err := moneyOf(v.GetValue())
	if err != nil {
		return Valuation{}, err
	}
	return Valuation{Date: d, Instrument: in, Value: m}, nil
}

// instrumentOf converts a read API instrument.
func instrumentOf(in *viewsv1.Instrument) (ledger.Instrument, error) {
	isin, err := ledger.ParseISIN(in.GetIsin())
	if err != nil {
		return ledger.Instrument{}, fmt.Errorf("%w: instrument: %w", ErrBadRecord, err)
	}
	return ledger.Instrument{ISIN: isin, Symbol: in.GetSymbol()}, nil
}

// moneyOf converts a read API amount.
func moneyOf(m *viewsv1.Money) (kernel.Money, error) {
	c, err := kernel.ParseCurrency(m.GetCurrency())
	if err != nil {
		return kernel.Money{}, fmt.Errorf("%w: currency: %w", ErrBadRecord, err)
	}
	out, err := kernel.ParseMoney(m.GetAmount(), c)
	if err != nil {
		return kernel.Money{}, fmt.Errorf("%w: amount: %w", ErrBadRecord, err)
	}
	return out, nil
}

// transactionKinds returns a map of every read API transaction kind to the ledger's.
func transactionKinds() map[viewsv1.TransactionKind]ledger.TransactionKind {
	return map[viewsv1.TransactionKind]ledger.TransactionKind{
		viewsv1.TransactionKind_TRANSACTION_KIND_BUY:           ledger.KindBuy,
		viewsv1.TransactionKind_TRANSACTION_KIND_SELL:          ledger.KindSell,
		viewsv1.TransactionKind_TRANSACTION_KIND_DIVIDEND:      ledger.KindDividend,
		viewsv1.TransactionKind_TRANSACTION_KIND_FEE:           ledger.KindFee,
		viewsv1.TransactionKind_TRANSACTION_KIND_FX_CONVERSION: ledger.KindFXConversion,
		viewsv1.TransactionKind_TRANSACTION_KIND_DEPOSIT:       ledger.KindDeposit,
		viewsv1.TransactionKind_TRANSACTION_KIND_WITHDRAWAL:    ledger.KindWithdrawal,
		viewsv1.TransactionKind_TRANSACTION_KIND_OPENING_LOT:   ledger.KindOpeningLot,
		viewsv1.TransactionKind_TRANSACTION_KIND_INTEREST:      ledger.KindInterest,
	}
}

// kindOf converts a read API transaction kind.
func kindOf(k viewsv1.TransactionKind) (ledger.TransactionKind, error) {
	if kind, ok := transactionKinds()[k]; ok {
		return kind, nil
	}
	return 0, fmt.Errorf("%w: transaction kind %s", ErrBadRecord, k)
}

// bookKindOf converts a read API ledger account kind.
func bookKindOf(k viewsv1.BookKind) (ledger.BookKind, error) {
	switch k {
	case viewsv1.BookKind_BOOK_KIND_CASH:
		return ledger.BookCash, nil
	case viewsv1.BookKind_BOOK_KIND_HOLDING:
		return ledger.BookHolding, nil
	case viewsv1.BookKind_BOOK_KIND_FEE:
		return ledger.BookFee, nil
	case viewsv1.BookKind_BOOK_KIND_WITHHOLDING_TAX:
		return ledger.BookWithholdingTax, nil
	case viewsv1.BookKind_BOOK_KIND_FX:
		return ledger.BookFX, nil
	case viewsv1.BookKind_BOOK_KIND_EQUITY:
		return ledger.BookEquity, nil
	case viewsv1.BookKind_BOOK_KIND_UNSPECIFIED:
	}
	return 0, fmt.Errorf("%w: ledger account kind %s", ErrBadRecord, k)
}
