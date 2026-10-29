package ingest

import (
	"errors"
	"fmt"

	"github.com/arkadiuss/openfief/internal/kernel"
)

// ErrInvalidIndexReading reports an index reading without a known index, a
// date or a source, or a CPI reading not dated on the first of its month.
var ErrInvalidIndexReading = errors.New("invalid index reading")

// Index is a published index that sets treasury bond interest after the first
// interest period.
type Index int

// The indices. The zero Index is not an index.
const (
	// IndexCPI is CPI year on year from GUS, published for one month.
	IndexCPI Index = iota + 1
	// IndexReferenceRate is the NBP reference rate, for the month whose
	// ROR and DOR interest periods it sets.
	IndexReferenceRate
)

// String returns "cpi" or "reference-rate", or a placeholder for an invalid
// index.
func (i Index) String() string {
	switch i {
	case IndexCPI:
		return "cpi"
	case IndexReferenceRate:
		return "reference-rate"
	}
	return fmt.Sprintf("Index(%d)", int(i))
}

// IndexReading is one published value of an index, as a connector submits
// it: a yearly fraction, so CPI of 104.9 (previous year = 100) is 0.049 and a
// 5.75% reference rate is 0.0575. Valuation picks the reading each interest
// period needs from these (pl.Indices).
type IndexReading struct {
	Index Index
	// Date is, for CPI, the first day of the month the reading measures and,
	// for the reference rate, the first day of the month whose interest
	// periods it sets: the rate in force on that month's reference-rate day
	// (pl.ReferenceRateDay). The core records each reading for the month of
	// its date.
	Date  kernel.Date
	Value kernel.Decimal
	// Source names who submitted the reading, such as a connector.
	Source string
}

// Validate reports whether r has a known index, a date and a source, and,
// for CPI, is dated on the first of its month.
func (r *IndexReading) Validate() error {
	switch {
	case r.Index != IndexCPI && r.Index != IndexReferenceRate:
		return fmt.Errorf("%w: %s", ErrInvalidIndexReading, r.Index)
	case r.Date.IsZero():
		return fmt.Errorf("%w: %s has no date", ErrInvalidIndexReading, r.Index)
	case r.Index == IndexCPI && r.Date.Day() != 1:
		return fmt.Errorf("%w: CPI dated %s, not the first of its month", ErrInvalidIndexReading, r.Date)
	case r.Source == "":
		return fmt.Errorf("%w: %s on %s has no source", ErrInvalidIndexReading, r.Index, r.Date)
	}
	return nil
}
