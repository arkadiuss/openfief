---
title: Through the real core and PostgreSQL, removals are refused while the ledger refers to them, and broker accounts reach only the owner
status: active
sourceFile: cmd/core/removals_test.go
verification_scope: integration
origin:
  kind: agent
  recorded_at: '2026-10-31T14:35:08.460Z'
id: TEST-core-removals
type: test
---
Runs from TestOnboardingServices after a connector sync booked a deposit from broker account U0000001 into ibkr-main: the read API's GET /v1/broker-accounts answers 501 naming no broker account while the AccountService lists them; deleting ibkr-main, remapping U0000001 and deleting the sleeve and kind in use are FailedPrecondition; an empty account takes and gives up U0000002 and is deleted twice; an empty sleeve and then its kind are deleted.