---
title: The XTB workbook reader keeps only cells with a value and refuses a workbook past any of its limits
status: active
sourceFile: internal/ingest/xtb/xlsx_test.go
verification_scope: unit
origin:
  kind: agent
  recorded_at: '2026-11-02T20:57:25.488Z'
id: TEST-ingest-xtb-xlsx
type: test
---
TestReadWorkbookBounds builds tiny synthetic workbooks in the test with archive/zip and reads them with readWorkbookWithin: a sheet of 1,000 rows each with an empty cell and an inline string in column XFD must keep exactly 1,000 cells and find the text in the last column; and with each limit shrunk to a few units (cells, text bytes, shared strings, rows, the run between markup characters, nesting depth and sheets) a workbook one past it must fail with ErrTooLarge. Before the fix the padded sheet allocated about 1.6 GB; it now allocates under a megabyte.