---
title: Caps, trimming, framing and summaries of the earlier turns are table tests; the model gets a transcript in the question message
status: active
origin:
  kind: agent
  ref: 'Implement D-224: the Ask box holds a conversation (design session brief, 2026-11-21); the author''s feedback that the Ask button must sit inline with the field'
  recorded_at: '2026-11-21T09:22:39.000Z'
sourceFile: internal/views/ui/askhistory_test.go
verification_scope: unit
id: TEST-ui-ask-history
type: test
---
Table tests of capHistory (turn cap, character cap with the oldest dropped first, field caps, empty turns, line breaks and control characters, idempotence), parseHistory (not JSON, not an array, trailing text, wrong types, oversize), turnOf and turnData (a command's card, a model's answer with the cards it showed, the cards that add no turn), historyMessage (no turns is the question alone; a JSON transcript with the untrusted preamble; no line imitating another message), and the loop: one user message in tier T0 holds the transcript, no turn is an assistant or tool message, a card the transcript named is no card of the answer, and the history of a request is capped again wherever it came from.