---
title: Synthetic Only = true
status: active
text_ref: docs/DESIGN.md#GR-12
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-gr-12
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: demo.dataset
property_key: synthetic_only
canonical_key: mcp-kibi-model-requirement:demo.dataset:synthetic_only:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-803FF4F690854ED9
claim_text: 'The demo dataset must hold no real account data: every instrument must have a synthetic ISIN with a valid check digit and the TESTD marker after its country code, IBKR statements must name only the placeholder account U0000000, account IDs must start with demo-, and the only real figures may be public ones: bond series terms, GUS CPI readings and NBP table A rates.'
origin:
  kind: agent
  recorded_at: '2026-10-28T09:31:20.910Z'
id: FACT-PROP-EAD7AED7D96A3512
type: fact
---
