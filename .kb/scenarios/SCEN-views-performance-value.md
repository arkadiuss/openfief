---
title: The portfolio's daily value counts holdings, cash and money in transit
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T20:54:36.854Z'
id: SCEN-views-performance-value
type: scenario
---
Given holdings with stored valuations, cash in PLN and USD and an opening lot of an account seeded after go-live, when the day is valued, then its value is the holdings' valuations plus the cash converted at the latest FX rate on or before the day, rounded half up to grosze, plus money in transit, and the opening lot flows in at its share of its holding's valuation.