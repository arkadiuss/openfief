package pl

import (
	"errors"
	"fmt"

	"github.com/arkadiuss/openfief/internal/kernel"
)

// ErrInvalidQuantity reports a bond quantity that is not a positive whole number.
var ErrInvalidQuantity = errors.New("bond quantity must be a positive whole number")

// grosze is the number of decimal places of a PLN amount.
const grosze = 2

// LotValue is the statutory value of a lot of one bond series on a date.
// Every amount is in PLN and covers the whole lot.
type LotValue struct {
	// Gross is the gross accrued value: nominal plus interest accrued and not
	// yet paid out (for coupon types, only the current period's interest).
	Gross kernel.Money
	// Interest is the accrued interest inside Gross.
	Interest kernel.Money
	// Fee is the early-redemption fee after its cap; zero at maturity.
	Fee kernel.Money
	// Tax is the Belka tax on Interest minus Fee; zero inside IKE and IKZE.
	Tax kernel.Money
	// Net is what redeeming the lot on the date would pay: Gross - Fee - Tax.
	Net kernel.Money
	// Matured reports a date on or after maturity; the value is then the
	// redemption value at maturity.
	Matured bool
}

// ValueLot returns the gross accrued value and the net-if-redeemed value on
// date on of quantity bonds of t.Series bought on purchased and held in
// wrapper. idx supplies the index readings of every period up to the one the
// date falls in.
//
// The value of one bond follows the issue letters (obligacjeskarbowe.pl,
// listy emisyjne): within interest period k, interest accrues as
// r_k × (period months / 12) × a / ACT, where a is the days from the period's
// first day to the date and ACT the actual days in the period. Capitalized
// types compound completed periods, W = N × Π(1 + r_n) × (1 + r_k × a/ACT),
// rounded once to grosze per bond; coupon types pay completed periods out, so
// a bond is worth N plus the current period's interest, rounded to grosze.
// Lot amounts are the per-bond amounts times the quantity.
//
// The early-redemption fee is capped at the bond's accrued interest, so it
// never eats into the nominal, except for coupon types after their first
// interest period, where the full fee is charged. Belka tax is due on the
// accrued interest minus the fee.
func ValueLot(t *Terms, idx Indices, quantity kernel.Quantity, purchased kernel.Date, wrapper TaxWrapper, on kernel.Date) (LotValue, error) {
	if err := t.validate(); err != nil {
		return LotValue{}, err
	}
	whole, err := quantity.Decimal().Round(0, kernel.RoundDown)
	if err != nil || quantity.Sign() <= 0 || !whole.Equal(quantity.Decimal()) {
		return LotValue{}, fmt.Errorf("%w: %s", ErrInvalidQuantity, quantity)
	}
	maturity, err := Maturity(t.Series, purchased)
	if err != nil {
		return LotValue{}, err
	}
	if on.Before(purchased) {
		return LotValue{}, fmt.Errorf("%w: %s bought on %s, valued on %s", ErrBeforePurchase, t.Series, purchased, on)
	}
	matured := !on.Before(maturity)
	if matured {
		on = maturity
	}
	acc, err := accrue(t, idx, purchased, on)
	if err != nil {
		return LotValue{}, err
	}
	var fee kernel.Decimal
	if !matured {
		fee = cappedFee(t, acc)
	}
	return lotValue(acc, fee, quantity.Decimal(), wrapper, matured)
}

// accrual is the value of one bond on a date, in PLN.
type accrual struct {
	value    kernel.Decimal // nominal plus unpaid interest, in grosze
	interest kernel.Decimal // value minus nominal
	period   int            // the interest period the date falls in
}

// cappedFee returns the early-redemption fee of one bond: at most its accrued
// interest, except for coupon types after the first interest period.
func cappedFee(t *Terms, acc accrual) kernel.Decimal {
	fee := t.EarlyRedemptionFee.Amount()
	st, err := t.Series.Type.Structure()
	if err == nil && st.Payout == Coupon && acc.period > 1 {
		return fee
	}
	if fee.Cmp(acc.interest) > 0 {
		return acc.interest
	}
	return fee
}

func lotValue(acc accrual, fee, q kernel.Decimal, wrapper TaxWrapper, matured bool) (LotValue, error) {
	gross, interest, fees := acc.value.Mul(q), acc.interest.Mul(q), fee.Mul(q)
	tax, err := belkaTax(interest.Sub(fees), wrapper)
	if err != nil {
		return LotValue{}, err
	}
	c := pln()
	return LotValue{
		Gross:    kernel.NewMoney(gross, c),
		Interest: kernel.NewMoney(interest, c),
		Fee:      kernel.NewMoney(fees, c),
		Tax:      kernel.NewMoney(tax, c),
		Net:      kernel.NewMoney(gross.Sub(fees).Sub(tax), c),
		Matured:  matured,
	}, nil
}

// period is one interest period of a bond: its number, first day, day count
// and yearly rate.
type period struct {
	n          int
	start, end kernel.Date
	rate       kernel.Decimal
}

func (t *Terms) period(idx Indices, purchased kernel.Date, n int) (period, error) {
	start, err := PeriodStart(t.Series.Type, purchased, n)
	if err != nil {
		return period{}, err
	}
	end, err := PeriodStart(t.Series.Type, purchased, n+1)
	if err != nil {
		return period{}, err
	}
	r, err := t.PeriodRate(idx, n, start)
	if err != nil {
		return period{}, err
	}
	return period{n: n, start: start, end: end, rate: r}, nil
}

// accrue returns the value of one bond bought on purchased, on date on, which
// must be on or before maturity. Every quantity stays exact until the single
// division that rounds the result to grosze.
func accrue(t *Terms, idx Indices, purchased, on kernel.Date) (accrual, error) {
	st, err := t.Series.Type.Structure()
	if err != nil {
		return accrual{}, err
	}
	periods := st.TenorMonths / st.PeriodMonths
	twelve := kernel.NewDecimal(monthsPerYear, 0)
	months := kernel.NewDecimal(int64(st.PeriodMonths), 0)
	// After the completed periods, a capitalized bond is worth growthNum / growthDen.
	growthNum, growthDen := Nominal().Amount(), kernel.NewDecimal(1, 0)
	for n := 1; n <= periods; n++ {
		p, err := t.period(idx, purchased, n)
		if err != nil {
			return accrual{}, err
		}
		if !on.Before(p.end) && n < periods {
			growthNum = growthNum.Mul(twelve.Add(p.rate.Mul(months)))
			growthDen = growthDen.Mul(twelve)
			continue
		}
		act := kernel.NewDecimal(int64(p.start.DaysUntil(p.end)), 0)
		a := kernel.NewDecimal(int64(p.start.DaysUntil(on)), 0)
		// Interest so far in period n per unit of base: r × months × a / (12 × ACT).
		partNum, partDen := p.rate.Mul(months).Mul(a), twelve.Mul(act)
		if st.Payout == Coupon {
			return roundAccrual(Nominal().Amount().Mul(partNum), partDen, n, false)
		}
		return roundAccrual(growthNum.Mul(partDen.Add(partNum)), growthDen.Mul(partDen), n, true)
	}
	return accrual{}, fmt.Errorf("%w: %s has no interest periods", ErrInvalidTerms, t.Series)
}

// roundAccrual rounds num / den half up to grosze. For a capitalized bond the
// quotient is the bond's value; for a coupon bond it is the current period's
// interest, added to the nominal.
func roundAccrual(num, den kernel.Decimal, n int, isValue bool) (accrual, error) {
	q, err := num.Quo(den, grosze, kernel.RoundHalfUp)
	if err != nil {
		return accrual{}, fmt.Errorf("bond accrual: %w", err)
	}
	nominal := Nominal().Amount()
	if isValue {
		return accrual{value: q, interest: q.Sub(nominal), period: n}, nil
	}
	return accrual{value: nominal.Add(q), interest: q, period: n}, nil
}
