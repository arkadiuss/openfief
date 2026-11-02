package pl

import (
	"fmt"

	"github.com/arkadiuss/openfief/internal/kernel"
)

// ratio is an exact quotient num/den with den above zero. OKI's averages
// divide by the days an account was held, which a decimal can't hold
// exactly, so they are kept as ratios and rounded only for the base, the
// tax and what is shown (D-144). The zero value is not valid: zeroRatio and
// ratioOf make one.
type ratio struct {
	num, den kernel.Decimal
}

// zeroRatio returns zero as a ratio.
func zeroRatio() ratio { return ratio{den: kernel.NewDecimal(1, 0)} }

// ratioOf returns d as a ratio.
func ratioOf(d kernel.Decimal) ratio { return ratio{num: d, den: kernel.NewDecimal(1, 0)} }

// normal returns r with a positive denominator; a zero denominator, which
// no caller makes, becomes zero.
func (r ratio) normal() ratio {
	switch r.den.Sign() {
	case 0:
		return zeroRatio()
	case -1:
		return ratio{num: r.num.Neg(), den: r.den.Neg()}
	}
	return r
}

// add returns r + o.
func (r ratio) add(o ratio) ratio {
	return ratio{num: r.num.Mul(o.den).Add(o.num.Mul(r.den)), den: r.den.Mul(o.den)}
}

// sub returns r - o.
func (r ratio) sub(o ratio) ratio {
	return r.add(ratio{num: o.num.Neg(), den: o.den})
}

// cmp compares r with o: -1, 0 or 1.
func (r ratio) cmp(o ratio) int {
	return r.num.Mul(o.den).Cmp(o.num.Mul(r.den))
}

// sign returns r's sign: -1, 0 or 1.
func (r ratio) sign() int { return r.num.Sign() }

// min returns the smaller of r and o.
func (r ratio) min(o ratio) ratio {
	if o.cmp(r) < 0 {
		return o
	}
	return r
}

// clamp returns r limited to 0..limit, zero when either is not positive.
func (r ratio) clamp(limit ratio) ratio {
	if r.sign() <= 0 || limit.sign() <= 0 {
		return zeroRatio()
	}
	return r.min(limit)
}

// round returns r rounded to places with mode.
func (r ratio) round(places int32, mode kernel.RoundingMode) (kernel.Decimal, error) {
	d, err := r.num.Quo(r.den, places, mode)
	if err != nil {
		return kernel.Decimal{}, fmt.Errorf("rounding %s/%s: %w", r.num, r.den, err)
	}
	return d, nil
}
