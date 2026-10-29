package pl

import (
	"errors"
	"fmt"

	"github.com/arkadiuss/openfief/internal/kernel"
)

// ErrInvalidQuantity reports a bond quantity that is not a positive whole number.
var ErrInvalidQuantity = errors.New("bond quantity must be a positive whole number")

// ErrInvalidPrice reports a purchase price per bond that is not a PLN amount
// above zero and at most the nominal.
var ErrInvalidPrice = errors.New("bond purchase price must be a PLN amount above zero and at most the nominal")

// grosze is the number of decimal places of a PLN amount.
const grosze = 2

// Purchase is how the bonds of a lot were bought: on which date and at what
// price per bond. A cash purchase pays the nominal, 100 PLN. A purchase by
// exchange (zamiana) of maturing bonds pays the series' exchange price,
// usually 99.90 PLN ("Cena zamiany jednej obligacji wynosi 99,90 zł", issue
// letters of ROR, DOR, TOS, COI and EDO, ust. 10; OTS exchanges at 100.00 PLN,
// OTS letter ust. 8), on the day the old bonds mature.
type Purchase struct {
	Date  kernel.Date
	Price kernel.Money
}

// LotValue is the statutory value of a lot of one bond series on a date,
// the parts its three values (D-77) are computed from with BondValues.
// Every amount is in PLN and covers the whole lot.
type LotValue struct {
	// Gross is the gross accrued value: nominal plus interest accrued and not
	// yet paid out (for coupon types, only the current period's interest).
	Gross kernel.Money
	// Interest is the accrued interest inside Gross.
	Interest kernel.Money
	// Discount is what the lot was bought below its nominal, by exchange;
	// it is taxed with the interest.
	Discount kernel.Money
	// Fee is the early-redemption fee redeeming the lot on the date would
	// charge, after its cap: zero at maturity and inside IKE and IKZE.
	Fee kernel.Money
	// Matured reports a date on or after maturity; the value is then the
	// redemption value at maturity.
	Matured bool
}

// ValueLot returns the statutory value on date on of quantity bonds of
// t.Series bought as p and held in wrapper: the gross accrued value, its
// interest, the purchase discount and the early-redemption fee redeeming it
// on that date would charge. idx supplies the index readings of every
// period up to the one the date falls in.
//
// The value of one bond follows the issue letters' early-redemption formula
// (Załącznik nr 3 of the EDO1036 letter, list emisyjny nr 99/2026, linked from
// https://www.obligacjeskarbowe.pl/oferta-obligacji/obligacje-10-letnie-edo/):
//
//	WP_k = N × (1 + r_1) × … × (1 + r_k-1) × (1 + r_k × a_k / ACT) − b
//
// "zaokrąglona do dwóch miejsc po przecinku", where a_k is the actual days
// "od pierwszego dnia [...] okresu odsetkowego, z włączeniem tego dnia, do
// dnia »d« z wyłączeniem dnia »d«" and ACT the actual days of the period,
// first day included and last excluded. The product stays exact and the
// value is rounded once per bond, half up: the letters do not name the
// direction, and half up matches every day of the published tables. Coupon types count only the
// current period: O = N × r × a / (D × F) with F = 12 for ROR and DOR
// (Załącznik nr 2 of their letters) and N × r × a / ACT for COI. OTS earns
// its fixed interest only at maturity. Lot amounts are the per-bond amounts
// times the quantity.
//
// The published interest tables (https://www.obligacjeskarbowe.pl/tabela-odsetkowa/)
// match this for every day: EDO1133 is worth 115.49 PLN on 2026-03-01, where
// rounding the base after each year would give 115.50. They also do so for
// TOS, whose letters word the base as rounded to grosze each year (TOS1029
// letter, Załącznik nr 2: "N_k-1 [...] zaokrąglona do dwóch miejsc po
// przecinku"): TOS0128 is worth 112.31 PLN on 2027-01-04 in the table and
// here, 112.30 with a rounded base. ValueLot follows the tables.
//
// The early-redemption fee b is charged per bond and capped at the bond's
// accrued interest ("dla WP_k < 100 WP_k = 100"), except for coupon types
// after their first interest period, where the full fee is charged ("w
// pierwszym okresie odsetkowym [...] nie wyższą niż", then "począwszy od
// drugiego okresu odsetkowego [...] pomniejszana o kwotę"; ROR, DOR and COI
// letters). It is not charged inside IKE or IKZE, assuming a withdrawal
// ("potrącenia nie dokonuje się wobec posiadaczy obligacji dokonujących
// wypłaty lub wypłaty transferowej z IKE lub IKZE"); it is charged in a
// taxable account and in OKI, which the letters don't name. The value on a date d is
// what an early redemption with interest counted to d pays; the issuer's
// worked examples take d as the fifth business day after the redemption
// order.
func ValueLot(t *Terms, idx Indices, quantity kernel.Quantity, p Purchase, wrapper TaxWrapper, on kernel.Date) (LotValue, error) {
	if err := t.validate(); err != nil {
		return LotValue{}, err
	}
	if err := validateLot(quantity, p); err != nil {
		return LotValue{}, err
	}
	maturity, err := Maturity(t.Series, p.Date)
	if err != nil {
		return LotValue{}, err
	}
	if on.Before(p.Date) {
		return LotValue{}, fmt.Errorf("%w: %s bought on %s, valued on %s", ErrBeforePurchase, t.Series, p.Date, on)
	}
	matured := !on.Before(maturity)
	if matured {
		on = maturity
	}
	acc, err := accrue(t, idx, p.Date, on, matured)
	if err != nil {
		return LotValue{}, err
	}
	charged, err := chargesEarlyFee(wrapper)
	if err != nil {
		return LotValue{}, err
	}
	var fee kernel.Decimal
	if !matured && charged {
		fee = cappedFee(t, acc)
	}
	q, c := quantity.Decimal(), pln()
	return LotValue{
		Gross:    kernel.NewMoney(acc.value.Mul(q), c),
		Interest: kernel.NewMoney(acc.interest.Mul(q), c),
		Discount: kernel.NewMoney(Nominal().Amount().Sub(p.Price.Amount()).Mul(q), c),
		Fee:      kernel.NewMoney(fee.Mul(q), c),
		Matured:  matured,
	}, nil
}

// chargesEarlyFee reports whether redeeming a bond early in wrapper pays
// the early-redemption fee: in a taxable account and in OKI, not on a
// withdrawal from IKE or IKZE.
func chargesEarlyFee(wrapper TaxWrapper) (bool, error) {
	switch wrapper {
	case Taxable, OKI:
		return true, nil
	case IKE, IKZE:
		return false, nil
	}
	return false, fmt.Errorf("%w: %d", ErrUnknownTaxWrapper, wrapper)
}

// validateLot checks that quantity is a positive whole number of bonds and
// that the purchase price per bond is a PLN amount in (0, nominal].
func validateLot(quantity kernel.Quantity, p Purchase) error {
	whole, err := quantity.Decimal().Round(0, kernel.RoundDown)
	if err != nil || quantity.Sign() <= 0 || !whole.Equal(quantity.Decimal()) {
		return fmt.Errorf("%w: %s", ErrInvalidQuantity, quantity)
	}
	if !p.Price.Currency().Equal(pln()) || p.Price.Sign() <= 0 || p.Price.Amount().Cmp(Nominal().Amount()) > 0 {
		return fmt.Errorf("%w: %s", ErrInvalidPrice, p.Price)
	}
	return nil
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
// must be on or before maturity; matured reports that on is the maturity.
// Every quantity stays exact until the single division that rounds the result
// to grosze.
func accrue(t *Terms, idx Indices, purchased, on kernel.Date, matured bool) (accrual, error) {
	st, err := t.Series.Type.Structure()
	if err != nil {
		return accrual{}, err
	}
	if st.Payout == AtMaturity && !matured {
		return accrual{value: Nominal().Amount(), interest: kernel.Decimal{}, period: 1}, nil
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
