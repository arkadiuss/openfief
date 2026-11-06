---
title: What an import did in each account
status: active
sourceFile: internal/ledger/app/peraccount_test.go
verification_scope: unit
origin:
  kind: agent
  recorded_at: '2026-11-06T11:51:14.032Z'
id: TEST-ledger-app-per-account
type: test
---
Table test of ImportSummary.PerAccount: each account's transactions by decision and its flagged checks, an account with only a flagged check included, in ID order; the import tests also check the per-account counts.