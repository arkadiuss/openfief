---
title: The usage and the questions are kept in the database, exactly, under concurrent calls
status: active
sourceFile: internal/modelrouter/postgres/guards_test.go
verification_scope: integration
origin:
  kind: agent
  recorded_at: '2026-11-14T20:37:52.360Z'
id: TEST-models-postgres-guards
type: test
---
Against PostgreSQL: the day's costs add up per currency and survive a new store, 25 concurrent records lose none, questions are counted per session and per address with exact reset times, 30 simultaneous questions let exactly the limit through, the address is stored only as a salted hash, old questions are forgotten and the migration makes one salt however often it runs.