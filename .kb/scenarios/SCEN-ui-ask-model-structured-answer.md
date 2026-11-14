---
title: A model's final answer is checked, asked for once more and retried once, and a second failure is an error
status: active
origin:
  kind: agent
  ref: 'AB3 brief from the design session: exact Ask box scorers, a structured final answer, host pinning in the bench, 2026-11-14'
  recorded_at: '2026-11-14T14:52:56.927Z'
id: SCEN-ui-ask-model-structured-answer
type: scenario
---
Given a model that answers with a JSON object of kind, text, cards, suggestions and candidates, when the loop reads it, then a valid one ends the question, a reply that names a card it did not read, offers something that is no command of the box, has no text, no kind or a clarify without candidates is not an answer; given a model that writes plain text in a round that offered tools, then it is asked once more in a round without tools for the structured answer, and a reply of that round that is not valid gets one retry whose reason is fixed words, after which the question ends as a card that says it could not answer; every call that asks for the answer offers no tools, and all of it is within the question's rounds and output tokens.