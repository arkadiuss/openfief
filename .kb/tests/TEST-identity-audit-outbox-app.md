---
title: RecordOwnerChange hands the entry to the outbox with the change
status: active
sourceFile: internal/identity/app/ownerchange_test.go
verification_scope: unit
origin:
  kind: agent
  ref: the S9 brief, from D-211, D-198 and D-67, 2026-11-14
  recorded_at: '2026-11-14T04:22:45.431Z'
id: TEST-identity-audit-outbox-app
type: test
---
Over an in-memory outbox: the entry, with the owner's session and the time of the change, is enqueued only when the change succeeds; a failed change leaves none; an outbox that can't take the entry, or none, means the change is not made and the error says the log is unavailable; an entry the domain refuses means no change.