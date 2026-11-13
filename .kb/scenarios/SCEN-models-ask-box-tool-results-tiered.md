---
title: A tool result above what the route may receive is withheld and named
status: active
origin:
  kind: agent
  ref: M2 brief (the Ask box's model, D-203) on the model router, 2026-11-13
  recorded_at: '2026-11-13T18:11:28.670Z'
id: SCEN-models-ask-box-tool-results-tiered
type: scenario
---
Given a model allowed to receive public data only, when it calls a tool whose card is T0, then the router is asked about the conversation's tiers together, the model gets a refusal tagged public with none of the card's data, the page names the withheld card and its tier without drawing it, and a public card still goes through.