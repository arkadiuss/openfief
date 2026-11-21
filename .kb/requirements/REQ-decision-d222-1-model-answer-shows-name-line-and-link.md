---
title: 'When a model answers in the Ask box, the answer shows the model''s display name from its settings entry, the line "AI can be wrong; check the cards. Not investment advice.", and a link to the docs page that reports how models did in the Ask box''s eval (D-217): each model and host with its verdict, failing cases, reliability, cost and run dates, regenerated from the bench''s reports.'
status: open
tags: []
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-222#1
  recorded_at: '2026-11-15T09:51:16.055Z'
text_ref: design-decisions:D-222#1
semantic_text: 'When a model answers in the Ask box, the answer shows the model''s display name from its settings entry, the line "AI can be wrong; check the cards. Not investment advice.", and a link to the docs page that reports how models did in the Ask box''s eval (D-217): each model and host with its verdict, failing cases, reliability, cost and run dates, regenerated from the bench''s reports.'
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 557f2fab83d0ddfe009df04004204599fe65908bed11b27d9eb72e3f697851fc
semantic_inventory:
  - claim_key: CLAIM-F9BBDE701438CEA4
    claim_text: When a model answers in the Ask box, the answer shows the model's display name from its settings entry, the line "AI can be wrong
    role: condition
    status: ontology_gap
    span:
      start: 0
      end: 129
    subjects:
      - ask_box.answer_foot_note
  - claim_key: CLAIM-94D67B578313D72A
    claim_text: check the cards
    role: descriptive
    status: missing
    span:
      start: 131
      end: 146
  - claim_key: CLAIM-042A565E60D43F42
    claim_text: 'Not investment advice.", and a link to the docs page that reports how models did in the Ask box''s eval (D-217): each model and host with its verdict, failing cases, reliability, cost and run dates, regenerated from the bench''s reports'
    role: descriptive
    status: missing
    span:
      start: 148
      end: 382
logic_claims:
  - CLAIM-F9BBDE701438CEA4
  - CLAIM-94D67B578313D72A
  - CLAIM-042A565E60D43F42
id: REQ-decision-d222-1-model-answer-shows-name-line-and-link
type: req
pair_reviews:
  - with: REQ-ui-ask-shows-model-and-disclaimer
    verdict: no_conflict
    separated_by: what the note must state (name from the settings entry, the disclaimer line, the docs link, not evaluated; D-222) against where it sits and how many cards it counts (the UI requirement); the UI requirement restates D-222's single note
    claims:
      - claim_key: CLAIM-042A565E60D43F42
      - claim_key: CLAIM-07955DF152B6858B
      - claim_key: CLAIM-94D67B578313D72A
      - claim_key: CLAIM-F9BBDE701438CEA4
    origin:
      kind: agent
      ref: 'The author approved the subjects and every pair verdict of the kibi 2.45 review: "Approve all, please continue", chat with the orchestrator, 2026-11-21'
      approved_by: author, 2026-11-21
---
When a model answers in the Ask box, the answer shows the model's display name from its settings entry, the line "AI can be wrong; check the cards. Not investment advice.", and a link to the docs page that reports how models did in the Ask box's eval (D-217): each model and host with its verdict, failing cases, reliability, cost and run dates, regenerated from the bench's reports.

## Context

Author decision D-222: the Ask box says who answered, that it can be wrong and where to read how the model did in the eval.

## Source

> **[D-222]** When a model answers in the Ask box, the answer shows the model's display name from its settings entry, the line "AI can be wrong; check the cards. Not investment advice.", and a link to the docs page that reports how models did in the Ask box's eval (D-217): each model and host with its verdict, failing cases, reliability, cost and run dates, regenerated from the bench's reports. A model the eval hasn't covered is named as not evaluated. Answers to the box's own commands (D-194) show none of this.
