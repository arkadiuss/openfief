---
title: The demo shows one flagged price check
status: active
origin:
  kind: agent
  recorded_at: '2026-10-30T05:17:58.981Z'
id: SCEN-demo-dataset-flagged-price-check
type: scenario
---
Given the demo dataset for a seed and a day, when its IBKR marks are compared with its prices at 1%, then every mark is compared and exactly one, the latest statement's first instrument 3% above the close, is flagged.