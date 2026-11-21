---
title: 'Only the Ask box''s own POST sends a question to the model: a question that is no command, asked over GET as the overview''s link or by htmx, gets a card that says to ask it from the box, and nothing is checked with Turnstile, admitted, counted or sent to a model; the commands still answer over GET.'
status: open
firmness: default
semantic_text: 'Only the Ask box''s own POST sends a question to the model: a question that is no command, asked over GET as the overview''s link or by htmx, gets a card that says to ask it from the box, and nothing is checked with Turnstile, admitted, counted or sent to a model; the commands still answer over GET.'
origin:
  kind: agent
  ref: 'Security review before the alpha.2 cut, known limit #9: a model question must need POST so a same-site GET can''t spend model budget (the coordinator''s follow-up, 2026-11-21)'
  recorded_at: '2026-11-21T19:04:28.239Z'
rationale: A same-site page could make a signed-in browser ask the model with a GET, spending its questions and the budget; only a same-origin POST may.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 9fa68ef9aac7c91260f2590b938aa3189985c040e610dfe4d283720af05290a8
semantic_inventory:
  - claim_key: CLAIM-AF8F5DC0933A6450
    claim_text: 'Only the Ask box''s own POST sends a question to the model: a question that is no command, asked over GET as the overview''s link or by htmx, gets a card that says to ask it from the box, and nothing is checked with Turnstile, admitted, counted or sent to a model'
    role: descriptive
    status: ontology_gap
    span:
      start: 0
      end: 261
    subjects:
      - ui.ask_model_question_method
    reason: It states which HTTP method may send a question to the model; no predicate of the ontology states a request method condition yet.
  - claim_key: CLAIM-B93BE2EB93ECE321
    claim_text: the commands still answer over GET
    role: descriptive
    status: ontology_gap
    span:
      start: 263
      end: 297
    subjects:
      - ui.ask_model_question_method
    reason: It states that the method condition leaves the commands alone; no predicate of the ontology states it yet.
logic_claims:
  - CLAIM-AF8F5DC0933A6450
  - CLAIM-B93BE2EB93ECE321
id: REQ-ui-ask-model-question-needs-post
type: req
---
Only the Ask box's own POST sends a question to the model: a question that is no command, asked over GET as the overview's link or by htmx, gets a card that says to ask it from the box, and nothing is checked with Turnstile, admitted, counted or sent to a model; the commands still answer over GET.

## Context

A same-site page could make a signed-in browser ask the model with a GET, spending its questions and the budget; only a same-origin POST may.

## Source

> Security review of the new surfaces before the alpha.2 cut: the multi-turn Ask box (2026-11-21); the coordinator asked to fix known limit #9.
