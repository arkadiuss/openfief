---
title: A clarifying answer of the Ask box offers only names that a card the model read holds, without regard to case; a name no card holds is dropped, and a clarifying answer left with no name is not an answer.
status: open
tags:
  - askbench
origin:
  kind: agent
  ref: 'Design session decision: the box owns a refusal''s command suggestions and vets clarify candidates (D-194), 2026-11-15'
  recorded_at: '2026-11-15T10:32:46.285Z'
semantic_text: A clarifying answer of the Ask box offers only names that a card the model read holds, without regard to case; a name no card holds is dropped, and a clarifying answer left with no name is not an answer.
logic_claims:
  - CLAIM-D0A087B8DD8D982C
  - CLAIM-E46980D52B174583
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: default
required_evidence: unverified
semantic_source_hash: 0076d78ca4ffef6c21f2bf1e7fdfb900aa7ba459dea55f380c8e9a3a8fffe577
semantic_inventory:
  - claim_key: CLAIM-D0A087B8DD8D982C
    claim_text: A clarifying answer of the Ask box offers only names that a card the model read holds, without regard to case
    role: descriptive
    status: ontology_gap
    span:
      start: 0
      end: 109
  - claim_key: CLAIM-E46980D52B174583
    claim_text: a name no card holds is dropped, and a clarifying answer left with no name is not an answer
    role: descriptive
    status: ontology_gap
    span:
      start: 111
      end: 202
id: REQ-ui-ask-clarify-names-come-from-cards
type: req
---
A clarifying answer of the Ask box offers only names that a card the model read holds, without regard to case; a name no card holds is dropped, and a clarifying answer left with no name is not an answer.

## Context

Candidates are chips that ask the box's own show command, so a name the model made up asks for something that does not exist. With the commands no longer the model's to write, the candidates are the one list left that it fills with names.

## Source

Design session decision, 2026-11-15: candidates stay the model's job, each validated against the shown cards' names, dropping unknown ones and rejecting a clarify left with none.
