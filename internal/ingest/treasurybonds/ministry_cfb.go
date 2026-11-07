package treasurybonds

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"unicode/utf16"
)

// ErrNotWorkbook reports a file that is not a readable Excel 97-2003 (.xls)
// workbook: no compound file, no Workbook stream, or records that don't
// parse.
var ErrNotWorkbook = errors.New("not an .xls workbook")

// The compound file binary format (MS-CFB), which holds an .xls workbook's
// streams in sectors chained by a file allocation table.
const (
	cfbHeaderSize    = 512
	cfbHeaderDIFAT   = 109
	dirEntrySize     = 128
	miniSectorSize   = 64
	maxRegularSector = 0xFFFFFFFA
	endOfChain       = 0xFFFFFFFE
	freeSector       = 0xFFFFFFFF
	sectorShiftV3    = 9
	sectorShiftV4    = 12
	versionV3        = 3
	versionV4        = 4
	entryStream      = 2
	entryRoot        = 5
	uint32Bytes      = 4
	maxNameBytes     = 64
	byteOrderMark    = 0xFFFE
)

// Offsets of the compound file header's fields and of a directory entry's
// (MS-CFB 2.2 and 2.6).
const (
	hdrVersion     = 26
	hdrByteOrder   = 28
	hdrSectorShift = 30
	hdrDirStart    = 48
	hdrCutoff      = 56
	hdrMiniFAT     = 60
	hdrDIFATStart  = 68
	hdrDIFAT       = 76
	entryType      = 66
	entryStart     = 116
	entrySize      = 120
)

// cfbSignature starts every compound file.
const cfbSignature = "\xD0\xCF\x11\xE0\xA1\xB1\x1A\xE1"

// isWorkbook reports whether data starts like a compound file, the
// container of an .xls workbook.
func isWorkbook(data []byte) bool { return bytes.HasPrefix(data, []byte(cfbSignature)) }

// compoundFile is a parsed compound file: its sectors, file allocation
// table and directory.
type compoundFile struct {
	data       []byte
	sectorSize int
	fat        []uint32
	dir        []dirEntry
	cutoff     uint64
	miniFAT    []uint32
	miniStream []byte
}

// dirEntry is one directory entry: a stream or storage.
type dirEntry struct {
	name  string
	typ   byte
	start uint32
	size  uint64
}

// readStream returns the stream named name, such as "Workbook", from the
// compound file data. Every chain is bounded by the file's own size, so a
// loop or an out-of-range sector is an error, never a hang.
func readStream(data []byte, name string) ([]byte, error) {
	f, err := parseCompoundFile(data)
	if err != nil {
		return nil, err
	}
	for _, e := range f.dir {
		if e.typ != entryStream || e.name != name {
			continue
		}
		if e.size < f.cutoff {
			return f.miniChain(e.start, e.size)
		}
		return f.chain(e.start, e.size)
	}
	return nil, fmt.Errorf("%w: no %s stream", ErrNotWorkbook, name)
}

// parseCompoundFile reads the header, the file allocation table, the
// directory and the mini stream.
func parseCompoundFile(data []byte) (*compoundFile, error) {
	if len(data) < cfbHeaderSize || !isWorkbook(data) {
		return nil, fmt.Errorf("%w: no compound file header", ErrNotWorkbook)
	}
	f := &compoundFile{data: data}
	if err := f.readHeader(); err != nil {
		return nil, err
	}
	dir, err := f.chain(le32(data, hdrDirStart), uint64(len(data)))
	if err != nil {
		return nil, fmt.Errorf("directory: %w", err)
	}
	f.dir = dirEntries(dir)
	if len(f.dir) == 0 || f.dir[0].typ != entryRoot {
		return nil, fmt.Errorf("%w: no root entry", ErrNotWorkbook)
	}
	if err := f.readMini(); err != nil {
		return nil, err
	}
	return f, nil
}

// readHeader checks the header and reads the file allocation table.
func (f *compoundFile) readHeader() error {
	data := f.data
	version, shift := le16(data, hdrVersion), le16(data, hdrSectorShift)
	if le16(data, hdrByteOrder) != byteOrderMark ||
		(version != versionV3 || shift != sectorShiftV3) && (version != versionV4 || shift != sectorShiftV4) {
		return fmt.Errorf("%w: unsupported compound file version %d, sector shift %d", ErrNotWorkbook, version, shift)
	}
	f.sectorSize = 1 << shift
	f.cutoff = uint64(le32(data, hdrCutoff))
	fatSectors, err := f.fatSectors()
	if err != nil {
		return err
	}
	for _, s := range fatSectors {
		sec, err := f.sector(s)
		if err != nil {
			return fmt.Errorf("file allocation table: %w", err)
		}
		f.fat = append(f.fat, uint32s(sec)...)
	}
	return nil
}

// fatSectors returns the sectors of the file allocation table: those the
// header lists, then those of the DIFAT chain. A table of more sectors than
// the file has fails, so the table is never larger than the file.
func (f *compoundFile) fatSectors() ([]uint32, error) {
	var out []uint32
	for i := range cfbHeaderDIFAT {
		if s := le32(f.data, hdrDIFAT+i*uint32Bytes); s <= maxRegularSector {
			out = append(out, s)
		}
	}
	perSector := f.sectorSize/uint32Bytes - 1
	next, seen := le32(f.data, hdrDIFATStart), 0
	for next <= maxRegularSector {
		seen++
		if seen > f.sectors() {
			return nil, fmt.Errorf("%w: DIFAT chain loops", ErrNotWorkbook)
		}
		sec, err := f.sector(next)
		if err != nil {
			return nil, fmt.Errorf("DIFAT: %w", err)
		}
		entries := uint32s(sec)
		for _, s := range entries[:perSector] {
			if s <= maxRegularSector {
				out = append(out, s)
			}
		}
		next = entries[perSector]
	}
	if len(out) > f.sectors() {
		return nil, fmt.Errorf("%w: %d FAT sectors in %d sectors", ErrNotWorkbook, len(out), f.sectors())
	}
	return out, nil
}

// readMini reads the mini stream, held in the root entry's chain, and the
// mini file allocation table, which streams smaller than the cutoff use.
func (f *compoundFile) readMini() error {
	root := f.dir[0]
	if root.size == 0 {
		return nil
	}
	ms, err := f.chain(root.start, root.size)
	if err != nil {
		return fmt.Errorf("mini stream: %w", err)
	}
	mf, err := f.chain(le32(f.data, hdrMiniFAT), uint64(len(f.data)))
	if err != nil {
		return fmt.Errorf("mini file allocation table: %w", err)
	}
	f.miniStream, f.miniFAT = ms, uint32s(mf)
	return nil
}

// sectors returns how many whole sectors follow the header.
func (f *compoundFile) sectors() int { return (len(f.data) - f.sectorSize) / f.sectorSize }

// sector returns the bytes of sector s.
func (f *compoundFile) sector(s uint32) ([]byte, error) {
	if int64(s) >= int64(f.sectors()) {
		return nil, fmt.Errorf("%w: sector %d beyond the file", ErrNotWorkbook, s)
	}
	off := (int(s) + 1) * f.sectorSize
	return f.data[off : off+f.sectorSize], nil
}

// chain returns the sectors chained from start, at most size bytes of
// them. A chain shorter than size is an error unless size is the file's
// length, which reads the whole chain.
func (f *compoundFile) chain(start uint32, size uint64) ([]byte, error) {
	var out []byte
	for s, steps := start, 0; s != endOfChain && uint64(len(out)) < size; s = f.fat[s] {
		steps++
		if steps > f.sectors() || s > maxRegularSector || int64(s) >= int64(len(f.fat)) {
			return nil, fmt.Errorf("%w: broken sector chain at %d", ErrNotWorkbook, s)
		}
		sec, err := f.sector(s)
		if err != nil {
			return nil, err
		}
		out = append(out, sec...)
	}
	return trim(out, size, uint64(len(f.data)))
}

// miniChain returns size bytes of the mini sectors chained from start.
func (f *compoundFile) miniChain(start uint32, size uint64) ([]byte, error) {
	var out []byte
	limit := len(f.miniStream) / miniSectorSize
	for s, steps := start, 0; s != endOfChain && uint64(len(out)) < size; s = f.miniFAT[s] {
		steps++
		if steps > limit || int64(s) >= int64(limit) || int64(s) >= int64(len(f.miniFAT)) {
			return nil, fmt.Errorf("%w: broken mini sector chain at %d", ErrNotWorkbook, s)
		}
		off := int(s) * miniSectorSize
		out = append(out, f.miniStream[off:off+miniSectorSize]...)
	}
	return trim(out, size, uint64(len(f.data)))
}

// trim cuts b to size, or fails when b is shorter; whole is the size that
// means "the whole chain".
func trim(b []byte, size, whole uint64) ([]byte, error) {
	if size == whole {
		return b, nil
	}
	if uint64(len(b)) < size {
		return nil, fmt.Errorf("%w: stream shorter than its size", ErrNotWorkbook)
	}
	return b[:size], nil
}

// dirEntries reads the directory's entries; an entry's size uses only its
// low 32 bits, as version 3 files may leave the high ones unset.
func dirEntries(dir []byte) []dirEntry {
	out := make([]dirEntry, 0, len(dir)/dirEntrySize)
	for off := 0; off+dirEntrySize <= len(dir); off += dirEntrySize {
		e := dir[off : off+dirEntrySize]
		n := min(int(le16(e, maxNameBytes)), maxNameBytes)
		units := make([]uint16, 0, n/wideBytes)
		for i := 0; i+1 < n; i += 2 {
			if u := le16(e, i); u != 0 {
				units = append(units, u)
			}
		}
		out = append(out, dirEntry{name: string(utf16.Decode(units)), typ: e[entryType], start: le32(e, entryStart), size: uint64(le32(e, entrySize))})
	}
	return out
}

// uint32s reads b as little-endian 32-bit values.
func uint32s(b []byte) []uint32 {
	out := make([]uint32, 0, len(b)/uint32Bytes)
	for i := 0; i+uint32Bytes <= len(b); i += uint32Bytes {
		out = append(out, le32(b, i))
	}
	return out
}

// le16 and le32 read little-endian values at off.
func le16(b []byte, off int) uint16 { return binary.LittleEndian.Uint16(b[off:]) }
func le32(b []byte, off int) uint32 { return binary.LittleEndian.Uint32(b[off:]) }
