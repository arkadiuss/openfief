---
title: Takes Fees From Reviewed Table = true
status: active
text_ref: design-decisions:D-97#2
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-97-2
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: plugins.treasury_bonds
property_key: takes_fees_from_reviewed_table
canonical_key: mcp-kibi-model-requirement:plugins.treasury_bonds:takes_fees_from_reviewed_table:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-647CE0F76DBE5AC3
claim_text: The treasury bond connector must take each series' early-redemption fee from a reviewed table per bond type, taken from the issue letters, with the dates each fee applies from.
origin:
  kind: agent
  recorded_at: '2026-11-07T12:37:37.321Z'
id: FACT-PROP-D24BE31B5E6E9DE6
type: fact
---
