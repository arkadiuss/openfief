package kernel

import "fmt"

// CurrencyMismatchError reports an operation on money in two different currencies.
type CurrencyMismatchError struct {
	Left, Right Currency
}

func (e *CurrencyMismatchError) Error() string {
	return fmt.Sprintf("currency mismatch: %s and %s", e.Left, e.Right)
}

// Money is an exact amount in one currency. It keeps its original currency:
// nothing in Money converts, and arithmetic between currencies is an error.
type Money struct {
	amount   Decimal
	currency Currency
}

// NewMoney returns amount in currency c.
func NewMoney(amount Decimal, c Currency) Money {
	return Money{amount: amount, currency: c}
}

// ParseMoney parses amount as a decimal number in currency c.
func ParseMoney(amount string, c Currency) (Money, error) {
	d, err := ParseDecimal(amount)
	if err != nil {
		return Money{}, err
	}
	return NewMoney(d, c), nil
}

// Amount returns the amount without its currency.
func (m Money) Amount() Decimal { return m.amount }

// Currency returns the currency of m.
func (m Money) Currency() Currency { return m.currency }

// Add returns m + o; both must be in the same currency.
func (m Money) Add(o Money) (Money, error) {
	if err := m.sameCurrency(o); err != nil {
		return Money{}, err
	}
	return NewMoney(m.amount.Add(o.amount), m.currency), nil
}

// Sub returns m - o; both must be in the same currency.
func (m Money) Sub(o Money) (Money, error) {
	if err := m.sameCurrency(o); err != nil {
		return Money{}, err
	}
	return NewMoney(m.amount.Sub(o.amount), m.currency), nil
}

// Neg returns -m.
func (m Money) Neg() Money { return NewMoney(m.amount.Neg(), m.currency) }

// Mul returns m × factor in m's currency, exactly; round it with Round.
// Use it for a price times a quantity (factor = Quantity.Decimal()).
func (m Money) Mul(factor Decimal) Money {
	return NewMoney(m.amount.Mul(factor), m.currency)
}

// Round returns m rounded to places digits after the decimal point.
func (m Money) Round(places int32, mode RoundingMode) (Money, error) {
	d, err := m.amount.Round(places, mode)
	if err != nil {
		return Money{}, err
	}
	return NewMoney(d, m.currency), nil
}

// Cmp compares m and o, which must be in the same currency.
func (m Money) Cmp(o Money) (int, error) {
	if err := m.sameCurrency(o); err != nil {
		return 0, err
	}
	return m.amount.Cmp(o.amount), nil
}

// Equal reports whether m and o have the same currency and equal amounts.
func (m Money) Equal(o Money) bool {
	return m.currency.Equal(o.currency) && m.amount.Equal(o.amount)
}

// Sign returns -1, 0 or +1 as the amount is negative, zero or positive.
func (m Money) Sign() int { return m.amount.Sign() }

// IsZero reports whether the amount is zero.
func (m Money) IsZero() bool { return m.amount.IsZero() }

// String formats m as "<amount> <currency>", such as "12.5 PLN".
func (m Money) String() string { return m.amount.String() + " " + m.currency.String() }

func (m Money) sameCurrency(o Money) error {
	if !m.currency.Equal(o.currency) {
		return &CurrencyMismatchError{Left: m.currency, Right: o.currency}
	}
	return nil
}
