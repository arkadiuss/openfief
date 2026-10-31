---
title: A lot bought on or before a hand-entered bond account's go-live day must be an opening lot.
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-119-2
  - lane:strict
text_ref: design-decisions:D-119#2
origin:
  kind: import
  ref: design-decisions:D-119#2
  recorded_at: '2026-10-31T12:05:38.000Z'
semantic_text: A lot bought on or before a hand-entered bond account's go-live day must be an opening lot.
logic_claims:
  - CLAIM-93D004B3AFD7B40E
semantic_clauses:
  - A lot bought on or before a hand-entered bond account's go-live day must be an opening lot.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: 280ccd0aea19d4f7e90a0414d35e2bdbefe4d097f0f674e57a6dc955dded6572
semantic_inventory:
  - claim_key: CLAIM-93D004B3AFD7B40E
    claim_text: A lot bought on or before a hand-entered bond account's go-live day must be an opening lot
    role: normative
    status: modeled
    span:
      start: 0
      end: 90
id: REQ-treasurybonds-lot-on-go-live-day-opening
type: req
---
A lot bought on or before a hand-entered bond account's go-live day must be an opening lot.

## Context

D-98 makes a lot bought before go-live an opening lot (REQ-treasurybonds-pre-go-live-opening-lot) and one bought after it a deposit and a buy (REQ-treasurybonds-post-go-live-deposit-buy), but didn't say how a lot bought on the go-live day itself is booked (question #18 on docs/implementation-questions.md). With a hand-entered account going live on the day its first lot is entered (REQ-ingest-bond-account-live-on-first-lot), the author decided on 2026-10-31 that a lot bought on that day is an opening lot too, as the implementation already did.

## Source

> A hand-entered bond account goes live on the day its first lot is entered, as a broker account does with its first statement; a lot bought on or before that day is an opening lot (D-98).

OpenFief author decisions (docs/DESIGN.md, section 3.2), design-decisions:D-119#2
