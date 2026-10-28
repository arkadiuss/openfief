package demo

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/arkadiuss/openfief/internal/ledger"
)

// ErrNotSynthetic reports an ISIN body that does not carry the synthetic
// marker, or for which no check digit makes a valid ISIN.
var ErrNotSynthetic = errors.New("not a synthetic demo ISIN")

// SyntheticMarker follows the country code of every demo ISIN. Real ISINs
// have a national number there, so this pattern marks the demo's as fake at
// a glance, like the PLTEST and XSFAKE ISINs of the tests.
const SyntheticMarker = "TESTD"

const (
	isinLen        = 12
	isinCountryLen = 2
	checkDigits    = 10
)

// SyntheticISIN returns the ISIN made of country, SyntheticMarker and four
// digits, with the check digit that makes it valid (ledger.ParseISIN).
func SyntheticISIN(country string, number int) (ledger.ISIN, error) {
	const maxNumber = 9999
	if number < 0 || number > maxNumber || len(country) != isinCountryLen {
		return ledger.ISIN{}, fmt.Errorf("%w: %s %d", ErrNotSynthetic, country, number)
	}
	body := fmt.Sprintf("%s%s%04d", country, SyntheticMarker, number)
	for d := range checkDigits {
		if isin, err := ledger.ParseISIN(body + strconv.Itoa(d)); err == nil {
			return isin, nil
		}
	}
	return ledger.ISIN{}, fmt.Errorf("%w: %s", ErrNotSynthetic, body)
}

// IsSynthetic reports whether isin is a demo ISIN: SyntheticMarker right
// after the country code.
func IsSynthetic(isin ledger.ISIN) bool {
	s := isin.String()
	return len(s) == isinLen && strings.HasPrefix(s[isinCountryLen:], SyntheticMarker)
}
