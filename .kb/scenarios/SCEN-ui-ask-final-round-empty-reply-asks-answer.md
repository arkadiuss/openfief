---
title: A model that stops after a tool result with an empty reply is asked to answer from the cards, not to restate a reply it never wrote
status: active
origin:
  kind: agent
  ref: Design session approval of the empty-reply fix after bench traces (D-205, D-217), 2026-11-15
  recorded_at: '2026-11-15T12:26:15.204Z'
id: SCEN-ui-ask-final-round-empty-reply-asks-answer
type: scenario
---
Given a model that calls a tool and then replies with no text or only blanks, when the loop asks for the structured answer, then the conversation holds no empty assistant message and the last message is the user's instruction to answer the question in the schema using only the cards above, not to restate a reply; and a bench question graded on that path passes when the model answers with the card.

## Context

The bench trace's common path: tool call, empty reply, final round.