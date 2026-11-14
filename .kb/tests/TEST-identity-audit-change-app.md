---
title: RecordOwnerChange keeps the entry with the change, and plugin observations enter once per generation
status: active
sourceFile: internal/identity/app/ownerchange_test.go
verification_scope: unit
origin:
  kind: agent
  ref: the S8 brief, from D-211, 2026-11-14
  recorded_at: '2026-11-14T02:39:41.602Z'
id: TEST-identity-audit-change-app
type: test
---
Runs the use cases over the in-memory store whose Atomically restores its tables on failure: the entry is kept when the change succeeds, not when it fails, and a log that is down or an entry the domain refuses means the change is not called; plugin observations enter an install, upgrade and removal once, ignore a resource being deleted or without a generation, and leave the generations as they were when the log is down.