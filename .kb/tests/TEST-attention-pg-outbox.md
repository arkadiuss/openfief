---
title: No outbox row holds a broker account number
status: active
verification_scope: integration
sourceFile: internal/attention/postgres/outbox_test.go
origin:
  kind: agent
  recorded_at: '2026-11-03T15:54:47.727Z'
id: TEST-attention-pg-outbox
type: test
---
Against a fresh PostgreSQL database with the attention migrations, appends an unmapped broker account's item whose subject is the instance and a synthetic broker account number, reads every outbox row's raw payload, and checks it names only the instance and never the number, while the stored item keeps its full subject. A payload quoting the number fails it.