---
title: An index reading needs a known index, a date, a source, and CPI on the first of its month
status: active
sourceFile: internal/ingest/index_test.go
origin:
  kind: agent
  recorded_at: '2026-10-28T08:54:20.427Z'
id: TEST-ingest-index-reading
type: test
verification_scope: unit
---
Table test of IndexReading.Validate with synthetic readings: CPI dated on the first of a month and a reference rate dated any day pass; CPI on the 15th, a zero or unknown index, a missing date and a missing source fail with ErrInvalidIndexReading. Index.String names cpi and reference-rate and a placeholder for invalid values.