package valuation

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/arkadiuss/openfief/internal/jurisdiction/pl"
	"github.com/arkadiuss/openfief/internal/kernel"
)

// ErrUnknownIndexKind reports an IndexKind outside the declared constants,
// or text that names none of them.
var ErrUnknownIndexKind = errors.New("unknown index kind")

// IndexKind says which published series an index reading belongs to.
type IndexKind int

const (
	// CPI is consumer prices year on year from GUS, keyed by the month it
	// measures. It sets the rates of inflation-indexed bond types.
	CPI IndexKind = iota + 1
	// ReferenceRate is the NBP reference rate, keyed by the month whose
	// interest periods it sets: the rate in force on the tenth business day
	// before that month's first day (pl.Indices), which pl.ReferenceRateDay
	// names. The treasury-bonds connector picks that rate when it records
	// the reading.
	ReferenceRate
)

// String returns the kind's name: cpi or reference_rate.
func (k IndexKind) String() string {
	switch k {
	case CPI:
		return "cpi"
	case ReferenceRate:
		return "reference_rate"
	}
	return "IndexKind(" + strconv.Itoa(int(k)) + ")"
}

// ParseIndexKind parses an index kind name as String writes it.
func ParseIndexKind(s string) (IndexKind, error) {
	for _, k := range []IndexKind{CPI, ReferenceRate} {
		if k.String() == s {
			return k, nil
		}
	}
	return 0, fmt.Errorf("%w: %q", ErrUnknownIndexKind, s)
}

// IndexReading is one published reading of a period index (glossary: Period
// index) for one month, as a yearly fraction: CPI of 104.9 is 0.049. There is
// at most one reading per kind and month.
type IndexReading struct {
	Kind   IndexKind
	Year   int
	Month  time.Month
	Value  kernel.Decimal
	Source string
}

// LotIndices returns the period index readings a lot of series s bought on
// purchased needs to be valued on on: one for each interest period after the
// first that has started by on, picked from readings. An inflation-indexed
// period takes the CPI of pl.CPIMonth of its start; a reference-rate period
// takes the ReferenceRate reading of the month it starts in. Fixed-rate
// types need none. A reading that is not among readings is left out, so
// pl.ValueLot then reports it as a *pl.MissingIndexError.
func LotIndices(s pl.Series, purchased, on kernel.Date, readings []IndexReading) (pl.Indices, error) {
	st, err := s.Type.Structure()
	if err != nil {
		return nil, fmt.Errorf("indices of %s: %w", s, err)
	}
	idx := pl.Indices{}
	if st.Indexation == pl.Fixed {
		return idx, nil
	}
	for n := 2; n <= st.TenorMonths/st.PeriodMonths; n++ {
		start, err := pl.PeriodStart(s.Type, purchased, n)
		if err != nil {
			return nil, fmt.Errorf("indices of %s: %w", s, err)
		}
		if start.After(on) {
			break
		}
		kind, year, month := ReferenceRate, start.Year(), start.Month()
		if st.Indexation == pl.Inflation {
			kind = CPI
			year, month = pl.CPIMonth(start)
		}
		if r, ok := findReading(readings, kind, year, month); ok {
			idx[n] = r.Value
		}
	}
	return idx, nil
}

// findReading returns the reading of kind for year and month in readings.
func findReading(readings []IndexReading, kind IndexKind, year int, month time.Month) (IndexReading, bool) {
	for _, r := range readings {
		if r.Kind == kind && r.Year == year && r.Month == month {
			return r, true
		}
	}
	return IndexReading{}, false
}
