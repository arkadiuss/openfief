---
title: core token issue and revoke enter the administrator's entries
status: active
sourceFile: cmd/core/tokenissue_test.go
verification_scope: integration
origin:
  kind: agent
  ref: the S9 brief, from D-211, D-198 and D-67, 2026-11-14
  recorded_at: '2026-11-14T04:22:43.861Z'
id: TEST-core-admin-audit
type: test
---
Under task test:pg: token issue with and without the household grant, its failure paths and token revoke leave entries of the admin actor, as TestTokenIssueGrantsTheHousehold, TestTokenRevokeEndsItsGrants and TestTokenIssueRevokesWhenItsGrantCantReachSpiceDB check; the grants cutover and the sweeps stay the system's (TestGrantsStartupCutsOverOnce); the identity postgres test checks the table takes the admin actor with its own ID and no other.