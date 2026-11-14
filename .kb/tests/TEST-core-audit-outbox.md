---
title: A change and its pending entry commit or roll back together, over a real database
status: active
sourceFile: cmd/core/outbox_test.go
verification_scope: integration
origin:
  kind: agent
  ref: the S9 brief, from D-211, D-198 and D-67, 2026-11-14
  recorded_at: '2026-11-14T04:22:42.321Z'
id: TEST-core-audit-outbox
type: test
---
Under task test:pg: a change that is made leaves its writes in two stores and one job with the change's time; a change that fails after its writes, or whose second store's statement the database refuses, leaves neither writes nor a job and the next change goes through; a step that fails inside a savepoint and is handled costs the change nothing else; an enqueue that fails fails the change with ErrAuditUnavailable and rolls its writes back; an entry can't be enqueued outside a change's transaction. TestEveryOwnerChangeWritesExactlyOneEntry drives every owner procedure through the real core and, with audit jobs refused, checks no change succeeds; TestOwnerChangesWaitForTheLogInTheOutbox takes the log away and checks the change is made, the entry waits, and the worker writes it once, with the change's time, when the log is back.