---
title: An upload's mode follows the account's seeding unless the owner overrides it; undeclared cross-check prices are one problem
status: active
sourceFile: internal/ingest/declarations_test.go
origin:
  kind: agent
  recorded_at: '2026-10-30T04:10:37.473Z'
id: TEST-ingest-declarations
type: test
---
Table test of ingest.UploadMode: an account not yet seeded gets opening, a seeded one transactions, and either override wins while an unknown mode is no override; and of UndeclaredCrossCheckPrices, a skipped problem that counts the prices left out. A mode chosen from anything but seeding would fail a row.