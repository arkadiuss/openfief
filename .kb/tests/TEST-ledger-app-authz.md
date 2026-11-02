---
title: A transfer decision needs the change of both legs' accounts
status: active
sourceFile: internal/ledger/app/transfers_test.go
verification_scope: unit
origin:
  kind: agent
  recorded_at: '2026-11-01T22:46:03.865Z'
id: TEST-ledger-app-authz
type: test
---
TestTransfersNeedTheChangeOfBothAccounts decides a pair over an in-memory ledger with the Owners fake denying either leg's account and checks the refusal and that no link is stored; TestTransferAccounts lists the accounts of stored legs once each.