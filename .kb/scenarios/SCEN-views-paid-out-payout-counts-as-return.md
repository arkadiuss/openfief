---
title: A paid-out coupon still counts as return
status: active
origin:
  kind: agent
  recorded_at: '2026-11-01T07:44:28.340Z'
id: SCEN-views-paid-out-payout-counts-as-return
type: scenario
---
Given an account whose holding rises from 1000 to 1100 PLN and that is paid a 50 PLN coupon withdrawn the same day, when the time-weighted return is computed, then the index is 115: the withdrawal is a cash flow, so the coupon counts as return.