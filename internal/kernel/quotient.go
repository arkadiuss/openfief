package kernel

import (
	"errors"
	"fmt"

	"github.com/shopspring/decimal"
)

// ErrDivisionByZero reports a division by zero.
var ErrDivisionByZero = errors.New("division by zero")

// Quo returns d ÷ e rounded once to places digits after the decimal point
// with mode. A quotient is rarely exact, so unlike Add, Sub and Mul it always
// takes a rounding: callers keep the numerator exact and divide last, so the
// result is rounded exactly once.
func (d Decimal) Quo(e Decimal, places int32, mode RoundingMode) (Decimal, error) {
	if e.IsZero() {
		return Decimal{}, fmt.Errorf("%w: %s / 0", ErrDivisionByZero, d)
	}
	// q is truncated toward zero; r has d's sign and |r| < |e| × 10^-places.
	q, r := d.v.QuoRem(e.v, places)
	unit := decimal.New(1, -places)
	// Compare the dropped part with half a unit: 2|r| against |e| × 10^-places.
	half := r.Abs().Add(r.Abs()).Cmp(e.v.Abs().Mul(unit))
	ulp := unit
	if d.v.Sign() != e.v.Sign() {
		ulp = unit.Neg()
	}
	switch mode {
	case RoundDown:
		return Decimal{v: q}, nil
	case RoundHalfUp:
		if half >= 0 {
			q = q.Add(ulp)
		}
		return Decimal{v: q}, nil
	case RoundHalfEven:
		if half > 0 || (half == 0 && lastDigitOdd(q, places)) {
			q = q.Add(ulp)
		}
		return Decimal{v: q}, nil
	}
	return Decimal{}, fmt.Errorf("%w: %d", ErrUnknownRoundingMode, mode)
}

// lastDigitOdd reports whether the digit of q at 10^-places is odd.
func lastDigitOdd(q decimal.Decimal, places int32) bool {
	units := q.Shift(places).Truncate(0).BigInt()
	return units.Bit(0) == 1
}
