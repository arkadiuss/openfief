---
title: A command is answered by the Ask box itself and a question beyond the commands goes to the model
status: active
origin:
  kind: agent
  ref: M2 brief (the Ask box's model, D-203) on the model router, 2026-11-13
  recorded_at: '2026-11-13T18:11:26.175Z'
id: SCEN-ui-ask-model-fixed-commands-first
type: scenario
---
Given a model is configured, when the owner asks what needs me, an allocation or a value, then the box answers at once from the read API and the router is not asked; when the owner asks a question that is no command, then it goes to the model, whose answer is shown with the model's name.