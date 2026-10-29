---
title: IBKR's marks are stored beside the prices and never value a holding
status: active
origin:
  kind: agent
  recorded_at: '2026-10-29T05:37:11.760Z'
id: SCEN-valuation-cross-check-prices
type: scenario
---
Given stored eodhd prices of a synthetic ETF, when cross-check prices for it arrive from two IBKR instances, then both are stored under their own source, a repeated one replaces the earlier, the valuation prices are unchanged and no revaluation is enqueued; when one of a batch has no source or a close that isn't positive, then nothing of the batch is stored.