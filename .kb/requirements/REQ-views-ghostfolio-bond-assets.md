---
title: 'The Ghostfolio view must make each treasury bond a MANUAL asset whose symbol is a UUID derived from its series code and whose name is the code, and must send its full price history on every sync: each stored valuation divided by the quantity held at the end of that day.'
status: closed
text_ref: docs/DESIGN.md#D-64
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-64
  - lane:strict
semantic_text: 'The Ghostfolio view must make each treasury bond a MANUAL asset whose symbol is a UUID derived from its series code and whose name is the code, and must send its full price history on every sync: each stored valuation divided by the quantity held at the end of that day.'
logic_claims:
  - CLAIM-DCB52589DDA330C5
semantic_clauses:
  - 'The Ghostfolio view must make each treasury bond a MANUAL asset whose symbol is a UUID derived from its series code and whose name is the code, and must send its full price history on every sync: each stored valuation divided by the quantity held at the end of that day.'
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 7618ee536928256e38155a370cc315664ae0cd0ac4b643ff06456b6c0046ce62
semantic_inventory:
  - claim_key: CLAIM-DCB52589DDA330C5
    claim_text: 'The Ghostfolio view must make each treasury bond a MANUAL asset whose symbol is a UUID derived from its series code and whose name is the code, and must send its full price history on every sync: each stored valuation divided by the quantity held at the end of that day'
    role: normative
    status: modeled
    span:
      start: 0
      end: 269
origin:
  kind: agent
  recorded_at: '2026-10-29T00:54:53.513Z'
id: REQ-views-ghostfolio-bond-assets
type: req
---
The Ghostfolio view must make each treasury bond a MANUAL asset whose symbol is a UUID derived from its series code and whose name is the code, and must send its full price history on every sync: each stored valuation divided by the quantity held at the end of that day.

## Context

Closed: superseded by REQ-views-ghostfolio-asset-keys. D-64 was amended: a bond's symbol derives from the instance too. Modeled by Claude Code for the Ghostfolio work package (impl/ghostfolio-2); the author has not reviewed it yet.
