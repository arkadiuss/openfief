package treasurybonds

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"

	"github.com/richardlehane/mscfb"
)

// ErrNotWorkbook reports a file that is not a readable Excel 97-2003 (.xls)
// workbook: no compound file, no Workbook stream, or records that don't
// parse.
var ErrNotWorkbook = errors.New("not an .xls workbook")

// uint32Bytes is the size of a little-endian 32-bit value.
const uint32Bytes = 4

// cfbSignature starts every compound file.
const cfbSignature = "\xD0\xCF\x11\xE0\xA1\xB1\x1A\xE1"

// isWorkbook reports whether data starts like a compound file, the
// container of an .xls workbook.
func isWorkbook(data []byte) bool { return bytes.HasPrefix(data, []byte(cfbSignature)) }

// readStream returns the top-level stream named name, such as "Workbook",
// from the compound file data. The compound file binary format (MS-CFB),
// which holds an .xls workbook's streams in sectors chained by a file
// allocation table, is read by github.com/richardlehane/mscfb, which
// detects chain loops and keeps its tables within the file. data is
// already size-limited by the client, and a stream that claims to be
// larger than data itself fails before anything is allocated for it, so a
// forged size or chain is an error, never a hang or a large allocation.
func readStream(data []byte, name string) ([]byte, error) {
	if !isWorkbook(data) {
		return nil, fmt.Errorf("%w: no compound file header", ErrNotWorkbook)
	}
	r, err := mscfb.New(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrNotWorkbook, err)
	}
	for f, nextErr := r.Next(); nextErr == nil; f, nextErr = r.Next() {
		if f.Name != name || len(f.Path) != 0 || f.FileInfo().IsDir() {
			continue
		}
		if f.Size < 0 || f.Size > int64(len(data)) {
			return nil, fmt.Errorf("%w: a %s stream of %d bytes in a file of %d", ErrNotWorkbook, name, f.Size, len(data))
		}
		out := make([]byte, f.Size)
		if _, err := io.ReadFull(f, out); err != nil {
			return nil, fmt.Errorf("%w: reading the %s stream: %w", ErrNotWorkbook, name, err)
		}
		return out, nil
	}
	return nil, fmt.Errorf("%w: no %s stream", ErrNotWorkbook, name)
}

// le16 and le32 read little-endian values at off.
func le16(b []byte, off int) uint16 { return binary.LittleEndian.Uint16(b[off:]) }
func le32(b []byte, off int) uint32 { return binary.LittleEndian.Uint32(b[off:]) }
