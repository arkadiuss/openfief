---
title: Every network client the core builds refuses a call made inside a change, and what a change defers runs after it commits
status: active
sourceFile: cmd/core/nonetwork_test.go
verification_scope: integration
origin:
  kind: agent
  ref: the S9b brief, author decisions of 2026-11-14 on D-211, D-198, D-196 and D-18
  recorded_at: '2026-11-14T06:29:54.169Z'
id: TEST-core-no-network-in-change
type: test
---
TestHTTPIsRefusedInsideAChange, TestPluginClientIsRefusedInsideAChange, TestGRPCIsRefusedInsideAChange and TestSpiceDBIsRefusedInsideAChange check each client fails a call made with a change's context with ErrNetworkInChange before anything is sent; TestPlacementsReachSpiceDBAfterTheChangeCommits and its two siblings check what SpiceDB is told is held until the commit; under task test:pg, TestAfterChangeRunsOnlyOnceTheOutermostChangeCommits, TestAfterChangeNeverRunsForAChangeThatRollsBack and TestAfterChangeThatFailsLeavesTheChangeCommitted check the follow-ups over a real database.