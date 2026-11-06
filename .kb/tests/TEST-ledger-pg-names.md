---
title: PostgreSQL keeps one delivered name per ISIN and every rename, append-only
status: active
sourceFile: internal/ledger/postgres/names_test.go
verification_scope: integration
origin:
  kind: agent
  recorded_at: '2026-11-06T06:58:54.248Z'
id: TEST-ledger-pg-names
type: test
---
Under task test:pg: two deliveries for one ISIN keep the first, another ISIN's is kept, an owner's rename is appended, all listed in order; UPDATE, DELETE and TRUNCATE are refused by the append-only triggers.