package treasurybonds

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"unicode/utf16"

	"github.com/arkadiuss/openfief/internal/kernel"
)

// BIFF8, the record format of an Excel 97-2003 workbook stream (MS-XLS):
// each record is a type, a length and its data. The reader keeps only the
// sheets' text and number cells.
const (
	recBOF        = 0x0809
	recEOF        = 0x000A
	recBoundSheet = 0x0085
	recSST        = 0x00FC
	recContinue   = 0x003C
	recLabelSST   = 0x00FD
	recLabel      = 0x0204
	recNumber     = 0x0203
	recRK         = 0x027E
	recMulRK      = 0x00BD
	recFormula    = 0x0006
	recString     = 0x0207

	biff8Version    = 0x0600
	bofGlobals      = 0x0005
	bofWorksheet    = 0x0010
	recordHeader    = 4
	cellHeader      = 6
	rkSize          = 6
	formulaResult   = 8
	maxSheetRows    = 65536
	maxSheetColumns = 256
	// maxCells caps the cells kept, all sheets together: the ministry's
	// file has about 25,000.
	maxCells = 200_000
	// maxSheets caps the worksheets read: the ministry's file has 18.
	maxSheets = 64
	// maxStrings caps the shared strings, and initialStrings is how many
	// the reader makes room for before reading them.
	maxStrings     = 200_000
	initialStrings = 1024
	// initialSegments is how many SST segments the reader makes room for:
	// the SST record and its CONTINUE records, about 20 in the ministry's
	// file.
	initialSegments = 32
	rkFixed         = 2
	rkTimes100      = 1
	rkShift         = 2
	rkSign          = 1 << 31
	rkRange         = 1 << 30
	rkFlags         = rkFixed | rkTimes100
	// rkHighBits shifts an RK double's 30 bits to the top of the 64.
	rkHighBits = 32
	// lengthAt is where a record header holds the data's length;
	// bofKindAt where a BOF record holds the substream's kind; resultTypeAt
	// where a FORMULA's cached result marks itself as no number.
	lengthAt     = 2
	bofKindAt    = 2
	resultTypeAt = 6
	// shortHeader and longHeader are the lengths of a string's header: a
	// length of one or two bytes, then a flags byte.
	shortHeader = 2
	longHeader  = 3
	// mulRKFixed is a MULRK record's size without its values: the row,
	// the first column and the last column.
	mulRKFixed = 6
	// mulRKFirst is where its first value starts.
	mulRKFirst = 4
	// rkValueAt is where an RK value starts after its format index.
	rkValueAt = 2
	// colAt is where a cell record holds its column, after its row.
	colAt = 2
	// wideBytes is how many bytes a wide character takes.
	wideBytes = 2
	highByte  = 1
	richText  = 0x08
	extString = 0x04
	// formulaString marks a formula whose cached result is text, held in the
	// next STRING record.
	formulaString = 0
	resultMarker  = 0xFFFF
)

// ErrTooLarge reports a download or workbook larger than the reader
// accepts: more bytes, cells or shared strings than the ministry's file
// could plausibly hold.
var ErrTooLarge = errors.New("too large")

// cell is one text or number cell. A number is kept as the exact decimal
// of the shortest text that reads back as the file's IEEE double, never as
// a float: rates are money (code-organization.md).
type cell struct {
	text   string
	num    kernel.Decimal
	number bool
}

// sheet is a worksheet's cells by row, then column.
type sheet map[int]map[int]cell

// workbook is the worksheets of an .xls workbook by name.
type workbook map[string]sheet

// readWorkbook reads the worksheets of the BIFF8 Workbook stream s.
func readWorkbook(s []byte) (workbook, error) {
	r := &records{b: s}
	sheets, sst, err := r.globals()
	if err != nil {
		return nil, err
	}
	wb := workbook{}
	cells := 0
	for _, bs := range sheets {
		sh, err := (&records{b: s, pos: bs.offset}).worksheet(sst, &cells)
		if err != nil {
			return nil, fmt.Errorf("sheet %q: %w", bs.name, err)
		}
		wb[bs.name] = sh
	}
	return wb, nil
}

// boundSheet is a worksheet the globals list: its name and where its
// records start in the stream.
type boundSheet struct {
	name   string
	offset int
}

// record is one record of the stream.
type record struct {
	typ  uint16
	data []byte
}

// records reads records from b, from pos on.
type records struct {
	b   []byte
	pos int
}

// errTruncated reports a record or field cut short.
var errTruncated = errors.New("truncated record")

// next returns the next record.
func (r *records) next() (record, error) {
	if r.pos+recordHeader > len(r.b) {
		return record{}, fmt.Errorf("%w: %w at %d", ErrNotWorkbook, errTruncated, r.pos)
	}
	typ, n := le16(r.b, r.pos), int(le16(r.b, r.pos+lengthAt))
	start := r.pos + recordHeader
	if start+n > len(r.b) {
		return record{}, fmt.Errorf("%w: %w at %d", ErrNotWorkbook, errTruncated, r.pos)
	}
	r.pos = start + n
	return record{typ: typ, data: r.b[start : start+n]}, nil
}

// bof reads a BIFF8 beginning-of-substream record of kind want.
func (r *records) bof(want uint16) error {
	rec, err := r.next()
	if err != nil {
		return err
	}
	if rec.typ != recBOF || len(rec.data) < bofKindAt+2 || le16(rec.data, 0) != biff8Version || le16(rec.data, bofKindAt) != want {
		return fmt.Errorf("%w: not a BIFF8 substream of kind %#x", ErrNotWorkbook, want)
	}
	return nil
}

// globals reads the workbook globals: the worksheets and the shared
// strings.
func (r *records) globals() ([]boundSheet, []string, error) {
	if err := r.bof(bofGlobals); err != nil {
		return nil, nil, err
	}
	var g globalRecords
	for {
		rec, err := r.next()
		if err != nil {
			return nil, nil, err
		}
		if rec.typ == recEOF {
			shared, err := sharedStrings(g.sst)
			return g.sheets, shared, err
		}
		if err := g.add(rec, len(r.b)); err != nil {
			return nil, nil, err
		}
	}
}

// globalRecords are the globals read so far: the worksheets, the SST
// record's data and its CONTINUE records', and whether the last record was
// one of those.
type globalRecords struct {
	sheets []boundSheet
	sst    [][]byte
	inSST  bool
}

// add reads one globals record of a stream of size bytes.
func (g *globalRecords) add(rec record, size int) error {
	continued := rec.typ == recContinue && g.inSST
	g.inSST = rec.typ == recSST || continued
	switch {
	case rec.typ == recBoundSheet:
		bs, ok, err := boundSheetOf(rec.data, size)
		if err != nil || !ok {
			return err
		}
		if g.sheets = append(g.sheets, bs); len(g.sheets) > maxSheets {
			return fmt.Errorf("%w: more than %d sheets", ErrTooLarge, maxSheets)
		}
	case rec.typ == recSST:
		g.sst = append(make([][]byte, 0, initialSegments), rec.data)
	case continued:
		g.sst = append(g.sst, rec.data)
	}
	return nil
}

// boundSheetOf reads a BOUNDSHEET record of a stream of size bytes, and
// reports whether it names a worksheet.
func boundSheetOf(d []byte, size int) (boundSheet, bool, error) {
	const minLen, typeAt, nameAt = 8, 5, 6
	if len(d) < minLen {
		return boundSheet{}, false, fmt.Errorf("%w: %w: sheet entry", ErrNotWorkbook, errTruncated)
	}
	off := int64(le32(d, 0))
	if off >= int64(size) {
		return boundSheet{}, false, fmt.Errorf("%w: sheet beyond the stream", ErrNotWorkbook)
	}
	name, err := shortString(d[nameAt:])
	if err != nil {
		return boundSheet{}, false, err
	}
	return boundSheet{name: name, offset: int(off)}, d[typeAt] == 0, nil
}

// shortString reads a string with a one-byte length and a flags byte.
func shortString(d []byte) (string, error) {
	if len(d) < shortHeader {
		return "", fmt.Errorf("%w: %w: string", ErrNotWorkbook, errTruncated)
	}
	return chars(d[shortHeader:], int(d[0]), d[1]&highByte != 0)
}

// longString reads a string with a two-byte length and a flags byte, as a
// LABEL or STRING record holds it.
func longString(d []byte) (string, error) {
	if len(d) < longHeader {
		return "", fmt.Errorf("%w: %w: string", ErrNotWorkbook, errTruncated)
	}
	return chars(d[longHeader:], int(le16(d, 0)), d[2]&highByte != 0)
}

// chars decodes n characters from d: two bytes each when wide, else one
// (Latin-1).
func chars(d []byte, n int, wide bool) (string, error) {
	size := n
	if wide {
		size *= wideBytes
	}
	if len(d) < size {
		return "", fmt.Errorf("%w: %w: characters", ErrNotWorkbook, errTruncated)
	}
	if !wide {
		runes := make([]rune, n)
		for i, b := range d[:n] {
			runes[i] = rune(b)
		}
		return string(runes), nil
	}
	units := make([]uint16, n)
	for i := range units {
		units[i] = le16(d, wideBytes*i)
	}
	return string(utf16.Decode(units)), nil
}

// worksheet reads one worksheet's cells, counting them in cells.
func (r *records) worksheet(sst []string, cells *int) (sheet, error) {
	if err := r.bof(bofWorksheet); err != nil {
		return nil, err
	}
	sh := sheet{}
	pending := -1 // the formula cell waiting for its STRING record, as row*columns+column
	for {
		rec, err := r.next()
		if err != nil {
			return nil, err
		}
		if rec.typ == recEOF {
			return sh, nil
		}
		got, next, err := cellsOf(rec, sst, pending)
		if err != nil {
			return nil, err
		}
		pending = next
		for _, c := range got {
			*cells++
			if *cells > maxCells {
				return nil, fmt.Errorf("%w: more than %d cells", ErrTooLarge, maxCells)
			}
			sh.put(c)
		}
	}
}

// placed is a cell with its row and column.
type placed struct {
	cell

	row, col int
}

// put stores c.
func (s sheet) put(c placed) {
	row, ok := s[c.row]
	if !ok {
		row = map[int]cell{}
		s[c.row] = row
	}
	row[c.col] = c.cell
}

// cellsOf returns the cells record rec holds, and the cell a formula's
// STRING record still has to fill (pending, -1 for none).
func cellsOf(rec record, sst []string, pending int) ([]placed, int, error) {
	d := rec.data
	switch rec.typ {
	case recLabelSST, recLabel, recNumber, recRK, recFormula, recMulRK:
		if len(d) < cellHeader {
			return nil, -1, fmt.Errorf("%w: %w: cell %#x", ErrNotWorkbook, errTruncated, rec.typ)
		}
	case recString:
		if pending < 0 {
			return nil, -1, nil
		}
		text, err := longString(d)
		return []placed{{row: pending / maxSheetColumns, col: pending % maxSheetColumns, cell: cell{text: text}}}, -1, err
	default:
		return nil, -1, nil
	}
	row, col := int(le16(d, 0)), int(le16(d, colAt))
	if col >= maxSheetColumns {
		return nil, -1, fmt.Errorf("%w: column %d", ErrNotWorkbook, col)
	}
	return oneCell(rec.typ, d, row, col, sst)
}

// oneCell reads a cell record of type typ at row and col.
func oneCell(typ uint16, d []byte, row, col int, sst []string) ([]placed, int, error) {
	at := func(c cell) []placed { return []placed{{row: row, col: col, cell: c}} }
	switch typ {
	case recLabelSST:
		const idxEnd = 10
		if len(d) < idxEnd || int64(le32(d, cellHeader)) >= int64(len(sst)) {
			return nil, -1, fmt.Errorf("%w: shared string out of range", ErrNotWorkbook)
		}
		return at(cell{text: sst[le32(d, cellHeader)]}), -1, nil
	case recLabel:
		text, err := longString(d[cellHeader:])
		return at(cell{text: text}), -1, err
	case recNumber:
		v, err := number(d[cellHeader:])
		return at(cell{num: v, number: true}), -1, err
	case recRK:
		if len(d) < cellHeader+uint32Bytes {
			return nil, -1, fmt.Errorf("%w: %w: RK", ErrNotWorkbook, errTruncated)
		}
		v, err := rk(le32(d, cellHeader))
		return at(cell{num: v, number: true}), -1, err
	case recFormula:
		return formula(d, row, col)
	}
	return mulRK(d, row, col)
}

// number reads an 8-byte IEEE double as a decimal.
func number(d []byte) (kernel.Decimal, error) {
	const size = 8
	if len(d) < size {
		return kernel.Decimal{}, fmt.Errorf("%w: %w: number", ErrNotWorkbook, errTruncated)
	}
	return decimalOf(uint64(le32(d, 0)) | uint64(le32(d, uint32Bytes))<<32)
}

// decimalOf is the decimal of the IEEE double with bits: the shortest
// decimal that reads back as the same double, so 0.0535 stored as the
// double nearest it reads as 0.0535. NaN and infinities are refused.
func decimalOf(bits uint64) (kernel.Decimal, error) {
	f := math.Float64frombits(bits)
	if math.IsNaN(f) || math.IsInf(f, 0) {
		return kernel.Decimal{}, fmt.Errorf("%w: a number that is not finite", ErrNotWorkbook)
	}
	d, err := kernel.ParseDecimal(strconv.FormatFloat(f, 'g', -1, 64))
	if err != nil {
		return kernel.Decimal{}, fmt.Errorf("%w: number: %w", ErrNotWorkbook, err)
	}
	return d, nil
}

// rk decodes an RK number: a 30-bit integer, exact, or the high 30 bits of
// a double, either one divided by 100 when flagged.
func rk(v uint32) (kernel.Decimal, error) {
	var (
		d   kernel.Decimal
		err error
	)
	if v&rkFixed != 0 {
		i := int64(v >> rkShift)
		if v&rkSign != 0 { // a negative 30-bit integer, in two's complement
			i -= rkRange
		}
		d = kernel.NewDecimal(i, 0)
	} else if d, err = decimalOf(uint64(v&^rkFlags) << rkHighBits); err != nil {
		return kernel.Decimal{}, err
	}
	if v&rkTimes100 != 0 {
		d = d.Mul(kernel.NewDecimal(1, -2))
	}
	return d, nil
}

// formula reads a FORMULA record's cached result: a number, or text that
// the next STRING record holds (returned as pending).
func formula(d []byte, row, col int) ([]placed, int, error) {
	if len(d) < cellHeader+formulaResult {
		return nil, -1, fmt.Errorf("%w: %w: formula", ErrNotWorkbook, errTruncated)
	}
	res := d[cellHeader : cellHeader+formulaResult]
	if le16(res, resultTypeAt) == resultMarker {
		if res[0] == formulaString && row < maxSheetRows {
			return nil, row*maxSheetColumns + col, nil
		}
		return nil, -1, nil
	}
	v, err := number(res)
	return []placed{{row: row, col: col, cell: cell{num: v, number: true}}}, -1, err
}

// mulRK reads a MULRK record: RK numbers in consecutive columns of one row.
func mulRK(d []byte, row, first int) ([]placed, int, error) {
	n := (len(d) - mulRKFixed) / rkSize
	if n < 0 || mulRKFixed+n*rkSize != len(d) || first+n > maxSheetColumns {
		return nil, -1, fmt.Errorf("%w: malformed MULRK", ErrNotWorkbook)
	}
	out := make([]placed, 0, n)
	for i := range n {
		v, err := rk(le32(d, mulRKFirst+i*rkSize+rkValueAt))
		if err != nil {
			return nil, -1, err
		}
		out = append(out, placed{row: row, col: first + i, cell: cell{num: v, number: true}})
	}
	return out, -1, nil
}
