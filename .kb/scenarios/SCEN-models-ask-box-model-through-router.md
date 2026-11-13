---
title: The Ask box's model call is the ask role's call for the asking session, audited with no text
status: active
origin:
  kind: agent
  ref: M2 brief (the Ask box's model, D-203) on the model router, 2026-11-13
  recorded_at: '2026-11-13T18:11:29.415Z'
id: SCEN-models-ask-box-model-through-router
type: scenario
---
Given a session of any scope, when the Ask box asks the router, then the call is for the ask role with the session's ID as actor and writes the sent and result entries with no question, answer or tool result; asking whether there is a model, or whether a tier may go, writes none and calls no model; a message without a tier is refused.