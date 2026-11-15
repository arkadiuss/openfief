---
title: The final round after an empty reply asks for the answer from the cards
status: active
origin:
  kind: agent
  ref: Design session approval of the empty-reply fix after bench traces (D-205, D-217), 2026-11-15
  recorded_at: '2026-11-15T12:26:14.635Z'
sourceFile: internal/views/ui/askfinal_test.go
verification_scope: unit
id: TEST-ui-ask-final-round-empty-reply
type: test
---
Runs the loop with a scripted model that calls a tool and then replies with no text or blanks, and checks that the final round holds the conversation so far, no empty assistant message and the answer instruction as the user's last message, with no restate wording.

## Context

Unit test of the final round's instruction.