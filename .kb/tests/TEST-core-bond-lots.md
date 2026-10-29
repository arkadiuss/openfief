---
title: The core takes an owner's bond lot, hands it to the bound connector and values the lot by the bond rules
status: active
sourceFile: cmd/core/bondlots_test.go
origin:
  kind: agent
  recorded_at: '2026-10-29T01:54:40.118Z'
id: TEST-core-bond-lots
type: test
---
Runs the real core on a disposable PostgreSQL (task test:pg) with a passkey owner session and the demo registry: before any connector client exists AddBondLot is UNAVAILABLE; with a connector that maps lots as the treasury-bonds plugin does, a synthetic 40-bond EDO0536 lot imports two new transactions, entering it again leaves both unchanged, an unknown session is UNAUTHENTICATED and an account without a bond connector FAILED_PRECONDITION; then the River workers value the lot at the hand-computed 4042.12 PLN net if redeemed on 2026-10-28. A false pass would need valuation to price the bond at market, which has no price.