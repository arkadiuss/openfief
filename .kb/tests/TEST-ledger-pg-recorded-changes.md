---
title: PostgreSQL lists the ledger's coupons and corrections newest first, a page at a time
status: active
sourceFile: internal/ledger/postgres/recorded_test.go
verification_scope: integration
origin:
  kind: agent
  recorded_at: '2026-11-06T01:12:17.513Z'
id: TEST-ledger-pg-recorded-changes
type: test
---
Runs under task test:pg on a fresh database with the ledger migrations. Appends the log fixture, which holds corrections and a deletion among first versions, and a synthetic coupon, then checks RecordedChanges returns the coupon and the corrections newest first with their postings and record times, marks each as coupon or correction, leaves every other first version out, and continues after a given change and row ID.