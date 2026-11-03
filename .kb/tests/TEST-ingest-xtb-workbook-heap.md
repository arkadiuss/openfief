---
title: Hostile and the largest allowed XTB workbooks grow the reader's heap a bounded amount
status: active
sourceFile: internal/ingest/xtb/heap_test.go
verification_scope: integration
origin:
  kind: agent
  recorded_at: '2026-11-03T07:17:55.180Z'
id: TEST-ingest-xtb-workbook-heap
type: test
---
TestWorkbookHeap builds each workbook in the test, streamed through the zip compressor so the files stay small, and reruns its own test binary per case with the file on standard input; the child TestWorkbookHeapChild reads it with readWorkbook and reports how far HeapSys grew. A 120 MiB attribute of '>', 100 sheets on one part of 200,000 empty rows, one sheet of 200,000 empty rows and 150,000 zip entries must each stay within 16 MiB, and half a million cells with 16 MiB of text within 112 MiB.