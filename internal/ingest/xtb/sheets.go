package xtb

import (
	"errors"
	"fmt"
	"math/big"
	"slices"
	"strings"
	"time"

	"github.com/arkadiuss/openfief/internal/kernel"
)

// ErrLayout reports a workbook that isn't laid out as an xStation export
// (R-24): a sheet, header or account number missing, or sheets naming
// different accounts. It never carries a cell's value.
var ErrLayout = errors.New("not an xStation account export")

// The export's sheets and labels (R-24).
const (
	sheetClosed      = "Closed Positions"
	sheetCash        = "Cash Operations"
	sheetOpen        = "Open Positions"
	labelAccount     = "Account number"
	labelReport      = "Data as of report generated"
	labelTotal       = "total"
	summaryMetric    = "Metric"
	summaryCurrency  = "Currency"
	maxHeaderRow     = 20
	maxReasonTextLen = 40
)

// errCellValue reports a cell whose value isn't what its column holds. It
// names the column, never the value.
var errCellValue = errors.New("unreadable value")

// table is a sheet's data under its header row: the column of each header
// label and the data rows' numbers, before any Total row.
type table struct {
	sheet  *sheet
	header int
	cols   map[string]int
	rows   []int
}

// findTable locates the header row of s, the first of rows 1 to 20 holding
// every label of required, and the data rows below it up to a Total row or
// the end. Empty rows are left out.
func findTable(s *sheet, required ...string) (*table, error) {
	for n := 1; n <= maxHeaderRow; n++ {
		cols := labels(s.row(n))
		if !hasAll(cols, required) {
			continue
		}
		t := &table{sheet: s, header: n, cols: cols}
		for r := n + 1; r <= s.last; r++ {
			cells := s.row(r)
			if first, ok := firstText(cells); ok && strings.HasPrefix(strings.ToLower(first), labelTotal) {
				break
			}
			if !empty(cells) {
				t.rows = append(t.rows, r)
			}
		}
		return t, nil
	}
	return nil, fmt.Errorf("%w: sheet %q has no header row with %s", ErrLayout, s.name, strings.Join(required, ", "))
}

// labels returns the column of each text label of a row, the first when a
// label repeats.
func labels(cells row) map[string]int {
	cols := map[string]int{}
	for _, e := range cells {
		l := strings.TrimSpace(e.cell.value)
		if _, dup := cols[l]; e.cell.kind == cellText && l != "" && !dup {
			cols[l] = e.col
		}
	}
	return cols
}

// hasAll reports whether cols has every label.
func hasAll(cols map[string]int, labels []string) bool {
	for _, l := range labels {
		if _, ok := cols[l]; !ok {
			return false
		}
	}
	return true
}

// firstText returns the first non-empty cell's text, if it is text.
func firstText(cells row) (string, bool) {
	for _, e := range cells {
		if e.cell.kind == cellEmpty {
			continue
		}
		return strings.TrimSpace(e.cell.value), e.cell.kind == cellText
	}
	return "", false
}

// empty reports whether every cell is empty or blank text.
func empty(cells row) bool {
	return !slices.ContainsFunc(cells, func(e entry) bool { return strings.TrimSpace(e.cell.value) != "" })
}

// record is one data row of a table.
type record struct {
	t *table
	n int
}

// records returns the table's data rows in sheet order.
func (t *table) records() []record {
	out := make([]record, len(t.rows))
	for i, n := range t.rows {
		out[i] = record{t: t, n: n}
	}
	return out
}

// has reports whether the table has column label.
func (t *table) has(label string) bool {
	_, ok := t.cols[label]
	return ok
}

// cell returns the record's cell in column label, empty when the table has
// no such column.
func (r record) cell(label string) cell {
	i, ok := r.t.cols[label]
	if !ok {
		return cell{}
	}
	return r.t.sheet.row(r.n).at(i)
}

// text returns the trimmed text of column label, or a number as written.
func (r record) text(label string) string { return strings.TrimSpace(r.cell(label).value) }

// errorf returns an error naming the record's column label.
func (r record) errorf(label string) error {
	return fmt.Errorf("%w in column %q", errCellValue, label)
}

// decimal returns the exact decimal written in column label: a number cell
// as the file writes it, or text holding a decimal, either of them possibly
// in scientific notation such as 1.5E-7, within kernel.ParseDecimal's bounds.
func (r record) decimal(label string) (kernel.Decimal, error) {
	c := r.cell(label)
	if c.kind != cellNumber && c.kind != cellText {
		return kernel.Decimal{}, r.errorf(label)
	}
	d, err := kernel.ParseDecimal(strings.TrimSpace(c.value))
	if err != nil {
		return kernel.Decimal{}, r.errorf(label)
	}
	return d, nil
}

// id returns the institution's ID in column label, text in Cash Operations
// and a number in Closed Positions (R-24), as digits.
func (r record) id(label string) (string, error) {
	c := r.cell(label)
	switch c.kind {
	case cellText:
		return strings.TrimSpace(c.value), nil
	case cellNumber:
		d, err := kernel.ParseDecimal(c.value)
		if err != nil {
			return "", r.errorf(label)
		}
		whole, err := d.Round(0, kernel.RoundDown)
		if err != nil || !whole.Equal(d) || d.Sign() < 0 {
			return "", r.errorf(label)
		}
		return whole.String(), nil
	case cellEmpty:
		return "", nil
	case cellBool, cellError, cellISODate:
	}
	return "", r.errorf(label)
}

// epoch returns day 0 of a workbook's date serials: 1899-12-30 in the 1900
// date system, 1904-01-01 in the 1904 one.
func epoch(date1904 bool) time.Time {
	if date1904 {
		return time.Date(1904, time.January, 1, 0, 0, 0, 0, time.UTC)
	}
	return time.Date(1899, time.December, 30, 0, 0, 0, 0, time.UTC)
}

// secondsPerDay converts a date serial's fraction to seconds.
const secondsPerDay = 86400

// textLayouts are the written date-times the reader takes from a text or
// ISO date cell, all in UTC.
func textLayouts() []string {
	return []string{
		"2006-01-02 15:04:05", "2006-01-02T15:04:05", "2006-01-02T15:04:05Z07:00",
		"2006-01-02 15:04", "2006-01-02", "02.01.2006 15:04:05", "02.01.2006",
	}
}

// date returns the UTC date of the date-time in column label: a date
// serial, rounded to the second, or a written date-time.
func (r record) date(label string, date1904 bool) (kernel.Date, error) {
	t, err := r.instant(label, date1904)
	if err != nil {
		return kernel.Date{}, err
	}
	d, ok := utcDate(t)
	if !ok {
		return kernel.Date{}, r.errorf(label)
	}
	return d, nil
}

// instant returns the UTC date-time in column label, to the second.
func (r record) instant(label string, date1904 bool) (time.Time, error) {
	t, ok := cellTime(r.cell(label), date1904)
	if !ok {
		return time.Time{}, r.errorf(label)
	}
	return t, nil
}

// cellTime returns the UTC date-time of c: a date serial, rounded to the
// second, or a written date-time; false for any other cell.
func cellTime(c cell, date1904 bool) (time.Time, bool) {
	var (
		t  time.Time
		ok bool
	)
	switch c.kind {
	case cellNumber:
		t, ok = serialTime(c.value, date1904)
	case cellText, cellISODate:
		t, ok = parseWritten(strings.TrimSpace(c.value))
	case cellEmpty, cellBool, cellError:
	}
	return t.UTC(), ok
}

// utcDate returns the date of t, which is in UTC.
func utcDate(t time.Time) (kernel.Date, bool) {
	d, err := kernel.NewDate(t.Year(), t.Month(), t.Day())
	return d, err == nil
}

// reportDate returns the UTC date of the report an Open Positions sheet s
// holds, the date-time after "Data as of report generated" in a row above
// its header (R-24), and false when no such row has one.
func reportDate(s *sheet, header int, date1904 bool) (kernel.Date, bool) {
	for n := 1; n < header; n++ {
		v, ok := labelled(s.row(n), labelReport)
		if !ok {
			continue
		}
		t, ok := cellTime(v, date1904)
		if !ok {
			return kernel.Date{}, false
		}
		return utcDate(t)
	}
	return kernel.Date{}, false
}

// labelled returns the first non-empty cell after the text cell label in
// cells, and false when cells hold no such pair.
func labelled(cells row, label string) (cell, bool) {
	i := slices.IndexFunc(cells, func(e entry) bool {
		return e.cell.kind == cellText && strings.TrimSpace(e.cell.value) == label
	})
	if i < 0 {
		return cell{}, false
	}
	j := slices.IndexFunc(cells[i+1:], func(e entry) bool { return e.cell.kind != cellEmpty })
	if j < 0 {
		return cell{}, false
	}
	return cells[i+1+j].cell, true
}

// maxSerialDays bounds a date serial: day 2958465 is 9999-12-31.
const maxSerialDays = 2958465

// serialTime returns the UTC time of date serial v, rounded half up to the
// second, computed exactly from its decimal digits.
func serialTime(v string, date1904 bool) (time.Time, bool) {
	serial, ok := new(big.Rat).SetString(v)
	if !ok || serial.Sign() < 0 || serial.Cmp(big.NewRat(maxSerialDays, 1)) > 0 {
		return time.Time{}, false
	}
	secs := new(big.Rat).Mul(serial, big.NewRat(secondsPerDay, 1))
	half := new(big.Int).Rsh(secs.Denom(), 1)
	whole := new(big.Int).Quo(new(big.Int).Add(secs.Num(), half), secs.Denom()).Int64()
	days, rest := whole/secondsPerDay, whole%secondsPerDay
	return epoch(date1904).AddDate(0, 0, int(days)).Add(time.Duration(rest) * time.Second), true
}

// parseWritten parses a written date-time in one of textLayouts.
func parseWritten(s string) (time.Time, bool) {
	for _, l := range textLayouts() {
		if t, err := time.ParseInLocation(l, s, time.UTC); err == nil {
			return t, true
		}
	}
	return time.Time{}, false
}

// accountNumber returns the account number of row 1, "Account number" and
// the number as text (R-24).
func accountNumber(s *sheet) (string, error) {
	cells := s.row(1)
	for i, e := range cells {
		if e.cell.kind != cellText || strings.TrimSpace(e.cell.value) != labelAccount {
			continue
		}
		for _, v := range cells[i+1:] {
			if n := strings.TrimSpace(v.cell.value); n != "" {
				return n, nil
			}
		}
	}
	return "", fmt.Errorf("%w: sheet %q has no account number in row 1", ErrLayout, s.name)
}

// summaryCurrencies returns the currencies of Open Positions' summary
// table above its header (R-24), each once.
func summaryCurrencies(s *sheet, header int) []string {
	for n := 1; n < header; n++ {
		cols := labels(s.row(n))
		col, ok := cols[summaryCurrency]
		if _, metric := cols[summaryMetric]; !ok || !metric {
			continue
		}
		var out []string
		for m := n + 1; m < header; m++ {
			if v := cellAt(s.row(m), col); v != "" && !slices.Contains(out, v) {
				out = append(out, v)
			}
		}
		return out
	}
	return nil
}

// cellAt returns the trimmed value of the cell in column col, or "".
func cellAt(cells row, col int) string {
	return strings.TrimSpace(cells.at(col).value)
}

// shown returns s for an ingest problem's reason: at most 40 printable
// characters of a label such as an operation type.
func shown(s string) string {
	var b strings.Builder
	for _, c := range s {
		if b.Len() >= maxReasonTextLen {
			b.WriteString("…")
			break
		}
		if c >= ' ' && c != 0x7f {
			b.WriteRune(c)
		}
	}
	return b.String()
}
