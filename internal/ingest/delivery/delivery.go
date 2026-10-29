// Package delivery is the connector side of the connector contract
// (D-49): it writes a connector's mapping, prices, FX rates, index
// readings, index levels and ingest mode as openfief.connector.v1 messages, the reverse of what the core's
// internal/ingest/connectrpc reads. First-party connector plugins under
// plugins/ use it, so their institution adapters stay in internal/ and
// their binaries stay thin.
//
// Every amount, rate and quantity is written as an exact decimal string,
// every date as YYYY-MM-DD, as the contract requires.
package delivery

import (
	"errors"
	"fmt"

	"github.com/arkadiuss/openfief/internal/ingest"
	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
	"github.com/arkadiuss/openfief/internal/valuation"
	connectorv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/connector/v1"
)

// Errors writing a delivery.
var (
	// ErrUnknownValue reports a domain value with no contract counterpart,
	// such as the zero transaction kind.
	ErrUnknownValue = errors.New("value has no connector contract form")
	// ErrNoMode reports a request without a known ingest mode.
	ErrNoMode = errors.New("no ingest mode")
)

// Mode reads the ingest mode of a Sync or ImportFile request.
func Mode(m connectorv1.Mode) (ingest.Mode, error) {
	switch m {
	case connectorv1.Mode_MODE_OPENING:
		return ingest.ModeOpening, nil
	case connectorv1.Mode_MODE_TRANSACTIONS:
		return ingest.ModeTransactions, nil
	case connectorv1.Mode_MODE_UNSPECIFIED:
	}
	return 0, fmt.Errorf("%w: %s", ErrNoMode, m)
}

// FromMapping writes a mapping as a delivery for batch from institution:
// its transactions, balance assertions and ingest problems, in order.
func FromMapping(batch, institution string, m *ingest.Mapping) (*connectorv1.Delivery, error) {
	d := &connectorv1.Delivery{BatchId: batch, Institution: institution}
	for i := range m.Transactions {
		t, err := transaction(&m.Transactions[i])
		if err != nil {
			return nil, fmt.Errorf("transactions[%d]: %w", i, err)
		}
		d.Transactions = append(d.Transactions, t)
	}
	for i := range m.Assertions {
		a, err := assertion(&m.Assertions[i])
		if err != nil {
			return nil, fmt.Errorf("balance_assertions[%d]: %w", i, err)
		}
		d.BalanceAssertions = append(d.BalanceAssertions, a)
	}
	for i, p := range m.Problems {
		pb, err := problem(p)
		if err != nil {
			return nil, fmt.Errorf("problems[%d]: %w", i, err)
		}
		d.Problems = append(d.Problems, pb)
	}
	return d, nil
}

// FXRates writes FX rates to PLN, in order. Their source is the
// connector's own and is not part of the contract.
func FXRates(rates []valuation.FXRate) []*connectorv1.FxRate {
	out := make([]*connectorv1.FxRate, 0, len(rates))
	for _, r := range rates {
		out = append(out, &connectorv1.FxRate{Currency: r.Currency.String(), Date: r.Date.String(), Rate: r.Rate.String()})
	}
	return out
}

// Price writes the end-of-day close of one unit of in on a date (D-21), in
// the currency the instrument is quoted in. Its source is the connector's
// own and is not part of the contract.
func Price(in ledger.Instrument, on kernel.Date, closing kernel.Money) *connectorv1.Price {
	return &connectorv1.Price{
		Instrument: &connectorv1.Instrument{Isin: in.ISIN.String(), Symbol: in.Symbol},
		Date:       on.String(), Close: money(closing),
	}
}

// IndexLevel writes the closing level of the market index named index on a
// date.
func IndexLevel(index string, on kernel.Date, level kernel.Decimal) *connectorv1.IndexLevel {
	return &connectorv1.IndexLevel{Index: index, Date: on.String(), Level: level.String()}
}

// IndexReadings writes monthly period index readings (D-9), in order, each
// dated on the first day of its month: for CPI the month it measures, for
// the NBP reference rate the month whose interest periods it sets. Their
// source is the connector's own and is not part of the contract.
func IndexReadings(readings []valuation.IndexReading) ([]*connectorv1.IndexReading, error) {
	out := make([]*connectorv1.IndexReading, 0, len(readings))
	for i := range readings {
		r := &readings[i]
		var index connectorv1.Index
		switch r.Kind {
		case valuation.CPI:
			index = connectorv1.Index_INDEX_CPI
		case valuation.ReferenceRate:
			index = connectorv1.Index_INDEX_REFERENCE_RATE
		default:
			return nil, fmt.Errorf("index_readings[%d]: %w: index %s", i, ErrUnknownValue, r.Kind)
		}
		first, err := kernel.NewDate(r.Year, r.Month, 1)
		if err != nil {
			return nil, fmt.Errorf("index_readings[%d]: %w", i, err)
		}
		out = append(out, &connectorv1.IndexReading{Index: index, Date: first.String(), Value: r.Value.String()})
	}
	return out, nil
}

// transaction writes one transaction.
func transaction(t *ledger.Transaction) (*connectorv1.Transaction, error) {
	kind, err := transactionKind(t.Kind)
	if err != nil {
		return nil, err
	}
	out := &connectorv1.Transaction{
		Key:       &connectorv1.TransactionKey{Source: t.Key.Source, ExternalId: t.Key.ExternalID},
		AccountId: string(t.Account), Kind: kind, Date: t.Date.String(),
		LotOpened: optionalDate(t.LotOpened), Description: t.Description,
	}
	for i := range t.Postings {
		p, err := posting(&t.Postings[i])
		if err != nil {
			return nil, fmt.Errorf("postings[%d]: %w", i, err)
		}
		out.Postings = append(out.Postings, p)
	}
	return out, nil
}

// posting writes one posting; only a holding posting carries a quantity.
func posting(p *ledger.Posting) (*connectorv1.Posting, error) {
	b, err := book(p.Book)
	if err != nil {
		return nil, err
	}
	out := &connectorv1.Posting{Book: b, Amount: money(p.Amount)}
	if p.Book.Kind == ledger.BookHolding {
		out.Quantity = p.Quantity.String()
	}
	return out, nil
}

// assertion writes one balance assertion.
func assertion(a *ledger.BalanceAssertion) (*connectorv1.BalanceAssertion, error) {
	b, err := book(a.Book)
	if err != nil {
		return nil, err
	}
	return &connectorv1.BalanceAssertion{
		AccountId: string(a.Account), Date: a.Date.String(), Book: b, Balance: a.Balance.String(),
	}, nil
}

// book writes a ledger account: a holding names its instrument, any other
// kind its currency.
func book(b ledger.Book) (*connectorv1.Book, error) {
	kind, err := bookKind(b.Kind)
	if err != nil {
		return nil, err
	}
	if b.Kind == ledger.BookHolding {
		return &connectorv1.Book{Kind: kind, Instrument: &connectorv1.Instrument{
			Isin: b.Instrument.ISIN.String(), Symbol: b.Instrument.Symbol,
		}}, nil
	}
	return &connectorv1.Book{Kind: kind, Currency: b.Currency.String()}, nil
}

// money writes an amount with its currency.
func money(m kernel.Money) *connectorv1.Money {
	return &connectorv1.Money{Amount: m.Amount().String(), Currency: m.Currency().String()}
}

// optionalDate writes a date, or "" for the zero date.
func optionalDate(d kernel.Date) string {
	if d.IsZero() {
		return ""
	}
	return d.String()
}

// bookKind writes a ledger account kind.
func bookKind(k ledger.BookKind) (connectorv1.BookKind, error) {
	switch k {
	case ledger.BookCash:
		return connectorv1.BookKind_BOOK_KIND_CASH, nil
	case ledger.BookHolding:
		return connectorv1.BookKind_BOOK_KIND_HOLDING, nil
	case ledger.BookFee:
		return connectorv1.BookKind_BOOK_KIND_FEE, nil
	case ledger.BookWithholdingTax:
		return connectorv1.BookKind_BOOK_KIND_WITHHOLDING_TAX, nil
	case ledger.BookFX:
		return connectorv1.BookKind_BOOK_KIND_FX, nil
	case ledger.BookEquity:
		return connectorv1.BookKind_BOOK_KIND_EQUITY, nil
	}
	return 0, fmt.Errorf("%w: ledger account kind %d", ErrUnknownValue, k)
}

// transactionKind writes a transaction kind.
func transactionKind(k ledger.TransactionKind) (connectorv1.TransactionKind, error) {
	switch k {
	case ledger.KindBuy:
		return connectorv1.TransactionKind_TRANSACTION_KIND_BUY, nil
	case ledger.KindSell:
		return connectorv1.TransactionKind_TRANSACTION_KIND_SELL, nil
	case ledger.KindDividend:
		return connectorv1.TransactionKind_TRANSACTION_KIND_DIVIDEND, nil
	case ledger.KindFee:
		return connectorv1.TransactionKind_TRANSACTION_KIND_FEE, nil
	case ledger.KindFXConversion:
		return connectorv1.TransactionKind_TRANSACTION_KIND_FX_CONVERSION, nil
	case ledger.KindDeposit:
		return connectorv1.TransactionKind_TRANSACTION_KIND_DEPOSIT, nil
	case ledger.KindWithdrawal:
		return connectorv1.TransactionKind_TRANSACTION_KIND_WITHDRAWAL, nil
	case ledger.KindOpeningLot:
		return connectorv1.TransactionKind_TRANSACTION_KIND_OPENING_LOT, nil
	}
	return 0, fmt.Errorf("%w: transaction kind %d", ErrUnknownValue, k)
}

// problem writes one ingest problem.
func problem(p ingest.Problem) (*connectorv1.Problem, error) {
	var kind connectorv1.ProblemKind
	switch p.Kind {
	case ingest.ProblemUnsupported:
		kind = connectorv1.ProblemKind_PROBLEM_KIND_UNSUPPORTED
	case ingest.ProblemSkipped:
		kind = connectorv1.ProblemKind_PROBLEM_KIND_SKIPPED
	default:
		return nil, fmt.Errorf("%w: problem kind %d", ErrUnknownValue, p.Kind)
	}
	return &connectorv1.Problem{Kind: kind, Ref: p.Ref, Reason: p.Reason}, nil
}
