---
title: Withdrawals are suggested with the nearest free deposit
status: active
origin:
  kind: agent
  recorded_at: '2026-10-30T06:20:11.497Z'
id: SCEN-views-transfer-suggestions
type: scenario
---
Given withdrawals and deposits across accounts, each withdrawal not already confirmed is suggested with the nearest deposit in date into another account of exactly the same amount and currency at most five days away, the earliest of equally near ones, not taken by a confirmed transfer or an earlier suggestion and not rejected with it; GET /v1/transfers lists suggestions, confirmed and rejected pairs with their legs' keys.