---
title: Period Rate Follows Index Plus Margin = true
status: active
text_ref: docs/DESIGN.md#D-9
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-9
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: pl.bond_value
property_key: period_rate_follows_index_plus_margin
canonical_key: mcp-kibi-model-requirement:pl.bond_value:period_rate_follows_index_plus_margin:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-5C34FC18CE32ECEA
claim_text: From the second interest period, an inflation-indexed bond's rate must be the CPI reading plus the series margin with a negative CPI counted as zero, a ROR or DOR bond's rate must be the NBP reference rate plus the margin, and OTS and TOS must keep their first-period rate.
origin:
  kind: agent
  recorded_at: '2026-10-28T05:21:52.427Z'
id: FACT-PROP-B326EF7E2849A6D2
type: fact
---
