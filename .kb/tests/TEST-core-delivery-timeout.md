---
title: A delivery that runs past its timeout fails whole with its own class, one in time commits
status: active
sourceFile: cmd/core/outbox_test.go
verification_scope: integration
origin:
  kind: agent
  ref: the S9b brief, author decisions of 2026-11-14 on D-211, D-198, D-196 and D-18
  recorded_at: '2026-11-14T06:29:52.639Z'
id: TEST-core-delivery-timeout
type: test
---
Under task test:pg: TestSyncDeliveryThatRunsPastItsTimeoutFailsWhole runs a delivery with a statement that outlasts the delivery timeout and checks it fails with ErrDeliveryTimeout, the attempt's class is delivery-timeout, and neither its rows nor its entry were written; TestSyncDeliveryInTimeCommitsAndLeavesNoTimeoutBehind checks a delivery in time commits with its entry and leaves the connection's own timeout; TestParseArgs in main_test.go checks the flag and its refusal of zero.