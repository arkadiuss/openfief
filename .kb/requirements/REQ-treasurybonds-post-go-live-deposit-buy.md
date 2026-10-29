---
title: A treasury bond lot bought after go-live must be a deposit of its cost and a buy on its purchase date.
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-98-3
  - lane:strict
text_ref: design-decisions:D-98#3
origin:
  kind: import
  ref: design-decisions:D-98#3
  recorded_at: '2026-10-29T09:48:32.218Z'
semantic_text: A treasury bond lot bought after go-live must be a deposit of its cost and a buy on its purchase date.
logic_claims:
  - CLAIM-E13DF037C41763FB
semantic_clauses:
  - A treasury bond lot bought after go-live must be a deposit of its cost and a buy on its purchase date.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: b1496342a7d0f429a9fa661a1ad9407f0f6433563d2e0fd1495c2f5e8c7d902c
semantic_inventory:
  - claim_key: CLAIM-E13DF037C41763FB
    claim_text: A treasury bond lot bought after go-live must be a deposit of its cost and a buy on its purchase date
    role: normative
    status: modeled
    span:
      start: 0
      end: 101
id: REQ-treasurybonds-post-go-live-deposit-buy
type: req
---
A treasury bond lot bought after go-live must be a deposit of its cost and a buy on its purchase date.

## Context

D-23 takes transactions from go-live, so D-98 books a treasury bond lot bought after go-live as transactions: a deposit of its cost and a buy on its purchase date. A lot bought before go-live is an opening lot instead (REQ-treasurybonds-pre-go-live-opening-lot). REQ-treasurybonds-lot-mapping, which the author has not reviewed, already books lots this way, but for every purchase date. The author recorded D-96 to D-99 after the design session that answered the implementation's questions.

## Source

> A treasury bond lot bought before go-live is an opening lot at go-live, keeping its own purchase date and cost (D-23). A manually entered account may receive its opening lots one at a time, and each is idempotent per series and purchase date. A lot bought after go-live is a deposit of its cost and a buy on its purchase date.

OpenFief author decisions (docs/DESIGN.md, section 3.2), design-decisions:D-98#3
