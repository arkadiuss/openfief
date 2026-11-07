package treasurybonds_test

import (
	"encoding/binary"
	"math"
	"slices"
	"strconv"
	"testing"
	"unicode/utf16"
)

// The synthetic workbooks the tests read are written here, record by
// record, as Excel 97-2003 does (MS-XLS, BIFF8, in an MS-CFB compound
// file), so no real spreadsheet is committed.

// cellKind is how a synthetic cell is written.
type cellKind int

const (
	sstText     cellKind = iota // LABELSST: text in the shared string table
	inlineText                  // LABEL: text in the cell record
	double                      // NUMBER: an IEEE double
	rkHundredth                 // RK: an integer number of hundredths
	rkInteger                   // MULRK: integers, two cells in one record
	formulaNum                  // FORMULA with a number result
	formulaText                 // FORMULA with a text result, then STRING
)

// xcell is one synthetic cell: its column, how it is written, and its
// value: text, or a number as decimal text.
type xcell struct {
	col  int
	kind cellKind
	val  string
}

// xsheet is a synthetic worksheet: its name and its rows' cells.
type xsheet struct {
	name string
	rows [][]xcell
}

// text and num are cells of the common kinds.
func text(col int, s string) xcell { return xcell{col: col, kind: sstText, val: s} }
func num(col int, s string) xcell  { return xcell{col: col, kind: double, val: s} }

// record is one BIFF record: its type and data.
func record(typ uint16, data []byte) []byte {
	out := binary.LittleEndian.AppendUint16(nil, typ)
	out = binary.LittleEndian.AppendUint16(out, to16(len(data)))
	return append(out, data...)
}

// u16s and u32s write little-endian values.
func u16s(vs ...int) []byte {
	var out []byte
	for _, v := range vs {
		out = binary.LittleEndian.AppendUint16(out, to16(v))
	}
	return out
}

func u32(v uint32) []byte { return binary.LittleEndian.AppendUint32(nil, v) }

// wide writes s as UTF-16LE.
func wide(s string) []byte {
	var out []byte
	for _, u := range utf16.Encode([]rune(s)) {
		out = binary.LittleEndian.AppendUint16(out, u)
	}
	return out
}

// latin1 reports whether every character of s fits one byte.
func latin1(s string) bool {
	for _, r := range s {
		if r > 0xFF {
			return false
		}
	}
	return true
}

// ieee is the IEEE double of the decimal text s.
func ieee(tb testing.TB, s string) []byte {
	tb.Helper()
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		tb.Fatalf("ParseFloat(%q): %v", s, err)
	}
	return binary.LittleEndian.AppendUint64(nil, math.Float64bits(f))
}

// rkOf is the RK value of the integer text s, divided by 100 when
// hundredths.
func rkOf(tb testing.TB, s string, hundredths bool) []byte {
	tb.Helper()
	i, err := strconv.Atoi(s)
	if err != nil {
		tb.Fatalf("Atoi(%q): %v", s, err)
	}
	v := to32(i)<<2 | 2
	if hundredths {
		v |= 1
	}
	return u32(v)
}

// sst writes the shared string table of strs as an SST record and as many
// CONTINUE records as records of at most maxData bytes take: a string's
// characters may run on into the next record, which then starts with the
// byte that says whether they are one or two bytes each.
func sst(strs []string, maxData int) []byte {
	segs := [][]byte{append(u32(to32(len(strs))), u32(to32(len(strs)))...)}
	add := func(b, flag []byte) {
		if len(segs[len(segs)-1])+len(b) > maxData {
			segs = append(segs, append([]byte(nil), flag...))
		}
		segs[len(segs)-1] = append(segs[len(segs)-1], b...)
	}
	for _, s := range strs {
		compressed := latin1(s)
		flag := byte(1)
		if compressed {
			flag = 0
		}
		add(append(u16s(len(utf16.Encode([]rune(s)))), flag), nil)
		for _, r := range s {
			if compressed {
				add([]byte{toByte(int(r))}, []byte{flag})
				continue
			}
			add(wide(string(r)), []byte{flag})
		}
	}
	out := record(0x00FC, segs[0])
	for _, seg := range segs[1:] {
		out = append(out, record(0x003C, seg)...)
	}
	return out
}

// cellRecords writes one cell's records at row, adding its text to strs.
func cellRecords(tb testing.TB, row int, c xcell, strs *[]string) []byte {
	tb.Helper()
	head := u16s(row, c.col, 0)
	switch c.kind {
	case sstText:
		*strs = append(*strs, c.val)
		return record(0x00FD, append(head, u32(to32(len(*strs)-1))...))
	case inlineText:
		return record(0x0204, append(append(head, append(u16s(units(c.val)), 1)...), wide(c.val)...))
	case double:
		return record(0x0203, append(head, ieee(tb, c.val)...))
	case rkHundredth:
		return record(0x027E, append(head, rkOf(tb, c.val, true)...))
	case rkInteger:
		data := u16s(row, c.col)
		data = append(append(data, u16s(0)...), rkOf(tb, c.val, false)...)
		data = append(append(data, u16s(0)...), rkOf(tb, c.val, false)...)
		return record(0x00BD, append(data, u16s(c.col+1)...))
	case formulaNum:
		data := append(append(head, ieee(tb, c.val)...), append(u16s(0), append(u32(0), u16s(0)...)...)...)
		return record(0x0006, data)
	case formulaText:
		result := []byte{0, 0, 0, 0, 0, 0, 0xFF, 0xFF}
		data := append(append(head, result...), append(u16s(0), append(u32(0), u16s(0)...)...)...)
		return append(record(0x0006, data), record(0x0207, append(append(u16s(units(c.val)), 1), wide(c.val)...))...)
	}
	tb.Fatalf("cell kind %d", c.kind)
	return nil
}

// workbookStream writes the BIFF8 Workbook stream of sheets, with the
// shared strings in records of at most sstMax bytes.
func workbookStream(tb testing.TB, sheets []xsheet, sstMax int) []byte {
	tb.Helper()
	var strs []string
	bodies := make([][]byte, len(sheets))
	for i, sh := range sheets {
		body := record(0x0809, append(u16s(0x0600, 0x0010), make([]byte, 12)...))
		for r, cells := range sh.rows {
			for _, c := range cells {
				body = append(body, cellRecords(tb, r, c, &strs)...)
			}
		}
		bodies[i] = append(body, record(0x000A, nil)...)
	}
	bof := record(0x0809, append(u16s(0x0600, 0x0005), make([]byte, 12)...))
	bound := func(off int, name string) []byte {
		return record(0x0085, append(append(u32(to32(off)), 0, 0, toByte(len(name)), 0), name...))
	}
	size := len(bof) + len(sst(strs, sstMax)) + len(record(0x000A, nil))
	for _, sh := range sheets {
		size += len(bound(0, sh.name))
	}
	out := bof
	off := size
	for i, sh := range sheets {
		out = append(out, bound(off, sh.name)...)
		off += len(bodies[i])
	}
	out = append(append(out, sst(strs, sstMax)...), record(0x000A, nil)...)
	for _, b := range bodies {
		out = append(out, b...)
	}
	return out
}

// compoundFile wraps stream as the Workbook stream of a version 3 compound
// file: the header, one FAT sector, one directory sector, then the stream,
// padded to at least the 4096-byte cutoff so it needs no mini stream.
func compoundFile(tb testing.TB, stream []byte) []byte {
	tb.Helper()
	const sector, cutoff = 512, 4096
	if len(stream) < cutoff {
		stream = append(stream, make([]byte, cutoff-len(stream))...)
	}
	n := (len(stream) + sector - 1) / sector
	if n+2 > sector/4 {
		tb.Fatalf("stream of %d bytes is too long for one FAT sector", len(stream))
	}
	header := append([]byte("\xD0\xCF\x11\xE0\xA1\xB1\x1A\xE1"), make([]byte, 16)...)
	header = append(header, u16s(0x3E, 3, 0xFFFE, 9, 6)...)
	header = append(header, make([]byte, 6)...)
	header = append(header, u32(0)...)          // directory sectors, 0 in version 3
	header = append(header, u32(1)...)          // FAT sectors
	header = append(header, u32(1)...)          // first directory sector
	header = append(header, u32(0)...)          // transaction signature
	header = append(header, u32(cutoff)...)     // mini stream cutoff
	header = append(header, u32(0xFFFFFFFE)...) // no mini FAT
	header = append(header, u32(0)...)
	header = append(header, u32(0xFFFFFFFE)...) // no DIFAT sectors
	header = append(header, u32(0)...)
	header = append(header, u32(0)...) // the FAT is sector 0
	for len(header) < sector {
		header = append(header, 0xFF)
	}
	fat := append(u32(0xFFFFFFFD), u32(0xFFFFFFFE)...)
	for i := range n {
		next := to32(i + 3)
		if i == n-1 {
			next = 0xFFFFFFFE
		}
		fat = append(fat, u32(next)...)
	}
	for len(fat) < sector {
		fat = append(fat, 0xFF)
	}
	dir := append(dirEntry("Root Entry", 5, 1, 0xFFFFFFFE, 0), dirEntry("Workbook", 2, 0xFFFFFFFF, 2, len(stream))...)
	dir = append(dir, make([]byte, sector-len(dir))...)
	body := slices.Concat(stream, make([]byte, n*sector-len(stream)))
	return append(append(append(header, fat...), dir...), body...)
}

// dirEntry is a 128-byte directory entry.
func dirEntry(name string, typ byte, child, start uint32, size int) []byte {
	e := append(wide(name), 0, 0)
	nameLen := len(e)
	e = append(e, make([]byte, 64-len(e))...)
	e = append(e, u16s(nameLen)...)
	e = append(e, typ, 1)
	e = append(e, u32(0xFFFFFFFF)...)
	e = append(e, u32(0xFFFFFFFF)...)
	e = append(e, u32(child)...)
	e = append(e, make([]byte, 36)...)
	e = append(e, u32(start)...)
	e = append(e, u32(to32(size))...)
	return append(e, make([]byte, 4)...)
}

// xls is the synthetic .xls file of sheets.
func xls(tb testing.TB, sheets ...xsheet) []byte {
	tb.Helper()
	return compoundFile(tb, workbookStream(tb, sheets, 64))
}

// units is how many UTF-16 code units s takes.
func units(s string) int { return len(utf16.Encode([]rune(s))) }

// to16, to32 and toByte convert v, which the synthetic workbooks keep in
// range; a value out of range becomes the largest, which the reader then
// refuses, failing the test.
func to16(v int) uint16 {
	if v < 0 || v > math.MaxUint16 {
		return math.MaxUint16
	}
	return uint16(v)
}

func to32(v int) uint32 {
	if v < 0 || v > math.MaxUint32 {
		return math.MaxUint32
	}
	return uint32(v)
}

func toByte(v int) byte {
	if v < 0 || v > math.MaxUint8 {
		return math.MaxUint8
	}
	return byte(v)
}
