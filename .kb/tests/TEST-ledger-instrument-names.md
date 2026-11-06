---
title: Display names are cleaned; the first delivered name stays until the owner renames
status: active
sourceFile: internal/ledger/instrumentname_test.go
verification_scope: unit
origin:
  kind: agent
  recorded_at: '2026-11-06T06:58:50.455Z'
id: TEST-ledger-instrument-names
type: test
---
Table tests of ledger.CleanInstrumentName (trimmed, collapsed, Polish letters kept, empty, control characters, invalid UTF-8 and over 200 characters refused), CurrentInstrumentNames (first delivered wins, owner's latest wins, a rename before any delivery) and FirstDeliveredNames (one per ISIN, invalid names and missing ISINs left out).