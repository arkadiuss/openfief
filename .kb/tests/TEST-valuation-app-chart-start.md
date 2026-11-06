---
title: The owner's charts' start setting is stored per portfolio and account, and the default clears it
status: active
sourceFile: internal/valuation/app/chartstart_test.go
verification_scope: unit
origin:
  kind: agent
  recorded_at: '2026-11-02T08:22:15.698Z'
id: TEST-valuation-app-chart-start
type: test
---
Runs valuation's ChartStarts over an in-memory store, a fake ledger and fake owners that refuse one account as the authorizer would. The portfolio's and an account's starts are stored apart, a zero start clears either, a read session, a refused account and an account that is not stored are refused with nothing stored, and the authorizer is asked before the accounts are listed, so a refused account fails alike whether stored or not.