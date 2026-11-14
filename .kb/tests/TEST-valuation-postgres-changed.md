---
title: Saving prices, FX rates and bond terms counts only the rows that were new or different
status: active
sourceFile: internal/valuation/postgres/changed_test.go
verification_scope: integration
origin:
  kind: agent
  ref: the S9 brief, from D-211, D-198 and D-67, 2026-11-14
  recorded_at: '2026-11-14T04:22:48.515Z'
id: TEST-valuation-postgres-changed
type: test
---
Under task test:pg: the same prices, rates and terms sent again count for none, however the amount is written; a different value, another source's row and the first send count; the stored row is replaced by the different one.