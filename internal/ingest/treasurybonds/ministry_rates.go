package treasurybonds

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/arkadiuss/openfief/internal/jurisdiction/pl"
	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
)

// ErrNoRates reports a workbook that holds no series rates in the layout
// the ministry's file uses: no sheet named after a bond type with the
// "Seria" and "Oprocentowanie" headers, a series listed twice, or one ISIN
// listed for two series.
var ErrNoRates = errors.New("no retail bond rates in the workbook")

// The headers of the ministry's sheets (D-97): each bond type's sheet, named
// by its code, has a header row naming the series column, the registry's
// ISIN column (D-214), the interest columns, whose first is the first
// period's rate, and, for the types that have one, the margin column.
const (
	headerSeries = "Seria"
	headerISIN   = "Kod ISIN"
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

// ParseWorkbook reads the first-period rate and margin, and the registry's
// ISIN (D-214), of every series the ministry's retail bond workbook (an
// .xls file) lists, in each bond type's sheet. A row without a series code
// of the sheet's type, such as a sub-header, is skipped; so are the rates
// of a row without a rate, or of an indexed type's row without a margin,
// such as a new series the file lists before its rates. A series' ISIN is
// read from its "Kod ISIN" cell when that holds a valid Polish ISIN other
// than a provisional one (pl.ProvisionalSeries), and left out otherwise.
// Rates are rounded half even to six decimal places; a rate or margin below
// zero or of 100% or more, which no fraction the file holds can be, fails,
// since the file would then hold percentages. A file that is no workbook
// fails with ErrNotWorkbook, one too large with ErrTooLarge, and one
// without any rate, listing a series twice or one ISIN for two series with
// ErrNoRates.
func ParseWorkbook(data []byte) ([]pl.PublishedRates, []pl.PublishedISIN, error) {
	stream, err := readStream(data, "Workbook")
	if err != nil {
		return nil, nil, err
	}
	wb, err := readWorkbook(stream)
	if err != nil {
		return nil, nil, err
	}
	var rows []seriesRow
	for _, t := range bondTypes() {
		sh, ok := wb[t.String()]
		if !ok {
			continue
		}
		sheet, err := sheetRows(sh, t)
		if err != nil {
			return nil, nil, fmt.Errorf("sheet %s: %w", t, err)
		}
		rows = append(rows, sheet...)
	}
	return publishedRows(rows)
}

// publishedRows returns the rates and ISINs rows list, failing with
// ErrNoRates when a series is listed twice, one ISIN for two series, or no
// rate at all.
func publishedRows(rows []seriesRow) ([]pl.PublishedRates, []pl.PublishedISIN, error) {
	var (
		rates []pl.PublishedRates
		isins []pl.PublishedISIN
	)
	seen := map[pl.Series]bool{}
	for _, r := range rows {
		if seen[r.series] {
			return nil, nil, fmt.Errorf("%w: %s is listed twice", ErrNoRates, r.series)
		}
		seen[r.series] = true
		if r.rates != nil {
			rates = append(rates, *r.rates)
		}
		if r.isin != "" {
			isins = append(isins, pl.PublishedISIN{Series: r.series, ISIN: r.isin})
		}
	}
	if len(rates) == 0 {
		return nil, nil, ErrNoRates
	}
	if err := uniqueISINs(isins); err != nil {
		return nil, nil, err
	}
	return rates, isins, nil
}

// uniqueISINs fails with ErrNoRates when isins lists one ISIN for two
// series, which the registry never assigns.
func uniqueISINs(isins []pl.PublishedISIN) error {
	of := map[string]pl.Series{}
	for _, i := range isins {
		if s, ok := of[i.ISIN]; ok {
			return fmt.Errorf("%w: ISIN %s is listed for %s and %s", ErrNoRates, i.ISIN, s, i.Series)
		}
		of[i.ISIN] = i.Series
	}
	return nil
}

// columns are the columns of a sheet's headers; isin and margin are -1 for
// none.
type columns struct {
	header, series, isin, rate, margin int
}

// headerColumns finds the header row of sh among its first rows and the
// columns of the series, its ISIN, the first rate and the margin.
func headerColumns(sh sheet) (columns, bool) {
	for row := range headerRows {
		cols := columns{header: row, series: -1, isin: -1, rate: -1, margin: -1}
		for col, c := range sh[row] {
			switch {
			case is(c, headerSeries):
				cols.series = col
			case is(c, headerISIN):
				cols.isin = col
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

// seriesRow is what one row of a sheet lists of its series: its rates, nil
// when the row has none yet, and its registry ISIN, empty for none.
type seriesRow struct {
	series pl.Series
	rates  *pl.PublishedRates
	isin   string
}

// sheetRows reads the series of type t in sh that list rates or an ISIN,
// in row order.
func sheetRows(sh sheet, t pl.BondType) ([]seriesRow, error) {
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
	var out []seriesRow
	for _, row := range rows {
		sr, ok, err := rowOf(sh[row], cols, t, indexed)
		if err != nil {
			return nil, fmt.Errorf("row %d: %w", row+1, err)
		}
		if ok {
			out = append(out, sr)
		}
	}
	return out, nil
}

// rowOf reads one row of a sheet of type t: its series' rates and ISIN,
// and whether it lists either.
func rowOf(row map[int]cell, cols columns, t pl.BondType, indexed bool) (seriesRow, bool, error) {
	code := row[cols.series]
	if code.number {
		return seriesRow{}, false, nil
	}
	s, ok := seriesOf(code.text, t)
	if !ok {
		return seriesRow{}, false, nil
	}
	r, ok, err := rowRates(row, cols, s, indexed)
	if err != nil {
		return seriesRow{}, false, err
	}
	sr := seriesRow{series: s, isin: rowISIN(row, cols)}
	if ok {
		sr.rates = &r
	}
	return sr, sr.rates != nil || sr.isin != "", nil
}

// rowISIN returns the registry ISIN in row's ISIN cell: a valid Polish
// ISIN that is no provisional one (D-214), or empty for anything else, such
// as no cell or a dash.
func rowISIN(row map[int]cell, cols columns) string {
	if cols.isin < 0 {
		return ""
	}
	c, ok := row[cols.isin]
	if !ok || c.number {
		return ""
	}
	text := strings.TrimSpace(c.text)
	isin, err := ledger.ParseISIN(text)
	if err != nil || pl.CheckBondISIN(text) != nil {
		return ""
	}
	if _, provisional := pl.ProvisionalSeries(text); provisional {
		return ""
	}
	return isin.String()
}

// rowRates reads the rates of series s in row, and reports whether the row
// holds them.
func rowRates(row map[int]cell, cols columns, s pl.Series, indexed bool) (pl.PublishedRates, bool, error) {
	rate := row[cols.rate]
	if !rate.number {
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
