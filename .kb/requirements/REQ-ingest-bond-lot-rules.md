---
title: A treasury bond lot entered by hand must be refused unless it has a purchase date no later than today, a margin of zero or more, and can be valued on its purchase date by the Polish bond rules as a cash purchase.
status: open
text_ref: docs/DESIGN.md#D-9
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-9
  - lane:strict
semantic_text: A treasury bond lot entered by hand must be refused unless it has a purchase date no later than today, a margin of zero or more, and can be valued on its purchase date by the Polish bond rules as a cash purchase.
logic_claims:
  - CLAIM-21FAF5872FD12268
semantic_clauses:
  - A treasury bond lot entered by hand must be refused unless it has a purchase date no later than today, a margin of zero or more, and can be valued on its purchase date by the Polish bond rules as a cash purchase.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: e6c75355a39378d3d3bb984c306a5b2033f85ee7825b89049d17a4e0b2ed153f
semantic_inventory:
  - claim_key: CLAIM-21FAF5872FD12268
    claim_text: A treasury bond lot entered by hand must be refused unless it has a purchase date no later than today, a margin of zero or more, and can be valued on its purchase date by the Polish bond rules as a cash purchase
    role: exception
    status: modeled
    span:
      start: 0
      end: 211
origin:
  kind: agent
  recorded_at: '2026-10-29T01:14:12.885Z'
id: REQ-ingest-bond-lot-rules
type: req
---
A treasury bond lot entered by hand must be refused unless it has a purchase date no later than today, a margin of zero or more, and can be valued on its purchase date by the Polish bond rules as a cash purchase.

## Context

D-9 names the fields (series, quantity, purchase date) and says valuation is computed; valuing needs the series terms too, so a lot carries them. Refusing a future date, reusing pl.ValueLot as the check, and assuming a cash purchase at the nominal are this task's narrow choices, listed for the author. Modeled by Claude Code for the manual bond lot task; the author has not reviewed it yet.

## Source

> **[D-9]** M0 data paths: XTB through uploaded xStation export files; treasury bonds through manual lot entry (series, quantity, purchase date) with valuation computed; an import of the bond service's own export is added if one exists (OQ-22).

docs/DESIGN.md#D-9
