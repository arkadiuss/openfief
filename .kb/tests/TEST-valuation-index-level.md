---
title: An index level needs an index name, a date and a positive level
status: active
sourceFile: internal/valuation/level_test.go
origin:
  kind: agent
  recorded_at: '2026-10-28T22:42:16.928Z'
id: TEST-valuation-index-level
type: test
verification_scope: unit
---
Table test of IndexLevel.Validate on the synthetic index FAKE20: a level of 1522.60 on a date passes, as does a 64-character name; no name, a 65-character name, no date, a zero and a negative level fail with ErrInvalidIndexLevel. A false pass would need Validate to accept a level the backtest would divide by.