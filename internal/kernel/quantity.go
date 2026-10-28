package kernel

// Quantity is an exact number of units of an instrument, such as shares or
// bonds. It may be fractional. Its zero value is 0.
type Quantity struct {
	d Decimal
}

// NewQuantity returns d units.
func NewQuantity(d Decimal) Quantity { return Quantity{d: d} }

// ParseQuantity parses a decimal number of units, such as "12.5".
func ParseQuantity(s string) (Quantity, error) {
	d, err := ParseDecimal(s)
	if err != nil {
		return Quantity{}, err
	}
	return Quantity{d: d}, nil
}

// Decimal returns q as a unitless number, for example to multiply a price.
func (q Quantity) Decimal() Decimal { return q.d }

// Add returns q + o.
func (q Quantity) Add(o Quantity) Quantity { return Quantity{d: q.d.Add(o.d)} }

// Sub returns q - o.
func (q Quantity) Sub(o Quantity) Quantity { return Quantity{d: q.d.Sub(o.d)} }

// Neg returns -q.
func (q Quantity) Neg() Quantity { return Quantity{d: q.d.Neg()} }

// Cmp returns -1, 0 or +1 as q is less than, equal to or greater than o.
func (q Quantity) Cmp(o Quantity) int { return q.d.Cmp(o.d) }

// Equal reports whether q and o are numerically equal.
func (q Quantity) Equal(o Quantity) bool { return q.d.Equal(o.d) }

// Sign returns -1, 0 or +1 as q is negative, zero or positive.
func (q Quantity) Sign() int { return q.d.Sign() }

// IsZero reports whether q is zero.
func (q Quantity) IsZero() bool { return q.d.IsZero() }

// String formats q like Decimal.String.
func (q Quantity) String() string { return q.d.String() }
