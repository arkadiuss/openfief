---
title: Index levels and index readings report only the rows that were new or different
status: active
sourceFile: internal/valuation/postgres/changed_test.go
verification_scope: integration
origin:
  kind: agent
  ref: the S9b brief, author decisions of 2026-11-14 on D-211, D-198, D-196 and D-18
  recorded_at: '2026-11-14T06:30:00.721Z'
id: TEST-valuation-postgres-market-data-changed
type: test
---
Under task test:pg: TestSaveIndexLevelsReportsOnlyWhatWasNewOrDifferent and TestSaveIndexReadingsReportsOnlyWhatWasNewOrDifferent check that a new row, or one that differs in its value or its source, counts, and the same row again, however its value is written, counts for none and leaves one row.