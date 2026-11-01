// Package xtbtest writes synthetic xStation "Historia operacji" exports
// laid out as R-24 describes, for tests only. Every value in them is made
// up: account 0000000, tickers such as TESTA.PL, operation IDs such as
// 9000000001. No real export is ever committed.
package xtbtest

import (
	"archive/zip"
	"bytes"
	"fmt"
	"math/big"
	"slices"
	"strconv"
	"strings"
	"time"
)

// Account is the synthetic XTB account number of every export.
const Account = "0000000"

// Cell is one cell: text, a number written as its decimal string (which
// may carry a float artifact, such as "12.300000000000001"), or a date-time
// written as a date serial.
type Cell struct {
	text   string
	number string
	isText bool
}

// Text returns a text cell.
func Text(s string) Cell { return Cell{text: s, isText: true} }

// Num returns a number cell written as s.
func Num(s string) Cell { return Cell{number: s} }

// secondsPerDay converts seconds to a date serial's fraction.
const (
	secondsPerDay = 86400
	serialPlaces  = 10
)

// Time returns a date-time cell of t, in UTC, as a date serial of the 1900
// date system, whose day 0 is 1899-12-30.
func Time(t time.Time) Cell {
	day0 := time.Date(1899, time.December, 30, 0, 0, 0, 0, time.UTC)
	secs := int64(t.UTC().Sub(day0) / time.Second)
	return Cell{number: new(big.Rat).SetFrac64(secs, secondsPerDay).FloatString(serialPlaces)}
}

// Sheet is a named sheet's rows from row 1; Inline writes its text as
// inline strings instead of shared strings.
type Sheet struct {
	Name   string
	Rows   [][]Cell
	Inline bool
}

// CashHeader is the header row of Cash Operations (R-24).
func CashHeader() []string {
	return strings.Split("Type|Instrument|Ticker|Category|Time|Amount|ID|Comment|Product|Position ID", "|")
}

// ClosedHeader is the header row of Closed Positions (R-24).
func ClosedHeader() []string {
	return strings.Split("Instrument|Ticker|Category|Type|Volume|Open Price|Open Time (UTC)|Close Price|"+
		"Close Time (UTC)|Product|Profit/Loss|Gross Profit|Purchase Value|Sale Value|Stop Loss|Take Profit|"+
		"Commission|Margin|Swap|Rollover|Open Conversion Rate|Close Conversion Rate|Close Origin|Position ID|Comment", "|")
}

// OpenHeader is the header row of Open Positions (R-24).
func OpenHeader() []string {
	return strings.Split("Product|Instrument/Position|Ticker|Category|Type|Volume|Value|Current price|Open price|"+
		"Open time (UTC)|Stop Loss|Take Profit|Net Profit %|Net Profit|Gross Profit|Margin|Open Commission|Swap|Rollover", "|")
}

// Export is a synthetic export of one account: its data rows under each
// table's header, Cash rows newest first as xStation writes them.
type Export struct {
	Account  string
	Currency string
	From, To time.Time
	Cash     [][]Cell
	Closed   [][]Cell
	Open     [][]Cell
	// NoTotals leaves out the Total rows that end the tables.
	NoTotals bool
	// Inline writes text as inline strings.
	Inline bool
	// NoReportDate leaves the report's date out of Open Positions.
	NoReportDate bool
}

// FileName returns xStation's name of the export,
// <currency>_<account>_<from>_<to>.xlsx.
func (e *Export) FileName() string {
	return fmt.Sprintf("%s_%s_%s_%s.xlsx", e.Currency, e.Account, e.From.Format(time.DateOnly), e.To.Format(time.DateOnly))
}

// texts returns a row of text cells.
func texts(labels []string) []Cell {
	out := make([]Cell, len(labels))
	for i, l := range labels {
		out[i] = Text(l)
	}
	return out
}

// top returns the rows above a table's header: the account number, the
// title and the period (R-24).
func top(e *Export, title string) [][]Cell {
	return [][]Cell{
		{Text("Account number"), Text(e.Account)},
		{Text(title)},
		{Text("Date from (UTC)"), Time(e.From)},
		{Text("Date to (UTC)"), Time(e.To)},
	}
}

// total returns rows with the Total row that ends a table.
func total(e *Export, rows [][]Cell) [][]Cell {
	if e.NoTotals {
		return rows
	}
	return append(rows, []Cell{Text("Total"), {}, {}, {}, {}, Num("0")})
}

// Bytes writes the export as an .xlsx file.
func (e *Export) Bytes() []byte {
	closed := append(top(e, "Closed Positions"), texts(ClosedHeader()))
	cash := append(top(e, "Cash Operations"), texts(CashHeader()))
	open := slices.Concat([][]Cell{
		{Text("Account number"), Text(e.Account)},
		{Text("Open Positions")},
		report(e),
		{Text("Product"), Text("Metric"), Text("Amount"), Text("Currency")},
		{Text("My Trades"), Text("Open position value"), Num("0"), Text(e.Currency)},
		{Text("My Trades"), Text("Open position profit"), Num("0"), Text(e.Currency)},
		{},
		{Text("Values in the account currency.")},
		texts(OpenHeader()),
	}, e.Open)
	return Workbook(
		Sheet{Name: "Closed Positions", Rows: total(e, append(closed, e.Closed...)), Inline: e.Inline},
		Sheet{Name: "Cash Operations", Rows: total(e, append(cash, e.Cash...)), Inline: e.Inline},
		Sheet{Name: "Open Positions", Rows: total(e, open), Inline: e.Inline},
	)
}

// report returns Open Positions' row with the report's date, the export's
// end, or an empty row with NoReportDate.
func report(e *Export) []Cell {
	if e.NoReportDate {
		return []Cell{}
	}
	return []Cell{Text("Data as of report generated"), Time(e.To)}
}

// CashRow returns a Cash Operations row in CashHeader's order.
func CashRow(kind, name, ticker, category string, at time.Time, amount, id, comment string) []Cell {
	return []Cell{
		Text(kind), Text(name), Text(ticker), Text(category), Time(at), Num(amount), Text(id), Text(comment),
		Text("My Trades"), Text(""),
	}
}

// Position is a position row of Closed Positions or Open Positions. An open
// position's Value is in the account's currency, "0" when empty, and its
// CurrentPrice is its OpenPrice when empty.
type Position struct {
	Name, Ticker, Category string
	Volume, OpenPrice      string
	Opened                 time.Time
	ClosePrice, Profit     string
	Closed                 time.Time
	PositionID             string
	Value, CurrentPrice    string
}

// ClosedRow returns a Closed Positions row in ClosedHeader's order, with
// the position ID as a number.
func ClosedRow(p *Position) []Cell {
	zero := Num("0")
	return []Cell{
		Text(p.Name), Text(p.Ticker), Text(p.Category), Text("BUY"), Num(p.Volume), Num(p.OpenPrice), Time(p.Opened),
		Num(p.ClosePrice), Time(p.Closed), Text("My Trades"), Num(p.Profit), Num(p.Profit), zero, zero, zero,
		zero, zero, zero, zero, zero, Num("1"), Num("1"), Text("xStation"), Num(p.PositionID), Text(""),
	}
}

// OpenRow returns an Open Positions row in OpenHeader's order.
func OpenRow(p *Position) []Cell {
	zero := Num("0")
	value, current := p.Value, p.CurrentPrice
	if value == "" {
		value = "0"
	}
	if current == "" {
		current = p.OpenPrice
	}
	return []Cell{
		Text("My Trades"), Text(p.Name), Text(p.Ticker), Text(p.Category), Text("BUY"), Num(p.Volume), Num(value),
		Num(current), Num(p.OpenPrice), Time(p.Opened), zero, zero, zero, zero, zero, zero, zero, zero, zero,
	}
}

// strTable collects a workbook's shared strings.
type strTable struct {
	list  []string
	index map[string]int
}

// at returns the index of s, adding it.
func (t *strTable) at(s string) int {
	if n, ok := t.index[s]; ok {
		return n
	}
	t.index[s] = len(t.list)
	t.list = append(t.list, s)
	return len(t.list) - 1
}

// Workbook writes sheets as an .xlsx file, as spreadsheet programs do: a
// workbook part naming the sheets, its relationships, shared strings and
// one worksheet part per sheet with cell references. It returns nil if the
// zip can't be written.
func Workbook(sheets ...Sheet) []byte {
	parts := map[string]string{}
	shared := &strTable{index: map[string]int{}}
	var wb, rels strings.Builder
	wb.WriteString(`<?xml version="1.0" encoding="UTF-8"?><workbook xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships"><sheets>`)
	rels.WriteString(`<?xml version="1.0" encoding="UTF-8"?><Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">`)
	for i := range sheets {
		id := "rId" + strconv.Itoa(i+1)
		fmt.Fprintf(&wb, `<sheet name="%s" sheetId="%d" r:id="%s"/>`, escape(sheets[i].Name), i+1, id)
		fmt.Fprintf(&rels, `<Relationship Id="%s" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/worksheet" Target="worksheets/sheet%d.xml"/>`, id, i+1)
		parts[fmt.Sprintf("xl/worksheets/sheet%d.xml", i+1)] = worksheet(&sheets[i], shared)
	}
	wb.WriteString(`</sheets></workbook>`)
	rels.WriteString(`</Relationships>`)
	parts["xl/workbook.xml"] = wb.String()
	parts["xl/_rels/workbook.xml.rels"] = rels.String()
	var ss strings.Builder
	fmt.Fprintf(&ss, `<?xml version="1.0" encoding="UTF-8"?><sst xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main" count="%d" uniqueCount="%d">`, len(shared.list), len(shared.list))
	for _, s := range shared.list {
		fmt.Fprintf(&ss, `<si><t>%s</t></si>`, escape(s))
	}
	ss.WriteString(`</sst>`)
	parts["xl/sharedStrings.xml"] = ss.String()
	parts["[Content_Types].xml"] = `<?xml version="1.0" encoding="UTF-8"?><Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types"/>`
	b, err := zipOf(parts)
	if err != nil {
		return nil
	}
	return b
}

// worksheet writes one sheet's part, its text as shared strings or inline.
func worksheet(s *Sheet, shared *strTable) string {
	var ws strings.Builder
	ws.WriteString(`<?xml version="1.0" encoding="UTF-8"?><worksheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main"><sheetData>`)
	for r, row := range s.Rows {
		fmt.Fprintf(&ws, `<row r="%d">`, r+1)
		for c, cell := range row {
			ref := column(c) + strconv.Itoa(r+1)
			switch {
			case cell.isText && cell.text == "":
			case cell.isText && s.Inline:
				fmt.Fprintf(&ws, `<c r="%s" t="inlineStr"><is><t>%s</t></is></c>`, ref, escape(cell.text))
			case cell.isText:
				fmt.Fprintf(&ws, `<c r="%s" t="s"><v>%d</v></c>`, ref, shared.at(cell.text))
			case cell.number != "":
				fmt.Fprintf(&ws, `<c r="%s"><v>%s</v></c>`, ref, cell.number)
			}
		}
		ws.WriteString(`</row>`)
	}
	ws.WriteString(`</sheetData></worksheet>`)
	return ws.String()
}

// zipOf writes parts into a zip, in name order.
func zipOf(parts map[string]string) ([]byte, error) {
	names := make([]string, 0, len(parts))
	for n := range parts {
		names = append(names, n)
	}
	slices.Sort(names)
	var buf bytes.Buffer
	z := zip.NewWriter(&buf)
	for _, n := range names {
		w, err := z.Create(n)
		if err != nil {
			return nil, fmt.Errorf("create %s: %w", n, err)
		}
		if _, err := w.Write([]byte(parts[n])); err != nil {
			return nil, fmt.Errorf("write %s: %w", n, err)
		}
	}
	if err := z.Close(); err != nil {
		return nil, fmt.Errorf("close zip: %w", err)
	}
	return buf.Bytes(), nil
}

// column returns the letters of 0-based column c.
func column(c int) string {
	const letters = 26
	s := ""
	for c++; c > 0; c = (c - 1) / letters {
		s = string(rune('A'+(c-1)%letters)) + s
	}
	return s
}

// escape returns s escaped for XML text and attributes.
func escape(s string) string {
	return strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;", `"`, "&quot;", "'", "&apos;").Replace(s)
}
