---
title: The GUS adapter must return only final year-on-year CPI readings of a year from GUS's public API, in month order, as exact fractions of the same month a year before, and none for a year with nothing published.
status: open
text_ref: docs/DESIGN.md#D-9
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-9
  - lane:strict
semantic_text: The GUS adapter must return only final year-on-year CPI readings of a year from GUS's public API, in month order, as exact fractions of the same month a year before, and none for a year with nothing published.
logic_claims:
  - CLAIM-D645395D0BF038BF
semantic_clauses:
  - The GUS adapter must return only final year-on-year CPI readings of a year from GUS's public API, in month order, as exact fractions of the same month a year before, and none for a year with nothing published.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 138ba7a701ceac8691a3d5cf2942b85bba03f3f154cc298229ab54a30c49d791
semantic_inventory:
  - claim_key: CLAIM-D645395D0BF038BF
    claim_text: The GUS adapter must return only final year-on-year CPI readings of a year from GUS's public API, in month order, as exact fractions of the same month a year before, and none for a year with nothing published
    role: normative
    status: modeled
    span:
      start: 0
      end: 208
origin:
  kind: agent
  recorded_at: '2026-10-29T03:56:36.629Z'
id: REQ-valuation-gus-cpi
type: req
---
The GUS adapter must return only final year-on-year CPI readings of a year from GUS's public API, in month order, as exact fractions of the same month a year before, and none for a year with nothing published.

## Context

D-9 decides that bond valuation is computed, and DESIGN.md's context diagram has the treasury bonds connector fetch CPI from GUS. Using the SDP API's indicators 1832 (up to 2025) and 2728 (from 2026, COICOP 2018) and skipping preliminary values are this task's choices, listed for the author. Modeled by Claude Code for the bond indices task; the author has not reviewed it yet.

## Source

> **[D-9]** M0 data paths: XTB through uploaded xStation export files; treasury bonds through manual lot entry (series, quantity, purchase date) with valuation computed; an import of the bond service's own export is added if one exists (OQ-22).

docs/DESIGN.md#D-9
