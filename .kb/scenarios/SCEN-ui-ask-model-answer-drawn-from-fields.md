---
title: 'The page draws a model''s answer from its fields: the text, the cards it names, a clarify''s choices and a refusal''s commands'
status: active
origin:
  kind: agent
  ref: 'AB3 brief from the design session: exact Ask box scorers, a structured final answer, host pinning in the bench, 2026-11-14'
  recorded_at: '2026-11-14T14:52:57.745Z'
id: SCEN-ui-ask-model-answer-drawn-from-fields
type: scenario
---
Given an answer of kind answer, then the page shows its text and exactly the cards it names, in its order, and not a card it read and did not name; given kind clarify, then its candidates are choices that each ask the box's own show command, answered by the box at once; given kind refuse, then its suggestions are the box's command chips; and a candidate or suggestion is drawn only for its own kind, escaped.