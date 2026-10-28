---
title: An ETF holding is valued at the latest close and FX rate in PLN
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T07:28:43.678Z'
id: SCEN-valuation-market-holding
type: scenario
---
Given 12.5 units of an ETF, a 101.37 USD close on Friday and a 4.0123 PLN per USD rate on Friday, when the holding is valued on Sunday, then its value is exactly 1267.125 USD and 5084.09 PLN rounded half up, with Friday as the price and rate dates; given a close in PLN, then no rate is needed.