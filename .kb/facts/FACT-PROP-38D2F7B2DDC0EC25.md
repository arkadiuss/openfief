---
title: Rebuilds Rows From Date = true
status: active
text_ref: docs/DESIGN.md#D-36
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-36
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: valuation.revaluation
property_key: rebuilds_rows_from_date
canonical_key: mcp-kibi-model-requirement:valuation.revaluation:rebuilds_rows_from_date:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-5EA93812331A6EF6
claim_text: A revaluation of an account from a date must rebuild its holding valuations for every day from that date, or its first transaction when later, to today, one per holding and day keyed by account, instrument, date and basis, replacing the stored ones from that date on, so that running it again gives the same rows.
origin:
  kind: agent
  recorded_at: '2026-10-28T07:29:12.443Z'
id: FACT-PROP-38D2F7B2DDC0EC25
type: fact
---
