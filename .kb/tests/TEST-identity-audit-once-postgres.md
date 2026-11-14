---
title: AppendAuditOnce writes an entry once per job, with the change's time
status: active
sourceFile: internal/identity/postgres/auditchange_test.go
verification_scope: integration
origin:
  kind: agent
  ref: the S9 brief, from D-211, D-198 and D-67, 2026-11-14
  recorded_at: '2026-11-14T04:22:46.230Z'
id: TEST-identity-audit-once-postgres
type: test
---
Under task test:pg, over the real store: an entry keyed by a job's ID is written once however often the worker runs, a second job's entry is written, both keep the change's time, an entry the domain refuses is not written, and the table's unique constraint refuses a second row under a key; the audit table takes the admin actor with its own ID and no other.