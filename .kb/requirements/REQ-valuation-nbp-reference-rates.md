---
title: The NBP archive adapter must return every change of the NBP reference rate with the day it took effect, as an exact fraction, in order, and fail on a document that is not NBP's interest rate archive.
status: open
text_ref: docs/DESIGN.md#D-9
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-9
  - lane:strict
semantic_text: The NBP archive adapter must return every change of the NBP reference rate with the day it took effect, as an exact fraction, in order, and fail on a document that is not NBP's interest rate archive.
logic_claims:
  - CLAIM-011869F3EC4626F0
semantic_clauses:
  - The NBP archive adapter must return every change of the NBP reference rate with the day it took effect, as an exact fraction, in order, and fail on a document that is not NBP's interest rate archive.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 05616d192c0829a4dbe75e9a6b9d8a4bf460196ea49892fa17ca4f7b9d212e22
semantic_inventory:
  - claim_key: CLAIM-011869F3EC4626F0
    claim_text: The NBP archive adapter must return every change of the NBP reference rate with the day it took effect, as an exact fraction, in order, and fail on a document that is not NBP's interest rate archive
    role: normative
    status: modeled
    span:
      start: 0
      end: 198
origin:
  kind: agent
  recorded_at: '2026-10-29T03:56:43.804Z'
id: REQ-valuation-nbp-reference-rates
type: req
---
The NBP archive adapter must return every change of the NBP reference rate with the day it took effect, as an exact fraction, in order, and fail on a document that is not NBP's interest rate archive.

## Context

DESIGN.md's context diagram has the treasury bonds connector fetch the NBP reference rate. NBP's Web API at api.nbp.pl serves no interest rates; NBP publishes them as XML on static.nbp.pl, which this reads. Modeled by Claude Code for the bond indices task; the author has not reviewed it yet.

## Source

> **[D-9]** M0 data paths: XTB through uploaded xStation export files; treasury bonds through manual lot entry (series, quantity, purchase date) with valuation computed; an import of the bond service's own export is added if one exists (OQ-22).

docs/DESIGN.md#D-9
