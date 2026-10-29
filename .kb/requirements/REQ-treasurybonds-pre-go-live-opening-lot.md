---
title: A treasury bond lot bought before go-live must be an opening lot at go-live, keeping its own purchase date and cost.
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-98-1
  - lane:strict
text_ref: design-decisions:D-98#1
origin:
  kind: import
  ref: design-decisions:D-98#1
  recorded_at: '2026-10-29T09:47:09.022Z'
semantic_text: A treasury bond lot bought before go-live must be an opening lot at go-live, keeping its own purchase date and cost.
logic_claims:
  - CLAIM-757EAF8221C22D26
semantic_clauses:
  - A treasury bond lot bought before go-live must be an opening lot at go-live, keeping its own purchase date and cost.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: 7804817282958d2a65335e6352a242804eb26e914943d1380a030cc191bc3170
semantic_inventory:
  - claim_key: CLAIM-757EAF8221C22D26
    claim_text: A treasury bond lot bought before go-live must be an opening lot at go-live, keeping its own purchase date and cost
    role: normative
    status: modeled
    span:
      start: 0
      end: 115
id: REQ-treasurybonds-pre-go-live-opening-lot
type: req
---
A treasury bond lot bought before go-live must be an opening lot at go-live, keeping its own purchase date and cost.

## Context

D-23 seeds accounts from current lot-level positions as opening lots, with transactions from go-live. D-98 applies that to treasury bonds entered by hand (D-9): a lot bought before go-live becomes an opening lot recorded at go-live that keeps its own purchase date and cost, while a lot bought after go-live is booked as it happened (REQ-treasurybonds-post-go-live-deposit-buy). REQ-treasurybonds-lot-mapping, which the author has not reviewed, maps every lot to a deposit and a buy on its purchase date, whether it was bought before or after go-live. The author recorded D-96 to D-99 after the design session that answered the implementation's questions.

## Source

> A treasury bond lot bought before go-live is an opening lot at go-live, keeping its own purchase date and cost (D-23). A manually entered account may receive its opening lots one at a time, and each is idempotent per series and purchase date. A lot bought after go-live is a deposit of its cost and a buy on its purchase date.

OpenFief author decisions (docs/DESIGN.md, section 3.2), design-decisions:D-98#1
