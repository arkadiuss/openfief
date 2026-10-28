package ledger

import (
	"errors"
	"fmt"
)

// ErrInvalidISIN reports text that is not a well-formed ISIN with a correct
// check digit.
var ErrInvalidISIN = errors.New("invalid ISIN")

// ErrInvalidInstrument reports an instrument without an ISIN or a venue
// symbol.
var ErrInvalidInstrument = errors.New("invalid instrument")

const (
	isinLen         = 12
	isinCountryLen  = 2
	luhnBase        = 10
	luhnDoubleLimit = 9
	letterBase      = 10 // ISIN letters count as A=10 ... Z=35
)

// ISIN is an International Securities Identification Number (ISO 6166): a
// two-letter country code, nine alphanumeric characters and a check digit.
// The zero value is no ISIN.
type ISIN struct {
	code string
}

// ParseISIN parses an upper-case ISIN and verifies its check digit.
func ParseISIN(s string) (ISIN, error) {
	if len(s) != isinLen {
		return ISIN{}, fmt.Errorf("%w: %q: want %d characters", ErrInvalidISIN, s, isinLen)
	}
	digits, err := isinDigits(s)
	if err != nil {
		return ISIN{}, err
	}
	if !luhnValid(digits) {
		return ISIN{}, fmt.Errorf("%w: %q: wrong check digit", ErrInvalidISIN, s)
	}
	return ISIN{code: s}, nil
}

// isinDigits returns the digits the ISIN check digit is computed over: each
// letter becomes its two-digit value (A=10 ... Z=35). Letters may appear only
// before the check digit and digits only after the country code.
func isinDigits(s string) ([]int, error) {
	var digits []int
	for i, r := range s {
		switch {
		case r >= 'A' && r <= 'Z' && i < isinLen-1:
			v := int(r-'A') + letterBase
			digits = append(digits, v/luhnBase, v%luhnBase)
		case r >= '0' && r <= '9' && i >= isinCountryLen:
			digits = append(digits, int(r-'0'))
		default:
			return nil, fmt.Errorf("%w: %q: bad character at %d", ErrInvalidISIN, s, i)
		}
	}
	return digits, nil
}

// luhnValid reports whether digits, ending with their check digit, pass the
// Luhn (mod 10) check.
func luhnValid(digits []int) bool {
	sum := 0
	for i := range digits {
		d := digits[len(digits)-1-i]
		if i%2 == 1 {
			d *= 2
			if d > luhnDoubleLimit {
				d -= luhnDoubleLimit
			}
		}
		sum += d
	}
	return sum%luhnBase == 0
}

// String returns the ISIN's 12 characters.
func (i ISIN) String() string { return i.code }

// IsZero reports whether i is the zero value, no ISIN.
func (i ISIN) IsZero() bool { return i.code == "" }

// Equal reports whether i and o are the same ISIN.
func (i ISIN) Equal(o ISIN) bool { return i.code == o.code }

// Instrument is something that can be held, identified by its ISIN plus the
// symbol it has on its trading venue. A bond series is an instrument.
// Instruments are comparable and can be map keys.
type Instrument struct {
	ISIN   ISIN
	Symbol string
}

// Validate reports whether in has an ISIN and a venue symbol.
func (in Instrument) Validate() error {
	if in.ISIN.IsZero() || in.Symbol == "" {
		return fmt.Errorf("%w: %q: needs an ISIN and a venue symbol", ErrInvalidInstrument, in.String())
	}
	return nil
}

// String returns "ISIN/symbol".
func (in Instrument) String() string { return in.ISIN.String() + "/" + in.Symbol }
