---
title: PostgreSQL keeps each instance's broker accounts, their mapping and their seeding
status: active
sourceFile: internal/ingest/postgres/brokers_test.go
origin:
  kind: agent
  recorded_at: '2026-10-29T11:35:35.718Z'
id: TEST-ingest-postgres-brokers
type: test
---
PostgreSQL tests under task test:pg: recording synthetic broker accounts twice keeps the last seen date and the last delivery flag, mapping is idempotent and refused for another account or another instance, seeding marks mapped accounts only, all instances list in order, a broker account with a space is refused, and an instance counts as seeded only while every mapped broker account is seeded.