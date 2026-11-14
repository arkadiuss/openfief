---
title: 'The structured answer: validation, the loop''s one retry and final round, and the page''s three kinds'
status: active
sourceFile: internal/views/ui/askstructured_test.go
verification_scope: unit
origin:
  kind: agent
  ref: 'AB3 brief from the design session: exact Ask box scorers, a structured final answer, host pinning in the bench, 2026-11-14'
  recorded_at: '2026-11-14T14:52:55.374Z'
id: TEST-ui-ask-structured
type: test
---
Parses structured answers against the schema, kind, text, card ids, suggestions, candidates and lists too long; keeps the JSON Schema and the struct one definition; runs the loop with a scripted model that answers at once in a round with tools, writes plain text there, writes an invalid answer in the final round and mends it on the retry, or fails twice, and checks the calls that ask for the answer offer no tools, the retry's reason is in fixed words, the retry is spent from the question's tokens, the cards shown are those the answer names in its order and each kind draws only its own fields; and renders the three kinds through the page over a fake core.
