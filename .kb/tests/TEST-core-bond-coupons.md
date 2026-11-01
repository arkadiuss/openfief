---
title: Through the real core and PostgreSQL, a bond account's sync carries its lot and books its coupons once
status: active
sourceFile: cmd/core/coupons_test.go
verification_scope: integration
origin:
  kind: agent
  recorded_at: '2026-10-31T13:25:15.152Z'
id: TEST-core-bond-coupons
type: test
---
Runs the core's connector commands against a disposable PostgreSQL with the demo registry, seeds 10 COI1028 bonds as an opening lot with their terms, binds a fake treasury-bonds connector served over Connect and syncs twice. The connector decodes the request with delivery.Held and books coupons with treasurybonds.DueCoupons and BookCoupons. The test checks the request named the account, its go-live and its lot, that two interest transactions with Belka withheld are in the ledger, and that the second sync left them unchanged.