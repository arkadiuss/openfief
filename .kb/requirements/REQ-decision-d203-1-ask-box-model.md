---
title: With a model configured, the Ask box must still answer its fixed commands itself, at once, send anything else to the model, which may only call the same read-only answerers as tools and must show the answer cards it used, never change anything, and never log the question and answer.
status: open
semantic_text: With a model configured, the Ask box must still answer its fixed commands itself, at once, send anything else to the model, which may only call the same read-only answerers as tools and must show the answer cards it used, never change anything, and never log the question and answer.
origin:
  kind: import
  ref: design-decisions:D-203#1
  recorded_at: '2026-11-13T18:11:36.163Z'
firmness: firm
text_ref: design-decisions:D-203#1
required_evidence: asserted
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: bb8706bd250ed2bbd545fa2ef12231fb41c175ab7eb71482125238d1ee7704f9
semantic_inventory:
  - claim_key: CLAIM-DC3ADD5D76F483D9
    claim_text: With a model configured, the Ask box must still answer its fixed commands itself, at once, send anything else to the model, which may only call the same read-only answerers as tools and must show the answer cards it used, never change anything, and never log the question and answer
    role: normative
    status: ontology_gap
    span:
      start: 0
      end: 282
logic_claims:
  - CLAIM-DC3ADD5D76F483D9
id: REQ-decision-d203-1-ask-box-model
type: req
---
With a model configured, the Ask box must still answer its fixed commands itself, at once, send anything else to the model, which may only call the same read-only answerers as tools and must show the answer cards it used, never change anything, and never log the question and answer.

## Context

Author decision D-203 (the read-only models slice, D-202).

## Source

> **[D-203]** With a model configured, the Ask box still answers its fixed commands itself, at once; anything else goes to the model, which may only call the same read-only answerers as tools and must show the answer cards it used. It never changes anything, and the question and answer are never logged (GR-12).

docs/DESIGN.md#D-203