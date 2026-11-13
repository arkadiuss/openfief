---
title: With no model the Ask box answers as it did before there was one
status: active
origin:
  kind: agent
  ref: M2 brief (the Ask box's model, D-203) on the model router, 2026-11-13
  recorded_at: '2026-11-13T18:11:27.839Z'
id: SCEN-ui-ask-model-fallback-as-before
type: scenario
---
Given the switch off, the Ask box's toggle off, no route or a core that can't say, when the owner asks a question that is no command, then the box answers I can't answer that yet with the commands it understands and the model is not asked; a model that can't answer gets a card with the commands and a code in the log.