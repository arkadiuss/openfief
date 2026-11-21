---
title: Given an Ask box with a model, when opinions are on for the owner, then the model's first message is the
status: active
origin:
  kind: agent
  ref: 'Implement D-225 and D-229: opinions and draft theses (design session brief, 2026-11-21)'
  recorded_at: '2026-11-21T15:00:00.000Z'
id: SCEN-ui-ask-opinions-mode
type: scenario
---
Given an Ask box with a model, when opinions are on for the owner, then the model's first message is the opinions prompt and its schema allows a thesis; when they are off, for a viewer or ineligible, it is the plain prompt, and an ineligible refusal says why; and every model answer's note says it is not investment advice.