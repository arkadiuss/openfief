package xtb

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"path"
	"strconv"
	"strings"
)

// Errors reading the workbook.
var (
	// ErrNotWorkbook reports a file that is not a readable .xlsx workbook.
	ErrNotWorkbook = errors.New("not an .xlsx workbook")
	// ErrTooLarge reports a workbook beyond the reader's limits.
	ErrTooLarge = errors.New("workbook too large")
)

// Limits of what the reader takes from one file, against hostile uploads
// such as a zip bomb.
const (
	// MaxFileBytes caps the uploaded file, as the core does (D-109).
	MaxFileBytes = 64 << 20
	// maxPartBytes caps one uncompressed part of the workbook.
	maxPartBytes = 128 << 20
	// maxTotalBytes caps all the parts read, uncompressed.
	maxTotalBytes = 256 << 20
	// maxParts caps the zip's entries.
	maxParts = 1000
	// maxRows caps the rows of one sheet.
	maxRows = 200_000
	// maxColumns caps the column of any cell (column XFD is 16384).
	maxColumns = 16384
	// letters is the number of column letters, A to Z.
	letters = 26
)

// cellKind is how a cell stores its value.
type cellKind int

// The kinds of cell the reader tells apart.
const (
	cellEmpty cellKind = iota
	cellText
	cellNumber
	cellBool
	cellError
	cellISODate
)

// cell is one cell's value as the workbook stores it: text as text, a
// number as the decimal string written in the file, never converted
// through a float.
type cell struct {
	kind  cellKind
	value string
}

// sheet is one worksheet's rows, by row number from 1, each a slice of
// cells by column from 0.
type sheet struct {
	name string
	rows map[int][]cell
	last int
}

// row returns the cells of row n, or none.
func (s *sheet) row(n int) []cell { return s.rows[n] }

// workbook is the sheets of a file by name, and whether its dates count
// from 1904.
type workbook struct {
	sheets   map[string]*sheet
	date1904 bool
}

// zipParts reads the zip's parts within the limits.
type zipParts struct {
	files map[string]*zip.File
	read  int64
}

// openZip returns the zip of data, with its parts by name.
func openZip(data []byte) (*zipParts, error) {
	if len(data) > MaxFileBytes {
		return nil, fmt.Errorf("%w: more than %d bytes", ErrTooLarge, MaxFileBytes)
	}
	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrNotWorkbook, err)
	}
	if len(zr.File) > maxParts {
		return nil, fmt.Errorf("%w: more than %d parts", ErrTooLarge, maxParts)
	}
	z := &zipParts{files: make(map[string]*zip.File, len(zr.File))}
	for _, f := range zr.File {
		z.files[strings.TrimPrefix(f.Name, "/")] = f
	}
	return z, nil
}

// part returns the uncompressed bytes of the part named name, or nil and
// no error when there is none.
func (z *zipParts) part(name string) ([]byte, error) {
	f, ok := z.files[name]
	if !ok {
		return nil, nil
	}
	rc, err := f.Open()
	if err != nil {
		return nil, fmt.Errorf("%w: %s: %w", ErrNotWorkbook, name, err)
	}
	b, err := io.ReadAll(io.LimitReader(rc, maxPartBytes+1))
	closeErr := rc.Close()
	switch {
	case err != nil:
		return nil, fmt.Errorf("%w: %s: %w", ErrNotWorkbook, name, err)
	case closeErr != nil:
		return nil, fmt.Errorf("%w: %s: %w", ErrNotWorkbook, name, closeErr)
	case len(b) > maxPartBytes:
		return nil, fmt.Errorf("%w: part %s is beyond %d bytes", ErrTooLarge, name, maxPartBytes)
	}
	z.read += int64(len(b))
	if z.read > maxTotalBytes {
		return nil, fmt.Errorf("%w: beyond %d bytes uncompressed", ErrTooLarge, maxTotalBytes)
	}
	return b, nil
}

// The parts of the workbook as SpreadsheetML (ECMA-376) names them.
type xWorkbook struct {
	Pr struct {
		Date1904 string `xml:"date1904,attr"`
	} `xml:"workbookPr"`
	Sheets []struct {
		Name string `xml:"name,attr"`
		RID  string `xml:"http://schemas.openxmlformats.org/officeDocument/2006/relationships id,attr"`
	} `xml:"sheets>sheet"`
}

type xRels struct {
	Rels []struct {
		ID     string `xml:"Id,attr"`
		Target string `xml:"Target,attr"`
	} `xml:"Relationship"`
}

// xText is a string item: plain text or rich text runs.
type xText struct {
	T    string `xml:"t"`
	Runs []struct {
		T string `xml:"t"`
	} `xml:"r"`
}

// String joins the item's text.
func (x *xText) String() string {
	if len(x.Runs) == 0 {
		return x.T
	}
	var b strings.Builder
	for _, r := range x.Runs {
		b.WriteString(r.T)
	}
	return b.String()
}

type xShared struct {
	Items []xText `xml:"si"`
}

// xCell is one cell of a worksheet part.
type xCell struct {
	R  string `xml:"r,attr"`
	T  string `xml:"t,attr"`
	V  string `xml:"v"`
	Is xText  `xml:"is"`
}

// xRow is one row of a worksheet part.
type xRow struct {
	R     int     `xml:"r,attr"`
	Cells []xCell `xml:"c"`
}

type xSheet struct {
	Rows []xRow `xml:"sheetData>row"`
}

// readWorkbook reads the sheets of an .xlsx file.
func readWorkbook(data []byte) (*workbook, error) {
	z, err := openZip(data)
	if err != nil {
		return nil, err
	}
	var wb xWorkbook
	if decodeErr := z.decode("xl/workbook.xml", &wb, true); decodeErr != nil {
		return nil, decodeErr
	}
	targets, err := z.targets()
	if err != nil {
		return nil, err
	}
	strs, err := z.sharedStrings()
	if err != nil {
		return nil, err
	}
	out := &workbook{sheets: map[string]*sheet{}, date1904: wb.Pr.Date1904 == "1" || wb.Pr.Date1904 == "true"}
	for _, s := range wb.Sheets {
		target, ok := targets[s.RID]
		if !ok {
			return nil, fmt.Errorf("%w: sheet %q has no part", ErrNotWorkbook, s.Name)
		}
		sh, err := z.sheet(s.Name, partName(target), strs)
		if err != nil {
			return nil, err
		}
		out.sheets[s.Name] = sh
	}
	return out, nil
}

// targets returns the workbook's relationship targets by ID.
func (z *zipParts) targets() (map[string]string, error) {
	var rels xRels
	if err := z.decode("xl/_rels/workbook.xml.rels", &rels, true); err != nil {
		return nil, err
	}
	targets := make(map[string]string, len(rels.Rels))
	for _, r := range rels.Rels {
		targets[r.ID] = r.Target
	}
	return targets, nil
}

// sharedStrings returns the workbook's shared strings, none without the
// part.
func (z *zipParts) sharedStrings() ([]string, error) {
	var shared xShared
	if err := z.decode("xl/sharedStrings.xml", &shared, false); err != nil {
		return nil, err
	}
	strs := make([]string, len(shared.Items))
	for i := range shared.Items {
		strs[i] = shared.Items[i].String()
	}
	return strs, nil
}

// partName resolves a relationship target of the workbook part.
func partName(target string) string {
	if strings.HasPrefix(target, "/") {
		return strings.TrimPrefix(path.Clean(target), "/")
	}
	return path.Clean(path.Join("xl", target))
}

// decode unmarshals part name into v; a missing part is an error only when
// required.
func (z *zipParts) decode(name string, v any, required bool) error {
	b, err := z.part(name)
	if err != nil {
		return err
	}
	if b == nil {
		if required {
			return fmt.Errorf("%w: no %s", ErrNotWorkbook, name)
		}
		return nil
	}
	if err := xml.Unmarshal(b, v); err != nil {
		return fmt.Errorf("%w: %s: %w", ErrNotWorkbook, name, err)
	}
	return nil
}

// sheet reads the worksheet part name with the shared strings strs.
func (z *zipParts) sheet(name, part string, strs []string) (*sheet, error) {
	var xs xSheet
	if err := z.decode(part, &xs, true); err != nil {
		return nil, err
	}
	if len(xs.Rows) > maxRows {
		return nil, fmt.Errorf("%w: sheet %q has more than %d rows", ErrTooLarge, name, maxRows)
	}
	s := &sheet{name: name, rows: make(map[int][]cell, len(xs.Rows))}
	next := 1
	for i := range xs.Rows {
		xr := &xs.Rows[i]
		n := xr.R
		if n == 0 {
			n = next
		}
		if n < 1 || n > maxRows {
			return nil, fmt.Errorf("%w: sheet %q: row %d out of range", ErrNotWorkbook, name, n)
		}
		next = n + 1
		cells, err := rowCells(xr, strs)
		if err != nil {
			return nil, fmt.Errorf("%w: sheet %q row %d: %w", ErrNotWorkbook, name, n, err)
		}
		s.rows[n] = cells
		s.last = max(s.last, n)
	}
	return s, nil
}

// rowCells returns a row's cells by column: a cell with a reference is in
// its column, one without follows the previous cell.
func rowCells(xr *xRow, strs []string) ([]cell, error) {
	var cells []cell
	col := 0
	for i := range xr.Cells {
		xc := &xr.Cells[i]
		if xc.R != "" {
			c, err := columnOf(xc.R)
			if err != nil {
				return nil, err
			}
			col = c
		}
		v, err := cellValue(xc, strs)
		if err != nil {
			return nil, err
		}
		for len(cells) <= col {
			cells = append(cells, cell{})
		}
		cells[col] = v
		col++
	}
	return cells, nil
}

// errCell reports a cell the reader can't read.
var errCell = errors.New("unreadable cell")

// cellValue reads a cell by its type.
func cellValue(xc *xCell, strs []string) (cell, error) {
	switch xc.T {
	case "s":
		return sharedString(xc.V, strs)
	case "inlineStr":
		return cell{kind: cellText, value: xc.Is.String()}, nil
	case "str":
		return cell{kind: cellText, value: xc.V}, nil
	case "b":
		return cell{kind: cellBool, value: xc.V}, nil
	case "e":
		return cell{kind: cellError, value: xc.V}, nil
	case "d":
		return cell{kind: cellISODate, value: xc.V}, nil
	case "", "n":
		return number(xc.V), nil
	}
	return cell{}, fmt.Errorf("%w: type %q", errCell, xc.T)
}

// number returns the number cell of v, or an empty cell when v is blank.
func number(v string) cell {
	if v = strings.TrimSpace(v); v == "" {
		return cell{}
	}
	return cell{kind: cellNumber, value: v}
}

// sharedString returns the text cell of shared string index v.
func sharedString(v string, strs []string) (cell, error) {
	i, err := strconv.Atoi(strings.TrimSpace(v))
	if err != nil || i < 0 || i >= len(strs) {
		return cell{}, fmt.Errorf("%w: shared string index", errCell)
	}
	return cell{kind: cellText, value: strs[i]}, nil
}

// columnOf returns the 0-based column of a cell reference such as "AB12".
func columnOf(ref string) (int, error) {
	col := 0
	i := 0
	for ; i < len(ref) && ref[i] >= 'A' && ref[i] <= 'Z'; i++ {
		col = col*letters + int(ref[i]-'A'+1)
		if col > maxColumns {
			return 0, fmt.Errorf("%w: reference", errCell)
		}
	}
	if i == 0 {
		return 0, fmt.Errorf("%w: reference", errCell)
	}
	return col - 1, nil
}
