---
title: The treasury bond connector must map each valid lot to a deposit of its cost and a buy of its bonds on its purchase date in its configured account, keyed by series and purchase date, and deliver each series' terms once under an instrument whose symbol is the series code.
status: open
text_ref: docs/DESIGN.md#D-9
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-9
  - lane:strict
semantic_text: The treasury bond connector must map each valid lot to a deposit of its cost and a buy of its bonds on its purchase date in its configured account, keyed by series and purchase date, and deliver each series' terms once under an instrument whose symbol is the series code.
logic_claims:
  - CLAIM-076844F381AA59B7
semantic_clauses:
  - The treasury bond connector must map each valid lot to a deposit of its cost and a buy of its bonds on its purchase date in its configured account, keyed by series and purchase date, and deliver each series' terms once under an instrument whose symbol is the series code.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: f99a7b2dcea1911cd1ff40e1b6e65810578aab083472de8d70b1d2e2e304b40f
semantic_inventory:
  - claim_key: CLAIM-076844F381AA59B7
    claim_text: The treasury bond connector must map each valid lot to a deposit of its cost and a buy of its bonds on its purchase date in its configured account, keyed by series and purchase date, and deliver each series' terms once under an instrument whose symbol is the series code
    role: normative
    status: modeled
    span:
      start: 0
      end: 270
origin:
  kind: agent
  recorded_at: '2026-10-29T01:14:20.029Z'
id: REQ-treasurybonds-lot-mapping
type: req
---
The treasury bond connector must map each valid lot to a deposit of its cost and a buy of its bonds on its purchase date in its configured account, keyed by series and purchase date, and deliver each series' terms once under an instrument whose symbol is the series code.

## Context

D-9 decides manual lot entry; D-18 makes keys stable so re-entry supersedes. Booking a deposit plus a buy dated the purchase date (rather than an opening lot at go-live), the derived ISIN and keying by series and day are this task's choices, listed for the author. Modeled by Claude Code for the manual bond lot task; the author has not reviewed it yet.

## Source

> **[D-9]** M0 data paths: XTB through uploaded xStation export files; treasury bonds through manual lot entry (series, quantity, purchase date) with valuation computed; an import of the bond service's own export is added if one exists (OQ-22).

docs/DESIGN.md#D-9
