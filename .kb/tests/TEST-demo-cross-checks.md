---
title: The demo dataset's IBKR marks give exactly one flagged price check
status: active
sourceFile: internal/demo/crosscheck_test.go
origin:
  kind: agent
  recorded_at: '2026-10-30T05:17:48.723Z'
id: TEST-demo-cross-checks
type: test
verification_scope: unit
---
Generates the demo dataset for the test seed and day, compares its synthetic IBKR mark prices with its prices through valuation.CheckPrices at the default 1%, and checks that every mark is compared, each from the ibkr source against the demo's own price, and that exactly one check, on the latest statement's day, is flagged.