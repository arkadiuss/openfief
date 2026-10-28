package kernel

import (
	"errors"
	"fmt"
	"time"
)

// ErrInvalidDate reports text or fields that are not a calendar date.
var ErrInvalidDate = errors.New("invalid date")

const hoursPerDay = 24

// Date is a calendar date with no time of day and no time zone, such as a
// trade date or an end-of-day price date. Its zero value is no date.
// Dates compare with == and Compare.
type Date struct {
	year  int
	month time.Month
	day   int
}

// NewDate returns the date year-month-day. It rejects dates that do not
// exist, such as 2026-02-30, instead of normalizing them.
func NewDate(year int, month time.Month, day int) (Date, error) {
	t := time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
	if t.Year() != year || t.Month() != month || t.Day() != day {
		return Date{}, fmt.Errorf("%w: %04d-%02d-%02d", ErrInvalidDate, year, int(month), day)
	}
	return Date{year: year, month: month, day: day}, nil
}

// ParseDate parses an ISO 8601 calendar date, YYYY-MM-DD.
func ParseDate(s string) (Date, error) {
	t, err := time.Parse(time.DateOnly, s)
	if err != nil {
		return Date{}, fmt.Errorf("%w: %q: %w", ErrInvalidDate, s, err)
	}
	return dateOf(t), nil
}

// Year returns the year.
func (d Date) Year() int { return d.year }

// Month returns the month.
func (d Date) Month() time.Month { return d.month }

// Day returns the day of the month.
func (d Date) Day() int { return d.day }

// AddDays returns the date n days after d; n may be negative.
func (d Date) AddDays(n int) Date { return dateOf(d.midnight().AddDate(0, 0, n)) }

// DaysUntil returns the number of days from d to e, negative when e is earlier.
func (d Date) DaysUntil(e Date) int {
	return int(e.midnight().Sub(d.midnight()) / (hoursPerDay * time.Hour))
}

// Compare returns -1, 0 or +1 as d is before, equal to or after e.
func (d Date) Compare(e Date) int { return d.midnight().Compare(e.midnight()) }

// Before reports whether d is earlier than e.
func (d Date) Before(e Date) bool { return d.Compare(e) < 0 }

// After reports whether d is later than e.
func (d Date) After(e Date) bool { return d.Compare(e) > 0 }

// Equal reports whether d and e are the same date.
func (d Date) Equal(e Date) bool { return d == e }

// IsZero reports whether d is the zero value, no date.
func (d Date) IsZero() bool { return d == Date{} }

// String formats d as YYYY-MM-DD.
func (d Date) String() string { return d.midnight().Format(time.DateOnly) }

func (d Date) midnight() time.Time {
	return time.Date(d.year, d.month, d.day, 0, 0, 0, 0, time.UTC)
}

func dateOf(t time.Time) Date {
	return Date{year: t.Year(), month: t.Month(), day: t.Day()}
}
