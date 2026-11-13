---
title: The model calls the commands as tools with the asker's session and the page shows exactly the cards it read
status: active
origin:
  kind: agent
  ref: M2 brief (the Ask box's model, D-203) on the model router, 2026-11-13
  recorded_at: '2026-11-13T18:11:26.982Z'
id: SCEN-ui-ask-model-read-only-tools
type: scenario
---
Given a model that asks for the allocation and the value as tools, then each tool reads the read API with the asking person's own session token, the model gets each card as JSON, the page shows its text and exactly those two cards, and a call of an unknown tool, with arguments the schema doesn't allow or beyond the loop's limits gets a failure and runs nothing.