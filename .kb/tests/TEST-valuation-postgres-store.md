---
title: Inputs and holding valuations round-trip through the valuation schema with their three values
status: active
sourceFile: internal/valuation/postgres/store_test.go
origin:
  kind: agent
  recorded_at: '2026-10-28T07:55:58.194Z'
id: TEST-valuation-postgres-store
type: test
---
PostgreSQL tests (task test:pg): inputs round-trip, a tax wrapper is stored as the pack's key and none reads as unrecorded; holding valuations round-trip with gross, net and liquidation values or each missing value's reason (marginal_rate, lots), replacing from a date keeps the days before, and a row on an unknown basis is refused.