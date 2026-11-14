---
title: The database takes every action, the operator, and enters an owner's change only with the change
status: active
sourceFile: internal/identity/postgres/auditchange_test.go
verification_scope: integration
origin:
  kind: agent
  ref: the S8 brief, from D-211, 2026-11-14
  recorded_at: '2026-11-14T02:39:42.458Z'
id: TEST-identity-audit-change-postgres
type: test
---
Under task test:pg: a row of every domain action and of the operator actor is accepted and a wrong actor ID refused; RecordOwnerChange over the real store keeps the entry of a change that was made, none for a change that failed, and makes no change when the log is gone; plugin generations are entered once each over the real store.