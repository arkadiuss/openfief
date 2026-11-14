---
title: The model's tools read as the asking person and the page shows the cards the answer names
status: active
origin:
  kind: agent
  ref: 'AB3 brief from the design session: exact Ask box scorers, a structured final answer, host pinning in the bench, 2026-11-14'
  recorded_at: '2026-11-14T14:53:02.584Z'
id: SCEN-ui-ask-model-read-only-tools
type: scenario
---
Given a model that asks for the allocation and the value as tools, then each tool reads the read API with the asking person's own session token, the model gets each card as JSON with an id, the page shows its text and exactly the cards its answer names, and a call of an unknown tool, with arguments the schema doesn't allow or beyond the loop's limits gets a failure and runs nothing.