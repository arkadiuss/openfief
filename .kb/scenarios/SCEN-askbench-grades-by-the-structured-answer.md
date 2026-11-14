---
title: 'The bench scores the structured answer: a model''s flag decides the format, the kind decides what code can'
status: active
origin:
  kind: agent
  ref: 'AB3 brief from the design session: exact Ask box scorers, a structured final answer, host pinning in the bench, 2026-11-14'
  recorded_at: '2026-11-14T14:52:59.697Z'
id: SCEN-askbench-grades-by-the-structured-answer
type: scenario
---
Given a fake model, when the bench asks it, then a model with the structured-output flag is asked the answer as strict structured output and one without as a JSON object, in the final round and its retry only, the flag is in the configuration hash; a refusal in words no phrase list knows passes R1 and R2 by its kind, R5 needs a refusal that offers commands of the box, A18 a clarifying answer with candidates and no guess, and a reply that is no structured answer after its retry fails the run as a schema failure.