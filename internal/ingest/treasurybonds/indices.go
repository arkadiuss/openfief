package treasurybonds

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/arkadiuss/openfief/internal/jurisdiction/pl"
	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/valuation"
)

// CPISource supplies GUS's year-on-year CPI readings for the months of a
// year, in month order, final ones only; internal/valuation/gus implements
// it.
type CPISource interface {
	CPI(ctx context.Context, year int) ([]valuation.IndexReading, error)
}

// ReferenceRateSource supplies every change of the NBP reference rate;
// internal/valuation/nbp's Archive implements it.
type ReferenceRateSource interface {
	ReferenceRates(ctx context.Context) ([]pl.RateChange, error)
}

const (
	// maxCPIYearsPerSync bounds the years of CPI one sync asks GUS for, one
	// request each, to stay under its anonymous limit of 5 requests a
	// second; a first sync's history arrives over several syncs.
	maxCPIYearsPerSync = 4
	// defaultIndexYears is how many years before today's a first sync's
	// readings start, without a configured start: ROD, the longest bond
	// type, runs 12 years, and its later periods take CPI from two months
	// before they start.
	defaultIndexYears = 13

	// referenceRateSource names NBP's reference rate on its readings; the
	// core records each reading under the connector instance's name anyway.
	referenceRateSource = "nbp-ref"
	monthsPerYear       = 12

	watermarkCPI  = "cpi"
	watermarkRate = "reference-rate"
	monthLayout   = "2006-01"
)

// ErrBadWatermark reports a sync watermark this connector never returned.
var ErrBadWatermark = errors.New("not a treasury-bonds index watermark")

// IndexWatermark is the treasury-bonds connector's sync watermark: the last
// month whose CPI and whose reference-rate reading it delivered, each as
// that month's first day, or zero before the first.
type IndexWatermark struct {
	CPI           kernel.Date
	ReferenceRate kernel.Date
}

// String writes the watermark as "cpi=2026-09;reference-rate=2026-11",
// leaving out an index with no month; the zero watermark is "".
func (w IndexWatermark) String() string {
	var parts []string
	if !w.CPI.IsZero() {
		parts = append(parts, watermarkCPI+"="+monthOf(w.CPI))
	}
	if !w.ReferenceRate.IsZero() {
		parts = append(parts, watermarkRate+"="+monthOf(w.ReferenceRate))
	}
	return strings.Join(parts, ";")
}

// monthOf writes d's year and month as YYYY-MM.
func monthOf(d kernel.Date) string { return fmt.Sprintf("%04d-%02d", d.Year(), int(d.Month())) }

// ParseIndexWatermark reads a watermark as String writes it. The empty
// watermark, a first sync's, is the zero one.
func ParseIndexWatermark(s string) (IndexWatermark, error) {
	var w IndexWatermark
	if s == "" {
		return w, nil
	}
	for part := range strings.SplitSeq(s, ";") {
		key, month, _ := strings.Cut(part, "=")
		first, ok := parseMonth(month)
		if !ok {
			return IndexWatermark{}, fmt.Errorf("%w: %q", ErrBadWatermark, s)
		}
		switch {
		case key == watermarkCPI && w.CPI.IsZero():
			w.CPI = first
		case key == watermarkRate && w.ReferenceRate.IsZero():
			w.ReferenceRate = first
		default:
			return IndexWatermark{}, fmt.Errorf("%w: %q", ErrBadWatermark, s)
		}
	}
	return w, nil
}

// parseMonth reads YYYY-MM as that month's first day.
func parseMonth(s string) (kernel.Date, bool) {
	t, err := time.Parse(monthLayout, s)
	if err != nil {
		return kernel.Date{}, false
	}
	return firstOfMonth(t.Year(), t.Month()), true
}

// DefaultIndexStart is the first month of readings a first sync delivers
// when no start is configured: January, defaultIndexYears years before
// today's year.
func DefaultIndexStart(today kernel.Date) kernel.Date {
	return firstOfMonth(today.Year()-defaultIndexYears, time.January)
}

// firstOfMonth is the first day of month in year, normalizing a month out
// of range into the next or previous year.
func firstOfMonth(year int, month time.Month) kernel.Date {
	t := time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)
	d, err := kernel.NewDate(t.Year(), t.Month(), 1)
	if err != nil { // the first of a normalized month is always valid
		return kernel.Date{}
	}
	return d
}

// ReferenceReadings returns one NBP reference-rate reading for each month
// from the month of from whose reference-rate day (pl.ReferenceRateDay) is
// before today: the rate in force on that day, keyed by the month whose ROR
// and DOR interest periods it sets. A month with no rate in force on its
// day, before the first change in changes, has none. The day must be over,
// not today, so that a change taking effect on it is already in changes.
func ReferenceReadings(changes []pl.RateChange, from, today kernel.Date) ([]valuation.IndexReading, error) {
	var out []valuation.IndexReading
	for m := firstOfMonth(from.Year(), from.Month()); ; m = firstOfMonth(m.Year(), m.Month()+1) {
		day, err := pl.ReferenceRateDay(m.Year(), m.Month())
		if err != nil {
			return nil, fmt.Errorf("reference rate for %s: %w", monthOf(m), err)
		}
		if !day.Before(today) {
			return out, nil
		}
		if rate, ok := pl.RateInForce(changes, day); ok {
			out = append(out, valuation.IndexReading{
				Kind: valuation.ReferenceRate, Year: m.Year(), Month: m.Month(), Value: rate, Source: referenceRateSource,
			})
		}
	}
}

// IndexSync fetches the period index readings treasury bonds need (D-9):
// GUS's CPI and the NBP reference rate, each for the months after the
// watermark.
type IndexSync struct {
	CPI   CPISource
	Rates ReferenceRateSource
	// Start is the first month a first sync delivers readings for.
	Start kernel.Date
}

// Readings returns the CPI and reference-rate readings for the months after
// watermark, as of today, and the watermark of the last month of each it
// returns. CPI comes for at most maxCPIYearsPerSync years and stops at the
// first month GUS has not published in final form; the reference rate comes
// for every month whose reference-rate day is over, and NBP is asked only
// when there is one. A watermark this connector never wrote fails with
// ErrBadWatermark; a failed request fails the whole sync.
func (s *IndexSync) Readings(ctx context.Context, watermark string, today kernel.Date) ([]valuation.IndexReading, string, error) {
	w, err := ParseIndexWatermark(watermark)
	if err != nil {
		return nil, "", err
	}
	cpi, err := s.cpi(ctx, s.after(w.CPI), today)
	if err != nil {
		return nil, "", err
	}
	rates, err := s.referenceRates(ctx, s.after(w.ReferenceRate), today)
	if err != nil {
		return nil, "", err
	}
	if len(cpi) > 0 {
		last := cpi[len(cpi)-1]
		w.CPI = firstOfMonth(last.Year, last.Month)
	}
	if len(rates) > 0 {
		last := rates[len(rates)-1]
		w.ReferenceRate = firstOfMonth(last.Year, last.Month)
	}
	return append(cpi, rates...), w.String(), nil
}

// after is the first month to deliver after last, the watermark of one
// index: Start's month before the first sync.
func (s *IndexSync) after(last kernel.Date) kernel.Date {
	if last.IsZero() {
		return firstOfMonth(s.Start.Year(), s.Start.Month())
	}
	return firstOfMonth(last.Year(), last.Month()+1)
}

// cpi returns GUS's CPI readings from the month of from, year by year, up
// to today's year and at most maxCPIYearsPerSync years; a year GUS has not
// published in full ends them.
func (s *IndexSync) cpi(ctx context.Context, from, today kernel.Date) ([]valuation.IndexReading, error) {
	var out []valuation.IndexReading
	for y := from.Year(); y <= today.Year() && y < from.Year()+maxCPIYearsPerSync; y++ {
		readings, err := s.CPI.CPI(ctx, y)
		if err != nil {
			return nil, fmt.Errorf("CPI for %d: %w", y, err)
		}
		for _, r := range readings {
			if !firstOfMonth(r.Year, r.Month).Before(from) {
				out = append(out, r)
			}
		}
		if len(readings) < monthsPerYear {
			break
		}
	}
	return out, nil
}

// referenceRates returns the reference-rate readings from the month of
// from, asking NBP only when that month's reference-rate day is over.
func (s *IndexSync) referenceRates(ctx context.Context, from, today kernel.Date) ([]valuation.IndexReading, error) {
	day, err := pl.ReferenceRateDay(from.Year(), from.Month())
	if err != nil {
		return nil, fmt.Errorf("reference rate for %s: %w", monthOf(from), err)
	}
	if !day.Before(today) {
		return nil, nil
	}
	changes, err := s.Rates.ReferenceRates(ctx)
	if err != nil {
		return nil, fmt.Errorf("reference rates: %w", err)
	}
	return ReferenceReadings(changes, from, today)
}
