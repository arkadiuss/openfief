---
title: PostgreSQL stores index levels per index and date, replaces them and revalues nothing
status: active
sourceFile: internal/valuation/postgres/level_test.go
origin:
  kind: agent
  recorded_at: '2026-10-28T22:42:18.787Z'
id: TEST-valuation-postgres-index-levels
type: test
verification_scope: integration
---
Under task test:pg on a disposable PostgreSQL: three synthetic levels of FAKE20 and OTHER are saved, one is saved again with another level and source and replaces the first, IndexLevels returns FAKE20's in date order and only through the given date, an unknown index returns none, nothing is enqueued in valuation.dirty or River, and the table refuses a zero level.