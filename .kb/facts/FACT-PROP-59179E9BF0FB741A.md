---
title: Keys Manual Assets By Instance And User = true
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
property_key: keys_manual_assets_by_instance_and_user
canonical_key: mcp-kibi-model-requirement:views.ghostfolio_plugin:keys_manual_assets_by_instance_and_user:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-D44C838B03B0EBDD
claim_text: 'The Ghostfolio view must make each treasury bond a MANUAL asset whose symbol is a UUID derived from the view instance''s name, the Ghostfolio user''s ID and the series code and whose name is the code, must key each currency''s fee asset the same way from the currency, and must send each bond''s full price history on every sync: each stored valuation divided by the quantity held at the end of that day.'
origin:
  kind: agent
  recorded_at: '2026-10-30T01:21:11.301Z'
id: FACT-PROP-59179E9BF0FB741A
type: fact
---
