package treasurybonds_test

import (
	"encoding/binary"
	"slices"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/arkadiuss/openfief/internal/ingest/treasurybonds"
	"github.com/arkadiuss/openfief/internal/jurisdiction/pl"
	"github.com/arkadiuss/openfief/internal/kernel"
)

// rates is one series' synthetic published rates.
func rates(tb testing.TB, code, first, margin string) pl.PublishedRates {
	tb.Helper()
	s, err := pl.ParseSeries(code)
	if err != nil {
		tb.Fatalf("ParseSeries(%q): %v", code, err)
	}
	f, err := kernel.ParseDecimal(first)
	if err != nil {
		tb.Fatalf("ParseDecimal(%q): %v", first, err)
	}
	m, err := kernel.ParseDecimal(margin)
	if err != nil {
		tb.Fatalf("ParseDecimal(%q): %v", margin, err)
	}
	return pl.PublishedRates{Series: s, FirstPeriodRate: f, Margin: m}
}

// header is a sheet's header row as the ministry's file lays it out: the
// series, then the rate columns from column 9, then the margin at
// marginCol, if any.
func header(marginCol int) []xcell {
	row := []xcell{text(0, "Seria"), text(1, "Kod ISIN"), text(2, "Data wykupu"), text(9, "Oprocentowanie")}
	if marginCol >= 0 {
		row = append(row, text(marginCol, "Marża"))
	}
	return row
}

// syntheticWorkbook is a workbook in the ministry's layout with made-up
// series, rates and registry ISINs: a description sheet, OTS with no margin
// column, an ISIN with a wrong check digit and one padded with spaces, EDO
// with a sub-header row, a row of another type, a provisional ISIN, a
// series listed with its ISIN before its rate and one without either yet,
// and DOR with its numbers written each way Excel writes them and no ISIN
// column.
func syntheticWorkbook(tb testing.TB) []byte {
	tb.Helper()
	return xls(tb,
		xsheet{name: "Opis", rows: [][]xcell{{text(0, "Dane dotyczące obligacji detalicznych")}, {text(0, "Seria"), num(1, "1")}}},
		xsheet{name: "OTS", rows: [][]xcell{
			header(-1),
			{text(0, "OTS0199"), text(1, "PL0000000001"), num(9, "0.0111")},
			{text(0, "OTS0299"), text(1, " PL9999000059 "), num(9, "0.0123")},
		}},
		xsheet{name: "EDO", rows: [][]xcell{
			header(20),
			{text(9, " w 1. roku"), text(10, " w 2. roku")},
			{text(0, "EDO0199"), text(1, "PL9999000018"), num(9, "0.0611"), num(10, "0.0711"), num(20, "0.0222")},
			{text(0, "EDO0299"), text(1, "PLEDO0299005"), num(9, "0.0612"), num(20, "0.0222")},
			{text(0, "COI0199"), text(1, "PL9999000026"), num(9, "0.05"), num(20, "0.01")},
			{text(0, "EDO0399"), text(1, "PL9999000034"), num(20, "0.0222")},
			{text(0, "EDO0599"), text(1, "-")},
			{text(0, "EDO0499"), num(9, "0.0613")},
		}},
		xsheet{name: "DOR", rows: [][]xcell{
			{text(0, "Seria"), {col: 9, kind: inlineText, val: "Oprocentowanie"}, {col: 57, kind: formulaText, val: "Marża"}},
			{{col: 0, kind: formulaText, val: "DOR0199"}, {col: 9, kind: rkHundredth, val: "7"}, {col: 57, kind: formulaNum, val: "0.0015"}},
			{{col: 0, kind: inlineText, val: "DOR0299"}, {col: 9, kind: rkInteger, val: "0"}, {col: 57, kind: double, val: "0.039999999999999994"}},
		}},
	)
}

// TestParseWorkbook reads a synthetic workbook: every series of the bond
// type sheets, rate and margin as fractions, fixed-rate types with no
// margin, rows without a series or a rate left out, numbers written each
// way exact, and a double's rounding error rounded away; and each series'
// registry ISIN (D-214), one listed before its rates too, but none that
// isn't a valid Polish ISIN or is a provisional one.
func TestParseWorkbook(t *testing.T) {
	t.Parallel()
	got, isins, err := treasurybonds.ParseWorkbook(syntheticWorkbook(t))
	if err != nil {
		t.Fatalf("ParseWorkbook: %v", err)
	}
	want := []pl.PublishedRates{
		rates(t, "OTS0199", "0.0111", "0"),
		rates(t, "OTS0299", "0.0123", "0"),
		rates(t, "DOR0199", "0.07", "0.0015"),
		rates(t, "DOR0299", "0", "0.04"),
		rates(t, "EDO0199", "0.0611", "0.0222"),
		rates(t, "EDO0299", "0.0612", "0.0222"),
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("ParseWorkbook (-want +got):\n%s", diff)
	}
	wantISINs := []pl.PublishedISIN{
		{Series: rates(t, "OTS0299", "0", "0").Series, ISIN: "PL9999000059"},
		{Series: rates(t, "EDO0199", "0", "0").Series, ISIN: "PL9999000018"},
		{Series: rates(t, "EDO0399", "0", "0").Series, ISIN: "PL9999000034"},
	}
	if diff := cmp.Diff(wantISINs, isins); diff != "" {
		t.Errorf("ParseWorkbook ISINs (-want +got):\n%s", diff)
	}
}

// TestParseWorkbookRefuses checks what the reader refuses: anything that
// is no workbook, a cut-short file, percentages instead of fractions, a
// series listed twice, one ISIN listed for two series, and a workbook
// without any rate.
func TestParseWorkbookRefuses(t *testing.T) {
	t.Parallel()
	good := syntheticWorkbook(t)
	tests := []struct {
		name    string
		data    []byte
		wantErr error
	}{
		{name: "not a compound file", data: []byte("<html>not a spreadsheet</html>"), wantErr: treasurybonds.ErrNotWorkbook},
		{name: "cut short", data: good[:1024], wantErr: treasurybonds.ErrNotWorkbook},
		{name: "a stream larger than the file", data: patched(good, workbookSizeAt, 0x7FFFFFFF), wantErr: treasurybonds.ErrNotWorkbook},
		{name: "a directory chain that loops", data: patched(good, directoryFATAt, 1), wantErr: treasurybonds.ErrNotWorkbook},
		{name: "percentages", wantErr: treasurybonds.ErrNotFraction, data: xls(t, xsheet{name: "TOS", rows: [][]xcell{
			header(-1), {text(0, "TOS0199"), num(9, "4.4")},
		}})},
		{name: "a series twice", wantErr: treasurybonds.ErrNoRates, data: xls(t, xsheet{name: "TOS", rows: [][]xcell{
			header(-1), {text(0, "TOS0199"), num(9, "0.044")}, {text(0, "TOS0199"), num(9, "0.045")},
		}})},
		{name: "one ISIN for two series", wantErr: treasurybonds.ErrNoRates, data: xls(t, xsheet{name: "TOS", rows: [][]xcell{
			header(-1), {text(0, "TOS0199"), text(1, "PL9999000018"), num(9, "0.044")}, {text(0, "TOS0299"), text(1, "PL9999000018")},
		}})},
		{name: "no headers", wantErr: treasurybonds.ErrNoRates, data: xls(t, xsheet{name: "TOS", rows: [][]xcell{
			{text(0, "TOS0199"), num(9, "0.044")},
		}})},
		{name: "no bond type sheet", wantErr: treasurybonds.ErrNoRates, data: xls(t, xsheet{name: "Opis", rows: [][]xcell{header(-1)}})},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, _, err := treasurybonds.ParseWorkbook(tt.data)
			if diff := cmp.Diff(tt.wantErr, err, cmpopts.EquateErrors()); diff != "" {
				t.Errorf("ParseWorkbook error (-want +got):\n%s", diff)
			}
		})
	}
}

// Offsets in the synthetic compound file (compoundFile): the Workbook
// directory entry's stream size, and the FAT entry of the directory's
// sector, sector 1.
const (
	workbookSizeAt = 512 + 512 + 128 + 120
	directoryFATAt = 512 + 4
)

// patched returns a copy of data with the 32-bit little-endian value v at
// off.
func patched(data []byte, off int, v uint32) []byte {
	out := slices.Clone(data)
	binary.LittleEndian.PutUint32(out[off:], v)
	return out
}

// FuzzParseWorkbook feeds the reader damaged workbooks: whatever the bytes,
// it returns rates or an error, and never panics or hangs.
func FuzzParseWorkbook(f *testing.F) {
	good := syntheticWorkbook(f)
	f.Add(good)
	f.Add(good[:600])
	f.Fuzz(func(t *testing.T, data []byte) {
		if _, _, err := treasurybonds.ParseWorkbook(data); err == nil && len(data) < 512 {
			t.Errorf("ParseWorkbook read rates from %d bytes", len(data))
		}
	})
}
