package kernel

import (
	"errors"
	"fmt"
)

// ErrInvalidCurrency reports text that is not an ISO 4217 alphabetic code.
var ErrInvalidCurrency = errors.New("invalid currency code")

const currencyCodeLen = 3

// Currency is an ISO 4217 alphabetic currency code, such as PLN or USD.
// Its zero value is no currency; ParseCurrency is the only way to make one.
type Currency struct {
	code string
}

// ParseCurrency accepts three upper-case ASCII letters. It checks the
// ISO 4217 format, not membership in the current list of codes.
func ParseCurrency(code string) (Currency, error) {
	if len(code) != currencyCodeLen {
		return Currency{}, fmt.Errorf("%w: %q", ErrInvalidCurrency, code)
	}
	for _, r := range code {
		if r < 'A' || r > 'Z' {
			return Currency{}, fmt.Errorf("%w: %q", ErrInvalidCurrency, code)
		}
	}
	return Currency{code: code}, nil
}

// Equal reports whether c and o are the same currency.
func (c Currency) Equal(o Currency) bool { return c.code == o.code }

// IsZero reports whether c is the zero value, no currency.
func (c Currency) IsZero() bool { return c.code == "" }

// String returns the ISO 4217 code.
func (c Currency) String() string { return c.code }
