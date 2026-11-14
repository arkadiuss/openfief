---
title: No PostgreSQL adapter imports a client that reaches the network
status: active
sourceFile: tools/archtest/rules_test.go
verification_scope: unit
origin:
  kind: agent
  ref: the S9b brief, author decisions of 2026-11-14 on D-211, D-198, D-196 and D-18
  recorded_at: '2026-11-14T06:29:55.600Z'
id: TEST-archtest-no-network-in-transactions
type: test
---
TestRules checks the rule no-network-in-transactions over fixtures: an adapter that imports net/http is a violation, one that doesn't, a package outside internal and a package that isn't an adapter are not; TestArchitecture runs the rule over the repository with an empty baseline.