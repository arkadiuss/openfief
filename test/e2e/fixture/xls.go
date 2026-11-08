package main

import (
	"encoding/binary"
	"math"
	"slices"
	"strconv"
	"unicode/utf16"
)

// The synthetic workbook the fixture serves as the Ministry of Finance's
// retail bond spreadsheet is written here, record by record, as Excel
// 97-2003 writes one (MS-XLS, BIFF8, in an MS-CFB compound file), so no real
// spreadsheet is committed. It follows the writer the treasury-bonds
// reader's own tests use (internal/ingest/treasurybonds/ministry_xls_test.go),
// with only the two cell kinds the fixture needs: text in LABEL records and
// numbers in NUMBER records.

// BIFF8 record types and fields the writer uses.
const (
	biffBOF        = 0x0809
	biffEOF        = 0x000A
	biffBoundSheet = 0x0085
	biffLabel      = 0x0204
	biffNumber     = 0x0203
	biffVersion    = 0x0600
	bofGlobals     = 0x0005
	bofWorksheet   = 0x0010
	// bofTail is the BOF record's bytes after its version and kind.
	bofTail = 12
	// wideFlag marks a string's characters as two bytes each.
	wideFlag = 1
)

// Compound file layout: 512-byte sectors, the 4096-byte cutoff below which
// a stream would go in the mini stream, and the special sector numbers.
const (
	cfbSector        = 512
	cfbCutoff        = 4096
	cfbDirEntry      = 128
	cfbNameBytes     = 64
	cfbEndOfChain    = 0xFFFFFFFE
	cfbFATSector     = 0xFFFFFFFD
	cfbNoStream      = 0xFFFFFFFF
	cfbHeaderVersion = 3
	cfbMinorVersion  = 0x3E
	cfbByteOrder     = 0xFFFE
	cfbSectorShift   = 9
	cfbMiniShift     = 6
	cfbStorage       = 5 // the root entry
	cfbStream        = 2
	cfbBlack         = 1
	cfbSignature     = "\xD0\xCF\x11\xE0\xA1\xB1\x1A\xE1"
	// cfbFirstStream is the stream's first sector: after the FAT (0) and
	// the directory (1).
	cfbFirstStream = 2
	// u32Bytes is the size of a 32-bit value.
	u32Bytes = 4
	// cfbCLSID and cfbReserved are the header's zeroed class ID and
	// reserved bytes; cfbFree fills what a sector doesn't use.
	cfbCLSID    = 16
	cfbReserved = 6
	cfbFree     = 0xFF
)

// xlsCell is one cell of a synthetic sheet: its column and its text, or,
// when number is set, its number as decimal text.
type xlsCell struct {
	col    int
	text   string
	number bool
}

// xlsSheet is a synthetic worksheet: its name and its rows' cells.
type xlsSheet struct {
	name string
	rows [][]xlsCell
}

// textCell and numberCell are cells of each kind.
func textCell(col int, s string) xlsCell   { return xlsCell{col: col, text: s} }
func numberCell(col int, s string) xlsCell { return xlsCell{col: col, text: s, number: true} }

// xlsFile is the .xls file of sheets. A number that doesn't parse is
// written as text, which the reader then doesn't take as a rate.
func xlsFile(sheets ...xlsSheet) []byte {
	return compoundFile(workbookStream(sheets))
}

// biffRecord is one BIFF record: its type and data.
func biffRecord(typ uint16, data []byte) []byte {
	out := binary.LittleEndian.AppendUint16(nil, typ)
	out = binary.LittleEndian.AppendUint16(out, clamp16(len(data)))
	return append(out, data...)
}

// le16s writes little-endian 16-bit values.
func le16s(vs ...int) []byte {
	var out []byte
	for _, v := range vs {
		out = binary.LittleEndian.AppendUint16(out, clamp16(v))
	}
	return out
}

// le32 writes a little-endian 32-bit value.
func le32(v uint32) []byte { return binary.LittleEndian.AppendUint32(nil, v) }

// utf16le writes s as UTF-16LE.
func utf16le(s string) []byte {
	var out []byte
	for _, u := range utf16.Encode([]rune(s)) {
		out = binary.LittleEndian.AppendUint16(out, u)
	}
	return out
}

// cellRecord writes cell c of row.
func cellRecord(row int, c xlsCell) []byte {
	head := le16s(row, c.col, 0)
	if c.number {
		if f, err := strconv.ParseFloat(c.text, 64); err == nil {
			return biffRecord(biffNumber, binary.LittleEndian.AppendUint64(head, math.Float64bits(f)))
		}
	}
	data := slices.Concat(head, le16s(len(utf16.Encode([]rune(c.text)))), []byte{wideFlag}, utf16le(c.text))
	return biffRecord(biffLabel, data)
}

// bof is a BOF record of kind.
func bof(kind int) []byte {
	return biffRecord(biffBOF, append(le16s(biffVersion, kind), make([]byte, bofTail)...))
}

// boundSheet is the BOUNDSHEET record of a worksheet named name whose BOF
// is at off: a visible worksheet, its name in one-byte characters.
func boundSheet(off int, name string) []byte {
	data := append(le32(clamp32(off)), 0, 0, clampByte(len(name)), 0)
	return biffRecord(biffBoundSheet, append(data, name...))
}

// workbookStream writes the BIFF8 Workbook stream of sheets: the globals,
// which list the sheets, then each sheet's records.
func workbookStream(sheets []xlsSheet) []byte {
	bodies := make([][]byte, len(sheets))
	for i, sh := range sheets {
		body := bof(bofWorksheet)
		for r, cells := range sh.rows {
			for _, c := range cells {
				body = append(body, cellRecord(r, c)...)
			}
		}
		bodies[i] = append(body, biffRecord(biffEOF, nil)...)
	}
	size := len(bof(bofGlobals)) + len(biffRecord(biffEOF, nil))
	for _, sh := range sheets {
		size += len(boundSheet(0, sh.name))
	}
	out := bof(bofGlobals)
	off := size
	for i, sh := range sheets {
		out = append(out, boundSheet(off, sh.name)...)
		off += len(bodies[i])
	}
	out = append(out, biffRecord(biffEOF, nil)...)
	for _, b := range bodies {
		out = append(out, b...)
	}
	return out
}

// compoundFile wraps stream as the Workbook stream of a version 3 compound
// file: the header, one FAT sector, one directory sector, then the stream,
// padded to at least the cutoff so it needs no mini stream. The fixture's
// workbook is a few kilobytes, well within the one FAT sector's 126 data
// sectors.
func compoundFile(stream []byte) []byte {
	if len(stream) < cfbCutoff {
		stream = append(stream, make([]byte, cfbCutoff-len(stream))...)
	}
	n := (len(stream) + cfbSector - 1) / cfbSector
	header := append([]byte(cfbSignature), make([]byte, cfbCLSID)...)
	header = append(header, le16s(cfbMinorVersion, cfbHeaderVersion, cfbByteOrder, cfbSectorShift, cfbMiniShift)...)
	header = append(header, make([]byte, cfbReserved)...)
	header = append(header, le32(0)...)             // directory sectors, 0 in version 3
	header = append(header, le32(1)...)             // FAT sectors
	header = append(header, le32(1)...)             // first directory sector
	header = append(header, le32(0)...)             // transaction signature
	header = append(header, le32(cfbCutoff)...)     // mini stream cutoff
	header = append(header, le32(cfbEndOfChain)...) // no mini FAT
	header = append(header, le32(0)...)
	header = append(header, le32(cfbEndOfChain)...) // no DIFAT sectors
	header = append(header, le32(0)...)
	header = append(header, le32(0)...) // the FAT is sector 0
	header = padded(header, cfbFree)
	fat := append(le32(cfbFATSector), le32(cfbEndOfChain)...)
	for i := range n {
		next := clamp32(i + cfbFirstStream + 1)
		if i == n-1 {
			next = cfbEndOfChain
		}
		fat = append(fat, le32(next)...)
	}
	fat = padded(fat, cfbFree)
	dir := append(dirEntry("Root Entry", cfbStorage, 1, cfbEndOfChain, 0), dirEntry("Workbook", cfbStream, cfbNoStream, cfbFirstStream, len(stream))...)
	dir = padded(dir, 0)
	body := slices.Concat(stream, make([]byte, n*cfbSector-len(stream)))
	return slices.Concat(header, fat, dir, body)
}

// padded fills b with fill up to a whole sector.
func padded(b []byte, fill byte) []byte {
	for len(b)%cfbSector != 0 {
		b = append(b, fill)
	}
	return b
}

// dirEntry is a 128-byte directory entry named name of type typ, with its
// child entry, first sector and size.
func dirEntry(name string, typ byte, child, start uint32, size int) []byte {
	e := append(utf16le(name), 0, 0)
	nameLen := len(e)
	e = append(e, make([]byte, cfbNameBytes-len(e))...)
	e = append(e, le16s(nameLen)...)
	e = append(e, typ, cfbBlack)
	e = append(e, le32(cfbNoStream)...) // left sibling
	e = append(e, le32(cfbNoStream)...) // right sibling
	e = append(e, le32(child)...)
	e = append(e, make([]byte, cfbDirEntry-len(e)-3*u32Bytes)...)
	e = append(e, le32(start)...)
	e = append(e, le32(clamp32(size))...)
	return append(e, make([]byte, u32Bytes)...)
}

// clamp16, clamp32 and clampByte convert v, which the fixture's workbook
// keeps in range; a value out of range becomes the largest, which the
// reader then refuses.
func clamp16(v int) uint16 {
	if v < 0 || v > math.MaxUint16 {
		return math.MaxUint16
	}
	return uint16(v)
}

func clamp32(v int) uint32 {
	if v < 0 || v > math.MaxUint32 {
		return math.MaxUint32
	}
	return uint32(v)
}

func clampByte(v int) byte {
	if v < 0 || v > math.MaxUint8 {
		return math.MaxUint8
	}
	return byte(v)
}
