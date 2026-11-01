package pl

import (
	"fmt"

	"github.com/arkadiuss/openfief/internal/kernel"
)

// CouponPayment is the interest one bond of a coupon series (Payout Coupon: ROR,
// DOR, COI) pays at the end of an interest period: on Date, the first day
// of the next period.
type CouponPayment struct {
	// Period is the interest period it pays, 1-based.
	Period int
	// Date is the day it falls due.
	Date kernel.Date
	// PerBond is what one bond is paid, in PLN, before tax.
	PerBond kernel.Money
}

// CouponDates returns the day each coupon of a bond of s bought on
// purchased falls due, in order, period n's at index n-1: the first day of
// period n+1. The last period's interest is paid with the redemption at
// maturity, not as a coupon, and a series that pays no coupons (Capitalized
// or AtMaturity) has none. A purchase date that doesn't match the series
// fails as Maturity does.
func CouponDates(s Series, purchased kernel.Date) ([]kernel.Date, error) {
	if _, err := Maturity(s, purchased); err != nil {
		return nil, err
	}
	st, err := s.Type.Structure()
	if err != nil {
		return nil, err
	}
	if st.Payout != Coupon {
		return nil, nil
	}
	periods := st.TenorMonths / st.PeriodMonths
	out := make([]kernel.Date, 0, periods-1)
	for n := 1; n < periods; n++ {
		due, err := PeriodStart(s.Type, purchased, n+1)
		if err != nil {
			return nil, err
		}
		out = append(out, due)
	}
	return out, nil
}

// Coupon returns the coupon of interest period n of one bond of t.Series
// bought on purchased: the nominal times the period's rate times its share
// of a year, months over 12, rounded half up to grosze once per bond, as
// ValueLot counts a whole period (O = N × r / 12 for ROR and DOR, N × r for
// COI, Załącznik nr 2 of their issue letters). idx supplies the period's
// index reading when n is after the first; without it Coupon fails with a
// *MissingIndexError. A period after the last coupon (CouponDates), or a
// series that pays no coupons, fails with ErrInvalidTerms.
func (t *Terms) Coupon(idx Indices, purchased kernel.Date, n int) (CouponPayment, error) {
	if err := t.validate(); err != nil {
		return CouponPayment{}, err
	}
	dates, err := CouponDates(t.Series, purchased)
	if err != nil {
		return CouponPayment{}, err
	}
	if n < 1 || n > len(dates) {
		return CouponPayment{}, fmt.Errorf("%w: %s pays no coupon for period %d", ErrInvalidTerms, t.Series, n)
	}
	p, err := t.period(idx, purchased, n)
	if err != nil {
		return CouponPayment{}, err
	}
	st, err := t.Series.Type.Structure()
	if err != nil {
		return CouponPayment{}, err
	}
	twelve := kernel.NewDecimal(monthsPerYear, 0)
	num := Nominal().Amount().Mul(p.rate).Mul(kernel.NewDecimal(int64(st.PeriodMonths), 0))
	perBond, err := num.Quo(twelve, grosze, kernel.RoundHalfUp)
	if err != nil {
		return CouponPayment{}, fmt.Errorf("coupon of %s: %w", t.Series, err)
	}
	return CouponPayment{Period: n, Date: dates[n-1], PerBond: kernel.NewMoney(perBond, pln())}, nil
}
