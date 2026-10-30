---
title: Suggestions Listed = true
status: active
text_ref: docs/DESIGN.md#D-105
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-105
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: views.transfers
property_key: suggestions_listed
canonical_key: mcp-kibi-model-requirement:views.transfers:suggestions_listed:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-7070D9D4423D1BE3
claim_text: The read API must list as suggested transfers each withdrawal that is not in a confirmed transfer, paired with the nearest deposit in date into another account of exactly the same amount and currency at most five calendar days apart that is not in a confirmed transfer, not already suggested and not rejected with it, together with the confirmed and rejected pairs.
origin:
  kind: agent
  recorded_at: '2026-10-30T06:20:27.676Z'
id: FACT-PROP-167C42F51FBA6FB2
type: fact
---
