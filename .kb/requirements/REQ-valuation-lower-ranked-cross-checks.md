---
title: A lower-ranked source's price must be compared with the winning price of its instrument and date like a cross-check price, and a cross-check price must never be ranked or win.
status: closed
text_ref: docs/DESIGN.md#D-94
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-94
  - lane:strict
semantic_text: A lower-ranked source's price must be compared with the winning price of its instrument and date like a cross-check price, and a cross-check price must never be ranked or win.
logic_claims:
  - CLAIM-C7E56B276B0BA899
semantic_clauses:
  - A lower-ranked source's price must be compared with the winning price of its instrument and date like a cross-check price, and a cross-check price must never be ranked or win.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 3fe9881365037ac24588e1f6ed9d79ca4f73493e811d1b91c746c1665fe4c404
semantic_inventory:
  - claim_key: CLAIM-C7E56B276B0BA899
    claim_text: A lower-ranked source's price must be compared with the winning price of its instrument and date like a cross-check price, and a cross-check price must never be ranked or win
    role: normative
    status: modeled
    span:
      start: 0
      end: 174
origin:
  kind: agent
  recorded_at: '2026-10-29T18:01:42.609Z'
id: REQ-valuation-lower-ranked-cross-checks
type: req
---
A lower-ranked source's price must be compared with the winning price of its instrument and date like a cross-check price, and a cross-check price must never be ranked or win.

## Context

An earlier refinement of D-94 by Claude Code (default firmness).

## Why this changed

D-94 was rewritten: each series has one source, so there is no winner per date; every other source's price checks the series source's latest price on or before its date. Superseded by REQ-valuation-other-source-price-checks.

## Source

docs/DESIGN.md#D-94
