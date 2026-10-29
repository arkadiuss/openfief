---
title: The price check threshold defaults to 1% and must be a fraction
status: active
origin:
  kind: agent
  recorded_at: '2026-10-29T05:37:13.895Z'
id: SCEN-valuation-price-check-threshold
type: scenario
---
Given a core started without a threshold, when price checks are read, then a difference beyond 0.01 is flagged; given a configured 0.003, then a 0.4% difference is flagged; given a threshold of 0, 1 or more, or a negative one, then the cross-check use cases refuse it with an invalid threshold error instead of flagging everything or nothing.