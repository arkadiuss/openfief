---
title: Time-weighted return, cash flows and transfer suggestions
status: active
sourceFile: internal/views/performance_test.go
origin:
  kind: agent
  recorded_at: '2026-10-28T20:54:27.083Z'
id: TEST-views-performance
type: test
verification_scope: unit
---
Table tests of Performance over synthetic deposits, withdrawals, buys and opening lots with stored valuations and USD rates: a price move counts, a deposit or withdrawal does not, a transfer the owner confirmed is money in transit at the day's rate in either booking order, an unconfirmed pair is two cash flows, gaps stop the series. SuggestTransfers pairs each withdrawal with the nearest deposit of the same amount and currency in another account within five days, skips taken and rejected pairs and opening cash; ConfirmedTransfers leaves only confirmed legs out of the cash flows and ignores links whose legs no longer match.