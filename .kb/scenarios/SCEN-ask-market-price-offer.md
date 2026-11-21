---
title: The price tool is offered only where a connector runs and the route may receive T2
status: active
origin:
  kind: agent
  ref: D-226, the author's decision of 2026-11-21 (docs/DESIGN.md)
  recorded_at: '2026-11-21T12:00:00.000Z'
id: SCEN-ask-market-price-offer
type: scenario
---
Given an install whose settings say a market-data connector runs and a route whose models may receive T0 and T2, when a question reaches the model, then the tool is told with the others and its card is sent as T2; given a route that may not receive T2, then the tool is left out and the prompt says that prices need a model cleared for market data; given no connector, as on a demo, then the tool is left out and a call of it is an unknown tool; and if the route stops allowing T2 while the question runs, the card is withheld. Assumes the router answers the tier check.
