---
title: 'The Ask box draws a model''s answer from its structured fields and nothing else the model wrote: its text, exactly the cards it names in the order it names them, a clarifying answer''s candidates as choices that ask the box''s own show command, and a refusing answer''s suggestions as the box''s command chips.'
status: open
firmness: default
required_evidence: unverified
semantic_text: 'The Ask box draws a model''s answer from its structured fields and nothing else the model wrote: its text, exactly the cards it names in the order it names them, a clarifying answer''s candidates as choices that ask the box''s own show command, and a refusing answer''s suggestions as the box''s command chips.'
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 730fc3ac745200df33dd9c8e92a6a11fa8acc35271128d2ac8b17d4299a0266c
semantic_inventory:
  - claim_key: CLAIM-D85E46758D9FA6F6
    claim_text: 'The Ask box draws a model''s answer from its structured fields and nothing else the model wrote: its text, exactly the cards it names in the order it names them, a clarifying answer''s candidates as choices that ask the box''s own show command, and a refusing answer''s suggestions as the box''s command chips'
    role: descriptive
    status: missing
    span:
      start: 0
      end: 304
logic_claims:
  - CLAIM-D85E46758D9FA6F6
origin:
  kind: agent
  ref: 'AB3 brief from the design session: exact Ask box scorers, a structured final answer, host pinning in the bench, 2026-11-14'
  recorded_at: '2026-11-14T14:53:06.990Z'
tags:
  - askbench
id: REQ-ui-ask-model-answer-drawn-from-fields
type: req
---
The Ask box draws a model's answer from its structured fields and nothing else the model wrote: its text, exactly the cards it names in the order it names them, a clarifying answer's candidates as choices that ask the box's own show command, and a refusing answer's suggestions as the box's command chips.

## Context

D-203 has the model show the answer cards it used. With the answer as data the page can show exactly those, offer the choices of a name that is several things, and offer the box's commands when the model declines, each of which the box answers itself, at once.

## Source

> The AB3 brief from the design session, 2026-11-14: (4) a structured final answer for the Ask box, with strict structured output where the model entry has the flag and JSON mode with one retry where it has not, validated cards and suggestions, the UI drawn from the fields, and the bench scoring by the schema where it decides.