---
title: A cross-check price in a different currency from the stored price must be shown as a pair with its own status instead of being skipped.
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-111-2
  - lane:strict
text_ref: design-decisions:D-111#2
origin:
  kind: import
  ref: design-decisions:D-111#2
  recorded_at: '2026-10-29T10:54:08.412Z'
semantic_text: A cross-check price in a different currency from the stored price must be shown as a pair with its own status instead of being skipped.
logic_claims:
  - CLAIM-ED654A91A5D07C2A
semantic_clauses:
  - A cross-check price in a different currency from the stored price must be shown as a pair with its own status instead of being skipped.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: ee891b2c4733faf887e75df79a06592a4969c0b2ac90080390dc4907af6e8570
semantic_inventory:
  - claim_key: CLAIM-ED654A91A5D07C2A
    claim_text: A cross-check price in a different currency from the stored price must be shown as a pair with its own status instead of being skipped
    role: normative
    status: modeled
    span:
      start: 0
      end: 134
id: REQ-valuation-price-check-currency-mismatch
type: req
---
A cross-check price in a different currency from the stored price must be shown as a pair with its own status instead of being skipped.

## Context

D-111 compares a cross-check price with the latest stored price of the same instrument (REQ-valuation-price-check-latest-price). A pair in different currencies gets its own status because the market-data connector may be pricing another listing than the one held. The author recorded D-109 to D-112 after the design session that answered the implementation's questions.

## Source

> A cross-check price is compared with the latest stored price of the same instrument on or before its date. When the two are in different currencies, the pair is shown with its own status instead of being skipped, because the market-data connector may be pricing another listing than the one held.

OpenFief author decisions (docs/DESIGN.md, section 3.2), design-decisions:D-111#2
