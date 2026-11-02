---
title: A hostile XTB workbook is refused, or read, within bounded memory
status: active
origin:
  kind: agent
  recorded_at: '2026-11-02T20:57:28.426Z'
id: SCEN-ingest-xtb-reader-bounded
type: scenario
---
Given an uploaded workbook whose every row has a cell in column XFD, when the reader reads it, then each row keeps one entry per cell with a value, never padding to the last column; given a workbook with more cells, shared strings, text, rows or sheets than the reader keeps, or a name, attribute list or text run longer than its limit, or elements nested deeper than its limit, then reading it fails with ErrTooLarge before anything is mapped. The parts are streamed, never held whole.