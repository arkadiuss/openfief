---
title: A hand-entered bond account must go live on the day its first lot is entered.
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-119-1
  - lane:strict
text_ref: design-decisions:D-119#1
origin:
  kind: import
  ref: design-decisions:D-119#1
  recorded_at: '2026-10-31T12:05:38.000Z'
semantic_text: A hand-entered bond account must go live on the day its first lot is entered.
logic_claims:
  - CLAIM-82F68F9F974DE457
semantic_clauses:
  - A hand-entered bond account must go live on the day its first lot is entered.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: 72c699a931d422ee5f18804ba607e828db46a52e4b406cd07dff0d9824827920
semantic_inventory:
  - claim_key: CLAIM-82F68F9F974DE457
    claim_text: A hand-entered bond account must go live on the day its first lot is entered
    role: normative
    status: modeled
    span:
      start: 0
      end: 76
id: REQ-ingest-bond-account-live-on-first-lot
type: req
---
A hand-entered bond account must go live on the day its first lot is entered.

## Context

D-98 books a bond lot bought before go-live as an opening lot at go-live and a later one as a deposit and a buy, but a hand-entered account has no broker statement whose date would be its go-live (question #18 on docs/implementation-questions.md). On 2026-10-31 the author decided it goes live on the day its first lot is entered, as a broker account does with its first statement (D-23), and that a lot bought on or before that day is an opening lot (REQ-treasurybonds-lot-on-go-live-day-opening).

## Source

> A hand-entered bond account goes live on the day its first lot is entered, as a broker account does with its first statement; a lot bought on or before that day is an opening lot (D-98).

OpenFief author decisions (docs/DESIGN.md, section 3.2), design-decisions:D-119#1
