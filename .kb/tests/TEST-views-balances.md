---
title: The balance check passes cash within 0.01 and holdings with equal units, keeps the latest assertions and the latest batch
status: active
sourceFile: internal/views/balances_test.go
verification_scope: unit
origin:
  kind: agent
  recorded_at: '2026-11-06T23:37:07.759Z'
id: TEST-views-balances
type: test
---
TestCheckBalance tables synthetic cash and holding assertions against derived balances: cash passes at exactly 0.01 over or under and fails just past it or a unit off, a holding passes only with equal units and fails a hundredth of a unit off, and each check carries reported minus derived as its difference. TestLatestBalanceChecks keeps the latest assertion of each account and ledger account in account, date and ledger account order and counts failures, and TestLatestUnsupported counts only each instance's latest completed batch, so open CFDs closed since drop out and the greater ID breaks a tie.