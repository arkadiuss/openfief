---
title: Bonds Are Manual Assets With History = true
status: active
text_ref: docs/DESIGN.md#D-64
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-64
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: views.ghostfolio_plugin
property_key: bonds_are_manual_assets_with_history
canonical_key: mcp-kibi-model-requirement:views.ghostfolio_plugin:bonds_are_manual_assets_with_history:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-DCB52589DDA330C5
claim_text: 'The Ghostfolio view must make each treasury bond a MANUAL asset whose symbol is a UUID derived from its series code and whose name is the code, and must send its full price history on every sync: each stored valuation divided by the quantity held at the end of that day.'
origin:
  kind: agent
  recorded_at: '2026-10-29T00:54:51.259Z'
id: FACT-PROP-9D2C280D4474EFF5
type: fact
---
