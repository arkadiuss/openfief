---
title: Valuation totals include each account's cash, valued in its wrapper
status: active
origin:
  kind: agent
  recorded_at: '2026-10-30T00:04:52.668Z'
id: SCEN-views-totals-with-cash
type: scenario
---
Given an account with 100 EUR of cash in IKZE and an EUR rate of 4.30, when its totals are read, then the day's gross total includes 430 PLN, its net 387, and cash in a currency with no rate counts only per currency.