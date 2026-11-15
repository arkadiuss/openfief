---
title: A model the eval hasn't covered is named as not evaluated.
status: open
tags: []
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-222#2
  recorded_at: '2026-11-15T09:51:18.780Z'
text_ref: design-decisions:D-222#2
semantic_text: A model the eval hasn't covered is named as not evaluated.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: d875955bac84293032d73b9343b0aeba82f8011f457c1b5b7ca8a7f8ac7bc0f4
semantic_inventory:
  - claim_key: CLAIM-EC7BADF52566806F
    claim_text: A model the eval hasn't covered is named as not evaluated
    role: descriptive
    status: missing
    span:
      start: 0
      end: 57
logic_claims:
  - CLAIM-EC7BADF52566806F
id: REQ-decision-d222-2-model-not-evaluated-is-named
type: req
---
A model the eval hasn't covered is named as not evaluated.

## Context

Author decision D-222: the Ask box says who answered, that it can be wrong and where to read how the model did in the eval.

## Source

> **[D-222]** When a model answers in the Ask box, the answer shows the model's display name from its settings entry, the line "AI can be wrong; check the cards. Not investment advice.", and a link to the docs page that reports how models did in the Ask box's eval (D-217): each model and host with its verdict, failing cases, reliability, cost and run dates, regenerated from the bench's reports. A model the eval hasn't covered is named as not evaluated. Answers to the box's own commands (D-194) show none of this.
