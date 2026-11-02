---
title: Stored bond terms and index levels name the instance that delivered them
status: active
verification_scope: integration
sourceFile: internal/valuation/postgres/deliveredby_test.go
origin:
  kind: agent
  recorded_at: '2026-11-02T22:24:39.931Z'
id: TEST-valuation-postgres-delivered-by
type: test
---
Against a fresh PostgreSQL database with every migration applied, stores one bond series' terms from one instance, then again from another, and reads the stored source back each time; storing terms or an index level with an empty source must fail. A store that kept the first instance or accepted no instance fails it.