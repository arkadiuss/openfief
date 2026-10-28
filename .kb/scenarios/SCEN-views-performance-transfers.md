---
title: A matched withdrawal and deposit between own accounts is a transfer, not a cash flow
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T20:54:39.162Z'
id: SCEN-views-performance-transfers
type: scenario
---
Given a withdrawal from one account and a deposit of the same amount and currency into another at most five days apart, in either order, when the series is computed, then they are one transfer, the money counts as in transit at each day's rate until both legs are booked, and neither is a cash flow; further apart, in the same account, of another amount or currency, or with opening cash, they stay cash flows.