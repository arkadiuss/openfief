package xtb

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"path"
	"slices"
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
// such as a zip bomb or a sheet of empty cells in its last column (D-184).
// The parts are streamed through the XML decoder, never held whole, and no
// token may outgrow maxTokenBytes, so what the reader allocates is bounded
// by what it keeps: the cells and their text. The limits are set so the
// xtb plugin maps the largest file it takes within its pod's 256 MiB memory
// limit (internal/plugins/operator/objects.go), file and delivery included.
const (
	// MaxFileBytes caps the uploaded file: an export of 50,000 operations
	// is a few MiB. The core takes files up to 64 MiB (D-109); the xtb
	// plugin refuses one over this cap.
	MaxFileBytes = 16 << 20
	// maxPartBytes caps one uncompressed part of the workbook.
	maxPartBytes = 128 << 20
	// maxTotalBytes caps all the parts read, uncompressed.
	maxTotalBytes = 256 << 20
	// maxParts caps the zip's entries, counted before the zip is opened:
	// opening it reads every entry of its directory.
	maxParts = 1000
	// maxSheets caps the sheets a workbook lists; an export has three.
	maxSheets = 100
	// maxRows caps the rows of one sheet.
	maxRows = 200_000
	// maxColumns caps the column of any cell (column XFD is 16384).
	maxColumns = 16384
	// maxCells caps the cells with a value the reader keeps, all sheets
	// together: an export of 50,000 operations, at ten columns a row, has
	// half as many. Only rows with a cell are kept, so it caps the rows
	// kept too. With maxTextBytes it bounds the workbook the reader keeps
	// to about 70 MiB of heap.
	maxCells = 500_000
	// maxSharedStrings caps the shared strings.
	maxSharedStrings = 500_000
	// maxTextBytes caps the text the reader keeps: the shared strings and
	// every cell's own value, all sheets together.
	maxTextBytes = 16 << 20
	// maxTokenBytes caps one XML token as written in the part: a start tag
	// with all its attributes, a text, a comment. It is what the XML
	// decoder holds at once. Excel's own limit for a cell is 32,767
	// characters.
	maxTokenBytes = 256 << 10
	// readAhead is how far the XML decoder reads past the token it is on:
	// the buffer it wraps a reader in.
	readAhead = 4096
	// maxDepth caps how deep elements nest; SpreadsheetML needs five.
	maxDepth = 32
	// letters is the number of column letters, A to Z.
	letters = 26
)

// limits are the reader's bounds on what it keeps and how it parses; tests
// shrink them to reach each one with a tiny workbook.
type limits struct {
	sheets, rows, cells, sharedStrings, textBytes, tokenBytes, depth int
}

// defaultLimits are the limits of every upload.
func defaultLimits() limits {
	return limits{
		sheets: maxSheets, rows: maxRows, cells: maxCells, sharedStrings: maxSharedStrings,
		textBytes: maxTextBytes, tokenBytes: maxTokenBytes, depth: maxDepth,
	}
}

// centralHeader is the signature of a zip central directory header.
const centralHeader = "PK\x01\x02"

// The namespace of a relationship ID attribute, r:id.
const nsRelationships = "http://schemas.openxmlformats.org/officeDocument/2006/relationships"

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

// entry is a cell with a value and its column, from 0.
type entry struct {
	col  int
	cell cell
}

// row is one row's cells with a value, in column order, each column once.
// An empty cell is not kept, so a cell in column XFD costs one entry, not
// 16,384.
type row []entry

// at returns the cell in column col, empty when the row has none there.
func (r row) at(col int) cell {
	i, ok := slices.BinarySearchFunc(r, col, func(e entry, c int) int { return e.col - c })
	if !ok {
		return cell{}
	}
	return r[i].cell
}

// sheet is one worksheet's rows, by row number from 1.
type sheet struct {
	name string
	rows map[int]row
	last int
}

// row returns the cells of row n, or none.
func (s *sheet) row(n int) row { return s.rows[n] }

// workbook is the sheets of a file by name, and whether its dates count
// from 1904.
type workbook struct {
	sheets   map[string]*sheet
	date1904 bool
}

// zipParts reads the zip's parts within the limits, and counts what the
// reader keeps against them.
type zipParts struct {
	files map[string]*zip.File
	lim   limits
	read  int64
	cells int
	text  int
}

// openZip returns the zip of data, with its parts by name, to read within
// lim.
func openZip(data []byte, lim limits) (*zipParts, error) {
	if len(data) > MaxFileBytes {
		return nil, fmt.Errorf("%w: more than %d bytes", ErrTooLarge, MaxFileBytes)
	}
	// Every entry zip.NewReader reads starts with the central directory
	// header's signature, so counting them bounds what it allocates.
	if bytes.Count(data, []byte(centralHeader)) > maxParts {
		return nil, fmt.Errorf("%w: more than %d parts", ErrTooLarge, maxParts)
	}
	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrNotWorkbook, err)
	}
	if len(zr.File) > maxParts {
		return nil, fmt.Errorf("%w: more than %d parts", ErrTooLarge, maxParts)
	}
	z := &zipParts{files: make(map[string]*zip.File, len(zr.File)), lim: lim}
	for _, f := range zr.File {
		z.files[strings.TrimPrefix(f.Name, "/")] = f
	}
	return z, nil
}

// keepText counts n more bytes of text kept, failing past the limit.
func (z *zipParts) keepText(n int) error {
	z.text += n
	if z.text > z.lim.textBytes {
		return fmt.Errorf("%w: more than %d bytes of text", ErrTooLarge, z.lim.textBytes)
	}
	return nil
}

// keepCell counts one more cell kept, failing past the limit.
func (z *zipParts) keepCell() error {
	z.cells++
	if z.cells > z.lim.cells {
		return fmt.Errorf("%w: more than %d cells", ErrTooLarge, z.lim.cells)
	}
	return nil
}

// partReader streams one part: it fails past maxPartBytes of the part,
// past maxTotalBytes of all parts, and once the XML token being read is
// longer than the limit's tokenBytes.
type partReader struct {
	r    io.Reader
	z    *zipParts
	name string
	n    int64
	// token is the offset in the part where the token being read starts.
	token int64
}

// Read reads from the part within the limits.
func (p *partReader) Read(b []byte) (int, error) {
	n, err := p.r.Read(b)
	p.n += int64(n)
	p.z.read += int64(n)
	switch {
	case p.n > maxPartBytes:
		return 0, fmt.Errorf("%w: part %s is beyond %d bytes", ErrTooLarge, p.name, maxPartBytes)
	case p.z.read > maxTotalBytes:
		return 0, fmt.Errorf("%w: beyond %d bytes uncompressed", ErrTooLarge, maxTotalBytes)
	}
	if p.n-p.token > int64(p.z.lim.tokenBytes+readAhead) {
		return 0, fmt.Errorf("%w: part %s has a token of more than %d bytes", ErrTooLarge, p.name, p.z.lim.tokenBytes)
	}
	switch {
	case err == nil:
		return n, nil
	case errors.Is(err, io.EOF):
		return n, io.EOF
	}
	return n, fmt.Errorf("%w: %s: %w", ErrNotWorkbook, p.name, err)
}

// walker reads a part's XML tokens and tracks the path of elements open,
// at most depth deep.
type walker struct {
	d     *xml.Decoder
	p     *partReader
	path  []string
	name  string
	depth int
}

// next returns the next token, io.EOF at the end of the part. A start
// element is on path when returned, an end element already off it.
func (w *walker) next() (xml.Token, error) {
	tok, err := w.d.Token()
	w.p.token = w.d.InputOffset()
	if errors.Is(err, io.EOF) {
		return nil, io.EOF
	}
	if err != nil {
		if errors.Is(err, ErrTooLarge) || errors.Is(err, ErrNotWorkbook) {
			return nil, fmt.Errorf("%s: %w", w.name, err)
		}
		return nil, fmt.Errorf("%w: %s: %w", ErrNotWorkbook, w.name, err)
	}
	switch t := tok.(type) {
	case xml.StartElement:
		if len(w.path) >= w.depth {
			return nil, fmt.Errorf("%w: %s nests deeper than %d", ErrTooLarge, w.name, w.depth)
		}
		w.path = append(w.path, t.Name.Local)
	case xml.EndElement:
		w.path = w.path[:len(w.path)-1]
	}
	return tok, nil
}

// in reports whether the innermost open elements are names, outermost
// first.
func (w *walker) in(names ...string) bool {
	if len(w.path) < len(names) {
		return false
	}
	return slices.Equal(w.path[len(w.path)-len(names):], names)
}

// walk calls fn with each token of part name, through a walker; a missing
// part is an error only when required.
func (z *zipParts) walk(name string, required bool, fn func(w *walker, tok xml.Token) error) error {
	f, ok := z.files[name]
	if !ok {
		if required {
			return fmt.Errorf("%w: no %s", ErrNotWorkbook, name)
		}
		return nil
	}
	rc, err := f.Open()
	if err != nil {
		return fmt.Errorf("%w: %s: %w", ErrNotWorkbook, name, err)
	}
	p := &partReader{r: rc, z: z, name: name}
	w := &walker{d: xml.NewDecoder(p), p: p, name: name, depth: z.lim.depth}
	err = w.each(fn)
	if closeErr := rc.Close(); closeErr != nil && err == nil {
		err = fmt.Errorf("%w: %s: %w", ErrNotWorkbook, name, closeErr)
	}
	return err
}

// each calls fn with each token to the end of the part, stopping at the
// first error.
func (w *walker) each(fn func(w *walker, tok xml.Token) error) error {
	for {
		tok, err := w.next()
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return err
		}
		if err := fn(w, tok); err != nil {
			return err
		}
	}
}

// attr returns the value of the attribute local in namespace space, "" for
// none.
func attr(t xml.StartElement, space, local string) string {
	for _, a := range t.Attr {
		if a.Name.Local == local && a.Name.Space == space {
			return a.Value
		}
	}
	return ""
}

// sheetRef is a sheet the workbook lists: its name and relationship ID.
type sheetRef struct {
	name, rID string
}

// readWorkbook reads the sheets of an .xlsx file within the default limits.
func readWorkbook(data []byte) (*workbook, error) {
	return readWorkbookWithin(data, defaultLimits())
}

// readWorkbookWithin reads the sheets of an .xlsx file within lim.
func readWorkbookWithin(data []byte, lim limits) (*workbook, error) {
	z, err := openZip(data, lim)
	if err != nil {
		return nil, err
	}
	refs, date1904, err := z.workbookPart()
	if err != nil {
		return nil, err
	}
	targets, err := z.targets()
	if err != nil {
		return nil, err
	}
	strs, err := z.sharedStrings()
	if err != nil {
		return nil, err
	}
	out := &workbook{sheets: map[string]*sheet{}, date1904: date1904}
	read := map[string]bool{}
	for _, s := range refs {
		target, ok := targets[s.rID]
		if !ok {
			return nil, fmt.Errorf("%w: sheet %q has no part", ErrNotWorkbook, s.name)
		}
		part := partName(target)
		if _, dup := out.sheets[s.name]; dup || read[part] {
			// Each sheet is read once: a part listed again would be read
			// again, as often as the workbook lists it.
			return nil, fmt.Errorf("%w: sheet %q or its part is listed twice", ErrNotWorkbook, s.name)
		}
		read[part] = true
		sh, err := z.sheet(s.name, part, strs)
		if err != nil {
			return nil, err
		}
		out.sheets[s.name] = sh
	}
	return out, nil
}

// workbookPart returns the sheets xl/workbook.xml lists and whether its
// dates count from 1904.
func (z *zipParts) workbookPart() ([]sheetRef, bool, error) {
	var (
		refs     []sheetRef
		date1904 bool
	)
	err := z.walk("xl/workbook.xml", true, func(w *walker, tok xml.Token) error {
		t, ok := tok.(xml.StartElement)
		switch {
		case !ok:
		case w.in("workbook", "workbookPr"):
			v := attr(t, "", "date1904")
			date1904 = v == "1" || v == "true"
		case w.in("sheets", "sheet"):
			if len(refs) == z.lim.sheets {
				return fmt.Errorf("%w: more than %d sheets", ErrTooLarge, z.lim.sheets)
			}
			refs = append(refs, sheetRef{name: attr(t, "", "name"), rID: attr(t, nsRelationships, "id")})
		}
		return nil
	})
	return refs, date1904, err
}

// targets returns the workbook's relationship targets by ID.
func (z *zipParts) targets() (map[string]string, error) {
	targets := map[string]string{}
	err := z.walk("xl/_rels/workbook.xml.rels", true, func(w *walker, tok xml.Token) error {
		t, ok := tok.(xml.StartElement)
		if !ok || !w.in("Relationships", "Relationship") {
			return nil
		}
		if len(targets) == maxParts {
			return fmt.Errorf("%w: more than %d relationships", ErrTooLarge, maxParts)
		}
		targets[attr(t, "", "Id")] = attr(t, "", "Target")
		return nil
	})
	return targets, err
}

// sharedStrings returns the workbook's shared strings, none without the
// part. An item's text is its plain text or its rich text runs joined;
// phonetic runs are left out.
func (z *zipParts) sharedStrings() ([]string, error) {
	r := &sharedReader{z: z}
	err := z.walk("xl/sharedStrings.xml", false, r.token)
	return r.strs, err
}

// sharedReader collects the shared strings from their part's tokens.
type sharedReader struct {
	z    *zipParts
	strs []string
	item strings.Builder
}

// token takes one token of the shared strings part.
func (r *sharedReader) token(w *walker, tok xml.Token) error {
	switch t := tok.(type) {
	case xml.StartElement:
		if w.in("sst", "si") {
			r.item.Reset()
		}
	case xml.CharData:
		if w.in("si", "t") || w.in("si", "r", "t") {
			return r.text(t)
		}
	case xml.EndElement:
		if t.Name.Local == "si" && w.in("sst") {
			return r.endItem()
		}
	}
	return nil
}

// text keeps t as part of the current item.
func (r *sharedReader) text(t xml.CharData) error {
	if err := r.z.keepText(len(t)); err != nil {
		return err
	}
	r.item.Write(t)
	return nil
}

// endItem keeps the item just read.
func (r *sharedReader) endItem() error {
	if len(r.strs) == r.z.lim.sharedStrings {
		return fmt.Errorf("%w: more than %d shared strings", ErrTooLarge, r.z.lim.sharedStrings)
	}
	r.strs = append(r.strs, r.item.String())
	return nil
}

// partName resolves a relationship target of the workbook part.
func partName(target string) string {
	if strings.HasPrefix(target, "/") {
		return strings.TrimPrefix(path.Clean(target), "/")
	}
	return path.Clean(path.Join("xl", target))
}

// xCell is one cell of a worksheet part as read: its reference, type,
// value and inline string.
type xCell struct {
	R, T, V, Is string
}

// sheetReader builds a sheet from its part's tokens.
type sheetReader struct {
	z    *zipParts
	s    *sheet
	strs []string
	// next is the number of the row after the last one read.
	next int
	// count is how many rows were read.
	count int
	rowN  int
	cells row
	col   int
	c     xCell
	v, is strings.Builder
}

// sheet reads the worksheet part name with the shared strings strs.
func (z *zipParts) sheet(name, part string, strs []string) (*sheet, error) {
	r := &sheetReader{z: z, s: &sheet{name: name, rows: map[int]row{}}, strs: strs, next: 1}
	if err := z.walk(part, true, r.token); err != nil {
		return nil, err
	}
	return r.s, nil
}

// token takes one token of the worksheet part.
func (r *sheetReader) token(w *walker, tok xml.Token) error {
	switch t := tok.(type) {
	case xml.StartElement:
		return r.start(w, t)
	case xml.CharData:
		return r.charData(w, t)
	case xml.EndElement:
		return r.end(w, t)
	}
	return nil
}

// start takes a start element: a row or a cell.
func (r *sheetReader) start(w *walker, t xml.StartElement) error {
	switch {
	case w.in("sheetData", "row"):
		return r.startRow(t)
	case w.in("row", "c"):
		r.c = xCell{R: attr(t, "", "r"), T: attr(t, "", "t")}
		r.v.Reset()
		r.is.Reset()
	}
	return nil
}

// charData takes text: a cell's value or inline string.
func (r *sheetReader) charData(w *walker, t xml.CharData) error {
	switch {
	case w.in("c", "v"):
		return r.text(&r.v, t)
	case w.in("c", "is", "t"), w.in("c", "is", "r", "t"):
		return r.text(&r.is, t)
	}
	return nil
}

// end takes an end element: of a cell or a row.
func (r *sheetReader) end(w *walker, t xml.EndElement) error {
	switch {
	case t.Name.Local == "c" && w.in("row"):
		return r.endCell()
	case t.Name.Local == "row" && w.in("sheetData"):
		r.endRow()
	}
	return nil
}

// text keeps t as part of the current cell's value in b.
func (r *sheetReader) text(b *strings.Builder, t xml.CharData) error {
	if err := r.z.keepText(len(t)); err != nil {
		return err
	}
	b.Write(t)
	return nil
}

// startRow starts the row t, numbered by its r attribute or following the
// previous one.
func (r *sheetReader) startRow(t xml.StartElement) error {
	n := r.next
	if v := attr(t, "", "r"); v != "" {
		var err error
		if n, err = strconv.Atoi(v); err != nil {
			return fmt.Errorf("%w: sheet %q: row number", ErrNotWorkbook, r.s.name)
		}
	}
	if n < 1 || n > r.z.lim.rows {
		return fmt.Errorf("%w: sheet %q: row %d out of range", ErrNotWorkbook, r.s.name, n)
	}
	r.count++
	if r.count > r.z.lim.rows {
		return fmt.Errorf("%w: sheet %q has more than %d rows", ErrTooLarge, r.s.name, r.z.lim.rows)
	}
	r.next = n + 1
	r.rowN = n
	r.cells = nil
	r.col = 0
	return nil
}

// endCell places the cell just read: a cell with a reference in its
// column, one without after the previous cell. A cell with a value is
// kept; an empty one only moves the column on.
func (r *sheetReader) endCell() error {
	if r.c.R != "" {
		c, err := columnOf(r.c.R)
		if err != nil {
			return fmt.Errorf("%w: sheet %q row %d: %w", ErrNotWorkbook, r.s.name, r.rowN, err)
		}
		r.col = c
	}
	r.c.V, r.c.Is = r.v.String(), r.is.String()
	v, err := cellValue(&r.c, r.strs)
	if err != nil {
		return fmt.Errorf("%w: sheet %q row %d: %w", ErrNotWorkbook, r.s.name, r.rowN, err)
	}
	if v.kind != cellEmpty {
		if err := r.z.keepCell(); err != nil {
			return err
		}
		r.cells = append(r.cells, entry{col: r.col, cell: v})
	}
	r.col++
	return nil
}

// endRow keeps the row just read, its cells in column order; when a column
// repeats, the last cell read in it wins. A row without a cell with a
// value is not kept: it reads as missing, the same as empty.
func (r *sheetReader) endRow() {
	cells := r.cells
	if len(cells) == 0 {
		return
	}
	slices.SortStableFunc(cells, func(a, b entry) int { return a.col - b.col })
	out := cells[:0]
	for _, e := range cells {
		if len(out) > 0 && out[len(out)-1].col == e.col {
			out[len(out)-1] = e
			continue
		}
		out = append(out, e)
	}
	r.s.rows[r.rowN] = out
	r.s.last = max(r.s.last, r.rowN)
}

// errCell reports a cell the reader can't read.
var errCell = errors.New("unreadable cell")

// cellValue reads a cell by its type.
func cellValue(xc *xCell, strs []string) (cell, error) {
	switch xc.T {
	case "s":
		return sharedString(xc.V, strs)
	case "inlineStr":
		return cell{kind: cellText, value: xc.Is}, nil
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
