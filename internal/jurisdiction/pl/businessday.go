package pl

import (
	"errors"
	"fmt"
	"time"

	"github.com/arkadiuss/openfief/internal/kernel"
)

// FirstCalendarYear is the first year the holiday calendar of
// IsBusinessDay covers. The act on days off work (ustawa z dnia 18 stycznia
// 1951 r. o dniach wolnych od pracy) took its list of fixed holidays,
// 3 May and 11 November included, in amendments in force from 1989-05-28
// and 1990-04-28 (Dz.U. 1989 poz. 154, Dz.U. 1990 poz. 159 and 160); every
// day off it lists from 1991 on is modeled.
const FirstCalendarYear = 1991

// Years from which the act lists a day it did not list before.
const (
	// epiphanyFrom: 6 January (Święto Trzech Króli), from 2011-01-01
	// (ustawa z dnia 24 września 2010 r., Dz.U. 2010 nr 224 poz. 1459, art. 2).
	epiphanyFrom = 2011
	// christmasEveFrom: 24 December (Wigilia Bożego Narodzenia), from
	// 2025-02-01 (ustawa z dnia 6 grudnia 2024 r., Dz.U. 2024 poz. 1965,
	// art. 1 and 5), so first on 24 December 2025.
	christmasEveFrom = 2025
)

// Days after Easter Sunday of the movable days off.
const (
	easterMonday   = 1
	pentecost      = 49 // pierwszy dzień Zielonych Świątek, the seventh Sunday after Easter
	corpusChristi  = 60 // Boże Ciało, the Thursday 60 days after Easter
	christmasEve   = 24
	epiphany       = 6
	assumption     = 15
	independence   = 11
	constitution   = 3
	firstChristmas = 25
	boxingDay      = 26
)

// referenceRateLookback is how many business days before the first day of
// a month the NBP reference rate that sets its ROR and DOR interest periods
// is read.
const referenceRateLookback = 10

// ErrOutsideCalendar reports a date before FirstCalendarYear, which the
// holiday calendar does not cover.
var ErrOutsideCalendar = errors.New("date outside the Polish holiday calendar")

// IsBusinessDay reports whether d is a business day (dzień roboczy) in the
// sense of the retail bond issue letters: neither a Saturday ("Do dni
// roboczych [...] nie wlicza się sobót", ROR1027 letter, list emisyjny nr
// 95/2026, ust. 29) nor a statutory day off work ("dzień ustawowo wolny od
// pracy", ust. 33). The days off are Sundays and the holidays art. 1 of the
// act on days off work lists (consolidated text Dz.U. 2025 poz. 296,
// https://api.sejm.gov.pl/eli/acts/DU/2025/296/text.pdf): 1 and 6 January,
// Easter Sunday and Monday, 1 and 3 May, Pentecost Sunday, Corpus Christi,
// 15 August, 1 and 11 November, and 24 to 26 December, each from the year
// the act lists it. A day off the Prime Minister may declare during an
// epidemic (art. 1a) is not modeled: none has been declared. It fails with
// ErrOutsideCalendar before FirstCalendarYear.
func IsBusinessDay(d kernel.Date) (bool, error) {
	if d.Year() < FirstCalendarYear {
		return false, fmt.Errorf("%w: %s, before %d", ErrOutsideCalendar, d, FirstCalendarYear)
	}
	switch weekday(d) {
	case time.Saturday, time.Sunday:
		return false, nil
	case time.Monday, time.Tuesday, time.Wednesday, time.Thursday, time.Friday:
	}
	return !isHoliday(d), nil
}

// isHoliday reports whether d is one of the holidays art. 1 pkt 1 of the act
// on days off work lists in d's year.
func isHoliday(d kernel.Date) bool {
	y := d.Year()
	fixed := []struct {
		month time.Month
		day   int
		from  int
	}{
		{time.January, 1, FirstCalendarYear},
		{time.January, epiphany, epiphanyFrom},
		{time.May, 1, FirstCalendarYear},
		{time.May, constitution, FirstCalendarYear},
		{time.August, assumption, FirstCalendarYear},
		{time.November, 1, FirstCalendarYear},
		{time.November, independence, FirstCalendarYear},
		{time.December, christmasEve, christmasEveFrom},
		{time.December, firstChristmas, FirstCalendarYear},
		{time.December, boxingDay, FirstCalendarYear},
	}
	for _, h := range fixed {
		if y >= h.from && d.Month() == h.month && d.Day() == h.day {
			return true
		}
	}
	easter := easterSunday(y)
	for _, after := range []int{0, easterMonday, pentecost, corpusChristi} {
		if easter.AddDays(after).Equal(d) {
			return true
		}
	}
	return false
}

// Constants of the anonymous Gregorian Easter algorithm (Meeus/Jones/Butcher,
// as given in Jean Meeus, Astronomical Algorithms, 2nd ed., chapter 8).
const (
	meeusGolden    = 19
	meeusCentury   = 100
	meeusQuarter   = 4
	meeusLeapShift = 8
	meeusLeapCycle = 25
	meeusThird     = 3
	meeusEpactBase = 15
	meeusLunar     = 30
	meeusWeekBase  = 32
	meeusDouble    = 2
	meeusWeek      = 7
	meeusMA        = 11
	meeusML        = 22
	meeusMDiv      = 451
	meeusMonthBase = 114
	meeusMonthDiv  = 31
)

// easterSunday returns the date of Easter Sunday in the Gregorian calendar
// for year.
func easterSunday(year int) kernel.Date {
	a := year % meeusGolden
	b, c := year/meeusCentury, year%meeusCentury
	d, e := b/meeusQuarter, b%meeusQuarter
	f := (b + meeusLeapShift) / meeusLeapCycle
	g := (b - f + 1) / meeusThird
	h := (meeusGolden*a + b - d - g + meeusEpactBase) % meeusLunar
	i, k := c/meeusQuarter, c%meeusQuarter
	l := (meeusWeekBase + meeusDouble*e + meeusDouble*i - h - k) % meeusWeek
	m := (a + meeusMA*h + meeusML*l) / meeusMDiv
	n := h + l - meeusWeek*m + meeusMonthBase
	month, day := time.Month(n/meeusMonthDiv), n%meeusMonthDiv+1
	t := time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
	return mustDate(t)
}

// weekday is the day of the week of d.
func weekday(d kernel.Date) time.Weekday {
	return time.Date(d.Year(), d.Month(), d.Day(), 0, 0, 0, 0, time.UTC).Weekday()
}

// mustDate is the calendar date of t, which is always a valid date.
func mustDate(t time.Time) kernel.Date {
	d, err := kernel.NewDate(t.Year(), t.Month(), t.Day())
	if err != nil {
		return kernel.Date{}
	}
	return d
}

// ReferenceRateDay returns the day whose NBP reference rate sets the ROR
// and DOR interest periods starting in month of year: the tenth business
// day before that month's first day ("stopa referencyjna NBP obowiązująca
// w dziesiątym dniu roboczym przed pierwszym dniem miesiąca
// kalendarzowego, w którym rozpoczyna się okres odsetkowy", ROR1027 letter
// ust. 16 and Załącznik nr 1; the DOR letters say the same). The first
// business day before the month is the last business day of the month
// before. It fails with ErrOutsideCalendar when that day falls before
// FirstCalendarYear.
func ReferenceRateDay(year int, month time.Month) (kernel.Date, error) {
	first, err := kernel.NewDate(year, month, 1)
	if err != nil {
		return kernel.Date{}, fmt.Errorf("reference rate day of %d-%02d: %w", year, int(month), err)
	}
	d := first
	for counted := 0; counted < referenceRateLookback; {
		d = d.AddDays(-1)
		ok, err := IsBusinessDay(d)
		if err != nil {
			return kernel.Date{}, err
		}
		if ok {
			counted++
		}
	}
	return d, nil
}

// RateChange is one change of the NBP reference rate: the yearly rate, as a
// fraction (5.75% is 0.0575), in force from From until the next change.
type RateChange struct {
	From kernel.Date
	Rate kernel.Decimal
}

// RateInForce returns the rate of the latest change in changes that is in
// force on day: the one with the latest From on or before day, whatever the
// order of changes. It reports false when none is.
func RateInForce(changes []RateChange, day kernel.Date) (kernel.Decimal, bool) {
	var best *RateChange
	for i := range changes {
		c := &changes[i]
		if c.From.After(day) {
			continue
		}
		if best == nil || c.From.After(best.From) {
			best = c
		}
	}
	if best == nil {
		return kernel.Decimal{}, false
	}
	return best.Rate, true
}
