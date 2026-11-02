---
title: Keeps Full Cpi Years = true
status: active
text_ref: docs/DESIGN.md#D-143
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-143
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: treasurybonds.coupons
property_key: keeps_full_cpi_years
canonical_key: mcp-kibi-model-requirement:treasurybonds.coupons:keeps_full_cpi_years:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-AB9F564322F20A87
claim_text: The treasury-bonds connector must keep each year of CPI that GUS served in full across syncs and ask GUS for at most four years it lacks per sync, so every coupon and maturity since each lot's purchase is booked within a few syncs however many years its lots span.
origin:
  kind: agent
  recorded_at: '2026-11-02T09:18:35.101Z'
id: FACT-PROP-4712F181BA381E8D
type: fact
---
