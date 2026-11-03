package xtb

import (
	"archive/zip"
	"bytes"
	"errors"
	"fmt"
	"strings"
	"testing"
)

// workbookOf zips a workbook whose xl/workbook.xml lists sheets, each a
// <sheet> element pointing at rId1, the worksheet sheetData, with the
// shared strings shared when not empty. Everything is synthetic.
func workbookOf(t *testing.T, sheets, sheetData, shared string) []byte {
	t.Helper()
	const ns = `xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main" ` +
		`xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships"`
	parts := map[string]string{
		"xl/workbook.xml": `<workbook ` + ns + `><sheets>` + sheets + `</sheets></workbook>`,
		"xl/_rels/workbook.xml.rels": `<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">` +
			`<Relationship Id="rId1" Target="worksheets/sheet1.xml"/></Relationships>`,
		"xl/worksheets/sheet1.xml": `<worksheet ` + ns + `><sheetData>` + sheetData + `</sheetData></worksheet>`,
	}
	if shared != "" {
		parts["xl/sharedStrings.xml"] = `<sst ` + ns + `>` + shared + `</sst>`
	}
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	for _, name := range []string{"xl/workbook.xml", "xl/_rels/workbook.xml.rels", "xl/worksheets/sheet1.xml", "xl/sharedStrings.xml"} {
		body, ok := parts[name]
		if !ok {
			continue
		}
		w, err := zw.Create(name)
		if err != nil {
			t.Fatalf("zip: %v", err)
		}
		if _, err := w.Write([]byte(body)); err != nil {
			t.Fatalf("zip: %v", err)
		}
	}
	if err := zw.Close(); err != nil {
		t.Fatalf("zip: %v", err)
	}
	return buf.Bytes()
}

// keptCells counts the cells s keeps, all rows together.
func keptCells(s *sheet) int {
	cells := 0
	for n := 1; n <= s.last; n++ {
		cells += len(s.row(n))
	}
	return cells
}

// TestReadWorkbookBounds: a cell in a row's last column, XFD, costs the
// reader one kept cell, not 16,384, and each of the reader's limits refuses
// a tiny workbook that passes it with ErrTooLarge (D-184). The limits are
// shrunk so every case stays a few bytes; defaultLimits sets the real ones.
func TestReadWorkbookBounds(t *testing.T) {
	t.Parallel()
	const one = `<sheet name="S" r:id="rId1"/>`
	var padded strings.Builder
	for n := 1; n <= 1000; n++ {
		fmt.Fprintf(&padded, `<row r="%d"><c r="XFD%d"/><c r="XFD%d" t="inlineStr"><is><t>x</t></is></c></row>`, n, n, n)
	}
	tiny := func(edit func(l *limits)) limits {
		l := defaultLimits()
		edit(&l)
		return l
	}
	tests := []struct {
		name                   string
		sheets, data, shared   string
		lim                    limits
		wantErr                error
		wantCells, wantColumns int
	}{
		{
			name: "a cell in the last column costs one cell", sheets: one, data: padded.String(), lim: defaultLimits(),
			wantCells: 1000, wantColumns: maxColumns,
		},
		{
			name: "cells over the limit", sheets: one, data: `<row><c><v>1</v></c><c><v>2</v></c><c><v>3</v></c></row>`,
			lim: tiny(func(l *limits) { l.cells = 2 }), wantErr: ErrTooLarge,
		},
		{
			name: "text over the limit", sheets: one, data: `<row><c t="inlineStr"><is><t>0123456789x</t></is></c></row>`,
			lim: tiny(func(l *limits) { l.textBytes = 10 }), wantErr: ErrTooLarge,
		},
		{
			name: "shared strings over the limit", sheets: one, shared: `<si><t>a</t></si><si><t>b</t></si><si><t>c</t></si>`,
			lim: tiny(func(l *limits) { l.sharedStrings = 2 }), wantErr: ErrTooLarge,
		},
		{
			name: "rows over the limit", sheets: one, data: strings.Repeat(`<row r="1"/>`, 6),
			lim: tiny(func(l *limits) { l.rows = 5 }), wantErr: ErrTooLarge,
		},
		{
			name: "text token over the limit", sheets: one, data: `<row><c><v>` + strings.Repeat("9", 5000) + `</v></c></row>`,
			lim: tiny(func(l *limits) { l.tokenBytes = 64 }), wantErr: ErrTooLarge,
		},
		{
			// '>' is legal in an attribute's value, so a token isn't bounded
			// by the markup characters in it.
			name: "attribute token over the limit", sheets: one, data: `<row><c r="A1" x="` + strings.Repeat(">", 5000) + `"/></row>`,
			lim: tiny(func(l *limits) { l.tokenBytes = 64 }), wantErr: ErrTooLarge,
		},
		{
			name: "nesting over the limit", sheets: one, data: strings.Repeat("<x>", 9) + strings.Repeat("</x>", 9),
			lim: tiny(func(l *limits) { l.depth = 8 }), wantErr: ErrTooLarge,
		},
		{
			name: "sheets over the limit", sheets: one + `<sheet name="T" r:id="rId2"/>`,
			lim: tiny(func(l *limits) { l.sheets = 1 }), wantErr: ErrTooLarge,
		},
		{
			name: "two sheets on one part", sheets: one + `<sheet name="T" r:id="rId1"/>`,
			lim: defaultLimits(), wantErr: ErrNotWorkbook,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			wb, err := readWorkbookWithin(workbookOf(t, tt.sheets, tt.data, tt.shared), tt.lim)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("readWorkbookWithin error = %v, want %v", err, tt.wantErr)
			}
			if err != nil {
				return
			}
			s := wb.sheets["S"]
			if cells := keptCells(s); cells != tt.wantCells {
				t.Errorf("kept %d cells, want %d", cells, tt.wantCells)
			}
			if got := s.row(1).at(tt.wantColumns - 1).value; got != "x" {
				t.Errorf("row 1, column %d = %q, want x", tt.wantColumns, got)
			}
		})
	}
}
