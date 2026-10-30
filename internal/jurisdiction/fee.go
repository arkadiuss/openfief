package jurisdiction

import (
	"errors"
	"fmt"

	"github.com/arkadiuss/openfief/internal/kernel"
)

// ErrInvalidFeeRule reports a fee rule whose rates are not fractions from 0
// up to 1, or whose minimum is negative or has no currency.
var ErrInvalidFeeRule = errors.New("invalid fee rule")

// ErrFeeCurrency reports a fee rule's minimum in another currency than the
// value it is charged on.
var ErrFeeCurrency = errors.New("the fee rule's minimum is in another currency than the value")

// FeeRulePlaces is how many decimal places a sale cost is rounded to, half
// up: the minor unit of the pack's tax currency (grosze for PLN).
const FeeRulePlaces = 2

// FeeRule is how an account's broker charges a sale (D-78): Rate of the
// value sold, at least Minimum, plus ConversionRate of the value for an
// instrument in a foreign currency, which the sale's proceeds are converted
// from. Rates are fractions: 0.29% is 0.0029. The zero rule, which every
// account has until the owner sets one, charges nothing.
type FeeRule struct {
	Rate           kernel.Decimal
	Minimum        kernel.Money
	ConversionRate kernel.Decimal
}

// IsZero reports whether r charges nothing.
func (r *FeeRule) IsZero() bool {
	return r.Rate.IsZero() && r.Minimum.IsZero() && r.ConversionRate.IsZero()
}

// Validate reports whether both rates are fractions from 0 up to but not
// including 1 and the minimum is not negative, with a currency unless it
// is zero.
func (r *FeeRule) Validate() error {
	one := kernel.NewDecimal(1, 0)
	switch {
	case r.Rate.Sign() < 0 || r.Rate.Cmp(one) >= 0:
		return fmt.Errorf("%w: rate %s is not from 0 to below 1", ErrInvalidFeeRule, r.Rate)
	case r.ConversionRate.Sign() < 0 || r.ConversionRate.Cmp(one) >= 0:
		return fmt.Errorf("%w: conversion rate %s is not from 0 to below 1", ErrInvalidFeeRule, r.ConversionRate)
	case r.Minimum.Sign() < 0:
		return fmt.Errorf("%w: minimum %s is negative", ErrInvalidFeeRule, r.Minimum)
	case !r.Minimum.IsZero() && r.Minimum.Currency().IsZero():
		return fmt.Errorf("%w: minimum %s has no currency", ErrInvalidFeeRule, r.Minimum.Amount())
	}
	return nil
}

// SaleCost returns what selling a holding worth gross would cost under r:
// the commission, Rate of gross but at least the minimum, plus, when
// foreign, ConversionRate of gross; rounded half up to FeeRulePlaces.
// minimum is r.Minimum converted into gross's currency by the caller, who
// knows the rates. Nothing is charged on a gross value that is not
// positive, and the zero rule charges nothing.
func (r *FeeRule) SaleCost(gross, minimum kernel.Money, foreign bool) (kernel.Money, error) {
	zero := kernel.NewMoney(kernel.Decimal{}, gross.Currency())
	if gross.Sign() <= 0 || r.IsZero() {
		return zero, nil
	}
	if !minimum.IsZero() && !minimum.Currency().Equal(gross.Currency()) {
		return kernel.Money{}, fmt.Errorf("%w: %s, not %s", ErrFeeCurrency, minimum, gross.Currency())
	}
	commission := gross.Amount().Mul(r.Rate)
	if commission.Cmp(minimum.Amount()) < 0 {
		commission = minimum.Amount()
	}
	if foreign {
		commission = commission.Add(gross.Amount().Mul(r.ConversionRate))
	}
	cost, err := commission.Round(FeeRulePlaces, kernel.RoundHalfUp)
	if err != nil {
		return kernel.Money{}, fmt.Errorf("sale cost: %w", err)
	}
	return kernel.NewMoney(cost, gross.Currency()), nil
}
