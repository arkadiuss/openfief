package treasurybonds

import (
	"fmt"
	"unicode/utf16"
)

// sharedStrings reads the shared string table: the SST record's data, then
// each CONTINUE record's, as segs. A string's characters may run on into
// the next segment, which then starts with a flags byte saying whether the
// rest is one or two bytes a character; anything else simply continues.
func sharedStrings(segs [][]byte) ([]string, error) {
	if len(segs) == 0 {
		return nil, nil
	}
	r := &segments{segs: segs}
	if _, err := r.u32(); err != nil { // the total count of strings in the workbook
		return nil, err
	}
	n, err := r.u32()
	if err != nil {
		return nil, err
	}
	if n > maxStrings {
		return nil, fmt.Errorf("%w: %d shared strings", ErrTooLarge, n)
	}
	out := make([]string, 0, min(n, initialStrings))
	for range n {
		s, err := r.richString()
		if err != nil {
			return nil, fmt.Errorf("%w: shared string %d: %w", ErrNotWorkbook, len(out), err)
		}
		out = append(out, s)
	}
	return out, nil
}

// segments reads across the data of an SST record and its CONTINUE
// records.
type segments struct {
	segs [][]byte
	i    int
	pos  int
}

// advance moves on to the next segment that isn't empty when the current
// one is used up, and reports whether it did.
func (r *segments) advance() (bool, error) {
	moved := false
	for r.pos >= len(r.segs[r.i]) {
		if r.i+1 >= len(r.segs) {
			return false, errTruncated
		}
		r.i, r.pos, moved = r.i+1, 0, true
	}
	return moved, nil
}

// byte1 reads one byte, moving on to the next segment as needed.
func (r *segments) byte1() (byte, error) {
	if _, err := r.advance(); err != nil {
		return 0, err
	}
	b := r.segs[r.i][r.pos]
	r.pos++
	return b, nil
}

// u16 and u32 read little-endian values byte by byte, so they may span
// segments.
func (r *segments) u16() (uint16, error) {
	lo, err1 := r.byte1()
	hi, err2 := r.byte1()
	if err1 != nil || err2 != nil {
		return 0, errTruncated
	}
	return uint16(lo) | uint16(hi)<<8, nil
}

func (r *segments) u32() (uint32, error) {
	lo, err1 := r.u16()
	hi, err2 := r.u16()
	if err1 != nil || err2 != nil {
		return 0, errTruncated
	}
	return uint32(lo) | uint32(hi)<<16, nil
}

// skip passes over n bytes, which may span segments.
func (r *segments) skip(n uint32) error {
	for left := int64(n); left > 0; {
		if _, err := r.advance(); err != nil {
			return err
		}
		k := min(left, int64(len(r.segs[r.i])-r.pos))
		r.pos += int(k)
		left -= k
	}
	return nil
}

// richString reads one XLUnicodeRichExtendedString: its length, flags,
// optional run and extension counts, characters, then the runs and the
// extension, which it skips.
func (r *segments) richString() (string, error) {
	n, err := r.u16()
	if err != nil {
		return "", err
	}
	flags, err := r.byte1()
	if err != nil {
		return "", err
	}
	var runs, ext uint32
	if flags&richText != 0 {
		c, cErr := r.u16()
		if cErr != nil {
			return "", cErr
		}
		runs = uint32(c)
	}
	if flags&extString != 0 {
		if ext, err = r.u32(); err != nil {
			return "", err
		}
	}
	s, err := r.chars(int(n), flags&highByte != 0)
	if err != nil {
		return "", err
	}
	const runBytes = 4
	if err := r.skip(runs * runBytes); err != nil {
		return "", err
	}
	return s, r.skip(ext)
}

// initialChars is how many characters a string makes room for before
// reading them, so a length the data can't hold allocates nothing large.
const initialChars = 256

// chars reads n characters; at a segment boundary, the next segment's first
// byte says whether the rest are wide.
func (r *segments) chars(n int, wide bool) (string, error) {
	units := make([]uint16, 0, min(n, initialChars))
	for len(units) < n {
		moved, err := r.advance()
		if err != nil {
			return "", err
		}
		if moved {
			flags := r.segs[r.i][r.pos]
			r.pos++
			wide = flags&highByte != 0
			continue
		}
		seg := r.segs[r.i]
		if !wide {
			units = append(units, uint16(seg[r.pos]))
			r.pos++
			continue
		}
		if r.pos+1 >= len(seg) {
			return "", errTruncated
		}
		units = append(units, le16(seg, r.pos))
		r.pos += 2
	}
	return string(utf16.Decode(units)), nil
}
