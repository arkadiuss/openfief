---
title: by the answer's kind, text, candidates and the box's own suggestions, the refusal R5 exactly (kind refuse, text, the box's commands) and nothing the model wrote in suggestions
status: open
firmness: default
required_evidence: unverified
semantic_text: by the answer's kind, text, candidates and the box's own suggestions, the refusal R5 exactly (kind refuse, text, the box's commands) and nothing the model wrote in suggestions
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: c1781a3d99c94b535833aa8740b2c8b3cc8cbfc94a972b64e0b628dfed67c088
semantic_inventory:
  - claim_key: CLAIM-A20E5858A006D0D8
    claim_text: by the answer's kind, text, candidates and the box's own suggestions, the refusal R5 exactly (kind refuse, text, the box's commands) and nothing the model wrote in suggestions
    role: descriptive
    status: ontology_gap
    span:
      start: 0
      end: 175
logic_claims:
  - CLAIM-A20E5858A006D0D8
origin:
  kind: agent
  ref: 'AB3 brief from the design session: exact Ask box scorers, a structured final answer, host pinning in the bench, 2026-11-14'
  recorded_at: '2026-11-14T14:53:10.059Z'
tags:
  - askbench
id: REQ-askbench-grades-by-the-structured-answer
type: req
---
The Ask box bench asks each model for the structured answer strictly or in JSON mode by its entry's flag, grades the refusals R1, R2 and R5 and the ambiguous name A18 by the answer's kind, text, candidates and the box's own suggestions, the refusal R5 exactly (kind refuse, text, the box's commands) and nothing the model wrote in suggestions, scores a reply that is no valid answer after its retry as a schema failure, holds the loop to the rounds that offer tools, the final round and its one retry, and puts the flag in the configuration hash.

## Context

The first runs failed refusals by their words (a Polish refusal, a refusal that says it can only answer questions about the portfolio) and A18 by whether it asked in words. With the answer as data these are decided exactly (2026-11-14). Changed 2026-11-15 by the design session (D-194): the box owns a refusal's suggestions, so the bench grades the box's and not the model's.

## Source

> The AB3 brief from the design session, 2026-11-14: (4) a structured final answer for the Ask box, with strict structured output where the model entry has the flag and JSON mode with one retry where it has not, validated cards and suggestions, the UI drawn from the fields, and the bench scoring by the schema where it decides.