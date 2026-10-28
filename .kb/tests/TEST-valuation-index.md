---
title: Index kinds and the period index readings a bond lot needs
status: active
sourceFile: internal/valuation/index_test.go
origin:
  kind: agent
  recorded_at: '2026-10-28T07:28:16.381Z'
id: TEST-valuation-index
type: test
---
Table tests parse and print the index kind names cpi and reference_rate and reject other spellings, then pick readings from a synthetic set for synthetic lots: an EDO lot bought 2024-03-15 needs none in its first year, January 2025 CPI from 2025-03-15 and January 2026 CPI from 2026-03-15; a TOS lot needs none; a ROR lot bought 2026-03-10 takes the April 2026 reference-rate reading from 2026-04-10 and leaves out the May reading that was not supplied. An unknown bond type fails.