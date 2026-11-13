---
title: A route that may not receive T0 is refused before anything is sent
status: active
origin:
  kind: agent
  ref: EV1 brief from the author's M2 corrections and D-208, 2026-11-13
  recorded_at: '2026-11-13T20:00:48.679Z'
id: SCEN-ui-ask-question-tagged-t0
type: scenario
---
Given a route whose models may receive public data only, when the owner asks a question beyond the commands, then the router refuses the conversation before sending anything, no tool runs, and the page says the question is T0 and nothing was sent, with the commands; given a route that may receive T0, the question goes tagged T0 with a system prompt tagged T2.