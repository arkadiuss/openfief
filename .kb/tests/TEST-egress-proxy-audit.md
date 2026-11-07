---
title: egress-proxy audit prints counts and destinations and fails after printing a problem
status: active
sourceFile: cmd/egress-proxy/audit_test.go
verification_scope: unit
origin:
  kind: agent
  recorded_at: '2026-11-07T03:30:42.566Z'
id: TEST-egress-proxy-audit
type: test
---
TestReport checks the audit's output for a clean audit, the counts and each allowed destination by instance and host followed by the passing line, and for one with a problem, which is printed and makes report return errAuditFailed. TestRunAuditRefusesBadFlags checks an unknown flag is refused before any cluster is read.