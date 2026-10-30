---
title: An acknowledgement must cover a price check only while its cross-check price and the price it is compared with keep their dates and closes, and only a flagged check may be acknowledged.
status: open
text_ref: docs/DESIGN.md#D-112
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-112
  - lane:strict
semantic_text: An acknowledgement must cover a price check only while its cross-check price and the price it is compared with keep their dates and closes, and only a flagged check may be acknowledged.
logic_claims:
  - CLAIM-469C8F886244B8F2
semantic_clauses:
  - An acknowledgement must cover a price check only while its cross-check price and the price it is compared with keep their dates and closes, and only a flagged check may be acknowledged.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: d8654800df2fc42aef6649161b5d5746e275c69eac220749b5e175c72904d319
semantic_inventory:
  - claim_key: CLAIM-469C8F886244B8F2
    claim_text: An acknowledgement must cover a price check only while its cross-check price and the price it is compared with keep their dates and closes, and only a flagged check may be acknowledged
    role: normative
    status: modeled
    span:
      start: 0
      end: 184
origin:
  kind: agent
  recorded_at: '2026-10-30T02:58:49.715Z'
id: REQ-valuation-price-check-acknowledgement-lapses
type: req
---
An acknowledgement must cover a price check only while its cross-check price and the price it is compared with keep their dates and closes, and only a flagged check may be acknowledged.

## Context

D-112 lets the owner acknowledge a flagged check, recorded append-only, without saying what an acknowledgement covers. The narrowest reading: the check as seen, so a changed price is flagged anew.

## Source

> Only a connector whose manifest declares it may send cross-check prices. The owner can acknowledge a flagged check, which is recorded append-only, and the UI shows the latest check per instrument with its history.

docs/DESIGN.md#D-112
