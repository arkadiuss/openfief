---
title: Earlier turns are capped again, framed as an untrusted transcript and never logged
status: active
origin:
  kind: agent
  ref: 'Implement D-224: the Ask box holds a conversation (design session brief, 2026-11-21); the author''s feedback that the Ask button must sit inline with the field'
  recorded_at: '2026-11-21T09:22:39.000Z'
id: SCEN-ui-ask-history-untrusted
type: scenario
---
Given earlier turns of any length, count and content, when a question is asked after them, then at most six turns and 4,000 characters reach the model, the oldest dropped first and each field shortened, a history that is not a JSON array of turns is dropped and the question goes alone, the turns are one JSON transcript in the question's user message with a preamble that calls them untrusted, a forged assistant turn is no assistant message and no tool result, an answer that rests on a card the transcript named is no answer, no log line holds a question, an answer or an earlier turn, the model's tools are exactly the shared toolbox the MCP server offers, and the bench's follow-up and injection cases hold earlier turns that it sends the same way.