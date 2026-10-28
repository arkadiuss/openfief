package kernel

import (
	"errors"
	"fmt"

	"github.com/shopspring/decimal"
)

// ErrInvalidDecimal reports text that is not a decimal number.
var ErrInvalidDecimal = errors.New("invalid decimal")

// ErrUnknownRoundingMode reports a RoundingMode outside the declared constants.
var ErrUnknownRoundingMode = errors.New("unknown rounding mode")

// RoundingMode says how Round treats the digits it drops.
type RoundingMode int

const (
	// RoundHalfEven rounds to the nearest value; a tie goes to the even digit.
	RoundHalfEven RoundingMode = iota + 1
	// RoundHalfUp rounds to the nearest value; a tie goes away from zero.
	RoundHalfUp
	// RoundDown drops the extra digits, rounding toward zero.
	RoundDown
)

// Decimal is an exact decimal number with no unit, such as an FX rate or a
// factor. Its zero value is 0. Add, Sub, Neg and Mul are exact.
type Decimal struct {
	v decimal.Decimal
}

// ParseDecimal parses a decimal number such as "-12.345".
func ParseDecimal(s string) (Decimal, error) {
	v, err := decimal.NewFromString(s)
	if err != nil {
		return Decimal{}, fmt.Errorf("%w: %q: %w", ErrInvalidDecimal, s, err)
	}
	return Decimal{v: v}, nil
}

// NewDecimal returns coefficient × 10^exponent, such as NewDecimal(1234, -2) for 12.34.
func NewDecimal(coefficient int64, exponent int32) Decimal {
	return Decimal{v: decimal.New(coefficient, exponent)}
}

// Add returns d + e.
func (d Decimal) Add(e Decimal) Decimal { return Decimal{v: d.v.Add(e.v)} }

// Sub returns d - e.
func (d Decimal) Sub(e Decimal) Decimal { return Decimal{v: d.v.Sub(e.v)} }

// Neg returns -d.
func (d Decimal) Neg() Decimal { return Decimal{v: d.v.Neg()} }

// Mul returns d × e, exactly.
func (d Decimal) Mul(e Decimal) Decimal { return Decimal{v: d.v.Mul(e.v)} }

// Round returns d rounded to places digits after the decimal point.
func (d Decimal) Round(places int32, mode RoundingMode) (Decimal, error) {
	switch mode {
	case RoundHalfEven:
		return Decimal{v: d.v.RoundBank(places)}, nil
	case RoundHalfUp:
		return Decimal{v: d.v.Round(places)}, nil
	case RoundDown:
		return Decimal{v: d.v.Truncate(places)}, nil
	}
	return Decimal{}, fmt.Errorf("%w: %d", ErrUnknownRoundingMode, mode)
}

// Cmp returns -1, 0 or +1 as d is less than, equal to or greater than e.
func (d Decimal) Cmp(e Decimal) int { return d.v.Cmp(e.v) }

// Equal reports whether d and e are numerically equal: 1.50 equals 1.5.
func (d Decimal) Equal(e Decimal) bool { return d.v.Equal(e.v) }

// Sign returns -1, 0 or +1 as d is negative, zero or positive.
func (d Decimal) Sign() int { return d.v.Sign() }

// IsZero reports whether d is zero.
func (d Decimal) IsZero() bool { return d.v.IsZero() }

// String formats d without an exponent and without trailing zeros.
func (d Decimal) String() string { return d.v.String() }
