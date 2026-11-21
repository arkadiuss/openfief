---
title: On PostgreSQL
status: active
sourceFile: internal/thesis/postgres/store_test.go
verification_scope: integration
origin:
  kind: agent
  ref: 'Implement D-225 and D-229: opinions and draft theses (design session brief, 2026-11-21)'
  recorded_at: '2026-11-21T15:00:00.000Z'
id: TEST-thesis-store
type: test
---
On PostgreSQL: versions kept in order, stale and missing saves store nothing, a failing hook rolls back, the list is the latest versions newest first, a saved version is never changed or removed, the table refuses what the domain refuses.