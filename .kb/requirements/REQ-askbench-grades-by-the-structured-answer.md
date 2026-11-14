---
title: The Ask box bench asks each model for the structured answer strictly or in JSON mode by its entry's flag, grades the refusals R1, R2 and R5 and the ambiguous name A18 by the answer's kind, text, suggestions and candidates, scores a reply that is no valid answer after its retry as a schema failure, holds the loop to the rounds that offer tools, the final round and its one retry, and puts the flag in the configuration hash.
status: open
firmness: default
required_evidence: unverified
semantic_text: The Ask box bench asks each model for the structured answer strictly or in JSON mode by its entry's flag, grades the refusals R1, R2 and R5 and the ambiguous name A18 by the answer's kind, text, suggestions and candidates, scores a reply that is no valid answer after its retry as a schema failure, holds the loop to the rounds that offer tools, the final round and its one retry, and puts the flag in the configuration hash.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: cbab43bd8af276336bf1b6f3d718578a103af083e22c4675baab5ff0644ee10c
semantic_inventory:
  - claim_key: CLAIM-B96D82AB54EDF957
    claim_text: The Ask box bench asks each model for the structured answer strictly or in JSON mode by its entry's flag, grades the refusals R1, R2 and R5 and the ambiguous name A18 by the answer's kind, text, suggestions and candidates, scores a reply that is no valid answer after its retry as a schema failure, holds the loop to the rounds that offer tools, the final round and its one retry, and puts the flag in the configuration hash
    role: descriptive
    status: missing
    span:
      start: 0
      end: 424
logic_claims:
  - CLAIM-B96D82AB54EDF957
origin:
  kind: agent
  ref: 'AB3 brief from the design session: exact Ask box scorers, a structured final answer, host pinning in the bench, 2026-11-14'
  recorded_at: '2026-11-14T14:53:10.059Z'
tags:
  - askbench
id: REQ-askbench-grades-by-the-structured-answer
type: req
---
The Ask box bench asks each model for the structured answer strictly or in JSON mode by its entry's flag, grades the refusals R1, R2 and R5 and the ambiguous name A18 by the answer's kind, text, suggestions and candidates, scores a reply that is no valid answer after its retry as a schema failure, holds the loop to the rounds that offer tools, the final round and its one retry, and puts the flag in the configuration hash.

## Context

The first runs failed refusals by their words (a Polish refusal, a refusal that says it can only answer questions about the portfolio) and A18 by whether it asked in words. With the answer as data these are decided exactly (2026-11-14).

## Source

> The AB3 brief from the design session, 2026-11-14: (4) a structured final answer for the Ask box, with strict structured output where the model entry has the flag and JSON mode with one retry where it has not, validated cards and suggestions, the UI drawn from the fields, and the bench scoring by the schema where it decides.