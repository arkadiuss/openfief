---
title: A tool result above what the route may receive is withheld and named
status: active
origin:
  kind: agent
  ref: EV1 brief from the author's M2 corrections and D-208, 2026-11-13
  recorded_at: '2026-11-13T20:00:51.812Z'
id: SCEN-models-ask-box-tool-results-tiered
type: scenario
---
Given a route whose model may receive T0 and public data but not one fief's data (TF), when it calls a tool whose card is TF, then the router is asked about the conversation's tiers together, the model gets a refusal tagged public with none of the card's data, the page names the withheld card and its tier without drawing it, and a T0 card still goes through. A route without T0 never reaches this: the question is T0 and is refused first.