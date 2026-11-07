---
title: Only the owner gets the balance report, over the accounts the authorizer lets it view
status: active
sourceFile: internal/views/app/balances_test.go
verification_scope: unit
origin:
  kind: agent
  recorded_at: '2026-11-06T23:37:08.562Z'
id: TEST-views-app-balances
type: test
---
TestBalanceReportOwnerOnly asks for the report as nobody, a script token, a view plugin and a demo session and expects ErrOwnerOnly each time. TestBalanceReport asks as the owner over fakes of the ledger and the sync history, with an authorizer that denies acc-2: the report holds acc-1's assertions only, the latest of each ledger account by default and every one with all, with the derived balances the ledger stored, and the unsupported rows of the XTB instance bound to acc-1, not those of the instance bound to acc-2.