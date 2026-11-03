package xtb

import (
	"archive/zip"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"testing"
)

// heapCaseEnv names the case TestWorkbookHeapChild reads, and heapSizeEnv
// the size of its file on standard input, in the child process
// TestWorkbookHeap starts.
const (
	heapCaseEnv = "OPENFIEF_WORKBOOK_HEAP_CASE"
	heapSizeEnv = "OPENFIEF_WORKBOOK_HEAP_SIZE"
)

// workbookHeapCase is one uploaded file and the most the reader's heap may
// grow reading it.
type workbookHeapCase struct {
	name    string
	file    func() ([]byte, error)
	wantErr error
	maxMiB  uint64
}

// workbookHeapCases are hostile files from the second security review and
// the largest workbook the reader keeps, each with how far the heap may
// grow while the reader reads it, with Go's default GC: the xtb plugin's
// pod may use 256 MiB, and it holds the file too. Measured before the
// fixes: 372 MiB, 1.5 GiB, 20 MiB and 32 MiB for the first four, and
// 132 MiB for a million cells, the old cap.
func workbookHeapCases() []workbookHeapCase {
	return []workbookHeapCase{
		{
			// '>' is legal in an attribute's value, so it never ended the
			// old run of bytes between markup characters.
			name: "an attribute of 120 MiB of '>'", wantErr: ErrTooLarge, maxMiB: 16,
			file: func() ([]byte, error) {
				return zipOf(`<sheet name="S" r:id="rId1"/>`,
					`<worksheet><sheetData><row><c r="A1" x="`, strings.Repeat(">", 1<<20), 120, `"/></row></sheetData></worksheet>`)
			},
		},
		{
			name: "100 sheets on one part of 200,000 empty rows", wantErr: ErrNotWorkbook, maxMiB: 16,
			file: func() ([]byte, error) {
				var sheets strings.Builder
				for i := range 100 {
					fmt.Fprintf(&sheets, `<sheet name="S%d" r:id="rId1"/>`, i)
				}
				return zipOf(sheets.String(),
					`<worksheet><sheetData>`, `<row/>`, maxRows, `</sheetData></worksheet>`)
			},
		},
		{
			name: "one sheet of 200,000 empty rows", maxMiB: 16,
			file: func() ([]byte, error) {
				return zipOf(`<sheet name="S" r:id="rId1"/>`,
					`<worksheet><sheetData>`, `<row/>`, maxRows, `</sheetData></worksheet>`)
			},
		},
		{
			name: "150,000 zip entries", wantErr: ErrTooLarge, maxMiB: 16,
			file: func() ([]byte, error) { return manyEntries(150_000) },
		},
		{
			// Half a million cells of 32 bytes of text each: every cell and
			// all the text the reader keeps.
			name: "the most cells and text kept", maxMiB: 112,
			file: func() ([]byte, error) {
				cells := strings.Repeat(`<c t="inlineStr"><is><t>`+strings.Repeat("x", 32)+`</t></is></c>`, 100)
				return zipOf(`<sheet name="S" r:id="rId1"/>`,
					`<worksheet><sheetData>`, `<row>`+cells+`</row>`, maxCells/100, `</sheetData></worksheet>`)
			},
		},
	}
}

// TestWorkbookHeap: whatever file the owner uploads, reading it grows the
// heap a bounded amount (D-184). Each case's file is built here and read
// alone in a child process, so neither building it nor any other test
// leaves memory the read could reuse unseen.
func TestWorkbookHeap(t *testing.T) {
	t.Parallel()
	self, err := os.Executable()
	if err != nil {
		t.Fatalf("test binary: %v", err)
	}
	for _, tc := range workbookHeapCases() {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			data, err := tc.file()
			if err != nil {
				t.Fatalf("build the file: %v", err)
			}
			grown := childHeapGrowth(t, self, tc.name, data)
			t.Logf("%s: peak heap grew %d MiB", tc.name, grown>>20)
			if grown > tc.maxMiB<<20 {
				t.Errorf("reading the file grew the heap %d MiB, want at most %d MiB", grown>>20, tc.maxMiB)
			}
		})
	}
}

// childHeapGrowth runs TestWorkbookHeapChild in the test binary self for
// case name, with data on its standard input, and returns how far its heap
// grew.
func childHeapGrowth(t *testing.T, self, name string, data []byte) uint64 {
	t.Helper()
	cmd := exec.CommandContext(t.Context(), self, "-test.run=^TestWorkbookHeapChild$", "-test.count=1", "-test.v")
	cmd.Env = append(os.Environ(), heapCaseEnv+"="+name, heapSizeEnv+"="+strconv.Itoa(len(data)))
	cmd.Stdin = bytes.NewReader(data)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("child: %v\n%s", err, out)
	}
	at := bytes.Index(out, []byte("heap-growth="))
	if at < 0 {
		t.Fatalf("child printed no heap growth:\n%s", out)
	}
	var grown uint64
	if _, err := fmt.Sscanf(string(out[at:]), "heap-growth=%d", &grown); err != nil {
		t.Fatalf("child output %q: %v", out, err)
	}
	return grown
}

// TestWorkbookHeapChild reads the file on its standard input and logs how
// far the heap grew. Outside TestWorkbookHeap's child process it has
// nothing to do.
func TestWorkbookHeapChild(t *testing.T) {
	t.Parallel()
	name := os.Getenv(heapCaseEnv)
	if name == "" {
		return
	}
	size, err := strconv.Atoi(os.Getenv(heapSizeEnv))
	if err != nil {
		t.Fatalf("file size: %v", err)
	}
	data := make([]byte, size)
	if _, readErr := io.ReadFull(os.Stdin, data); readErr != nil {
		t.Fatalf("read the file: %v", readErr)
	}
	for _, tc := range workbookHeapCases() {
		if tc.name != name {
			continue
		}
		var wb *workbook
		grown := heapGrowth(func() { wb, err = readWorkbook(data) })
		runtime.KeepAlive(wb)
		t.Logf("heap-growth=%d", grown)
		if !errors.Is(err, tc.wantErr) {
			t.Fatalf("readWorkbook error = %v, want %v", err, tc.wantErr)
		}
		return
	}
	t.Fatalf("no case %q", name)
}

// heapGrowth returns how much heap the process obtained while f ran: the
// runtime never returns heap address space, so the growth of HeapSys is
// the peak heap f needed beyond what was free before it. Stacks come out
// of the heap, so they are counted with it.
func heapGrowth(f func()) uint64 {
	var before, after runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&before)
	f()
	runtime.ReadMemStats(&after)
	end, start := after.HeapSys+after.StackSys, before.HeapSys+before.StackSys
	return max(end, start) - start
}

// zipOf zips a workbook whose xl/workbook.xml lists sheets, each pointing
// at rId1, and whose part for rId1, xl/worksheets/sheet1.xml, is head,
// repeat written n times, and tail, streamed through the compressor so the
// part is never held whole.
func zipOf(sheets, head, repeat string, n int, tail string) ([]byte, error) {
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	parts := []struct{ name, body string }{
		{"xl/workbook.xml", `<workbook xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships">` +
			`<sheets>` + sheets + `</sheets></workbook>`},
		{"xl/_rels/workbook.xml.rels", `<Relationships><Relationship Id="rId1" Target="worksheets/sheet1.xml"/></Relationships>`},
	}
	for _, p := range parts {
		if err := writePart(zw, p.name, strings.NewReader(p.body)); err != nil {
			return nil, err
		}
	}
	body := io.MultiReader(strings.NewReader(head), &repeated{s: repeat, n: n}, strings.NewReader(tail))
	if err := writePart(zw, "xl/worksheets/sheet1.xml", body); err != nil {
		return nil, err
	}
	if err := zw.Close(); err != nil {
		return nil, fmt.Errorf("zip: %w", err)
	}
	return bytes.Clone(buf.Bytes()), nil
}

// writePart writes one part of r to zw.
func writePart(zw *zip.Writer, name string, r io.Reader) error {
	w, err := zw.Create(name)
	if err != nil {
		return fmt.Errorf("zip %s: %w", name, err)
	}
	if _, err := io.Copy(w, r); err != nil {
		return fmt.Errorf("zip %s: %w", name, err)
	}
	return nil
}

// repeated reads s n times.
type repeated struct {
	s   string
	n   int
	off int
}

// Read reads the next bytes of s repeated.
func (r *repeated) Read(b []byte) (int, error) {
	if r.n == 0 {
		return 0, io.EOF
	}
	k := copy(b, r.s[r.off:])
	r.off += k
	if r.off == len(r.s) {
		r.off = 0
		r.n--
	}
	return k, nil
}

// manyEntries is a zip of n empty stored entries.
func manyEntries(n int) ([]byte, error) {
	buf := bytes.NewBuffer(make([]byte, 0, n*96))
	zw := zip.NewWriter(buf)
	for i := range n {
		if _, err := zw.CreateHeader(&zip.FileHeader{Name: fmt.Sprintf("%x", i), Method: zip.Store}); err != nil {
			return nil, fmt.Errorf("zip: %w", err)
		}
	}
	if err := zw.Close(); err != nil {
		return nil, fmt.Errorf("zip: %w", err)
	}
	return buf.Bytes(), nil
}
