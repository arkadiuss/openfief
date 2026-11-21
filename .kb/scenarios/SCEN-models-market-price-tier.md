---
title: A price needs T2 and the box's own text needs no tier
status: active
origin:
  kind: agent
  ref: D-226, the author's decision of 2026-11-21 (docs/DESIGN.md), and the design session's ruling on it
  recorded_at: '2026-11-21T14:00:00.000Z'
id: SCEN-models-market-price-tier
type: scenario
---
Given a route whose models are allowed T0 only, when the Ask box asks about the portfolio's value, then it is answered with the question and cards as T0 and the fixed text tagged instructions adding no tier, and a price is declined with "prices need a model cleared for market data (T2)"; given a route allowed T0 and T2, the model gets the price tool and a price is T2; a request of nothing but instructions has no tier and is refused as untagged.
