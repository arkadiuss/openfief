---
title: The audit-log-behind check raises one condition with the count, or none
status: active
sourceFile: internal/attention/app/auditlog_test.go
verification_scope: unit
origin:
  kind: agent
  ref: the S9b brief, author decisions of 2026-11-14 on D-211, D-198, D-196 and D-18
  recorded_at: '2026-11-14T06:29:57.050Z'
id: TEST-attention-app-audit-log-behind
type: test
---
TestAuditLogBehind checks no condition while nothing is stuck, one condition of the one audit log with the detail for one and for several changes, and an error, not a cleared item, when the queue can't be read.