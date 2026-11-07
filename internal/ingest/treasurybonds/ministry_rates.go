package treasurybonds

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/arkadiuss/openfief/internal/jurisdiction/pl"
	"github.com/arkadiuss/openfief/internal/kernel"
)

// ErrNoRates reports a workbook that holds no series rates in the layout
// the ministry's file uses: no sheet named after a bond type with the
// "Seria" and "Oprocentowanie" headers, or a series listed twice.
var ErrNoRates = errors.New("no retail bond rates in the workbook")

// The headers of the ministry's sheets (D-97): each bond type's sheet, named
// by its code, has a header row naming the series column, the interest
// columns, whose first is the first period's rate, and, for the types that
// have one, the margin column.
const (
	headerSeries = "Seria"
	headerRate   = "Oprocentowanie"
	headerMargin = "Marża"
	// headerRows is how many rows from the top of a sheet the header row
	// may be on.
	headerRows = 5
	// ratePlaces is the decimal places a rate is rounded to: the file holds
	// doubles, and a fraction of a percent has at most four.
	ratePlaces = 6
)

// bondTypes returns the retail bond types whose sheets are read: those on
// sale since the oldest series the fee table covers
// (pl.EarlyRedemptionFees).
func bondTypes() []pl.BondType {
	return []pl.BondType{pl.OTS, pl.ROR, pl.DOR, pl.TOS, pl.COI, pl.EDO, pl.ROS, pl.ROD}
}

// ParseWorkbook reads the first-period rate and margin of every series the
// ministry's retail bond workbook (an .xls file) lists, in each bond type's
// sheet. A row without a series code of the sheet's type, or without a
// rate, such as a sub-header, is skipped; so is an indexed type's row
// without a margin. Rates are rounded half even to six decimal places; a
// rate or margin below zero or of 100% or more, which no fraction the file
// holds can be, fails, since the file would then hold percentages. A file
// that is no workbook fails with ErrNotWorkbook, one too large with
// ErrTooLarge, and one without any rate with ErrNoRates.
func ParseWorkbook(data []byte) ([]pl.PublishedRates, error) {
	stream, err := readStream(data, "Workbook")
	if err != nil {
		return nil, err
	}
	wb, err := readWorkbook(stream)
	if err != nil {
		return nil, err
	}
	var out []pl.PublishedRates
	seen := map[pl.Series]bool{}
	for _, t := range bondTypes() {
		sh, ok := wb[t.String()]
		if !ok {
			continue
		}
		rates, err := sheetRates(sh, t)
		if err != nil {
			return nil, fmt.Errorf("sheet %s: %w", t, err)
		}
		for _, r := range rates {
			if seen[r.Series] {
				return nil, fmt.Errorf("%w: %s is listed twice", ErrNoRates, r.Series)
			}
			seen[r.Series] = true
			out = append(out, r)
		}
	}
	if len(out) == 0 {
		return nil, ErrNoRates
	}
	return out, nil
}

// columns are the columns of a sheet's headers; margin is -1 for none.
type columns struct {
	header, series, rate, margin int
}

// headerColumns finds the header row of sh among its first rows and the
// columns of the series, the first rate and the margin.
func headerColumns(sh sheet) (columns, bool) {
	for row := range headerRows {
		cols := columns{header: row, series: -1, rate: -1, margin: -1}
		for col, c := range sh[row] {
			switch {
			case is(c, headerSeries):
				cols.series = col
			case is(c, headerRate):
				cols.rate = col
			case is(c, headerMargin):
				cols.margin = col
			}
		}
		if cols.series >= 0 && cols.rate >= 0 {
			return cols, true
		}
	}
	return columns{}, false
}

// is reports whether c is the header text h, ignoring case and spaces
// around it.
func is(c cell, h string) bool { return !c.number && strings.EqualFold(strings.TrimSpace(c.text), h) }

// sheetRates reads the rates of the series of type t in sh, in row order.
func sheetRates(sh sheet, t pl.BondType) ([]pl.PublishedRates, error) {
	cols, ok := headerColumns(sh)
	if !ok {
		return nil, fmt.Errorf("%w: no %q and %q headers", ErrNoRates, headerSeries, headerRate)
	}
	st, err := t.Structure()
	if err != nil {
		return nil, fmt.Errorf("bond type %s: %w", t, err)
	}
	indexed := st.Indexation != pl.Fixed
	rows := make([]int, 0, len(sh))
	for row := range sh {
		if row > cols.header {
			rows = append(rows, row)
		}
	}
	slices.Sort(rows)
	var out []pl.PublishedRates
	for _, row := range rows {
		r, ok, err := rowRates(sh[row], cols, t, indexed)
		if err != nil {
			return nil, fmt.Errorf("row %d: %w", row+1, err)
		}
		if ok {
			out = append(out, r)
		}
	}
	return out, nil
}

// rowRates reads one row's series and rates, and reports whether the row
// holds them.
func rowRates(row map[int]cell, cols columns, t pl.BondType, indexed bool) (pl.PublishedRates, bool, error) {
	code, rate := row[cols.series], row[cols.rate]
	if code.number || !rate.number {
		return pl.PublishedRates{}, false, nil
	}
	s, ok := seriesOf(code.text, t)
	if !ok {
		return pl.PublishedRates{}, false, nil
	}
	var err error
	out := pl.PublishedRates{Series: s}
	if out.FirstPeriodRate, err = fraction(rate.num); err != nil {
		return pl.PublishedRates{}, false, fmt.Errorf("%s first-period rate: %w", s, err)
	}
	if !indexed {
		return out, true, nil
	}
	margin, ok := row[cols.margin]
	if cols.margin < 0 || !ok || !margin.number {
		return pl.PublishedRates{}, false, nil
	}
	if out.Margin, err = fraction(margin.num); err != nil {
		return pl.PublishedRates{}, false, fmt.Errorf("%s margin: %w", s, err)
	}
	return out, true, nil
}

// seriesOf reads text as a series code of type t, and reports whether it
// is one.
func seriesOf(text string, t pl.BondType) (pl.Series, bool) {
	s, err := pl.ParseSeries(strings.TrimSpace(text))
	return s, err == nil && s.Type == t
}

// ErrNotFraction reports a rate that is no yearly fraction from 0 up to 1.
var ErrNotFraction = errors.New("not a yearly fraction from 0 up to 1")

// fraction returns d rounded half even to ratePlaces, which must be from
// zero up to, not including, one.
func fraction(d kernel.Decimal) (kernel.Decimal, error) {
	r, err := d.Round(ratePlaces, kernel.RoundHalfEven)
	if err != nil {
		return kernel.Decimal{}, fmt.Errorf("rounding: %w", err)
	}
	if r.Sign() < 0 || r.Cmp(kernel.NewDecimal(1, 0)) >= 0 {
		return kernel.Decimal{}, fmt.Errorf("%w: %s", ErrNotFraction, r)
	}
	return r, nil
}
