---
title: Owner Confirms Or Overrides Bond Fee = true
status: active
text_ref: docs/DESIGN.md#D-97
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-97
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: valuation.settings
property_key: owner_confirms_or_overrides_bond_fee
canonical_key: mcp-kibi-model-requirement:valuation.settings:owner_confirms_or_overrides_bond_fee:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-3234B71D811BBE05
claim_text: The owner must be able to confirm a series' assumed early-redemption fee, which revalues nothing, or override any series' fee, which revalues every account that held the series when the amount changes, and confirming a fee that is not assumed must be refused.
origin:
  kind: agent
  recorded_at: '2026-11-07T13:26:58.784Z'
id: FACT-PROP-483D9C11B56FD72F
type: fact
---
