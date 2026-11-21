---
title: A link that asks the model asks nothing
status: active
origin:
  kind: agent
  ref: 'Security review before the alpha.2 cut, known limit #9: a model question must need POST so a same-site GET can''t spend model budget (the coordinator''s follow-up, 2026-11-21)'
  recorded_at: '2026-11-21T19:04:45.798Z'
id: SCEN-ui-ask-model-question-needs-post
type: scenario
---
Given a signed-in person and a model on the Ask box's route, when their browser opens /?q=how+diversified+am+I%3F or GETs /ask with that question, then the page shows a card saying to ask it from the box and nothing is checked, counted or sent; when they type it in the box and press Ask, the box posts it and the model answers; 'allocation by currency' answers from a link as before.