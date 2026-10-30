package ledger

import (
	"fmt"

	"github.com/arkadiuss/openfief/internal/kernel"
)

// TransactionHeader holds what every transaction constructor needs besides
// the event's own amounts. Set a description on the result.
type TransactionHeader struct {
	Key     TransactionKey
	Account AccountID
	Date    kernel.Date
}

// transaction returns the transaction of kind with postings, not yet
// validated.
func (h TransactionHeader) transaction(kind TransactionKind, postings ...Posting) Transaction {
	return Transaction{Key: h.Key, Account: h.Account, Kind: kind, Date: h.Date, Postings: postings}
}

// build returns the validated transaction of kind with postings.
func (h TransactionHeader) build(kind TransactionKind, postings ...Posting) (Transaction, error) {
	t := h.transaction(kind, postings...)
	return validated(&t)
}

// validated returns t if it is valid.
func validated(t *Transaction) (Transaction, error) {
	if err := t.Validate(); err != nil {
		return Transaction{}, err
	}
	return *t, nil
}

// NewBuy books the purchase of qty units of in for cost, plus an optional
// fee: the holding gains the units at cost, cash pays cost and the fee. A
// zero Money fee means no fee; it may be in another currency than cost.
func NewBuy(h TransactionHeader, in Instrument, qty kernel.Quantity, cost, fee kernel.Money) (Transaction, error) {
	if err := needPositive(h, "quantity", qty.Decimal()); err != nil {
		return Transaction{}, err
	}
	if err := needNotNegative(h, "cost", cost); err != nil {
		return Transaction{}, err
	}
	fees, err := feePostings(h, fee)
	if err != nil {
		return Transaction{}, err
	}
	return h.build(KindBuy, append([]Posting{
		{Book: HoldingBook(in), Amount: cost, Quantity: qty},
		{Book: CashBook(cost.Currency()), Amount: cost.Neg()},
	}, fees...)...)
}

// NewSell books the sale of qty units of in for proceeds, less an optional
// fee: the holding gives up the units at the proceeds' value and cash
// receives the proceeds and pays the fee. Which lots the units come from is
// derived later, FIFO.
func NewSell(h TransactionHeader, in Instrument, qty kernel.Quantity, proceeds, fee kernel.Money) (Transaction, error) {
	if err := needPositive(h, "quantity", qty.Decimal()); err != nil {
		return Transaction{}, err
	}
	if err := needNotNegative(h, "proceeds", proceeds); err != nil {
		return Transaction{}, err
	}
	fees, err := feePostings(h, fee)
	if err != nil {
		return Transaction{}, err
	}
	return h.build(KindSell, append([]Posting{
		{Book: HoldingBook(in), Amount: proceeds.Neg(), Quantity: qty.Neg()},
		{Book: CashBook(proceeds.Currency()), Amount: proceeds},
	}, fees...)...)
}

// NewDividend books a dividend of gross on in, of which withholding was
// withheld at source: cash receives the net, withholding tax the withheld
// part, and the holding is the income's source with no change in units. A
// zero Money withholding means none was withheld.
func NewDividend(h TransactionHeader, in Instrument, gross, withholding kernel.Money) (Transaction, error) {
	postings, err := incomePostings(h, "gross dividend", in, gross, withholding)
	if err != nil {
		return Transaction{}, err
	}
	return h.build(KindDividend, postings...)
}

// NewInterest books interest of gross paid on in, such as a treasury bond's
// coupon (D-101), of which withholding was withheld at source, such as
// Belka tax: posted like a dividend (D-19), cash receives the net,
// withholding tax the withheld part, and the holding is the income's source
// with no change in units. A zero Money withholding means none was withheld.
func NewInterest(h TransactionHeader, in Instrument, gross, withholding kernel.Money) (Transaction, error) {
	postings, err := incomePostings(h, "gross interest", in, gross, withholding)
	if err != nil {
		return Transaction{}, err
	}
	return h.build(KindInterest, postings...)
}

// incomePostings returns the postings of income gross on in, named what, of
// which withholding was withheld at source.
func incomePostings(h TransactionHeader, what string, in Instrument, gross, withholding kernel.Money) ([]Posting, error) {
	if err := needPositive(h, what, gross.Amount()); err != nil {
		return nil, err
	}
	postings := []Posting{{Book: HoldingBook(in), Amount: gross.Neg()}}
	net := gross
	if !withholding.IsZero() {
		if err := needNotNegative(h, "withholding tax", withholding); err != nil {
			return nil, err
		}
		var err error
		if net, err = gross.Sub(withholding); err != nil {
			return nil, fmt.Errorf("%w %s: %w", ErrInvalidTransaction, h.Key, err)
		}
		if net.Sign() < 0 {
			return nil, fmt.Errorf("%w %s: withholding tax exceeds the %s", ErrInvalidTransaction, h.Key, what)
		}
		postings = append(postings, Posting{Book: WithholdingTaxBook(withholding.Currency()), Amount: withholding})
	}
	if !net.IsZero() {
		postings = append(postings, Posting{Book: CashBook(net.Currency()), Amount: net})
	}
	return postings, nil
}

// NewFee books a fee charged on its own, such as an account fee: the fee
// ledger account gains it and cash pays it.
func NewFee(h TransactionHeader, fee kernel.Money) (Transaction, error) {
	if err := needPositive(h, "fee", fee.Amount()); err != nil {
		return Transaction{}, err
	}
	fees, err := feePostings(h, fee)
	if err != nil {
		return Transaction{}, err
	}
	return h.build(KindFee, fees...)
}

// NewFXConversion books selling sold to buy bought, plus an optional fee.
// Each currency balances on its own through its FX ledger account: cash pays
// sold into FX, and FX pays bought into cash. No rate is stored; it is
// implied by the two amounts.
func NewFXConversion(h TransactionHeader, sold, bought, fee kernel.Money) (Transaction, error) {
	if err := needPositive(h, "sold amount", sold.Amount()); err != nil {
		return Transaction{}, err
	}
	if err := needPositive(h, "bought amount", bought.Amount()); err != nil {
		return Transaction{}, err
	}
	if sold.Currency().Equal(bought.Currency()) {
		return Transaction{}, fmt.Errorf("%w %s: conversion from %s to itself", ErrInvalidTransaction, h.Key, sold.Currency())
	}
	fees, err := feePostings(h, fee)
	if err != nil {
		return Transaction{}, err
	}
	return h.build(KindFXConversion, append([]Posting{
		{Book: CashBook(sold.Currency()), Amount: sold.Neg()},
		{Book: FXBook(sold.Currency()), Amount: sold},
		{Book: FXBook(bought.Currency()), Amount: bought.Neg()},
		{Book: CashBook(bought.Currency()), Amount: bought},
	}, fees...)...)
}

// NewDeposit books cash paid into the account from outside it.
func NewDeposit(h TransactionHeader, amount kernel.Money) (Transaction, error) {
	if err := needPositive(h, "deposit", amount.Amount()); err != nil {
		return Transaction{}, err
	}
	return h.build(KindDeposit,
		Posting{Book: CashBook(amount.Currency()), Amount: amount},
		Posting{Book: EquityBook(amount.Currency()), Amount: amount.Neg()},
	)
}

// NewWithdrawal books cash paid out of the account.
func NewWithdrawal(h TransactionHeader, amount kernel.Money) (Transaction, error) {
	if err := needPositive(h, "withdrawal", amount.Amount()); err != nil {
		return Transaction{}, err
	}
	return h.build(KindWithdrawal,
		Posting{Book: CashBook(amount.Currency()), Amount: amount.Neg()},
		Posting{Book: EquityBook(amount.Currency()), Amount: amount},
	)
}

// NewOpeningLot books a lot seeded from a broker's current lot-level
// positions at go-live (D-23): qty units of in, first acquired on opened at
// cost. The header's date is the go-live date the lot is recorded on; equity
// is the counterpart.
func NewOpeningLot(h TransactionHeader, in Instrument, qty kernel.Quantity, cost kernel.Money, opened kernel.Date) (Transaction, error) {
	if err := needPositive(h, "quantity", qty.Decimal()); err != nil {
		return Transaction{}, err
	}
	if err := needNotNegative(h, "cost", cost); err != nil {
		return Transaction{}, err
	}
	t := h.transaction(KindOpeningLot,
		Posting{Book: HoldingBook(in), Amount: cost, Quantity: qty},
		Posting{Book: EquityBook(cost.Currency()), Amount: cost.Neg()},
	)
	t.LotOpened = opened
	return validated(&t)
}

// NewDeletion books the deletion of the transaction the header's key names,
// as a correction (D-100): a new version of that key with no postings, which
// supersedes the current version and keeps it in the log (D-18). Holdings,
// lots and balances then leave the transaction out. Date it on the deleted
// transaction's date.
func NewDeletion(h TransactionHeader) (Transaction, error) {
	return h.build(KindDeletion)
}

// feePostings returns the postings that pay fee from cash, or none for a zero
// fee.
func feePostings(h TransactionHeader, fee kernel.Money) ([]Posting, error) {
	if fee.IsZero() {
		return nil, nil
	}
	if err := needNotNegative(h, "fee", fee); err != nil {
		return nil, err
	}
	return []Posting{
		{Book: FeeBook(fee.Currency()), Amount: fee},
		{Book: CashBook(fee.Currency()), Amount: fee.Neg()},
	}, nil
}

// needPositive fails unless d > 0.
func needPositive(h TransactionHeader, what string, d kernel.Decimal) error {
	if d.Sign() <= 0 {
		return fmt.Errorf("%w %s: %s must be positive, got %s", ErrInvalidTransaction, h.Key, what, d)
	}
	return nil
}

// needNotNegative fails unless m >= 0 and has a currency.
func needNotNegative(h TransactionHeader, what string, m kernel.Money) error {
	if m.Sign() < 0 || m.Currency().IsZero() {
		return fmt.Errorf("%w %s: %s must be a non-negative amount with a currency", ErrInvalidTransaction, h.Key, what)
	}
	return nil
}
