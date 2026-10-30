---
title: A price check's difference is rounded for display only
status: active
origin:
  kind: agent
  recorded_at: '2026-10-30T02:58:30.851Z'
id: SCEN-valuation-price-check-rounding
type: scenario
---
Given a close of 3 USD and a mark of 4 USD, when they are compared at a 50% threshold, then the relative difference reads 0.333333, rounded half-even to six places; and given a close of 100 USD and a mark of 101 USD at 1%, the check is matched because the exact difference equals the threshold, while 101.0001 USD is flagged.