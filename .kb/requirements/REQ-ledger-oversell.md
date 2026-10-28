---
title: Disposing of more units than an account holds must fail with an error giving the requested and available quantities, and must consume no lot.
status: open
text_ref: docs/DESIGN.md#D-18
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-18
  - lane:strict
semantic_text: Disposing of more units than an account holds must fail with an error giving the requested and available quantities, and must consume no lot.
logic_claims:
  - CLAIM-E5B6D5236ADB1FB3
semantic_clauses:
  - Disposing of more units than an account holds must fail with an error giving the requested and available quantities, and must consume no lot.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 1957d3735ead383b741f65b5ac9e6e32f4e8b6093652ba1ee70d565b9993aebf
semantic_inventory:
  - claim_key: CLAIM-E5B6D5236ADB1FB3
    claim_text: Disposing of more units than an account holds must fail with an error giving the requested and available quantities, and must consume no lot
    role: normative
    status: modeled
    span:
      start: 0
      end: 140
origin:
  kind: agent
  recorded_at: '2026-10-28T05:39:29.506Z'
id: REQ-ledger-oversell
type: req
---
Disposing of more units than an account holds must fail with an error giving the requested and available quantities, and must consume no lot.

## Context

Short positions are out of scope for M0 (stocks, ETFs and bonds held long). A sale with no lots to cover it means a missing or wrong import, such as a buy before go-live without an opening lot; deriving lots anyway would produce wrong tax lots, so it is a typed error that names what to fix. Modeled by Claude Code for the M0 ledger task; the author has not reviewed it yet.

## Source

> Holdings and tax lots are derived from it.

docs/DESIGN.md D-18
