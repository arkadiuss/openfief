---
title: The core raises the item at five failures, follows the count and clears it
status: active
sourceFile: cmd/core/auditbehind_test.go
verification_scope: integration
origin:
  kind: agent
  ref: the S9b brief, author decisions of 2026-11-14 on D-211, D-198, D-196 and D-18
  recorded_at: '2026-11-14T06:29:57.801Z'
id: TEST-core-audit-log-behind
type: test
---
Under task test:pg: TestAuditLogBehindItemIsRaisedAtFiveFailures holds audit jobs in a real queue and checks that jobs that failed four times raise nothing and two that failed five times raise the one open item of data coming in with the count; TestAuditLogBehindItemClearsWhenTheStuckEntriesGetThrough checks the count follows the jobs still stuck and the item resolves when none is.