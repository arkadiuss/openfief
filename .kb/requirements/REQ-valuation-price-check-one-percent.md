---
title: A cross-check price must be flagged when it differs from the stored price by more than 1% of it, in either direction.
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-110-1
  - lane:strict
text_ref: design-decisions:D-110#1
origin:
  kind: import
  ref: design-decisions:D-110#1
  recorded_at: '2026-10-29T10:51:40.179Z'
semantic_text: A cross-check price must be flagged when it differs from the stored price by more than 1% of it, in either direction.
logic_claims:
  - CLAIM-DB5F1E85CDE3405B
semantic_clauses:
  - A cross-check price must be flagged when it differs from the stored price by more than 1% of it, in either direction.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: 43ecb84c49c77bf91a2e23986ee59b8f4a194c17edef568fff6ed48e3ddf56af
semantic_inventory:
  - claim_key: CLAIM-DB5F1E85CDE3405B
    claim_text: A cross-check price must be flagged when it differs from the stored price by more than 1% of it, in either direction
    role: normative
    status: modeled
    span:
      start: 0
      end: 116
id: REQ-valuation-price-check-one-percent
type: req
---
A cross-check price must be flagged when it differs from the stored price by more than 1% of it, in either direction.

## Context

D-22 makes IBKR Flex prices a cross-check of the market-data prices, and D-94 keeps every source's values so the lower-ranked ones serve as cross-checks. D-110 sets when a cross-check price is flagged; REQ-valuation-price-check-instrument-threshold lets the owner change the threshold for one instrument. The author recorded D-109 to D-112 after the design session that answered the implementation's questions.

## Source

> A cross-check price is flagged when it differs from the stored price by more than 1% of it, in either direction. The owner can set a different threshold for a single instrument, such as a thinly traded Warsaw listing.

OpenFief author decisions (docs/DESIGN.md, section 3.2), design-decisions:D-110#1
