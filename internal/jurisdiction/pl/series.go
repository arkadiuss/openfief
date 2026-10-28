package pl

import (
	"errors"
	"fmt"
	"strconv"
	"time"
)

// ErrInvalidSeries reports text that is not a retail treasury bond series code.
var ErrInvalidSeries = errors.New("invalid bond series")

// ErrUnknownBondType reports a BondType outside the declared constants.
var ErrUnknownBondType = errors.New("unknown bond type")

const (
	seriesCodeLen   = 7    // three letters, then MMYY
	typeCodeLen     = 3    // OTS, ROR, ...
	centuryBase     = 2000 // series codes carry two-digit years of this century
	monthsPerYear   = 12
	yearsPerCentury = 100
	decimalBase     = 10
)

// Terms in months of each bond type.
const (
	otsTenor = 3
	rorTenor = 12
	dorTenor = 24
	tosTenor = 36
	coiTenor = 48
	rosTenor = 72
	edoTenor = 120
	rodTenor = 144
)

// BondType is the product a series belongs to, named by the first three
// letters of its code.
type BondType int

const (
	// OTS is a 3-month fixed-rate bond; interest is paid at maturity.
	OTS BondType = iota + 1
	// ROR is a 1-year bond with monthly coupons at the NBP reference rate plus a margin.
	ROR
	// DOR is a 2-year bond with monthly coupons at the NBP reference rate plus a margin.
	DOR
	// TOS is a 3-year fixed-rate bond; interest is capitalized yearly.
	TOS
	// COI is a 4-year inflation-indexed bond with yearly coupons.
	COI
	// EDO is a 10-year inflation-indexed bond; interest is capitalized yearly.
	EDO
	// ROS is a 6-year inflation-indexed family bond; interest is capitalized yearly.
	ROS
	// ROD is a 12-year inflation-indexed family bond; interest is capitalized yearly.
	ROD
)

// Indexation says what sets the rate of interest periods after the first.
type Indexation int

const (
	// Fixed keeps the first-period rate for the whole term (OTS, TOS).
	Fixed Indexation = iota + 1
	// Inflation uses CPI year on year from GUS plus the series margin (COI, EDO, ROS, ROD).
	Inflation
	// ReferenceRate uses the NBP reference rate plus the series margin (ROR, DOR).
	ReferenceRate
)

// Payout says what happens to interest at the end of an interest period.
type Payout int

const (
	// Capitalized adds each period's interest to the base of the next; all of
	// it is paid at redemption (OTS, TOS, EDO, ROS, ROD).
	Capitalized Payout = iota + 1
	// Coupon pays each period's interest out at the period's end (ROR, DOR, COI).
	Coupon
)

// Structure is the shape every series of one bond type shares.
type Structure struct {
	TenorMonths  int
	PeriodMonths int
	Indexation   Indexation
	Payout       Payout
}

// Structure returns the term, interest period, indexation and payout of b,
// as described on each product's offer page at obligacjeskarbowe.pl.
func (b BondType) Structure() (Structure, error) {
	const year = monthsPerYear
	switch b {
	case OTS:
		return Structure{TenorMonths: otsTenor, PeriodMonths: otsTenor, Indexation: Fixed, Payout: Capitalized}, nil
	case ROR:
		return Structure{TenorMonths: rorTenor, PeriodMonths: 1, Indexation: ReferenceRate, Payout: Coupon}, nil
	case DOR:
		return Structure{TenorMonths: dorTenor, PeriodMonths: 1, Indexation: ReferenceRate, Payout: Coupon}, nil
	case TOS:
		return Structure{TenorMonths: tosTenor, PeriodMonths: year, Indexation: Fixed, Payout: Capitalized}, nil
	case COI:
		return Structure{TenorMonths: coiTenor, PeriodMonths: year, Indexation: Inflation, Payout: Coupon}, nil
	case EDO:
		return Structure{TenorMonths: edoTenor, PeriodMonths: year, Indexation: Inflation, Payout: Capitalized}, nil
	case ROS:
		return Structure{TenorMonths: rosTenor, PeriodMonths: year, Indexation: Inflation, Payout: Capitalized}, nil
	case ROD:
		return Structure{TenorMonths: rodTenor, PeriodMonths: year, Indexation: Inflation, Payout: Capitalized}, nil
	}
	return Structure{}, fmt.Errorf("%w: %d", ErrUnknownBondType, b)
}

// String returns the three-letter code, such as "EDO".
func (b BondType) String() string {
	for _, t := range bondTypes() {
		if t.typ == b {
			return t.code
		}
	}
	return "BondType(" + strconv.Itoa(int(b)) + ")"
}

type bondTypeCode struct {
	code string
	typ  BondType
}

func bondTypes() []bondTypeCode {
	return []bondTypeCode{
		{"OTS", OTS},
		{"ROR", ROR},
		{"DOR", DOR},
		{"TOS", TOS},
		{"COI", COI},
		{"EDO", EDO},
		{"ROS", ROS},
		{"ROD", ROD},
	}
}

// Series is a retail treasury bond series, such as EDO0536: the bond type
// and the month the series matures. Every bond of a series is sold in the
// same month, and each bond's term runs from its own purchase date.
type Series struct {
	Type          BondType
	MaturityYear  int
	MaturityMonth time.Month
}

// ParseSeries parses a series code: the bond type's three letters, then the
// maturity month and two-digit year, so EDO0536 is EDO maturing in May 2036.
func ParseSeries(code string) (Series, error) {
	if len(code) != seriesCodeLen {
		return Series{}, fmt.Errorf("%w: %q: want 3 letters and MMYY", ErrInvalidSeries, code)
	}
	var typ BondType
	for _, t := range bondTypes() {
		if t.code == code[:typeCodeLen] {
			typ = t.typ
		}
	}
	if typ == 0 {
		return Series{}, fmt.Errorf("%w: %q: unknown bond type %q", ErrInvalidSeries, code, code[:typeCodeLen])
	}
	for _, r := range code[typeCodeLen:] {
		if r < '0' || r > '9' {
			return Series{}, fmt.Errorf("%w: %q: bad maturity MMYY", ErrInvalidSeries, code)
		}
	}
	month, year := twoDigits(code[typeCodeLen:typeCodeLen+2]), twoDigits(code[typeCodeLen+2:])
	if month < 1 || month > monthsPerYear {
		return Series{}, fmt.Errorf("%w: %q: bad maturity MMYY", ErrInvalidSeries, code)
	}
	return Series{Type: typ, MaturityYear: centuryBase + year, MaturityMonth: time.Month(month)}, nil
}

// twoDigits reads two ASCII digits as a number.
func twoDigits(s string) int { return int(s[0]-'0')*decimalBase + int(s[1]-'0') }

// String returns the series code, such as "EDO0536".
func (s Series) String() string {
	return fmt.Sprintf("%s%02d%02d", s.Type, int(s.MaturityMonth), s.MaturityYear%yearsPerCentury)
}
